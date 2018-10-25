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

local functionControllerContainer =
  container.default("kubeless-function-controller", "kubeless/function-controller:latest") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local httpTriggerControllerContainer =
  container.default("http-trigger-controller", "bitnami/http-trigger-controller:v1.0.0-alpha.9") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local cronjobTriggerContainer =
  container.default("cronjob-trigger-controller", "bitnami/cronjob-trigger-controller:v1.0.0-alpha.9") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local kubelessLabel = {kubeless: "controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kubeless-controller-manager", [functionControllerContainer, httpTriggerControllerContainer, cronjobTriggerContainer], namespace) +
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
  },
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("httptriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "httptriggers", singular: "httptrigger", kind: "HTTPTrigger"}},
  },
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("cronjobtriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "cronjobtriggers", singular: "cronjobtrigger", kind: "CronJobTrigger"}},
  }
];

local deploymentConfig = '{}';

local runtime_images ='[
  {
    "ID": "python",
    "compiled": false,
    "versions": [
      {
        "name": "python27",
        "version": "2.7",
        "runtimeImage": "kubeless/python@sha256:34332f4530508a810f491838a924c36ceac0ec7cab487520e2db2b037800ecda",
        "initImage": "python:2.7"
      },
      {
        "name": "python34",
        "version": "3.4",
        "runtimeImage": "kubeless/python@sha256:5c93a60b83dba9324ad8358e66952232746ef9d477266d6a199617d7344c2053",
        "initImage": "python:3.4"
      },
      {
        "name": "python36",
        "version": "3.6",
        "runtimeImage": "kubeless/python@sha256:8c49bfa1c6aa5fbcd0f7d99d97280c161247fc94c06d26c04e39ac341c3f75e5",
        "initImage": "python:3.6"
      }
    ],
    "depName": "requirements.txt",
    "fileNameSuffix": ".py"
  },
  {
    "ID": "nodejs",
    "compiled": false,
    "versions": [
      {
        "name": "node6",
        "version": "6",
        "runtimeImage": "kubeless/nodejs@sha256:556ff930c7a609d1ad90322d41c8b562cb42313898486fed9674fb2647e4b42f",
        "initImage": "node:6.10"
      },
      {
        "name": "node8",
        "version": "8",
        "runtimeImage": "kubeless/nodejs@sha256:5c9c5e36f9845f2cf8e9e0d55993796d82e34a2b8c0f8a508c9d3c04b2041076",
        "initImage": "node:8"
      }
    ],
    "depName": "package.json",
    "fileNameSuffix": ".js"
  },
  {
    "ID": "nodejs_distroless",
    "compiled": false,
    "versions": [
      {
        "name": "node8",
        "version": "8",
        "runtimeImage": "henrike42/kubeless/runtimes/nodejs/distroless:0.0.2",
        "initImage": "node:8"
      }
    ],
    "depName": "package.json",
    "fileNameSuffix": ".js"
  },
  {
    "ID": "ruby",
    "compiled": false,
    "versions": [
      {
        "name": "ruby23",
        "version": "2.3",
        "runtimeImage": "kubeless/ruby@sha256:67870b57adebc4dc749a8a19795da801da2d05fc6e8324168ac1b227bb7c77f7",
        "initImage": "bitnami/ruby:2.3"
      },
      {
        "name": "ruby24",
        "version": "2.4",
        "runtimeImage": "kubeless/ruby@sha256:aaa9c7f3dfd4f866a527c04171c32dae2efa420d770a6af9c517771137ab4011",
        "initImage": "bitnami/ruby:2.4"
      },
      {
        "name": "ruby25",
        "version": "2.5",
        "runtimeImage": "kubeless/ruby@sha256:577e35724996ba340ff0a18366bce99586b0be58e4d27fa3e8038f977caf1559",
        "initImage": "bitnami/ruby:2.5"
      }
    ],
    "depName": "Gemfile",
    "fileNameSuffix": ".rb"
  },
  {
    "ID": "php",
    "compiled": false,
    "versions": [
      {
        "name": "php72",
        "version": "7.2",
        "runtimeImage": "kubeless/php@sha256:9b86066b2640bedcd88acb27f43dfaa2b338f0d74d9d91131ea781402f7ec8ec",
        "initImage": "composer:1.6"
      }
    ],
    "depName": "composer.json",
    "fileNameSuffix": ".php"
  },
  {
    "ID": "go",
    "compiled": true,
    "versions": [
      {
        "name": "go1.10",
        "version": "1.10",
        "runtimeImage": "kubeless/go@sha256:e2fd49f09b6ff8c9bac6f1592b3119ea74237c47e2955a003983e08524cb3ae5",
        "initImage": "kubeless/go-init@sha256:983b3f06452321a2299588966817e724d1a9c24be76cf1b12c14843efcdff502"
      }
    ],
    "depName": "Gopkg.toml",
    "fileNameSuffix": ".go"
  },
  {
    "ID": "dotnetcore",
    "compiled": true,
    "versions": [
      {
        "name": "dotnetcore2.0",
        "version": "2.0",
        "runtimeImage": "allantargino/kubeless-dotnetcore@sha256:1699b07d9fc0276ddfecc2f823f272d96fd58bbab82d7e67f2fd4982a95aeadc",
        "initImage": "allantargino/aspnetcore-build@sha256:0d60f845ff6c9c019362a68b87b3920f3eb2d32f847f2d75e4d190cc0ce1d81c"
      }
    ],
    "depName": "project.csproj",
    "fileNameSuffix": ".cs"
  },
  {
    "ID": "java",
    "compiled": true,
    "versions": [
      {
        "name": "java1.8",
        "version": "1.8",
        "runtimeImage": "kubeless/java@sha256:debf9502545f4c0e955eb60fabb45748c5d98ed9365c4a508c07f38fc7fefaac",
        "initImage": "kubeless/java-init@sha256:7e5e4376d3ab76c336d4830c9ed1b7f9407415feca49b8c2bf013e279256878f"
      }
    ],
    "depName": "pom.xml",
    "fileNameSuffix": ".java"
  },
  {
     "ID": "ballerina",
     "compiled": true,
     "versions": [
       {
          "name": "ballerina0.981.0",
          "version": "0.981.0",
          "runtimeImage": "ballerina/kubeless-ballerina@sha256:a025841010cfdf8136396efef31d4155283770d331ded6a9003e6e55f02db2e5",
          "initImage": "ballerina/kubeless-ballerina-init@sha256:a04ca9d289c62397d0b493876f6a9ff4cc425563a47aa7e037c3b850b8ceb3e8"
       }
     ],
     "depName": "",
     "fileNameSuffix": ".bal"
  },
  {
    "ID": "jvm",
    "compiled": true,
    "versions": [
      {
        "name": "jvm1.8",
        "version": "1.8",
        "runtimeImage": "caraboides/jvm@sha256:2870c4f48df4feb2ee7478a152b44840d781d4b1380ad3fa44b3c7ff314faded",
        "initImage": "caraboides/jvm-init@sha256:e57dbf3f56570a196d68bce1c0695102b2dbe3ae2ca6d1c704476a7a11542f1d"
      }
    ],
    "depName": "",
    "fileNameSuffix": ".jar"
  }
]';

local kubelessConfig  = configMap.default("kubeless-config", namespace) +
    configMap.data({"ingress-enabled": "false"}) +
    configMap.data({"service-type": "ClusterIP"})+
    configMap.data({"deployment": std.toString(deploymentConfig)})+
    configMap.data({"runtime-images": std.toString(runtime_images)})+
    configMap.data({"enable-build-step": "false"})+
    configMap.data({"function-registry-tls-verify": "true"})+
    configMap.data({"provision-image": "kubeless/unzip@sha256:f162c062973cca05459834de6ed14c039d45df8cdb76097f50b028a1621b3697"})+
    configMap.data({"provision-image-secret": ""})+
    configMap.data({"builder-image": "kubeless/function-image-builder:latest"})+
    configMap.data({"builder-image-secret": ""});

{
  controllerAccount: k.util.prune(controllerAccount),
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  cfg: k.util.prune(kubelessConfig),
}
