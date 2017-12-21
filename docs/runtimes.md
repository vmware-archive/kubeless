# Runtimes support

Right now Kubeless has support for the following runtimes:
 - Python: For the branch 2.7.X
 - NodeJS: For the branches 6.X and 8.X
 - Ruby: For the branch 2.4.X

Each runtime is encapsulated in a container image. The reference to these images are injected in the Kubeless controller. You can find source code of all runtimes in `docker/runtime`.

Runtimes have a maximum timeout set by the environment variable FUNC_TIMEOUT. This environment variable can be set using the CLI option `--timeout`. The default value is 180 seconds. If a function takes more than that in being executed, the process will be terminated.

## Configuring Default Runtime Container Images
The Kubeless controller defines a set of default container images per supported runtime variant.
These default container images can be configured via Kubernetes environment variables on the Kubeless controller's deployment container.
Some examples:

| Runtime Variant | Environment Variable Name |
| --- | --- |
| NodeJS 6.X HTTP Trigger | NODEJS6_RUNTIME |
| NodeJS 8.X Event Trigger | NODEJS8_PUBSUB_RUNTIME |
| Python 2.7.x HTTP Trigger | PYTHON2.7_RUNTIME |
| Ruby 2.4.x HTTP Trigger | RUBY2.4_RUNTIME |
| ... | ... |

# Runtime variants
## HTTP Trigger
This variant is used when the function is meant to be triggered through HTTP. For doing so we use a web framework that is in charge of receiving request and redirect them to the function. This kind of trigger is supported for all the runtimes.

### Node.js HTTP Trigger
For the Node.js runtime we start an [Express](http://expressjs.com) server and we include the routes for serving the health check and exposing the monitoring metrics. Apart from that we enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) requests and [Morgan](https://github.com/expressjs/morgan) for handling the logging in the server. Monitoring is supported if the function is synchronous or if it uses promises.

When using the Node.js runtime, it is possible to configure a [custom registry or scope](https://docs.npmjs.com/misc/scope#associating-a-scope-with-a-registry) in case a function needs to install modules from a different source. For doing so it is necessary to set up the environment variables *NPM_REGISTRY* and *NPM_SCOPE* when deploying the function:
```console
$ kubeless function deploy myFunction --runtime nodejs6 \
                                --env NPM_REGISTRY=http://my-registry.com \
                                --env NPM_SCOPE=@myorg \
                                --dependencies package.json \
                                --handler test.foobar \
                                --from-file test.js \
                                --trigger-http
```

### Python HTTP Trigger
For python we use [Bottle](https://bottlepy.org) and we also add routes for health check and monitoring metrics.

### Ruby HTTP Trigger
For the case of Ruby we use [Sinatra](http://www.sinatrarb.com) as web framework and we add the routes required for the function and the health check. Monitoring is currently not supported yet for this framework. PR is welcome :-)

## Event trigger
This variant is used when the function is meant to be triggered through message events in a pre-deployed [Kafka](https://kafka.apache.org) system. We include a set of kafka/zookeeper in the deployment manifest of [Kubeless release package](https://github.com/kubeless/kubeless/releases) that will be deployed together with the Kubeless controller. Basically the runtimes are Kafka consumers which listen messages in a specific kafka topic and execute the injected function.

Right now the runtimes that support this kind of events are Python, NodeJS and Ruby.

The pods are deployed using the function handler as [group ID](https://kafka.apache.org/documentation/#intro_consumers). That means that the load will be balanced. If a function is scaled and its deployment has more than one replica only one pod will process a published message.

## Scheduled Trigger

This is meant for functions that should be triggered following a certain schedule. For specifying the execution frequency  we use the [Cron](https://en.wikipedia.org/wiki/Cron) format. Every time a scheduled function is executed, a [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) is started. This Job will do a HTTP GET request to the function service and will be successful as far as the function returns 200 OK.

For executing scheduled functions we use Kubernetes [CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) using mostly the default options which means:
 - If a Job fails, it won't be restarted but it will be retried in the next scheduled event. The maximum time that a Job will exist is specified with the function timeout (180 seconds by default).
 - The concurrency policy is set to `Allow` so concurrent jobs may exists.
 - The history limit is set to maintain as maximum three successful jobs (and one failed).

If for some reason you want to modify one of the default values for a certain function you can execute `kubectl edit cronjob trigger-<func_name>` (where `func_name` is the name of your function) and modify the fields required. Once it is saved the CronJob will be updated.

## Monitoring functions
Kubeless runtimes are exposing metrics at `/metrics` endpoint and these metrics will be collected by Prometheus. We also include a prometheus setup in [`manifests/monitoring`](https://github.com/kubeless/kubeless/blob/master/manifests/monitoring/prometheus.yaml) to help you easier set it up. The metrics collected are: Number of calls, succeeded and error executions and the time spent per call.

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
$ kubeless function deploy --runtime-image tuna/kubeless-python:0.0.6 --from-file ./handler.py --handler handler.hello --runtime python2.7 --trigger-http hello
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

Note that it is possible to specify `--dependencies` as well when using custom images and install them using an Init container but that is only possible for the supported runtimes. You can get the list of supported runtimes executing `kubeless function deploy --help`.

When using a runtime not supported your function will be stored as `/kubeless/function` without extension. For example, injecting a file `my-function.jar` would result in the file being mounted as `/kubeless/my-fuction`).
