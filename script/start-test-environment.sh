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
  echo "Delete the current profile executing 'minikube delete' or create a new one"
  echo "executing 'minikube profile new_profile'"
  exit 1
fi

minikube start --extra-config=apiserver.Authorization.Mode=RBAC --insecure-registry 0.0.0.0/0
eval $(minikube docker-env)

CONTEXT=$(kubectl config current-context)

# Both RBAC'd dind and minikube seem to be missing rules to make kube-dns work properly
# add some (granted) broad ones:
kubectl --context=${CONTEXT} get clusterrolebinding kube-dns-admin >& /dev/null || \
    kubectl --context=${CONTEXT} create clusterrolebinding kube-dns-admin --serviceaccount=kube-system:default --clusterrole=cluster-admin

docker run --privileged -it \
  -v $ROOTDIR:/go/src/github.com/kubeless/kubeless \
  -v $HOME/.kube:/root/.kube \
  -v $HOME/.minikube:$HOME/.minikube \
  -e TEST_CONTEXT=$(kubectl config current-context) \
  -e TEST_DEBUG=1 \
  kubeless/dev-environment:latest bash -c "$COMMAND"
