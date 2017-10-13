#!/usr/bin/env bash

kubectl get po --all-namespaces
kubectl delete statefulsets --namespace kubeless kafka
kubectl delete statefulsets --namespace kubeless zoo
kubectl delete deployment --namespace kubeless kubeless-controller
kubectl delete svc --namespace kubeless broker
kubectl delete svc --namespace kubeless kafka
kubectl delete svc --namespace kubeless zoo
kubectl delete svc --namespace kubeless zookeeper
#kubectl delete customresourcedefinition function.k8s.io
