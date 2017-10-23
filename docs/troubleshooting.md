# Troubleshooting

## Installation

If installing using
```
kubectl create -f kubeless.yaml --namespace kubeless
```
gives the following error:
```
customresourcedefinition "functions.k8s.io" created
error: error validating "kubeless.yaml": error validating data: unknown object type schema.GroupVersionKind{Group:"", Version:"v1", Kind:"Service"}; if you choose to ignore these errors, turn validation off with --validate=false
```

You probably have an older version of Kubernetes. Make sure
you are using at least version `1.7`.
