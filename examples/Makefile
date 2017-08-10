get-python:
	kubeless function deploy get-python --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get-python-verify:
	kubeless function call get-python |egrep hello.world

get-nodejs:
	kubeless function deploy get-nodejs --trigger-http --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-nodejs/"

get-python-metadata:
	kubeless function deploy get-python --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --env foo:bar,bar=foo,foo --mem 128Mi --label --label foo:bar,bar=foo,foobar
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get: get-python get-nodejs get-python-metadata

get-nodejs-verify:
	kubeless function call get-nodejs |egrep hello.world

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

topic:
	kubeless function publish --topic demo --data "s3"
