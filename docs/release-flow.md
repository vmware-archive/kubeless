# Kubeless release flow

Kubeless leverages [travis-ci](https://travis-ci.org/) to construct an automated release flow. A release package includes kubeless binaries for multiple platforms (linux and osx are supported) and one yaml file to deploy kubeless controller.

A release is triggered by [Travis Github Releases](https://docs.travis-ci.com/user/deployment/releases/) and based on github tagging. Once a commit in the master branch is tagged, a travis job will be started to build and upload assets to Github release page under a new release with the tag name. The setup is described at `before_deploy` and `deploy` sections in `.travis.yaml`:

```yaml
before_deploy:
  - make binary-cross
  - for d in bundles/kubeless_*; do zip -r9 $d.zip $d/; done
  - |
    if [ "$TRAVIS_OS_NAME" = linux ]; then
      docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
      docker push $CONTROLLER_IMAGE
      NEW_DIGEST=$(./script/find_digest.sh ${CONTROLLER_IMAGE_NAME} ${TRAVIS_TAG} | cut -d " " -f 2)
      OLD_DIGEST=$(cat kubeless.yaml | grep ${CONTROLLER_IMAGE_NAME} |  cut -d "@" -f 2)
      sed 's/'"$OLD_DIGEST"'/'"$NEW_DIGEST"'/g' kubeless.yaml > kubeless-${TRAVIS_TAG}.yaml
    fi
```

`before_deploy` defines commands executed before releasing. At this stage, we prepare assets which will be uploaded including kubeless binaries and the yaml file. The yaml file is converted from [kubeless.jsonnet](https://github.com/kubeless/kubeless/blob/master/kubeless.jsonnet) file using [kubecfg](https://github.com/ksonnet/kubecfg). The kubeless-controller is built in format of docker image and push to [Bitnami repository](https://hub.docker.com/r/bitnami/kubeless-controller/) on DockerHub. Because we use sha256 digest for labeling docker images to be deployed when installing kubeless, we need to update these digests for the new release.

```yaml
deploy:
  provider: releases
  api_key:
    secure: "daSVcUjHGO5L1SlyX3+GOz/Hv4KqkC/ObdRAbr7n7bvZUQewN5ll9Q5z5gvOI9z+zY9nnYFVppaOLfdWGjSldxjepwdHFB0F/oBXVwD+l7WOWacpq8+0+Zm71gG31YMl9ImkgmFll9WwXcG/2MLSGUsyBInGLzyRT0l5OhG6PNWLO3p7YZBKr9ihF3nxeRucfwn2uMUuyVnlEyBCyqmpxeKMNK3J9FY9j3jTmTmRQCBc9UEGGcX2p7R+LtLEvg/nxllVZyHQYc6UhuXWC6imx1mo6NAsVNSvbgJ0ufWrwcd+43XmU5eO8IPMlkb/EVgzp40XWoMrsxQZLEhnkLuApLFMdgn3ioJynBmtYiP3mQZ67Dt5zQiA7+UW/D+1CVjb7hVTzk76Vm4oEEq1KUs2xnA+kaUKprTLrlW4EusiSqKPZWG+hrqlC6VGYTSpQBtx9Hgo5yHZ0r5zoNCrXy3ISsCf8ublH+eiQ5tQ4pFD42AlLKGOQEpL6cvu2zl7+KZfWDqXG8dnjni7aguuPEdXnUm/bsv7JWKwFFV8CHfmBuWruU+VG1iVvSOmATlwQgb72ZsE+bJw5nFWTbrsVeK+hXSBk1tiWWDJzYEC9e81hFKxw8ltPHnG+dos941FXjNVG81L6TcuCgc42Ca4Kf7ICirjdmCKT3e7FW60IFw8SLM="
  file_glob: true
  file:
    - kubeless-${TRAVIS_TAG}.yaml
    - bundles/kubeless_*.zip
  skip_cleanup: true
  overwrite: true
  on:
    tags: true
    repo: kubeless/kubeless
    go: 1.8
    os: linux
```

`deploy` defines configuration for a github release. API key is encrypted version of our Github token with scope `public_repo`. The condition for a release to be triggered is defined at `on` section:
- it will be triggered once a commit is tagged
- the repository is `kubeless/kubeless`
- only travis job for `os: linux` and `go: 1.8` can do the release
