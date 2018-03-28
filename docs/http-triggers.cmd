# HTTP triggers

Kubeless leverages [Kubernetes ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to provide routing for functions. By default, a deployed function will be matched to a Kubernetes service using ClusterIP as the service. That means that the function is not exposed publicly. Because of that, we provide the `kubeless route` command that can make a function publicly available. This guide provides a quick sample on how to do it.

## Ingress controller

In order to create routes for functions in Kubeless, you must have an Ingress controller running. There are several options to deploy it. You can deploy it manually via the [manifest](https://github.com/kubeless/kubeless/blob/master/manifests/ingress/ingress-controller-http-only.yaml) we provide.

If you are on GKE, you can try [this](https://github.com/kubernetes/ingress-gce). Minikube also provide an addon for ingress, you can enable it executing `minikube addons enable ingress`. Please note that if you're intending to use our provided manifest on minikube, please disable the ingress addon.

## Deploy function with Kubeless CLI

Try our example:

```console
$ cd examples
$ kubeless function deploy get-python \
                    --runtime python2.7 \
                    --handler helloget.foo \
                    --from-file python/helloget.py

$ kubectl get po
NAME                          READY     STATUS    RESTARTS   AGE
get-python-1796153810-krrf3   1/1       Running   0          2s

$ kubectl get svc
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
get-python   10.0.0.26    <none>        8080/TCP   44s
```

## Create HTTP trigger

Kubeless support command to create HTTP trigger using which you can control route to function.

```console
$ kubeless trigger http create --help
Create a http trigger

Usage:
  kubeless trigger http create <http_trigger_name> FLAG [flags]

Flags:
      --enableTLSAcme          If true, routing rule will be configured for use with kube-lego
      --function-name string   Name of the function to be associated with trigger
  -h, --help                   help for create
      --hostname string        Specify a valid hostname for the function
      --namespace string       Specify namespace for the function
      --path string            Ingress path for the function
```

We will create a http trigger to `get-python` function:

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python
```

This command will create an ingress object. We can see it with kubectl (this guide is run on minikube):

```console
$ kubectl get ing
NAME           HOSTS                              ADDRESS          PORTS     AGE
get-python    get-python.192.168.99.100.nip.io    192.168.99.100   80        59s
```

Kubeless creates a default hostname in form of <function-name>.<master-address>.nip.io. Alternatively, you can provide a real hostname with `--hostname` flag like this:

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python --hostname example.com
$ kubectl get ing
NAME          HOSTS                              ADDRESS          PORTS     AGE
get-python    example.com                                          80        6s
```

But you have to make sure your hostname is configured properly.

You can test the created HTTP trigger with the following command:

```console
$ curl --data '{"Another": "Echo"}' \
  --header "Host: get-python.192.168.99.100.nip.io" \
  --header "Content-Type:application/json" \
  192.168.99.100/get-python
{"Another": "Echo"}
```

## Enable TLS

By default, Kubeless doesn't take care of setting up TLS for its functions. You can do it manually by following the [standard procedure](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) of securing ingress. There is also a [general guideline](https://docs.bitnami.com/kubernetes/how-to/secure-kubernetes-services-with-ingress-tls-letsencrypt/) to enable TLS for your Kubernetes services using LetsEncrypt and Kube-lego written by Bitnami folks. When you have running Kube-lego, you can deploy function and create route with flag `--enableTLSAcme` enabled as below:

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python --enableTLSAcme
```

Running the above command, Kubeless will automatically create a ingress object with annotation `kubernetes.io/tls-acme: 'true'` set which will be used by Kube-lego to configure the service certificate.
