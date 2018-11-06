get-python:
	kubeless function deploy get-python --runtime python2.7 --handler helloget.foo --from-file python/helloget.py

get-python-verify:
	kubeless function call get-python |egrep hello.world
	kubeless function top --function get-python --out yaml |egrep total_calls.*[1-100000]

get-python-update:
	$(eval TMPDIR := $(shell mktemp -d))
	printf 'def foo(event, context):\n%4sreturn "hello world updated"\n' > $(TMPDIR)/hello-updated.py
	kubeless function update get-python --from-file $(TMPDIR)/hello-updated.py
	rm -rf $(TMPDIR)

get-python-update-verify:
	kubeless function call get-python |egrep hello.world.updated

get-python-deps:
	cd python && zip hellowithdeps.zip hellowithdeps.py  hellowithdepshelper.py && cd ..
	kubeless function deploy get-python-deps --runtime python2.7 --handler hellowithdeps.foo --from-file python/hellowithdeps.zip --dependencies python/requirements.txt

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

get-python-url-deps:
	kubeless function deploy get-python-url-deps --runtime python2.7 --handler helloget.foo --from-file https://raw.githubusercontent.com/kubeless/kubeless/v1.0.0-alpha.1/examples/python/hellowithdeps.py --dependencies https://raw.githubusercontent.com/kubeless/kubeless/v1.0.0-alpha.1/examples/python/requirements.txt

get-python-url-deps-verify:
	kubeless function call get-python-url-deps |egrep Google	

get-node-url-zip:
	kubeless function deploy get-node-url-zip --runtime nodejs6 --handler index.helloGet --from-file https://github.com/kubeless/kubeless/blob/master/examples/nodejs/helloFunctions.zip?raw=true

get-node-url-zip-verify:
	kubeless function call get-node-url-zip |egrep hello.world

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

get-nodejs-stream:
	kubeless function deploy get-nodejs-stream --runtime nodejs6 --handler hellostream.foo --from-file nodejs/hellostream.js --dependencies nodejs/package.json

get-nodejs-stream-verify:
	kubeless function call get-nodejs-stream |egrep hello.world

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
	echo $$logs | grep -q "Z" && \
	echo $$logs | grep -q "application/json" && \
	echo $$logs | grep -q "cli.kubeless.io"

get-python-metadata:
	kubeless function deploy get-python-metadata --runtime python2.7 --handler helloget.foo --from-file python/helloget.py --env foo:bar,bar=foo --memory 128Mi --label foo:bar,bar=foo,foobar

get-python-metadata-verify:
	kubeless function call get-python-metadata |egrep hello.world
	kubectl get po -o jsonpath='{.items[0].spec.containers[0].env}' -l function=get-python-metadata | grep "name:foo value:bar"
	kubectl get po -o jsonpath='{.items[0].spec.containers[0].env}' -l function=get-python-metadata | grep "name:bar value:foo"

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
	kubeless function deploy get-dotnetcore --runtime dotnetcore2.0 --handler module.handler --from-file dotnetcore/helloget.cs

get-dotnetcore-verify:
	kubeless function call get-dotnetcore |egrep hello.world
	kubeless function top --function get-dotnetcore --out yaml |egrep "Function does not expose metrics" 

get-dotnetcore-dependency:
	kubeless function deploy get-dotnetcore-dependency --runtime dotnetcore2.0 --handler module.handler --from-file dotnetcore/dependency-yaml.cs --dependencies dotnetcore/dependency-yaml.csproj

get-dotnetcore-dependency-verify:
	kubeless function call get-dotnetcore-dependency |egrep Name:\ Michael

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
	echo $$logs | grep -q "event-time.*Z" && \
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
	echo $$logs | grep -q "event-time.*Z" && \
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
	echo $$logs | grep -q "event-time.*Z" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*cli.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

post-dotnetcore:
	kubeless function deploy post-dotnetcore --runtime dotnetcore2.0 --handler module.handler --from-file dotnetcore/hellowithdata.cs

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

