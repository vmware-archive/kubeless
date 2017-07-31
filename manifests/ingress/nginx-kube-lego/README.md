# Ingress with HTTPS enable

- Kube-lego for certificate retrieval from Let's Encrypt
- nginx ingress-controller

## Create nginx ingress-controller

```
kubectl apply -f nginx/ingress-controller.yaml
```

The nginx service uses NodePort to publish the service (for minikube only). Consider to change service type to LoadBalancer in public k8s cluster.

```
kubectl describe svc nginx --namespace kubeless
```

So you have to point your domains to the [nodeIP]:[nodePort]

## Create an example app (echoserver) and ingress-rule for it

```
kubectl apply -f echoserver/echoserver.yaml
```

- Make sure the echo service is reachable through http://echo.example.com

## Enable kube-lego

```
kubectl apply -f lego/kube-lego.yaml
```
- Change the email address in `lego/configmap` object before creating the
  kubernetes resource
- Please be aware that kube-lego creates it's related service on its own


## For non-tls for echoserver ingress

```
kubectl apply -f echoserver/ingress-notls.yaml
```

## Get debug information

- Look at the log output of the nginx pod
- Look at the log output of the ingress pods
- Sometimes after acquiring a new certificate nginx needs to be restarted (as
  it's not watching change events for secrets)
