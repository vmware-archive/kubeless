
build2.7:
	docker build -t kubeless/python:2.7$$RUNTIME_TAG_MODIFIER -f Dockerfile.2.7 .

build3.4:
	docker build -t kubeless/python:3.4$$RUNTIME_TAG_MODIFIER -f Dockerfile.3.4 .

build3.6:
	docker build -t kubeless/python:3.6$$RUNTIME_TAG_MODIFIER -f Dockerfile.3.6 .

push2.7:
	docker push kubeless/python:2.7$$RUNTIME_TAG_MODIFIER

push3.4:
	docker push kubeless/python:3.4$$RUNTIME_TAG_MODIFIER

push3.6:
	docker push kubeless/python:3.6$$RUNTIME_TAG_MODIFIER

# Mandatory jobs
build-all: build2.7 build3.4 build3.6
push-all: push2.7 push3.4 push3.6