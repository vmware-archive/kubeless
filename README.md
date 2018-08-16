# <img src="https://cloud.githubusercontent.com/assets/4056725/25480209/1d5bf83c-2b48-11e7-8db8-bcd650f31297.png" alt="Kubeless logo" width="400">

[![CircleCI](https://circleci.com/gh/kubeless/kubeless.svg?style=svg)](https://circleci.com/gh/kubeless/kubeless)
[![Slack](https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg)](http://slack.k8s.io)

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

## Quick start

Check out the instructions for quickly set up Kubeless [here](http://kubeless.io/docs/quick-start).

## Building

Consult the [developer's guide](docs/dev-guide.md) for a complete set of instruction
to build kubeless.

## Comparison

There are other solutions, like [fission](http://fission.io) and [funktion](https://github.com/fabric8io/funktion). There is also an incubating project at the ASF: [OpenWhisk](https://github.com/openwhisk/openwhisk). We believe however, that Kubeless is the most Kubernetes native of all.

Kubeless uses k8s primitives, there is no additional API server or API router/gateway. Kubernetes users will quickly understand how it works and be able to leverage their existing logging and monitoring setup as well as their troubleshooting skills.

## Compatibility Matrix with Kubernetes

Kubeless fully supports two major versions of Kubernetes (1.8 and 1.9) at the moment. For other versions some of the features in Kubeless may not be available. Our CI run tests against two different platforms: GKE (1.8) and Minikube (1.9). Other platforms are supported but fully compatibiliy cannot be assured. This is the summary of the features and versions supported:

| Platform | Kubernetes Version | HTTP functions | Scheduled functions | PubSub (Kafka) functions | PubSub (NATS) functions | Autoscaling (CPU) |
| ------------- | ----- | - | - | - | - | - |
| GKE           | 1.7.X | ✓ | X | ✓ | ✓ | X |
| GKE           | 1.8.X | ✓ | ✓ | ✓ | ✓ | ✓ |
| GKE (CI)      | 1.9.X | ✓ | ✓ | ✓ | ✓ | ✓ |
| Minikube      | 1.7.X | ✓ | X | ✓ | ✓ | ✓ |
| Minikube      | 1.8.X | ✓ | ✓ | ✓ | ✓ | ✓ |
| Minikube (CI) | 1.9.X | ✓ | ✓ | ✓ | ✓ | ✓ |

## _Roadmap_

We would love to get your help, feel free to lend a hand. We are currently looking to implement the following high level features:

* Add other runtimes, currently Python, NodeJS, Ruby, PHP, .NET and Ballerina are supported. We are also providing a way to use custom runtime. Please check [this doc](./docs/runtimes.md) for more details.
* Investigate other messaging bus (e.g SQS, rabbitMQ)
* Optimize for functions startup time
* Add distributed tracing (maybe using istio)

## Community

**Issues**: If you find any issues, please [file it](https://github.com/kubeless/kubeless/issues).

**Slack**: We're fairly active on [slack](http://slack.k8s.io) and you can find us in the #kubeless channel.
