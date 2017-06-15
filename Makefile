KUBECFG = kubecfg
DOCKER = docker
CONTROLLER_IMAGE = kubeless-controller
BUNDLES = bundles/kubeless_linux-amd64/

.PHONY: all

KUBELESS_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \

BIND_DIR := bundles

default: binary

all:
	CGO_ENABLED=1 ./script/make.sh

binary:
	CGO_ENABLED=1 ./script/make.sh binary

binary-cross:
	./script/make.sh binary-cross

kubeless.yaml: kubeless.jsonnet
	$(KUBECFG) show -o yaml $< > $@

docker/controller: controller-build
	cp $(BUNDLES)/$(CONTROLLER_IMAGE) $@

controller-build:
	./script/binary-cross -os=linux -arch=amd64

controller-image: docker/controller
	$(DOCKER) build -t $(CONTROLLER_IMAGE):latest $<