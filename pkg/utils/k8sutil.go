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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/skippbox/kubeless/pkg/spec"
	"github.com/skippbox/kubeless/version"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
)

const (
	TIMEOUT          = 300
	CONTROLLER_IMAGE = "bitnami/kubeless-controller"
	KAFKA_IMAGE      = "wurstmeister/kafka"
	PYTHON_RUNTIME   = "skippbox/kubeless-python:0.0.4"
	PUBSUB_RUNTIME   = "skippbox/kubeless-event-consumer:0.0.4"
	NODEJS_RUNTIME   = "rosskukulinski/kubeless-nodejs:0.0.0"
)

func GetFactory() *cmdutil.Factory {
	factory := cmdutil.NewFactory(nil)
	return factory
}

func IsKubernetesResourceAlreadyExistError(err error) bool {
	se, ok := err.(*apierrors.StatusError)
	if !ok {
		return false
	}
	if se.Status().Code == http.StatusConflict && se.Status().Reason == unversionedAPI.StatusReasonAlreadyExists {
		return true
	}
	return false
}

func ListResources(host, ns string, httpClient *http.Client) (*http.Response, error) {
	if host == "localhost" {
		//return httpClient.Get(fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/namespaces/%s/lambdas",
		//	host, ns))
		return httpClient.Get(fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/lambdas",
			host))
	} else {
		//return httpClient.Get(fmt.Sprintf("%s/apis/k8s.io/v1/namespaces/%s/lambdas",
		//	host, ns))
		return httpClient.Get(fmt.Sprintf("%s/apis/k8s.io/v1/lambdas",
			host))
	}

}

func WatchResources(host, ns string, httpClient *http.Client, resourceVersion string) (*http.Response, error) {
	if host == "localhost" {
		//return httpClient.Get(fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/namespaces/%s/lambdas?watch=true&resourceVersion=%s",
		//	host, ns, resourceVersion))
		return httpClient.Get(fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/lambdas?watch=true&resourceVersion=%s",
			host, resourceVersion))
	} else {
		//return httpClient.Get(fmt.Sprintf("https://%s:8443/apis/k8s.io/v1/namespaces/%s/lambdas?watch=true&resourceVersion=%s",
		//	host, ns, resourceVersion))
		return httpClient.Get(fmt.Sprintf("https://%s:8443/apis/k8s.io/v1/lambdas?watch=true&resourceVersion=%s",
			host, resourceVersion))
	}
}

func submitResource(host, ns string, httpClient *http.Client, body io.Reader) (*http.Response, error) {
	if host == "localhost" {
		return httpClient.Post(fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/namespaces/%s/lambdas",
			host, ns), "application/json", body)
	} else {
		return httpClient.Post(fmt.Sprintf("%s/apis/k8s.io/v1/namespaces/%s/lambdas",
			host, ns), "application/json", body)
	}
}

func deleteResource(host, ns, funcName string, httpClient *http.Client) (*http.Response, error) {
	var (
		req *http.Request
		err error
	)

	if host == "localhost" {
		req, err = http.NewRequest("DELETE", fmt.Sprintf("http://%s:8080/apis/k8s.io/v1/namespaces/%s/lambdas/%s",
			host, ns, funcName), nil)
	} else {
		req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/apis/k8s.io/v1/namespaces/%s/lambdas/%s",
			host, ns, funcName), nil)
	}

	if err != nil {
		return nil, err
	}
	resq, err := httpClient.Do(req)
	return resq, err
}

