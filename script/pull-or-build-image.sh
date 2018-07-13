#!/bin/bash

set -e

TARGET=${1:?}

function push() {
    local image=${1:?}
    if [[ -n "$DOCKER_USERNAME" && -n "$DOCKER_PASSWORD" ]]; then
        docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD" 
        docker push $image
    fi
}

case "${TARGET}" in
    "function-controller")
      image=${CONTROLLER_IMAGE:?}
      docker pull $image || make $TARGET CONTROLLER_IMAGE=$image
      push $image
      ;;
    "function-image-builder")
      image=${FUNCTION_IMAGE_BUILDER:?}
      docker pull $image || make $TARGET FUNCTION_IMAGE_BUILDER=$image
      push $image
      ;;
    "default")
      echo "Unsupported target"
      exit 1
esac
