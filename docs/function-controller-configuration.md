# Controller configurations for Functions

## Using ConfigMap

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

## Install kubeless in different namespace

If you have installed kubeless into some other namespace (which is not called `kubeless`) or changed the name of the config file from kubeless-config to something else, then you have to export the kubeless namespace and the name of kubeless config as environment variables before using kubless cli. This can be done as follows:

```bash
$ export KUBELESS_NAMESPACE=<name of namespace>
$ export KUBELESS_CONFIG=<name of config file>
```

or the following information can be added to `functions.kubeless.io` `CustomResourceDefinition` as `annotations`. E.g. below `CustomResourceDefinition` will signify `kubeless-controller` is installed in namespace `kubless-new-namespace` and config name is `kubeless-config-new-name`

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
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

### Install several instances of kubeless (multi-tenancy)

It is possible to install Kubeless in several namespaces. This allow administrators to have several instances of Kubeless that can be configured differently (for example using different runtime images or with different Docker credentials).

In order to install Kubeless in a custom namespace (or in several ones) it's necessary to:

 - Install the `CustomResourceDefinitions` and `ClusterRoles` as in the default scenario. These resources are not namespaced which means that you need to install them just once. It is also recommendable to split the current rules of the `ClusterRole` into two different roles: one just for accessing cluster-wide resources like `CustomResourceDefinitions` and a second one with the rest of resources. That way it's possible to attach the first `ClusterRole` to a `ClusterRoleBinding` as the default scenario but attaching the second one with a namespaced `RoleBinding` to avoind unauthorized access to other namespaces. More information about `RBAC` [here](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).
 - The rest of the resources you can find in the installation manifest (`Deployment`, `ConfigMap`, `ServiceAccount`...) are namespaced. This means that it's required to modify the `metadata.namespace` of each one of those to target the correct namespace.
 - The next step is to set in the Kubeless ConfigMap the namespace in which the controller should listen for functions. This is set in the variable `functions-namespace`. If this value is empty it will try to find functions in all namespaces.

This is an example of a manifest (simplified) for a Kubeless instance deployed in the namespace "test":

```yaml
# RBAC Configuration
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kubeless-controller-read
rules:
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kubeless-controller-deployer
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  verbs:
  ... # The rest of the ClusterRole has been omitted
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-acct
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubeless-controller-read-test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubeless-controller-read
subjects:
- kind: ServiceAccount
  name: controller-acct
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: kubeless-controller-deployer
  namespace: test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubeless-controller-deployer
subjects:
- kind: ServiceAccount
  name: controller-acct

# Kubeless Configuration
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeless-config
  namespace: test
data:
  functions-namespace: "test"
  ...  # The rest of the ConfigMap data has been omitted

# Kubeless core controller
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    kubeless: controller
  name: kubeless-controller-manager
  namespace: test
spec:
  ... # The rest of the Deployment has been omitted
```

The same process should be followed for any trigger controller installed (Kafka, Nats, ...): Adapt the RBAC configuration and change the resources namespace. These controllers will read the `functions-namespace` property from the main ConfigMap.

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
 
## Authenticate Kubeless Function Controller using OAuth Bearer Token

In some non-RBAC k8s deployments using webhook authorization, service accounts may have insufficient privileges to perform all k8s operations that the Kubeless Function Controller requires for interacting with the cluster. It's possible to override the default behavior of the Kubeless Function Controller using a k8s serviceaccount for authentication with the cluster and instead use a provided OAuth Bearer token for all k8s operations. 

This can be done by creating a k8s secret and mounting that secret as a volume on controller pods, then setting the environmental variable `KUBELESS_TOKEN_FILE_PATH` to the filepath of that secret. Be sure to set this environmental variable on the controller template spec or to every pod created in the deployment.

For example, if the bearer token is mounted at /mnt/secrets/bearer-token, this k8s spec can use it:

```yaml
# Kubeless core controller
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: kubeless-controller-manager
  namespace: kubeless
  labels:
    kubeless: controller
spec:
  template:
    metadata:
      labels:
        kubeless: controller
    spec:
      containers:
      - env:
        - name: KUBELESS_TOKEN_FILE_PATH
          value: /mnt/secrets/bearer-token
  ... # The rest of the Deployment has been omitted
```

