# What is Kafka?

Apache Kafka™ is a distributed streaming platform. What exactly does that mean?

1. It lets you publish and subscribe to streams of records. In this respect it is similar to a message queue or enterprise messaging system.
2. It lets you store streams of records in a fault-tolerant way.
3. It lets you process streams of records as they occur.

# What is Kafka good for?

It gets used for two broad classes of application:

1. Building real-time streaming data pipelines that reliably get data between systems or applications
2. Building real-time streaming applications that transform or react to the streams of data


# How is Kafka user for in Kubeless?

Kafka is used to provide a rich messaging framework to connect heterogeneous systems. 

# How to install 

Just execute:

```
kubectl create -f namespace.yaml
kubectl create -f kafka-headless-svc.yaml
kubectl create -f kafka-svc.yaml
kubectl create -f kafka-deployment.yaml
```
