# Installation

Installation is made of three steps:

* Download the `kubeless` CLI from the [release page](https://github.com/kubeless/kubeless/releases).
* Create a `kubeless` namespace (used by default)
* Then use one of the YAML manifests found in the release page to deploy kubeless. It will create a _functions_ Custom Resource Definition and launch a controller.

There are several kubeless manifests being shipped for multiple k8s environments (non-rbac, rbac and openshift), pick the one that corresponds to your environment:

* `kubeless-$RELEASE.yaml` is used for RBAC Kubernetes cluster.
* `kubeless-non-rbac-$RELEASE.yaml` is used for non-RBAC Kubernetes cluster.
* `kubeless-openshift-$RELEASE.yaml` is used to deploy Kubeless to OpenShift (1.5+).

For example, this below is a show case of deploying kubeless to a Kubernetes cluster (with RBAC available).

```console
$ export RELEASE=$(curl -s https://api.github.com/repos/kubeless/kubeless/releases/latest | grep tag_name | cut -d '"' -f 4)
$ kubectl create ns kubeless
$ kubectl create -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless-$RELEASE.yaml

$ kubectl get pods -n kubeless
NAME                                           READY     STATUS    RESTARTS   AGE
kubeless-controller-manager-567dcb6c48-ssx8x   1/1       Running   0          1h

$ kubectl get deployment -n kubeless
NAME                          DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kubeless-controller-manager   1         1         1            1           1h

$ kubectl get customresourcedefinition
NAME                          AGE
cronjobtriggers.kubeless.io   1h
functions.kubeless.io         1h
httptriggers.kubeless.io      1h
```

> Details on [installing kubeless in a different namespace](/docs/function-controller-configuration#install-kubeless-in-different-namespace) can be found here.

For installing `kubeless` CLI using execute:

#### Linux and macOS

```console
export OS=$(uname -s| tr '[:upper:]' '[:lower:]')
curl -OL https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless_$OS-amd64.zip && \
  unzip kubeless_$OS-amd64.zip && \
  sudo mv bundles/kubeless_$OS-amd64/kubeless /usr/local/bin/
```

Binaries for x86 architectures can be found as well [in the releases page](https://github.com/kubeless/kubeless/releases).

#### Windows

1. Download the latest release from [the releases page](https://github.com/kubeless/kubeless/releases).
2. Extract the content and add the `kubeless` binary to the system PATH.

You are now ready to create functions.

# Sample function

You can use the CLI to create a function. Here is a toy:

```python
def hello(event, context):
  print event
  return event['data']
```

Functions in Kubeless have the same format regardless of the language of the function or the event source. In general, every function:

 - Receives an object `event` as their first parameter. This parameter includes all the information regarding the event source. In particular, the key 'data' should contain the body of the function request.
 - Receives a second object `context` with general information about the function.
 - Returns a string/object that will be used as response for the caller.

You can find more details about the function interface [here](/docs/kubeless-functions#functions-interface)

You create it with:

```console
$ kubeless function deploy hello --runtime python2.7 \
                                --from-file test.py \
                                --handler test.hello
INFO[0000] Deploying function...
INFO[0000] Function hello submitted for deployment
INFO[0000] Check the deployment status executing 'kubeless function ls hello'
```

Let's dissect the command:

* `hello`: This is the name of the function we want to deploy.
* `--runtime python2.7`: This is the runtime we want to use to run our function. Available runtimes are shown in the help information.
* `--from-file test.py`: This is the file containing the function code. It is supported to specify a zip file as far as it doesn't exceed the maximum size for an etcd entry (1 MB).
* `--handler test.foobar`: This specifies the file and the exposed function that will be used when receiving requests. In this example we are using the function `foobar` from the file `test.py`.

You can find the rest of options available when deploying a function executing `kubeless function deploy --help`

You will see the function custom resource created:

```console
$ kubectl get functions
NAME         AGE
hello        1h

$ kubeless function ls
NAME           	NAMESPACE	HANDLER       RUNTIME  	DEPENDENCIES	STATUS
hello         	default  	helloget.foo  python2.7	            	1/1 READY
```

You can then call the function with:

```console
$ kubeless function call hello --data 'Hello world!'
Hello world!
```

Or you can curl directly with `kubectl proxy`using an [apiserver proxy URL](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#manually-constructing-apiserver-proxy-urls).
For example:

```console
$ kubectl proxy -p 8080 &

$ curl -L --data '{"Another": "Echo"}' \
  --header "Content-Type:application/json" \
  localhost:8080/api/v1/namespaces/default/services/hello:http-function-port/proxy/
{"Another": "Echo"}
```

Kubeless also supports [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) which means you can provide your custom URL to the function. Please refer to [this doc](/docs/http-triggers) for more details.

## Clean up

You can delete the function and uninstall Kubeless:

```console
$ kubeless function delete hello

$ kubeless function ls
NAME        NAMESPACE   HANDLER     RUNTIME     DEPENDENCIES    STATUS

$ kubectl delete -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless-$RELEASE.yaml
```

## Examples

See the [examples](https://github.com/kubeless/kubeless/tree/master/examples) directory for a list of simple examples in all the languages supporetd. NodeJS, Python, Golang etc ...

Also checkout the [functions repository](https://github.com/kubeless/functions), where we're building a library of ready to use kubeless examples, including an [incubator](https://github.com/kubeless/functions/tree/master/incubator) to encourage contributions from the community - **your PR is welcome** ! :)
