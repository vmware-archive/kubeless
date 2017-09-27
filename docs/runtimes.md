# Runtimes support

Right now Kubeless has support for the following runtimes:
 - Python: For the branch 2.7.X
 - NodeJS: For the branches 6.X and 8.X
 - Ruby: For the branch 2.4.X

Each runtime is encapsulated in a container image. The reference to these images are injected in the Kubeless controller. You can find source code of all runtimes in `docker/runtime`.

# Runtime variants
## HTTP Trigger
This variant is used when the function is meant to be triggered through HTTP. For doing so we use a web framework that is in charge of receiving request and redirect them to the function. This kind of trigger is supported for all the runtimes.

### NodeJS HTTP Trigger
For the NodeJS runtime we start an [Express](http://expressjs.com) server and we include the routes for serving the health check and exposing the monitoring metrics. Apart from that we enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) requests and [Morgan](https://github.com/expressjs/morgan) for handling the logging in the server. Monitoring is supported if the function is synchronous or if it uses promises.

### Python HTTP Trigger
For python we use [Bottle](https://bottlepy.org) and we also add routes for health check and monitoring metrics.

### Ruby HTTP Trigger
For the case of Ruby we use [Sinatra](http://www.sinatrarb.com) as web framework and we add the routes required for the function and the health check. Monitoring is currently not supported yet for this framework. PR is welcome :-)

## Event trigger
This variant is used when the function is meant to be triggered through message events in a pre-deployed [Kafka](https://kafka.apache.org) system. We include a set of kafka/zookeeper in the deployment manifest of [Kubeless release package](https://github.com/kubeless/kubeless/releases) that will be deployed together with the Kubeless controller. Basically the runtimes are Kafka consumers which listen messages in a specific kafka topic and execute the injected function.

Right now the runtimes that support this kind of events are Python and NodeJS.

## Monitoring functions
Kubeless runtimes are exposing metrics at `/metrics` endpoint and these metrics will be collected by Prometheus. We also include a prometheus setup in [`manifests/monitoring`](https://github.com/kubeless/kubeless/blob/master/manifests/monitoring/prometheus.yaml) to help you easier set it up. The metrics collected are: Number of calls, succeeded and error executions and the time spent per call.

# Custom Runtime
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

Note that you can also use your own image for one of the supported runtimes. That means you have to manage how your runtime starts and looks for the injected function and executes it. Kubeless injects the function into runtime container via a [Kubernetes ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/) object mounted at `/kubeless` folder, so make sure your runtime looks for function at that folder.

You can deploy your function just specifying the runtime image:
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
