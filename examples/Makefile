get-python:
	kubeless function deploy get-python --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get-python-verify:
	kubeless function call get-python |egrep hello.world

get-python-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo():\n%4sreturn "hello world updated"\n' > $(TMPDIR)/hello-updated.py
	kubeless function update get-python --from-file $(TMPDIR)/hello-updated.py
	rm -rf $(TMPDIR)
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get-python-update-verify:
	kubeless function call get-python |egrep hello.world.updated

get-python-deps:
	kubeless function deploy get-python-deps --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/hellowithdeps.py --dependencies python/requirements.txt

get-python-deps-verify:
	kubeless function call get-python-deps |egrep Google

get-python-custom-port:
	kubeless function deploy get-python-custom-port --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --port 8081

get-python-custom-port-verify:
	kubeless function call get-python-custom-port |egrep hello.world

get-python-deps-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'bs4\ntwitter\n' > $(TMPDIR)/requirements.txt
	kubeless function update get-python-deps --dependencies $(TMPDIR)/requirements.txt
	rm -rf $(TMPDIR)

get-python-deps-update-verify:
	$(eval pod := $(shell kubectl get pod -l function=get-python-deps -o go-template -o custom-columns=:metadata.name --no-headers=true))
	kubectl exec -it $(pod) pip freeze | grep -q "twitter=="

get-python-34:
	kubeless function deploy get-python --trigger-http --runtime python3.4 --handler helloget.foo --from-file python/helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python/"

get-python-34-verify:
	kubeless function call get-python |egrep hello.world

get-python-36:
	kubeless function deploy get-python-36 --trigger-http --runtime python3.6 --handler helloget.foo --from-file python/helloget.py
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python-36/"

get-python-36-verify:
	kubeless function call get-python-36 |egrep hello.world

scheduled-get-python:
	kubeless function deploy scheduled-get-python --schedule "* * * * *" --runtime python2.7 --handler helloget.foo --from-file python/helloget.py

scheduled-get-python-verify:
	number="1"; \
	timeout="70"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=scheduled-get-python`; \
		logs=`kubectl logs $$pod | grep "GET / HTTP/1.1\" 200 11 \"\""`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

timeout-python:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo():\n%4swhile 1: pass\n%4sreturn "hello world"\n' > $(TMPDIR)/hello-loop.py
	kubeless function deploy timeout-python --trigger-http --runtime python2.7 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.py --timeout 3
	rm -rf $(TMPDIR)

timeout-python-verify:
	$(eval MSG := $(shell kubeless function call timeout-python 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

get-nodejs:
	kubeless function deploy get-nodejs --trigger-http --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-nodejs/"

get-nodejs-verify:
	kubeless function call get-nodejs |egrep hello.world

get-nodejs-custom-port:
	kubeless function deploy get-nodejs-custom-port --trigger-http --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js --port 8083
	echo "curl localhost:8083/api/v1/proxy/namespaces/default/services/get-nodejs-custom-port/"

get-nodejs-custom-port-verify:
	kubeless function call get-nodejs-custom-port |egrep hello.world

timeout-nodejs:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'module.exports = { foo: function (req, res) { while(true) {} } }\n' > $(TMPDIR)/hello-loop.js
	kubeless function deploy timeout-nodejs --trigger-http --runtime nodejs6 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.js --timeout 4
	rm -rf $(TMPDIR)

timeout-nodejs-verify:
	$(eval MSG := $(shell kubeless function call timeout-nodejs 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

get-nodejs-deps:
	kubeless function deploy get-nodejs-deps --trigger-http --runtime nodejs6 --handler helloget.handler --from-file nodejs/hellowithdeps.js --dependencies nodejs/package.json

get-nodejs-deps-verify:
	kubeless function call get-nodejs-deps --data '{"hello": "world"}' | grep -q '"hello":"world","date"'

get-nodejs-multi:
	cd nodejs; zip helloFunctions.zip *js
	kubeless function deploy get-nodejs-multi --trigger-http --runtime nodejs6 --handler index.helloGet --from-file nodejs/helloFunctions.zip
	rm nodejs/helloFunctions.zip

get-nodejs-multi-verify:
	kubeless function call get-nodejs-multi |egrep hello.world

get-python-metadata:
	kubeless function deploy get-python-metadata --trigger-http --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --env foo:bar,bar=foo,foo --memory 128Mi --label foo:bar,bar=foo,foobar
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-python-metadata/"

get-python-metadata-verify:
	kubeless function call get-python-metadata |egrep hello.world

get-ruby:
	kubeless function deploy get-ruby --trigger-http --runtime ruby2.4 --handler helloget.foo --from-file ruby/helloget.rb
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-ruby/"

get-ruby-verify:
	kubeless function call get-ruby |egrep hello.world

get-ruby-deps:
	kubeless function deploy get-ruby-deps --trigger-http --runtime ruby2.4 --handler hellowithdeps.foo --from-file ruby/hellowithdeps.rb --dependencies ruby/Gemfile
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-ruby-deps/"

get-ruby-deps-verify:
	kubeless function call get-ruby-deps |egrep hello.world

get-ruby-custom-port:
	kubeless function deploy get-ruby-custom-port --trigger-http --runtime ruby2.4 --handler helloget.foo --from-file ruby/helloget.rb --port 8082
	echo "curl localhost:8082/api/v1/proxy/namespaces/default/services/get-ruby-custom-port/"

get-ruby-custom-port-verify:
	kubeless function call get-ruby-custom-port |egrep hello.world

timeout-ruby:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo(c)\n%4swhile true do;sleep(1);end\n%4s"hello world"\nend' > $(TMPDIR)/hello-loop.rb
	kubeless function deploy timeout-ruby --trigger-http --runtime ruby2.4 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.rb --timeout 4
	rm -rf $(TMPDIR)

timeout-ruby-verify:
	$(eval MSG := $(shell { time kubeless function call timeout-ruby; } 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded
	echo $(MSG) | egrep "real\s*0m4."

get-dotnetcore:
	kubeless function deploy get-dotnetcore --trigger-http --runtime dotnetcore2.0 --handler helloget.foo --from-file dotnetcore/helloget.cs
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/get-dotnetcore/"

get-dotnetcore-verify:
	kubeless function call get-dotnetcore |egrep hello.world

custom-get-python:
	kubeless function deploy --runtime-image kubeless/get-python-example@sha256:a922942597ce617adbe808b9f0cdc3cf7ff987a1277adf0233dd47be5d85082a custom-get-python

custom-get-python-verify:
	kubeless function call custom-get-python |egrep hello.world

custom-get-python-update:
	kubeless function update --runtime-image kubeless/get-python-example@sha256:d2ca4ab086564afbac6d30c29614f1623ddb9163b818537742c42fd785fcf2ce custom-get-python

custom-get-python-update-verify:
	kubeless function call custom-get-python |egrep hello.world.updated

get: get-python get-nodejs get-python-metadata get-ruby get-ruby-deps get-python-custom-port

post-python:
	kubeless function deploy post-python --trigger-http --runtime python2.7 --handler hellowithdata.handler --from-file python/hellowithdata.py
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post-python/ --header \"Content-Type:application/json\""

post-python-verify:
	kubeless function call post-python --data '{"it-s": "alive"}'|egrep "it.*alive"

post-python-custom-port:
	kubeless function deploy post-python-custom-port --trigger-http --runtime python2.7 --handler hellowithdata.handler --from-file python/hellowithdata.py --port 8081
	echo "curl --data '{\"hello\":\"world\"}' localhost:8081/api/v1/proxy/namespaces/default/services/post-python-custom-port/ --header \"Content-Type:application/json\""

post-python-custom-port-verify:
	kubeless function call post-python-custom-port --data '{"it-s": "alive"}'|egrep "it.*alive"

post-nodejs:
	kubeless function deploy post-nodejs --trigger-http --runtime nodejs6 --handler hellowithdata.handler --from-file nodejs/hellowithdata.js
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post-nodejs/ --header \"Content-Type:application/json\""

post-nodejs-verify:
	kubeless function call post-nodejs --data '{"it-s": "alive"}'|egrep "it.*alive"

post-ruby:
	kubeless function deploy post-ruby --trigger-http --runtime ruby2.4 --handler hellowithdata.handler --from-file ruby/hellowithdata.rb
	echo "curl --data '{\"hello\":\"world\"}' localhost:8080/api/v1/proxy/namespaces/default/services/post-ruby/ --header \"Content-Type:application/json\""

post-ruby-verify:
	kubeless function call post-ruby --data '{"it-s": "alive"}'|egrep "it.*alive"

post-dotnetcore:
	kubeless function deploy post-dotnetcore --runtime dotnetcore2.0 --handler hellowithdata.handler --from-file dotnetcore/hellowithdata.cs --trigger-http

post-dotnetcore-verify:
	kubeless function call post-dotnetcore --data '{"it-s": "alive"}'|egrep "it.*alive"

post: post-python post-nodejs post-ruby post-python-custom-port

pubsub-python:
	kubeless topic create s3-python || true
	kubeless function deploy pubsub-python --trigger-topic s3-python --runtime python2.7 --handler pubsub.handler --from-file python/pubsub.py

# Generate a random string to inject into s3 topic,
# then "tail -f" until it shows (with timeout)
pubsub-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data "$(DATA)"
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-python34:
	kubeless topic create s3-python34 || true
	kubeless function deploy pubsub-python34 --trigger-topic s3-python34 --runtime python3.4 --handler pubsub-python.handler --from-file python/pubsub.py

pubsub-python34-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python34 --data "$(DATA)"
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python34`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-python36:
	kubeless topic create s3-python36 || true
	kubeless function deploy pubsub-python36 --trigger-topic s3-python36 --runtime python3.6 --handler pubsub-python.handler --from-file python/pubsub.py

pubsub-python36-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python36 --data "$(DATA)"
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python36`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-nodejs:
	kubeless topic create s3-nodejs || true
	kubeless function deploy pubsub-nodejs --trigger-topic s3-nodejs --runtime nodejs6 --handler pubsub-nodejs.handler --from-file nodejs/helloevent.js

pubsub-nodejs-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-nodejs --data '{"test": "$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-nodejs`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-nodejs-update:
	kubeless topic create s3-nodejs-2 || true
	kubeless function update pubsub-nodejs --trigger-topic s3-nodejs-2

pubsub-nodejs-update-verify:
	kubectl describe $$(kubectl get po -oname|grep pubsub-nodejs) | grep -e "TOPIC_NAME:\s*s3-nodejs-2"

pubsub-ruby:
	kubeless topic create s3-ruby || true
	kubeless function deploy pubsub-ruby --trigger-topic s3-ruby --runtime ruby2.4 --handler pubsub-ruby.handler --from-file ruby/helloevent.rb

pubsub-ruby-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-ruby --data "$(DATA)"
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-ruby`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

post: pubsub-python pubsub-nodejs pubsub-ruby
