#!/bin/bash

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

# This script is a wrapper around https://github.com/projectatomic/skopeo
# and /layer_builder.js that pulls an image and push a new one including a tar file
# as a new layer.
# Usage example:
#   $ image_builder.sh docker://node:6 docker://andresmgot/my-node:6 /my_files.tar user:pass user:pass 
# TODO: Subsitute for a go binary with proper flags

set -e

# Kubernetes ImagePullSecrets uses .dockerconfigjson as the file name
# for storing credentials but skopeo requires it to be named config.json
if [ -f $DOCKER_CONFIG_FOLDER/.dockerconfigjson ]; then
    echo "Creating $HOME/.docker/config.json"
    mkdir -p $HOME/.docker
    ln -s $DOCKER_CONFIG_FOLDER/.dockerconfigjson $HOME/.docker/config.json
fi

skopeoCopy() {
    local command="skopeo copy"
    if [ ! -z $src_credentials ]; then
        command+=" --src-creds '$src_credentials'"
    fi
    if [ ! -z $dst_credentials ]; then
        command+=" --dest-creds '$dst_credentials'"
    fi
    command+=" ${@}"
    /bin/sh -c "$command"
}

if [ "$#" -lt 3 ]; then
  echo >&2 "USAGE: $0 src_image dst_image new_layer.tar src[user:password] dst[user:password] [work_dir]"
  exit 1
fi

src_image=${1:?}
dst_image=${2:?}
new_layer=${3:?}
src_credentials=${4-""}
dst_credentials=${5-""}
work_dir=${6-""}

if [ -z $work_dir ]; then
    work_dir=$(mktemp -u)
fi
mkdir -p $work_dir

skopeoCopy $src_image dir:$work_dir

node /layer_builder.js $work_dir $new_layer

skopeoCopy dir:$work_dir $dst_image
