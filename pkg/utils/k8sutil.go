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
	"strconv"
	"strings"

	"github.com/kubeless/kubeless/pkg/langruntime"

	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"

	// Adding explicitely the GCP auth plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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

// GetRestClient returns a k8s restclient to the request from inside of cluster
func GetRestClient() (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return restClient, nil
}

// GetCRDClient returns crd client to the request from inside of cluster
func GetCRDClient() (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	crdConfig := configureClient("k8s.io", "v1", "/apis", config)

	crdclient, err := rest.RESTClientFor(crdConfig)
	if err != nil {
		return nil, err
	}

	return crdclient, nil
}

// GetRestClientOutOfCluster returns a REST client based on a group, API version and path
func GetRestClientOutOfCluster(group, apiVersion, apiPath string) (*rest.RESTClient, error) {
	config, err := BuildOutOfClusterConfig()
	if err != nil {
		return nil, err
	}

	crdConfig := configureClient(group, apiVersion, apiPath, config)

	client, err := rest.RESTClientFor(crdConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetCRDClientOutOfCluster returns crd client to the request from outside of cluster
func GetCRDClientOutOfCluster() (*rest.RESTClient, error) {
	return GetRestClientOutOfCluster("k8s.io", "v1", "/apis")
}

//GetDefaultNamespace returns the namespace set in current cluster context
func GetDefaultNamespace() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if ns, _, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).Namespace(); err == nil {
		return ns
	}
	return api.NamespaceDefault
}

// GetFunction returns specification of a function
func GetFunction(funcName, ns string) (spec.Function, error) {
	var f spec.Function

	crdClient, err := GetCRDClientOutOfCluster()
	if err != nil {
		return spec.Function{}, err
	}

	err = crdClient.Get().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			logrus.Fatalf("Function %s is not found", funcName)
		}
		return spec.Function{}, err
	}

	return f, nil
}

// CreateK8sCustomResource will create a custom function object
func CreateK8sCustomResource(crdClient rest.Interface, f *spec.Function) error {
	err := crdClient.Post().
		Resource("functions").
		Namespace(f.Metadata.Namespace).
		Body(f).
		Do().Error()
	if err != nil {
		return err
	}

	return nil
}

// UpdateK8sCustomResource applies changes to the function custom object
func UpdateK8sCustomResource(crdClient rest.Interface, f *spec.Function) error {
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return crdClient.Patch(types.MergePatchType).
		Namespace(f.Metadata.Namespace).
		Resource("functions").
		Name(f.Metadata.Name).
		Body(data).
		Do().Error()
}

// DeleteK8sCustomResource will delete custom function object
func DeleteK8sCustomResource(crdClient *rest.RESTClient, funcName, ns string) error {
	err := crdClient.Delete().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Error()

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

func getProvisionContainer(function, checksum, fileName, handler, contentType, runtime string, runtimeVolume, depsVolume v1.VolumeMount) (v1.Container, error) {
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
		destFileName, err := getFileName(handler, contentType, runtime)
		if err != nil {
			return v1.Container{}, err
		}
		dest := path.Join(runtimeVolume.MountPath, destFileName)
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("cp %s %s", originFile, dest),
		)
	}

	// Copy deps file to the installation path
	runtimeInf, err := langruntime.GetRuntimeInfo(runtime)
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

// configureClient configures crd client
func configureClient(group, version, apiPath string, config *rest.Config) *rest.Config {
	var result rest.Config
	result = *config
	groupversion := schema.GroupVersion{
		Group:   group,
		Version: version,
	}

	result.GroupVersion = &groupversion
	result.APIPath = apiPath
	result.ContentType = runtime.ContentTypeJSON
	result.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&spec.Function{},
				&spec.FunctionList{},
			)
			return nil
		})
	metav1.AddToGroupVersion(api.Scheme, groupversion)
	schemeBuilder.AddToScheme(api.Scheme)
	return &result
}

// addInitContainerAnnotation is a hot fix to add annotation to deployment for init container to run
func addInitContainerAnnotation(dpm *v1beta1.Deployment) error {
	if len(dpm.Spec.Template.Spec.InitContainers) > 0 {
		value, err := json.Marshal(dpm.Spec.Template.Spec.InitContainers)
		if err != nil {
			return err
		}
		if dpm.Spec.Template.Annotations == nil {
			dpm.Spec.Template.Annotations = make(map[string]string)
		}
		dpm.Spec.Template.Annotations[v1.PodInitContainersAnnotationKey] = string(value)
		dpm.Spec.Template.Annotations[v1.PodInitContainersBetaAnnotationKey] = string(value)
	}
	return nil
}

