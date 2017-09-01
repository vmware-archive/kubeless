# Runtimes support

Right now Kubeless has support for the following runtimes:
 - Python: For the branch 2.7.X
 - NodeJS: For the branches 6.X and 8.X
 - Ruby: For the branch 2.4.X

Each runtime is encapsulated in a container image. The reference to these images are injected in the Kubeless controller.

 # Runtime variants
 ## HTTP Trigger
This variant is used when the function is meant to be triggered through HTTP. For doing so we use a web framework that is in charge of receiving request and redirect them to the function. This kind of trigger is supported for all the runtimes.

### NodeJS HTTP Trigger
For the NodeJS runtime we start an [Express](http://expressjs.com) server and we include the routes for serving the health check and the metrics. Apart from that we enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) requests and [Morgan](https://github.com/expressjs/morgan) for handling the logging in the server.

Monitoring is supported if the function is synchronous or if it uses promises. The statistics collected are: Number of calls, successed and errored executions and the time spent per call. They are availble when accesing the route `/metrics`.

### Python HTTP Trigger
TO BE FULFILLED

### Ruby HTTP Trigger
For the case of Ruby we use Sinatra as web framework and we add the routes required for the function and the health check. Monitoring is not supported yet for this framework.

## Event trigger
This variant is used when the function is meant to be triggered through message events in a certain [Kafka](https://kafka.apache.org) topic.

Right now the runtimes that support this kind of events are Python and NodeJS.