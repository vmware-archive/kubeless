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

After that,Please check the specific guide `here` to create appropriate disks and PVs.If you are running Kubernetes in
GKE,you can provision those persistent volumes mannually deploying the manifests present in the misc folder.If you use
other cloud provider,check [kubernetes docs](https://kubernetes.io/docs/concepts/storage/volumes/) to create these required volumes.
