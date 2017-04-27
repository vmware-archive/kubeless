# Set up the demo environment

This demo needs at least Helm 2.3.0 

First, we need to pack the code that we can locate at https://github.com/bitnami/charts/tree/tika-server/incubator/tika-server (tika-server branch). Clone the repo, checkout to the tika-server branch a and execute helm package tika-server. This will create a tika-server-0.10.tgz file.

Fetch this file and create an empty dir somewhere. After put the tika-server-0.10.tgz inside this directory, execute helm server directory. The last command will cause that all chart packaged in that directory are being shared as an HTTP repo. We can use that local repo by executing  helm repo add local http://127.0.0.1:8879

To verify that all is OK :

$ helm search tika
NAME             	VERSION	DESCRIPTION
local/tika-server	0.1.0  	A Apache Tika Server Helm chart for Kubernetes

Everything is OK if you do a helm dep up and the dependent packages will be pushed to the charts directory in your helm package source code.


# Install Kubeless
Install latest Kubeless version from https://github.com/bitnami/kubeless/releases

 $ kubeless install 

Install helm and initializar Tiller into the cluster with: 
 $ helm init


Deploy the helm chart with: 
  $ helm install ./ocr-pipeline -n ocr

Execute `make all` to populate the function on kubeless
  $ make all 

Install minio-mc  and configure it as needed:
You can obtain the endpoint of  the service `ocr-minio-svc` by this command:
  $ MINIOENDPOINT=$(kubectl describe svc/ocr-minio-svc |awk '/Ingress/{ print $3}')

  $ mc config host add ocr http://$MINIOENDPOINT:9000   AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY  S3V4

  Where:
    `ocr`: is an alias for the server in the minio-mc tool
    `http://MINIOENDPOINT:9000`: Is the ocr-minio-svc external endpoint provided by deploying the helm chart
    `AKIA....EXAMPLE`: Example access key 
    `wJal...EXAMPLEKEY`: Example secret access key
    `S3V4`: API-signature version

Create a bucket called `ocr` and a subfolder called `input`: 
  $ mc mb ocr/input 

Finally, enable bucket event notification  for TXT files 
  $ mc events add ocr/input arn:minio:sqs:us-east-1:1:webhook --events put 

Fetch a sample PDF to feed the OCR system powered by Apache Tika. 

This sample file includes most common tags used in PDF files.

   http://www.tra.org.bh/media/document/sample10.pdf

#   This one makes a deep OCR recognition, since the text it's embedded into a JPEG image

#   https://courses.cs.vt.edu/csonline/AI/Lessons/VisualProcessing/OCRscans_files/bowers.jpg

To test the minio webhook plumbing we need to make a quick test by uploading a file:
  $ mc cp /tmp/sample10.pdf  ocr/input

 This must trigger an webhook on minio, calling to the K8s service named `slack` created by  `kubeless function create ...`  since in values.yaml we defined this: 

```
 webhook:
   enable: true
   endpoint: "http://minio-ocr:8080"
```

If all is correct, you must receive a JSON document like this:

```
{
   "Key" : "input/log.txt",
   "time" : "2017-04-25T13:22:59Z",
   "level" : "info",
   "msg" : "",
   "Records" : [
      {
         "awsRegion" : "us-east-1",
         "eventVersion" : "2.0",
         "eventName" : "s3:ObjectCreated:Put",
         "eventSource" : "aws:s3",
         "s3" : {
            "configurationId" : "Config",
            "bucket" : {
               "ownerIdentity" : {
                  "principalId" : "AKIAIOSFODNN7EXAMPLE"
               },
               "arn" : "arn:aws:s3:::input",
               "name" : "input"
            },
            "s3SchemaVersion" : "1.0",
            "object" : {
               "sequencer" : "14B8A6B679C32273",
               "eTag" : "6c01cd97cc315d6378dce9a2e6e61db9",
               "size" : 140,
               "key" : "log.txt"
            }
         },
         "eventTime" : "2017-04-25T13:22:59Z",
         "userIdentity" : {
            "principalId" : "AKIAIOSFODNN7EXAMPLE"
         },
         "responseElements" : {
            "x-amz-request-id" : "14B8A6B679C32273",
            "x-minio-origin-endpoint" : "http://10.92.1.2:9000"
         },
         "requestParameters" : {
            "sourceIPAddress" : "10.92.1.1:62337"
         }
      }
   ],
   "EventType" : "s3:ObjectCreated:Put"
}
```



