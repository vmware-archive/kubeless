# <img src="https://cloud.githubusercontent.com/assets/4056725/25480209/1d5bf83c-2b48-11e7-8db8-bcd650f31297.png" alt="Kubeless logo" width="400">

[![Build Status](https://travis-ci.org/kubeless/kubeless.svg?branch=master)](https://travis-ci.org/kubeless/kubeless)
[![Slack](https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg)](http://slack.oss.bitnami.com)

`kubeless` is a Kubernetes-native serverless framework that lets you deploy small bits of code without having to worry about the underlying infrastructure plumbing. It leverages Kubernetes resources to provide auto-scaling, API routing, monitoring, troubleshooting and more.

Kubeless stands out as we use a ThirdPartyResource (now called [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/)) to be able to create functions as custom kubernetes resources. We then run an in-cluster controller that watches these custom resources and launches _runtimes_ on-demand. The controller dynamically injects the functions code into the runtimes and make them available over HTTP or via a PubSub mechanism.

As event system, we currently use [Kafka](https://kafka.apache.org) and bundle a kafka setup in the Kubeless namespace for development. Help to support additional event framework like [nats.io](http://nats.io/) would be more than welcome.

Kubeless is purely open-source and non-affiliated to any commercial organization. Chime in at anytime, we would love the help and feedback !

## Screencasts

Click on the picture below to see a screencast demonstrating event based function triggers with kubeless.

[![screencast](https://img.youtube.com/vi/AxZuQIJUX4s/0.jpg)](https://www.youtube.com/watch?v=AxZuQIJUX4s)

## Tools

* A [UI](https://github.com/kubeless/kubeless-ui) available. It can run locally or in-cluster.
* A [serverless framework plugin](https://github.com/serverless/serverless-kubeless) is available.

## Installation

Download `kubeless` cli from the [release page](https://github.com/kubeless/kubeless/releases). Then using one of yaml manifests found in the release package to deploy kubeless. It will create a _kubeless_ namespace and a _function_ ThirdPartyResource. You will see a _kubeless_ controller, and _kafka_, _zookeeper_ statefulset running.

There are several kubeless manifests being shipped for multiple k8s environments (non-rbac, rbac and openshift), please consider to pick up the correct one:

* `kubeless-$RELEASE.yaml` is used for non-RBAC Kubernetes cluster.
* `kubeless-rbac-$RELEASE.yaml` is used for RBAC-enabled Kubernetes cluster.
* `kubeless-openshift-$RELEASE.yaml` is used to deploy Kubeless to OpenShift (1.5+).

For example, this below is a show case of deploying kubeless to a non-RBAC Kubernetes cluster.

```console
$ export RELEASE=v0.1.0
$ kubectl create ns kubeless
$ kubectl create -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless-$RELEASE.yaml

$ kubectl get pods -n kubeless
NAME                                   READY     STATUS    RESTARTS   AGE
kafka-0                                1/1       Running   0          1m
kubeless-controller-3331951411-d60km   1/1       Running   0          1m
zoo-0                                  1/1       Running   0          1m

$ kubectl get deployment -n kubeless
NAME                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kubeless-controller   1         1         1            1           1m

$ kubectl get statefulset -n kubeless
NAME      DESIRED   CURRENT   AGE
kafka     1         1         1m
zoo       1         1         1m

$ kubectl get thirdpartyresource
NAME             DESCRIPTION                                     VERSION(S)
function.k8s.io   Kubeless: Serverless framework for Kubernetes   v1

$ kubectl get functions
```

You are now ready to create functions.

## Usage

You can use the CLI to create a function. Functions have two possible types:

* http trigger (function will expose an HTTP endpoint)
* pubsub trigger (function will consume event on a specific topic)

### HTTP function

Here is a toy:

```python
def foobar(context):
   print context.json
   return context.json
```

You create it with:

```console
$ kubeless function deploy get-python --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-http
```

You will see the function custom resource created:

```console
$ kubectl get functions
NAME          KIND
get-python    Function.v1.k8s.io

$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     TYPE    TOPIC
get-python  default     test.foobar python2.7   HTTP
```

You can then call the function with:

```console
$ kubeless function call get-python --data '{"echo": "echo echo"}'
Connecting to function...
Forwarding from 127.0.0.1:30000 -> 8080
Forwarding from [::1]:30000 -> 8080
Handling connection for 30000
{"echo": "echo echo"}
```

Or you can curl directly with `kubectl proxy`, for example:

```console
$ kubectl proxy -p 8080 &

$ curl --data '{"Another": "Echo"}' localhost:8080/api/v1/proxy/namespaces/default/services/get-python/ --header "Content-Type:application/json"
{"Another": "Echo"}
```

Kubeless also supports [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) which means you can provide your custom URL to the function. Please refer to [this doc](./docs/ingress.md) for more details.

### PubSub function

Messages need to be JSON messages. A function can be as simple as:

```python
def foobar(context):
    print context.json
    return context.json
```

You create it the same way than an _HTTP_ function except that you specify a `--trigger-topic`.

```console
$ kubeless function deploy test --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-topic test-topic
```

### Other commands

You can delete and list functions:

```console
$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     TYPE    TOPIC
test        default     test.foobar python2.7   PubSub  test-topic

$ kubeless function delete test

$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     TYPE    TOPIC
```

You can create, list and delete PubSub topics:

```console
$ kubeless topic create another-topic
Created topic "another-topic".

$ kubeless topic delete another-topic

$ kubeless topic ls
```

## Examples

See the [examples](./examples) directory for a list of various examples. Minio, SLACK, Twitter etc ...

Also checkout the [functions repository](https://github.com/kubeless/functions).

## Building

### Building with go

* You need go v1.7+
* To cross compile the kubeless command you will need [gox](https://github.com/mitchellh/gox) set up in your environment
* We use make to build the project. 
* Ensure you have a GOPATH setup

fect the project:
```console
$ got get -d github.com/kubeless/kubeless
$ cd $GOPATH/src/github.com/kubeless/kubeless/
```

build for your local system kubeless and kubeless-controller:
```console
$ make binary
```

You can build kubeless (cli) for multiple platforms with:
```console
$ make binary-cross
```
To build a deployable kubeless-controller docker image:
```console
$ make controller-image
$ docker tag kubeless-controller [your_image_name] 
```

## Comparison

There are other solutions, like [fission](http://fission.io) and [funktion](https://github.com/fabric8io/funktion). There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk). We believe however, that Kubeless is the most Kubernetes native of all.

Kubeless uses k8s primitives, there is no additional API server or API router/gateway. Kubernetes users will quickly understand how it works and be able to leverage their existing logging and monitorig setup as well as their troubleshooting skills.

## _Roadmap_

We would love to get your help, feel free to land a hand. We are currently looking to implement the following high level features:

* Add other runtimes, currently Python, NodeJS and Ruby are supported. We are also providing a way to use custom runtime. Please check [this doc](./docs/runtimes.md) for more details.
* Investigate other messaging bus (e.g nats.io)
* Instrument the runtimes via Prometheus to be able to create pod autoscalers automatically (e.g use custom metrics not just CPU)
* Optimize for functions startup time
* Add distributed tracing (maybe using istio)
* Break out the triggers and runtimes
