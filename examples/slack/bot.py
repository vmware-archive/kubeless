import json

from slackclient import SlackClient
from kubernetes import client, config

config.load_incluster_config()

v1=client.CoreV1Api()

for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'slack':
        token = base64.b64decode(secrets.data['token'])

sc = SlackClient(token)

def handler(context):
    return sc.api_call(
                       "chat.postMessage",
                       channel="#bot",
                       text=context.json['msg']
                      )
