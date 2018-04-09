
build6:
	docker build -t kubeless/nodejs:6$$RUNTIME_TAG_MODIFIER -f Dockerfile.6 .

build8:
	docker build -t kubeless/nodejs:8$$RUNTIME_TAG_MODIFIER -f Dockerfile.8 .

push6:
	docker push kubeless/nodejs:6$$RUNTIME_TAG_MODIFIER

push8:
	docker push kubeless/nodejs:8$$RUNTIME_TAG_MODIFIER

# Mandatory jobs
build-all: build6 build8
push-all: push6 push8