#!/usr/bin/env bash

# Copyright (c) 2016-2017 Bitnami
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Special case: if ./ksonnet-lib exists, set KUBECFG_JPATH
test -d $PWD/ksonnet-lib && export KUBECFG_JPATH=$PWD/ksonnet-lib

# We require below env
: ${GOPATH:?} ${KUBECFG_JPATH:?}
export PATH=${PATH}:${GOPATH}/bin

# Default kubernetes context - if it's "dind" or "minikube" will
# try to bring up a local (dockerized) cluster
test -n "${TRAVIS_K8S_CONTEXT}" && set -- ${TRAVIS_K8S_CONTEXT}
# minikube seems to be more stable than dind, sp for kafka
INTEGRATION_TESTS_CTX=${1:-minikube}

# Check for some needed tools, install (some) if missing
which bats > /dev/null || {
   echo "ERROR: 'bats' is required to run these tests," \
        "install it from https://github.com/sstephenson/bats"
   exit 255
}

install_bin() {
    local exe=${1:?}
    test -n "${TRAVIS}" && sudo install -v ${exe} /usr/local/bin || install ${exe} ${GOPATH:?}/bin
}

which kubectl || {
    KUBECTL_VERSION=$(wget -qO- https://storage.googleapis.com/kubernetes-release/release/stable.txt)
    wget --no-clobber \
        https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
    install_bin ./kubectl
}

# Start a k8s cluster (minikube, dind) if not running
kubectl get nodes --context=${INTEGRATION_TESTS_CTX:?} || {
    cluster_up=./script/cluster-up-${INTEGRATION_TESTS_CTX}.sh
    test -f ${cluster_up} || {
        echo "FATAL: bringing up k8s cluster '${INTEGRATION_TESTS_CTX}' not supported"
        exit 255
    }
    ${cluster_up}
}

# Both RBAC'd dind and minikube seem to be missing rules to make kube-dns work properly
# add some (granted) broad ones:
kubectl --context=${INTEGRATION_TESTS_CTX:?} get clusterrolebinding kube-dns-admin >& /dev/null || \
    kubectl --context=${INTEGRATION_TESTS_CTX:?} create clusterrolebinding kube-dns-admin --serviceaccount=kube-system:default --clusterrole=cluster-admin

k8s_context_save