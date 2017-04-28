install: 
	
	helm install -n ocr ./ocr-pipeline 
	@echo "IMPORTANT: "
	@echo "Remove livenessProve section  from kubectl edit deployment ocr-ocr-pipeline"

clean: 
	-helm delete ocr --purge 
	-kubectl delete deployment ocr-ocr-pipeline   
	-kubectl delete svc ocr-ocr-pipeline   
	-kubectl delete configmap ocr-ocr-pipeline   
	-kubectl get pods --namespace=kubeless| awk '/kubeless-controller/{print $$1}' | xargs kubectl delete  --namespace=kubeless pod

setup:
	yes | kubeless install
	helm init 


test:   
	mc config host add ocr $(shell minikube service ocr-minio-svc --url)  AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
	-mc rm ocr/input 
	-mc mb ocr/input 
	-mc mb ocr/done
	-mc events add ocr/input arn:minio:sqs:us-east-1:1:webhook --events put
	curl -O http://www.tra.org.bh/media/document/sample10.pdf
	mc cp sample10.pdf  ocr/input
	rm sample10.pdf
	@echo "Wait a minute and then execute: kubectl exec -it ocr-mongodb-XXXX-XXXXX  --  mongo --eval \"printjson(db.processed.findOne())\" localhost/ocr"
	@echo "Find the mongodb pod name with: kubectl get pods -l app=ocr-mongodb  | awk '/ocr-mongo/{print \$$1}'"


all: clean install

