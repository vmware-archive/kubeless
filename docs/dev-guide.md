# Kubeless developer guide

This will cover the steps need to be done in order to build your local developement environment for Kubeless.

## Setting things up

As Kubeless project is mainly developed in the Go Programming Language, the first thing you should do is guarantee that Go is installed and all environment variables are properly set.

In this example we will use Ubuntu Linux 16.04.2 LTS as the target host on where the project will be built.

### Installing Go

* Visit https://golang.org/dl/
* Download the most recent Go version (here we used 1.9) and unpack the file
* Check the installation process on https://golang.org/doc/install
* Set the Go environment variables

````
export GOROOT=/GoDir/go
export GOPATH=/GoDir/go/bin
export PATH=$GOPATH:$PATH
````

### Create a working directory for the project

````
working_dir=$GOROOT/src/github.com/kubeless/
mkdir -p $working_dir
````

### Fork the repository

1. Visit the repo: https://github.com/kubeless/kubeless
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone from your fork

```
working_dir = $GOROOT/src/github.com/kubeless/
cd $working_dir
git clone https://github.com/<YOUR FORK>
cd $working_dir/kubeless
git remote add upstream https://github.com/kubeless/kubeless.git

# Never push to upstream master
git remote set-url --push upstream no_push
# Checking your remote set correctly
git remote -v
```

### Building local binaries

To make the binaries for your platform, run:

```
cd $working_dir/kubeless
make binary
make controller-image
```

This will instruct "make" to run the scripts to build the kubeless client and the kubeless controller image.

You can build kubeless for multiple platforms with:

```
$ make binary-cross
```

The binaries accordingly located at `bundles/kubeless_$OS_$arch` folder.

### Uploading your kubeless image to Docker Hub

Usually you will need to upload your controller image to a repository so you can make it available for your Kubernettes cluster, whenever it is running.

To do so, run the commands:

````
docker login -u=<dockerhubuser> -e=<e-mail>
docker tag kubeless-controller <your-docker-hub-repo>/kubeless-test:latest
docker push <your-docker-hub-repo>/MyKubelessController:latest 
````
In order to upload your kubeless controller image to Kubernettes, you should use kubectl as follows, informing the yaml file with the required descriptions of your deployment.

````
kubectl create -f <path-to-yaml-file>
````
Make sure your image repository is correctly referenced in the "containers" session on the yaml file.

```
      containers:
      - image: fabriciosanchez/kubeless-test:latest
        imagePullPolicy: Always
        name: kubeless-controller
      serviceAccountName: controller-acct
```

**Hint:** take a look at the `imagePullPolicy` configuration if you are sending images with tags (e. g. "latest") to the Kubernettes cluster. This option controls the image caching mechanism for Kubernettes and you may enconter problems if new images enters the cluster with the same name. They might not be properly pulled for example. 

### Working on your local branch

Branch from it:

```
$ git checkout -b myfeature
```

Then start working on your `myfeature` branch.

**Keep your branch in sync**

```
# While on your myfeature branch
$ git fetch upstream
$ git rebase upstream/master
```

**Commit your changes**

```
$ git commit
```

Likely you go back and edit/build/test some more then `commit --amend` in a few cycles.

**Push to your origin first**

```
$ git push origin myfeature
```

### Creat a pull request

1. Visit your fork at https://github.com/$your_github_username/kubeless.
2. Click the `Compare & pull request` button next to your `myfeature` branch.
3. Make sure you fill up clearly the description, point out the particular issue your PR is mitigating, and ask for code review.

## Testing kubeless with local minikube

The simplest way to try kubeless is deploying it with [minikube](https://github.com/kubernetes/minikube)

```
$ minikube start
```

You can choose to start minikube vm with your preferred VM driver (virtualbox xhyve vmwarefusion)

```
$ kubeless install

# Verify the installation
$ kubectl get po --all-namespaces
NAMESPACE     NAME                                   READY     STATUS    RESTARTS   AGE
kube-system   heapster-ptbc4                         1/1       Running   27         55d
kube-system   influxdb-grafana-ml60t                 2/2       Running   54         55d
kube-system   kube-addon-manager-minikube            1/1       Running   27         55d
kube-system   kube-dns-v20-p8k3c                     3/3       Running   81         55d
kube-system   kubernetes-dashboard-s9dv1             1/1       Running   27         55d
kubeless      kafka-controller-2158053540-3q6pv      2/2       Running   2          11h
kubeless      kubeless-controller-2330759281-d0l4v   2/2       Running   4          11h
```

Make sure your have `kubeless-controller` and `kafka-controller` running at the default `kubeless` namespace. For problems you might have with kubeless-controller, remember to take a look at its log:

```
$ kubectl logs kubeless-controller-2330759281-d0l4v -n kubeless -c kubeless -f
```

Or you might wanna check log of a particular function deployed to kubeless.

```
$ kubeless function logs <function_name> -f
```

## Scripting build and publishing

Example of shell script to setup a local environment, build the kubeless binaries and make it available on kubernettes.

```
#!/bin/bash


# Please set GOROOT and GOPATH appropriately before running!
#rm -rf $GOROOT/src/github.com

#export GOROOT=
#export GOPATH=
#export PATH=$GOPATH:$PATH

#working_dir=$GOPATH/src/github.com/kubeless/
#mkdir -p $working_dir
#cd $working_dir
#git clone https://github.com/<INCLUDE HERE YOUR FORK AND UNCOMMENT>
#cd $working_dir/kubeless
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

We use [Glide](https://github.com/Masterminds/glide) to vendor the dependencies. Take a quick look on README to understand how it works. Packages that Kubeless relying on is listed at [glide.yaml](https://github.com/kubeless/kubeless/blob/master/glide.yaml)

Happy hacking!