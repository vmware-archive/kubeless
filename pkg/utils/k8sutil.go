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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/batch/v2alpha1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	python27Http   = "bitnami/kubeless-python@sha256:6789266df0c97333f76e23efd58cf9c7efe24fa3e83b5fc826fd5cc317699b55"
	python27Pubsub = "bitnami/kubeless-event-consumer@sha256:5ce469529811acf49c4d20bcd8a675be7aa029b43cf5252a8c9375b170859d83"
	node6Http      = "bitnami/kubeless-nodejs@sha256:b3c7cec77f973bf7a48cbbb8ea5069cacbaee7044683a275c6f78fa248de17b4"
	node6Pubsub    = "bitnami/kubeless-nodejs-event-consumer@sha256:b027bfef5f99c3be68772155a1feaf1f771ab9a3c7bb49bef2e939d6b766abec"
	node8Http      = "bitnami/kubeless-nodejs@sha256:1eff2beae6fcc40577ada75624c3e4d3840a854588526cd8616d66f4e889dfe6"
	node8Pubsub    = "bitnami/kubeless-nodejs-event-consumer@sha256:4d005c9c0b462750d9ab7f1305897e7a01143fe869d3b722ed3330560f9c7fb5"
	ruby24Http     = "bitnami/kubeless-ruby@sha256:98e95c41652a7a0149421157c2dfb64b31e0d406b8c46c8bc89bd54e50f9898d"
	busybox        = "tuna/busybox@sha256:3651f7ee3b1e779471e338dcc43e6a5e69e0a8c7a4d08fd5702531cbbacc3269"
	pubsubFunc     = "PubSub"
	schedFunc      = "Scheduled"
)

type runtimeVersion struct {
	runtimeID   string
	version     string
	httpImage   string
	pubsubImage string
}

var python, node, ruby []runtimeVersion

func init() {
	python27 := runtimeVersion{runtimeID: "python", version: "2.7", httpImage: python27Http, pubsubImage: python27Pubsub}
	python = []runtimeVersion{python27}

	node6 := runtimeVersion{runtimeID: "nodejs", version: "6", httpImage: node6Http, pubsubImage: node6Pubsub}
	node8 := runtimeVersion{runtimeID: "nodejs", version: "8", httpImage: node8Http, pubsubImage: node8Pubsub}
	node = []runtimeVersion{node6, node8}

	ruby24 := runtimeVersion{runtimeID: "ruby", version: "2.4", httpImage: ruby24Http, pubsubImage: ""}
	ruby = []runtimeVersion{ruby24}
}

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
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
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

// GetTPRClient returns tpr client to the request from inside of cluster
func GetTPRClient() (*rest.RESTClient, error) {
	tprconfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	configureClient(tprconfig)

	tprclient, err := rest.RESTClientFor(tprconfig)
	if err != nil {
		return nil, err
	}

	return tprclient, nil
}

// GetTPRClientOutOfCluster returns tpr client to the request from outside of cluster
func GetTPRClientOutOfCluster() (*rest.RESTClient, error) {
	tprconfig, err := BuildOutOfClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	configureClient(tprconfig)

	tprclient, err := rest.RESTClientFor(tprconfig)
	if err != nil {
		return nil, err
	}

	return tprclient, nil
}

