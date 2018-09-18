# How to add a new event source as Trigger

Kubeless [architecture](/docs/architecture) is built on core concepts of Functions, Triggers and Runtime. A _Trigger_ in Kubeless represents association between an event source and functions that need to be invoked on an event in the event source. Kubeless fully leverages the Kubernetes concepts of [custom resource definition](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)(CRD) and [custom controllers](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers). Each trigger is expected to be modelled as Kubernetes CRD. A trigger specific custom resource controller is expected to be written that realizes how deployed functions are invoked when event occurs. Following sections document how one can add a new event source as _Trigger_ into Kubeless.

## Triggers development repository

Each Kubeless trigger controller is being developed on its own repository. You can find more information about those controllers in their repositories. If you want to create a new trigger you will need to create a new repository for that. These are the triggers currently available that can be used as templates for new ones:

 - [HTTP Trigger](https://github.com/kubeless/http-trigger)
 - [CronJob Trigger](https://github.com/kubeless/cronjob-trigger)
 - [Kafka Trigger](https://github.com/kubeless/kafka-trigger)
 - [NATS Trigger](https://github.com/kubeless/nats-trigger)

## Model event source as CRD

First step is to create a new CRD for the event source. CRD for the new triggers will be largely similar to the existing ones. For example below is the CRD for Kafka trigger

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
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

For e.g below is the definition of [Kafka Trigger](https://github.com/kubeless/kafka-trigger/blob/master/pkg/apis/kubeless/v1beta1/kafka_trigger.go)

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

Now you can auto-generate the clientset, lister and informers for the new API resource object as well but running `make update` or `./hack/update-codegen.sh` within the trigger repository. Auto-generated clientset, lister and informers comes handy in writing the controller in next step.

## CRD controller

Here is the most important step, i.e. writing controller itself. As far as the skeleton of controller goes, it would be pretty similar to existing controllers like Kafka trigger controller, nats trigger controller or http trigger controller. Functionally controller does two important things

- watch Kuberentes API server for CRUD operations on the new trigger object and take appropriate actions.
- when an event occurs in the event source trigger the associated functions.

Please read the code and logic for the existing [Kafka controller](https://github.com/kubeless/kafka-trigger/tree/master/pkg/controller) as a reference.

## Building controller binary and docker image

Ensure you controller is an independent binary that can be built from the Makefile. Please follow one of the existing controller [cmd](https://github.com/kubeless/kafka-trigger/tree/master/cmd) as referance. Also ensure there is corresponding `Dockerfile` to build the controller image. Please see the dockerfile for other trigger controller as a [reference](https://github.com/kubeless/kafka-trigger/tree/master/docker).

Add appropriate Makefile targets so that controller binary and docker image can be built.

## Manifest

Create a jsonnet file for the new trigger and ensure that generated yaml file has CRD definition, deployment for the trigger controller and necessary RBAC rules. Again most of the stuff is common with Kafka or HTTP triggers, so take existing jsonnet manifests as a referance.

## CI

Once the new trigger is working it's important to add tests to ensure and preserve the trigger functionality. Each trigger should contain:

 - Unit tests covering the basic functionality.
 - End-to-end tests that ensure the compatibility with the latest image of the Kubeless core.

The CI used to run the tests is TravisCI. You can check examples of how Travis is configured [here](https://github.com/kubeless/kafka-trigger/blob/master/.circleci/config.yml). This file should define at least 4 jobs:

 - One for building the binaries and manifests.
 - Another one to test the functionality end-to-end in a Minikube scenario.
 - A third one to push the image used in the tests as `latest`.
 - A final one to auto generate a release in Github in case it's building a new tag.

Most of the functionality for the above depends on scripts that have been already developed so you just need to change some data and names from the YAML to make it work.

The tests to run are defined in the folder `tests/` of each repository. These are [`bats`](https://github.com/sstephenson/bats) that loads a common library (`script/libtest.bash`) and execute some simple scenarios. Again you can take Kafka as an example for some useful scenarios to test.
