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

### Further reading

[Custom Metrics API](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/custom-metrics-api.md)

[Support for custom metrics](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
