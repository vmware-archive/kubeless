minio-ocr: 
	kubeless function create minio-ocr --trigger-http --runtime python27 --handler minio-ocr.handler --from-file minio-ocr.py --dependencies requirements.txt

clean: 
	kubeless function delete minio-ocr 

setup:
	yes | kubeless install
	helm init ; sleep 30
	helm ls 
	helm install -n ocr ./ocr-pipeline --replace
	make all

all: clean minio-ocr

