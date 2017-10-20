# Autoscaling with custom metrics on k8s 1.7

This walkthrough will go over the step-by-step of setting up the prometheus-based custom API server on your cluster and configuring autoscaler (HPA) to use application metrics sourced from prometheus instance.

This walkthrough is done in [kubeadm-dind-cluster v1.7](https://github.com/Mirantis/kubeadm-dind-cluster)

## Cluster configuration

Before getting started, ensure that the main components of your cluster are configured for autoscaling on custom metrics. As of Kubernetes 1.7, this requires enabling the aggregation layer on the API server and configuring the controller manager to use the metrics APIs via their REST clients.

Read more about the aggregation and autoscaling in the Kubernetes documentations:
- aggregation layer in v1.7: https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/

**Start the cluster**
```
$ wget https://cdn.rawgit.com/Mirantis/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.7.sh
$ chmod +x dind-cluster-v1.7.sh
$ ./dind-cluster-v1.7.sh up
```

Checking the state of the cluster:
```
$ docker ps
$ kubectl cluster-info
```

**Configuration**

The manifests of kubernetes components in `kubeadm-dind-cluster` locate at `/etc/kubernetes/manifests`, you can just jump in the master "container" and edit them directly; and kubeadm manages to recreate the components immediately.

These below configurations must be set:

1) enable and configure aggregation layer in v1.7 in `kube-apiserver`: https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/
```
--requestheader-client-ca-file=<path to aggregator CA cert>
--requestheader-allowed-names=aggregator
--requestheader-extra-headers-prefix=X-Remote-Extra-
--requestheader-group-headers=X-Remote-Group
--requestheader-username-headers=X-Remote-User
--proxy-client-cert-file=<path to aggregator proxy cert>
--proxy-client-key-file=<path to aggregator proxy key>
```

2) configure controller manager to use the metrics APIs via their REST clients by these settings in `kube-controller-manager`:
```
--horizontal-pod-autoscaler-use-rest-clients=true
--horizontal-pod-autoscaler-sync-period=10s
--master=<apiserver-address>:<port> //port should be 8080
```

The `horizontal-pod-autoscaler-sync-period` parameter set the interval time (in second) that the HPA controller synchronizes the number of pods. By default it's 30s. Sometimes we might want to optimize this parameter to make the HPA controller reacts faster.

3) the autoscaling for custom metrics is supported in HPA since v1.7 via `autoscaling/v2alpha1` API. It needs to be enabled by setting the `runtime-config` in `kube-apiserver`:
```
--runtime-config=api/all=true
```
Once the kube-apiserver is configured and up and running, `kubectl` will auto-discover all API groups. Check it using this below command and you will see the `autoscaling/v2alpha1` is enabled:
```
$ kubectl api-versions
admissionregistration.k8s.io/v1alpha1
apiextensions.k8s.io/v1beta1
apiregistration.k8s.io/v1beta1
apps/v1beta1
authentication.k8s.io/v1
authentication.k8s.io/v1beta1
authorization.k8s.io/v1
authorization.k8s.io/v1beta1
autoscaling/v1
**autoscaling/v2alpha1**
batch/v1
batch/v2alpha1
certificates.k8s.io/v1beta1
extensions/v1beta1
networking.k8s.io/v1
policy/v1beta1
rbac.authorization.k8s.io/v1alpha1
rbac.authorization.k8s.io/v1beta1
settings.k8s.io/v1alpha1
storage.k8s.io/v1
storage.k8s.io/v1beta1
v1
```

## Deploy Prometheus to monitor services
The Prometheus setup contains a Prometheus operator and a Prometheus instance

```
$ kubectl create -f prometheus-operator.yaml
clusterrole "prometheus-operator" created
serviceaccount "prometheus-operator" created
clusterrolebinding "prometheus-operator" created
deployment "prometheus-operator" created

$ kubectl create -f sample-prometheus-instance.yaml
clusterrole "prometheus" created
serviceaccount "prometheus" created
clusterrolebinding "prometheus" created
prometheus "sample-metrics-prom" created
service "sample-metrics-prom" created

$ kubectl get svc
NAME                  CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
kubernetes            10.96.0.1       <none>        443/TCP          6d
prometheus-operated   None            <none>        9090/TCP         1h
```

