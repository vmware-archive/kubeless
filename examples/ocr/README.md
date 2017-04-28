# Set up the demo environment

This demo needs at least Helm 2.3.0  and Kubeless 0.0.11

This demo shows how to interconnect services with kubeless functions, all the architecture is packaged with Helm, 
a Kubernetes package manager which provide us 

# Install Kubeless
Install latest Kubeless version from https://github.com/bitnami/kubeless/releases


Execute `make setup` to install kubeless and helm Tiller into your minikube cluster
  $ make setup

Wait a couple of minutes while the pods starts

Execute `make install` to install the `ocr=pipeline` helm chart and associated dependencies
 $ make install

 Quickly edit the deployment `ocr-ocr-pipeline` with kubectl edit deployment ocr-ocr-pipeline 
and remove this lines.  Save the document and quit to apply changes.   By this method we're
removing health checks for this pod, since OCR is a time consuming process. This is a workaround
and need to be fixed ASAP. 

        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 3
          periodSeconds: 3
          successThreshold: 1
          timeoutSeconds: 1

Wait a couple of minutes while the pods starts. Minio pod tends to crash until our function is ready
so it can crash 5 o 6 times.

Execute `make install` to install the `ocr=pipeline` helm chart and associated dependencies
 $ make test 

It fetchs a sample PDF to feed the OCR system powered by Apache Tika. 

This sample file includes most common tags used in PDF files.

   http://www.tra.org.bh/media/document/sample10.pdf


Wait a few seconds while Tika parses the document and execute:

POD=`kubectl get pods -l app=ocr-mongodb  | awk '/ocr-mongo/{print $1}'`
kubectl exec -it $POD --  mongo --eval "printjson(db.processed.findOne())" localhost/ocr



