import base64
import tempfile

#pip install kubernetes
from kubernetes import client, config

# pip install minio
from minio import Minio
from minio.error import ResponseError

config.load_incluster_config()

v1=client.CoreV1Api()

for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'minio':
        access_key = base64.b64decode(secrets.data['access_key'])
        secret_key = base64.b64decode(secrets.data['secret_key'])

# Replace the DNS below with the minio service name (helm release name -svc)
client = Minio('invited-crab-minio-svc:9000',
                  access_key=access_key,
                  secret_key=secret_key,
                  secure=False)

def ocr(context):
    if context['EventType'] == "s3:ObjectCreated:Put" and context['Key'].split('/')[0] == 'foobar':

        tf = tempfile.NamedTemporaryFile(delete=False)
        bucket = context['Key'].split('/')[0]
        filename = context['Key'].split('/')[1]

        try:
            client.fget_object(bucket, filename, tf.name)
        except ResponseError as err:
            print err
        

        # puts the same file in a OCR bucket
        ocr_name = filename + '.ocr'
        try:
            client.fput_object('ocr',ocr_name,tf.name)
        except ResponseError as err:
            print err

    else:
        print "Minio file deletion event"

    return "OCR called"
