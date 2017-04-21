# Kubeless

`kubeless` is a proof of concept to develop a serverless framework for Kubernetes.

There are other solutions, like [fission](http://fission.io) from Platform9, [funktion](https://github.com/fabric8io/funktion) from Fabric8. There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk).

Kubeless stands out as we use a ThirdPartyResource to be able to create functions as custom resources. We then run an in-cluster controller that watches these custom resources and launches _runtimes_ on-demand. These runtimes, dynamically inject the functions and make them available over HTTP or via a PubSub mechanism.

For PubSub we use [Kafka](https://kafka.apache.org). Currently we start Kafka and Zookeeper in a non-persistent setup. With `kubeless` you can create topics, and publish events that get consumed by the runtime.

## Screencasts

From the December 8th 2016 Kubernetes Community meeting

[![screencast](https://img.youtube.com/vi/gRVuFupq1Y4/0.jpg)](https://www.youtube.com/watch?v=gRVuFupq1Y4)

## Usage

Download `kubeless` from the release page. Then launch the controller. It will ask you if you are OK to do it. It will create a _kubeless_ namespace and a _function_ ThirdPartyResource. You will see a _kubeless_ controller and a _kafka_ controller running.

```console
$ kubeless install
We are going to install the controller in the kubeless namespace. Are you OK with this: [Y/N]
Y
INFO[0002] Initializing Kubeless controller...           pkg=controller
INFO[0002] Installing Kubeless controller into Kubernetes deployment...  pkg=controller
INFO[0002] Kubeless controller installation finished!    pkg=controller
INFO[0002] Installing Message Broker into Kubernetes deployment...  pkg=controller
INFO[0002] Message Broker installation finished!         pkg=controller

$ kubectl get pods --namespace=kubeless
NAME                                   READY     STATUS              RESTARTS   AGE
kafka-controller-2158053540-a7n0v      0/2       ContainerCreating   0          12s
kubeless-controller-1801423959-yow3t   0/2       ContainerCreating   0          12s

$ kubectl get thirdpartyresource
NAME             DESCRIPTION                                     VERSION(S)
function.k8s.io   Kubeless: Serverless framework for Kubernetes   v1

$ kubectl get functions
```

**Note** We provide `--kafka-version` flag for specifying the kafka version will be installed and `--controller-image` in case you are willing to install your customized Kubeless controller. Without the flags, we will install the newest release of `bitnami/kubeless-controller` and the latest `wurstmeister/kafka`. Check `kubeless install --help` for more detail.

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
$ kubeless function deploy test --runtime python27 \
                              --handler test.foobar \
                              --from-file test.py \
                              --trigger-http
```

You will see the function custom resource created:

```console
$ kubectl get functions
NAME      LABELS    DATA
test      <none>    {"apiVersion":"k8s.io/v1","kind":"Function","metadat...
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

To test your endpoints you can call the function directly with the `kubeless` CLI:

```
$ kubeless function call test --data '{"kube":"coodle"}'
```

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

## Download kubeless package

```
$ go get -u github.com/bitnami/kubeless
```

## _Roadmap_

This is still currently a POC, feel free to land a hand. We need to implement the following high level features:

* Add other runtimes, currently only Python and NodeJS is supported
* Deploy Kafka and Zookeeper using StatefulSets for persistency
* Instrument the runtimes via Prometheus to be able to create pod autoscalers automatically
