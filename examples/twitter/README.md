# Send Tweets via kubeless

This cool example shows off two great features of kubeless:

* How we can load dependencies of functions at runtime via an init container.
* How by using the Kubernetes python client you can retrieve a secret to talk to an external API

The function is then exposed as service and we can POST data to this service to send a tweet.

## Create a secret that holds your Twitter API tokens

You need to have created a [Twitter application](https://dev.twitter.com) and obtained secrets and tokens to use it.

Store those tokens and keys in a Kubernetes secrets. For example you can use `kubectl` directly like so:

```
kubectl create secret generic twitter --from-literal=consumer_key=<> --from-literal=consumer_secret=<> --from-literal=token_key=<> --from-literal=token_secret=<>
```

## Write a `requirements.txt` file

The function we will write uses this [Python twitter](https://github.com/bear/python-twitter) module as well as the Kubernetes [Python client](https://github.com/kubernetes-incubator/client-python).
Define those dependencies in a `requirements.txt` file.

```
python-twitter
kubernetes
```

## Write a function to retrieve the secret and send a tweet

```
import base64
import twitter

from kubernetes import client, config

config.load_incluster_config()

v1=client.CoreV1Api()

for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'twitter':
        consumer_key = base64.b64decode(secrets.data['consumer_key'])
        consumer_secret = base64.b64decode(secrets.data['consumer_secret'])
        token_key = base64.b64decode(secrets.data['token_key'])
        token_secret = base64.b64decode(secrets.data['token_secret'])

api = twitter.Api(consumer_key=consumer_key,
                  consumer_secret=consumer_secret,
                  access_token_key=token_key,
                  access_token_secret=token_secret)

def tweet(context):
    msg = context.json
    status = api.PostUpdate(msg['tweet'])
```

The `context` object is a request that comes from the `bottle` wrapper of the kubeless python runtime.

Store this function in a `send-tweet.py` file

## Deploy the function with `kubeless`

```
kubeless function deploy tweet --trigger-http --runtime python2.7 --handler send-tweet.tweet --from-file send-tweet.py --dependencies requirements.txt
```

The `handler` referes to the file name of the function and the method name that we will trigger. Note that the dependency file is loaded with the `--dependencies` option.

## Call the function via its Kubernetes service

kubeless will launch a container. An init container will first load all the dependencies and then your function will be wrapped by the kubeless runtime.

If you run a proxy you can then call your function:

```
curl --data '{"tweet":"this rocks from kubeless"}' localhost:8080/api/v1/proxy/namespaces/default/services/tweet/ --header "Content-Type:application/json"
```

Your container will log a `200 OK` response and your twitter timeline will be populated with the message.

```
$ kubectl logs tweet-1222248550-17g1w
Bottle v0.12.10 server starting up (using WSGIRefServer())...
Listening on http://0.0.0.0:8080/
Hit Ctrl-C to quit.

172.17.0.1 - - [16/Mar/2017 10:52:45] "POST / HTTP/1.1" 200 0
```

## Explore

What happened ?

You created a secret. Then by creating the function with `kubeless`, a deployment was created, which launched the container (and the init container before that).
The runtime wrapped your function with an HTTP REST endpoint and exposed it via a service.

No need to build a container before end, `kubeless` does all of it.

```
kubectl get pods,deployments,svc,configmaps,secrets
NAME                          READY     STATUS    RESTARTS   AGE
po/tweet-1222248550-17g1w   1/1       Running   0          1h

NAME             DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/tweet   1         1         1            1           1h

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)          AGE
svc/kubernetes   10.0.0.1     <none>        443/TCP          2h
svc/tweet      10.0.0.137   <nodes>       8080:31753/TCP   1h

NAME         DATA      AGE
cm/tweet   3         1h

NAME                          TYPE                                  DATA      AGE
secrets/default-token-21dmt   kubernetes.io/service-account-token   3         2h
secrets/twitter               Opaque                                4         2h
```
