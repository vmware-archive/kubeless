# Kubeless architecture

This doc covers the architectural design of Kubeless and directory structure of the repository.

## Architecture

Kubeless leverages multiple concepts of Kubernetes in order to support deploy functions on top of it. In details, we have been using:

- Custom Resource Definitions (CRD) to simulate function's metadata
- Deployment / Pod to run the corresponding runtime.
- Config map to inject function's code to the runtime pod.
- Service to expose function.
- Init container to load the dependencies that function might have.

When install kubeless, there is a CRD endpoint being deploy called `function.kubeless.io`:

```yaml
$ kubectl get customresourcedefinition -o yaml
apiVersion: v1
items:
- apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    ...
    name: functions.kubeless.io
    ...
    selfLink: /apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/functions.kubeless.io
  spec:
    group: kubeless.io
    names:
      kind: Function
      listKind: FunctionList
      plural: functions
      singular: function
    scope: Namespaced
    version: v1beta1
  status:
    acceptedNames:
      kind: Function
      listKind: FunctionList
      plural: functions
      singular: function
    conditions:
      ...
kind: List
```

Then function custom objects will be created under this CRD endpoint. A function object looks like this:

```yaml
$ kubectl get function -o yaml
apiVersion: v1
items:
- apiVersion: kubeless.io/v1beta1
  kind: Function
  metadata:
    clusterName: ""
    creationTimestamp: 2018-02-21T16:21:15Z
    labels:
      created-by: kubeless
    name: get-python
    namespace: default
    resourceVersion: "378"
    selfLink: /apis/kubeless.io/v1beta1/namespaces/default/functions/get-python
    uid: 3d544cf3-1723-11e8-ac5d-080027ae63b7
  spec:
    checksum: sha256:0807cad785f8bdb0d272bfd26cf34b44fa124189644ad7852c2c5a071ccd4a66
    deployment:
      metadata:
        creationTimestamp: null
      spec:
        strategy: {}
    template:
      metadata:
        creationTimestamp: null
      spec:
        containers:
        - name: ""
          resources: {}
      status: {}
    deps: ""
    function: |
      def foo():
          return "hello world"
    function-content-type: text
    handler: helloget.foo
    horizontalPodAutoscaler:
      metadata:
        creationTimestamp: null
      spec:
        maxReplicas: 0
        scaleTargetRef:
          kind: ""
          name: ""
      status:
        conditions: null
        currentMetrics: null
        currentReplicas: 0
        desiredReplicas: 0
    runtime: python2.7
    schedule: ""
    service:
      ports:
      - name: http-function-port
        port: 8080
        protocol: TCP
        targetPort: 8080
      selector:
        created-by: kubeless
        function: get-python
      type: ClusterIP
    timeout: "180"
    topic: ""
    type: HTTP
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

`function.spec` contains function's metadata including code, handler, runtime, type (http or pubsub) and probably its dependency file.

We write a custom controller named `Kubeless-controller` to continuously watch changes of function objects and react accordingly to deploy/delete K8S deployment/svc/configmap. By default Kubeless-controller is installed into `kubeless` namespace. Here we use configmap to inject the function code from function.spec.function into the corresponding k8s runtime pod.

There are currently three type of functions supported in Kubeless: http-based, scheduled and pubsub-based. A set of Kafka and Zookeeper is installed into the `kubeless` namespace to handle the pubsub-based functions.

## Kubeless command-line client

Together with `kubeless-controller`, we provide `kubeless` cli which enables users to interact with Kubeless system. At this moment, Kubeless cli provides these below actions:

```console
$ kubeless --help
Serverless framework for Kubernetes

Usage:
  kubeless [command]

Available Commands:
  autoscale         manage autoscale to function on Kubeless
  function          function specific operations
  get-server-config Print the current configuration of the controller
  help              Help about any command
  route             manage route to function on Kubeless
  topic             manage message topics in Kubeless
  version           Print the version of Kubeless

Flags:
  -h, --help   help for kubeless

Use "kubeless [command] --help" for more information about a command.
```

## Implementation

Kubeless controller is written in Go programming language, and uses the Kubernetes go client to interact with the Kubernetes API server.

Kubeless CLI is written in Go as well, using the popular cli library `github.com/spf13/cobra`. Basically it is a bundle of HTTP requests and kubectl commands. We send http requests to the Kubernetes API server in order to 'crud' CRD objects. Checkout [the cmd folder](https://github.com/kubeless/kubeless/tree/master/cmd/kubeless) for more details.

## Directory structure

In order to help you getting a better feeling before you start diving into the project, we would give you the 10,000 foot view of the source code directory structure.

- chart: chart to deploy Kubeless with Helm
- cmd: contains kubeless cli implementation and kubeless-controller.
- docker: contains artifacts for building the kubeless-controller and runtime images.
- docs: contains documentations.
- examples: contains some samples of running function with kubeless.
- manifests: collection of manifests for additional features
- pkg: contains shared packages.
- script: contains build scripts.
- vendor: contains dependencies packages.
