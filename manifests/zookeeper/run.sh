kubectl create -f namespace.yaml
kubectl create -f zk.headless.svc.yaml
kubectl create -f zk.svc.yaml
kubectl create -f zk-stateful.yaml


