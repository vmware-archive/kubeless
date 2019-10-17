local k = import "ksonnet.beta.1/k.libsonnet";
local runtimesSrc = import "runtimes.jsonnet";

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
  container.default("http-trigger-controller", "kubeless/http-trigger-controller:v1.0.1") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local cronjobTriggerContainer =
  container.default("cronjob-trigger-controller", "kubeless/cronjob-trigger-controller:v1.0.1") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(controllerEnv);

local kubelessLabel = {kubeless: "controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kubeless-controller-manager", [functionControllerContainer, httpTriggerControllerContainer, cronjobTriggerContainer], namespace) +
  {apiVersion: "apps/v1"} +
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

local kubelessConfig  = configMap.default("kubeless-config", namespace) +
    configMap.data({"ingress-enabled": "false"}) +
    configMap.data({"service-type": "ClusterIP"})+
    configMap.data({"deployment": std.toString(deploymentConfig)})+
    configMap.data({"runtime-images": std.toString(runtimesSrc)})+
    configMap.data({"enable-build-step": "false"})+
    configMap.data({"function-registry-tls-verify": "true"})+
    configMap.data({"provision-image": "kubeless/unzip@sha256:4863100364496255de9bf8722a220dad7143ab277ac72435876eb8c93936e9d7"})+
    configMap.data({"provision-image-secret": ""})+
    configMap.data({"builder-image": "kubeless/function-image-builder:latest"})+
    configMap.data({"builder-image-secret": ""});

{
  controllerAccount: k.util.prune(controllerAccount),
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  cfg: k.util.prune(kubelessConfig),
}
