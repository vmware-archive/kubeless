import base64
import twitter

from kubernetes import client, config

config.load_incluster_config()

v1=client.CoreV1Api()

for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'twitter':
        consumer_key = base64.b64decode(secrets.data['consumer_key'])
        consumer_secret = base64.b64decode(secrets.data['consumer_secret'])
        token_key = base64.b64decode(secrets.data['token_key'])
        token_secret = base64.b64decode(secrets.data['token_secret'])

api = twitter.Api(consumer_key=consumer_key,
                  consumer_secret=consumer_secret,
                  access_token_key=token_key,
                  access_token_secret=token_secret)

def tweet(context):
    msg = context.json
    status = api.PostUpdate(msg['tweet'])