// GetFunction returns specification of a function
func GetFunction(funcName, ns string) (spec.Function, error) {
	var f spec.Function

	tprClient, err := GetTPRClientOutOfCluster()
	if err != nil {
		return spec.Function{}, err
	}

	err = tprClient.Get().
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

func getAvailableRuntimes(imageType string) []string {
	runtimeObjList := [][]runtimeVersion{python, node, ruby}
	var runtimeList []string
	for i := range runtimeObjList {
		for j := range runtimeObjList[i] {
			if (imageType == "PubSub" && runtimeObjList[i][j].pubsubImage != "") || (imageType == "HTTP" && runtimeObjList[i][j].httpImage != "") {
				runtimeList = append(runtimeList, runtimeObjList[i][j].runtimeID+runtimeObjList[i][j].version)
			}
		}
	}
	return runtimeList
}

// GetFunctionData given a runtime returns an Image ID, the dependencies filename and the function filename
func GetFunctionData(runtime, ftype, modName string) (imageName, depName, fileName string, err error) {
	err = nil
	imageName = ""
	depName = ""
	fileName = ""

	runtimeID := regexp.MustCompile("[a-zA-Z]+").FindString(runtime)
	version := regexp.MustCompile("[0-9.]+").FindString(runtime)

	var versionsDef []runtimeVersion
	switch {
	case runtimeID == "python":
		fileName = modName + ".py"
		versionsDef = python
		depName = "requirements.txt"
	case runtimeID == "nodejs":
		fileName = modName + ".js"
		versionsDef = node
		depName = "package.json"
	case runtimeID == "ruby":
		fileName = modName + ".rb"
		versionsDef = ruby
		depName = "Gemfile"
	default:
		err = errors.New("The given runtime is not valid")
		return
	}
	imageNameEnvVar := ""
	if ftype == pubsubFunc {
		imageNameEnvVar = strings.ToUpper(runtime) + "_PUBSUB_RUNTIME"
	} else {
		imageNameEnvVar = strings.ToUpper(runtime) + "_RUNTIME"
	}
	if imageName = os.Getenv(imageNameEnvVar); imageName == "" {
		rVersion := runtimeVersion{"", "", "", ""}
		for i := range versionsDef {
			if versionsDef[i].version == version {
				rVersion = versionsDef[i]
				break
			}
		}
		if ftype == pubsubFunc {
			if rVersion.pubsubImage == "" {
				err = errors.New("The given runtime and version '" + runtime + "does not have a valid image for event based functions. Available runtimes are: " + strings.Join(getAvailableRuntimes("PubSub")[:], ", "))
			} else {
				imageName = rVersion.pubsubImage
			}
		} else {
			if rVersion.httpImage == "" {
				err = errors.New("The given runtime and version '" + runtime + "' does not have a valid image for HTTP based functions. Available runtimes are: " + strings.Join(getAvailableRuntimes("HTTP")[:], ", "))
			} else {
				imageName = rVersion.httpImage
			}
		}
	}
	return
}

type metadata struct {
	Name       string
	Namespace  string
	Handler    string
	Module     string
	Image      string
	File       string
	Dependency string
	Labels     map[string]string
}

// EnsureK8sResources creates/updates k8s objects (deploy, svc, configmap) for the function
func EnsureK8sResources(ns, name string, funcObj *spec.Function, client kubernetes.Interface) error {
	str := strings.Split(funcObj.Spec.Handler, ".")
	if len(str) != 2 {
		return errors.New("Failed: incorrect handler format. It should be module_name.handler_name")
	}

	funcHandler := str[1]
	modName := str[0]

	fileName := modName
	imageName, depName, fileName, err := GetFunctionData(funcObj.Spec.Runtime, funcObj.Spec.Type, modName)
	if err != nil {
		return err
	}

	labels := map[string]string{
		"function": name,
	}
	for k, v := range funcObj.Metadata.Labels {
		labels[k] = v
	}

	funcMeta := metadata{
		Name:       name,
		Namespace:  ns,
		Handler:    funcHandler,
		Module:     modName,
		Image:      imageName,
		File:       fileName,
		Dependency: depName,
		Labels:     labels,
	}

	t := true
	or := []metav1.OwnerReference{
		{
			Kind:               "Function",
			APIVersion:         "k8s.io",
			Name:               funcMeta.Name,
			UID:                funcObj.Metadata.UID,
			BlockOwnerDeletion: &t,
		},
	}

	err = ensureFuncConfigMap(client, funcMeta, funcObj, or)
	if err != nil {
		return err
	}

	err = ensureFuncService(client, funcMeta, or)
	if err != nil {
		return err
	}

	err = ensureFuncDeployment(client, funcMeta, funcObj, or)
	if err != nil {
		return err
	}

	if funcObj.Spec.Type == schedFunc {
		err = ensureFuncJob(client, funcMeta, funcObj, or)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteK8sResources removes k8s objects of the function
func DeleteK8sResources(ns, name string, client kubernetes.Interface) error {
	//check if func is scheduled or not
	_, err := client.BatchV2alpha1().CronJobs(ns).Get(fmt.Sprintf("trigger-%s", name), metav1.GetOptions{})
	if err == nil {
		err = client.BatchV2alpha1().CronJobs(ns).Delete(fmt.Sprintf("trigger-%s", name), &metav1.DeleteOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
	} else {
		deletePolicy := metav1.DeletePropagationForeground
		err := client.Extensions().Deployments(ns).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
		// delete svc
		err = client.Core().Services(ns).Delete(name, &metav1.DeleteOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
	}

	// delete cm
	err = client.Core().ConfigMaps(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	return nil
}

// CreateK8sCustomResource will create a custom function object
func CreateK8sCustomResource(tprClient rest.Interface, f *spec.Function) error {
	err := tprClient.Get().
		Resource("functions").
		Namespace(f.Metadata.Namespace).
		Name(f.Metadata.Name).
		Do().Into(f)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			var result spec.Function
			err = tprClient.Post().
				Resource("functions").
				Namespace(f.Metadata.Namespace).
				Body(f).
				Do().Into(&result)

			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Function has already existed")
	}

	return nil
}

// UpdateK8sCustomResource applies changes to the function custom object
func UpdateK8sCustomResource(f *spec.Function) error {
	fa := cmdutil.NewFactory(nil)
	funcJSON, err := json.Marshal(f)
	if err != nil {
		return err
	}

	// TODO: looking for a way to not writing to temp file
	filename := os.TempDir() + "/.func.json"
	err = ioutil.WriteFile(filename, funcJSON, 0644)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	buferr := bytes.NewBuffer([]byte{})
	applyCmd := cmd.NewCmdApply(fa, buf, buferr)

	applyCmd.Flags().Set("filename", filename)
	applyCmd.Flags().Set("output", "name")
	applyCmd.Run(applyCmd, []string{})

	// remove temp func file
	err = os.Remove(filename)
	if err != nil {
		return err
	}

	return err
}

// DeleteK8sCustomResource will delete custom function object
func DeleteK8sCustomResource(tprClient *rest.RESTClient, funcName, ns string) error {
	var f spec.Function
	err := tprClient.Get().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return fmt.Errorf("The function doesn't exist")
		}
	}

	err = tprClient.Delete().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)

	if err != nil {
		return err
	}

	return nil
}

func getResource() v1.ResourceList {
	r := make(map[v1.ResourceName]resource.Quantity)
	r[v1.ResourceStorage], _ = resource.ParseQuantity("1Gi")
	return r
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
	return v1.Pod{}, errors.New("There is no pod ready")
}

// specify image for the init container
func getInitImage(runtime string) string {
	switch {
	case strings.Contains(runtime, "python"):
		return "tuna/python-pillow:2.7.11-alpine"
	case strings.Contains(runtime, "nodejs"):
		return "node:6.10-alpine"
	default:
		return ""
	}
}

// specify command for the init container
func getCommand(runtime string) []string {
	switch {
	case strings.Contains(runtime, "python"):
		return []string{"pip", "install", "--prefix=/pythonpath", "-r", "/requirements/requirements.txt"}
	case strings.Contains(runtime, "nodejs"):
		return []string{"/bin/sh", "-c", "cp package.json /nodepath && npm install --prefix=/nodepath"}
	default:
		return []string{}
	}
}

// specify volumes for the init container
func getVolumeMounts(name, runtime string) []v1.VolumeMount {
	switch {
	case strings.Contains(runtime, "python"):
		return []v1.VolumeMount{
			{
				Name:      "pythonpath",
				MountPath: "/pythonpath",
			},
			{
				Name:      name,
				MountPath: "/requirements",
			},
		}
	case strings.Contains(runtime, "nodejs"):
		return []v1.VolumeMount{
			{
				Name:      "nodepath",
				MountPath: "/nodepath",
			},
			{
				Name:      name,
				MountPath: "/requirements",
			},
		}
	default:
		return []v1.VolumeMount{}
	}
}

// update deployment object in case of custom runtime
func updateDeployment(dpm *v1beta1.Deployment, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "PYTHONPATH",
			Value: "/opt/kubeless/pythonpath/lib/python2.7/site-packages",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "pythonpath",
			MountPath: "/opt/kubeless/pythonpath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "pythonpath",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	case strings.Contains(runtime, "nodejs"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "NODE_PATH",
			Value: "/opt/kubeless/nodepath/node_modules",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "nodepath",
			MountPath: "/opt/kubeless/nodepath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "nodepath",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}
}

// update cronjob object in case of custom runtime
func updateJob(job *v2alpha1.CronJob, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env = append(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "PYTHONPATH",
			Value: "/opt/kubeless/pythonpath/lib/python2.7/site-packages",
		})
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "pythonpath",
			MountPath: "/opt/kubeless/pythonpath",
		})
		job.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(job.Spec.JobTemplate.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "pythonpath",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	case strings.Contains(runtime, "nodejs"):
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env = append(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "NODE_PATH",
			Value: "/opt/kubeless/nodepath/node_modules",
		})
		job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "nodepath",
			MountPath: "/opt/kubeless/nodepath",
		})
		job.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(job.Spec.JobTemplate.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "nodepath",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}
}

