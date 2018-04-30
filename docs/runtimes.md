# Kubeless Runtime Variants

By default Kubeless has support for the following runtimes:

 - Python: For the branches 2.7, 3.4 and 3.6
 - NodeJS: For the branches 6 and 8
 - Ruby: For the branch 2.4
 - PHP: For the branch 7.2
 - Golang: For the branch 1.10

You can see the list of supported runtimes executing:

```console
$ kubeless get-server-config
INFO[0000] Current Server Config:
INFO[0000] Supported Runtimes are: python2.7, python3.4, python3.6, nodejs6, nodejs8, ruby2.4, php7.2, go1.10
```

Each runtime is encapsulated in a container image. The reference to these images are injected in the Kubeless configuration. You can find the source code of all runtimes in [`docker/runtime`](https://github.com/kubeless/kubeless/tree/master/docker/runtime).

### NodeJS

#### Example

```js
module.exports = {
  foo: function (event, context) {
    console.log(event);
    return event.data;
  }
}
```

#### Description

NodeJS functions should export the desired method using `module.exports`. You can specify dependencies using a `package.json` file. It is also possible to return an object instead of a string, this object will be stringified before returning.

When using the Node.js runtime, it is possible to configure a [custom registry or scope](https://docs.npmjs.com/misc/scope#associating-a-scope-with-a-registry) in case a function needs to install modules from a different source. For doing so it is necessary to set up the environment variables *NPM_REGISTRY* and *NPM_SCOPE* when deploying the function:

```console
$ kubeless function deploy myFunction --runtime nodejs6 \
                                --env NPM_REGISTRY=http://my-registry.com \
                                --env NPM_SCOPE=@myorg \
                                --dependencies package.json \
                                --handler test.foobar \
                                --from-file test.js
```

#### Server implementation

For the Node.js runtime we start an [Express](http://expressjs.com) server and we include the routes for serving the health check and exposing the monitoring metrics. Apart from that we enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) requests and [Morgan](https://github.com/expressjs/morgan) for handling the logging in the server. Monitoring is supported if the function is synchronous or if it uses promises.

### Python

#### Example

```py
def handler(event, context):
    print (event)
    return event['data']
```

#### Description

Python functions should define the desired method. You can specify dependencies using a `requirements.txt` file.

#### Server implementation

For python we use [Bottle](https://bottlepy.org) and we also add routes for health check and monitoring metrics.

### Ruby

#### Example

```rb
def handler(event, context)
  puts event
  JSON.generate(event[:data])
end
```

#### Description

Ruby functions should define the desired method. You can specify dependencies using a `Gemfile` file.

#### Server implementation

For the case of Ruby we use [Sinatra](http://www.sinatrarb.com) as web framework and we add the routes required for the function and the health check. Monitoring is currently not supported yet for this framework. PR is welcome :-)

### Go

#### Example

```go
package kubeless

import "github.com/kubeless/kubeless/pkg/functions"

func Handler(event functions.Event, context functions.Context) (string, error) {
	return event.Data, nil
}
```

#### Description

Ruby functions requires to import the package `github.com/kubeless/kubeless/pkg/functions` that is used to define the input parameters. The desired method should be exported in the package. You can specify dependencies using a `Gopkg.toml` file, dependencies are installed using [`dep`](https://github.com/golang/dep).

#### Server implementation

The Go HTTP server doesn't include any framework since the native packages includes enough functionality to fit our needs. Since there is not a standard package that manages server logs that functionality is implemented in the same server. It is also required to implement the `ResponseWriter` interface in order to retrieve the Status Code of the response.

#### Debugging compilation

If there is an error during the compilation of a function, the error message will be dumped to the termination log. If you see that the pod is crashed in a init container:

```console
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

#### Timeout handling

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

## Use a custom runtime

The Kubeless configuration defines a set of default container images per supported runtime variant.

These default container images can be configured via Kubernetes environment variables on the Kubeless controller's deployment container. Or modifying the `kubeless-config` ConfigMap that is deployed along with the Kubeless controller. For more information about how to modify the Kubeless configuration check [this guide](/docs/function-controller-configuration).

Apart than changing the configuration, it is possible to use a custom runtime specifying the image that the function will use. This way you are able to use any language or any binary with Kubeless as far as the image satisfies the following conditions:

 - It runs a web server listening in the port 8080
 - It exposes the endpoint `/healthz` to perform the container [liveness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)

You can check the list of desired features that a runtime image should have [in this document](/docs/implementing-new-runtime).

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
