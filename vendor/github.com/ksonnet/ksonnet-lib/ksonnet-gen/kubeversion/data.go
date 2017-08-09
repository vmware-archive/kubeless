package kubeversion

//-----------------------------------------------------------------------------
// Kubernetes version-specific data for customizing code that's
// emitted.
//-----------------------------------------------------------------------------

var versions = map[string]versionData{
	"v1.7.0": versionData{
		idAliases: map[string]string{
			// Properties of objects. Stuff like `cinder.volumeId`.
			"hostIPC":                         "hostIpc",
			"hostPID":                         "hostPid",
			"targetCPUUtilizationPercentage":  "targetCpuUtilizationPercentage",
			"externalID":                      "externalId",
			"podCIDR":                         "podCidr",
			"providerID":                      "providerId",
			"bootID":                          "bootId",
			"machineID":                       "machineId",
			"systemUUID":                      "systemUuid",
			"volumeID":                        "volumeId",
			"diskURI":                         "diskUri",
			"targetWWNs":                      "targetWwns",
			"datasetUUID":                     "datasetUuid",
			"pdID":                            "pdId",
			"scaleIO":                         "scaleIo",
			"podIP":                           "podIp",
			"hostIP":                          "hostIp",
			"clusterIP":                       "clusterIp",
			"externalIPs":                     "externalIps",
			"loadBalancerIP":                  "loadBalancerIp",
			"containerID":                     "containerId",
			"imageID":                         "imageId",
			"serverAddressByClientCIDRs":      "serverAddressByClientCidrs",
			"clientCIDR":                      "clientCidr",
			"nonResourceURLs":                 "nonResourceUrls",
			"currentCPUUtilizationPercentage": "currentCpuUtilizationPercentage",
			"downwardAPI":                     "downwardApi",

			// Types. These have capitalized first letters, and exist in
			// places like `core.v1.AWSElasticBlockStoreVolumeSource`.
			"AWSElasticBlockStoreVolumeSource": "awsElasticBlockStoreVolumeSource",
			"CephFSVolumeSource":               "cephFsVolumeSource",
			"DownwardAPIProjection":            "downwardApiProjection",
			"DownwardAPIVolumeFile":            "downwardApiVolumeFile",
			"DownwardAPIVolumeSource":          "downwardApiVolumeSource",
			"FCVolumeSource":                   "fcVolumeSource",
			"GCEPersistentDiskVolumeSource":    "gcePersistentDiskVolumeSource",
			"HTTPGetAction":                    "httpGetAction",
			"HTTPHeader":                       "httpHeader",
			"ISCSIVolumeSource":                "iscsiVolumeSource",
			"NFSVolumeSource":                  "nfsVolumeSource",
			"RBDVolumeSource":                  "rbdVolumeSource",
			"SELinuxOptions":                   "seLinuxOptions",
			"ScaleIOVolumeSource":              "scaleIoVolumeSource",
			"TCPSocketAction":                  "tcpSocketAction",
			"APIVersion":                       "apiVersion",
			"FSGroupStrategyOptions":           "fsGroupStrategyOptions",
			"HTTPIngressPath":                  "httpIngressPath",
			"HTTPIngressRuleValue":             "httpIngressRuleValue",
			"IDRange":                          "idRange",
			"IngressTLS":                       "ingressTls",
			"SELinuxStrategyOptions":           "seLinuxStrategyOptions",
			"APIGroup":                         "apiGroup",
			"APIGroupList":                     "apiGroupList",
			"APIResource":                      "apiResource",
			"APIResourceList":                  "apiResourceList",
			"APIVersions":                      "apiVersions",
			"ServerAddressByClientCIDR":        "serverAddressByClientCidr",
		},
		constructorSpecs: map[string][]CustomConstructorSpec{
			"io.k8s.kubernetes.pkg.api.v1.Container": []CustomConstructorSpec{
				newConstructor("new", newParam("name"), newParam("image")),
			},
			"io.k8s.kubernetes.pkg.api.v1.ContainerPort": []CustomConstructorSpec{
				newConstructor("new", newParam("containerPort")),
				newConstructor("newNamed", newParam("name"), newParam("containerPort")),
			},
			"io.k8s.kubernetes.pkg.api.v1.EnvVar": []CustomConstructorSpec{
				newConstructor("new", newParam("name"), newParam("value")),
				newConstructor(
					"fromSecretRef",
					newParam("name"),
					newParamNestedRef("secretRefName", "mixin.valueFrom.secretKeyRef.name"),
					newParamNestedRef("secretRefKey", "mixin.valueFrom.secretKeyRef.key")),
				newConstructor(
					"fromFieldPath",
					newParam("name"),
					newParamNestedRef("fieldPath", "mixin.valueFrom.fieldRef.fieldPath")),
			},
			"io.k8s.kubernetes.pkg.api.v1.KeyToPath": []CustomConstructorSpec{
				newConstructor("new", newParam("key"), newParam("path")),
			},
			"io.k8s.kubernetes.pkg.api.v1.Service": []CustomConstructorSpec{
				newConstructor(
					"new",
					newParamNestedRef("name", "mixin.metadata.name"),
					newParamNestedRef("selector", "mixin.spec.selector"),
					newParamNestedRef("ports", "mixin.spec.ports")),
			},
			"io.k8s.kubernetes.pkg.api.v1.ServicePort": []CustomConstructorSpec{
				newConstructor("new", newParam("port"), newParam("targetPort")),
				newConstructor("newNamed", newParam("name"), newParam("port"), newParam("targetPort")),
			},
			"io.k8s.kubernetes.pkg.api.v1.Volume": []CustomConstructorSpec{
				newConstructor(
					"fromConfigMap",
					newParam("name"),
					newParamNestedRef("configMapName", "mixin.configMap.name"),
					newParamNestedRef("configMapItems", "mixin.configMap.items")),
				newConstructor(
					"fromEmptyDir",
					newParam("name"),
					newParamNestedRefDefault("emptyDir", "mixin.emptyDir.mixinInstance", "{}")),
				newConstructor(
					"fromPersistentVolumeClaim",
					newParam("name"),
					newParamNestedRef("claimName", "mixin.persistentVolumeClaim.claimName")),
				newConstructor(
					"fromHostPath",
					newParam("name"),
					newParamNestedRef("hostPath", "mixin.hostPath.path")),
			},
			"io.k8s.kubernetes.pkg.api.v1.VolumeMount": []CustomConstructorSpec{
				newConstructor("new", newParam("name"), newParam("mountPath"), newParamWithDefault("readOnly", "false")),
			},
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.Deployment": []CustomConstructorSpec{
				newConstructor(
					"new",
					newParamNestedRef("name", "mixin.metadata.name"),
					newParamNestedRef("replicas", "mixin.spec.replicas"),
					newParamNestedRef("containers", "mixin.spec.template.spec.containers"),
					newParamNestedRefDefault("podLabels", "mixin.spec.template.metadata.labels", "{}")),
			},
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.Deployment": []CustomConstructorSpec{
				newConstructor(
					"new",
					newParamNestedRef("name", "mixin.metadata.name"),
					newParamNestedRef("replicas", "mixin.spec.replicas"),
					newParamNestedRef("containers", "mixin.spec.template.spec.containers"),
					newParamNestedRefDefault("podLabels", "mixin.spec.template.metadata.labels", "{}")),
			},
		},
		propertyBlacklist: map[string]propertySet{
			// Metadata fields.
			"io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta": newPropertySet(
				"creationTimestamp", "deletionTimestamp", "generation",
				"ownerReferences", "resourceVersion", "selfLink", "uid",
			),

			// Fields whose types are
			// `io.k8s.apimachinery.pkg.apis.meta.v1.ListMeta`.
			"io.k8s.kubernetes.pkg.api.v1.ComponentStatusList":                              newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.ConfigMapList":                                    newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.EndpointsList":                                    newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.EventList":                                        newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.LimitRangeList":                                   newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.NamespaceList":                                    newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.NodeList":                                         newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.PersistentVolumeClaimList":                        newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.PersistentVolumeList":                             newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.PodList":                                          newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.PodTemplateList":                                  newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.ReplicationControllerList":                        newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.ResourceQuotaList":                                newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.SecretList":                                       newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.ServiceAccountList":                               newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.api.v1.ServiceList":                                      newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentList":                        newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.StatefulSetList":                       newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.autoscaling.v1.HorizontalPodAutoscalerList":         newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.autoscaling.v2alpha1.HorizontalPodAutoscalerList":   newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.batch.v1.JobList":                                   newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.batch.v2alpha1.CronJobList":                         newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.certificates.v1beta1.CertificateSigningRequestList": newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.DaemonSetList":                   newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.DeploymentList":                  newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.IngressList":                     newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.NetworkPolicyList":               newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.PodSecurityPolicyList":           newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.ReplicaSetList":                  newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.ThirdPartyResourceList":          newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudgetList":             newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1alpha1.ClusterRoleBindingList":               newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1alpha1.ClusterRoleList":                      newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1alpha1.RoleBindingList":                      newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1alpha1.RoleList":                             newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1beta1.ClusterRoleBindingList":                newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1beta1.ClusterRoleList":                       newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1beta1.RoleBindingList":                       newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.rbac.v1beta1.RoleList":                              newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.settings.v1alpha1.PodPresetList":                    newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.storage.v1.StorageClassList":                        newPropertySet("metadata"),
			"io.k8s.kubernetes.pkg.apis.storage.v1beta1.StorageClassList":                   newPropertySet("metadata"),

			// Status fields.
			"io.k8s.kubernetes.pkg.api.v1.Namespace":                                    newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.Node":                                         newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.NodeCondition":                                newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.PersistentVolume":                             newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.PersistentVolumeClaim":                        newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.Pod":                                          newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.PodCondition":                                 newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.ReplicationController":                        newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.ReplicationControllerCondition":               newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.ResourceQuota":                                newPropertySet("status"),
			"io.k8s.kubernetes.pkg.api.v1.Service":                                      newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.Deployment":                        newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentCondition":               newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.Scale":                             newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.apps.v1beta1.StatefulSet":                       newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authentication.v1.TokenReview":                  newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authentication.v1beta1.TokenReview":             newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1.LocalSubjectAccessReview":      newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1.SelfSubjectAccessReview":       newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1.SubjectAccessReview":           newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1beta1.LocalSubjectAccessReview": newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1beta1.SelfSubjectAccessReview":  newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.authorization.v1beta1.SubjectAccessReview":      newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.autoscaling.v1.HorizontalPodAutoscaler":         newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.autoscaling.v1.Scale":                           newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.autoscaling.v2alpha1.HorizontalPodAutoscaler":   newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.batch.v1.Job":                                   newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.batch.v1.JobCondition":                          newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.batch.v2alpha1.CronJob":                         newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.certificates.v1beta1.CertificateSigningRequest": newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.DaemonSet":                   newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.Deployment":                  newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.DeploymentCondition":         newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.Ingress":                     newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.ReplicaSet":                  newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.ReplicaSetCondition":         newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.Scale":                       newPropertySet("status"),
			"io.k8s.kubernetes.pkg.apis.policy.v1beta1.PodDisruptionBudget":             newPropertySet("status"),

			// TODO: Find a more principled way to omit "status" types.
			// Currently we emit these in the `local hidden` in the `root`,
			// so that we can type aliases. To get around the fact that some
			// of their function names collide with Jsonnet keywords, we
			// simply choose not to emit them. Eventually we will approach
			// this problem in a more principled manner.
			"io.k8s.kubernetes.pkg.api.v1.ComponentCondition":                     newPropertySet("error", "status"),
			"io.k8s.kubernetes.pkg.apis.authentication.v1.TokenReviewStatus":      newPropertySet("error"),
			"io.k8s.kubernetes.pkg.apis.authentication.v1beta1.TokenReviewStatus": newPropertySet("error"),

			// Has both status and a property with type
			// `io.k8s.apimachinery.pkg.apis.meta.v1.ListMeta`.
			"io.k8s.apimachinery.pkg.apis.meta.v1.Status": newPropertySet("status", "metadata"),

			// Misc.
			"io.k8s.kubernetes.pkg.apis.extensions.v1beta1.DaemonSetSpec": newPropertySet("templateGeneration"),
		},
	},
}
