#!/bin/bash

if [ ! -d "$GOPATH/src/github.com/kubeless/kubeless" ]; then
    echo "Kubeless directory not found"
    exit 1
fi    

if [ ! -d "$GOPATH/src/github.com/kubeless/kubeless/ksonnet-lib" ]; then
    # Ksonnet-lib is required in the same folder than Kubeless
    git clone --depth=1 https://github.com/ksonnet/ksonnet-lib.git "$GOPATH/src/github.com/kubeless/kubeless/ksonnet-lib"
fi
export KUBECFG_JPATH="$GOPATH/src/github.com/kubeless/kubeless/ksonnet-lib"

dockerd > /dev/null 2>&1 &

cd "$GOPATH/src/github.com/kubeless/kubeless"

"$@"
