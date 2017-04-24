/*
Copyright 2016 Skippbox, Ltd.

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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/bitnami/kubeless/pkg/spec"
	"github.com/bitnami/kubeless/version"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	kerrors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	controllerImage = "bitnami/kubeless-controller"
	kafkaImage      = "wurstmeister/kafka"
	pythonRuntime   = "skippbox/kubeless-python:0.0.5"
	pubsubRuntime   = "skippbox/kubeless-event-consumer:0.0.5"
	nodejsRuntime   = "rosskukulinski/kubeless-nodejs:0.0.0"
	rubyRuntime     = "jbianquettibitnami/kubeless-ruby:0.0.0"
)

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() *kubernetes.Clientset {
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

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
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
	tprconfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
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
		if kerrors.IsNotFound(err) {
			logrus.Fatalf("Function %s is not found", funcName)
		}
		return spec.Function{}, err
	}

	return f, nil
}

// WatchResources looking for changes of custom function objects
func WatchResources(httpClient *http.Client, resourceVersion string) (*http.Response, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}

	return httpClient.Get(fmt.Sprintf("https://%s/apis/k8s.io/v1/functions?watch=true&resourceVersion=%s",
		net.JoinHostPort(host, port), resourceVersion))

}

// CreateK8sResources deploys k8s objects (deploy, svc, configmap) for the function
func CreateK8sResources(ns, name string, spec *spec.FunctionSpec, client *kubernetes.Clientset) error {
	str := strings.Split(spec.Handler, ".")
	if len(str) != 2 {
		return errors.New("Failed: incorrect handler format. It should be module_name.handler_name")
	}
	funcHandler := str[1]
	modName := str[0]
	fileName := modName

	//TODO: Only python and nodejs supported. Add more...
	imageName := ""
	depName := ""
	switch {
	case strings.Contains(spec.Runtime, "python"):
		fileName = modName + ".py"
		imageName = pythonRuntime
		if spec.Type == "PubSub" {
			imageName = pubsubRuntime
		}
		depName = "requirements.txt"
	case strings.Contains(spec.Runtime, "go"):
		fileName = modName + ".go"
	case strings.Contains(spec.Runtime, "nodejs"):
		fileName = modName + ".js"
		imageName = nodejsRuntime
		depName = "package.json"
	case strings.Contains(spec.Runtime, "ruby"):
		fileName = modName + ".rb"
		imageName = rubyRuntime
		depName = "Gemfile"
	}

	//add configmap
	labels := map[string]string{
		"function": name,
	}
	data := map[string]string{
		"handler": spec.Handler,
		fileName:  spec.Function,
		depName:   spec.Deps,
	}
	configMap := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: data,
	}

	_, err := client.Core().ConfigMaps(ns).Create(configMap)
	if err != nil {
		return err
	}

	//add service
	svc := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     v1.ServiceTypeNodePort,
		},
	}
	_, err = client.Core().Services(ns).Create(svc)
	if err != nil {
		return err
	}

	//prepare init-container for custom runtime
	initContainer := []v1.Container{}
	if spec.Deps != "" {
		initContainer = append(initContainer, v1.Container{
			Name:            "install",
			Image:           getInitImage(spec.Runtime),
			Command:         getCommand(spec.Runtime),
			VolumeMounts:    getVolumeMounts(name, spec.Runtime),
			WorkingDir:      "/requirements",
			ImagePullPolicy: v1.PullIfNotPresent,
		})
	}

	//add deployment
	dpm := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					InitContainers: initContainer,
					Containers: []v1.Container{
						{
							Name:            name,
							Image:           imageName,
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "FUNC_HANDLER",
									Value: funcHandler,
								},
								{
									Name:  "MOD_NAME",
									Value: modName,
								},
								{
									Name:  "TOPIC_NAME",
									Value: spec.Topic,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      name,
									MountPath: "/kubeless",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: name,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// update deployment for custom runtime
	if spec.Deps != "" {
		updateDeployment(dpm, spec.Runtime)
	}

	_, err = client.Extensions().Deployments(ns).Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

// DeleteK8sResources removes k8s objects of the function
func DeleteK8sResources(ns, name string, client *kubernetes.Clientset) error {
	replicas := int32(0)
	// scale deployment to 0
	oldScale, err := client.Scales(ns).Get("Deployment", name)
	if err != nil {
		return err
	}

	oldScale.Spec = v1beta1.ScaleSpec{Replicas: replicas}
	_, err = client.Scales(ns).Update("Deployment", oldScale)

	// and delete it
	err = client.Extensions().Deployments(ns).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	// delete svc
	err = client.Core().Services(ns).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	// delete cm
	err = client.Core().ConfigMaps(ns).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateK8sCustomResource will create a custom function object
func CreateK8sCustomResource(runtime, handler, file, funcName, funcType, topic, ns, deps string) error {
	var f spec.Function

	tprClient, err := GetTPRClientOutOfCluster()
	if err != nil {
		return err
	}

	err = tprClient.Get().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)
	if err != nil {
		if kerrors.IsNotFound(err) {
			f := &spec.Function{
				TypeMeta: unversioned.TypeMeta{
					Kind:       "Function",
					APIVersion: "k8s.io/v1",
				},
				Metadata: api.ObjectMeta{
					Name:      funcName,
					Namespace: ns,
				},
				Spec: spec.FunctionSpec{
					Handler:  handler,
					Runtime:  runtime,
					Type:     funcType,
					Function: readFile(file),
					Topic:    topic,
				},
			}
			// add dependencies file to func spec
			if deps != "" {
				f.Spec.Deps = readFile(deps)
			}

			var result spec.Function
			err = tprClient.Post().
				Resource("functions").
				Namespace(ns).
				Body(f).
				Do().Into(&result)

			if err != nil {
				return err
			}
		}
	} else {
		//FIXME: improve the error message
		return fmt.Errorf("Can't create the function")
	}

	return nil
}

// DeleteK8sCustomResource will delete custom function object
func DeleteK8sCustomResource(funcName, ns string) error {
	var f spec.Function

	tprClient, err := GetTPRClientOutOfCluster()
	if err != nil {
		return err
	}

	err = tprClient.Get().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return fmt.Errorf("The function doesn't exist")
		}
	}

	err = tprClient.Delete().
		Resource("functions").
		Namespace(ns).
		Name(funcName).
		Do().Into(&f)

	if err != nil {
		fmt.Println("error here")
		return err
	}

	return nil
}

// DeployKubeless deploys kubeless controller to k8s
func DeployKubeless(client *kubernetes.Clientset, ctlImage string, ctlNamespace string) error {
	//add deployment
	labels := map[string]string{
		"app": "kubeless-controller",
	}
	dpm := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:   "kubeless-controller",
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            ctlNamespace,
							Image:           getImage(ctlImage),
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	//create Kubeless namespace if it's not exists
	_, err := client.Core().Namespaces().Get(ctlNamespace)
	if err != nil {
		ns := &v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: ctlNamespace,
			},
		}
		_, err = client.Core().Namespaces().Create(ns)
		if err != nil {
			return err
		}
	}

	//deploy Kubeless controller
	_, err = client.Extensions().Deployments(ctlNamespace).Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

// DeployMsgBroker deploys kafka-controller
func DeployMsgBroker(client *kubernetes.Clientset, kafkaVer string, ctlNamespace string) error {
	labels := map[string]string{
		"app": "kafka",
	}

	//add zookeeper svc
	svc := &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:   "zookeeper",
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "zookeeper-port",
					Port:       2181,
					TargetPort: intstr.FromInt(2181),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}

	_, err := client.Core().Services(ctlNamespace).Create(svc)
	if err != nil {
		return err
	}

	//add kafka svc
	svc = &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:   "kafka",
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "kafka-port",
					Port:       9092,
					TargetPort: intstr.FromInt(9092),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}

	_, err = client.Core().Services(ctlNamespace).Create(svc)
	if err != nil {
		return err
	}

	//add deployment
	dpm := &v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:   "kafka-controller",
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "kafka",
							Image:           getKafkaImage(kafkaVer),
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "KAFKA_ADVERTISED_HOST_NAME",
									Value: "kafka." + ctlNamespace,
								},
								{
									Name:  "KAFKA_ADVERTISED_PORT",
									Value: "9092",
								},
								{
									Name:  "KAFKA_PORT",
									Value: "9092",
								},
								{
									Name:  "KAFKA_ZOOKEEPER_CONNECT",
									Value: "zookeeper." + ctlNamespace + ":2181",
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9092,
								},
							},
						},
						{
							Name:            "zookeeper",
							Image:           "wurstmeister/zookeeper",
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 2181,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	_, err = client.Extensions().Deployments(ctlNamespace).Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

// GetPodName returns pod to which the function is deployed
func GetPodName(c *kubernetes.Clientset, ns, funcName string) (string, error) {
	po, err := c.Core().Pods(ns).List(v1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, item := range po.Items {
		if strings.Index(item.Name, funcName) == 0 {
			return item.Name, nil
		}
	}

	return "", errors.New("Can't find pod starting with the function name")
}

// getImage returns runtime image of the function
func getImage(v string) string {
	switch v {
	case "":
		return fmt.Sprintf("%s:%s", controllerImage, version.VERSION)
	default:
		return v
	}
}

// getKafkaImage returns corresponding kafka image
func getKafkaImage(v string) string {
	switch v {
	case "":
		return kafkaImage
	default:
		return fmt.Sprintf("%s:%s", kafkaImage, v)
	}
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

func readFile(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Can not read file: %s. The file may not exist", file)
	}
	return string(data[:])
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

// configureClient configures tpr client
func configureClient(config *rest.Config) {
	groupversion := unversioned.GroupVersion{
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
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)
}
