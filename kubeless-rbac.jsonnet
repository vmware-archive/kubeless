# Add RBAC role and binding on top of kubeless.jsonnet, to allow
# kubeless controller to deploy/update/etc functions on any namespace
local k = import "ksonnet.beta.1/k.libsonnet";
local objectMeta = k.core.v1.objectMeta;

local kubeless = import "kubeless.jsonnet";
local controller_account = kubeless.controller_account;
local controller_roles = [{
  apiGroups: ["*"],
  resources: ["services", "deployments", "functions", "configmaps"],
  verbs: ["*"]
}];

local controllerAccount = kubeless.controllerAccount;

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
    subjects: [s + {namespace: s.metadata.namespace, name: s.metadata.name} for s in subjects],
    roleRef: role + {name: role.metadata.name},
};

local controllerClusterRole = clusterRole(
  "kubeless-controller-deployer", controller_roles);

local controllerClusterRoleBinding = clusterRoleBinding(
  "kubeless-controller-deployer", controllerClusterRole, [controllerAccount]
);

kubeless + {
  controllerClusterRole: controllerClusterRole,
  controllerClusterRoleBinding: controllerClusterRoleBinding,
}