python-kinesis:
	kubeless function deploy python-kinesis --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py
	$(eval NODEPORT := $(shell kubectl get svc kinesis -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubectl create secret generic ec2 --from-literal=aws_access_key_id=kinesalite --from-literal=aws_secret_access_key=kinesalite
	kubeless trigger kinesis create-stream --aws-region kinesalite --secret ec2 --endpoint http://$(MINIKUBE_IP):$(NODEPORT) --shard-count 1 --stream-name kubeless-stream
	kubeless trigger kinesis create kinesis-trigger --function-name python-kinesis --aws-region kinesalite --shard-id shardId-000000000000 --stream kubeless-stream --secret ec2 --endpoint http://$(MINIKUBE_IP):$(NODEPORT)

python-kinesis-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc kinesis -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger kinesis publish --aws-region kinesalite --secret ec2  --endpoint http://$(MINIKUBE_IP):$(NODEPORT) --partition-key key1 --stream  kubeless-stream --records '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=python-kinesis`; \
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
	logs=`kubectl logs -l function=python-kinesis`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*kinesistriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

python-kinesis-multi-record:
	kubeless function deploy python-kinesis-multi-record --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py
	$(eval NODEPORT := $(shell kubectl get svc kinesis -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger kinesis create kinesis-trigger-mr --function-name python-kinesis-multi-record --aws-region kinesalite --shard-id shardId-000000000000 --stream kubeless-stream --secret ec2 --endpoint http://$(MINIKUBE_IP):$(NODEPORT)

python-kinesis-multi-record-verify:
	$(eval DATA1 := $(shell mktemp -u -t XXXXXXXX))
	$(eval DATA2 := $(shell mktemp -u -t XXXXXXXX))
	$(eval NODEPORT := $(shell kubectl get svc kinesis -o jsonpath="{.spec.ports[0].nodePort}"))
	$(eval MINIKUBE_IP := $(shell minikube ip))
	kubeless trigger kinesis publish --aws-region kinesalite --secret ec2  --endpoint http://$(MINIKUBE_IP):$(NODEPORT) --partition-key key1 --stream  kubeless-stream --records '{"payload":"$(DATA1)"}' --records '{"payload":"$(DATA2)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=python-kinesis-multi-record`; \
		logs1=`kubectl logs $$pod | grep $(DATA1)`; \
		logs2=`kubectl logs $$pod | grep $(DATA2)`; \
		if [ "$$logs1" != "" ] && [ "$$logs2" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=python-kinesis-multi-record`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*kinesistriggers.kubeless.io" && \
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

get-java:
	kubeless function deploy get-java --runtime java1.8 --handler Foo.foo --from-file java/HelloGet.java

get-java-verify:
	kubeless function call get-java |egrep Hello.world

post-java:
	kubeless function deploy post-java --runtime java1.8  --handler Foo.foo --from-file java/HelloWithData.java

post-java-verify:
	kubeless function call post-java --data '{"its": "alive"}'| egrep "its.*alive"

get-java-deps:
	kubeless function deploy get-java-deps --runtime java1.8 --handler Hello.sayHello --from-file java/HelloWithDeps.java --dependencies java/pom.xml

get-java-deps-verify:
	kubeless function call get-java-deps --data '{"hello": "world"}'
	kubectl logs -l function=get-java-deps | grep -q '.*Hello.*world! Current local time is:'

get-jvm-java:
	kubeless function deploy get-jvm-java --runtime jvm1.8 --from-file jvm/java/test-java-jvm.jar --handler io_ino_Handler.sayHello

get-jvm-java-verify:
	kubeless function call get-jvm-java | grep "Hello world"

get-nodejs-distroless:
	kubeless function deploy get-nodejs-distroless --runtime nodejs_distroless8 --handler helloget.foo --from-file nodejs/helloget.js

get-nodejs-distroless-verify:
	kubeless function call get-nodejs-distroless |egrep hello.world

get-nodejs-distroless-deps:
	kubeless function deploy get-nodejs-distroless-deps --runtime nodejs_distroless8 --handler helloget.handler --from-file nodejs/hellowithdeps.js --dependencies nodejs/package.json

get-nodejs-distroless-deps-verify:
	kubeless function call get-nodejs-distroless-deps --data '{"hello": "world"}' | grep -q 'hello.*world.*date.*UTC'

get-ballerina:
	kubeless function deploy get-ballerina --runtime ballerina0.981.0 --from-file ballerina/helloget.bal --handler helloget.foo

get-ballerina-verify:
	kubeless function call get-ballerina |egrep Hello.World.Ballerina

get-ballerina-custom-port:
	kubeless function deploy get-ballerina-custom-port --runtime ballerina0.981.0 --handler helloget.foo --from-file ballerina/helloget.bal --port 8083

get-ballerina-custom-port-verify:
	kubectl get svc get-ballerina-custom-port -o yaml | grep 'targetPort: 8083'
	kubeless function call get-ballerina-custom-port |egrep Hello.World.Ballerina

get-ballerina-data:
	kubeless function deploy get-ballerina-data --runtime ballerina0.981.0 --from-file ballerina/hellowithdata.bal --handler hellowithdata.foo

get-ballerina-data-verify:
	kubeless function call get-ballerina-data --data '{"hello":"world"}' |egrep hello

get-ballerina-conf:
	zip -r -j ballerina/bar.zip ballerina/hello_with_conf/
	kubeless function deploy get-ballerina-conf --runtime ballerina0.981.0 --from-file ballerina/bar.zip --handler hello_with_conf.bar
	rm ballerina/bar.zip

get-ballerina-conf-verify:
	kubeless function call get-ballerina-conf | egrep john
