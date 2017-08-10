# Remove volumeClaimTemplates from kafkaSts to enable testing kubeless
# on simple clusters deploys like kubeadm-dind-cluster
local kubeless_rbac = import "kubeless-rbac.jsonnet";

kubeless_rbac + {
  kafkaSts+:
   {spec+: {volumeClaimTemplates: []}} +
   {spec+: {template+: {spec+: {volumes: [{name: "datadir", emptyDir: {}}]}}}}
}
