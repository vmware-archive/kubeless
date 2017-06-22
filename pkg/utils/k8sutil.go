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
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"

	appsv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	controllerImage = "bitnami/kubeless-controller@sha256:d07986d575a80179ae15205c6fa5eb3bf9f4f4f46235c79ad6b284e5d3df22d0"
	pythonRuntime   = "bitnami/kubeless-python@sha256:2d0412e982a8e831dee056aee49089e1d5edd65470e479dcbc7d60bb56ea2b71"
	pubsubRuntime   = "bitnami/kubeless-event-consumer@sha256:9c29b8ec6023040492226a55b19781bc3a8911d535327c773ee985895515e905"
	kafkaImage      = "bitnami/kafka@sha256:0b7c8b790546ddb9dcd7e8ff4d50f030fc496176238f36789537620bb13fb54c"
	zookeeperImage  = "bitnami/zookeeper@sha256:2244fba9d7c35df85f078ffdbf77ec9f9b44dad40752f15dd619a85d70aec22d"
	nodejsRuntime   = "rosskukulinski/kubeless-nodejs:0.0.0"
	rubyRuntime     = "jbianquettibitnami/kubeless-ruby:0.0.0"
	pubsubFunc      = "PubSub"
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

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster() kubernetes.Interface {
	config, err := buildOutOfClusterConfig()
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
	tprconfig, err := buildOutOfClusterConfig()
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

// EnsureK8sResources creates/updates k8s objects (deploy, svc, configmap) for the function
func EnsureK8sResources(ns, name string, funcObj *spec.Function, client kubernetes.Interface) error {
	str := strings.Split(funcObj.Spec.Handler, ".")
	if len(str) != 2 {
		return errors.New("Failed: incorrect handler format. It should be module_name.handler_name")
	}
	funcHandler := str[1]
	modName := str[0]
	fileName := modName

	imageName := ""
	depName := ""
	switch {
	case strings.Contains(funcObj.Spec.Runtime, "python"):
		fileName = modName + ".py"
		imageName = pythonRuntime
		if funcObj.Spec.Type == "PubSub" {
			imageName = pubsubRuntime
		}
		depName = "requirements.txt"
	case strings.Contains(funcObj.Spec.Runtime, "go"):
		fileName = modName + ".go"
	case strings.Contains(funcObj.Spec.Runtime, "nodejs"):
		fileName = modName + ".js"
		imageName = nodejsRuntime
		depName = "package.json"
	case strings.Contains(funcObj.Spec.Runtime, "ruby"):
		fileName = modName + ".rb"
		imageName = rubyRuntime
		depName = "Gemfile"
	}

	//add configmap
	labels := map[string]string{
		"function": name,
	}

	t := true
	or := []metav1.OwnerReference{
		{
			Kind:               "Function",
			APIVersion:         "k8s.io",
			Name:               name,
			UID:                funcObj.Metadata.UID,
			BlockOwnerDeletion: &t,
		},
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
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: or,
		},
		Data: map[string]string{
			"handler": funcObj.Spec.Handler,
			fileName:  funcObj.Spec.Function,
			depName:   funcObj.Spec.Deps,
		},
	}

	_, err := client.Core().ConfigMaps(ns).Create(configMap)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = client.Core().ConfigMaps(ns).Update(configMap)
	}
	if err != nil {
		return err
	}

	//add service
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
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
			Selector: labels,
			Type:     v1.ServiceTypeNodePort,
		},
	}
	_, err = client.Core().Services(ns).Create(svc)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = client.Core().Services(ns).Update(svc)
	}
	if err != nil {
		return err
	}

	//prepare init-container for custom runtime
	initContainer := []v1.Container{}
	if funcObj.Spec.Deps != "" {
		initContainer = append(initContainer, v1.Container{
			Name:            "install",
			Image:           getInitImage(funcObj.Spec.Runtime),
			Command:         getCommand(funcObj.Spec.Runtime),
			VolumeMounts:    getVolumeMounts(name, funcObj.Spec.Runtime),
			WorkingDir:      "/requirements",
			ImagePullPolicy: v1.PullIfNotPresent,
		})
	}
	//add deployment
	maxUnavailable := intstr.FromInt(0)
	dpm := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: or,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: podAnnotations,
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
									Value: funcObj.Spec.Topic,
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
			Strategy: v1beta1.DeploymentStrategy{
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
				},
			},
		},
	}

	// update deployment for custom runtime
	if funcObj.Spec.Deps != "" {
		updateDeployment(dpm, funcObj.Spec.Runtime)
		//TODO: remove this when init containers becomes a stable feature
		addInitContainerAnnotation(dpm)
	}

	if funcObj.Spec.Type != pubsubFunc {
		livenessProbe := &v1.Probe{
			InitialDelaySeconds: int32(3),
			PeriodSeconds:       int32(3),
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(8080),
				},
			},
		}
		dpm.Spec.Template.Spec.Containers[0].LivenessProbe = livenessProbe
	}

	_, err = client.Extensions().Deployments(ns).Create(dpm)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = client.Extensions().Deployments(ns).Update(dpm)

		// kick existing function pods then it will be recreated
		// with the new data mount from updated configmap.
		// TODO: This is a workaround.  Do something better.
		pods, err := GetPodsByLabel(client, ns, "function", name)
		for _, pod := range pods.Items {
			err = client.Core().Pods(ns).Delete(pod.Name, &metav1.DeleteOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				// non-fatal
				logrus.Warnf("Unable to delete pod %s/%s, may be running stale version of function: %v", ns, pod.Name, err)
			}
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// DeleteK8sResources removes k8s objects of the function
func DeleteK8sResources(ns, name string, client kubernetes.Interface) error {
	deploy, err := client.Extensions().Deployments(ns).Get(name, metav1.GetOptions{})
	if err == nil {
		//scale deployment to 0
		replicas := int32(0)
		deploy.Spec.Replicas = &replicas
		_, _ = client.Extensions().Deployments(ns).Update(deploy)
	}

	// delete deployment
	err = client.Extensions().Deployments(ns).Delete(name, &metav1.DeleteOptions{})
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
		if k8sErrors.IsNotFound(err) {
			f := &spec.Function{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Function",
					APIVersion: "k8s.io/v1",
				},
				Metadata: metav1.ObjectMeta{
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

// UpdateK8sCustomResource applies changes to the function custom object
func UpdateK8sCustomResource(runtime, handler, file, funcName, ns string) error {
	fa := cmdutil.NewFactory(nil)

	if ns == "" {
		ns, _, _ = fa.DefaultNamespace()
	}

	f := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: ns,
		},
		Spec: spec.FunctionSpec{
			Handler:  handler,
			Function: readFile(file),
			Runtime:  runtime,
		},
	}

	funcJSON, err := json.Marshal(f)
	if err != nil {
		return err
	}

	// TODO: looking for a way to not writing to temp file
	err = ioutil.WriteFile(".func.json", funcJSON, 0644)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	buferr := bytes.NewBuffer([]byte{})
	applyCmd := cmd.NewCmdApply(fa, buf, buferr)

	applyCmd.Flags().Set("filename", ".func.json")
	applyCmd.Flags().Set("output", "name")
	applyCmd.Run(applyCmd, []string{})

	// remove temp func file
	err = os.Remove(".func.json")
	if err != nil {
		return err
	}

	return err
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
		fmt.Println("error here")
		return err
	}

	return nil
}

// DeployKubeless deploys kubeless controller to k8s
func DeployKubeless(client kubernetes.Interface, ctlNamespace string) error {
	//add deployment
	labels := map[string]string{
		"kubeless": "controller",
	}
	dpm := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubeless-controller",
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            ctlNamespace,
							Image:           getImage("kubeless"),
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	//create Kubeless namespace if it's not exists
	_, err := client.Core().Namespaces().Get(ctlNamespace, metav1.GetOptions{})
	if err != nil {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
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

func getResource() v1.ResourceList {
	r := make(map[v1.ResourceName]resource.Quantity)
	r[v1.ResourceStorage], _ = resource.ParseQuantity("1Gi")
	return r
}

// DeployMsgBroker deploys kafka-controller
func DeployMsgBroker(client kubernetes.Interface, ctlNamespace string) error {
	labels := map[string]string{
		"kubeless": "kafka",
	}

	//add kafka svc
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kafka",
			//Labels: labels,
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

	_, err := client.Core().Services(ctlNamespace).Create(svc)
	if err != nil {
		return err
	}

	//add kafka headless svc
	svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "broker",
			//Labels: labels,
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
			Selector:  labels,
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
		},
	}

	_, err = client.Core().Services(ctlNamespace).Create(svc)
	if err != nil {
		return err
	}

	//add kafka statefulset
	sts := &appsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kafka",
			//Labels: labels,
		},
		Spec: appsv1beta1.StatefulSetSpec{
			ServiceName: "broker",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "broker",
							Image:           getImage("kafka"),
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "KAFKA_ADVERTISED_HOST_NAME",
									Value: "broker." + ctlNamespace,
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
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/opt/bitnami/kafka/data",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "datadir",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: getResource(),
						},
					},
				},
			},
		},
	}

	_, err = client.Apps().StatefulSets(ctlNamespace).Create(sts)
	if err != nil {
		return err
	}

	labels = map[string]string{
		"kubeless": "zookeeper",
	}

	//add zookeeper svc
	svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zookeeper",
			//Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "client",
					Port:       2181,
					TargetPort: intstr.FromInt(2181),
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

	//add zookeeper headless svc
	svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zoo",
			//Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "peer",
					Port:       2888,
					TargetPort: intstr.FromInt(2888),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "leader-election",
					Port:       3888,
					TargetPort: intstr.FromInt(3888),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector:  labels,
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
		},
	}

	_, err = client.Core().Services(ctlNamespace).Create(svc)
	if err != nil {
		return err
	}

	//add zookeeper statefulset
	sts = &appsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zoo",
			//Labels: labels,
		},
		Spec: appsv1beta1.StatefulSetSpec{
			ServiceName: "zoo",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "zoo",
							Image:           getImage("zookeeper"),
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "ZOO_SERVERS",
									Value: "server.1=zoo-0.zoo:2888:3888:participant",
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 2181,
									Name:          "client",
								},
								{
									ContainerPort: 2888,
									Name:          "peer",
								},
								{
									ContainerPort: 3888,
									Name:          "leader-election",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "zookeeper",
									MountPath: "/bitnami/zookeeper",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
					Volumes: []v1.Volume{
						{
							Name: "zookeeper",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.Apps().StatefulSets(ctlNamespace).Create(sts)
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

// getImage returns corresponding controller image
func getImage(v string) string {
	switch v {
	case "kafka":
		img := os.Getenv("KAFKA_IMAGE")
		if img != "" {
			return img
		}
		return kafkaImage
	case "zookeeper":
		img := os.Getenv("ZOOKEEPER_IMAGE")
		if img != "" {
			return img
		}
		return zookeeperImage
	case "kubeless":
		img := os.Getenv("KUBELESS_CONTROLLER_IMAGE")
		if img != "" {
			return img
		}
		return controllerImage
	default:
		return ""
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
