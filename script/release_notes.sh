#!/bin/bash
set -e

REPO_NAME=kubeless
REPO_DOMAIN=kubeless

function check_tag {
  local tag=$1
  published_tags=`curl -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME/tags`
  already_published=`echo $published_tags | jq ".[] | select(.name == \"$tag\")"`
  echo $already_published
}

function commit_list {
  local tag=$1
  git fetch --tags
  local last_tag=`curl -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME/tags | jq --raw-output '.[0].name'`
  local release_notes=`git log $last_tag..HEAD --oneline`
  local parsed_release_notes=$(echo "$release_notes" | sed -n -e 'H;${x;s/\n/\\n- /g;s/^\\n//;p;}')
  echo $parsed_release_notes
}

if [[ -z "$REPO_NAME" || -z "$REPO_DOMAIN" ]]; then
  echo "Github repository not specified" > /dev/stderr
  exit 1
fi

repo_check=`curl -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME`
if [[ $repo_check == *"Not Found"* ]]; then
  echo "Not found a Github repository for $REPO_DOMAIN/$REPO_NAME, it is not possible to publish it" > /dev/stderr
  exit 1
else
  tag=$1
  already_published=`check_tag $tag`
  if [[ -z $already_published ]]; then
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
EOF)
    echo -e "${notes}"
  else
    echo "Unable to produce relase notes since $tag was already released"
    exit 1
  fi
fi
