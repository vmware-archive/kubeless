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
	python27Http    = "bitnami/kubeless-python@sha256:6789266df0c97333f76e23efd58cf9c7efe24fa3e83b5fc826fd5cc317699b55"
	python27Pubsub  = "bitnami/kubeless-event-consumer@sha256:5ce469529811acf49c4d20bcd8a675be7aa029b43cf5252a8c9375b170859d83"
	python34Http    = "bitnami/kubeless-python:test@sha256:686cd28cda5fe7bc6db60fa3e8a9a2c57a5eff6a58e66a60179cc1d3fcf1035b"
	python34Pubsub  = "bitnami/kubeless-python-event-consumer@sha256:8f92397258836e9c39948814aa5324c29d96ff3624b66dd70fdbad1ce0a1615e"
	node6Http       = "bitnami/kubeless-nodejs@sha256:b3c7cec77f973bf7a48cbbb8ea5069cacbaee7044683a275c6f78fa248de17b4"
	node6Pubsub     = "bitnami/kubeless-nodejs-event-consumer@sha256:b027bfef5f99c3be68772155a1feaf1f771ab9a3c7bb49bef2e939d6b766abec"
	node8Http       = "bitnami/kubeless-nodejs@sha256:1eff2beae6fcc40577ada75624c3e4d3840a854588526cd8616d66f4e889dfe6"
	node8Pubsub     = "bitnami/kubeless-nodejs-event-consumer@sha256:4d005c9c0b462750d9ab7f1305897e7a01143fe869d3b722ed3330560f9c7fb5"
	ruby24Http      = "bitnami/kubeless-ruby@sha256:97b18ac36bb3aa9529231ea565b339ec00d2a5225cf7eb010cd5a6188cf72ab5"
	ruby24Pubsub    = "bitnami/kubeless-ruby-event-consumer@sha256:938a860dbd9b7fb6b4338248a02c92279315c6e42eed0700128b925d3696b606"
	dotnetcore2Http = "allantargino/kubeless-dotnetcore@sha256:d321dc4b2c420988d98cdaa22c733743e423f57d1153c89c2b99ff0d944e8a63"
	busybox         = "busybox@sha256:be3c11fdba7cfe299214e46edc642e09514dbb9bbefcd0d3836c05a1e0cd0642"
	pubsubFunc      = "PubSub"
	schedFunc       = "Scheduled"
)

type runtimeVersion struct {
	runtimeID   string
	version     string
	httpImage   string
	pubsubImage string
}

var python, node, ruby, dotnetcore []runtimeVersion

func init() {
	python27 := runtimeVersion{runtimeID: "python", version: "2.7", httpImage: python27Http, pubsubImage: python27Pubsub}
	python34 := runtimeVersion{runtimeID: "python", version: "3.4", httpImage: python34Http, pubsubImage: python34Pubsub}
	python = []runtimeVersion{python27, python34}

	node6 := runtimeVersion{runtimeID: "nodejs", version: "6", httpImage: node6Http, pubsubImage: node6Pubsub}
	node8 := runtimeVersion{runtimeID: "nodejs", version: "8", httpImage: node8Http, pubsubImage: node8Pubsub}
	node = []runtimeVersion{node6, node8}

	ruby24 := runtimeVersion{runtimeID: "ruby", version: "2.4", httpImage: ruby24Http, pubsubImage: ruby24Pubsub}
	ruby = []runtimeVersion{ruby24}

	dotnetcore2 := runtimeVersion{runtimeID: "dotnetcore", version: "2.0", httpImage: dotnetcore2Http, pubsubImage: ""}
	dotnetcore = []runtimeVersion{dotnetcore2}
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

func getRuntimeProperties(runtime, modName string) (fileName, depName string, versionsDef []runtimeVersion, err error) {
	runtimeID := regexp.MustCompile("[a-zA-Z]+").FindString(runtime)
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
	case runtimeID == "dotnetcore":
		fileName = modName + ".cs"
		versionsDef = dotnetcore
		depName = "requirements.xml"
	default:
		err = errors.New("The given runtime is not valid")
		return
	}
	return
}

