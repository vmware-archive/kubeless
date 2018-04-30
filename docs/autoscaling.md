# Autoscaling function deployment in Kubeless

This document gives you an overview of how we do autoscaling for functions in Kubeless and also give you a walkthrough how to configure it for custom metric.

## Overview

Kubernetes introduces [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for pod autoscaling. In kubeless, each function is deployed into a separate Kubernetes deployment, so naturally we leverage HPA to automatically scale function based on defined workload metrics.

If you're on Kubeless CLI, this below command gives you an idea how to setup autoscaling for deployed function:

```console
$ kubeless autoscale --help
autoscale command allows user to list, create, delete autoscale rule
for function on Kubeless

Usage:
  kubeless autoscale SUBCOMMAND [flags]
  kubeless autoscale [command]

Available Commands:
  create      automatically scale function based on monitored metrics
  delete      delete an autoscale from Kubeless
  list        list all autoscales in Kubeless

Flags:
  -h, --help   help for autoscale

Use "kubeless autoscale [command] --help" for more information about a command.
```

Once you create an autoscaling rule for a specific function (with `kubeless autoscale create`), the corresponding HPA object will be added to the system which is going to monitor your function and auto-scale its pods based on the autoscaling rule you define in the command. The default metric is CPU, but you have option to do autoscaling with custom metrics. At this moment, Kubeless supports `qps` which stands for number of incoming requests to function per second.

```console
$ kubeless autoscale create --help
automatically scale function based on monitored metrics

Usage:
  kubeless autoscale create <name> FLAG [flags]

Flags:
  -h, --help               help for create
      --max int32          maximum number of replicas (default 1)
      --metric string      metric to use for calculating the autoscale. Supported
      metrics: cpu, qps (default "cpu")
      --min int32          minimum number of replicas (default 1)
  -n, --namespace string   Specify namespace for the autoscale
      --value string       value of the average of the metric across all replicas.
      If metric is cpu, value is a number represented as percentage. If metric
      is qps, value must be in format of Quantity
```

The below part will walk you though setup need to be done in order to make function auto-scaled based on `qps` metric.

## Autoscaling based on CPU usage

To autoscale based on CPU usage, it is *required* that your function has been deployed with CPU request limits.

To do this, use the `--cpu` parameter when deploying your function. Please see the [Meaning of CPU](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) for the format of the value that should be passed. 

## Autoscaling with custom metrics

It is possible to use custom metrics (like queries per second) to scale your functions. We are [looking for help](https://github.com/kubeless/kubeless/issues/647) in order to document the required steps to do so with the different Kubernetes providers for newer versions of Kubernetes. If you want to contribute to this guide PRs are more than welcome :).

**Warning** This walkthrough is done in [kubeadm-dind-cluster v1.7](https://github.com/Mirantis/kubeadm-dind-cluster) it may not work for other versions or platforms.

### Cluster configuration

Before getting started, ensure that the main components of your cluster are configured for autoscaling on custom metrics. As of Kubernetes 1.7, this requires enabling the aggregation layer on the API server and configuring the controller manager to use the metrics APIs via their REST clients.

Read more about the aggregation and autoscaling in the Kubernetes documentations:

- [Aggregation layer in v1.7](https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/)

#### Start the cluster

```console
wget https://cdn.rawgit.com/Mirantis/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.7.sh
chmod +x dind-cluster-v1.7.sh
./dind-cluster-v1.7.sh up
```

Checking the state of the cluster:

```console
docker ps
kubectl cluster-info
```

#### Configuration

The manifests of kubernetes components in `kubeadm-dind-cluster` locate at `/etc/kubernetes/manifests`, you can just jump in the master "container" and edit them directly; and kubeadm manages to recreate the components immediately.

These below configurations must be set:

- Enable and configure [aggregation layer](https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/) in v1.7 in `kube-apiserver`:

```
--requestheader-client-ca-file=<path to aggregator CA cert>
--requestheader-allowed-names=aggregator
--requestheader-extra-headers-prefix=X-Remote-Extra-
--requestheader-group-headers=X-Remote-Group
--requestheader-username-headers=X-Remote-User
--proxy-client-cert-file=<path to aggregator proxy cert>
--proxy-client-key-file=<path to aggregator proxy key>
```

- Configure controller manager to use the metrics APIs via their REST clients by these settings in `kube-controller-manager`:

```
--horizontal-pod-autoscaler-use-rest-clients=true
--horizontal-pod-autoscaler-sync-period=10s
--master=<apiserver-address>:<port> //port should be 8080
```

The `horizontal-pod-autoscaler-sync-period` parameter set the interval time (in second) that the HPA controller synchronizes the number of pods. By default it's 30s. Sometimes we might want to optimize this parameter to make the HPA controller reacts faster.

- The autoscaling for custom metrics is supported in HPA since v1.7 via `autoscaling/v2alpha1` API. It needs to be enabled by setting the `runtime-config` in `kube-apiserver`:

```
--runtime-config=api/all=true
```

Once the kube-apiserver is configured and up and running, `kubectl` will auto-discover all API groups. Check it using this below command and you will see the `autoscaling/v2alpha1` is enabled:

```console
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

### Deploy Prometheus to monitor services

The Prometheus setup contains a Prometheus operator and a Prometheus instance

```console
$ kubectl create -f $KUBELESS_REPO/manifests/autoscaling/prometheus-operator.yaml
clusterrole "prometheus-operator" created
serviceaccount "prometheus-operator" created
clusterrolebinding "prometheus-operator" created
deployment "prometheus-operator" created

$ kubectl create -f $KUBELESS_REPO/manifests/autoscaling/sample-prometheus-instance.yaml
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

### Deploy a custom API server

When the aggregator enabled and configured properly, one can deploy and register a custom API server that provides the `custom-metrics.metrics.k8s.io/v1alpha1` API group/version and let the HPA controller queries custom metrics from that.

The custom API server we are using here is basically [a Prometheus adapter](https://github.com/directxman12/k8s-prometheus-adapter) which can collect metrics from Prometheus and send to HPA controller via REST queries (that's why we must configure HPA controller to use REST client via the `--horizontal-pod-autoscaler-use-rest-clients` flag)

```console
$ kubectl create -f $KUBELESS_REPO/manifests/autoscaling/custom-metrics.yaml
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

```console
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
{"kind":"APIResourceList","apiVersion":"v1",
"groupVersion":"custom-metrics.metrics.k8s.io/v1alpha1","resources":[]}
```

### Deploy a sample app

Now we can deploy a sample app and sample HPA rule to do the autoscale with `http_requests` metric collected and exposed via Prometheus.

```console
$ cat $KUBELESS_REPO/manifests/autoscaling/sample-metrics-app.yaml
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

$ kubectl create -f $KUBELESS_REPO/manifests/autoscaling/sample-metrics-app.yaml
deployment "sample-metrics-app" created
service "sample-metrics-app" created
servicemonitor "sample-metrics-app" created
horizontalpodautoscaler "sample-metrics-app-hpa" created

$ kubectl get hpa
```

Try to increase some loads by hitting the sample app service, then you can see the HPA scales it up.

## Autoscaling on GKE

Let's say you are running Kubeless on GKE. At this moment you can only do autoscaling with default metric (CPU). For custom metrics, GKE team says that it will be supported from GKE 1.9+. So stay tuned.

### Further reading

[Custom Metrics API](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/custom-metrics-api.md)

[Support for custom metrics](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
