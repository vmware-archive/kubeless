get:
	kubeless function create get --trigger-http --runtime python27 --handler helloget.foo --from-file helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get/"

post:
	kubeless function create post --trigger-http --runtime python27 --handler hellowithdata.handler --from-file hellowithdata.py
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post/ --header "Content-Type:application/json""

pubsub:
	kubeless function create pubsub --trigger-topic s3 --runtime python27 --handler pubsub.handler --from-file pubsub.py

topic:
	kubeless function publish --topic demo --data "s3"

