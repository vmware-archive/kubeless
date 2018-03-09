package utils

import (
	"os"
)

// GetConfigLocation returns a map which has information on the namespace where Kubeless controller is installed and the name of the ConfigMap which stores Kubeless configurations
func GetConfigLocation() map[string]string {
	var configLocation map[string]string
	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	kubelessConfig := os.Getenv("KUBELESS_CONFIG")
	apiExtensionsClientset := GetAPIExtensionsClientOutOfCluster()
	annotationsCRD, _ := GetAnnotationsFromCRD(apiExtensionsClientset, "functions.kubeless.io")
	if len(controllerNamespace) == 0 {
		if ns, ok := annotationsCRD["kubeless.io/namespace"]; ok {
			controllerNamespace = ns
		} else {
			controllerNamespace = "kubeless"
		}
	}
	configLocation["namespace"] = controllerNamespace
	if len(kubelessConfig) == 0 {
		if config, ok := annotationsCRD["kubeless.io/config"]; ok {
			kubelessConfig = config
		} else {
			kubelessConfig = "kubeless-config"
		}
	}
	configLocation["name"] = kubelessConfig
	return configLocation
}
