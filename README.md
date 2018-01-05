# <img src="https://cloud.githubusercontent.com/assets/4056725/25480209/1d5bf83c-2b48-11e7-8db8-bcd650f31297.png" alt="Kubeless logo" width="400">

[![Build Status](https://travis-ci.org/kubeless/kubeless.svg?branch=master)](https://travis-ci.org/kubeless/kubeless)
[![Slack](https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg)](http://slack.oss.bitnami.com)

`kubeless` is a Kubernetes-native serverless framework that lets you deploy small bits of code without having to worry about the underlying infrastructure plumbing. It leverages Kubernetes resources to provide auto-scaling, API routing, monitoring, troubleshooting and more.

Kubeless stands out as we use a [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/) to be able to create functions as custom kubernetes resources. We then run an in-cluster controller that watches these custom resources and launches _runtimes_ on-demand. The controller dynamically injects the functions code into the runtimes and make them available over HTTP or via a PubSub mechanism.

Kubeless is purely open-source and non-affiliated to any commercial organization. Chime in at anytime, we would love the help and feedback !

## Screencasts

Click on the picture below to see a screencast demonstrating event based function triggers with kubeless.

[![screencast](https://img.youtube.com/vi/AxZuQIJUX4s/0.jpg)](https://www.youtube.com/watch?v=AxZuQIJUX4s)

Click on this next picture to see a screencast demonstrating our [serverless](https://serverless.com/framework/docs/providers/kubeless/) plugin:

[![serverless](https://img.youtube.com/vi/ROA7Ig7tD5s/0.jpg)](https://www.youtube.com/watch?v=ROA7Ig7tD5s)

## Tools

* A [UI](https://github.com/kubeless/kubeless-ui) available. It can run locally or in-cluster.
* A [serverless framework plugin](https://github.com/serverless/serverless-kubeless) is available.

## Installation

Installation is made of three steps:

* Download the `kubeless` CLI from the [release page](https://github.com/kubeless/kubeless/releases). (OSX users can also use [brew](https://brew.sh/): `brew install kubeless/tap/kubeless`).
* Create a `kubeless` namespace (used by default)
* Then use one of the YAML manifests found in the release page to deploy kubeless. It will create a _functions_ Custom Resource Definition and launch a controller. You will see a _kubeless_ controller, a _kafka_ and a _zookeeper_ statefulset appear and shortly get in running state.

There are several kubeless manifests being shipped for multiple k8s environments (non-rbac, rbac and openshift), pick the one that corresponds to your environment:

* [`kubeless-$RELEASE.yaml`](https://github.com/kubeless/kubeless/releases/download/v0.2.4/kubeless-v0.2.4.yaml) is used for non-RBAC Kubernetes cluster.
* [`kubeless-rbac-$RELEASE.yaml`](https://github.com/kubeless/kubeless/releases/download/v0.2.4/kubeless-rbac-v0.2.4.yaml) is used for RBAC-enabled Kubernetes cluster.
* [`kubeless-openshift-$RELEASE.yaml`](https://github.com/kubeless/kubeless/releases/download/v0.2.4/kubeless-openshift-v0.2.4.yaml) is used to deploy Kubeless to OpenShift (1.5+).

For example, this below is a show case of deploying kubeless to a non-RBAC Kubernetes cluster.

```console
$ export RELEASE=v0.3.3
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

$ kubectl get customresourcedefinition
NAME               KIND
functions.k8s.io   CustomResourceDefinition.v1beta1.apiextensions.k8s.io

$ kubectl get functions
```

NOTE: Kafka statefulset uses a PVC (persistent volume claim). Depending on the configuration of your cluster you may need to provision a PV (Persistent Volume) that matches the PVC or configure dynamic storage provisioning. Otherwise Kafka pod will fail to get scheduled. Also note that Kafka is only required for PubSub functions, you can still use http triggered functions. Please refer to [PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) documentation on how to provision storage for PVC.

You are now ready to create functions.

## Usage

You can use the CLI to create a function. Functions have three possible types:

* http triggered (function will expose an HTTP endpoint)
* pubsub triggered (function will consume event on a specific topic)
* schedule triggered (function will be called on a cron schedule)

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
                                --from-file test.py \
                                --handler test.foobar \
                                --trigger-http
INFO[0000] Deploying function...
INFO[0000] Function get-python submitted for deployment
INFO[0000] Check the deployment status executing 'kubeless function ls get-python'
```

Let's dissect the command:

* `get-python`: This is the name of the function we want to deploy.
* `--runtime python2.7`: This is the runtime we want to use to run our function. Available runtimes are shown in the help information.
* `--from-file test.py`: This is the file containing the function code. It is supported to specify a zip file as far as it doesn't exceed the maximum size for an etcd entry (1 MB).
* `--handler test.foobar`: This specifies the file and the exposed function that will be used when receiving requests. In this example we are using the function `foobar` from the file `test.py`.
* `--env` to pass env vars to the function like `--env foo=bar,bar=foo`. See the [detail](https://github.com/kubeless/kubeless/pull/316#issuecomment-332172876)
* `--trigger-http`: This sets the function trigger.

Other available trigger options (defaults to `--trigger-http`) are:

* `--trigger-http` to trigger the function using HTTP requests.
* `--trigger-topic` to trigger the function with a certain Kafka topic. See the [next example](#pubsub-function).
* `--timeout string` to specify the timeout (in seconds) for the function to complete its execution (default "180")
* `--schedule` to trigger the function following a certain schedule using Cron notation. F.e. `--schedule "*/10 * * * *"` would trigger the function every 10 minutes.

You can find the rest of options available when deploying a function executing `kubeless function deploy --help`

You will see the function custom resource created:

```console
$ kubectl get functions
NAME          KIND
get-python    Function.v1.k8s.io

$ kubeless function ls
NAME           	NAMESPACE	HANDLER              	RUNTIME  	TYPE  	TOPIC      	DEPENDENCIES	STATUS
get-python     	default  	helloget.foo         	python2.7	HTTP  	           	            	1/1 READY
```

You can then call the function with:

```console
$ kubeless function call get-python --data '{"echo": "echo echo"}'
{"echo": "echo echo"}
```

Or you can curl directly with `kubectl proxy`
using an [apiserver proxy
URL](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#manually-constructing-apiserver-proxy-urls).
For example:

```console
$ kubectl proxy -p 8080 &

$ curl -L --data '{"Another": "Echo"}' localhost:8080/api/v1/proxy/namespaces/default/services/get-python:function-port/ --header "Content-Type:application/json"
{"Another": "Echo"}
```

Kubeless also supports [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) which means you can provide your custom URL to the function. Please refer to [this doc](./docs/routing.md) for more details.

### PubSub function

A function can be as simple as:

```python
def foobar(context):
    print context
    return context
```

You create it the same way than an _HTTP_ function except that you specify a `--trigger-topic`.

```console
$ kubeless function deploy test --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-topic test-topic
```

After that you can invoke them publishing messages in that topic. To allow you to easily manage topics `kubeless` provides a convenience function `kubeless topic`. You can create/delete and publish to a topic easily.

```console
$ kubeless topic create test-topic
$ kubeless topic publish --topic test-topic --data "Hello World!"
```

You can check the result in the pod logs:

```console
$ kubectl logs test-695251588-cxwmc
Hello World!
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

Also checkout the [functions repository](https://github.com/kubeless/functions),
where we're building a library of ready to use kubeless examples, including an
[incubator](https://github.com/kubeless/functions/tree/master/incubator)
to encourage contributions from the community - **your PR is welcome** ! :)

## Documentation

More details can be found in the complete [Documentation](docs/)

## Building

Consult the [developer's guide](docs/dev-guide.md) for a complete set of instruction
to build kubeless.

## Comparison

There are other solutions, like [fission](http://fission.io) and [funktion](https://github.com/fabric8io/funktion). There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk). We believe however, that Kubeless is the most Kubernetes native of all.

Kubeless uses k8s primitives, there is no additional API server or API router/gateway. Kubernetes users will quickly understand how it works and be able to leverage their existing logging and monitoring setup as well as their troubleshooting skills.

## _Roadmap_

We would love to get your help, feel free to land a hand. We are currently looking to implement the following high level features:

* Add other runtimes, currently Python, NodeJS, Ruby and .Net Core are supported. We are also providing a way to use custom runtime. Please check [this doc](./docs/runtimes.md) for more details.
* Investigate other messaging bus (e.g nats.io)
* Use a standard interface for events
* Optimize for functions startup time
* Add distributed tracing (maybe using istio)
* Decouple the triggers and runtimes

## Community

**Issues**: If you find any issues, please [file it](https://github.com/kubeless/kubeless/issues).

**Meetings**: [Thursday at 10:30 UTC](https://meet.google.com/rbr-gcjp-xxz) (Weekly). [Convert to your timezone](http://www.thetimezoneconverter.com/?t=10:30&tz=UTC)

Meeting notes and agenda can be found [here](https://docs.google.com/document/d/1-OsikjjQVHVFoXBHUbkRogrzzZijQ9MumFpLfWCCjwk/edit). Meeting records can be found [here](https://www.youtube.com/user/bitrock5/)

**Slack**: We're fairly active on [slack](http://slack.oss.bitnami.com) and you can find us in the #kubeless channel.
