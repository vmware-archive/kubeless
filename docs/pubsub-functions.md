# PubSub events

You can trigger any Kubeless function by a PubSub mechanism. The PubSub function is expected to consume input messages from a predefined topic from a messaging system. Kubeless currently supports using events from Kafka and NATS messaging systems.

## Kafka

In Kafka [release page](https://github.com/kubeless/kafka-trigger/releases), you can find the manifest to quickly deploy a collection of Kafka and Zookeeper statefulsets. If you have a Kafka cluster already running in the same Kubernetes environment, you can also deploy PubSub function with it. Check out [this tutorial](/docs/use-existing-kafka) for more details how to do that.

If you want to deploy the manifest we provide to deploy Kafka and Zookeeper execute the following command:

```console
$ export RELEASE=$(curl -s https://api.github.com/repos/kubeless/kafka-trigger/releases/latest | grep tag_name | cut -d '"' -f 4)
$ kubectl create -f https://github.com/kubeless/kafka-trigger/releases/download/$RELEASE/kafka-zookeeper-$RELEASE.yaml
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

A function can be as simple as:

```python
def foobar(event, context):
  print event['data']
  return event['data']
```

Now you can deploy a pubsub function. 

```console
$ kubeless function deploy test --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py
```

You need to create a _Kafka_ trigger that lets you associate a function with a topic specified by `--trigger-topic` as below:

```console
$ kubeless trigger kafka create test --function-selector created-by=kubeless,function=test --trigger-topic test-topic
```

After that you can invoke the function by publishing messages in that topic. To allow you to easily manage topics `kubeless` provides a convenience function `kubeless topic`. You can create/delete and publish to a topic easily.

```console
$ kubeless topic create test-topic
$ kubeless topic publish --topic test-topic --data "Hello World!"
```

You can check the result in the pod logs:

```console
$ kubectl logs test-695251588-cxwmc
...
Hello World!
```
## NATS

If you do not have NATS cluster its pretty easy to setup a NATS cluster. Run below command to deploy a [NATS operator](https://github.com/nats-io/nats-operator)

```console
$ kubectl apply -f https://raw.githubusercontent.com/nats-io/nats-operator/master/example/deployment-rbac.yaml
```

Once NATS operator is up and running run below command to deploy a NATS cluster

```console
echo '
apiVersion: "nats.io/v1alpha2"
kind: "NatsCluster"
metadata:
  name: "nats"
spec:
  size: 3
  version: "1.1.0"
' | kubectl apply -f - -n nats-io
```

Above command will create NATS cluster IP service `nats.nats-io.svc.cluster.local:4222` which is the default URL Kubeless NATS trigger contoller expects.

Now use this manifest to deploy Kubeless NATS triggers controller.

```console
$ export RELEASE=$(curl -s https://api.github.com/repos/kubeless/nats-trigger/releases/latest | grep tag_name | cut -d '"' -f 4)
$ kubectl create -f https://github.com/kubeless/nats-trigger/releases/download/$RELEASE/nats-$RELEASE.yaml
```

By default NATS trigger controller expects NATS cluster is available as Kubernetes cluster service `nats.nats-io.svc.cluster.local:4222`. You can overide the default NATS cluster url used by setting the environment variable `NATS_URL` in the manifest. Once NATS trigger controller is setup you can deploy the function and associate function with a topic on the NATS cluster.

```console
$ kubeless function deploy pubsub-python-nats --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py
```

After function is deployed you can use `kubeless trigger nats` CLI command to  associate function with a topic on NATS cluster as below.

```console
$ kubeless trigger nats create pubsub-python-nats --function-selector created-by=kubeless,function=pubsub-python-nats --trigger-topic test
```

At this point you are all set try Kubeless NATS triggers.

You could quickly test the functionality by publishing a message to the topic, and verifying that message is seen by the pod running the function.

```console
$ kubeless trigger nats publish --url nats://nats-server-ip:4222 --topic test --message "Hello World!"
```

You can check the result in the pod logs:

```console
$ kubectl logs pubsub-python-nats-5b9c849fc-tvq2l
...
Hello World!
```

## Other commands

You can create, list and delete PubSub topics (for Kafka):

```console
$ kubeless topic create another-topic
Created topic "another-topic".

$ kubeless topic delete another-topic

$ kubeless topic ls
```