// CreateIngress creates ingress rule for a specific function
func CreateIngress(client kubernetes.Interface, funcObj *spec.Function, ingressName, hostname, ns string, enableTLSAcme bool) error {
	or, err := GetOwnerReference(funcObj)
	if err != nil {
		return err
	}

	if len(funcObj.Spec.ServiceSpec.Ports) == 0 {
		return fmt.Errorf("can't create route due to service port isn't defined")
	}

	port := funcObj.Spec.ServiceSpec.Ports[0].TargetPort
	if port.IntVal <= 0 || port.IntVal > 65535 {
		return fmt.Errorf("Invalid port number %d specified", port.IntVal)
	}

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ingressName,
			Namespace:       ns,
			OwnerReferences: or,
			Labels:          funcObj.Metadata.Labels,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: funcObj.Metadata.Name,
										ServicePort: funcObj.Spec.ServiceSpec.Ports[0].TargetPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if enableTLSAcme {
		// add annotations and TLS configuration for kube-lego
		ingressAnnotations := map[string]string{
			"kubernetes.io/tls-acme":             "true",
			"ingress.kubernetes.io/ssl-redirect": "true",
		}
		ingress.ObjectMeta.Annotations = ingressAnnotations

		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{hostname},
				SecretName: ingressName + "-tls",
			},
		}
	}

	_, err = client.ExtensionsV1beta1().Ingresses(ns).Create(ingress)
	if err != nil {
		return err
	}
	return nil
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
func getFileName(handler, funcContentType, runtime string) (string, error) {
	modName, _, err := splitHandler(handler)
	if err != nil {
		return "", err
	}
	filename := modName
	if funcContentType == "text" || funcContentType == "" {
		// We can only guess the extension if the function is specified as plain text
		runtimeInf, err := langruntime.GetRuntimeInfo(runtime)
		if err == nil {
			filename = modName + runtimeInf.FileNameSuffix
		}
	}
	return filename, nil
}

// EnsureFuncConfigMap creates/updates a config map with a function specification
func EnsureFuncConfigMap(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	configMapData := map[string]string{}
	var err error
	if funcObj.Spec.Handler != "" {
		fileName, err := getFileName(funcObj.Spec.Handler, funcObj.Spec.FunctionContentType, funcObj.Spec.Runtime)
		if err != nil {
			return err
		}
		configMapData = map[string]string{
			"handler": funcObj.Spec.Handler,
			fileName:  funcObj.Spec.Function,
		}
		runtimeInfo, err := langruntime.GetRuntimeInfo(funcObj.Spec.Runtime)
		if err == nil && runtimeInfo.DepName != "" {
			configMapData[runtimeInfo.DepName] = funcObj.Spec.Deps
		}
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.Metadata.Name,
			Labels:          funcObj.Metadata.Labels,
			OwnerReferences: or,
		},
		Data: configMapData,
	}

	_, err = client.Core().ConfigMaps(funcObj.Metadata.Namespace).Create(configMap)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the ConfigMap already exists we should update
		// just certain fields (to avoid race conditions)
		var newConfigMap *v1.ConfigMap
		newConfigMap, err = client.Core().ConfigMaps(funcObj.Metadata.Namespace).Get(funcObj.Metadata.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newConfigMap.ObjectMeta.Labels = funcObj.Metadata.Labels
		newConfigMap.ObjectMeta.OwnerReferences = or
		newConfigMap.Data = configMap.Data
		_, err = client.Core().ConfigMaps(funcObj.Metadata.Namespace).Update(newConfigMap)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The configmap may already exist and there is nothing to update
			return nil
		}
	}

	return err
}

// this function resolves backward incompatibility in case user uses old client which doesn't include serviceSpec into funcSpec.
// if serviceSpec is empty, we will use the default serviceSpec whose port is 8080
func serviceSpec(funcObj *spec.Function) v1.ServiceSpec {
	if len(funcObj.Spec.ServiceSpec.Ports) != 0 && len(funcObj.Spec.ServiceSpec.Selector) != 0 {
		return funcObj.Spec.ServiceSpec
	}

	return v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:       "function-port",
				Protocol:   v1.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		},
		Selector: funcObj.Metadata.Labels,
		Type:     v1.ServiceTypeClusterIP,
	}
}

