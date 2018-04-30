# Controller configurations for Functions

### Using ConfigMap

Configurations for functions can be done in `ConfigMap`: `kubeless-config` which is a part of `Kubeless` deployment manifests.

Deployments for function can be configured in `data` inside the `ConfigMap`, using key `deployment`, which takes a string in the form of `yaml/json` and is driven by the structure of [v1beta2.Deployment](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.9/#deployment-v1beta2-apps):

E.g. In the below configuration, new **annotations** are added globally to all function deployments and podTemplates and **replicas** for each function pod will be `2`.

```yaml
apiVersion: v1
data:
  deployment: |-
    {
      "metadata": {
          "annotations":{
            "annotation-to-deployment": "value"
          }
      },
      "spec": {
        "replicas": 2,
        "template": {
          "annotations": {
            "annotations-to-pod": "value"
          }
        }
      }
    }
  ingress-enabled: "false"
  service-type: ClusterIP
kind: ConfigMap
metadata:
  name: kubeless-config
  namespace: kubeless
```

It is **recommended** to have controlled custom configurations on the following **items** (*but is not limited to just these*):

> Warning: You should know what you are doing.

- v1beta2.Deployment.ObjectMeta.Annotations
- v1beta2.Deployment.Spec.replicas
- v1beta2.Deployment.Spec.Strategy
- v1beta2.Deployment.Spec.Template.ObjectMeta.Annotations
- v1beta2.Deployment.Spec.Template.Spec.NodeSelector
- v1beta2.Deployment.Spec.Template.Spec.NodeName

Having said all that, if one wants to override configurations from the `ConfigMap` then in `Function` manifest one needs to provide the details as follows:

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: testfunc
spec:
  deployment:  ### Definition as per v1beta2.Deployment
    metadata:
      annotations:
        "annotation-to-deploy": "final-value-in-deployment"
    spec:
      replicas: 2  ### Final deployment gets Replicas as 2
      template:
        metadata:
          annotations:
            "annotation-to-pod": "value"
  deps: ""
  function: |
    module.exports = {
      foo: function (req, res) {
            res.end('hello world updated!!!')
      }
    }
  function-content-type: text
  handler: hello.foo
  runtime: nodejs8
  service:
    ports:
    - name: http-function-port
      port: 8080
      protocol: TCP
      targetPort: 8080
    type: ClusterIP
```

### Install kubeless in different namespace

If you have installed kubeless into some other namespace (which is not called `kubeless`) or changed the name of the config file from kubeless-config to something else, then you have to export the kubeless namespace and the name of kubeless config as environment variables before using kubless cli. This can be done as follows:

```bash
$ export KUBELESS_NAMESPACE=<name of namespace>
$ export KUBELESS_CONFIG=<name of config file>
```

or the following information can be added to `functions.kubeless.io` `CustomResourceDefinition` as `annotations`. E.g. below `CustomResourceDefinition` will signify `kubeless-controller` is installed in namespace `kubless-new-namespace` and config name is `kubeless-config-new-name`

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
description: Kubernetes Native Serverless Framework
kind: CustomResourceDefinition
metadata:
  name: functions.kubeless.io
  annotations:
    kubeless.io/namespace: kubless-new-namespace
    kubeless.io/config: kubeless-config-new-name
spec:
  group: kubeless.io
  names:
    kind: Function
    plural: functions
    singular: function
  scope: Namespaced
  version: v1beta1
```

The priority of deciding the `namespace` and `config name` (highest to lowest) is:

- Environment variables
- Annotations in `functions.kubeless.io` CRD
- default: `namespace` is `kubeless` and `ConfigMap` is `kubeless-config`

## Using custom images

It is possible to configure the different images that Kubeless uses for deploy and execute functions. In this ConfigMap you can configure:

 - Different or additional runtimes. For doing so it is possible to modify/add a runtime in the field `runtimeImages`. Runtimes are categorized by major version. See the guide for [implementing a new runtime](/docs/implementing-new-runtime) for more information. Each major version has:
  - Name: Unique ID of the runtime. It should contain the runtime name and version.
  - Version: Major and minor version of the runtime.
  - Runtime Image: Image used to execute the function.
  - Init Image: Image used for installing the function and/or dependencies.
  - (Optional) Image Pull Secrets: Secret required to pull the image in case the repository is private.
 - The image used to populate the base image with the function. This is called `provision-image`. This image should have at least `unzip` and `curl`. It is also possible to specify `provision-image-secret` to specify a secret to pull that image from a private registry. 
 - The image used to build function images. This is called `builder-image`. This image is optional since its usage can be disabled with the property `enable-build-step`. A Dockerfile to build this image can be found [here](https://github.com/kubeless/kubeless/tree/master/docker/function-image-builder). It is also possible to specify `builder-image-secret` to specify a secret to pull that image from a private registry.
 