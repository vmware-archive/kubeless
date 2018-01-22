# Controller configurations for Functions
### Using `ConfigMap`
Configurations for functions can be done in `ConfigMap`: `kubeless-config` which is a part of `Kubeless` deployment manifests.

Deployments for function can be configured in `data` inside the `ConfigMap`, using key `deployment`, which takes a string in the form of `yaml/json` and is driven by the structure of [v1beta2.Deployment](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.9/#deployment-v1beta2-apps):
E.g.
```yaml
---
apiVersion: v1
data:
  deployment: |-
    {
      "metadata": {
          "annotation":
            "annotation-to-deployment": "value"
      },
      "spec": {
        "replicas": 1
        "template":
          "annotations":
            "annotations-to-pod": "value"
      }
    }
  ingress-enabled: "false"
  service-type: ClusterIP
kind: ConfigMap
metadata:
  name: kubeless-config
  namespace: kubeless
```

This in turn will add the config to all the function `deployments` created by controller.

Having said all that, if one wants to override configurations from the `ConfigMap` then in `Function` manifest one needs to provide the details as follows: 
```
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