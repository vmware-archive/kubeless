# Deploy Kubeless with Helm

Create a `kubeless` namespace:

```console
$ kubectl create ns kubeless
```

Note that you could install kubeless in the default namespace or any other namespace.

Install kubeless with [helm](https://github.com/kubernetes/helm)

```console
helm init
helm install --name kubeless --namespace kubeless ./kubeless
```

