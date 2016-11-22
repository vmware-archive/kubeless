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
	"net/http"
	"strings"
	"time"

	"github.com/skippbox/kubeless/pkg/spec"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/kubectl"
)

const TIMEOUT = 300

func GetClient() (*client.Client, string, error) {
	factory := cmdutil.NewFactory(nil)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		return nil, "", err
	}
	namespace, _, err := factory.DefaultNamespace()
	if err != nil {
		return nil, "", err
	}
	client := client.NewOrDie(clientConfig)
	return client, namespace, nil
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
	return httpClient.Get(fmt.Sprintf("https://%s:8443/apis/k8s.io/v1/namespaces/%s/lambdas",
		host, ns))
}

func WatchResources(host, ns string, httpClient *http.Client, resourceVersion string) (*http.Response, error) {
	return httpClient.Get(fmt.Sprintf("https://%s:8443/apis/k8s.io/v1/namespaces/%s/lambdas?watch=true&resourceVersion=%s",
		host, ns, resourceVersion))
}

func CreateK8sResources(ns, name string, spec *spec.FunctionSpec, client *client.Client) error {
	str := strings.Split(spec.Handler, ".")
	if len(str) != 2 {
		return errors.New("Failed: incorrect handler format. It should be module_name.handler_name")
	}
	funcHandler := str[1]
	modName := str[0]
	fileName := modName
	switch {
	case strings.Contains(spec.Runtime, "python"):
		fileName = modName + ".py"
	case strings.Contains(spec.Runtime, "go"):
		fileName = modName + ".go"
	case strings.Contains(spec.Runtime, "nodejs"):
		fileName = modName + ".js"
	}

	//add configmap
	labels := map[string]string{
		"lambda": name,
	}
	data := map[string]string{
		"handler": spec.Handler,
		fileName:  spec.Lambda,
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
					Containers: []api.Container{
						{
							Name:  name,
							Image: "runseb/kubeless:0.0.5",
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