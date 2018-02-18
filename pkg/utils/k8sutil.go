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
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kubeless/kubeless/pkg/langruntime"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/sirupsen/logrus"

	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"

	// Adding explicitely the GCP auth plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/imdario/mergo"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
)

const (
	pubsubFunc     = "PubSub"
	busybox        = "busybox@sha256:be3c11fdba7cfe299214e46edc642e09514dbb9bbefcd0d3836c05a1e0cd0642"
	unzip          = "kubeless/unzip@sha256:f162c062973cca05459834de6ed14c039d45df8cdb76097f50b028a1621b3697"
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

// GetFunctionClientInCluster returns function clientset to the request from inside of cluster
func GetFunctionClientInCluster() (versioned.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubelessClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kubelessClient, nil
}

// GetFunctionClientOutCluster returns function clientset to the request from outside of cluster
func GetFunctionClientOutCluster() (versioned.Interface, error) {
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
	kubelessClient, err := GetFunctionClientOutCluster()
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

// CreateK8sCustomResource will create a custom function object
func CreateK8sCustomResource(kubelessClient versioned.Interface, f *kubelessApi.Function) error {
	_, err := kubelessClient.KubelessV1beta1().Functions(f.Namespace).Create(f)
	if err != nil {
		return err
	}
	return nil
}

// UpdateK8sCustomResource applies changes to the function custom object
func UpdateK8sCustomResource(kubelessClient versioned.Interface, f *kubelessApi.Function) error {
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}

	_, err = kubelessClient.KubelessV1beta1().Functions(f.Namespace).Patch(f.Name, types.MergePatchType, data)
	return err
}

// DeleteK8sCustomResource will delete custom function object
func DeleteK8sCustomResource(kubelessClient versioned.Interface, funcName, ns string) error {
	err := kubelessClient.KubelessV1beta1().Functions(ns).Delete(funcName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
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
		if pod.Status.ContainerStatuses[0].Ready {
			return pod, nil
		}
	}
	return v1.Pod{}, fmt.Errorf("there is no pod ready")
}

func appendToCommand(orig string, command ...string) string {
	if len(orig) > 0 {
		return fmt.Sprintf("%s && %s", orig, strings.Join(command, " && "))
	}
	return strings.Join(command, " && ")
}

func getProvisionContainer(function, checksum, fileName, handler, contentType, runtime string, runtimeVolume, depsVolume v1.VolumeMount, lr *langruntime.Langruntimes) (v1.Container, error) {
	prepareCommand := ""
	originFile := path.Join(depsVolume.MountPath, fileName)

	// Prepare Function file and dependencies
	if strings.Contains(contentType, "base64") {
		// File is encoded in base64
		prepareCommand = appendToCommand(prepareCommand, fmt.Sprintf("base64 -d < %s > %s.decoded", originFile, originFile))
		originFile = originFile + ".decoded"
	} else if strings.Contains(contentType, "text") || contentType == "" {
		// Assumming that function is plain text
		// So we don't need to preprocess it
	} else {
		return v1.Container{}, fmt.Errorf("Unable to prepare function of type %s: Unknown format", contentType)
	}

	// Validate checksum
	if checksum == "" {
		// DEPRECATED: Checksum may be empty
	} else {
		checksumInfo := strings.Split(checksum, ":")
		switch checksumInfo[0] {
		case "sha256":
			shaFile := originFile + ".sha256"
			prepareCommand = appendToCommand(prepareCommand,
				fmt.Sprintf("echo '%s  %s' > %s", checksumInfo[1], originFile, shaFile),
				fmt.Sprintf("sha256sum -c %s", shaFile),
			)
			break
		default:
			return v1.Container{}, fmt.Errorf("Unable to verify checksum %s: Unknown format", checksum)
		}
	}

	// Extract content in case it is a Zip file
	if strings.Contains(contentType, "zip") {
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("unzip -o %s -d %s", originFile, runtimeVolume.MountPath),
		)
	} else {
		// Copy the target as a single file
		destFileName, err := getFileName(handler, contentType, runtime, lr)
		if err != nil {
			return v1.Container{}, err
		}
		dest := path.Join(runtimeVolume.MountPath, destFileName)
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("cp %s %s", originFile, dest),
		)
	}

	// Copy deps file to the installation path
	runtimeInf, err := lr.GetRuntimeInfo(runtime)
	if err == nil && runtimeInf.DepName != "" {
		depsFile := path.Join(depsVolume.MountPath, runtimeInf.DepName)
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("cp %s %s", depsFile, runtimeVolume.MountPath),
		)
	}

	return v1.Container{
		Name:            "prepare",
		Image:           unzip,
		Command:         []string{"sh", "-c"},
		Args:            []string{prepareCommand},
		VolumeMounts:    []v1.VolumeMount{runtimeVolume, depsVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
	}, nil
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

// DeleteIngress deletes an ingress rule
func DeleteIngress(client kubernetes.Interface, name, ns string) error {
	err := client.ExtensionsV1beta1().Ingresses(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

func splitHandler(handler string) (string, string, error) {
	str := strings.Split(handler, ".")
	if len(str) != 2 {
		return "", "", fmt.Errorf("failed: incorrect handler format. It should be module_name.handler_name")
	}

	return str[0], str[1], nil
}

// getFileName returns a file name based on a handler identifier
func getFileName(handler, funcContentType, runtime string, lr *langruntime.Langruntimes) (string, error) {
	modName, _, err := splitHandler(handler)
	if err != nil {
		return "", err
	}
	filename := modName
	if funcContentType == "text" || funcContentType == "" {
		// We can only guess the extension if the function is specified as plain text
		runtimeInf, err := lr.GetRuntimeInfo(runtime)
		if err == nil {
			filename = modName + runtimeInf.FileNameSuffix
		}
	}
	return filename, nil
}

// EnsureFuncConfigMap creates/updates a config map with a function specification
func EnsureFuncConfigMap(client kubernetes.Interface, funcObj *kubelessApi.Function, or []metav1.OwnerReference, lr *langruntime.Langruntimes) error {
	configMapData := map[string]string{}
	var err error
	if funcObj.Spec.Handler != "" {
		fileName, err := getFileName(funcObj.Spec.Handler, funcObj.Spec.FunctionContentType, funcObj.Spec.Runtime, lr)
		if err != nil {
			return err
		}
		configMapData = map[string]string{
			"handler": funcObj.Spec.Handler,
			fileName:  funcObj.Spec.Function,
		}
		runtimeInfo, err := lr.GetRuntimeInfo(funcObj.Spec.Runtime)
		if err == nil && runtimeInfo.DepName != "" {
			configMapData[runtimeInfo.DepName] = funcObj.Spec.Deps
		}
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.ObjectMeta.Name,
			Labels:          funcObj.ObjectMeta.Labels,
			OwnerReferences: or,
		},
		Data: configMapData,
	}

	_, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Create(configMap)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the ConfigMap already exists we should update
		// just certain fields (to avoid race conditions)
		var newConfigMap *v1.ConfigMap
		newConfigMap, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newConfigMap.ObjectMeta.Labels = funcObj.ObjectMeta.Labels
		newConfigMap.ObjectMeta.OwnerReferences = or
		newConfigMap.Data = configMap.Data
		_, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Update(newConfigMap)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The configmap may already exist and there is nothing to update
			return nil
		}
	}

	return err
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

// CreateServiceMonitor creates a Service Monitor for the given function
func CreateServiceMonitor(smclient monitoringv1alpha1.MonitoringV1alpha1Client, funcObj *kubelessApi.Function, ns string, or []metav1.OwnerReference) error {
	_, err := smclient.ServiceMonitors(ns).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			s := &monitoringv1alpha1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      funcObj.ObjectMeta.Name,
					Namespace: ns,
					Labels: map[string]string{
						"service-monitor": "function",
					},
					OwnerReferences: or,
				},
				Spec: monitoringv1alpha1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"function": funcObj.ObjectMeta.Name,
						},
					},
					Endpoints: []monitoringv1alpha1.Endpoint{
						{
							Port: "http-function-port",
						},
					},
				},
			}
			_, err = smclient.ServiceMonitors(ns).Create(s)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("service monitor has already existed")
}

// GetOwnerReference returns ownerRef for appending to objects's metadata
// created by kubeless-controller one a function is deployed.
func GetOwnerReference(funcObj *kubelessApi.Function) ([]metav1.OwnerReference, error) {
	if funcObj.ObjectMeta.Name == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("function name can't be empty")
	}
	if funcObj.ObjectMeta.UID == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("uid of function %s can't be empty", funcObj.ObjectMeta.Name)
	}
	return []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
			Name:       funcObj.ObjectMeta.Name,
			UID:        funcObj.ObjectMeta.UID,
		},
	}, nil
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