func CreateK8sResources(ns, name string, spec *spec.FunctionSpec, client *client.Client) error {
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
		imageName = PYTHON_RUNTIME
		if spec.Type == "PubSub" {
			imageName = PUBSUB_RUNTIME
		}
		depName = "requirements.txt"
	case strings.Contains(spec.Runtime, "go"):
		fileName = modName + ".go"
	case strings.Contains(spec.Runtime, "nodejs"):
		fileName = modName + ".js"
		imageName = NODEJS_RUNTIME
		depName = "package.json"
	}

	//add configmap
	labels := map[string]string{
		"lambda": name,
	}
	data := map[string]string{
		"handler": spec.Handler,
		fileName:  spec.Lambda,
		depName:   spec.Deps,
	}
	configMap := &api.ConfigMap{
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: data,
	}

	_, err := client.ConfigMaps(ns).Create(configMap)
	if err != nil {
		return err
	}

	//add service
	svc := &api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   api.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     api.ServiceTypeNodePort,
		},
	}
	_, err = client.Services(ns).Create(svc)
	if err != nil {
		return err
	}

	//prepare init-container for custom runtime
	initContainer := []api.Container{}
	if spec.Deps != "" {
		initContainer = append(initContainer, api.Container{
			Name:            "install",
			Image:           getInitImage(spec.Runtime),
			Command:         getCommand(spec.Runtime),
			VolumeMounts:    getVolumeMounts(name, spec.Runtime),
			WorkingDir:      "/requirements",
			ImagePullPolicy: api.PullIfNotPresent,
		})
	}

	//add deployment
	dpm := &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: labels,
				},
				Spec: api.PodSpec{
					InitContainers: initContainer,
					Containers: []api.Container{
						{
							Name:            name,
							Image:           imageName,
							ImagePullPolicy: api.PullAlways,
							Ports: []api.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []api.EnvVar{
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
							VolumeMounts: []api.VolumeMount{
								{
									Name:      name,
									MountPath: "/kubeless",
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: name,
							VolumeSource: api.VolumeSource{
								ConfigMap: &api.ConfigMapVolumeSource{
									LocalObjectReference: api.LocalObjectReference{
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

	_, err = client.Deployments(ns).Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

func DeleteK8sResources(ns, name string, client *client.Client) error {
	rp, err := kubectl.ReaperFor(extensions.Kind("Deployment"), client)
	if err != nil {
		return err
	}
	err = rp.Stop(ns, name, TIMEOUT*time.Second, nil)
	if err != nil {
		return err
	}

	rp, err = kubectl.ReaperFor(api.Kind("Service"), client)
	if err != nil {
		return err
	}
	err = rp.Stop(ns, name, TIMEOUT*time.Second, nil)
	if err != nil {
		return err
	}

	err = client.ConfigMaps(ns).Delete(name)
	if err != nil {
		return err
	}

	return nil
}

func CreateK8sCustomResource(runtime, handler, file, funcName, funcType, topic, ns, deps string) error {
	f := &spec.Function{
		TypeMeta: unversionedAPI.TypeMeta{
			Kind:       "LambDa",
			APIVersion: "k8s.io/v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: funcName,
		},
		Spec: spec.FunctionSpec{
			Handler: handler,
			Runtime: runtime,
			Type:    funcType,
			Lambda:  readFile(file),
			Topic:   topic,
		},
	}

	// add dependencies file to func spec
	if deps != "" {
		f.Spec.Deps = readFile(deps)
	}

	fa := GetFactory()
	kClient, err := fa.Client()
	if err != nil {
		return err
	}
	if ns == "" {
		ns, _, err = fa.DefaultNamespace()
		if err != nil {
			return err
		}
	}
	cfg, err := fa.ClientConfig()
	if err != nil {
		return err
	}
	host := cfg.Host

	funcJson, err := json.Marshal(f)
	if err != nil {
		return err
	}

	_, err = submitResource(host, ns, kClient.RESTClient.Client, bytes.NewBuffer(funcJson))
	return err
}

func DeleteK8sCustomResource(funcName, ns string) error {
	fa := GetFactory()
	kClient, err := fa.Client()
	if err != nil {
		return err
	}
	if ns == "" {
		ns, _, err = fa.DefaultNamespace()
		if err != nil {
			return err
		}
	}

	cfg, err := fa.ClientConfig()
	if err != nil {
		return err
	}
	host := cfg.Host

	_, err = deleteResource(host, ns, funcName, kClient.RESTClient.Client)
	return err
}

func DeployKubeless(client *client.Client, ctlImage string) error {
	//add deployment
	labels := map[string]string{
		"app": "kubeless-controller",
	}
	dpm := &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name:   "kubeless-controller",
			Labels: labels,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: labels,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:            "kubeless",
							Image:           getImage(ctlImage),
							ImagePullPolicy: api.PullAlways,
						},
						{
							Name:            "kubectl",
							Image:           "kelseyhightower/kubectl:1.4.0",
							Args:            []string{"proxy", "-p", "8080"},
							ImagePullPolicy: api.PullAlways,
						},
					},
					RestartPolicy: api.RestartPolicyAlways,
				},
			},
		},
	}

	//create Kubeless namespace if it's not exists
	_, err := client.Namespaces().Get("kubeless")
	if err != nil {
		ns := &api.Namespace{
			ObjectMeta: api.ObjectMeta{
				Name: "kubeless",
			},
		}
		_, err = client.Namespaces().Create(ns)
		if err != nil {
			return err
		}
	}

	//deploy Kubeless controller
	_, err = client.Deployments("kubeless").Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

func DeployMsgBroker(client *client.Client, kafkaVer string) error {
	labels := map[string]string{
		"app": "kafka",
	}

	//add zookeeper svc
	svc := &api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   "zookeeper",
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{
					Name:       "zookeeper-port",
					Port:       2181,
					TargetPort: intstr.FromInt(2181),
					Protocol:   api.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     api.ServiceTypeClusterIP,
		},
	}

	_, err := client.Services("kubeless").Create(svc)
	if err != nil {
		return err
	}

	//add kafka svc
	svc = &api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   "kafka",
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{
					Name:       "kafka-port",
					Port:       9092,
					TargetPort: intstr.FromInt(9092),
					Protocol:   api.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     api.ServiceTypeClusterIP,
		},
	}

	_, err = client.Services("kubeless").Create(svc)
	if err != nil {
		return err
	}

	//add deployment
	dpm := &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name:   "kafka-controller",
			Labels: labels,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: labels,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:            "kafka",
							Image:           getKafkaImage(kafkaVer),
							ImagePullPolicy: api.PullIfNotPresent,
							Env: []api.EnvVar{
								{
									Name:  "KAFKA_ADVERTISED_HOST_NAME",
									Value: "kafka.kubeless",
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
									Value: "zookeeper.kubeless:2181",
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 9092,
								},
							},
						},
						{
							Name:            "zookeeper",
							Image:           "wurstmeister/zookeeper",
							ImagePullPolicy: api.PullIfNotPresent,
							Ports: []api.ContainerPort{
								{
									ContainerPort: 2181,
								},
							},
						},
					},
					RestartPolicy: api.RestartPolicyAlways,
				},
			},
		},
	}

	_, err = client.Deployments("kubeless").Create(dpm)
	if err != nil {
		return err
	}

	return nil
}

