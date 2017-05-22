# Kubeless architecture

This doc covers the architectural design of Kubeless and directory structure of the repository.

## Architecture

Kubeless leverages multiple concepts of Kubernetes in order to support deploy functions on top of it. In details, we have been using:

- Third party resources (TPR) to simulate function's metadata
- Deployment / Pod to run the corresponding runtime.
- Config map to inject function's code to the runtime pod.
- Service to expose function.
- Init container to load the dependencies that function might have.

When install kubeless, there is a TPR endpoint being deploy called `function.k8s.io`:

```
$ kubectl get thirdpartyresource -o yaml
apiVersion: v1
items:
- apiVersion: extensions/v1beta1
  description: 'Kubeless: Serverless framework for Kubernetes'
  kind: ThirdPartyResource
  metadata:
    creationTimestamp: 2017-03-23T11:55:41Z
    name: function.k8s.io
    resourceVersion: "123126"
    selfLink: /apis/extensions/v1beta1/thirdpartyresourcesfunction.k8s.io
    uid: a3933b1f-0fbf-11e7-b235-12d427b26198
  versions:
  - name: v1
kind: List
metadata: {}
```

Then function custom objects will be created under this thirdparty endpoint. A function object looks like this:

```
$ kubectl get function -o yaml
apiVersion: v1
items:
- apiVersion: k8s.io/v1
  kind: Function
  metadata:
    creationTimestamp: 2017-03-27T13:56:02Z
    name: get-python
    namespace: default
    resourceVersion: "146417"
    selfLink: /apis/k8s.io/v1/namespaces/default/functions/get-python
    uid: 1d08377f-12f5-11e7-953d-0eaf474e070b
  spec:
    deps: ""
    function: |
      import json
      def foo():
          return "hello world"
    handler: helloget.foo
    runtime: python27
    topic: ""
    type: HTTP
kind: List
metadata: {}
```

`function.spec` contains function's metadata including code, handler, runtime, type (http or pubsub) and probably its dependency file.

We write a custom controller named `Kubeless-controller` to continuously watch changes of function objects and react accordingly to deploy/delete K8S deployment/svc/configmap. By default Kubeless-controller is installed into `kubeless` namespace. Here we use configmap to inject the function code from function.spec.function into the corresponding k8s runtime pod.

There are currently two type of functions supported in Kubeless: http-based and pubsub-based. A set of Kafka and Zookeeper is installed into the `kubeless` namespace to handle the pubsub-based functions.

## Kubeless command-line client

Together with `kubeless-controller`, we provide `kubeless` cli which enables users to interact with Kubeless system. At this moment, Kubeless cli provides these below actions:

```
$ kubeless --help
Serverless framework for Kubernetes

Usage:
  kubeless [command]

Available Commands:
  function    function specific operations
  install     Install Kubeless controller
  topic       manage message topics in Kubeless
  version     Print the version of Kubeless

Use "kubeless [command] --help" for more information about a command.
```

Diving into these above commands for more details.

## Implementation

Kubeless controller is written in Go programming language, and uses the Kubernetes go client to interact with the Kubernetes API server. Detailed implementation of the controller could be found at https://github.com/kubeless/kubeless/blob/master/pkg/controller/controller.go#L113

Kubeless CLI is written in Go as well, using the popular cli library `github.com/spf13/cobra`. Basically it is a bundle of HTTP requests and kubectl commands. We send http requests to the Kubernetes API server in order to 'crud' TPR objects. Other actions are just wrapping up `kubectl port-forward`, `kubectl exec` and `kubectl logs` in order to make direct call to function, inject message to Kafka controller or get function log. Checkout https://github.com/kubeless/kubeless/tree/master/cmd/kubeless for more details.

## Directory structure

In order to help you getting a better feeling before you start diving into the project, we would give you the 10,000 foot view of the source code directory structure.

```
- cmd: contains kubeless cli implementation and kubeless-controller.
- docker: contains artifacts for building the kubeless-controller and runtime images.
- docs: contains documentations.
- examples: contains some samples of running function with kubeless.
- pkg: contains shared packages.
    - controller: kubeless controller implementation.
    - function: kubeless function crud.
    - spec: kubeless function spec.
    - utils: k8s utilities which helps to interact with k8s apiserver.
- script: contains build scripts.
- vendor: contains dependencies packages.
- version: nothing but kubeless version.
```