package utils

import (
	"os"

	clientsetAPIExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

// GetConfigLocation returns a map which has information on the namespace where Kubeless controller is installed and the name of the ConfigMap which stores kubeless configurations
func GetConfigLocation(apiExtensionsClientset clientsetAPIExtensions.Interface) (map[string]string, error) {
	configLocation := make(map[string]string)
	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	kubelessConfig := os.Getenv("KUBELESS_CONFIG")

	annotationsCRD, err := GetAnnotationsFromCRD(apiExtensionsClientset, "functions.kubeless.io")
	if err != nil {
		return nil, err
	}
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
	return configLocation, nil
}
