#!/bin/bash
set -e

# TODO: Use kubeless
REPO_NAME=kubeless
REPO_DOMAIN=kubeless
TAG=${1:?}

PROJECT_DIR=$(cd $(dirname $0)/.. && pwd)

source $(dirname $0)/release_utils.sh

if [[ -z "$REPO_NAME" || -z "$REPO_DOMAIN" ]]; then
  echo "Github repository not specified" > /dev/stderr
  exit 1
fi

if [[ -z "$ACCESS_TOKEN" ]]; then
  echo "Unable to release: Github Token not specified" > /dev/stderr
  exit 1
fi

repo_check=`curl -H "Authorization: token $ACCESS_TOKEN" -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME`
if [[ $repo_check == *"Not Found"* ]]; then
  echo "Not found a Github repository for $REPO_DOMAIN/$REPO_NAME, it is not possible to publish it" > /dev/stderr
  exit 1
else
  RELEASE_ID=$(release_tag $1 $REPO_DOMAIN $REPO_NAME | jq '.id')
fi

manifests=(
  kubeless kubeless-non-rbac kubeless-openshift
  kafka-zookeeper kafka-zookeeper-openshift
  nats
  )
for f in "${manifests[@]}"; do
  cp ${PROJECT_DIR}/${f}.yaml ${PROJECT_DIR}/${f}-${TAG}.yaml
  upload_asset $REPO_DOMAIN $REPO_NAME "$RELEASE_ID" "${PROJECT_DIR}/${f}-${TAG}.yaml"
done
for f in `ls ${PROJECT_DIR}/bundles/kubeless_*.zip`; do
  upload_asset $REPO_DOMAIN $REPO_NAME $RELEASE_ID $f
done
