local k = import "ksonnet.beta.1/k.libsonnet";
local util = import "ksonnet.beta.1/util.libsonnet";

local objectMeta = k.core.v1.objectMeta;
local deployment = k.apps.v1beta1.deployment;
local statefulset = k.apps.v1beta1.statefulSet;
local container = k.core.v1.container;
local service = k.core.v1.service;

local namespace = "kubeless";

local controllerContainer =
  container.default("kubeless-controller", "bitnami/kubeless-controller:0.0.13") +
  container.imagePullPolicy("IfNotPresent");

local kafkaEnv = [
  {
    name: "KAFKA_ADVERTISED_HOST_NAME",
    value: "broker.kubeless"
  },
  {
    name: "KAFKA_ADVERTISED_PORT",
    value: "9092"
  },
  {
    name: "KAFKA_PORT",
    value: "9092"
  },
  {
    name: "KAFKA_ZOOKEEPER_CONNECT",
    value: "zookeeper.kubeless:2181"
  }
];

local zookeeperEnv = [
  {
    name: "ZOO_SERVERS",
    value: "server.1=zoo-0.zoo:2888:3888:participant"
  }
];

local zookeeperPorts = [
  {
    containerPort: 2181,
    name: "client"
  },
  {
    containerPort: 2888,
    name: "peer"
  },
  {
    containerPort: 3888,
    name: "leader-election"
  }
];

local kafkaContainer =
  container.default("broker", "bitnami/kafka@sha256:9ef14a3a2348ae24072c73caa4d4db06c77a8c0383a726d02244ea0e43723355") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(kafkaEnv) +
  container.ports({containerPort: 9092}) +
  container.volumeMounts([
    {
      name: "datadir",
      mountPath: "/opt/bitnami/kafka/data"
    }
  ]);

local zookeeperContainer =
  container.default("zookeeper", "bitnami/zookeeper@sha256:0bbf6503e45fc7d5236513987702b0533d2c777144dfd5022feed9ad89dc6318") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(zookeeperEnv) +
  container.ports(zookeeperPorts) +
  container.volumeMounts([
    {
      name: "zookeeper",
      mountPath: "/bitnami/zookeeper"
    }
  ]);

local kubelessLabel = {kubeless: "controller"};
local kafkaLabel = {app: "kafka"};
local zookeeperLabel = {app: "zookeeper"};

local controllerDeployment =
  deployment.default("kubeless-controller", controllerContainer, namespace) +
  {spec+: {template+: {metadata: {labels: kubelessLabel}}}};

local kafkaVolumeCT = [
  {
    "metadata": {
      "name": "datadir"
    },
    "spec": {
      "accessModes": [
        "ReadWriteOnce"
      ],
      "resources": {
        "requests": {
          "storage": "1Gi"
        }
      }
    }
  }
];

local kafkaSts =
  statefulset.default("kafka", namespace) +
  statefulset.spec({serviceName: "broker"}) +
  {spec+: {template+: {metadata: {labels: kafkaLabel}}}} +
  {spec+: {volumeClaimTemplates: kafkaVolumeCT}} +
  {spec+: {template+: {spec: {containers: [kafkaContainer]}}}};

local zookeeperSts =
  statefulset.default("zookeeper", namespace) +
  statefulset.spec({serviceName: "zoo"}) +
  {spec+: {template+: {metadata: {labels: zookeeperLabel}}}} +
  {spec+: {template+: {spec: {containers: [zookeeperContainer], volumes: [{name: "zookeeper", emptyDir: {}}]}}}};

local kafkaSvc =
  service.default("kafka", namespace) +
  service.mixin.spec.ports({port: 9092}) +
  service.mixin.spec.selector({app: "kafka"});

local kafkaHeadlessSvc =
  service.default("broker", namespace) +
  service.mixin.spec.ports({port: 9092}) +
  service.mixin.spec.selector({app: "kafka"}) +
  {spec+: {clusterIP: "None"}};

local zookeeperSvc =
  service.default("zookeeper", namespace) +
  service.mixin.spec.ports({port: 2181, name: "client"}) +
  service.mixin.spec.selector({app: "zookeeper"});

local zookeeperHeadlessSvc =
  service.default("zoo", namespace) +
  service.mixin.spec.ports([{port: 9092, name: "peer"},{port: 3888, name: "leader-election"}]) +
  service.mixin.spec.selector({app: "zookeeper"}) +
  {spec+: {clusterIP: "None"}};

local tpr = {
  apiVersion: "extensions/v1beta1",
  kind: "ThirdPartyResource",
  metadata: objectMeta.name("function.k8s.io"),
  versions: [{name: "v1alpha1"}],
  description: "Kubernetes Native Serverless Framework",
};

{
  controller: util.prune(controllerDeployment),
  kafkaSts: util.prune(kafkaSts),
  zookeeperSts: util.prune(zookeeperSts),
  kafkaSvc: util.prune(kafkaSvc),
  kafkaHeadlessSvc: util.prune(kafkaHeadlessSvc),
  zookeeperSvc: util.prune(zookeeperSvc),
  zookeeperHeadlessSvc: util.prune(zookeeperHeadlessSvc),
  tpr: util.prune(tpr),
}
