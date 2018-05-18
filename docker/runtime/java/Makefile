init-image:
	docker build -f Dockerfile.init -t kubeless/java-init:1.8 .

runtime-image:
	docker build -f Dockerfile -t kubeless/java:1.8 .

push-init:
	docker push kubeless/java-init:1.8

push-runtime:
	docker push kubeless/java:1.8

build-all: init-image runtime-image
push-all: push-init push-runtime