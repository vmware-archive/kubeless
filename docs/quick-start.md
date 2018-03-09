# Installation

Installation is made of three steps:

* Download the `kubeless` CLI from the [release page](https://github.com/kubeless/kubeless/releases). (OSX users can also use [brew](https://brew.sh/): `brew install kubeless`).
* Create a `kubeless` namespace (used by default)
* Then use one of the YAML manifests found in the release page to deploy kubeless. It will create a _functions_ Custom Resource Definition and launch a controller.

There are several kubeless manifests being shipped for multiple k8s environments (non-rbac, rbac and openshift), pick the one that corresponds to your environment:

* `kubeless-$RELEASE.yaml` is used for non-RBAC Kubernetes cluster.
* `kubeless-rbac-$RELEASE.yaml` is used for RBAC-enabled Kubernetes cluster.
* `kubeless-openshift-$RELEASE.yaml` is used to deploy Kubeless to OpenShift (1.5+).

For example, this below is a show case of deploying kubeless to a non-RBAC Kubernetes cluster.

```console
$ export RELEASE=v0.4.0
$ kubectl create ns kubeless
$ kubectl create -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless-$RELEASE.yaml

$ kubectl get pods -n kubeless
NAME                                   READY     STATUS    RESTARTS   AGE
kubeless-controller-3331951411-d60km   1/1       Running   0          1m

$ kubectl get deployment -n kubeless
NAME                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kubeless-controller   1         1         1            1           1m

$ kubectl get customresourcedefinition
NAME                    AGE
functions.kubeless.io   1h

$ kubectl get functions
NAME         AGE
get-python   1d
```

If you have installed Kubeless into some other namespace (which is not called `kubeless`) or changed the name of the config file from kubeless-config to something else, then you have to export the kubeless namespace and the name of kubeless config as environment variables before using kubless cli. This can be done as follows:

```bash
$ export KUBELESS_NAMESPACE=<name of namespace>
$ export KUBELESS_CONFIG=<name of config file>
```

or the following information can be added to `functions.kubeless.io` `CustomResourceDefinition` as `annotations`. E.g. below `CustomResourceDefinition` will signify `kubeless-controller` is installed in namespace `kubless-new-namespace` and config name is `kubeless-config-new-name`

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
description: Kubernetes Native Serverless Framework
kind: CustomResourceDefinition
metadata:
  name: functions.kubeless.io
  annotations:
    kubeless.io/namespace: kubless-new-namespace
    kubeless.io/config: kubeless-config-new-name
spec:
  group: kubeless.io
  names:
    kind: Function
    plural: functions
    singular: function
  scope: Namespaced
  version: v1beta1
```

The priority of deciding the `namespace` and `config name` (highest to lowest) is:

- Environment variables
- Annotations in `functions.kubeless.io` CRD
- default: `namespace` is `kubeless` and `ConfigMap` is `kubeless-config`  

You are now ready to create functions.

# Usage

You can use the CLI to create a function. Functions have three possible types:

* http triggered (function will expose an HTTP endpoint)
* pubsub triggered (function will consume event on a specific topic; a running kafka cluster on your k8s is required)
* schedule triggered (function will be called on a cron schedule)

## HTTP function

Here is a toy:

```python
def foobar(context):
   print context.json
   return context.json
```

You create it with:

```console
$ kubeless function deploy get-python --runtime python2.7 \
                                --from-file test.py \
                                --handler test.foobar \
                                --trigger-http
INFO[0000] Deploying function...
INFO[0000] Function get-python submitted for deployment
INFO[0000] Check the deployment status executing 'kubeless function ls get-python'
```

Let's dissect the command:

* `get-python`: This is the name of the function we want to deploy.
* `--runtime python2.7`: This is the runtime we want to use to run our function. Available runtimes are shown in the help information.
* `--from-file test.py`: This is the file containing the function code. It is supported to specify a zip file as far as it doesn't exceed the maximum size for an etcd entry (1 MB).
* `--handler test.foobar`: This specifies the file and the exposed function that will be used when receiving requests. In this example we are using the function `foobar` from the file `test.py`.
* `--env` to pass env vars to the function like `--env foo=bar,bar=foo`. See the [detail](https://github.com/kubeless/kubeless/pull/316#issuecomment-332172876)
* `--trigger-http`: This sets the function trigger.

Other available options are:

* `--trigger-http` to trigger the function using HTTP requests.
* `--trigger-topic` to trigger the function with a certain Kafka topic. See the [next example](#pubsub-function).
* `--timeout` to specify the timeout (in seconds) for the function to complete its execution (default "180")
* `--schedule` to trigger the function following a certain schedule using Cron notation. F.e. `--schedule "*/10 * * * *"` would trigger the function every 10 minutes.
* `--secrets`: This sets a list of Secrets to be mounted as Volumes to the functions pod. They will be available in the path `/<secret_name>`.

You can find the rest of options available when deploying a function executing `kubeless function deploy --help`

You will see the function custom resource created:

```console
$ kubectl get functions
NAME         AGE
get-python   1h

$ kubeless function ls
NAME           	NAMESPACE	HANDLER       RUNTIME  	TYPE  	TOPIC      	DEPENDENCIES	STATUS
get-python     	default  	helloget.foo  python2.7	HTTP  	           	            	1/1 READY
```

You can then call the function with:

```console
$ kubeless function call get-python --data '{"echo": "echo echo"}'
{"echo": "echo echo"}
```

Or you can curl directly with `kubectl proxy`
using an [apiserver proxy
URL](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#manually-constructing-apiserver-proxy-urls).
For example:

```console
$ kubectl proxy -p 8080 &

$ curl -L --data '{"Another": "Echo"}' \
  --header "Content-Type:application/json" \
  localhost:8080/api/v1/proxy/namespaces/default/services/get-python:http-function-port/
{"Another": "Echo"}
```

Kubeless also supports [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) which means you can provide your custom URL to the function. Please refer to [this doc](./routing) for more details.

## PubSub function

We provide several [PubSub runtimes](https://hub.docker.com/r/kubeless/), which has suffix `event-consumer`, which help you to quickly deploy your function with PubSub mechanism. The PubSub function will expect to consume input messages from a predefined Kafka topic which means Kafka is required. In Kubeless [release page](https://github.com/kubeless/kubeless/releases), you can find the manifest to quickly deploy a collection of Kafka and Zookeeper statefulsets. If you have a Kafka cluster already running in the same Kubernetes environment, you can also deploy PubSub function with it. Check out [this tutorial](./use-existing-kafka.md) for more details how to do that.

If you want to deploy the manifest we provide to deploy Kafka and Zookeper execute the following command:

```console
$ kubectl create -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kafka-zookeeper-$RELEASE.yaml
```

> NOTE: Kafka statefulset uses a PVC (persistent volume claim). Depending on the configuration of your cluster you may need to provision a PV (Persistent Volume) that matches the PVC or configure dynamic storage provisioning. Otherwise Kafka pod will fail to get scheduled. Also note that Kafka is only required for PubSub functions, you can still use http triggered functions. Please refer to [PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) documentation on how to provision storage for PVC.

Once deployed, you can verify two statefulsets up and running:

```
$ kubectl -n kubeless get statefulset
NAME      DESIRED   CURRENT   AGE
kafka     1         1         40s
zoo       1         1         42s

$ kubectl -n kubeless get svc
NAME        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
broker      ClusterIP   None            <none>        9092/TCP            1m
kafka       ClusterIP   10.55.250.89    <none>        9092/TCP            1m
zoo         ClusterIP   None            <none>        9092/TCP,3888/TCP   1m
zookeeper   ClusterIP   10.55.249.102   <none>        2181/TCP            1m
```

Now you can deploy a pubsub function. A function can be as simple as:

```python
def foobar(context):
    print context
    return context
```

You create it the same way than an _HTTP_ function except that you specify a `--trigger-topic`.

```console
$ kubeless function deploy test --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-topic test-topic
```

After that you can invoke them publishing messages in that topic. To allow you to easily manage topics `kubeless` provides a convenience function `kubeless topic`. You can create/delete and publish to a topic easily.

```console
$ kubeless topic create test-topic
$ kubeless topic publish --topic test-topic --data "Hello World!"
```

You can check the result in the pod logs:

```console
$ kubectl logs test-695251588-cxwmc
Hello World!
```

## Other commands

You can delete and list functions:

```console
$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     TYPE    TOPIC
test        default     test.foobar python2.7   PubSub  test-topic

$ kubeless function delete test

$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     TYPE    TOPIC
```

You can create, list and delete PubSub topics:

```console
$ kubeless topic create another-topic
Created topic "another-topic".

$ kubeless topic delete another-topic

$ kubeless topic ls
```

You can also see the list of supported runtimes:

```console
$ kubeless get-server-config
INFO[0000] Current Server Config:
INFO[0000] Supported Runtimes are: python2.7, python3.4, python3.6, nodejs6, nodejs8, ruby2.4, dotnetcore2.0
```

## Examples

See the [examples](https://github.com/kubeless/kubeless/tree/master/examples) directory for a list of various examples. Minio, SLACK, Twitter etc ...

Also checkout the [functions repository](https://github.com/kubeless/functions),
where we're building a library of ready to use kubeless examples, including an
[incubator](https://github.com/kubeless/functions/tree/master/incubator)
to encourage contributions from the community - **your PR is welcome** ! :)
