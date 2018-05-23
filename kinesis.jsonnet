local k = import "ksonnet.beta.1/k.libsonnet";
local container = k.core.v1.container;

local deployment = k.apps.v1beta1.deployment;
local serviceAccount = k.core.v1.serviceAccount;
local objectMeta = k.core.v1.objectMeta;

local namespace = "kubeless";
local controller_account_name = "controller-acct";

local crd = [
  {
    apiVersion: "apiextensions.k8s.io/v1beta1",
    kind: "CustomResourceDefinition",
    metadata: objectMeta.name("kinesistriggers.kubeless.io"),
    spec: {group: "kubeless.io", version: "v1beta1", scope: "Namespaced", names: {plural: "kinesistriggers", singular: "kinesistrigger", kind: "KinesisTrigger"}},
    description: "CRD object for Kinesis trigger type",
  },
];

local controllerContainer =
  container.default("kinesis-trigger-controller", "bitnami/kinesis-trigger-controller:latest") +
  container.imagePullPolicy("IfNotPresent");

local kubelessLabel = {kubeless: "kinesis-trigger-controller"};

local controllerAccount =
  serviceAccount.default(controller_account_name, namespace);

local controllerDeployment =
  deployment.default("kinesis-trigger-controller", controllerContainer, namespace) +
  {metadata+:{labels: kubelessLabel}} +
  {spec+: {selector: {matchLabels: kubelessLabel}}} +
  {spec+: {template+: {spec+: {serviceAccountName: controllerAccount.metadata.name}}}} +
  {spec+: {template+: {metadata: {labels: kubelessLabel}}}};

local controller_roles = [
  {
    apiGroups: [""],
    resources: ["services", "configmaps", "secrets"],
    verbs: ["get", "list"],
  },
  {
    apiGroups: ["kubeless.io"],
    resources: ["functions", "kinesistriggers"],
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
  "kinesis-controller-deployer", controller_roles);

local controllerClusterRoleBinding = clusterRoleBinding(
  "kinesis-controller-deployer", controllerClusterRole, [controllerAccount]
);

{
  controller: k.util.prune(controllerDeployment),
  crd: k.util.prune(crd),
  controllerClusterRole: k.util.prune(controllerClusterRole),
  controllerClusterRoleBinding: k.util.prune(controllerClusterRoleBinding),
}
