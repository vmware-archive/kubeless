* Install Kubeless

Create `kubeless` namespace and create zookeeper

```
kubectl create -f zookeeper/namespace.yaml 
kubectl create -f zookeeper/zk-headless-svc.yaml
kubectl create -f zookeeper/zk-stateful.yaml 
kubectl create -f zookeeper/zk-svc.yaml
```

Create TPR For functions

```
kubectl create -f tpr/function.yaml
```

Create Kafka broker

```
kubectl create -f kafka/kafka-deployment.yaml
kubectl create -f kafka/kafka-headless-svc.yaml
kubectl create -f kafka/kafka-svc.yaml
```

Launch kubeless controller and UI

```
kubectl create -f controller/controller-deployment.yaml
kubectl create -f ui/ui-deployment.yaml
kubectl create -f ui/ui-svc.yaml 
```
