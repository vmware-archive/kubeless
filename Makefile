GO = go
GO_FLAGS =
GOFMT = gofmt
VERSION = dev-$(shell date +%FT%T%z)

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
	CGO_ENABLED=1 ./script/binary $(VERSION)

binary-cross:
	./script/binary-cli $(VERSION)


%.yaml: %.jsonnet
	$(KUBECFG) show -o yaml $< > $@.tmp
	mv $@.tmp $@

all-yaml: kubeless.yaml kubeless-rbac.yaml kubeless-openshift.yaml

kubeless.yaml: kubeless.jsonnet

kubeless-rbac.yaml: kubeless-rbac.jsonnet kubeless.jsonnet

kubeless-openshift.yaml: kubeless-openshift.jsonnet kubeless-rbac.jsonnet

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

integration-tests:
	./script/integration-tests

minikube-rbac-test:
	./script/integration-test-rbac minikube

fmt:
	$(GOFMT) -s -w $(GO_FILES)
