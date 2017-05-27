# What is Zoookeeper?

ZooKeeper is a centralized service for maintaining configuration information, naming, providing distributed synchronization, and providing group services.

# What is Zoookeeper good for?

Apache ZooKeeper is an effort to develop and maintain an open-source server which enables highly reliable distributed coordination.

# How is Zookeeper used in Kubeless?

Zookeeper is used to provide the basic fundations to run and operate the Kafka distributed messaging system. 

# How to install

Just execute : 

```
kubectl create -f namespace.yaml
kubectl create -f zk-headless-svc.yaml
kubectl create -f zk-svc.yaml
kubectl create -f zk-stateful.yaml
```

