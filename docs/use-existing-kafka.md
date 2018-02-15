# How to deploy Kubeless PubSub function with an existing Kafka cluster in Kubernetes

In Kubeless [release page](https://github.com/kubeless/kubeless/releases), we provide along with Kubeless manifests a collection of Kafka and Zookeeper statefulsets which helps user to quickly deploying PubSub function. These statefulsets are deployed in `kubeless` namespace. However, if you have a Kafka cluster already running in the same Kubernetes cluster, this doc will walk you through how to deploy Kubeless PubSub function with it.

Let's assume that you have Kafka cluster running at `pubsub` namespace like below:

```
$ kubectl -n pubsub get po
NAME      READY     STATUS    RESTARTS   AGE
kafka-0   1/1       Running   0          7h
zoo-0     1/1       Running   0          7h

$ kubectl -n pubsub get svc
NAME        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
kafka       ClusterIP   10.55.253.151   <none>        9092/TCP            7h
zookeeper   ClusterIP   10.55.248.146   <none>        2181/TCP            7h
```

And Kubeless already running at `kubeless` namespace:

```
$ kubectl -n kubeless get po
NAME                                   READY     STATUS    RESTARTS   AGE
kubeless-controller-58676964bb-l79gh   1/1       Running   0          5d
```

Kubeless provide several [PubSub runtimes](https://hub.docker.com/r/kubeless/),which has suffix `event-consumer`, specified for languages that help you to quickly deploy your function with PubSub mechanism. Those runtimes are configured to read Kafka configuration at two environment variables:

- KUBELESS_KAFKA_SVC: which points to kafka service name in Kubernetes cluster.
- KUBELESS_KAFKA_NAMESPACE: which declares the namespace that Kafka is running on.

In this example, when deploying function we will declare two environment variables `KUBELESS_KAFKA_SVC=kafka` and `KUBELESS_KAFKA_NAMESPACE=pubsub`.

We now try to deploy a provided function in `examples` folder with command as below:

```
$ kubeless function deploy pubsub-python --trigger-topic s3-python --runtime python2.7 --handler pubsub.handler --from-file examples/python/pubsub.py --env KUBELESS_KAFKA_SVC=kafka --env KUBELESS_KAFKA_NAMESPACE=pubsub
```

The `pubsub-python` function will just print out messages it receive from `s3-python` topic. Checking if the function is up and running:

```
$ kubectl get po
NAME                             READY     STATUS        RESTARTS   AGE
pubsub-python-5445bdcb64-48bv2   1/1       Running       0          4s
```

Now we need to create `s3-python` topic and try to publish some messages. You can do it on your own kafka client. In this example, I will try to use the bundled binaries in the kafka container:

```
# create s3-python topic
$ kubectl -n pubsub exec -it kafka-0 -- /opt/bitnami/kafka/bin/kafka-topics.sh --create --zookeeper zookeeper.pubsub:2181 --replication-factor 1 --partitions 1 --topic s3-python

# send test message to s3-python topic
$ kubectl -n pubsub exec -it kafka-0 -- /opt/bitnami/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic s3-python
> hello world
```

Open another terminal and check for the pubsub function log to see if it receives the message:

```
$ kubectl logs -f pubsub-python-5445bdcb64-48bv2
hello world
```
