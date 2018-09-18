# Kubeless developer guide

This will cover the steps need to be done in order to build your local development environment for Kubeless.

## Setting things up

As Kubeless project is mainly developed in the Go Programming Language, the first thing you should do is guarantee that Go is installed and all environment variables are properly set.

In this example we will use Ubuntu Linux 16.04.2 LTS as the target host on where the project will be built.

### Installing Go

* Visit [https://golang.org/dl/](https://golang.org/dl/)
* Download the most recent Go version (here we used 1.9) and unpack the file
* Check the installation process on [https://golang.org/doc/install](https://golang.org/doc/install)
* Set the Go environment variables

```bash
export GOROOT=/GoDir/go
export GOPATH=/GoDir/go/bin
export PATH=$GOPATH:$PATH
```

### Create a working directory for the project

```bash
export KUBELESS_WORKING_DIR=$GOROOT/src/github.com/kubeless/
mkdir -p $KUBELESS_WORKING_DIR
```

### Fork the repository

1. Visit the repo: [https://github.com/kubeless/kubeless](https://github.com/kubeless/kubeless)
1. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone from your fork

```bash
cd $KUBELESS_WORKING_DIR
git clone https://github.com/<YOUR FORK>
cd $KUBELESS_WORKING_DIR/kubeless
git remote add upstream https://github.com/kubeless/kubeless.git

# Never push to upstream master
git remote set-url --push upstream no_push
# Checking your remote set correctly
git remote -v
```

### Bootstrapping your local dev environment

To get all the needed tools to build and test, run:

```bash
cd $KUBELESS_WORKING_DIR/kubeless
make bootstrap
```

Or if you want to use a containerized environment you can use [minikube](https://github.com/kubernetes/minikube).
If you already have minikube use the following script to set it up:

```bash
cd $KUBELESS_WORKING_DIR/kubeless
./script/start-test-environment.sh
```

This will start a new minikube virtual machine and will open a bash shell in which you can build any local binary or execute the tests. Note that the Kubeless code will be mounted from outside so you can still edit your files with your favourite text editor.

### Building local binaries

To make the binaries for your platform, run:

```bash
cd $KUBELESS_WORKING_DIR/kubeless
make binary
make controller-image
```

This will instruct "make" to run the scripts to build the kubeless client and the kubeless controller image.

You can build kubeless for multiple platforms with:

```bash
make binary-cross
```

The binaries accordingly located at `bundles/kubeless_$OS_$arch` folder.

### Building Trigger Controllers

Each Kubeless trigger controller is being developed on its own repository. You can find more information about those controllers in their repositories:

 - [HTTP Trigger](https://github.com/kubeless/http-trigger)
 - [CronJob Trigger](https://github.com/kubeless/cronjob-trigger)
 - [Kafka Trigger](https://github.com/kubeless/kafka-trigger)
 - [NATS Trigger](https://github.com/kubeless/nats-trigger)

### Building k8s manifests file

To regenerate the most updated k8s manifests file, run:

> Note that you will need the [`kubecfg`](https://github.com/ksonnet/kubecfg/releases/) in your `PATH` in order to generate the Kubeless manifests.

```bash
cd $KUBELESS_WORKING_DIR
export KUBECFG_JPATH=$PWD/ksonnet-lib
git clone --depth=1 https://github.com/ksonnet/ksonnet-lib.git
cd $KUBELESS_WORKING_DIR/kubeless
make all-yaml
```

If everything is ok, you'll have generated manifests file under the `$KUBELESS_WORKING_DIR` root directory:

```
kubeless-openshift.yaml
kubeless-non-rbac.yaml
kubeless.yaml
```

You can also generate them separated using the following commands:

```bash
make kubeless-openshift.yaml
make kubeless-non-rbac.yaml
make kubeless.yaml
```

### Uploading your kubeless image to Docker Hub

Usually you will need to upload your controller image to a repository so you can make it available for your Kubernetes cluster, whenever it is running.

To do so, run the commands:

```bash
docker login -u=<dockerhubuser> -e=<e-mail>
docker tag kubeless-controller-manager <your-docker-hub-repo>/kubeless-test:latest
docker push <your-docker-hub-repo>/kubeless-test:latest
```

Make sure your image repository is correctly referenced in the "containers" session on the yaml file.

```yaml
      containers:
      - image: fabriciosanchez/kubeless-test:latest
        imagePullPolicy: Always
        name: kubeless-controller
      serviceAccountName: controller-acct
```

**Hint:** take a look at the `imagePullPolicy` configuration if you are sending images with tags (e. g. "latest") to the Kubernetes cluster. This option controls the image caching mechanism for Kubernetes and you may encounter problems if new images enters the cluster with the same name. They might not be properly pulled for example.

In order to upload your kubeless controller image to Kubernetes, you should use kubectl as follows, informing the yaml file with the required descriptions of your deployment.

```bash
kubectl create ns kubeless
kubectl create -f <path-to-yaml-file>/kubeless.yaml
```

### Working on your local branch

Branch from it:

```bash
git checkout -b myfeature
```

Then start working on your `myfeature` branch.

#### Keep your branch in sync

```bash
# While on your myfeature branch
git fetch upstream
git rebase upstream/master
```

#### Commit your changes

```bash
git commit
```

Likely you go back and edit/build/test some more then `commit --amend` in a few cycles.

#### Push to your origin first

```bash
git push origin myfeature
```

### Updating generated files

There are several files that are automatically generated by Kubernetes [code-generator](https://github.com/kubernetes/code-generator) based on the API [specification](https://github.com/kubeless/kubeless/tree/master/pkg/apis/kubeless) in the repository.

These include:

* Clientset
* Listers
* Shared informers
* Deepcopy functions

If you make any changes to API specification, you will need to run `make update` to regenerate clientset, informers, lister and deepcopy functions.

### Testing kubeless with local minikube

The simplest way to try kubeless is deploying it with [minikube](https://github.com/kubernetes/minikube)

You can start working with the local minikube VM and test your changes building the controller image and running your tests. Once you are happy with the result and you are ready to send a pull request you should run the unit and end-to-end tests (to spot possible issues with your changes):

```bash
make validation
make test
make build_and_test
```

Note that for running the end-to-end tests you need to provide a clean profile of minikube (you can create a specific profile for the tests with `minikube profile tests`).

Any new feature/bug fix made to the code should be accompanied by a unit or end to end test.

### Create a pull request

1. Visit your fork at [https://github.com/$your_github_username/kubeless](https://github.com/$your_github_username/kubeless).
1. Click the `Compare & pull request` button next to your `myfeature` branch.
1. Make sure you fill up clearly the description, point out the particular
   issue your PR is mitigating, and ask for code review.

## Scripting build and publishing

Example of shell script to setup a local environment, build the kubeless binaries and make it available on kubernetes.

```bash
#!/bin/bash


# Please set GOROOT and GOPATH appropriately before running!
#rm -rf $GOROOT/src/github.com

#export GOROOT=
#export GOPATH=
#export PATH=$GOPATH:$PATH

#KUBELESS_WORKING_DIR=$GOPATH/src/github.com/kubeless/
#mkdir -p $KUBELESS_WORKING_DIR
#cd $KUBELESS_WORKING_DIR
#git clone https://github.com/<INCLUDE HERE YOUR FORK AND UNCOMMENT>
#cd $KUBELESS_WORKING_DIR/kubeless
#git remote add upstream https://github.com/DXBrazil/kubeless
#git remote set-url --push upstream no_push
#git remote -v
# git checkout <INCLUDE HERE YOUR BRANCH AND UNCOMMENT>
#git fetch

#make binary
#make controller-image

#docker login -u=<your docker hub user> -e=<your e-mail>
#docker tag kubeless-controller <yourrepo>/<your-image>
#docker push <your repo>/<your-image>

#kubectl delete -f <path-to-yaml>
#kubectl delete namespace kubeless

#a=Terminating

#while [ $a == Terminating ]
#do

#a=`kubectl get ns | grep Termina | awk '{print $2}'`
#sleep 5

#done

#kubectl create namespace kubeless
#kubectl create -f <path-to-yaml>
```

## Manage dependencies

We use [dep](https://github.com/golang/dep) to vendor the dependencies. Take a quick look at the README to understand how it works. Packages that Kubeless relies on are listed at [Gopkg.toml](https://github.com/golang/dep/blob/master/docs/Gopkg.toml.md).

Happy hacking!
