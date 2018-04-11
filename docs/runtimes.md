# Runtimes support

Right now Kubeless has support for the following runtimes:

 - Python: For the branch 2.7.X
 - NodeJS: For the branches 6.X and 8.X
 - Ruby: For the branch 2.4.X
 - PHP: For the branch 7.2.X

Each runtime is encapsulated in a container image. The reference to these images are injected in the Kubeless controller. You can find source code of all runtimes in `docker/runtime`.

Runtimes have a maximum timeout set by the environment variable FUNC_TIMEOUT. This environment variable can be set using the CLI option `--timeout`. The default value is 180 seconds. If a function takes more than that in being executed, the process will be terminated.

## Runtimes interface

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
                                                # (specific properties will depend on the function language)
context:
    function-name: "pubsub-nodejs"
    timeout: "180"
    runtime: "nodejs6"
    memory-limit: "128M"
```

You can check basic examples of every language supported in the [examples](https://github.com/kubeless/kubeless/tree/master/examples) folder.

### Runtime user

As a [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) functions are configured to run with an unprivileged user (UID 1000) by default. This prevent functions from having root privileges and ensure compatibility with OpenShift clusters. This default behaviour can be overriden specifying a different Security Context in the `Deployment` template that is part of the Function Spec.

## Configuring Default Runtime Container Images

The Kubeless controller defines a set of default container images per supported runtime variant.

These default container images can be configured via Kubernetes environment variables on the Kubeless controller's deployment container. Or modifying the `kubeless-config` ConfigMap that is deployed along with the Kubeless controller.

If you want to obtain the list of possible runtimes execute:

```console
$ kubeless get-server-config
INFO[0000] Current Server Config:
INFO[0000] Supported Runtimes are: python2.7, python3.4, python3.6, nodejs6, nodejs8, ruby2.4, dotnetcore2.0
```

If you want to deploy a custom runtime using an environment variable these are some examples:

| Runtime Variant | Environment Variable Name |
| --- | --- |
| NodeJS 6.X HTTP Trigger | `NODEJS6_RUNTIME` |
| NodeJS 8.X Event Trigger | `NODEJS8_PUBSUB_RUNTIME` |
| Python 2.7.x HTTP Trigger | `PYTHON2.7_RUNTIME` |
| Ruby 2.4.x HTTP Trigger | `RUBY2.4_RUNTIME` |
| ... | ... |

# Runtime variants

Every runtime use a web framework that is in charge of receiving requests and redirect them to the function.

### Node.js Trigger

For the Node.js runtime we start an [Express](http://expressjs.com) server and we include the routes for serving the health check and exposing the monitoring metrics. Apart from that we enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) requests and [Morgan](https://github.com/expressjs/morgan) for handling the logging in the server. Monitoring is supported if the function is synchronous or if it uses promises.

When using the Node.js runtime, it is possible to configure a [custom registry or scope](https://docs.npmjs.com/misc/scope#associating-a-scope-with-a-registry) in case a function needs to install modules from a different source. For doing so it is necessary to set up the environment variables *NPM_REGISTRY* and *NPM_SCOPE* when deploying the function:

```console
$ kubeless function deploy myFunction --runtime nodejs6 \
                                --env NPM_REGISTRY=http://my-registry.com \
                                --env NPM_SCOPE=@myorg \
                                --dependencies package.json \
                                --handler test.foobar \
                                --from-file test.js
```

### Python HTTP Trigger

For python we use [Bottle](https://bottlepy.org) and we also add routes for health check and monitoring metrics.

### Ruby HTTP Trigger

For the case of Ruby we use [Sinatra](http://www.sinatrarb.com) as web framework and we add the routes required for the function and the health check. Monitoring is currently not supported yet for this framework. PR is welcome :-)

### Go HTTP Trigger

The Go HTTP server doesn't include any framework since the native packages includes enough functionality to fit our needs. Since there is not a standard package that manages server logs that functionality is implemented in the same server. It is also required to implement the `ResponseWritter` interface in order to retrieve the Status Code of the response.

### Debugging compilation

If there is an error during the compilation of a function, the error message will be dumped to the termination log. If you see that the pod is crashed in a init container:

```
NAME                      READY     STATUS                  RESTARTS   AGE
get-go-6774465f95-x55lw   0/1       Init:CrashLoopBackOff   1          1m
```

That can mean that the compilation failed. You can obtain the compilation logs executing:

```console
$ kubectl get pod -l function=get-go -o yaml
...
    - containerID: docker://253fb677da4c3106780d8be225eeb5abf934a961af0d64168afe98159e0338c0
      image: andresmgot/go-init:1.10
      lastState:
        terminated:
          containerID: docker://253fb677da4c3106780d8be225eeb5abf934a961af0d64168afe98159e0338c0
          exitCode: 2
          finishedAt: 2018-04-06T09:01:16Z
          message: |
            # kubeless
            /go/src/kubeless/handler.go:6:1: syntax error: non-declaration statement outside function body
