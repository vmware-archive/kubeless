get-python:
	kubeless function deploy get-python --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get-python-verify:
	kubeless function call get-python |egrep hello.world

get-nodejs:
	kubeless function deploy get-nodejs --trigger-http --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-nodejs/"

get-python-metadata:
	kubeless function deploy get-python-metadata --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --env foo:bar,bar=foo,foo --memory 128Mi --label foo:bar,bar=foo,foobar
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python-metadata/"

get: get-python get-nodejs get-python-metadata

get-nodejs-verify:
	kubeless function call get-nodejs |egrep hello.world

get-python-metadata-verify:
	kubeless function call get-python-metadata |egrep hello.world

post-python:
	kubeless function deploy post-python --trigger-http --runtime python2.7 --handler hellowithdata.handler --from-file python/hellowithdata.py
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post-python/ --header \"Content-Type:application/json\""

post-python-verify:
	kubeless function call post-python --data '{"it-s": "alive"}'|egrep "it.*alive"

post-nodejs:
	kubeless function deploy post-nodejs --trigger-http --runtime nodejs6 --handler hellowithdata.handler --from-file nodejs/hellowithdata.js
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post-nodejs/ --header \"Content-Type:application/json\""

post-nodejs-verify:
	kubeless function call post-nodejs --data '{"it-s": "alive"}'|egrep "it.*alive"

post: post-python post-nodejs

pubsub:
	kubeless topic create s3
	kubeless function deploy pubsub --trigger-topic s3 --runtime python2.7 --handler pubsub.handler --from-file python/pubsub.py

# Generate a random string to inject into s3 topic,
# then "tail -f" until it shows (with timeout)
pubsub-verify:
	$(eval DATA := $(shell mktemp -u -p entry -t XXXXXXXX))
	kubeless topic publish --topic s3 --data "$(DATA)"
	bash -c 'grep -q "$(DATA)" <(timeout 60 kubectl logs -f $$(kubectl get po -oname|grep pubsub))'
