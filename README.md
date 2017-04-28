# <img src="https://cloud.githubusercontent.com/assets/4056725/25480209/1d5bf83c-2b48-11e7-8db8-bcd650f31297.png" alt="Kubeless logo" width="400">

`kubeless` is a Kubernetes-native serverless framework. It is currently under active development, the most up-to-date version is `HEAD`. If you experience any problems during this growing phase please file an [issue](https://github.com/bitnami/kubeless/issues) and we will get back to you as quickly as we can.

There are other solutions, like [fission](http://fission.io) from Platform9, [funktion](https://github.com/fabric8io/funktion) from Fabric8. There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk). We believe however, that Kubeless is the most Kubernetes native of all.

Kubeless stands out as we use a **ThirdPartyResource** to be able to create functions as custom resources. We then run an in-cluster controller that watches these custom resources and launches _runtimes_ on-demand. These runtimes, dynamically inject the functions and make them available over HTTP or via a PubSub mechanism.

For PubSub we use [Kafka](https://kafka.apache.org). Currently we start Kafka and Zookeeper in a non-persistent setup. With `kubeless` you can create topics, and publish events that get consumed by the runtime and your even triggered functions.

## Screencasts

Demo of event based function triggers with Minio.

[![screencast](https://img.youtube.com/vi/AxZuQIJUX4s/0.jpg)](https://www.youtube.com/watch?v=AxZuQIJUX4s)

Also check our [UI](https://github.com/bitnami/kubeless-ui) project

## Usage

Download `kubeless` from the release page. Then launch the controller. It will ask you if you are OK to do it. It will create a _kubeless_ namespace and a _function_ ThirdPartyResource. You will see a _kubeless_ controller and a _kafka_ controller running.

```console
$ kubeless install
We are going to install the controller in the kubeless namespace. Are you OK with this: [Y/n]

$ kubectl get pods --namespace=kubeless
NAME                                   READY     STATUS              RESTARTS   AGE
kafka-controller-2158053540-a7n0v      0/2       ContainerCreating   0          12s
kubeless-controller-1801423959-yow3t   0/2       ContainerCreating   0          12s

$ kubectl get thirdpartyresource
NAME             DESCRIPTION                                     VERSION(S)
function.k8s.io   Kubeless: Serverless framework for Kubernetes   v1

$ kubectl get functions
```

**Note** We provide `--kafka-version` flag for specifying the kafka version will be installed and `--controller-image` in case you are willing to install your customized Kubeless controller. Without the flags, we will install the newest release of `bitnami/kubeless-controller`. Check `kubeless install --help` for more detail.

You are now ready to create functions. Then you can use the CLI to create a function. Functions have two possible types:

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

```
$ kubeless function deploy get-python --runtime python27 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-http
```

You will see the function custom resource created:

```console
$ kubectl get functions
NAME          KIND
get-python    Function.v1.k8s.io
```

### PubSub function

Messages need to be JSON messages. A function can be as simple as:

```python
def foobar(context):
    print context.json
    return context.json
```

You create it the same way than an _HTTP_ function except that you specify a `--trigger-topic`.

```
$ kubeless function deploy test --runtime python27 \
                                --handler test.foobar \
                                --from-file test.py \
                                --trigger-topic <topic_name>
```

### Other commands

You can delete and list functions:

```
$ kubeless function delete <function_name>
$ kubeless function ls
```

You can create, list and delete PubSub topics:

```
$ kubeless topic create <topic_name>
$ kubeless topic delete <topic_name>
$ kubeless topic ls
```

## Examples

See the [examples](./examples) directory for a list of various examples. Minio, SLACK, Twitter etc ...

## Building

### Building with go

- you need go v1.7+
- if your working copy is not in your `GOPATH`, you need to set it accordingly.
- we provided Makefile.

```
$ make binary
```

You can build kubeless for multiple platforms with:

```
$ make binary-cross
```

### Download kubeless package

```
$ go get -u github.com/bitnami/kubeless
```

## _Roadmap_

This is still currently a POC, feel free to land a hand. We need to implement the following high level features:

* Add other runtimes, currently only Python and NodeJS are supported
* Deploy Kafka and Zookeeper using StatefulSets for persistency
* Investigate other messaging bus
* Instrument the runtimes via Prometheus to be able to create pod autoscalers automatically
* Optimize for functions startup time