## Deploy a custom API server
When the aggregator enabled and configured properly, one can deploy and register a custom API server that provides the `custom-metrics.metrics.k8s.io/v1alpha1` API group/version and let the HPA controller queries custom metrics from that.

The custom API server we are using here is basically [a Prometheus adapter](https://github.com/directxman12/k8s-prometheus-adapter) which can collect metrics from Prometheus and send to HPA controller via REST queries (that's why we must configure HPA controller to use REST client via the `--horizontal-pod-autoscaler-use-rest-clients` flag)

```
$ kubectl create -f custom-metrics.yaml
namespace "custom-metrics" created
serviceaccount "custom-metrics-apiserver" created
clusterrolebinding "custom-metrics:system:auth-delegator" created
rolebinding "custom-metrics-auth-reader" created
clusterrole "custom-metrics-read" created
clusterrolebinding "custom-metrics-read" created
deployment "custom-metrics-apiserver" created
service "api" created
apiservice "v1alpha1.custom-metrics.metrics.k8s.io" created
clusterrole "custom-metrics-server-resources" created
clusterrolebinding "hpa-controller-custom-metrics" created
```

At this step, the custom API server is deployed and registered to API aggregator, so we can see it:

```
$ kubectl api-versions
admissionregistration.k8s.io/v1alpha1
apiextensions.k8s.io/v1beta1
apiregistration.k8s.io/v1beta1
apps/v1beta1
authentication.k8s.io/v1
authentication.k8s.io/v1beta1
authorization.k8s.io/v1
authorization.k8s.io/v1beta1
autoscaling/v1
autoscaling/v2alpha1
batch/v1
batch/v2alpha1
certificates.k8s.io/v1beta1
**custom-metrics.metrics.k8s.io/v1alpha1**
extensions/v1beta1
monitoring.coreos.com/v1alpha1
networking.k8s.io/v1
policy/v1beta1
rbac.authorization.k8s.io/v1alpha1
rbac.authorization.k8s.io/v1beta1
settings.k8s.io/v1alpha1
storage.k8s.io/v1
storage.k8s.io/v1beta1
v1

$ kubectl get po -n custom-metrics
NAME                                        READY     STATUS    RESTARTS   AGE
custom-metrics-apiserver-2956926076-wcgmw   1/1       Running   0          1h

$ kubectl get --raw /apis/custom-metrics.metrics.k8s.io/v1alpha1
{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"custom-metrics.metrics.k8s.io/v1alpha1","resources":[]}
```

## Deploy a sample app

Now we can deploy a sample app and sample HPA rule to do the autoscale with `http_requests` metric collected and exposed via Prometheus.

```
$ cat sample-metrics-app.yaml
...
---
kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2alpha1
metadata:
  name: sample-metrics-app-hpa
spec:
  scaleTargetRef:
    kind: Deployment
    name: sample-metrics-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Object
    object:
      target:
        kind: Service
        name: sample-metrics-app
      metricName: http_requests
      targetValue: 100

$ kubectl create -f sample-metrics-app.yaml
deployment "sample-metrics-app" created
service "sample-metrics-app" created
servicemonitor "sample-metrics-app" created
horizontalpodautoscaler "sample-metrics-app-hpa" created

$ kubectl get hpa
NAME                     REFERENCE                       TARGETS      MINPODS   MAXPODS   REPLICAS   AGE
sample-metrics-app-hpa   Deployment/sample-metrics-app   866m / 100   2         10        2          1h
```

Try to increase some loads by hitting the sample app service, then you can see the HPA scales it up.


## Further reading

https://github.com/kubernetes/community/blob/master/contributors/design-proposals/custom-metrics-api.md
https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics
