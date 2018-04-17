get-python:
	kubeless function deploy get-python --runtime python2.7 --handler helloget.foo --from-file python/helloget.py

get-python-verify:
	kubeless function call get-python |egrep hello.world

get-python-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo(event, context):\n%4sreturn "hello world updated"\n' > $(TMPDIR)/hello-updated.py
	kubeless function update get-python --from-file $(TMPDIR)/hello-updated.py
	rm -rf $(TMPDIR)

get-python-update-verify:
	kubeless function call get-python |egrep hello.world.updated

get-python-deps:
	kubeless function deploy get-python-deps --runtime python2.7 --handler helloget.foo --from-file python/hellowithdeps.py --dependencies python/requirements.txt

get-python-deps-verify:
	kubeless function call get-python-deps |egrep Google

get-python-custom-port:
	kubeless function deploy get-python-custom-port --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --port 8081

get-python-custom-port-verify:
	kubectl get svc get-python-custom-port -o yaml | grep 'targetPort: 8081'
	kubeless function call get-python-custom-port |egrep hello.world

get-python-deps-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'bs4\ntwitter\n' > $(TMPDIR)/requirements.txt
	kubeless function update get-python-deps --dependencies $(TMPDIR)/requirements.txt
	rm -rf $(TMPDIR)

get-python-deps-update-verify:
	pod=`kubectl get pod -l function=get-python-deps -o go-template -o custom-columns=:metadata.name --no-headers=true`; \
	echo "Checking updated deps of $$pod"; \
	kubectl exec -it $$pod pip freeze | grep -q "twitter=="

get-python-34:
	kubeless function deploy get-python --runtime python3.4 --handler helloget.foo --from-file python/helloget.py

get-python-34-verify:
	kubeless function call get-python |egrep hello.world

get-python-36:
	kubeless function deploy get-python-36 --runtime python3.6 --handler helloget.foo --from-file python/helloget.py

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
	printf 'def foo(event, context):\n%4swhile 1: pass\n%4sreturn "hello world"\n' > $(TMPDIR)/hello-loop.py
	kubeless function deploy timeout-python --runtime python2.7 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.py --timeout 3
	rm -rf $(TMPDIR)

timeout-python-verify:
	$(eval MSG := $(shell kubeless function call timeout-python 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

get-nodejs:
	kubeless function deploy get-nodejs --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js

get-nodejs-verify:
	kubeless function call get-nodejs |egrep hello.world

get-nodejs-custom-port:
	kubeless function deploy get-nodejs-custom-port --runtime nodejs6 --handler helloget.foo --from-file nodejs/helloget.js --port 8083

get-nodejs-custom-port-verify:
	kubectl get svc get-nodejs-custom-port -o yaml | grep 'targetPort: 8083'
	kubeless function call get-nodejs-custom-port |egrep hello.world

timeout-nodejs:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'module.exports = { foo: function (event, context) { while(true) {} } }\n' > $(TMPDIR)/hello-loop.js
	kubeless function deploy timeout-nodejs --runtime nodejs6 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.js --timeout 4
	rm -rf $(TMPDIR)

timeout-nodejs-verify:
	$(eval MSG := $(shell kubeless function call timeout-nodejs 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

get-nodejs-deps:
	kubeless function deploy get-nodejs-deps --runtime nodejs6 --handler helloget.handler --from-file nodejs/hellowithdeps.js --dependencies nodejs/package.json

get-nodejs-deps-verify:
	kubeless function call get-nodejs-deps --data '{"hello": "world"}' | grep -q 'hello.*world.*date.*UTC'

get-nodejs-multi:
	cd nodejs; zip helloFunctions.zip *js
	kubeless function deploy get-nodejs-multi --runtime nodejs6 --handler index.helloGet --from-file nodejs/helloFunctions.zip
	rm nodejs/helloFunctions.zip

get-nodejs-multi-verify:
	kubeless function call get-nodejs-multi |egrep hello.world

get-go:
	kubeless function deploy get-go --runtime go1.10 --handler handler.Foo --from-file golang/helloget.go

get-go-verify:
	kubeless function call get-go |egrep Hello.world

get-go-custom-port:
	kubeless function deploy get-go-custom-port --runtime go1.10 --handler helloget.Foo --from-file golang/helloget.go --port 8083

get-go-custom-port-verify:
	kubectl get svc get-go-custom-port -o yaml | grep 'targetPort: 8083'
	kubeless function call get-go-custom-port |egrep Hello.world

timeout-go:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'package kubeless\nimport "github.com/kubeless/kubeless/pkg/functions"\nfunc Foo(event functions.Event, context functions.Context) (string, error) {\nfor{\n}\nreturn "", nil\n}' > $(TMPDIR)/hello-loop.js
	kubeless function deploy timeout-go --runtime go1.10 --handler helloget.Foo  --from-file $(TMPDIR)/hello-loop.js --timeout 4
	rm -rf $(TMPDIR)

timeout-go-verify:
	$(eval MSG := $(shell kubeless function call timeout-go 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

get-go-deps:
	kubeless function deploy get-go-deps --runtime go1.10 --handler helloget.Hello --from-file golang/hellowithdeps.go --dependencies golang/Gopkg.toml

get-go-deps-verify:
	kubeless function call get-go-deps --data '{"hello": "world"}'
	kubectl logs -l function=get-go-deps | grep -q 'level=info msg=.*hello.*world'

post-go:
	kubeless function deploy post-go --runtime go1.10 --handler hellowithdata.Handler --from-file golang/hellowithdata.go

post-go-verify:
	kubeless function call post-go --data '{"it-s": "alive"}'| egrep "it.*alive"
	# Verify event context
	logs=`kubectl logs -l function=post-go`; \
	echo $$logs | grep -q "it.*alive" && \
	echo $$logs | grep -q "UTC" && \
	echo $$logs | grep -q "application/json" && \
	echo $$logs | grep -q "cli.kubeless.io"

get-python-metadata:
	kubeless function deploy get-python-metadata --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --env foo:bar,bar=foo,foo --memory 128Mi --label foo:bar,bar=foo,foobar

get-python-metadata-verify:
	kubeless function call get-python-metadata |egrep hello.world

get-python-secrets:
	kubectl create secret generic test-secret --from-literal=key=MY_KEY || true
	kubeless function deploy get-python-secrets --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --secrets test-secret

get-python-secrets-verify:
	$(eval pod := $(shell kubectl get pod -l function=get-python-secrets -o go-template -o custom-columns=:metadata.name --no-headers=true))
	kubectl exec -it $(pod) cat /test-secret/key | egrep "MY_KEY"

get-ruby:
	kubeless function deploy get-ruby --runtime ruby2.4 --handler helloget.foo --from-file ruby/helloget.rb

get-ruby-verify:
	kubeless function call get-ruby |egrep hello.world

get-ruby-deps:
	kubeless function deploy get-ruby-deps --runtime ruby2.4 --handler hellowithdeps.foo --from-file ruby/hellowithdeps.rb --dependencies ruby/Gemfile

get-ruby-deps-verify:
	kubeless function call get-ruby-deps |egrep hello.world

get-ruby-custom-port:
	kubeless function deploy get-ruby-custom-port --runtime ruby2.4 --handler helloget.foo --from-file ruby/helloget.rb --port 8082

get-ruby-custom-port-verify:
	kubectl get svc get-ruby-custom-port -o yaml | grep 'targetPort: 8082'
	kubeless function call get-ruby-custom-port |egrep hello.world

get-php:
	kubeless function deploy get-php --runtime php7.2 --handler helloget.foo --from-file php/helloget.php

get-php-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf '<?php\n function foo() { return "hello world updated"; } \n' > $(TMPDIR)/hello-updated.php
	kubeless function update get-php --from-file $(TMPDIR)/hello-updated.php
	rm -rf $(TMPDIR)

get-php-update-verify:
	kubeless function call get-php | egrep "hello.world.updated"

get-php-verify:
	kubeless function call get-php | egrep "hello world"

get-php-deps:
	kubeless function deploy get-php-deps --runtime php7.2 --handler hellowithdeps.foo --from-file php/hellowithdeps.php --dependencies php/composer.json

get-php-deps-verify:
	kubeless function call get-php-deps &> /dev/null
	kubectl logs -l function=get-php-deps | egrep "Hello"

get-php-deps-update:
	$(eval TMPDIR := $(shell mktemp -d))
	sed "s/1\.23/1\.20/" php/composer.json > $(TMPDIR)/composer.json
	kubeless function update get-php-deps --dependencies $(TMPDIR)/composer.json

get-php-deps-update-verify:
	$(eval pod := $(shell kubectl get pod -l function=get-php-deps -o go-template -o custom-columns=:metadata.name --no-headers=true))
	kubectl exec -it $(pod) cat /kubeless/composer.json | egrep "1.20"

post-php:
	kubeless function deploy post-php --runtime php7.2 --handler hellowithdata.foo --from-file php/hellowithdata.php

post-php-verify:
	kubeless function call post-php --data '{"it-s": "alive"}'| egrep "it.*alive"

timeout-php:
	$(eval TMPDIR := $(shell mktemp -d))
	printf '<?php\n function foo() { while(1) {} } \n' > $(TMPDIR)/hello-loop.php
	kubeless function deploy timeout-php --runtime php7.2 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.php --timeout 4
	rm -rf $(TMPDIR)

timeout-php-verify:
	$(eval MSG := $(shell kubeless function call timeout-php 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded

timeout-ruby:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo(event, context)\n%4swhile true do;sleep(1);end\n%4s"hello world"\nend' > $(TMPDIR)/hello-loop.rb
	kubeless function deploy timeout-ruby --runtime ruby2.4 --handler helloget.foo  --from-file $(TMPDIR)/hello-loop.rb --timeout 4
	rm -rf $(TMPDIR)

timeout-ruby-verify:
	$(eval MSG := $(shell { time kubeless function call timeout-ruby; } 2>&1 || true))
	echo $(MSG) | egrep Request.timeout.exceeded
	echo $(MSG) | egrep "real\s*0m4."

get-dotnetcore:
	kubeless function deploy get-dotnetcore --runtime dotnetcore2.0 --handler helloget.foo --from-file dotnetcore/helloget.cs

get-dotnetcore-verify:
	kubeless function call get-dotnetcore |egrep hello.world

custom-get-python:
	kubeless function deploy --runtime-image kubeless/get-python-example@sha256:6a14400f14e26d46a971445b7a850af533fe40cb75a67297283bdf536e09ca5e custom-get-python

custom-get-python-verify:
	kubeless function call custom-get-python |egrep hello.world

custom-get-python-update:
	kubeless function update --runtime-image kubeless/get-python-example@sha256:174beab98e6fa454e21121302395375e90a324e9276367296aab0eb5b4aa8922 custom-get-python

custom-get-python-update-verify:
	kubeless function call custom-get-python |egrep hello.world.updated

get: get-python get-nodejs get-python-metadata get-ruby get-ruby-deps get-python-custom-port

post-python:
	kubeless function deploy post-python --runtime python2.7 --handler hellowithdata.handler --from-file python/hellowithdata.py

post-python-verify:
	kubeless function call post-python --data '{"it-s": "alive"}'|egrep "it.*alive"
	# Verify event context
	logs=`kubectl logs -l function=post-python`; \
	echo $$logs | grep -q "it.*alive" && \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*cli.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

post-python-custom-port:
	kubeless function deploy post-python-custom-port --runtime python2.7 --handler hellowithdata.handler --from-file python/hellowithdata.py --port 8081

post-python-custom-port-verify:
	kubectl get svc post-python-custom-port -o yaml | grep 'targetPort: 8081'
	kubeless function call post-python-custom-port --data '{"it-s": "alive"}'|egrep "it.*alive"

post-nodejs:
	kubeless function deploy post-nodejs --runtime nodejs6 --handler hellowithdata.handler --from-file nodejs/hellowithdata.js

post-nodejs-verify:
	kubeless function call post-nodejs --data '{"it-s": "alive"}'|egrep "it.*alive"
	# Verify event context
	logs=`kubectl logs -l function=post-nodejs`; \
	echo $$logs | grep -q "it.*alive" && \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*cli.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

post-ruby:
	kubeless function deploy post-ruby --runtime ruby2.4 --handler hellowithdata.handler --from-file ruby/hellowithdata.rb

post-ruby-verify:
	kubeless function call post-ruby --data '{"it-s": "alive"}'|egrep "it.*alive"
	# Verify event context
	logs=`kubectl logs -l function=post-ruby`; \
	echo $$logs | grep -q "it.*alive" && \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*cli.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

post-dotnetcore:
	kubeless function deploy post-dotnetcore --runtime dotnetcore2.0 --handler hellowithdata.handler --from-file dotnetcore/hellowithdata.cs

post-dotnetcore-verify:
	kubeless function call post-dotnetcore --data '{"it-s": "alive"}'|egrep "it.*alive"

post: post-python post-nodejs post-ruby post-python-custom-port

pubsub-python:
	kubeless topic create s3-python || true
	kubeless function deploy pubsub-python  --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py
	kubeless trigger kafka create pubsub-python --function-selector created-by=kubeless,function=pubsub-python --trigger-topic s3-python

# Generate a random string to inject into s3 topic,
# then "tail -f" until it shows (with timeout)
pubsub-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
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
	# Verify event context
	logs=`kubectl logs -l function=pubsub-python`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*kafkatriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

python-nats:
	kubeless function deploy python-nats --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py
	kubeless trigger nats create python-nats --function-selector created-by=kubeless,function=python-nats --trigger-topic test

python-nats-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc nats -n nats-io -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger nats publish --url nats://$(MINIKUBE_IP):$(NODEPORT) --topic test --message '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=python-nats`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=python-nats`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*natstriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

nats-python-func1-topic-test:
	kubeless function deploy nats-python-func1-topic-test --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py  --label topic=nats-topic-test

nats-python-func2-topic-test:
	kubeless function deploy nats-python-func2-topic-test --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py  --label topic=nats-topic-test

nats-python-func-multi-topic:
	kubeless function deploy nats-python-func-multi-topic --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py  --label func=nats-python-func-multi-topic

nats-python-trigger-topic-test:
	kubeless trigger nats create nats-python-trigger-topic-test --function-selector created-by=kubeless,topic=nats-topic-test --trigger-topic topic-test

nats-python-trigger-topic1:
	kubeless trigger nats create nats-python-trigger-topic1 --function-selector created-by=kubeless,func=nats-python-func-multi-topic --trigger-topic topic1

nats-python-trigger-topic2:
	kubeless trigger nats create nats-python-trigger-topic2 --function-selector created-by=kubeless,func=nats-python-func-multi-topic --trigger-topic topic2

nats-python-func1-topic-test-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc nats -n nats-io -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger nats publish --url nats://$(MINIKUBE_IP):$(NODEPORT) --topic topic-test --message '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=nats-python-func1-topic-test`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=nats-python-func1-topic-test`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*natstriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

nats-python-func2-topic-test-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc nats -n nats-io -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger nats publish --url nats://$(MINIKUBE_IP):$(NODEPORT) --topic topic-test --message '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=nats-python-func2-topic-test`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=nats-python-func2-topic-test`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*natstriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

nats-python-func-multi-topic-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc nats -n nats-io -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger nats publish --url nats://$(MINIKUBE_IP):$(NODEPORT) --topic topic1 --message '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=nats-python-func-multi-topic`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=nats-python-func-multi-topic`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*natstriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

	kubeless trigger nats publish --url nats://$(MINIKUBE_IP):$(NODEPORT) --topic topic2 --message '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=nats-python-func-multi-topic`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=nats-python-func-multi-topic`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*natstriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

kafka-python-func1-topic-s3-python:
	kubeless topic create s3-python || true
	kubeless function deploy kafka-python-func1-topic-s3-python --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py --label topic=s3-python

kafka-python-func1-topic-s3-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=kafka-python-func1-topic-s3-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

kafka-python-func2-topic-s3-python:
	kubeless topic create s3-python || true
	kubeless function deploy kafka-python-func2-topic-s3-python --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py --label topic=s3-python

kafka-python-func2-topic-s3-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=kafka-python-func2-topic-s3-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

s3-python-kafka-trigger:
	kubeless trigger kafka create s3-python-kafka-trigger --function-selector created-by=kubeless,topic=s3-python --trigger-topic s3-python

pubsub-python34:
	kubeless topic create s3-python34 || true
	kubeless function deploy pubsub-python34 --runtime python3.4 --handler pubsub-python.handler --from-file python/hellowithdata34.py
	kubeless trigger kafka create pubsub-python34 --function-selector created-by=kubeless,function=pubsub-python34 --trigger-topic s3-python34

pubsub-python34-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python34 --data '{"payload":"$(DATA)"}'
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
	kubeless function deploy pubsub-python36 --runtime python3.6 --handler pubsub-python.handler --from-file python/pubsub.py
	kubeless trigger kafka create pubsub-python36 --function-selector created-by=kubeless,function=pubsub-python36 --trigger-topic s3-python36

pubsub-python36-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python36 --data '{"payload":"$(DATA)"}'
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
	kubeless function deploy pubsub-nodejs --runtime nodejs6 --handler pubsub-nodejs.handler --from-file nodejs/hellowithdata.js
	kubeless trigger kafka create pubsub-nodejs --function-selector created-by=kubeless,function=pubsub-nodejs --trigger-topic s3-nodejs

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
	kubeless trigger kafka update pubsub-nodejs --trigger-topic s3-nodejs-2

pubsub-nodejs-update-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-nodejs-2  --data '{"test": "$(DATA)"}'
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

pubsub-ruby:
	kubeless topic create s3-ruby || true
	kubeless function deploy pubsub-ruby --runtime ruby2.4 --handler pubsub-ruby.handler --from-file ruby/hellowithdata.rb
	kubeless trigger kafka create pubsub-ruby --function-selector created-by=kubeless,function=pubsub-ruby --trigger-topic s3-ruby

pubsub-ruby-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-ruby --data '{"payload":"$(DATA)"}'
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

pubsub-go:
	kubeless topic create s3-go || true
	kubeless function deploy pubsub-go --runtime go1.10 --handler pubsub-go.Handler --from-file golang/hellowithdata.go
	kubeless trigger kafka create pubsub-go --function-selector created-by=kubeless,function=pubsub-go --trigger-topic s3-go

pubsub-go-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-go --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-go`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found


pubsub: pubsub-python pubsub-nodejs pubsub-ruby
