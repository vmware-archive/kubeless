# SLACK example

Trigger the function and it send a message to your SLACK channel.

You need a SLACK_API_TOKEN

## Store token

Store your SLACK API TOKEN in a Kubernetes secret

```
kubectl create secret generic slack --from-literal=token=<your_token>
```

## Launch the function

Edit `bot.py` to specify the proper channel.

```
make slack
```

## Send a slack message

With a local proxy running:

```
curl --data '{"msg":"This is a message to SLACK"}' localhost:8080/api/v1/proxy/namespaces/default/services/slack/ --header "Content-Type:application/json"
```

## Listen to Kubernetes events in SLACK

Launch the Kafka event sync:

```
kubeless topic create k8s
kubectl run events --image=skippbox/k8s-events:0.10.12
```

Deploy the function to get triggered on k8s events and send message to SLACK

```
make slack
```
