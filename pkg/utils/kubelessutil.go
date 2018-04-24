package utils

import (
	"fmt"
	"os"

	"k8s.io/api/core/v1"
	clientsetAPIExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getConfigLocation(apiExtensionsClientset clientsetAPIExtensions.Interface) (ConfigLocation, error) {
	configLocation := ConfigLocation{}
	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	kubelessConfig := os.Getenv("KUBELESS_CONFIG")

	annotationsCRD, err := GetAnnotationsFromCRD(apiExtensionsClientset, "functions.kubeless.io")
	if err != nil {
		return configLocation, err
	}
	if len(controllerNamespace) == 0 {
		if ns, ok := annotationsCRD["kubeless.io/namespace"]; ok {
			controllerNamespace = ns
		} else {
			controllerNamespace = "kubeless"
		}
	}
	configLocation.Namespace = controllerNamespace
	if len(kubelessConfig) == 0 {
		if config, ok := annotationsCRD["kubeless.io/config"]; ok {
			kubelessConfig = config
		} else {
			kubelessConfig = "kubeless-config"
		}
	}
	configLocation.Name = kubelessConfig
	return configLocation, nil
}

// GetKubelessConfig Returns Kubeless ConfigMap
func GetKubelessConfig(cli kubernetes.Interface) (*v1.ConfigMap, error) {
	apiExtensionsClientset := GetAPIExtensionsClientInCluster()
	configLocation, err := getConfigLocation(apiExtensionsClientset)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching config location: %v", err)
	}
	controllerNamespace := configLocation.Namespace
	kubelessConfig := configLocation.Name
	config, err := cli.CoreV1().ConfigMaps(controllerNamespace).Get(kubelessConfig, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to read the configmap: %s", err)
	}
	return config, nil
}
