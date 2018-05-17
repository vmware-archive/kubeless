# Debug Kubeless Functions

In this document we will show how you can debug your function in order to spot possible errors. There could be several reasons that causes a wrong deployment. For learning how to successfully debug a function it is important to know what is the process of deploying a Kubeless function. In this guide we are going to assume that you are using the `kubeless` CLI tool to deploy your functions. If that is the case, this is the process to run a function:

 1. The `kubeless` CLI read the parameters you give to it and produces a [Function](/docs/advanced-function-deployment) object that submits to the Kubernetes API server.
 2. The Kubeless Function Controller detects that a new `Function` has been created and reads its content. From the function content it generates: a `ConfigMap` with the function code and its dependencies, a `Service` to make the function reachable through HTTP and a `Deployment` with the base image and all the required steps to install and run your functions. It is important to know this order because if the controller fails to deploy the `ConfigMap` or the `Service` it will never create the `Deployment`. A failure in any step will abort the process.
 3. Once the `Deployment` has been created a `Pod` should be generated with your function. When a Pod starts it dinamically reads the content of your function (in case of interpreted languages).

After all the above you are ready to call your function. Let's see some common mistakes and how to fix them.

## "kubeless function deploy" fails

The first failure that can appear is an error in the parameters that we give to the `kubeless function deploy` command. Hopefully this errors are pretty easy to debug:

```console
$ kubeless function deploy --runtime node8 \
  --from-file hello.js \
  --handler todos.create \
  --dependencies package.json \
  hello
FATA[0000] Invalid runtime: node8. Supported runtimes are: python2.7, python3.4, python3.6, nodejs6, nodejs8, ruby2.4, php7.2, go1.10
```

In the above we can see that we have a typo in the runtime. It should be `nodejs8` instead of `node8`.

## "kubeless function ls" returns "MISSING: Check controller logs"

There will be cases in which the validations done in the CLI won't be enough to spot a problem in the given parameters. If that is the case the function `Deployment` will never appear. To debug this kind of issues it is necessary to check what is the error in the controller logs. To retrieve these logs execute:

```
$ kubeless function deploy foo --from-file hellowithdata.py --handler hello,foo --runtime python3.6
INFO[0000] Deploying function...
INFO[0000] Function foo submitted for deployment
INFO[0000] Check the deployment status executing 'kubeless function ls foo'
$ kubeless function ls
NAME 	NAMESPACE	HANDLER  	RUNTIME  	DEPENDENCIES	STATUS
foo  	default  	hello,foo	python3.6	            	MISSING: Check controller logs
$ kubectl logs -n kubeless -l kubeless=controller
time="2018-04-27T15:12:28Z" level=info msg="Processing update to function object foo Namespace: default" controller=cronjob-trigger-controller
time="2018-04-27T15:12:28Z" level=error msg="Function can not be created/updated: failed: incorrect handler format. It should be module_name.handler_name" pkg=function-controller
time="2018-04-27T15:12:28Z" level=error msg="Error processing default/foo (will retry): failed: incorrect handler format. It should be module_name.handler_name" pkg=function-controller
```

From the logs we can see that there is a problem with the handler: we specified `hello,foo` while the correct value is `hello.foo`.

## Function pod is crashing

The most common error is finding that the `Deployment` is generated successfully but the function remains with the status `0/1 Not ready`. This is usually caused by a syntax error in our function or in the dependencies we specify.

If our function doesn't start we should check the status of the pods executing:

```
$ kubectl get pods -l function=foo
```

### Function pod crashes with Init:CrashLoopBackOff

If our function fails with an `Init` error that could mean that:

 - It fails to retrieve the function content.
 - It fails to install dependencies.
 - It fails to compile our function (in compiled languages).

For any of the above we should first identify which container is failing (since each step is performed in a different container):

```console
$ kubectl get pods -l function=foo
NAME                   READY     STATUS                  RESTARTS   AGE
foo-74978bbf45-9xb4p   0/1       Init:CrashLoopBackOff   1         6m
$ kubectl get pods -l function=foo -o yaml
...
      name: install
      ready: false
      restartCount: 2
...
```

From the above we can see that is the container `install` is the one with the problem. Depending on the runtime the logs of the container will be shown as well so we can directly spot the issue. Unfortunately that is not the case so let's retrieve manually the logs of the `install` container:

```console
$ kubectl logs foo-74978bbf45-9xb4p -c install --previous
...
Collecting twiter (from -r /kubeless/requirements.txt (line 1))
  Retrying (Retry(total=4, connect=None, read=None, redirect=None, status=None)) after connection broken by 'NewConnectionError('<pip._vendor.urllib3.connection.VerifiedHTTPSConnection object at 0x7f10eb4d7400>: Failed to establish a new connection: [Errno -3] Temporary failure in name resolution',)': /simple/twiter/
```

Now we can spot that the problem is a typo in our requirements: `twiter` should be `twitter`.

### Function pod crashes with CrashLoopBackOff

In the case the Pod remains in that state we should retrieve the logs of the runtime container:

```console
$ kubectl get pods -l function=bar
NAME                   READY     STATUS             RESTARTS   AGE
bar-7d458f6d7c-2gsh7   0/1       CrashLoopBackOff   7          15m
$ kubectl logs -l function=bar
kubectl logs -l function=bar
Traceback (most recent call last):
...
  File "/kubeless/hello.py", line 2
    return Hello world
                     ^
SyntaxError: invalid syntax
```

We can see that we have a syntax error: `return Hello world` should be modified with `return "Hello world"`.

### Function returns an "Internal Server Error"

There will be cases in which the pod doesn't crash but the function returns an error:

```console
$ kubectl get pods -l function=test
NAME                    READY     STATUS    RESTARTS   AGE
test-6845ff45cb-6q865   1/1       Running   0          1m
$ kubeless function call test --data '{"username": "test"}'
ERRO[0000]
FATA[0000] an error on the server ("Internal Server Error") has prevented the request from succeeding
```

This usually means that the function is syntactically correct but it has a bug. Again for spotting the issue we should check the function logs:

```console
$ kubectl logs -l function=test
...
[27/Apr/2018:15:45:33 +0000] "GET /healthz HTTP/1.1" 200 2 "-" "kube-probe/."
Function failed to execute: TypeError: Cannot read property 'name' of undefined
    at handler (/kubeless/hello.js:3:39)
    ...
```

We can see that it is raising an error in the line 3 of our function:

```js
module.exports = {
  handler: (event, context) => {
    return "Hello " + event.data.user.name;
  },
};
```

We are trying to access the property `name` of the property `user` while we are giving the function `username` instead.

## Conclusion

These are just some tips to quickly identify what's gone wrong with a function. If after checking the controller and function logs (or any other information that Kubernetes may provide) you are not able to spot the error you can open an [Issue in our GitHub repository](https://github.com/kubeless/kubeless/issues) or contact us through [slack](http://slack.k8s.io) in the #kubeless channel.
