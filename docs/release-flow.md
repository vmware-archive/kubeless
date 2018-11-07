# Introduction

Kubeless leverages [travis-ci](https://travis-ci.org/) to construct an automated release flow. A release package includes kubeless binaries for multiple platforms (linux and osx are supported) and one yaml file to deploy kubeless controller.

# Checks before releasing

Before releasing it is necessary to check that the rest of projects of the Kubeless environment do not present regressions for the new changes. Before creating a new release, deploy Kubeless using the latest commit of master (using the tag "latest" for the controller image). Make sure that the latest image build in Travis for the Kubeless controller is being used. After that, ensure that the following projects support the new version:

 - [Serverless Plugin](https://github.com/serverless/serverless-kubeless)
 - [Kubeless UI](https://github.com/kubeless/kubeless-ui)

If any error is found after doing some manual testing, make sure the error is addressed before doing a release.

# Kubeless release flow

A release is triggered by [Travis Github Releases](https://docs.travis-ci.com/user/deployment/releases/) and based on GitHub tagging. Once a commit in the master branch is tagged, a travis job will be started to build and upload assets to Github release page under a new release with the tag name. The setup is described at `before_deploy` and `deploy` sections in `.travis.yaml`.

`before_deploy` defines commands executed before releasing. At this stage, we prepare assets which will be uploaded including kubeless binaries and the yaml file. The yaml file is converted from [kubeless.jsonnet](https://github.com/kubeless/kubeless/blob/master/kubeless.jsonnet) file using [kubecfg](https://github.com/ksonnet/kubecfg). The kubeless-controller is built in format of docker image and push to [Bitnami repository](https://hub.docker.com/r/bitnami/kubeless-controller/) on DockerHub. Because we use sha256 digest for labeling docker images to be deployed when installing kubeless, we need to update these digests for the new release.

`deploy` defines configuration for a github release. API key is encrypted version of our Github token with scope `public_repo`. The condition for a release to be triggered is defined at `on` section:
- it will be triggered once a commit is tagged
- the repository is `kubeless/kubeless`
- only travis job for `os: linux` and `go: 1.8` can do the release

Once the release job has finished a `Draft` with the release notes will appear in the [releases page](https://github.com/kubeless/kubeless/releases). Review the notes and include a summary of the changes included in the release. Delete information that is not useful for the users. Make sure that breaking changes are properly highlighted. After that click on "Publish" for making the new release available for anyone.

# Update the rest of projects to use the new version

_Note: These steps are suitable for being automated in the Travis release job_

Once the new version is available, there are several projects/files that require to be updated in order to point to the latest version:
 
 - Kubeless docs site: To point to the latest version in the docs of http://kubeless.io rebuild the last build on https://travis-ci.org/kubeless/kubeless-website.
 - Kubeless chart: Update the references for the different images or any other required change in the `chart` folder of this repository.
 - Serverless plugin: Update the `KUBELESS_VERSION` environment variable in the `.travis` file to point to the latest version.
 - [Optional] Brew recipes: An automated PR will be generated in the `homebrew-core` repository with the new version and commit ID. Unless the recipe should contain breaking changes the update will be handled by the homebrew team. If it is not the case the [recipe](https://github.com/Homebrew/homebrew-core/blob/master/Formula/kubeless.rb) manually.
