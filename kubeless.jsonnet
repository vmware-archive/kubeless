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
  container.default("kubeless-controller-manager", "bitnami/kubeless-controller-manager:latest") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local kubelessLabel = {kubeless: "controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kubeless-controller-manager", controllerContainer, namespace) +
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
        "runtimeImage": "andresmgot/python@sha256:cdba8db2ab40c2e1b954f133c4b3c2dc8f63044e490bcccee88621a61d2d06d5",
        "initImage": "python:2.7"
      },
      {
        "name": "python34",
        "version": "3.4",
        "runtimeImage": "andresmgot/python@sha256:f5adcac7eeb24689786f600c7d9b6620b43f6c1027588e275360349b931d086b",
        "initImage": "python:3.4"
      },
      {
        "name": "python36",
        "version": "3.6",
        "runtimeImage": "andresmgot/python@sha256:7874ec2b71008fed0a0b18d4f907b7375452bbaf7ee8c000c5a6d18a93260ff7",
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
        "runtimeImage": "andresmgot/nodejs@sha256:a9d8b31c403a014d83e8820868421750f7513a6897a430454abfd0a76a22bb71",
        "initImage": "node:6.10"
      },
      {
        "name": "node8",
        "version": "8",
        "runtimeImage": "andresmgot/nodejs@sha256:7138b35e510ab824b6d453af0b6903b4f4c1e3b81ac8a2b69447306a79ba12a4",
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
        "runtimeImage": "andresmgot/ruby@sha256:a5fdab9ab42a973971f9902f0af8a5fa233d728bc786ce6c1d5e22094f04099d",
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
        "runtimeImage": "andresmgot/php@sha256:2abb820ad23b59aa9ebdd9457a6596ffd3f8bc0ddc9d91bb72b2a46a3069d82c",
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