// EnsureFuncService creates/updates a function service
func EnsureFuncService(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.Metadata.Name,
			Labels:          funcObj.Metadata.Labels,
			OwnerReferences: or,
		},
		Spec: serviceSpec(funcObj),
	}

	_, err := client.Core().Services(funcObj.Metadata.Namespace).Create(svc)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the SVC already exists we should update
		// just certain fields (to avoid race conditions)
		var newSvc *v1.Service
		newSvc, err = client.Core().Services(funcObj.Metadata.Namespace).Get(funcObj.Metadata.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newSvc.ObjectMeta.Labels = funcObj.Metadata.Labels
		newSvc.ObjectMeta.OwnerReferences = or
		newSvc.Spec.Ports = svc.Spec.Ports
		newSvc.Spec.Selector = svc.Spec.Selector
		_, err = client.Core().Services(funcObj.Metadata.Namespace).Update(newSvc)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The service may already exist and there is nothing to update
			return nil
		}
	}
	return err
}

func svcPort(funcObj *spec.Function) int32 {
	if len(funcObj.Spec.ServiceSpec.Ports) != 0 {
		return funcObj.Spec.ServiceSpec.Ports[0].Port
	}
	return int32(8080)
}

// EnsureFuncDeployment creates/updates a function deployment
func EnsureFuncDeployment(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	runtimeVolumeName := funcObj.Metadata.Name
	depsVolumeName := funcObj.Metadata.Name + "-deps"
	podAnnotations := map[string]string{
		// Attempt to attract the attention of prometheus.
		// For runtimes that don't support /metrics,
		// prometheus will get a 404 and mostly silently
		// ignore the pod (still displayed in the list of
		// "targets")
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   strconv.Itoa(int(svcPort(funcObj))),
	}

	//add deployment
	maxUnavailable := intstr.FromInt(0)
	dpm := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.Metadata.Name,
			Labels:          funcObj.Metadata.Labels,
			OwnerReferences: or,
		},
		Spec: v1beta1.DeploymentSpec{
			Strategy: v1beta1.DeploymentStrategy{
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
				},
			},
		},
	}

	//copy all func's Spec.Template to the deployment
	tmplCopy, err := api.Scheme.DeepCopy(funcObj.Spec.Template)
	if err != nil {
		return err
	}
	dpm.Spec.Template = tmplCopy.(v1.PodTemplateSpec)

	//append data to dpm spec
	if len(dpm.Spec.Template.ObjectMeta.Labels) == 0 {
		dpm.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range funcObj.Metadata.Labels {
		dpm.Spec.Template.ObjectMeta.Labels[k] = v
	}
	if len(dpm.Spec.Template.ObjectMeta.Annotations) == 0 {
		dpm.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	for k, v := range podAnnotations {
		//only append k-v from podAnnotations if it doesn't exist in deployment podTemplateSpec annotation
		if _, ok := dpm.Spec.Template.ObjectMeta.Annotations[k]; !ok {
			dpm.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name: runtimeVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		v1.Volume{
			Name: depsVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: funcObj.Metadata.Name,
					},
				},
			},
		})

	if len(dpm.Spec.Template.Spec.Containers) == 0 {
		dpm.Spec.Template.Spec.Containers = append(dpm.Spec.Template.Spec.Containers, v1.Container{})
	}

	if funcObj.Spec.Handler != "" {
		modName, handlerName, err := splitHandler(funcObj.Spec.Handler)
		if err != nil {
			return err
		}
		//only resolve the image name if it has not been already set
		if dpm.Spec.Template.Spec.Containers[0].Image == "" {
			imageName, err := langruntime.GetFunctionImage(funcObj.Spec.Runtime, funcObj.Spec.Type)
			if err != nil {
				return err
			}
			dpm.Spec.Template.Spec.Containers[0].Image = imageName
		}
		timeout := funcObj.Spec.Timeout
		if timeout == "" {
			// Set default timeout to 180 seconds
			timeout = defaultTimeout
		}
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  "FUNC_HANDLER",
				Value: handlerName,
			},
			v1.EnvVar{
				Name:  "MOD_NAME",
				Value: modName,
			},
			v1.EnvVar{
				Name:  "FUNC_TIMEOUT",
				Value: timeout,
			},
		)
	}

	dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  "FUNC_PORT",
			Value: strconv.Itoa(int(svcPort(funcObj))),
		},
	)

	dpm.Spec.Template.Spec.Containers[0].Name = funcObj.Metadata.Name
	dpm.Spec.Template.Spec.Containers[0].Ports = append(dpm.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{
		ContainerPort: svcPort(funcObj),
	})
	dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  "TOPIC_NAME",
			Value: funcObj.Spec.Topic,
		})

	runtimeVolumeMount := v1.VolumeMount{
		Name:      runtimeVolumeName,
		MountPath: "/kubeless",
	}

	dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, runtimeVolumeMount)

	// prepare init-containers if some function is specified
	if funcObj.Spec.Function != "" {
		fileName, err := getFileName(funcObj.Spec.Handler, funcObj.Spec.FunctionContentType, funcObj.Spec.Runtime)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		srcVolumeMount := v1.VolumeMount{
			Name:      depsVolumeName,
			MountPath: "/src",
		}
		provisionContainer, err := getProvisionContainer(
			funcObj.Spec.Function,
			funcObj.Spec.Checksum,
			fileName,
			funcObj.Spec.Handler,
			funcObj.Spec.FunctionContentType,
			funcObj.Spec.Runtime,
			runtimeVolumeMount,
			srcVolumeMount,
		)
		if err != nil {
			return err
		}
		dpm.Spec.Template.Spec.InitContainers = []v1.Container{provisionContainer}
	}
	// ensure that the runtime is supported for installing dependencies
	_, err = langruntime.GetRuntimeInfo(funcObj.Spec.Runtime)
	if funcObj.Spec.Deps != "" && err != nil {
		return fmt.Errorf("Unable to install dependencies for the runtime %s", funcObj.Spec.Runtime)
	} else if funcObj.Spec.Deps != "" {
		buildContainer, err := langruntime.GetBuildContainer(funcObj.Spec.Runtime, dpm.Spec.Template.Spec.Containers[0].Env, runtimeVolumeMount)
		if err != nil {
			return err
		}
		dpm.Spec.Template.Spec.InitContainers = append(
			dpm.Spec.Template.Spec.InitContainers,
			buildContainer,
		)
		// update deployment for loading dependencies
		langruntime.UpdateDeployment(dpm, runtimeVolumeMount.MountPath, funcObj.Spec.Runtime)
	}
	//TODO: remove this when init containers becomes a stable feature
	addInitContainerAnnotation(dpm)

	// add liveness Probe to deployment
	if funcObj.Spec.Type != pubsubFunc {
		livenessProbe := &v1.Probe{
			InitialDelaySeconds: int32(3),
			PeriodSeconds:       int32(30),
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(int(svcPort(funcObj))),
				},
			},
		}
		dpm.Spec.Template.Spec.Containers[0].LivenessProbe = livenessProbe
	}

	_, err = client.Extensions().Deployments(funcObj.Metadata.Namespace).Create(dpm)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the Deployment already exists we should update
		// just certain fields (to avoid race conditions)
		var newDpm *v1beta1.Deployment
		newDpm, err = client.Extensions().Deployments(funcObj.Metadata.Namespace).Get(funcObj.Metadata.Name, metav1.GetOptions{})
		newDpm.ObjectMeta.Labels = funcObj.Metadata.Labels
		newDpm.ObjectMeta.OwnerReferences = or
		newDpm.Spec = dpm.Spec
		_, err = client.Extensions().Deployments(funcObj.Metadata.Namespace).Update(newDpm)
		if err != nil {
			return err
		}

		// kick existing function pods then it will be recreated
		// with the new data mount from updated configmap.
		// TODO: This is a workaround.  Do something better.
		var pods *v1.PodList
		pods, err = GetPodsByLabel(client, funcObj.Metadata.Namespace, "function", funcObj.Metadata.Name)
		if err != nil {
			return err
		}
		for _, pod := range pods.Items {
			err = client.Core().Pods(funcObj.Metadata.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				// non-fatal
				logrus.Warnf("Unable to delete pod %s/%s, may be running stale version of function: %v", funcObj.Metadata.Namespace, pod.Name, err)
			}
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

// EnsureFuncCronJob creates/updates a function cron job
func EnsureFuncCronJob(client rest.Interface, funcObj *spec.Function, or []metav1.OwnerReference, groupVersion string) error {
	var maxSucccessfulHist, maxFailedHist int32
	maxSucccessfulHist = 3
	maxFailedHist = 1
	var timeout int
	if funcObj.Spec.Timeout != "" {
		var err error
		timeout, err = strconv.Atoi(funcObj.Spec.Timeout)
		if err != nil {
			return fmt.Errorf("Unable convert %s to a valid timeout", funcObj.Spec.Timeout)
		}
	} else {
		timeout, _ = strconv.Atoi(defaultTimeout)
	}
	activeDeadlineSeconds := int64(timeout)
	jobName := fmt.Sprintf("trigger-%s", funcObj.Metadata.Name)
	job := &batchv2alpha1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       funcObj.Metadata.Namespace,
			Labels:          funcObj.Metadata.Labels,
			OwnerReferences: or,
		},
		Spec: batchv2alpha1.CronJobSpec{
			Schedule:                   funcObj.Spec.Schedule,
			SuccessfulJobsHistoryLimit: &maxSucccessfulHist,
			FailedJobsHistoryLimit:     &maxFailedHist,
			JobTemplate: batchv2alpha1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds: &activeDeadlineSeconds,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: busybox,
									Name:  "trigger",
									Args:  []string{"wget", "-qO-", fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", funcObj.Metadata.Name, funcObj.Metadata.Namespace)},
								},
							},
							RestartPolicy: v1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	// We need to use directly the REST API since the endpoint
	// for CronJobs changes from Kubernetes 1.8
	err := doRESTReq(client, groupVersion, "create", "cronjobs", jobName, funcObj.Metadata.Namespace, job, nil)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		newCronJob := batchv2alpha1.CronJob{}
		err = doRESTReq(client, groupVersion, "get", "cronjobs", jobName, funcObj.Metadata.Namespace, nil, &newCronJob)
		if err != nil {
			return err
		}
		newCronJob.ObjectMeta.Labels = funcObj.Metadata.Labels
		newCronJob.ObjectMeta.OwnerReferences = or
		newCronJob.Spec = job.Spec
		err = doRESTReq(client, groupVersion, "update", "cronjobs", jobName, funcObj.Metadata.Namespace, &newCronJob, nil)
	}
	return err
}

