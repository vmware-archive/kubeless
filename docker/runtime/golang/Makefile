init-image:
	docker build -f Dockerfile.init -t kubeless/go-init:1.10 ../../../

runtime-image:
	docker build -f Dockerfile -t kubeless/go:1.10 .

push-init:
	docker push kubeless/go-init:1.10

push-runtime:
	docker push kubeless/go:1.10

# Mandatory jobs
build-all: init-image runtime-image
push-all: push-init push-runtime