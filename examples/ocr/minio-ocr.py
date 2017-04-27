import base64
import tempfile

#pip install kubernetes
from kubernetes import client, config

# pip install requests
import requests

# pip install minio
from minio import Minio
from minio.error import ResponseError

config.load_incluster_config()

v1=client.CoreV1Api()

for secrets in v1.list_secret_for_all_namespaces().items:
    if secrets.metadata.name == 'ocr-minio-user':
        access_key =  base64.b64decode(secrets.data['accesskey'])
        secret_key =  base64.b64decode(secrets.data['secretkey'])

client = Minio('ocr-minio-svc:9000', 
                  access_key=access_key,
                  secret_key=secret_key, 
                  secure=False)


#events = client.listen_bucket_notification('ocr', 'input/', '.pdf', ['s3:ObjectCreated:*', 's3:ObjectRemoved:*', 's3:ObjectAccessed:*']) 

def handler(context):
    if context['EventType'] == "s3:ObjectCreated:Put": 

        tf = tempfile.NamedTemporaryFile(delete=False)
        bucket = context['Key'].split('/')[0]
        filename = context['Key'].split('/')[1]

        try:
            client.fget_object(bucket, filename, tf.name)
        except ResponseError as err:
            print err


        try:
            files = {'file': open(tf.name, 'rb')}
            r = requests.post('http://ocr-tika-server/tika', files=files )
            if r.stauts_code == '200':
               ocr_output = r.text
        except ResponseError as err:
            print err
        

        # puts the same file in a done OCR bucket
        ocr_name = filename + '.ocr'
        file = open(ocr_name,'w')
        file.write(ocr_output)
        file.close()

        try:
            client.fput_object('done',ocr_name,tf.name)
        except ResponseError as err:
            print err

    else:
        print "Minio file deletion event"

    return "OCR called"