...
```

You can see there that there is a syntax error in the line 6 of the function. You can also retrieve the same information with this one-liner:

```console
$ kubectl get pod -l function=get-go -o go-template="{{range .items}}{{range .status.initContainerStatuses}}{{.lastState.terminated.message}}{{end}}{{end}}"

<no value># kubeless
/go/src/kubeless/handler.go:6:1: syntax error: non-declaration statement outside function body
```

### Timeout handling

One peculiarity of the Go runtime is that the user has a `Context` object as part of the `Event.Extensions` parameter. This can be used to handle timeouts in the function. For example:

```go
func Foo(event functions.Event, context functions.Context) (string, error) {
	select {
	case <-event.Extensions.Context.Done():
		return "", nil
  case <-time.After(5 * time.Second):
	}
	return "Function returned after 5 seconds", nil
}
```

If the function above has a timeout smaller than 5 seconds it will exit and the code after the `select{}` won't be executed. 

# Scheduled Trigger

This is meant for functions that should be triggered following a certain schedule. For specifying the execution frequency  we use the [Cron](https://en.wikipedia.org/wiki/Cron) format. Every time a scheduled function is executed, a [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) is started. This Job will do a HTTP GET request to the function service and will be successful as far as the function returns 200 OK.

For executing scheduled functions we use Kubernetes [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) using mostly the default options which means:
 - If a Job fails, it won't be restarted but it will be retried in the next scheduled event. The maximum time that a Job will exist is specified with the function timeout (180 seconds by default).
 - The concurrency policy is set to `Allow` so concurrent jobs may exists.
 - The history limit is set to maintain as maximum three successful jobs (and one failed).

If for some reason you want to modify one of the default values for a certain function you can execute `kubectl edit cronjob trigger-<func_name>` (where `func_name` is the name of your function) and modify the fields required. Once it is saved the CronJob will be updated.

# Monitoring functions

Some Kubeless runtimes expose metrics at `/metrics` endpoint and these metrics will be collected by Prometheus. We also include a prometheus setup in [`manifests/monitoring`](https://github.com/kubeless/kubeless/blob/master/manifests/monitoring/prometheus.yaml) to help you easier set it up. The metrics collected are: Number of calls, succeeded and error executions and the time spent per call.

# Custom Runtime (Alpha)

> NOTE: This feature is under heavy development and may change in the future

We are providing a way to define custom runtime in form of a container image. This way you are able to use any language or any binary with Kubeless as far as the image satisfies the following conditions:
 - It runs a web server listening in the port 8080
 - It exposes the endpoint `/healthz` to perform the container [liveness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
 - (Optional) It serves [Prometheus](https://prometheus.io) metrics in the endpoint `/metrics`

To deploy the container image you just need to specify it using the Kubeless CLI:

```console
$ kubeless function deploy --runtime-image bitnami/tomcat:9.0 webserver
$ kubeless function ls
NAME     	NAMESPACE	HANDLER         	RUNTIME  	TYPE	TOPIC
webserver	default  	                	         	HTTP
```

Now you can call your function like any other:

```console
$ kubeless function call webserver
...
<h2>If you're seeing this, you've successfully installed Tomcat. Congratulations!</h2>
```

Note that you can also use your own image and inject different functions. That means you have to manage how your runtime starts and looks for the injected function and executes it. Kubeless injects the function into the runtime container via a [Kubernetes ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/) object mounted at `/kubeless` folder, so make sure your runtime looks for function at that folder. Let's see an example:

First we need to create a base image. For this example we will use the Python web server that you can find [in the runtimes folder](../docker/runtime/python-2.7/http-trigger/kubeless.py). We will use the following Dockerfile:

```dockerfile
FROM python:2.7-slim

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client lxml

ADD kubeless.py /

EXPOSE 8080
CMD ["python", "/kubeless.py"]
```

Once you have built the image you need to push it to a registry to make it available within your cluster. Finally you can call the `deploy` command specifying the custom runtime image and the function you want to inject:

```console
$ kubeless function deploy \
  --runtime-image tuna/kubeless-python:0.0.6 \
  --from-file ./handler.py \
  --handler handler.hello \
  --runtime python2.7 \
  hello
$ kubeless function ls
NAME      	NAMESPACE	HANDLER     	RUNTIME  	TYPE	TOPIC
get-python	default  	foo.foo	      python2.7	HTTP
$ kubeless function call get-python
Connecting to function...
Forwarding from 127.0.0.1:30000 -> 8080
Forwarding from [::1]:30000 -> 8080
Handling connection for 30000
hello world
```

Note that it is possible to specify `--dependencies` as well when using custom images and install them using an Init container but that is only possible for the supported runtimes. You can get the list of supported runtimes executing `kubeless get-server-config`.

When using a runtime not supported your function will be stored as `/kubeless/function` without extension. For example, injecting a file `my-function.jar` would result in the file being mounted as `/kubeless/my-fuction`).
