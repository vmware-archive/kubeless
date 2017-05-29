kubectl create -f namespace.yaml
kubectl create -f kafka-headless-svc.yaml
kubectl create -f kafka-svc.yaml
kubectl create -f kafka-stateful.yaml
