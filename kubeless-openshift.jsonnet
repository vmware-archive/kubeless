# Builds on kubeless-rbac.ksonnet to produce a deployable manifest on OpenShift 1.5
# Modifies apiVersion for kubeless-controller Deployment to extensions/v1beta1
# Modifies ClusterRole and ClusterRoleBinding apiVersions to v1
local kubeless = import "kubeless-rbac.jsonnet";

kubeless + {
  controller: kubeless.controller + { apiVersion: "extensions/v1beta1" },
  controllerClusterRole: kubeless.controllerClusterRole + { apiVersion: "v1" },
  controllerClusterRoleBinding: kubeless.controllerClusterRoleBinding + { apiVersion: "v1" },
  kafkaSts: kubeless.kafkaSts  + {spec+: {template+: {spec+: { initContainers: [] }}}},
  zookeeperSts: kubeless.zookeeperSts  + {spec+: {template+: {spec+: { initContainers: [] }}}}
}
