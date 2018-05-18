#!/bin/bash
set -e

REPO_NAME=kubeless
REPO_DOMAIN=kubeless

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
  update_release_tag $1 $REPO_DOMAIN $REPO_DOMAIN
fi
