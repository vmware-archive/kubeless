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

local crd = {
  apiVersion: "apiextensions.k8s.io/v1beta1",
  kind: "CustomResourceDefinition",
  metadata: objectMeta.name("functions.kubeless.io"),
  spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "functions", singular: "function", kind: "Function"}},
  description: "Kubernetes Native Serverless Framework",
};

local deploymentConfig = '{}';

local runtime_images ='[
  {
    "ID": "python",
    "versions": [
      {
        "name": "python27",
        "version": "2.7",
        "httpImage": "kubeless/python@sha256:0f3b64b654df5326198e481cd26e73ecccd905aae60810fc9baea4dcbb61f697",
        "pubsubImage": "kubeless/python-event-consumer@sha256:1aeb6cef151222201abed6406694081db26fa2235d7ac128113dcebd8d73a6cb",
        "initImage": "tuna/python-pillow:2.7.11-alpine"
      },
      {
        "name": "python34",
        "version": "3.4",
        "httpImage": "kubeless/python@sha256:e502078dc9580bb73f823504a6765dfc98f000979445cdf071900350b938c292",
        "pubsubImage": "kubeless/python-event-consumer@sha256:d963e4cd58229d662188d618cd87503b3c749b126b359ce724a19a375e4b3040",
        "initImage": "python:3.4"
      },
      {
        "name": "python36",
        "version": "3.6",
        "httpImage": "kubeless/python@sha256:6300c2513ca51653ae698a31eacf6b2b8a16d2737dd3e244a8c9c11f6408fd35",
        "pubsubImage": "kubeless/python-event-consumer@sha256:0a2f9162de56b7966b02b70a5a0bcff03badfd9d87b8ae3d13e5381abd00220f",
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
        "httpImage": "kubeless/nodejs@sha256:2b25d7380d6ed06ad817f4ee1e177340a282788596b34464173bb8a967d83c02",
        "pubsubImage": "kubeless/nodejs-event-consumer@sha256:1861c32d6a46b2fdfc3e3996daf690ff2c3d5ca19a605abd2af503011d68e221",
        "initImage": "node:6.10"
      },
      {
        "name": "node8",
        "version": "8",
        "httpImage": "kubeless/nodejs@sha256:f1426efe274ea8480d95270c98f6007ac64645e36291dbfa36d759b5c8b7b733",
        "pubsubImage": "kubeless/nodejs-event-consumer@sha256:b301b02e463b586d9a32d5c1cb5a68c2a11e4fba9514e28d900fc50a78759af9",
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
        "httpImage": "kubeless/ruby@sha256:738e4cdeb5f5feece236bbf4e46902024e4b9fc16db4f3791404fa27e8b0db15",
        "pubsubImage": "kubeless/ruby-event-consumer@sha256:f9f50be51d93a98ae30689d87b067c181905a8757d339fb0fa9a81c6268c4eea",
        "initImage": "bitnami/ruby:2.4"
      }
    ],
    "depName": "Gemfile",
    "fileNameSuffix": ".rb"
  },
  {
    "ID": "dotnetcore",
    "versions": [
      {
        "name": "dotnetcore2",
        "version": "2.0",
        "httpImage": "allantargino/kubeless-dotnetcore@sha256:d321dc4b2c420988d98cdaa22c733743e423f57d1153c89c2b99ff0d944e8a63",
        "pubsubImage": "kubeless/ruby-event-consumer@sha256:f9f50be51d93a98ae30689d87b067c181905a8757d339fb0fa9a81c6268c4eea",
        "initImage": "microsoft/aspnetcore-build:2.0"
      }
    ],
    "depName": "requirements.xml",
    "fileNameSuffix": ".cs"
  },
    {
    "ID": "php",
    "versions": [
      {
        "name": "php72",
        "version": "7.2",
        "httpImage": "paolomainardi/kubeless-php@sha256:34285a739b6fa0730b8ac349740d20305ae629095dab85cfb10b874f6a14fe45",
        "pubsubImage": "",
        "initImage": "composer:1.6"
      }
    ],
    "depName": "composer.json",
    "fileNameSuffix": ".php"
  }
]';

local provision_container_image = '{
  Image: "kubeless/unzip@sha256:f162c062973cca05459834de6ed14c039d45df8cdb76097f50b028a1621b3697"
}';

local kubelessConfig  = configMap.default("kubeless-config", namespace) +
    configMap.data({"ingress-enabled": "false"}) +
    configMap.data({"service-type": "ClusterIP"})+
    configMap.data({"deployment": std.toString(deploymentConfig)})+
    configMap.data({"runtime-images": std.toString(runtime_images)})+
    configMap.data({"provision-container-image": std.toString(provision_container_image)});

{
  controllerAccount: k.util.prune(controllerAccount),
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  cfg: k.util.prune(kubelessConfig),
}
