#!/bin/bash
set -e
SCRIPT=$0
if [ -h $SCRIPT ]; then
    SCRIPT=`readlink $SCRIPT`
fi
ROOTDIR=`cd $(dirname $SCRIPT)/.. && pwd`

COMMAND="${@:-bash}"

if ! minikube version | grep -q "v0.22.3"; then
  # Until #399 is fixed
  echo "Only minikube v0.22.3 is supported"
  exit 1
fi

if ! minikube status | grep -q "minikube: $"; then
  echo "Unable to start the test environment with an existing instance of minikube"
  echo "Delete the current profile executing 'minikube delete' or create a new one"
  echo "executing 'minikube profile new_profile'"
  exit 1
fi

minikube start --extra-config=apiserver.Authorization.Mode=RBAC
eval $(minikube docker-env)

docker run --privileged -it \
  -v $ROOTDIR:/go/src/github.com/kubeless/kubeless \
  -v $HOME/.kube:/root/.kube \
  -v $HOME/.minikube:$HOME/.minikube \
  kubeless/dev-environment:latest bash -c "$COMMAND"
