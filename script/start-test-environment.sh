#!/bin/bash
set -e
SCRIPT=$0
if [ -h $SCRIPT ]; then
    SCRIPT=`readlink $SCRIPT`
fi
ROOTDIR=`cd $(dirname $SCRIPT)/.. && pwd`

COMMAND="${@:-bash}"

if ! minikube status | grep -q "minikube: $"; then
    echo "Unable to start the test environment with an existing instance of minikube"
    exit 1
fi

minikube start --extra-config=apiserver.Authorization.Mode=RBAC
eval $(minikube docker-env)

docker run --privileged -it \
  -v $ROOTDIR:/go/src/github.com/kubeless/kubeless \
  -v $HOME/.kube:/root/.kube \
  -v $HOME/.minikube:$HOME/.minikube \
  kubeless/dev-environment:latest bash -c "$COMMAND"
