/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	kafkaApi "github.com/kubeless/kafka-trigger/pkg/apis/kubeless/v1beta1"
	"github.com/sirupsen/logrus"

	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	clientsetAPIExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"

	// Auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/imdario/mergo"
	"github.com/kubeless/kafka-trigger/pkg/client/clientset/versioned"
)

const (
	defaultTimeout = "180"
)

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not create kubernetes client: %v", err)
	}

	return clientset
}

// BuildOutOfClusterConfig returns k8s config
func BuildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			for _, h := range []string{"HOME", "USERPROFILE"} {
				if home = os.Getenv(h); home != "" {
					break
				}
			}
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// GetTriggerClientInCluster returns function clientset to the request from inside of cluster
func GetTriggerClientInCluster() (versioned.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kafkaClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kafkaClient, nil
}

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster() kubernetes.Interface {
	config, err := BuildOutOfClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not get kubernetes client: %v", err)
	}

	return clientset
}

// GetAPIExtensionsClientOutOfCluster returns a k8s clientset to access APIExtensions from outside of cluster
func GetAPIExtensionsClientOutOfCluster() clientsetAPIExtensions.Interface {
	config, err := BuildOutOfClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}
	clientset, err := clientsetAPIExtensions.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not get kubernetes client: %v", err)
	}
	return clientset
}

// GetAPIExtensionsClientInCluster returns a k8s clientset to access APIExtensions from inside of cluster
func GetAPIExtensionsClientInCluster() clientsetAPIExtensions.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}
	clientset, err := clientsetAPIExtensions.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not get kubernetes client: %v", err)
	}
	return clientset
}

// GetKubelessClientOutCluster returns kubeless clientset to make kubeless API request from outside of cluster
func GetKubelessClientOutCluster() (versioned.Interface, error) {
	config, err := BuildOutOfClusterConfig()
	if err != nil {
		return nil, err
	}
	kubelessClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kubelessClient, nil
}

// CreateKafkaTriggerCustomResource will create a custom function object
func CreateKafkaTriggerCustomResource(kubelessClient versioned.Interface, kafkaTrigger *kafkaApi.KafkaTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().KafkaTriggers(kafkaTrigger.Namespace).Create(kafkaTrigger)
	if err != nil {
		return err
	}
	return nil
}

// UpdateKafkaTriggerCustomResource applies changes to the function custom object
func UpdateKafkaTriggerCustomResource(kubelessClient versioned.Interface, kafkaTrigger *kafkaApi.KafkaTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().KafkaTriggers(kafkaTrigger.Namespace).Update(kafkaTrigger)
	return err
}

// DeleteKafkaTriggerCustomResource will delete custom function object
func DeleteKafkaTriggerCustomResource(kubelessClient versioned.Interface, kafkaTriggerName, ns string) error {
	err := kubelessClient.KubelessV1beta1().KafkaTriggers(ns).Delete(kafkaTriggerName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetKafkaTriggerCustomResource will get CronJobTrigger custom resource object
func GetKafkaTriggerCustomResource(kubelessClient versioned.Interface, kafkaTriggerName, ns string) (*kafkaApi.KafkaTrigger, error) {
	kafkaCRD, err := kubelessClient.KubelessV1beta1().KafkaTriggers(ns).Get(kafkaTriggerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return kafkaCRD, nil
}

func doRESTReq(restIface rest.Interface, groupVersion, verb, resource, elem, namespace string, body interface{}, result interface{}) error {
	var req *rest.Request
	bodyJSON := []byte{}
	var err error
	if body != nil {
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	switch verb {
	case "get":
		req = restIface.Get().Name(elem)
		break
	case "create":
		req = restIface.Post().Body(bodyJSON)
		break
	case "update":
		req = restIface.Put().Name(elem).Body(bodyJSON)
		break
	default:
		return fmt.Errorf("Verb %s not supported", verb)
	}
	rawResponse, err := req.AbsPath("apis", groupVersion, "namespaces", namespace, resource).DoRaw()
	if err != nil {
		return err
	}
	if result != nil {
		err = json.Unmarshal(rawResponse, result)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateAutoscale creates HPA object for function
func CreateAutoscale(client kubernetes.Interface, hpa v2beta1.HorizontalPodAutoscaler) error {
	_, err := client.AutoscalingV2beta1().HorizontalPodAutoscalers(hpa.ObjectMeta.Namespace).Create(&hpa)
	if err != nil {
		return err
	}

	return err
}

// DeleteAutoscale deletes an autoscale rule
func DeleteAutoscale(client kubernetes.Interface, name, ns string) error {
	err := client.AutoscalingV2beta1().HorizontalPodAutoscalers(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

// DeleteServiceMonitor cleans the sm if it exists
func DeleteServiceMonitor(smclient monitoringv1alpha1.MonitoringV1alpha1Client, name, ns string) error {
	err := smclient.ServiceMonitors(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	return nil
}

// InitializeEmptyMapsInDeployment initializes all nil maps in a Deployment object
// This is done to counteract with side-effects of github.com/imdario/mergo which panics when provided with a nil map in a struct
func initializeEmptyMapsInDeployment(deployment *v1beta1.Deployment) {
	if deployment.ObjectMeta.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = make(map[string]string)
	}
	if deployment.Spec.Selector != nil && deployment.Spec.Selector.MatchLabels == nil {
		deployment.ObjectMeta.Labels = make(map[string]string)
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	}
	if deployment.Spec.Template.Spec.NodeSelector == nil {
		deployment.Spec.Template.Spec.NodeSelector = make(map[string]string)
	}
}

// MergeDeployments merges two deployment objects
func MergeDeployments(destinationDeployment *v1beta1.Deployment, sourceDeployment *v1beta1.Deployment) error {
	// Initializing nil maps in deployment objects else github.com/imdario/mergo panics
	initializeEmptyMapsInDeployment(destinationDeployment)
	initializeEmptyMapsInDeployment(sourceDeployment)
	return mergo.Merge(destinationDeployment, sourceDeployment)
}

// GetAnnotationsFromCRD gets annotations from a CustomResourceDefinition
func GetAnnotationsFromCRD(clientset clientsetAPIExtensions.Interface, name string) (map[string]string, error) {
	crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return crd.GetAnnotations(), nil
}

// GetRandString returns a random string of lenght N
func GetRandString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GetSecretsAsLocalObjectReference returns a list of LocalObjectReference based on secret names
func GetSecretsAsLocalObjectReference(secrets ...string) []v1.LocalObjectReference {
	res := []v1.LocalObjectReference{}
	for _, secret := range secrets {
		if secret != "" {
			res = append(res, v1.LocalObjectReference{Name: secret})
		}
	}
	return res
}