// GetFunctionData given a runtime returns an Image ID, the dependencies filename and the function filename
func GetFunctionData(runtime, ftype, modName string) (imageName, depName, fileName string, err error) {
	err = nil
	imageName = ""
	var versionsDef []runtimeVersion

	version := regexp.MustCompile("[0-9.]+").FindString(runtime)

	fileName, depName, versionsDef, err = getRuntimeProperties(runtime, modName)
	if err != nil {
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

// EnsureK8sResources creates/updates k8s objects (deploy, svc, configmap) for the function
func EnsureK8sResources(funcObj *spec.Function, client kubernetes.Interface) error {
	if len(funcObj.Metadata.Labels) == 0 {
		funcObj.Metadata.Labels = make(map[string]string)
	}
	funcObj.Metadata.Labels["function"] = funcObj.Metadata.Name

	t := true
	or := []metav1.OwnerReference{
		{
			Kind:               "Function",
			APIVersion:         "k8s.io",
			Name:               funcObj.Metadata.Name,
			UID:                funcObj.Metadata.UID,
			BlockOwnerDeletion: &t,
		},
	}

	err := ensureFuncConfigMap(client, funcObj, or)
	if err != nil {
		return err
	}

	err = ensureFuncService(client, funcObj, or)
	if err != nil {
		return err
	}

	err = ensureFuncDeployment(client, funcObj, or)
	if err != nil {
		return err
	}

	if funcObj.Spec.Type == schedFunc {
		err = ensureFuncJob(client, funcObj, or)
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
	}

	// delete deployment
	deletePolicy := metav1.DeletePropagationForeground
	err = client.Extensions().Deployments(ns).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	// delete svc
	err = client.Core().Services(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
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
		return fmt.Errorf("Function %s already existed", f.Metadata.Name)
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

// extract the branch number from the runtime string
func getBranchFromRuntime(runtime string) string {
	re := regexp.MustCompile("[0-9.]+")
	return re.FindString(runtime)
}

// specify image for the init container
func getInitImage(runtime string) string {
	switch {
	case strings.Contains(runtime, "python"):
		branch := getBranchFromRuntime(runtime)
		if branch == "2.7" {
			// TODO: Migrate the image for python 2.7 to an official source (not alpine-based)
			return "tuna/python-pillow:2.7.11-alpine"
		}
		return "python:" + branch
	case strings.Contains(runtime, "nodejs"):
		return "node:6.10-alpine"
	case strings.Contains(runtime, "ruby"):
		return "bitnami/ruby:2.4"
	case strings.Contains(runtime, "dotnetcore"):
		return "microsoft/aspnetcore-build:2.0"
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
	case strings.Contains(runtime, "ruby"):
		return []string{"bundle", "install", "--path", "/rubypath"}
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
	case strings.Contains(runtime, "ruby"):
		return []v1.VolumeMount{
			{
				Name:      "rubypath",
				MountPath: "/rubypath",
			},
			{
				Name:      name,
				MountPath: "/requirements",
			},
		}
	case strings.Contains(runtime, "dotnetcore"):
		return []v1.VolumeMount{
			{
				Name:      "dotnetcorepath",
				MountPath: "/dotnetcorepath",
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
			Value: "/opt/kubeless/pythonpath/lib/python" + getBranchFromRuntime(runtime) + "/site-packages",
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
	case strings.Contains(runtime, "ruby"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "GEM_HOME",
			Value: "/opt/kubeless/rubypath/ruby/2.4.0",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "rubypath",
			MountPath: "/opt/kubeless/rubypath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "rubypath",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	case strings.Contains(runtime, "dotnetcore"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "DOTNETCORE_HOME",
			Value: "/usr/bin/",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "dotnetcorepath",
			MountPath: "/opt/kubeless/dotnetcorepath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
			Name: "dotnetcorepath",
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

	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		return "", err
	}

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
		return "", "", errors.New("Failed: incorrect handler format. It should be module_name.handler_name")
	}

	return str[0], str[1], nil
}

func ensureFuncConfigMap(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	configMapData := map[string]string{}
	var err error
	if funcObj.Spec.Handler != "" {
		modName, _, err := splitHandler(funcObj.Spec.Handler)
		if err != nil {
			return err
		}
		_, depName, fileName, err := GetFunctionData(funcObj.Spec.Runtime, funcObj.Spec.Type, modName)
		if err != nil {
			return err
		}

		configMapData = map[string]string{
			"handler": funcObj.Spec.Handler,
			fileName:  funcObj.Spec.Function,
			depName:   funcObj.Spec.Deps,
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
		var data []byte
		data, err = json.Marshal(configMap)
		if err != nil {
			return err
		}
		_, err = client.Core().ConfigMaps(funcObj.Metadata.Namespace).Patch(configMap.Name, types.StrategicMergePatchType, data)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The configmap may already exist and there is nothing to update
			return nil
		}
	}

	return err
}

func ensureFuncService(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.Metadata.Name,
			Labels:          funcObj.Metadata.Labels,
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
			Selector: funcObj.Metadata.Labels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}
	_, err := client.Core().Services(funcObj.Metadata.Namespace).Create(svc)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		var data []byte
		data, err = json.Marshal(svc)
		if err != nil {
			return err
		}
		_, err = client.Core().Services(funcObj.Metadata.Namespace).Patch(svc.Name, types.StrategicMergePatchType, data)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The service may already exist and there is nothing to update
			return nil
		}
	}
	return err
}

func ensureFuncDeployment(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
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

	dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, v1.Volume{
		Name: funcObj.Metadata.Name,
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

	//only resolve the image name if it has not been already set
	if funcObj.Spec.Handler != "" {
		modName, handlerName, err := splitHandler(funcObj.Spec.Handler)
		if err != nil {
			return err
		}
		if dpm.Spec.Template.Spec.Containers[0].Image == "" {
			imageName, _, _, err := GetFunctionData(funcObj.Spec.Runtime, funcObj.Spec.Type, modName)
			if err != nil {
				return err
			}
			dpm.Spec.Template.Spec.Containers[0].Image = imageName
		}
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  "FUNC_HANDLER",
				Value: handlerName,
			},
			v1.EnvVar{
				Name:  "MOD_NAME",
				Value: modName,
			})
	}

	dpm.Spec.Template.Spec.Containers[0].Name = funcObj.Metadata.Name
	dpm.Spec.Template.Spec.Containers[0].Ports = append(dpm.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{
		ContainerPort: 8080,
	})
	dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  "TOPIC_NAME",
			Value: funcObj.Spec.Topic,
		})
	dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      funcObj.Metadata.Name,
		MountPath: "/kubeless",
	})

	//prepare init-container for custom runtime
	initContainer := v1.Container{}
	if funcObj.Spec.Deps != "" {
		// ensure that the runtime is supported for installing dependencies
		_, _, _, err = getRuntimeProperties(funcObj.Spec.Runtime, "")
		if err != nil {
			return err
		}
		initContainer = v1.Container{
			Name:            "install",
			Image:           getInitImage(funcObj.Spec.Runtime),
			Command:         getCommand(funcObj.Spec.Runtime),
			VolumeMounts:    getVolumeMounts(funcObj.Metadata.Name, funcObj.Spec.Runtime),
			WorkingDir:      "/requirements",
			ImagePullPolicy: v1.PullIfNotPresent,
		}
		initContainer.Env = dpm.Spec.Template.Spec.Containers[0].Env
		// update deployment for custom runtime
		updateDeployment(dpm, funcObj.Spec.Runtime)
		//TODO: remove this when init containers becomes a stable feature
		addInitContainerAnnotation(dpm)
	}
	// only append non-empty initContainer
	if initContainer.Name != "" {
		dpm.Spec.Template.Spec.InitContainers = append(dpm.Spec.Template.Spec.InitContainers, initContainer)
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

	_, err = client.Extensions().Deployments(funcObj.Metadata.Namespace).Create(dpm)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		var data []byte
		data, err = json.Marshal(dpm)
		if err != nil {
			return err
		}
		_, err = client.Extensions().Deployments(funcObj.Metadata.Namespace).Patch(dpm.Name, types.StrategicMergePatchType, data)
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

func ensureFuncJob(client kubernetes.Interface, funcObj *spec.Function, or []metav1.OwnerReference) error {
	job := &v2alpha1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("trigger-%s", funcObj.Metadata.Name),
			Labels:          funcObj.Metadata.Labels,
			OwnerReferences: or,
		},
		Spec: v2alpha1.CronJobSpec{
			Schedule: funcObj.Spec.Schedule,
			JobTemplate: v2alpha1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: busybox,
									Name:  "trigger",
									Args:  []string{"wget", "-qO-", fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", funcObj.Metadata.Name, funcObj.Metadata.Namespace)},
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	_, err := client.BatchV2alpha1().CronJobs(funcObj.Metadata.Namespace).Create(job)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		var data []byte
		data, err = json.Marshal(job)
		if err != nil {
			return err
		}
		_, err = client.BatchV2alpha1().CronJobs(funcObj.Metadata.Namespace).Patch(job.Name, types.StrategicMergePatchType, data)
	}

	return err
}
