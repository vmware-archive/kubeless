.PHONY: build-all
init-image:
	docker build -f Dockerfile.init -t kubeless/jvm-init:1.8 .

runtime-image:
	docker build -f Dockerfile -t kubeless/jvm:1.8 .

push-init:
	docker push kubeless/jvm-init:1.8

push-runtime:
	docker push kubeless/jvm:1.8

build-all: init-image runtime-image
push-all: push-init push-runtime
