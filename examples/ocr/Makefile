minio-ocr: 
	kubeless function create minio-ocr --trigger-http --runtime python27 --handler minio-ocr.handler --from-file minio-ocr.py --dependencies requirements.txt

clean: 
	kubeless function delete minio-ocr 

all: clean minio-ocr

