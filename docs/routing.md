# Add route to Kubeless function

Kubeless leverages [Kubernetes ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to provide routing for function. By default, a deployed function will be matched to a Kubernetes service with service type is ClusterIP that means it is not exposed publicly. We provides `kubeless route` command to make the function published. This guide give you a quick sample on how to do it.

## Ingress controller

In order to create route for function in Kubeless, you must have an Ingress controller running. There are several options to deploy it. You can deploy it manually via the [manifest](https://github.com/kubeless/kubeless/blob/master/manifests/ingress/ingress-controller-http-only.yaml) we provide. If you are on GKE, you can try [this](https://github.com/kubernetes/ingress-gce). Minikube also provide an addon for ingress, enable it with this command `minikube addons enable ingress`. Please note that if you're intending to use our provided manifest on minikube, please disable the ingress addon.

## Deploy function with Kubeless CLI

Try our example:
```
$ cd examples
$ kubeless function deploy get-python \
                    --trigger-http \
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

## Create route

Kubeless support ingress command to create route to function.
```
$ kubeless route --help
ingress command allows user to list, create, delete routing rule for function on Kubeless

Usage:
  kubeless route SUBCOMMAND [flags]
  kubeless route [command]

Available Commands:
  create      create a route to function
  delete      delete a route from Kubeless
  list        list all routes in Kubeless

Use "kubeless route [command] --help" for more information about a command.
```

We will create a route to `get-python` function:

```
$ kubeless route create route1 --function get-python
```

This command will create an ingress object. We can see it with kubectl (this guide is run on minikube):

```
$ kubectl get ing
NAME      HOSTS                              ADDRESS          PORTS     AGE
route1    get-python.192.168.99.100.nip.io   192.168.99.100   80        59s
```

Kubeless creates a default hostname in form of <function-name>.<master-address>.nip.io. Alternatively, you can provide a real hostname with `--hostname` flag like this:

```
$ kubeless route create route2 --function get-python --hostname example.com
$ kubectl get ing
NAME      HOSTS                              ADDRESS          PORTS     AGE
route1    get-python.192.168.99.100.nip.io   192.168.99.100   80        3m
route2    example.com                                         80        6s
```

But you have to make sure your hostname is configured properly.

You can test the new route with the following command:
```
$ curl --data '{"Another": "Echo"}' --header "Host: get-python.192.168.99.100.nip.io" 192.168.99.100/ --header "Content-Type:application/json"
{"Another": "Echo"}
```

## Enable TLS

By default, Kubeless doesn't take care of TLS setup for function. But you can do it manually by following the [standard procedure](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) of securing ingress. There is also [general guideline](https://docs.bitnami.com/kubernetes/how-to/secure-kubernetes-services-with-ingress-tls-letsencrypt/) to enable TLS for your Kubernetes service using LetsEncrypt and kube-lego written by Bitnami folks. When you have Kube-lego setup, you can deploy function and create route with flag `--enableTLSAcme` enabled as below:

```
$ kubeless route create route1 --function get-python --enableTLSAcme
```

Running above command, Kubeless will automatically create ingress object with annotation `kubernetes.io/tls-acme: 'true'` set.