// CreateAutoscale creates HPA object for function
func CreateAutoscale(client kubernetes.Interface, hpa v2alpha1.HorizontalPodAutoscaler) error {
	_, err := client.AutoscalingV2alpha1().HorizontalPodAutoscalers(hpa.ObjectMeta.Namespace).Create(&hpa)
	if err != nil {
		return err
	}

	return err
}

// DeleteAutoscale deletes an autoscale rule
func DeleteAutoscale(client kubernetes.Interface, name, ns string) error {
	err := client.AutoscalingV2alpha1().HorizontalPodAutoscalers(ns).Delete(name, &metav1.DeleteOptions{})
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
func CreateServiceMonitor(smclient monitoringv1alpha1.MonitoringV1alpha1Client, funcObj *spec.Function, ns string, or []metav1.OwnerReference) error {
	_, err := smclient.ServiceMonitors(ns).Get(funcObj.Metadata.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			s := &monitoringv1alpha1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      funcObj.Metadata.Name,
					Namespace: ns,
					Labels: map[string]string{
						"service-monitor": "function",
					},
					OwnerReferences: or,
				},
				Spec: monitoringv1alpha1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"function": funcObj.Metadata.Name,
						},
					},
					Endpoints: []monitoringv1alpha1.Endpoint{
						{
							Port: "function-port",
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
func GetOwnerReference(funcObj *spec.Function) ([]metav1.OwnerReference, error) {
	if funcObj.Metadata.Name == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("function name can't be empty")
	}
	if funcObj.Metadata.UID == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("uid of function %s can't be empty", funcObj.Metadata.Name)
	}
	t := true
	return []metav1.OwnerReference{
		{
			Kind:               "Function",
			APIVersion:         "k8s.io",
			Name:               funcObj.Metadata.Name,
			UID:                funcObj.Metadata.UID,
			BlockOwnerDeletion: &t,
		},
	}, nil
}
