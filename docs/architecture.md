# Kubeless architecture

This doc covers the architectural design of Kubeless and directory structure of the repository.

## Architecture

Kubeless leverages multiple concepts of Kubernetes in order to support deploy functions on top of it. In details, we have been using:

- Custom Resource Definitions (CRD) to simulate function's metadata
- Deployment / Pod to run the corresponding runtime.
- Config map to inject function's code to the runtime pod.
- Service to expose function.
- Init container to load the dependencies that function might have.

When install kubeless, there is a CRD endpoint being deploy called `function.k8s.io`:

```
$ kubectl get customresourcedefinition -o yaml
apiVersion: v1
items:
- apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"apiextensions.k8s.io/v1beta1","description":"Kubernetes Native Serverless Framework","kind":"CustomResourceDefinition","metadata":{"annotations":{},"name":"functions.k8s.io","namespace":""},"spec":{"group":"k8s.io","names":{"kind":"Function","plural":"functions","singular":"function"},"scope":"Namespaced","version":"v1"}}
    creationTimestamp: 2017-10-10T14:51:37Z
    name: functions.k8s.io
    namespace: ""
    resourceVersion: "7943"
    selfLink: /apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/functions.k8s.io
    uid: 846770f8-adca-11e7-b48e-0a580a160058
  spec:
    group: k8s.io
    names:
      kind: Function
      listKind: FunctionList
      plural: functions
      singular: function
    scope: Namespaced
    version: v1
  status:
    acceptedNames:
      kind: Function
      listKind: FunctionList
      plural: functions
      singular: function
    conditions:
    - lastTransitionTime: null
      message: no conflicts found
      reason: NoConflicts
      status: "True"
      type: NamesAccepted
    - lastTransitionTime: 2017-10-10T14:51:37Z
      message: the initial names have been accepted
      reason: InitialNamesAccepted
      status: "True"
      type: Established
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

Then function custom objects will be created under this CRD endpoint. A function object looks like this:

```
$ kubectl get function -o yaml
apiVersion: v1
items:
- apiVersion: k8s.io/v1
  kind: Function
  metadata:
    clusterName: ""
    creationTimestamp: 2017-10-10T16:09:06Z
    deletionGracePeriodSeconds: null
    deletionTimestamp: null
    name: get-python
    namespace: default
    resourceVersion: "13299"
    selfLink: /apis/k8s.io/v1/namespaces/default/functions/get-python
    uid: 5759e3f3-add5-11e7-b48e-0a580a160058
  spec:
    deps: ""
    function: |
      def foobar(context):
         print context.json
         return context.json
    handler: test.foobar
    runtime: python2.7
    schedule: ""
    template:
      metadata:
        creationTimestamp: null
      spec:
        containers:
        - name: ""
          resources: {}
    topic: ""
    type: HTTP
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
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

Kubeless CLI is written in Go as well, using the popular cli library `github.com/spf13/cobra`. Basically it is a bundle of HTTP requests and kubectl commands. We send http requests to the Kubernetes API server in order to 'crud' CRD objects. Other actions are just wrapping up `kubectl port-forward`, `kubectl exec` and `kubectl logs` in order to make direct call to function, inject message to Kafka controller or get function log. Checkout https://github.com/kubeless/kubeless/tree/master/cmd/kubeless for more details.

## Directory structure

In order to help you getting a better feeling before you start diving into the project, we would give you the 10,000 foot view of the source code directory structure.

```
- chart: chart to deploy Kubeless with Helm
- cmd: contains kubeless cli implementation and kubeless-controller.
- docker: contains artifacts for building the kubeless-controller and runtime images.
- docs: contains documentations.
- examples: contains some samples of running function with kubeless.
- manifests: collection of manifests for development
- pkg: contains shared packages.
    - controller: kubeless controller implementation.
    - function: kubeless function crud.
    - spec: kubeless function spec.
    - utils: k8s utilities which helps to interact with k8s apiserver.
- script: contains build scripts.
- vendor: contains dependencies packages.
- version: nothing but kubeless version.
```
