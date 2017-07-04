# Container to feed k8s events to kafka

`events.py` is a Python 3.4 script, that uses `asyncio` and the Kubernetes python client plus a Kafka client to watch for k8s events and send those events onto the kubeless Kafka _k8s_ topic.

The Dockerfile just builds an image to start this as a deployment in a k8s cluster running kubeless.

## Usage

Create the `k8s` topic in kubeless:

```
kubeless topic create k8s
```

Then launch the event sync

```
kubectl run event --image=skippbox/k8s-events:0.10.12
```
