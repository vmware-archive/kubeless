# Kubeless
[![Join us on Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](https://skippbox.herokuapp.com)

`kubeless` is a proof of concept to develop a serverless framework for Kubernetes.

They are other solutions, like [fission](http://fission.io) form Platform9, [funktion](https://github.com/fabric8io/funktion) from Fabric8. There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk).

Kubeless stands out as we use a ThirdPartyResource to be able to create functions as custom resources. We then run an in-cluster controller that watches these custome resources and launches _runtimes_ on-demand. These runtimes, dynamically inject the functions and make them available over HTTP or via a PubSub mechanism.

For PubSub we use [Kafka](https://kafka.apache.org). Currently we start Kafka and Zookeeper in a non-persistent setup. With `kubeless` you can create topics, and publish events that get consumed by the runtime.

## Screencasts

From the December 8th 2016 Kubernetes Community meeting

<iframe width="560" height="315" src="https://www.youtube.com/embed/gRVuFupq1Y4" frameborder="0" allowfullscreen></iframe>

## Usage

Download `kubeless` from the release page. Then launch the controller. It will ask you if you are OK to do it. It will create a _kubeless_ namespace and a _lambda_ ThirdPartyResource. You will see a _kubeless_ controller and a _kafka_ controller running.

```
$ kubeless install
We are going to install the controller in the default namespace. Are you OK with this: [Y/N]
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
lamb-da.k8s.io   Kubeless: Serverless framework for Kubernetes   v1

$ kubectl get lambdas
```

You are now ready to create functions. You need to run a proxy locally (soon to be fixed). Then you can use the CLI to create a function. Functions have two possible types:

* http trigger (function will expose an HTTP endpoint)
* pubsub trigger (function will consume event on a specific topic)

### HTTP function

Here is a toy:

```
def foobar(context):
   print context.json
   return context.json
```

You create it with:

```
kubeless function create test --runtime python27 \
                              --handler test.foobar \
                              --from-file test.py \
                              --trigger-http
```

You will see the lambda custom resource created:

```
$ kubectl get lambdas
NAME      LABELS    DATA
test      <none>    {"apiVersion":"k8s.io/v1","kind":"LambDa","metadat...
```

### PubSub function

Messages need to be JSON messages. A function cane be as simple as:

```
def foobar(context):
   print context.json
   return context.json
```

You create it the same way than an _HTTP_ function except that you specify a `--trigger-topic`.

```
kubeless function create test --runtime python27 \
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
$ kubeless function call test --data {'kube':'coodle'}
```

## Building

### Building with go

- you need go v1.5 or later.
- if your working copy is not in your `GOPATH`, you need to set it accordingly.

```console
$ go build -o kubeless main.go
```

## Download kubeless package

```console
$ go get -u github.com/skippbox/kubeless
```

## _Roadmap_

This is still currently a POC, feel free to land a hand. We need to implement the following high level features:

* Add other runtimes, currently only Python is supported
* Deploy Kafka and Zookeeper using StatefulSets for persistency
* Instrument the runtimes via Prometheus to be able to create pod autoscalers automatically
* Get rid off the need for a proxy by switching k8s clients to monitor the custom resources.
