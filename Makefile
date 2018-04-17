GO = go
GO_FLAGS =
GOFMT = gofmt
KUBECFG = kubecfg
DOCKER = docker
CONTROLLER_IMAGE = kubeless-controller-manager:latest
FUNCTION_IMAGE_BUILDER = kubeless-function-image-builder:latest
KAFKA_CONTROLLER_IMAGE = kafka-trigger-controller:latest
NATS_CONTROLLER_IMAGE = nats-trigger-controller:latest
OS = linux
ARCH = amd64
BUNDLES = bundles
GO_PACKAGES = ./cmd/... ./pkg/...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

export KUBECFG_JPATH := $(CURDIR)/ksonnet-lib
export PATH := $(PATH):$(CURDIR)/bats/bin

.PHONY: all

KUBELESS_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \

default: binary

all:
	CGO_ENABLED=1 ./script/make.sh

binary:
	CGO_ENABLED=1 ./script/binary

binary-cross:
	./script/binary-cli


%.yaml: %.jsonnet
	$(KUBECFG) show -o yaml $< > $@.tmp
	mv $@.tmp $@

all-yaml: kubeless.yaml kubeless-non-rbac.yaml kubeless-openshift.yaml kafka-zookeeper.yaml kafka-zookeeper-openshift.yaml nats.yaml

kubeless.yaml: kubeless.jsonnet

kubeless-non-rbac.yaml: kubeless-non-rbac.jsonnet

kubeless-openshift.yaml: kubeless-openshift.jsonnet

kafka-zookeeper.yaml: kafka-zookeeper.jsonnet

nats.yaml: nats.jsonnet

docker/controller-manager: controller-build
	cp $(BUNDLES)/kubeless_$(OS)-$(ARCH)/kubeless-controller-manager $@

controller-build:
	./script/binary-controller -os=$(OS) -arch=$(ARCH)

controller-image: docker/controller-manager
	$(DOCKER) build -t $(CONTROLLER_IMAGE) $<

docker/function-image-builder: function-image-builder-build
	cp $(BUNDLES)/kubeless_$(OS)-$(ARCH)/imbuilder $@

function-image-builder-build:
	./script/binary-controller -os=$(OS) -arch=$(ARCH) imbuilder github.com/kubeless/kubeless/pkg/function-image-builder

function-image-builder: docker/function-image-builder
	$(DOCKER) build -t $(FUNCTION_IMAGE_BUILDER) $<

docker/kafka-controller: kafka-controller-build
	cp $(BUNDLES)/kubeless_$(OS)-$(ARCH)/kafka-controller $@

kafka-controller-build:
	./script/kafka-controller.sh -os=$(OS) -arch=$(ARCH)

kafka-controller-image: docker/kafka-controller
	$(DOCKER) build -t $(KAFKA_CONTROLLER_IMAGE) $<

nats-controller-build:
	./script/binary-controller -os=$(OS) -arch=$(ARCH) nats-controller github.com/kubeless/kubeless/cmd/nats-trigger-controller

nats-controller-image: docker/nats-controller
	$(DOCKER) build -t $(NATS_CONTROLLER_IMAGE) $<

docker/nats-controller: nats-controller-build
	cp $(BUNDLES)/kubeless_$(OS)-$(ARCH)/nats-controller $@

update:
	./hack/update-codegen.sh

test:
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

validation:
	./script/validate-vet
	./script/validate-lint
	./script/validate-gofmt
	./script/validate-git-marks

integration-tests:
	./script/integration-tests minikube deployment
	./script/integration-tests minikube basic

minikube-rbac-test:
	./script/integration-test-rbac minikube

fmt:
	$(GOFMT) -s -w $(GO_FILES)

bats:
	git clone --depth=1 https://github.com/sstephenson/bats.git

ksonnet-lib:
	git clone --depth=1 https://github.com/ksonnet/ksonnet-lib.git

.PHONY: bootstrap
bootstrap: bats ksonnet-lib

	go get github.com/mitchellh/gox
	go get github.com/golang/lint/golint

	@if ! which kubecfg >/dev/null; then \
	wget -q -O $$GOPATH/bin/kubecfg https://github.com/ksonnet/kubecfg/releases/download/v0.6.0/kubecfg-$$(go env GOOS)-$$(go env GOARCH); \
	chmod +x $$GOPATH/bin/kubecfg; \
	fi

	@if ! which kubectl >/dev/null; then \
	KUBECTL_VERSION=$$(wget -qO- https://storage.googleapis.com/kubernetes-release/release/stable.txt); \
	wget -q -O $$GOPATH/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$$KUBECTL_VERSION/bin/$$(go env GOOS)/$$(go env GOARCH)/kubectl; \
	chmod +x $$GOPATH/bin/kubectl; \
	fi

build_and_test:
	./script/start-test-environment.sh "make binary && make controller-image CONTROLLER_IMAGE=bitnami/kubeless-controller-manager:latest && make integration-tests"
