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

# From minikube howto
export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir -p ~/.kube
touch ~/.kube/config

export KUBECONFIG=$HOME/.kube/config
export PATH=${PATH}:${GOPATH:?}/bin

MINIKUBE_VERSION=v0.22.3

install_bin() {
    local exe=${1:?}
    test -n "${TRAVIS}" && sudo install -v ${exe} /usr/local/bin || install ${exe} ${GOPATH:?}/bin
}

# Travis ubuntu trusty env doesn't have nsenter, needed for VM-less minikube
# (--vm-driver=none, runs dockerized)
check_or_build_nsenter() {
    which nsenter >/dev/null && return 0
    echo "INFO: Building 'nsenter' ..."
cat <<-EOF | docker run -i --rm -v "$(pwd):/build" ubuntu:14.04 >& nsenter.build.log
        apt-get update
        apt-get install -qy git bison build-essential autopoint libtool automake autoconf gettext pkg-config
        git clone --depth 1 git://git.kernel.org/pub/scm/utils/util-linux/util-linux.git /tmp/util-linux
        cd /tmp/util-linux
        ./autogen.sh
        ./configure --without-python --disable-all-programs --enable-nsenter
        make nsenter
        cp -pfv nsenter /build
EOF
    if [ ! -f ./nsenter ]; then
        echo "ERROR: nsenter build failed, log:"
        cat nsenter.build.log
        return 1
    fi
    echo "INFO: nsenter build OK, installing ..."
    install_bin ./nsenter
}
check_or_install_minikube() {
    which minikube || {
        wget --no-clobber -O minikube \
            https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64
        install_bin ./minikube
    }
}

# Install nsenter if missing
check_or_build_nsenter
# Install minikube if missing
check_or_install_minikube
MINIKUBE_BIN=$(which minikube)

# Start minikube
sudo -E ${MINIKUBE_BIN} start --vm-driver=none \
    --extra-config=apiserver.Authorization.Mode=RBAC

# Wait til settles
echo "INFO: Waiting for minikube cluster to be ready ..."
typeset -i cnt=120
until kubectl --context=minikube get pods >& /dev/null; do
    ((cnt=cnt-1)) || exit 1
    sleep 1
done
exit 0
# vim: sw=4 ts=4 et si
