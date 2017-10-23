import json
import base64
import tempfile

#pip install slackclient
from slackclient import SlackClient

#pip install kubernetes
from kubernetes import client, config

config.load_incluster_config()

v1=client.CoreV1Api()

#Get slack secret
for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'slack':
        token = base64.b64decode(secrets.data['token'])

sc = SlackClient(token)

def handler(context):
    if context['EventType'] == "s3:ObjectCreated:Put":
        bucket = context['Key'].split('/')[0]
        filename = context['Key'].split('/')[1]

        msg = "An object called %s was uploaded to bucket %s" % (filename,bucket)

        r = sc.api_call(
                    "chat.postMessage",
                    channel="#bot",
                    text=msg
                   )
        if r.get('ok'):
            return "Notification successfully sent to Slack"
        else:
            return "Error while sending notification to Slack: " + r.get('error')

