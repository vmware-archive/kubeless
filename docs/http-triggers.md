# Expose and secure Kubeless functions

Kubeless leverages [Kubernetes ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to provide routing for functions. By default, a deployed function will be matched to a Kubernetes service using ClusterIP as the service. That means that the function is not exposed publicly. Because of that, we provide the `kubeless trigger http` command that can make a function publicly available. This guide provides a quick sample on how to do it.

## Ingress controller

In order to create routes for functions in Kubeless, you must have an Ingress controller running. There are several options to deploy it. In this document we point to several different solutions that you can choose:

> Note: In case Kubeless is running in a GKE cluster you will need to disable the default Ingress controller provided by GKE. The native controller doesn't work with services that have a type different  than NodePort (see [this issue](https://github.com/kubernetes/ingress-nginx/issues/1417)). In order to expose a Kubeless function, disable the default controller and deploy one of the options described below.

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

[Traefik](http://traefik.io) provides an Ingress controller as well. To deploy it follow the steps described at [this guide](https://docs.traefik.io/user-guide/kubernetes/). As a result, you will be able to see the traefik controller running in the `kube-system` namespace:

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

Once you have one of the supported Ingress Controller it is possible to enable TLS using a certificate:

 - Automatically generated using Let's Encrypt and [cert-manager](https://github.com/jetstack/cert-manager)
 - Self signed
 - Provided by a certificate issuer

### Using Let’s Encrypt’s CA

When you have running Kube-lego, you can deploy function and create an HTTP trigger with flag `--enableTLSAcme` enabled as below:

```console
$ kubeless trigger http create get-python --function-name get-python --path get-python --enableTLSAcme
```

Running the above command, Kubeless will automatically create a ingress object with annotation `kubernetes.io/tls-acme: 'true'` set which will be used by Kube-lego to configure the service certificate.

### Create a self-signed certificate

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

### Use an existing certificate

Now that you have a certificate, you can use it to setup TLS for the HTTP trigger, there by securing functions:

```console
$ kubeless trigger http create get-python --function-name get-python --hostname foo.bar.com --tls-secret secret-name
```

Once the Ingress rule has been deployed you can verify that the function is accessible trough HTTPS:

```console
$ kubectl get ingress
NAME             HOSTS            ADDRESS          PORTS     AGE
get-python       foo.bar.com      192.168.99.100   80, 443   4m
$ curl -k https://192.168.99.100 --header 'Host: foo.bar.com'
hello world
```

## Enable Basic Authentication

Once you have one of the supported Ingress Controller it is possible to enable Basic Authentication either:

 - Creating a secret with the content of the user to authenticate. This is valid for the Nginx and Traefik controllers.
 - Adding the Kong plugin for basic authentication.

### Enable Basic Authentication with Nginx or Traefik

For enabling authentication for a function, the first thing is creating a secret with the user and password:

```console
$ htpasswd -cb auth foo bar
Adding password for user foo
$ kubectl create secret generic basic-auth --from-file=auth
secret "basic-auth" created
```

Now you just need to create a HTTP trigger using that secret.

```console
$ kubeless trigger http create get-python --function-name get-python --basic-auth-secret basic-auth --gateway nginx
INFO[0000] HTTP trigger get-python created in namespace default successfully!
```

> Note: The command is the same for the case of Traefik, just use `--gateway traefik` instead

Once the Ingress rule has been deployed you can verify that the function is accessible just for the proper user and password:

```console
$ kubectl get ingress
NAME         HOSTS                              ADDRESS          PORTS     AGE
get-python   get-python.192.168.99.100.nip.io   192.168.99.100   80        1m
$ curl --header 'Host: get-python.192.168.99.100.nip.io' 192.168.99.100
<html>
<head><title>401 Authorization Required</title></head>
<body bgcolor="white">
<center><h1>401 Authorization Required</h1></center>
<hr><center>nginx/1.13.7</center>
</body>
</html>
$ curl -u foo:bar --header 'Host: get-python.192.168.99.100.nip.io' 192.168.99.100
hello world
```

### Enable Basic Authentication with Kong

It is not yet supported to create an HTTP trigger with basic authentication using Kong as backend but the steps to do it manually are pretty simple. It is possible to do so using Kong plugins. In the [next section](#enable-kong-security-plugins) we explain how to enable any of the available Kong plugins and in particular we explain how to enable the basic-auth plugin.

## Enable Kong Security plugins

Kong has available several free [plugins](https://konghq.com/plugins/) that can be used along with the Kong Ingress controller for securing the access to Kubeless functions. In particular, the list of security plugins that can be used is:

 - Basic Authentication
 - Key Authentication
 - OAuth 2.0
 - JWT
 - ACL
 - HMAC Authentication
 - LDAP Authentication

Once you have Kong and its Ingress controller running in your cluster the generic steps to use any plugin are:

 - Deploy a basic HTTP trigger for the target function using `--gateway kong`.
 - Create a Kubernetes object for the plugin you want to use.
 - Add a Kong Consumer.
 - Create the specific credentials or follow any additional steps that the plugin may require.
 - Associate the credentials/plugin with the Ingress object created in the first step.

The specific steps that are required to use a plugin can be found in the [plugins](https://konghq.com/plugins/) page. As an example we will configure the plugin [basic-auth](https://getkong.org/plugins/basic-authentication/) for our function `get-python`.

### Deploy a basic HTTP trigger

First we need to create a HTTP trigger to generate the Ingress object that will expose our function.

```console
$ kubeless trigger http create get-python --function-name get-python --gateway kong --hostname foo.bar.com
INFO[0000] HTTP trigger get-python created in namespace default successfully!
```

### Add the basic-auth plugin

The next step is creating the Custom Resource related to the Kong basic authentication plugin. You can see the possible configuration options available in the [plugin documentation](https://getkong.org/plugins/basic-authentication).

```console
$ echo "
apiVersion: configuration.konghq.com/v1
kind: KongPlugin
metadata:
  name: basic-auth
consumerRef: basic-auth
config:
  hide_credentials: false
" | kubectl create -f -
kongplugin "basic-auth" created
```

#### Create a Consumer

Now we need a [`Consumer`](https://getkong.org/docs/0.13.x/getting-started/adding-consumers/#adding-consumers) for the plugin.

```console
$ echo "
apiVersion: configuration.konghq.com/v1
kind: KongConsumer
metadata:
  name: basic-auth
username: user
" | kubectl create -f -
kongconsumer "basic-auth" created
```

#### Create user credentials

Now that we have a consumer we need to create the basic authentication credentials that the function is going to use:

```console
$ echo "
apiVersion: configuration.konghq.com/v1
kind: KongCredential
metadata:
  name: basic-auth
consumerRef: basic-auth
type: basic-auth
config:
  username: user
  password: pass
" | kubectl create -f -
kongcredential "basic-auth" created
```

#### Associate the credentials with the Ingress object

The final step is to enable the credentials and the plugin for the function. For doing so we just need to add an `Annotation` in the Ingress object that we generated in the first step:

```console
$ kubectl patch ingress get-python \
 -p '{"metadata":{"annotations":{"basic-auth.plugin.konghq.com":"basic-auth"}}}'
ingress "get-python" patched
```

Now that the plugin has been enabled we can verify that it is working:

```console
$ export PROXY_IP=$(minikube   service -n kong kong-proxy --url --format "{{ .IP }}" | head -1)
$ export HTTP_PORT=$(minikube  service -n kong kong-proxy --url --format "{{ .Port }}" | head -1)
$ curl --header "Host: foo.bar.com" ${PROXY_IP}:${HTTP_PORT}
{"message":"Unauthorized"}
$ curl -u user:pass --header "Host: foo.bar.com" ${PROXY_IP}:${HTTP_PORT}
hello world
```
