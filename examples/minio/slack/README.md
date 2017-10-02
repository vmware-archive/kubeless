# Send messages on objects upload to a bucket

This function is triggered when an object is added to a bucket and sends a Slack message

You need a [SLACK_API_TOKEN](https://api.slack.com/custom-integrations/legacy-tokens)

## Store token

Store your SLACK API TOKEN in a Kubernetes secret

```
kubectl create secret generic slack --from-literal=token=<your_token>
```

Now edit `bot.py` to specify the proper channel.

## Launch the function

Once you have done the steps above you can deploy your function:

```
kubeless function deploy slack --trigger-topic s3 --from-file bot.py --handler bot.handler --runtime python2.7 --dependencies requirements.txt
```
 
Go to the Minio application and create a file in the `foobar` bucket we created in the [Minio example](../README.md). You will see the slack message in the channel configured.
