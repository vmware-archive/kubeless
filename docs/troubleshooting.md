# Troubleshooting

## Installation

If installing using

```console
kubectl create -f kubeless.yaml --namespace kubeless
```

gives the following error:

```console
customresourcedefinition "functions.k8s.io" created error: error validating
"kubeless.yaml": error validating data: unknown object type 
schema.GroupVersionKind{Group:"", Version:"v1", Kind:"Service"}; if you
choose to ignore these errors, turn validation off with --validate=false
```

You probably have an older version of Kubernetes. Make sure
you are using at least version `1.7`.

## Kafka and Zookeeper Persistent Volume creation

Since Kubeless 0.5, there is a standalone manifest for deploying Kafka and Zookeeper. In some platforms, the Persistent Volumes that these applications require are not automatically generated. If that is your case you will see the deployments and Persistent Volume Claims as Pending:

```
$ kubectl get pods -n kubeless
NAME                                           READY     STATUS    RESTARTS   AGE
kafka-0                                        1/1       Pending   0          1h
kafka-trigger-controller-7f4f458f8b-l6f5m      1/1       Running   0          1h
kubeless-controller-manager-58d78fff74-g7fsd   1/1       Running   0          1h
zoo-0                                          1/1       Pending   0          1h
$ kubectl get pvc -n kubeless
NAME              STATUS    VOLUME    CAPACITY   ACCESSMODES   STORAGECLASS   AGE
datadir-kafka-0   Pending                                                     2m
zookeeper-zoo-0   Pending                                                     2m

```

If you are running Kubernetes in GKE check the specific guide [here](/docs/GKE-deployment) to create the required disks and PVs. In other case, check the provider documentation of how to create these required volumes. Note that `kafka` and `zookeeper` are only needed when working with Kafka events, you can still use Kubeless to trigger functions using HTTP requests.
