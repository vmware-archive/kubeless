# Expose and secure Kubeless functions

Kubeless leverages [Kubernetes ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to provide routing for functions. By default, a deployed function will be matched to a Kubernetes service using ClusterIP as the service. That means that the function is not exposed publicly. Because of that, we provide the `kubeless trigger http` command that can make a function publicly available. This guide provides a quick sample on how to do it.

## Ingress controller

In order to create routes for functions in Kubeless, you must have an Ingress controller running. There are several options to deploy it. In this document we point to several different solutions that you can choose:

> Note: In case Kubeless is running in a GKE cluster you will need to disable the default Ingress controller provided by GKE since it doesn't work with service with a type different than NodePort (see [this issue](https://github.com/kubernetes/ingress-nginx/issues/1417)). In order to expose a Kubeless function, disable the default controller and deploy one of the options described below.

### Minikube Ingress addon

If your cluster is running in Minikube you can enable the Ingress controller just executing:

```console
minikube addons enable ingress
```

After a couple of minutes you should be able to see the controller running in the `kube-system` namespace:

```console
$ kubectl get pod -n kube-system -l app=nginx-ingress-controller
NAME                             READY     STATUS    RESTARTS   AGE
nginx-ingress-controller-pj2pz   1/1       Running   0          25s
```

### Nginx Ingress

You can deploy a Nginx Ingress controller manually (it is the same controller than in the Minikube addon) following the instructions that can be found [here](https://github.com/kubernetes/ingress-nginx/blob/master/deploy/README.md).

### Kong Ingress

[Kong](https://getkong.org) have an Ingress controller that can be used to expose functions and secure them. You can check the deployment instructions in [their repository](https://github.com/Kong/kubernetes-ingress-controller/blob/master/deploy/README.md). Once Kong is deployed you should be able to see the controller in the `kong` namespace:

```console
kubectl get pods -n kong
NAME                                       READY     STATUS    RESTARTS   AGE
kong-56c4cc55c9-78srh                      1/1       Running   0          1h
kong-ingress-controller-79f48dd4d7-ql4vw   2/2       Running   0          1h
postgres-0                                 1/1       Running   1          22h
```

### Traefik Ingress

[Traefik](http://traefik.io) provides an Ingress Controller as well. To deploy it follow the steps described at [this guide](https://docs.traefik.io/user-guide/kubernetes/). As a result you will be able to see the traefik controller running in the `kube-system` namespace:

```console
kubectl get pod -n kube-system -l name=traefik-ingress-lb
NAME                                          READY     STATUS    RESTARTS   AGE
traefik-ingress-controller-57b4767f99-g42n2   1/1       Running   0          1m
```

## Deploy function with Kubeless CLI

Once you have a Ingress Controller running you should be able to start deploying functions and expose them publicly. First deploy a function:

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

## Expose a function

In order to expose a function, it is necessary to create a HTTP Trigger object. The Kubeless CLI provides the commands required to do so:

```console
$ kubeless trigger http create --help
Create a http trigger

Usage:
  kubeless trigger http create <http_trigger_name> FLAG [flags]

Flags:
      --basic-auth-secret string   Specify an existing secret name for basic authentication
      --enableTLSAcme              If true, routing rule will be configured for use with kube-lego
      --function-name string       Name of the function to be associated with trigger
      --gateway string             Specify a valid gateway for the Ingress. Supported: nginx, traefik, kong (default "nginx")
  -h, --help                       help for create
      --hostname string            Specify a valid hostname for the function
      --namespace string           Specify namespace for the HTTP trigger
      --path string                Ingress path for the function
      --tls-secret string          Specify an existing secret that contains a TLS private key and certificate to secure ingress
```

We will create a http trigger to `get-python` function:

```console
$ kubeless trigger http create get-python --function-name get-python
```

This command will create an ingress object. We can see it with kubectl (this guide is run on minikube):

```console
$ kubectl get ing
NAME           HOSTS                              ADDRESS          PORTS     AGE
get-python    get-python.192.168.99.100.nip.io    192.168.99.100   80        59s
```

Kubeless creates a default hostname in form of <function-name>.<master-address>.nip.io. Alternatively, you can provide a real hostname with `--hostname` flag or use a different `--path` like this:

```console
$ kubeless trigger http create get-python --function-name get-python --path echo --hostname example.com
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
  192.168.99.100/echo
{"Another": "Echo"}
```

## Enable TLS

By default, Kubeless doesn't take care of setting up TLS for its functions. You can do it manually by following the [standard procedure](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) of securing ingress. There is also a [general guideline](https://docs.bitnami.com/kubernetes/how-to/secure-kubernetes-services-with-ingress-tls-letsencrypt/) to enable TLS for your Kubernetes services using LetsEncrypt and Kube-lego written by Bitnami folks.

### Using Let’s Encrypt’s CA

When you have running Kube-lego, you can deploy function and create route with flag `--enableTLSAcme` enabled as below:

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python --enableTLSAcme
```

Running the above command, Kubeless will automatically create a ingress object with annotation `kubernetes.io/tls-acme: 'true'` set which will be used by Kube-lego to configure the service certificate.

### Using existing Certificate and Private Key

#### Create a self-signed certificate

If you don't have a working certificate it is possible to generate a dummy one to be able to use TLS with your functions. To generate the certificate and its secret execute the following:

```console
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=foo.bar.com"
Generating a 2048 bit RSA private key
..........................................................................+++
.......................................................+++
writing new private key to 'tls.key'
-----
$ kubectl create secret tls tls-secret --key tls.key --cert tls.crt
secret "tls-secret" created
```

#### Use an existing certificate

Now that you have a certificate, you can use them to setup TLS for ingress, there by securing functions:

```console
$ kubeless trigger http create get-python --function-name get-python --hostname foo.bar.com --tls-secret secret-name
```

Once the Ingress rule has been deployed you can verify that the function is accessible trough https:

```console
$ kubectl get ingress
NAME             HOSTS            ADDRESS          PORTS     AGE
get-python       foo.bar.com      192.168.99.100   80, 443   4m
$ curl -k https://192.168.99.100 --header 'Host: foo.bar.com'
hello world
```

## Enable Basic Authentication

By default, Kubeless doesn't take care about securing its exposed functions.
You can do it manually depending on your Ingress controller, some examples are:
* [Nginx](https://github.com/kubernetes/ingress-nginx/blob/master/docs/examples/auth/basic/README.md)
* [Traefik](https://docs.traefik.io/user-guide/kubernetes/#basic-authentication)

When you have a running Nginx or Traefik ingress controller, you can deploy function and create a basic authentication secured route as shown below:
The Kubernetes secret specified by `--basic-auth-secret` must exist and located within the same namespace as the http trigger.

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python --basic-auth-secret get-python-secret --gateway nginx
```

Running the above command, Kubeless will automatically create an ingress object with annotations:
* `kubernetes.io/ingress.class: nginx`
* `ingress.kubernetes.io/auth-secret: get-python-secret`
* `ingress.kubernetes.io/auth-type: basic`


