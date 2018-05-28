#!/bin/bash
set -e

function commit_list {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  git fetch --tags
  local previous_tag=`curl -H "Authorization: token $ACCESS_TOKEN" -s https://api.github.com/repos/$repo_domain/$repo_name/tags | jq --raw-output '.[1].name'`
  local release_notes=`git log $previous_tag..$tag --oneline`
  local parsed_release_notes=$(echo "$release_notes" | sed -n -e 'H;${x;s/\n/\\n- /g;s/^\\n//;s/"/\\"/g;p;}')
  echo $parsed_release_notes
}

function get_release_notes {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  commits=`commit_list $tag $repo_domain $repo_name`
  notes=$(echo "\
This release includes the following commits and features:\\n\
$commits\\n\\n\
To install this latest version, use the manifest that is part of the release:\\n\
\\n\
**WITH RBAC ENABLED:**\\n\
\\n\
\`\`\`console\\n\
kubectl create ns kubeless\\n\
kubectl create -f https://github.com/kubeless/kubeless/releases/download/$tag/kubeless-$tag.yaml \\n\
\`\`\`\\n\
\\n\
**WITHOUT RBAC:**\\n\
\\n\
\`\`\`console\\n\
kubectl create ns kubeless\\n\
kubectl create -f https://github.com/kubeless/kubeless/releases/download/$tag/kubeless-non-rbac-$tag.yaml \\n\
\`\`\`\\n\
**OPENSHIFT:**\\n\
\\n\
\`\`\`console\\n\
oc create ns kubeless\\n\
oc create -f https://github.com/kubeless/kubeless/releases/download/$tag/kubeless-openshift-$tag.yaml \\n\
# Kafka\\n\
oc create -f https://github.com/kubeless/kubeless/releases/download/$tag/kafka-zookeeper-openshift-$tag.yaml \\n\
\`\`\`\\n\
")
  echo "${notes}"
}

function get_release_body {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local release_notes=$(get_release_notes $tag $repo_domain $repo_name)
  echo '{
    "tag_name": "'$tag'",
    "target_commitish": "master",
    "name": "'$tag'",
    "body": "'$release_notes'",
    "draft": true,
    "prerelease": false
  }'
}

function update_release_tag {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local release_id=$(curl -H "Authorization: token $ACCESS_TOKEN" -s https://api.github.com/repos/$repo_domain/$repo_name/releases | jq  --raw-output '.[0].id')
  local body=$(get_release_body $tag $repo_domain $repo_name)
  local release=`curl -H "Authorization: token $ACCESS_TOKEN" -s --request PATCH --data $body  https://api.github.com/repos/$repo_domain/$repo_name/releases/$release_id`
  echo $release
}

function release_tag {
  local tag=$1
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local body=$(get_release_body $tag $repo_domain $repo_name)
  local release=`curl -H "Authorization: token $ACCESS_TOKEN" -s --request POST --data "$body" https://api.github.com/repos/$repo_domain/$repo_name/releases`
  echo $release
}

function upload_asset {
  local repo_domain=${1:?}
  local repo_name=${2:?}
  local release_id=${3:?}
  local asset=${4:?}
  local filename=$(basename $asset)
  if [[ "$filename" == *".zip" ]]; then
    local content_type="application/zip"
  elif [[ "$filename" == *".yaml" ]]; then
    local content_type="text/yaml"
  fi
  curl -H "Authorization: token $ACCESS_TOKEN" \
    -H "Content-Type: $content_type" \
    --data-binary @"$asset" \
    "https://uploads.github.com/repos/$repo_domain/$repo_name/releases/$release_id/assets?name=$filename"
}
