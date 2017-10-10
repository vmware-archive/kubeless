#!/bin/bash
set -e

REPO_NAME=kubeless
REPO_DOMAIN=kubeless

function commit_list {
  local tag=$1
  git fetch --tags
  local previous_tag=`curl -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME/tags | jq --raw-output '.[1].name'`
  local release_notes=`git log $previous_tag..$tag --oneline`
  local parsed_release_notes=$(echo "$release_notes" | sed -n -e 'H;${x;s/\n/\\n- /g;s/^\\n//;p;}')
  echo $parsed_release_notes
}

function get_release_notes {
  commits=`commit_list $tag`
  notes=$(cat << EOF
This release includes the following commits and features:\n
$commits\n
To install this latest version, use the manifest that is part of the release:

**NO RBAC:**

\`\`\`console
kubectl create ns kubeless
curl -sL https://github.com/kubeless/kubeless/releases/download/$tag/kubeless-$tag.yaml | kubectl create -f -
\`\`\`

**WITH RBAC ENABLED:**

\`\`\`console
kubectl create ns kubeless
curl -sL https://github.com/kubeless/kubeless/releases/download/$tag/kubeless-rbac-$tag.yaml | kubectl create -f -
\`\`\`
EOF )
  echo -e "${notes}"
}

function release_tag {
  local tag=$1
  local get_release_notes=$(get_release_notes $tag)
  release=`curl -H "Authorization: token $ACCESS_TOKEN" --request PATCH -s --data "{
    \"tag_name\": \"$tag\",
    \"target_commitish\": \"master\",
    \"name\": \"$tag\",
    \"body\": \"Release $tag includes the following commits: \n$release_notes\",
    \"draft\": true,
    \"prerelease\": false
  }" https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME/releases`
  echo $release | jq ".id"
}

if [[ -z "$REPO_NAME" || -z "$REPO_DOMAIN" ]]; then
  echo "Github repository not specified" > /dev/stderr
  exit 1
fi

if [[ -z "$ACCESS_TOKEN" ]]; then
  echo "Unable to release: Github Token not specified" > /dev/stderr
  exit 1
fi

repo_check=`curl -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME`
if [[ $repo_check == *"Not Found"* ]]; then
  echo "Not found a Github repository for $REPO_DOMAIN/$REPO_NAME, it is not possible to publish it" > /dev/stderr
  exit 1
else
  release_tag $1
fi