// configureClient configures tpr client
func configureClient(config *rest.Config) {
	groupversion := schema.GroupVersion{
		Group:   "k8s.io",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

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

// addInitContainerAnnotationToJob is a hot fix to add annotation to cronjob for init container to run
func addInitContainerAnnotationToJob(job *v2alpha1.CronJob) error {
	if len(job.Spec.JobTemplate.Spec.Template.Spec.InitContainers) > 0 {
		value, err := json.Marshal(job.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
		if err != nil {
			return err
		}
		if job.Spec.JobTemplate.Spec.Template.Annotations == nil {
			job.Spec.JobTemplate.Spec.Template.Annotations = make(map[string]string)
		}
		job.Spec.JobTemplate.Spec.Template.Annotations[v1.PodInitContainersAnnotationKey] = string(value)
		job.Spec.JobTemplate.Spec.Template.Annotations[v1.PodInitContainersBetaAnnotationKey] = string(value)
	}
	return nil
}

// CreateIngress creates ingress rule for a specific function
func CreateIngress(client kubernetes.Interface, ingressName, funcName, hostname, ns string, enableTLSAcme bool) error {

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: ns,
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
										ServiceName: funcName,
										ServicePort: intstr.FromInt(8080),
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

	_, err := client.ExtensionsV1beta1().Ingresses(ns).Create(ingress)
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

	host, _, _ := net.SplitHostPort(url.Host)

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

func ensureFuncConfigMap(client kubernetes.Interface, funcMeta metadata, funcObj *spec.Function, or []metav1.OwnerReference) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcMeta.Name,
			Labels:          funcMeta.Labels,
			OwnerReferences: or,
		},
		Data: map[string]string{
			"handler":           funcObj.Spec.Handler,
			funcMeta.File:       funcObj.Spec.Function,
			funcMeta.Dependency: funcObj.Spec.Deps,
		},
	}

	_, err := client.Core().ConfigMaps(funcMeta.Namespace).Create(configMap)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		data, _ := json.Marshal(configMap)
		_, err = client.Core().ConfigMaps(funcMeta.Namespace).Patch(configMap.Name, types.StrategicMergePatchType, data)
	}

	return err
}

