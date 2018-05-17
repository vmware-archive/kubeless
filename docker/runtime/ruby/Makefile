
build2.4:
	env GOOS=linux GOARCH=amd64 go build $$GOPATH/src/github.com/kubeless/kubeless/pkg/function-proxy/proxy.go
	docker build -t kubeless/ruby:2.4$$RUNTIME_TAG_MODIFIER -f Dockerfile .

push2.4:
	docker push kubeless/ruby:2.4$$RUNTIME_TAG_MODIFIER

# Mandatory jobs
build-all: build2.4
push-all: push2.4