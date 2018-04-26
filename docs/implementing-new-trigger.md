# How to add a new event source as Trigger

Kubeless [architecture](/docs/architecture) is built on core concepts of Functions, Triggers and Runtime. A _Trigger_ in Kubeless represents association between an event source and functions that need to be invoked on an event in the event source. Kubeless fully leverages the Kubernetes concepts of [custom resource definition](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)(CRD) and [custom controllers](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers). Each trigger is expected to be modelled as Kubernetes CRD. A trigger specific custom resource controller is expected to be written that realizes how deployed functions are invoked when event occurs. Following sections document how one can add a new event source as _Trigger_ into Kubeless.

**Note**: Eventual [vision](https://github.com/kubeless/kubeless/issues/695) of the Kubeless architecture is to let any one develop a Trigger for any event source independent of Kubeless code base. Please see [695](https://github.com/kubeless/kubeless/issues/695) for details. There still work to be done to achieve the end goal. But as a interim solution below guidelines provide a way to add new triggers.

## Model event source as CRD

First step is to create a new CRD for the event source. CRD for the new triggers will be largely similar to the existing ones. For example below is the CRD for Kafka trigger

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
description: CRD object for Kafka trigger type
kind: CustomResourceDefinition
metadata:
  name: kafkatriggers.kubeless.io
spec:
  group: kubeless.io
  names:
    kind: KafkaTrigger
    plural: kafkatriggers
    singular: kafkatrigger
  scope: Namespaced
  version: v1beta1
```

Give appropriate and intutive name to the event source.

## Model the CRD spec

Once CRD is defined, you need to model the event source and its attributes as resource object spec. Please see [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md) for key attributes of Kubernetes API resource object. Except fot `Spec` part rest of the needed parts to define a Trigger are pretty similar to other Triggers.

For e.g below is the definition of [Kafka Trigger](https://github.com/kubeless/kubeless/blob/master/pkg/apis/kubeless/v1beta1/kafka_trigger.go)

```go
type KafkaTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              KafkaTriggerSpec `json:"spec"`
}
```

You need to model the event source attributes in to Spec attribute of new trigger. Depending on the nature of event source you may want to associate single function or multiple functions with the event source. Use appropriate mechanism to represent the association. For e.g Kafka trigger uses Kubernetes label selector to associate any function with matching label with the event source.

## Code Generation

Once you have definged the new trigger, please ensure its placed in `pkg/apis/kubeless/v1beta1/` path, and update `register.go` go to include new trigger type.

Now you can auto-generate the clientset, lister and informers for the new API resource object as well but running `make update` or `./hack/update-codegen.sh`. Auto-generated clientset, lister and informers comes handy in writing the controller in next step.

## CRD controller

Here is the most important step, i.e. writing controller itself. As far as the skeleton of controller goes, it would be pretty similar to existing controllers like Kafka trigger controller, nats trigger controller or http trigger controller. Functionally controller does two important things

- watch Kuberentes API server for CRUD operations on the new trigger object and take appropriate actions.
- when an event occurs in the event source trigger the associated functions.

Please read the code and logic for existing [controllers](https://github.com/kubeless/kubeless/tree/master/pkg/controller) as a reference.

## Building controller binary and docker image

Ensure you controller is an independent binary that can be built from the Makefile. Please follow one of the existing controller [cmd](https://github.com/kubeless/kubeless/tree/master/cmd) as referance. Also ensure there is corresponding `Dockerfile` to build the controller image. Please see the dockerfile for other trigger controller as a [reference](https://github.com/kubeless/kubeless/tree/master/docker).

Add appropriate Makefile targets so that controller binary and docker image can be built.

## Manifest

Create a jsonnet file for the new trigger and ensure that generated yaml file has CRD definition, deployment for the trigger controller and necessary RBAC rules. Again most of the stuff is common with Kafka or HTTP triggers, so take existing jsonnet manifests as a referance.

## CI
Please add unit tests and integration tests to tests the new functionality related to adding a trigger in to Kubeless. Also make necessary changes to .travis.yaml so that new trigger controller image can be built, deployed and integration tests can be run against the new trigger functionality. Please consider adding integration tests that conver following scenarios.

- add an integration test to enusre function gets invoked when an event occurs. For example, we have a Kafka integration test to publish a message to a topic, and test also verifies that associated function got invoked by the verifying the pods logs.
- if your Trigger supports 1:n association between Trigger and Functions, then add integration test to verify multiple functions get invoked on the occurrence of the event.
- add an integration tests to ensure `Trigger` CRD object update scenarios work fine.