func ensureFuncService(client kubernetes.Interface, funcMeta metadata, or []metav1.OwnerReference) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcMeta.Name,
			Labels:          funcMeta.Labels,
			OwnerReferences: or,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: funcMeta.Labels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}
	_, err := client.Core().Services(funcMeta.Namespace).Create(svc)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		data, _ := json.Marshal(svc)
		_, err = client.Core().Services(funcMeta.Namespace).Patch(svc.Name, types.StrategicMergePatchType, data)

	}
	return err
}

func ensureFuncDeployment(client kubernetes.Interface, funcMeta metadata, funcObj *spec.Function, or []metav1.OwnerReference) error {
	//prepare init-container for custom runtime
	initContainer := v1.Container{}
	if funcObj.Spec.Deps != "" {
		initContainer = v1.Container{
			Name:            "install",
			Image:           getInitImage(funcObj.Spec.Runtime),
			Command:         getCommand(funcObj.Spec.Runtime),
			VolumeMounts:    getVolumeMounts(funcMeta.Name, funcObj.Spec.Runtime),
			WorkingDir:      "/requirements",
			ImagePullPolicy: v1.PullIfNotPresent,
		}
	}

	podAnnotations := map[string]string{
		// Attempt to attract the attention of prometheus.
		// For runtimes that don't support /metrics,
		// prometheus will get a 404 and mostly silently
		// ignore the pod (still displayed in the list of
		// "targets")
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   "8080",
	}

	//add deployment
	maxUnavailable := intstr.FromInt(0)
	dpm := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcMeta.Name,
			Labels:          funcMeta.Labels,
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
	for k, v := range funcMeta.Labels {
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

	// only append non-empty initContainer
	if initContainer.Name != "" {
		dpm.Spec.Template.Spec.InitContainers = append(dpm.Spec.Template.Spec.InitContainers, initContainer)
	}

	dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
		Name: funcMeta.Name,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: funcMeta.Name,
				},
			},
		},
	})
	if len(dpm.Spec.Template.Spec.Containers) == 0 {
		dpm.Spec.Template.Spec.Containers = append(dpm.Spec.Template.Spec.Containers, v1.Container{})
	}
	dpm.Spec.Template.Spec.Containers[0].Image = funcMeta.Image
	dpm.Spec.Template.Spec.Containers[0].Name = funcMeta.Name
	dpm.Spec.Template.Spec.Containers[0].Ports = append(dpm.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{
		ContainerPort: 8080,
	})
	dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  "FUNC_HANDLER",
			Value: funcMeta.Handler,
		},
		v1.EnvVar{
			Name:  "MOD_NAME",
			Value: funcMeta.Module,
		},
		v1.EnvVar{
			Name:  "TOPIC_NAME",
			Value: funcObj.Spec.Topic,
		})
	dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      funcMeta.Name,
		MountPath: "/kubeless",
	})

	// update deployment for custom runtime
	if funcObj.Spec.Deps != "" {
		updateDeployment(dpm, funcObj.Spec.Runtime)
		//TODO: remove this when init containers becomes a stable feature
		addInitContainerAnnotation(dpm)
	}

	// add liveness Probe to deployment
	if funcObj.Spec.Type != pubsubFunc {
		livenessProbe := &v1.Probe{
			InitialDelaySeconds: int32(3),
			PeriodSeconds:       int32(30),
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(8080),
				},
			},
		}
		dpm.Spec.Template.Spec.Containers[0].LivenessProbe = livenessProbe
	}

	_, err = client.Extensions().Deployments(funcMeta.Namespace).Create(dpm)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		data, _ := json.Marshal(dpm)
		_, err = client.Extensions().Deployments(funcMeta.Namespace).Patch(dpm.Name, types.StrategicMergePatchType, data)
		if err != nil {
			return err
		}

		// kick existing function pods then it will be recreated
		// with the new data mount from updated configmap.
		// TODO: This is a workaround.  Do something better.
		pods, err := GetPodsByLabel(client, funcMeta.Namespace, "function", funcMeta.Name)
		for _, pod := range pods.Items {
			err = client.Core().Pods(funcMeta.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				// non-fatal
				logrus.Warnf("Unable to delete pod %s/%s, may be running stale version of function: %v", funcMeta.Namespace, pod.Name, err)
			}
		}
	}

	return err
}

func ensureFuncJob(client kubernetes.Interface, funcMeta metadata, funcObj *spec.Function, or []metav1.OwnerReference) error {
	podAnnotations := map[string]string{
		// Attempt to attract the attention of prometheus.
		// For runtimes that don't support /metrics,
		// prometheus will get a 404 and mostly silently
		// ignore the pod (still displayed in the list of
		// "targets")
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   "8080",
	}

	job := &v2alpha1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("trigger-%s", funcMeta.Name),
			Labels:          funcMeta.Labels,
			OwnerReferences: or,
		},
		Spec: v2alpha1.CronJobSpec{
			Schedule: funcObj.Spec.Schedule,
			JobTemplate: v2alpha1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: podAnnotations,
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: busybox,
									Name:  "trigger",
									Args:  []string{"curl", fmt.Sprintf("http://%s.%s.svc:8080", funcMeta.Name, funcMeta.Namespace)},
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	_, err := client.BatchV2alpha1().CronJobs(funcMeta.Namespace).Create(job)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		data, _ := json.Marshal(job)
		_, err = client.BatchV2alpha1().CronJobs(funcMeta.Namespace).Patch(job.Name, types.StrategicMergePatchType, data)
		if err != nil {
			return err
		}
	}

	return err
}
