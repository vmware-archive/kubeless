# Kubeless developer guide

This will cover the steps need to be done in order to build your local developement environment for Kubeless.

## Workflow

### Fork the repo
1. Visit the repo: https://github.com/bitnami/kubeless
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Setup the local directory
First setting your $GOPATH accordingly. Then following the [Golang workspace instruction](https://golang.org/doc/code.html#Workspaces) to make sure your local Kubeless directory is set correctly. It should be located at:

```
working_dir = $GOPATH/src/github.com/bitnami/
```

Your $PATH should also be updated:

```
$ export $PATH=$PATH:$GOPATH/bin
```

**Create your clone:**

```
$ mkdir -p working_dir
$ cd $working_dir
$ git clone https://github.com/$your_github_username/kubeless.git
$ cd $working_dir/kubeless
$ git remote add upstream https://github.com/bitnami/kubeless.git

# Never push to upstream master
$ git remote set-url --push upstream no_push
# Checking your remote set correctly
$ git remote -v
```

### Working on your local
**Branching**

Get your local master up-to-date

```
$ cd $working_dir/kubeless
$ git fetch upstream
$ git checkout master
$ git rebase upstream/master
```

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

## Building Kubeless

We provided `Makefile` to help building the project.

```
$ cd $working_dir/kubeless
$ make binary
```

You can build kubeless for multiple platforms with:

```
$ make binary-cross
```

The binaries accordingly located at `bundles/kubeless_$OS_$arch` folder.

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

Happy hacking!