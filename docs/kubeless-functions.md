# Kubeless Functions

Functions are the main entity in Kubeless. It is possible to write Functions in different languages but all of them share common properties like the generic interface, the default timeout or the runtime UID. In this document we are going to explain some these common properties and different runtimes availables in Kubeless. You can find in depth details about the Function specification [here](/docs/advanced-function-deployment). 

## Functions Interface

Every function receives two arguments: `event` and `context`. The first argument contains information about the source of the event that the function has received. The second contains general information about the function like its name or maximum timeout. This is a representation in YAML of a Kafka event:

```yaml
event:                                  
  data:                                         # Event data
    foo: "bar"                                  # The data is parsed as JSON when required
  event-id: "2ebb072eb24264f55b3fff"            # Event ID
  event-type: "application/json"                # Event content type
  event-time: "2009-11-10 23:00:00 +0000 UTC"   # Timestamp of the event source
  event-namespace: "kafkatriggers.kubeless.io"  # Event emitter
  extensions:                                   # Optional parameters
    request: ...                                # Reference to the request received 
    response: ...                               # Reference to the response to send 
                                                # (specific properties will depend on the function language)
context:
    function-name: "pubsub-nodejs"
    timeout: "180"
    runtime: "nodejs6"
    memory-limit: "128M"
```

Functions should return a string that will be used as the HTTP response for the caller. Some runtimes may support different types (like objects) for the returned values.

You can check basic examples of every language supported in the [examples](https://github.com/kubeless/kubeless/tree/master/examples) folder.

## Functions Timeout

Runtimes have a maximum timeout set by the environment variable FUNC_TIMEOUT. This environment variable can be set using the CLI option `--timeout`. The default value is 180 seconds. If a function takes more than that in being executed, the process will be terminated.

## Runtime User

As a [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) functions are configured to run with an unprivileged user (UID 1000) by default (except for OpenShift where the UID is automatically set). This prevent functions from having root privileges. This default behaviour can be overridden specifying a different Security Context in the `Deployment` template that is part of the Function Spec.

## Scheduled functions

It is possible to deploy functions that should be triggered following a certain schedule. For specifying the execution frequency we use the [Cron](https://en.wikipedia.org/wiki/Cron) format. Every time a scheduled function is executed, a [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) is started. This Job will do a HTTP GET request to the function service and will be successful as far as the function returns 200 OK.

For executing scheduled functions we use Kubernetes [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) using mostly the default options which means:
 - If a Job fails, it won't be restarted but it will be retried in the next scheduled event. The maximum time that a Job will exist is specified with the function timeout (180 seconds by default).
 - The concurrency policy is set to `Allow` so concurrent jobs may exists.
 - The history limit is set to maintain as maximum three successful jobs (and one failed).

If for some reason you want to modify one of the default values for a certain function you can execute `kubectl edit cronjob trigger-<func_name>` (where `func_name` is the name of your function) and modify the fields required. Once it is saved the CronJob will be updated.

## Monitoring functions

Some Kubeless runtimes expose metrics at `/metrics` endpoint and these metrics will be collected by Prometheus. We also include a prometheus setup in [`manifests/monitoring`](https://github.com/kubeless/kubeless/blob/master/manifests/monitoring/prometheus.yaml) to help you easier set it up. The metrics collected are: Number of calls, succeeded and error executions and the time spent per call.

## Runtime variants

Check [this document](/docs/runtimes) to get more details about supported runtimes and languages.
