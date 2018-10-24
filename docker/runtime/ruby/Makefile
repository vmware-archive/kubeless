
build2.3:
	env GOOS=linux GOARCH=amd64 go build $$GOPATH/src/github.com/kubeless/kubeless/pkg/function-proxy/proxy.go
	docker build -t kubeless/ruby:2.3$$RUNTIME_TAG_MODIFIER -f Dockerfile.2.3 .

build2.4:
	env GOOS=linux GOARCH=amd64 go build $$GOPATH/src/github.com/kubeless/kubeless/pkg/function-proxy/proxy.go
	docker build -t kubeless/ruby:2.4$$RUNTIME_TAG_MODIFIER -f Dockerfile.2.4 .

build2.5:
	env GOOS=linux GOARCH=amd64 go build $$GOPATH/src/github.com/kubeless/kubeless/pkg/function-proxy/proxy.go
	docker build -t kubeless/ruby:2.5$$RUNTIME_TAG_MODIFIER -f Dockerfile.2.5 .

push2.3:
	docker push kubeless/ruby:2.3$$RUNTIME_TAG_MODIFIER

push2.4:
	docker push kubeless/ruby:2.4$$RUNTIME_TAG_MODIFIER

push2.5:
	docker push kubeless/ruby:2.5$$RUNTIME_TAG_MODIFIER

# Mandatory jobs
build-all: build2.3 build2.4 build2.5
push-all: push2.3 push2.4 push2.5
