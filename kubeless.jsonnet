local k = import "ksonnet.beta.1/k.libsonnet";

local objectMeta = k.core.v1.objectMeta;
local deployment = k.apps.v1beta1.deployment;
local container = k.core.v1.container;
local service = k.core.v1.service;
local serviceAccount = k.core.v1.serviceAccount;
local configMap = k.core.v1.configMap;

local namespace = "kubeless";
local controller_account_name = "controller-acct";

local controllerEnv = [
  {
    name: "KUBELESS_INGRESS_ENABLED",
    valueFrom: {configMapKeyRef: {"name": "kubeless-config", key: "ingress-enabled"}}
  },
  {
    name: "KUBELESS_SERVICE_TYPE",
    valueFrom: {configMapKeyRef: {"name": "kubeless-config", key: "service-type"}}
    },
  {
     name: "KUBELESS_NAMESPACE",
     valueFrom: {fieldRef: {fieldPath: "metadata.namespace"}}
   },
   {
     name: "KUBELESS_CONFIG",
     value: "kubeless-config"
   },
];

local controllerContainer =
  container.default("kubeless-controller", "bitnami/kubeless-controller:latest") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local kubelessLabel = {kubeless: "controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kubeless-controller", controllerContainer, namespace) +
  {metadata+:{labels: kubelessLabel}} +
  {spec+: {selector: {matchLabels: kubelessLabel}}} +
  {spec+: {template+: {spec+: {serviceAccountName: controllerAccount.metadata.name}}}} +
  {spec+: {template+: {metadata: {labels: kubelessLabel}}}};

local crd = [
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("functions.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "functions", singular: "function", kind: "Function"}},
    description: "Kubernetes Native Serverless Framework",
  },
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("kafkatriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "kafkatriggers", singular: "kafkatrigger", kind: "KafkaTrigger"}},
    description: "CRD object for Kafka trigger type",  
  },
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("httptriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "httptriggers", singular: "httptrigger", kind: "HTTPTrigger"}},
    description: "CRD object for HTTP trigger type",  
  },
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("cronjobtriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "cronjobtriggers", singular: "cronjobtrigger", kind: "CronJobTrigger"}},
    description: "CRD object for HTTP trigger type",  
  }    
];

local deploymentConfig = '{}';

local runtime_images ='[
  {
    "ID": "python",
    "versions": [
      {
        "name": "python27",
        "version": "2.7",
        "runtimeImage": "andresmgot/python@sha256:0138bf2f1123b9c60d7ccb9a4dd189f10b48b15cc51f80cee4d13650b2235c6f",
        "initImage": "python:2.7"
      },
      {
        "name": "python34",
        "version": "3.4",
        "runtimeImage": "andresmgot/python@sha256:98d4c7a1c3e32d693f8d2d374cbf287acdcbe46dcceec27fafaafa2260482829",
        "initImage": "python:3.4"
      },
      {
        "name": "python36",
        "version": "3.6",
        "runtimeImage": "andresmgot/python@sha256:6e5115adf7b142429a3a038515b09648eeee364837173e55a117724b203da980",
        "initImage": "python:3.6"
      }
    ],
    "depName": "requirements.txt",
    "fileNameSuffix": ".py"
  },
  {
    "ID": "nodejs",
    "versions": [
      {
        "name": "node6",
        "version": "6",
        "runtimeImage": "andresmgot/nodejs@sha256:8a5bf9a7d1a14b6c5a38e31c5b927fb1f8b3ba2bb1efca8a9ca30e450b340249",
        "initImage": "node:6.10"
      },
      {
        "name": "node8",
        "version": "8",
        "runtimeImage": "andresmgot/nodejs@sha256:c39936e6d459461ee5679c3a844b9d81db28cd3e4190930b2cbf54a7e477ced1",
        "initImage": "node:8"
      }
    ],
    "depName": "package.json",
    "fileNameSuffix": ".js"
  },
  {
    "ID": "ruby",
    "versions": [
      {
        "name": "ruby24",
        "version": "2.4",
        "runtimeImage": "andresmgot/ruby@sha256:1e208b3b27eb51881fc6e5f271aa7d719a5274a8a767d0cb2c35ced313156078",
        "initImage": "bitnami/ruby:2.4"
      }
    ],
    "depName": "Gemfile",
    "fileNameSuffix": ".rb"
  },
  {
    "ID": "php",
    "versions": [
      {
        "name": "php72",
        "version": "7.2",
        "runtimeImage": "andresmgot/php@sha256:bf87850ae47234b7a46d979bea25ba4862ee8d7d928e951851ce8e94874b074b",
        "pubsubImage": "",
        "initImage": "composer:1.6"
      }
    ],
    "depName": "composer.json",
    "fileNameSuffix": ".php"
  }
]';

local kubelessConfig  = configMap.default("kubeless-config", namespace) +
    configMap.data({"ingress-enabled": "false"}) +
    configMap.data({"service-type": "ClusterIP"})+
    configMap.data({"deployment": std.toString(deploymentConfig)})+
    configMap.data({"runtime-images": std.toString(runtime_images)});

{
  controllerAccount: k.util.prune(controllerAccount),
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  cfg: k.util.prune(kubelessConfig),
}
