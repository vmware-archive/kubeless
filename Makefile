GO = go
GO_FLAGS =
GOFMT = gofmt

KUBECFG = kubecfg
DOCKER = docker
CONTROLLER_IMAGE = kubeless-controller:latest
OS = linux
ARCH = amd64
BUNDLES = bundles
GO_PACKAGES = ./cmd/... ./pkg/... ./version/...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

.PHONY: all

KUBELESS_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \

default: binary

all:
	CGO_ENABLED=1 ./script/make.sh

binary:
	CGO_ENABLED=1 ./script/make.sh binary

binary-cross:
	./script/make.sh binary-cli

kubeless.yaml: kubeless.jsonnet
	$(KUBECFG) show -o yaml $< > $@

docker/controller: controller-build
	cp $(BUNDLES)/kubeless_$(OS)-$(ARCH)/kubeless-controller $@

controller-build:
	./script/binary-controller -os=$(OS) -arch=$(ARCH)

controller-image: docker/controller
	$(DOCKER) build -t $(CONTROLLER_IMAGE) $<

test:
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

validation:
	./script/validate-vet
	./script/validate-lint
	./script/validate-gofmt
	./script/validate-git-marks

fmt:
	$(GOFMT) -s -w $(GO_FILES)
