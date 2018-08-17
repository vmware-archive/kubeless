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
	"net/url"
	"os"
	"path/filepath"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
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
	"k8s.io/apimachinery/pkg/types"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"

	// Auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/imdario/mergo"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
)

const (
	defaultTimeout = "180"
)

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := GetInClusterConfig()
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
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv == "" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			for _, h := range []string{"HOME", "USERPROFILE"} {
				if home = os.Getenv(h); home != "" {
					break
				}
			}
		}
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		loadingRules.ExplicitPath = kubeconfigPath
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
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
	config, err := GetInClusterConfig()

	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}
	clientset, err := clientsetAPIExtensions.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not get kubernetes client: %v", err)
	}
	return clientset
}

// GetFunctionClientInCluster returns function clientset to the request from inside of cluster
func GetFunctionClientInCluster() (versioned.Interface, error) {
	config, err := GetInClusterConfig()
	if err != nil {
		return nil, err
	}

	kubelessClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kubelessClient, nil
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

//GetDefaultNamespace returns the namespace set in current cluster context
func GetDefaultNamespace() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if ns, _, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).Namespace(); err == nil {
		return ns
	}
	return v1.NamespaceDefault
}

// GetFunction returns specification of a function
func GetFunction(funcName, ns string) (kubelessApi.Function, error) {
	kubelessClient, err := GetKubelessClientOutCluster()
	if err != nil {
		return kubelessApi.Function{}, err
	}

	f, err := kubelessClient.KubelessV1beta1().Functions(ns).Get(funcName, metav1.GetOptions{})

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			logrus.Fatalf("Function %s is not found", funcName)
		}
		return kubelessApi.Function{}, err
	}

	return *f, nil
}

// CreateFunctionCustomResource will create a custom function object
func CreateFunctionCustomResource(kubelessClient versioned.Interface, f *kubelessApi.Function) error {
	_, err := kubelessClient.KubelessV1beta1().Functions(f.Namespace).Create(f)
	if err != nil {
		return err
	}
	return nil
}

// UpdateFunctionCustomResource applies changes to the function custom object
func UpdateFunctionCustomResource(kubelessClient versioned.Interface, f *kubelessApi.Function) error {
	_, err := kubelessClient.KubelessV1beta1().Functions(f.Namespace).Update(f)
	return err
}

// PatchFunctionCustomResource applies changes to the function custom object
func PatchFunctionCustomResource(kubelessClient versioned.Interface, f *kubelessApi.Function) error {
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	_, err = kubelessClient.KubelessV1beta1().Functions(f.Namespace).Patch(f.Name, types.MergePatchType, data)
	return err
}

// DeleteFunctionCustomResource will delete custom function object
func DeleteFunctionCustomResource(kubelessClient versioned.Interface, funcName, ns string) error {
	err := kubelessClient.KubelessV1beta1().Functions(ns).Delete(funcName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetFunctionCustomResource will delete custom function object
func GetFunctionCustomResource(kubelessClient versioned.Interface, funcName, ns string) (*kubelessApi.Function, error) {
	functionObj, err := kubelessClient.KubelessV1beta1().Functions(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return functionObj, nil
}

// GetPodsByLabel returns list of pods which match the label
// We use this to returns pods to which the function is deployed or pods running controllers
func GetPodsByLabel(c kubernetes.Interface, ns, k, v string) (*v1.PodList, error) {
	pods, err := c.Core().Pods(ns).List(metav1.ListOptions{
		LabelSelector: k + "=" + v,
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

// GetReadyPod returns the first pod has passed the liveness probe check
func GetReadyPod(pods *v1.PodList) (v1.Pod, error) {
	for _, pod := range pods.Items {
		isPodRunning := true
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				isPodRunning = false
				break
			}
		}
		if isPodRunning {
			return pod, nil
		}
	}
	return v1.Pod{}, fmt.Errorf("there is no pod ready")
}

// GetLocalHostname returns hostname
func GetLocalHostname(config *rest.Config, funcName string) (string, error) {
	url, err := url.Parse(config.Host)
	if err != nil {
		return "", err
	}
	host := url.Hostname()

	return fmt.Sprintf("%s.%s.nip.io", funcName, host), nil
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

// FunctionObjAddFinalizer add specified finalizer string to function object
func FunctionObjAddFinalizer(kubelessClient versioned.Interface, funcObj *kubelessApi.Function, finalizerString string) error {
	funcObjClone := funcObj.DeepCopy()
	funcObjClone.ObjectMeta.Finalizers = append(funcObjClone.ObjectMeta.Finalizers, finalizerString)
	return UpdateFunctionCustomResource(kubelessClient, funcObjClone)
}

// FunctionObjHasFinalizer checks if function object already has the Function controller finalizer
func FunctionObjHasFinalizer(funcObj *kubelessApi.Function, finalizerString string) bool {
	currentFinalizers := funcObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == finalizerString {
			return true
		}
	}
	return false
}

// FunctionObjRemoveFinalizer removes the finalizer from the function object
func FunctionObjRemoveFinalizer(kubelessClient versioned.Interface, funcObj *kubelessApi.Function, finalizerString string) error {
	funcObjClone := funcObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range funcObj.ObjectMeta.Finalizers {
		if item == finalizerString {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	funcObjClone.ObjectMeta.Finalizers = newSlice
	err := UpdateFunctionCustomResource(kubelessClient, funcObjClone)
	return err
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
