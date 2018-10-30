# Deploying Kubeless Functions using Kubernetes API

Apart from using the `kubeless` CLI, it is possible to deploy Kubeless Functions directly using the Kubernetes API and creating Function objects. A minimal Function might look like:

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: get-python
  namespace: default
  label:
    created-by: kubeless
    function: get-python
spec:
  runtime: python2.7
  timeout: "180"
  handler: helloget.foo
  deps: ""
  checksum: sha256:d251999dcbfdeccec385606fd0aec385b214cfc74ede8b6c9e47af71728f6e9a
  function-content-type: text
  function: |
    def foo(event, context):
        return "hello world"
```

The fields that a Function specification can contain are:

 - Runtime: Runtime ID and version that the function will use. It should match one of the availables in the [Kubeless configuration](/docs/function-controller-configuration).
 - Timeout: Maximum timeout for the given function. After that time, the function execution will be terminated.
 - Handler: Pair of `<file_name>.<function_name>`. When using `zip` in `function-content-type` the `<file_name>` will be used to find the file with the function to expose. In other case it will be used just as a final file name. `<function_name>` is used to select the function to run from the exported functions of `<file_name>`. This field is mandatory and should match with an exported function.
 - Deps: Dependencies of the function. The format of this field will depend on the runtime, e.g. a `package.json` for NodeJS functions or a `Gemfile` for Ruby.
 - Checksum: SHA256 of the function content.
 - Function content type: Content type of the function. Current supported values are `base64`, `url` or `text`. If the content is zipped the suffix `+zip` should be added.
 - Function: Function content.

Apart from the basic parameters, it is possible to add the specification of a `Deployment`, a `Service` or an `Horizontal Pod Autoscaler` that Kubeless will use to generate them.

## Deploying large functions

As any Kubernetes object, function objects have a maximum size of 1.5MiB (due to the [maximum size](https://github.com/etcd-io/etcd/blob/master/Documentation/dev-guide/limit.md#request-size-limit) of an etcd entry). Because of that, it's not possible to specify in the `function` field of the YAML content that surpasses that size. To workaround this issue it's possible to specify an URL in the `function` field. This file will be downloaded at build time (extracted if necessary) and the checksum will be checked. Doing this we avoid any limitation regarding the file size. It's also possible to include the function dependencies in this file and skip the dependency installation step. Note that since the file will be downloaded in a pod the URL should be accessible from within the cluster:

```yaml
  checksum: sha256:d1f84e9f0a8ce27e7d9ce6f457126a8f92e957e5109312e7996373f658015547
  function: https://github.com/kubeless/kubeless/blob/master/examples/nodejs/helloFunctions.zip?raw=true
  function-content-type: url+zip
```

## Custom Deployment

It is possible to specify a [`Deployment` spec](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment) in the Function spec that will be merged with default values set by the Kubeless controller. It is not necessary to specify all the fields of the deployment, just the fields you are interested on overwriting. For example:

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: get-python
...
spec:
...
  deployment:
    spec:
      template:
        spec:
          containers:
          - env:
            - name: FOO
              value: bar
            name: ""
            resources:
              limits:
                cpu: 100m
                memory: 100Mi
              requests:
                cpu: 100m
                memory: 100Mi
            volumeMounts:
            - mountPath: /my_secret
              name: my-secret-vol
          volumes:
          - name: my-secret-vol
            secret:
              secretName: my-secret
```

Would create a function with the environment variable `FOO`, using CPU and memory limits and mounting the secret `my-secret` as a volume. Note that you can also specify a default template for a Deployment spec in the [controller configuration](/docs/function-controller-configuration).

## Custom Service

As with a deployment, it is possible to specify custom values for a [Service](https://kubernetes.io/docs/concepts/services-networking/service). This would be an example:

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: get-python
...
spec:
...
  service:
    clusterIP: None
    ports:
    - name: http-function-port
      port: 9090
      protocol: TCP
      targetPort: 9090
    selector:
      created-by: kubeless
      function: get-python
    type: ClusterIP
```

The example above will create a headless service running in the port 9090.

## Horizontal Pod Autoscaler

For configuring the [autoscale feature](/docs/autoscaling) it is possible to attach an [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to a function:

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: get-python
...
spec:
...
  horizontalPodAutoscaler:
    apiVersion: autoscaling/v2beta1
    kind: HorizontalPodAutoscaler
    metadata:
      name: get-python
      namespace: default
    spec:
      maxReplicas: 3
      metrics:
      - resource:
          name: cpu
          targetAverageUtilization: 70
        type: Resource
      minReplicas: 1
      scaleTargetRef:
        apiVersion: apps/v1beta1
        kind: Deployment
        name: get-python
```

The above specification will create a Horizontal Pod Autoscaler using CPU metrics.
