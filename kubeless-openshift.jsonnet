# Builds on kubeless.ksonnet to produce a deployable manifest on OpenShift 1.5
# Modifies apiVersion for kubeless-controller Deployment to extensions/v1beta1
# Modifies ClusterRole and ClusterRoleBinding apiVersions to v1
local k = import "ksonnet.beta.1/k.libsonnet";
local kubeless = import "kubeless.jsonnet";

local config = kubeless.cfg + k.core.v1.configMap.data({"deployment":'{"spec":{"template":{"spec":{"securityContext":{}}}}}'});

kubeless + {
  controller: kubeless.controller + { apiVersion: "extensions/v1beta1" },
  controllerClusterRole: kubeless.controllerClusterRole + { apiVersion: "v1" },
  controllerClusterRoleBinding: kubeless.controllerClusterRoleBinding + { apiVersion: "v1" },
  cfg: config,
}
