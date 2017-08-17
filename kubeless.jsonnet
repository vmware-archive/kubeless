local k = import "ksonnet.beta.1/k.libsonnet";

local objectMeta = k.core.v1.objectMeta;
local deployment = k.apps.v1beta1.deployment;
local statefulset = k.apps.v1beta1.statefulSet;
local container = k.core.v1.container;
local service = k.core.v1.service;
local serviceAccount = k.core.v1.serviceAccount;

local namespace = "kubeless";
local controller_account_name = "controller-acct";

local controllerContainer =
  container.default("kubeless-controller", "bitnami/kubeless-controller@sha256:d07986d575a80179ae15205c6fa5eb3bf9f4f4f46235c79ad6b284e5d3df22d0") +
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
  },
  {
    name: "ALLOW_PLAINTEXT_LISTENER",
    value: "yes"
  }
];

local zookeeperEnv = [
  {
    name: "ZOO_SERVERS",
    value: "server.1=zoo-0.zoo:2888:3888:participant"
  },
  {
    name: "ALLOW_ANONYMOUS_LOGIN",
    value: "yes"
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
  container.default("broker", "bitnami/kafka@sha256:ef0b1332408c0361d457852622d3a180f3609b9d98f1a85a9a809adaecfe9b52") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(kafkaEnv) +
  container.ports({containerPort: 9092}) +
  container.livenessProbe({tcpSocket: {port: 9092}, initialDelaySeconds: 30}) +
  container.volumeMounts([
    {
      name: "datadir",
      mountPath: "/bitnami/kafka/data"
    }
  ]);

local zookeeperContainer =
  container.default("zookeeper", "bitnami/zookeeper@sha256:f66625a8a25070bee18fddf42319ec58f0c49c376b19a5eb252e6a4814f07123") +
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
local kafkaLabel = {kubeless: "kafka"};
local zookeeperLabel = {kubeless: "zookeeper"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kubeless-controller", controllerContainer, namespace) +
  {metadata+:{labels: kubelessLabel}} +
  {spec+: {selector: {matchLabels: kubelessLabel}}} +
  {spec+: {template+: {spec+: {serviceAccountName: controllerAccount.metadata.name}}}} +
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
  {spec+: {template: {metadata: {labels: kafkaLabel}}}} +
  {spec+: {volumeClaimTemplates: kafkaVolumeCT}} +
  {spec+: {template+: {spec: {containers: [kafkaContainer]}}}};

local zookeeperSts =
  statefulset.default("zoo", namespace) +
  statefulset.spec({serviceName: "zoo"}) +
  {spec+: {template: {metadata: {labels: zookeeperLabel}}}} +
  {spec+: {template+: {spec: {containers: [zookeeperContainer], volumes: [{name: "zookeeper", emptyDir: {}}]}}}};

local kafkaSvc =
  service.default("kafka", namespace) +
  service.spec(k.core.v1.serviceSpec.default()) +
  service.mixin.spec.ports({port: 9092}) +
  service.mixin.spec.selector({kubeless: "kafka"});

local kafkaHeadlessSvc =
  service.default("broker", namespace) +
  service.spec(k.core.v1.serviceSpec.default()) +
  service.mixin.spec.ports({port: 9092}) +
  service.mixin.spec.selector({kubeless: "kafka"}) +
  {spec+: {clusterIP: "None"}};

local zookeeperSvc =
  service.default("zookeeper", namespace) +
  service.spec(k.core.v1.serviceSpec.default()) +
  service.mixin.spec.ports({port: 2181, name: "client"}) +
  service.mixin.spec.selector({kubeless: "zookeeper"});

local zookeeperHeadlessSvc =
  service.default("zoo", namespace) +
  service.spec(k.core.v1.serviceSpec.default()) +
  service.mixin.spec.ports([{port: 9092, name: "peer"},{port: 3888, name: "leader-election"}]) +
  service.mixin.spec.selector({kubeless: "zookeeper"}) +
  {spec+: {clusterIP: "None"}};

local tpr = {
  apiVersion: "extensions/v1beta1",
  kind: "ThirdPartyResource",
  metadata: objectMeta.name("function.k8s.io"),
  versions: [{name: "v1"}],
  description: "Kubernetes Native Serverless Framework",
};

{
  controllerAccount: k.util.prune(controllerAccount),
  controller: k.util.prune(controllerDeployment),
  kafkaSts: k.util.prune(kafkaSts),
  zookeeperSts: k.util.prune(zookeeperSts),
  kafkaSvc: k.util.prune(kafkaSvc),
  kafkaHeadlessSvc: k.util.prune(kafkaHeadlessSvc),
  zookeeperSvc: k.util.prune(zookeeperSvc),
  zookeeperHeadlessSvc: k.util.prune(zookeeperHeadlessSvc),
  tpr: k.util.prune(tpr),
}
