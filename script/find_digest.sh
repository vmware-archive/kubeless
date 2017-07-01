#!/bin/bash

REPOSITORY=$1
TARGET_TAG=$2

# get authorization token
TOKEN=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$REPOSITORY:pull" | jq -r .token)

# find all tags
ALL_TAGS=$(curl -s -H "Authorization: Bearer $TOKEN" https://index.docker.io/v2/$REPOSITORY/tags/list | jq -r .tags[])

# get image digest for target
TARGET_DIGEST=$(curl -s -D - -H "Authorization: Bearer $TOKEN" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" https://index.docker.io/v2/$REPOSITORY/manifests/$TARGET_TAG | grep Docker-Content-Digest | cut -d ' ' -f 2)

# for each tags
for tag in ${ALL_TAGS[@]}; do
  # get image digest
  digest=$(curl -s -D - -H "Authorization: Bearer $TOKEN" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" https://index.docker.io/v2/$REPOSITORY/manifests/$tag | grep Docker-Content-Digest | cut -d ' ' -f 2)

  # check digest
  if [[ $TARGET_DIGEST = $digest ]]; then
    echo "$tag $digest"
  fi
done