func GetPodName(c *client.Client, ns, funcName string) (string, error) {
	po, err := c.Pods(ns).List(api.ListOptions{})
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

func getImage(v string) string {
	switch v {
	case "":
		return fmt.Sprintf("%s:%s", CONTROLLER_IMAGE, version.VERSION)
	default:
		return v
	}
}

func getKafkaImage(v string) string {
	switch v {
	case "":
		return KAFKA_IMAGE
	default:
		return fmt.Sprintf("%s:%s", KAFKA_IMAGE, v)
	}
}

// specify image for the init container
func getInitImage(runtime string) string {
	switch {
	case strings.Contains(runtime, "python"):
		return "python:2.7.11-alpine"
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
func getVolumeMounts(name, runtime string) []api.VolumeMount {
	switch {
	case strings.Contains(runtime, "python"):
		return []api.VolumeMount{
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
		return []api.VolumeMount{
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
		return []api.VolumeMount{}
	}
}

func readFile(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("Can not read file: %s. The file may not exist", file)
	}
	return string(data[:])
}

// update deployment object in case of custom runtime
func updateDeployment(dpm *extensions.Deployment, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, api.EnvVar{
			Name:  "PYTHONPATH",
			Value: "/opt/kubeless/pythonpath/lib/python2.7/site-packages",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, api.VolumeMount{
			Name:      "pythonpath",
			MountPath: "/opt/kubeless/pythonpath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, api.Volume{
			Name: "pythonpath",
			VolumeSource: api.VolumeSource{
				EmptyDir: &api.EmptyDirVolumeSource{},
			},
		})
	case strings.Contains(runtime, "nodejs"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, api.EnvVar{
			Name:  "NODE_PATH",
			Value: "/opt/kubeless/nodepath/node_modules",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, api.VolumeMount{
			Name:      "nodepath",
			MountPath: "/opt/kubeless/nodepath",
		})
		dpm.Spec.Template.Spec.Volumes = append(dpm.Spec.Template.Spec.Volumes, api.Volume{
			Name: "nodepath",
			VolumeSource: api.VolumeSource{
				EmptyDir: &api.EmptyDirVolumeSource{},
			},
		})
	}
}
