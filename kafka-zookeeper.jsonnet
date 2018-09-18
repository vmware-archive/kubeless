local k = import "ksonnet.beta.1/k.libsonnet";

local statefulset = k.apps.v1beta1.statefulSet;
local container = k.core.v1.container;
local service = k.core.v1.service;
local deployment = k.apps.v1beta1.deployment;
local serviceAccount = k.core.v1.serviceAccount;
local objectMeta = k.core.v1.objectMeta;

local namespace = "kubeless";
local controller_account_name = "controller-acct";

local crd = [
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("kafkatriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "kafkatriggers", singular: "kafkatrigger", kind: "KafkaTrigger"}},
  },
];

local controllerContainer =
  container.default("kafka-trigger-controller", "bitnami/kafka-trigger-controller:v1.0.0-alpha.9") +
  container.imagePullPolicy("IfNotPresent");

local kubelessLabel = {kubeless: "kafka-trigger-controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kafka-trigger-controller", controllerContainer, namespace) +
  {metadata+:{labels: kubelessLabel}} +
  {spec+: {selector: {matchLabels: kubelessLabel}}} +
  {spec+: {template+: {spec+: {serviceAccountName: controllerAccount.metadata.name}}}} +
  {spec+: {template+: {metadata: {labels: kubelessLabel}}}};

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
    name: "KAFKA_DELETE_TOPIC_ENABLE",
    value: "true"
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
  container.default("broker", "bitnami/kafka:1.1.0-r0") +
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

local kafkaInitContainer =
  container.default("volume-permissions", "busybox") +
  container.imagePullPolicy("IfNotPresent") +
  container.command(["sh", "-c", "chmod -R g+rwX /bitnami"]) +
  container.volumeMounts([
    {
      name: "datadir",
      mountPath: "/bitnami/kafka/data"
    }
  ]);

local zookeeperContainer =
  container.default("zookeeper", "bitnami/zookeeper:3.4.10-r12") +
  container.imagePullPolicy("IfNotPresent") +
  container.env(zookeeperEnv) +
  container.ports(zookeeperPorts) +
  container.volumeMounts([
    {
      name: "zookeeper",
      mountPath: "/bitnami/zookeeper"
    }
  ]);

local zookeeperInitContainer =
  container.default("volume-permissions", "busybox") +
  container.imagePullPolicy("IfNotPresent") +
  container.command(["sh", "-c", "chmod -R g+rwX /bitnami"]) +
  container.volumeMounts([
    {
      name: "zookeeper",
      mountPath: "/bitnami/zookeeper"
    }
  ]);

local kafkaLabel = {kubeless: "kafka"};
local zookeeperLabel = {kubeless: "zookeeper"};

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

local zooVolumeCT = [
  {
    "metadata": {
      "name": "zookeeper"
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
  {spec+: {template+: {spec: {containers: [kafkaContainer], initContainers: [kafkaInitContainer]}}}};

local zookeeperSts =
  statefulset.default("zoo", namespace) +
  statefulset.spec({serviceName: "zoo"}) +
  {spec+: {template: {metadata: {labels: zookeeperLabel}}}} +
  {spec+: {volumeClaimTemplates: zooVolumeCT}} +
  {spec+: {template+: {spec: {containers: [zookeeperContainer], initContainers: [zookeeperInitContainer]}}}};

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

local controller_roles = [
  {
    apiGroups: [""],
    resources: ["services", "configmaps"],
    verbs: ["get", "list"],
  },
  {
    apiGroups: ["kubeless.io"],
    resources: ["functions", "kafkatriggers"],
    verbs: ["get", "list", "watch", "update", "delete"],
  },
];

local clusterRole(name, rules) = {
    apiVersion: "rbac.authorization.k8s.io/v1beta1",
    kind: "ClusterRole",
    metadata: objectMeta.name(name),
    rules: rules,
};

local clusterRoleBinding(name, role, subjects) = {
    apiVersion: "rbac.authorization.k8s.io/v1beta1",
    kind: "ClusterRoleBinding",
    metadata: objectMeta.name(name),
    subjects: [{kind: s.kind, namespace: s.metadata.namespace, name: s.metadata.name} for s in subjects],
    roleRef: {kind: role.kind, apiGroup: "rbac.authorization.k8s.io", name: role.metadata.name},
};

local controllerClusterRole = clusterRole(
  "kafka-controller-deployer", controller_roles);

local controllerClusterRoleBinding = clusterRoleBinding(
  "kafka-controller-deployer", controllerClusterRole, [controllerAccount]
);

{
  kafkaSts: k.util.prune(kafkaSts),
  zookeeperSts: k.util.prune(zookeeperSts),
  kafkaSvc: k.util.prune(kafkaSvc),
  kafkaHeadlessSvc: k.util.prune(kafkaHeadlessSvc),
  zookeeperSvc: k.util.prune(zookeeperSvc),
  zookeeperHeadlessSvc: k.util.prune(zookeeperHeadlessSvc),
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  controllerClusterRole: k.util.prune(controllerClusterRole),
  controllerClusterRoleBinding: k.util.prune(controllerClusterRoleBinding),
}
