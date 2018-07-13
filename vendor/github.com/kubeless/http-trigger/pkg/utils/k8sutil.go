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
	"path/filepath"

	httptriggerapi "github.com/kubeless/http-trigger/pkg/apis/kubeless/v1beta1"
	kubelessApi "github.com/kubeless/http-trigger/pkg/apis/kubeless/v1beta1"
	"k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sirupsen/logrus"

	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/types"

	// Auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/kubeless/http-trigger/pkg/client/clientset/versioned"
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

// GetTriggerClientInCluster returns function clientset to the request from inside of cluster
func GetTriggerClientInCluster() (versioned.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	httpTriggerClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return httpTriggerClient, nil
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

// CreateHTTPTriggerCustomResource will create a HTTP trigger custom resource object
func CreateHTTPTriggerCustomResource(kubelessClient versioned.Interface, httpTrigger *httptriggerapi.HTTPTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().HTTPTriggers(httpTrigger.Namespace).Create(httpTrigger)
	if err != nil {
		return err
	}
	return nil
}

// UpdateHTTPTriggerCustomResource applies changes to the HTTP trigger custom resource object
func UpdateHTTPTriggerCustomResource(kubelessClient versioned.Interface, httpTrigger *httptriggerapi.HTTPTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().HTTPTriggers(httpTrigger.Namespace).Update(httpTrigger)
	return err
}

// PatchHTTPTriggerCustomResource applies changes to the function custom object
func PatchHTTPTriggerCustomResource(kubelessClient versioned.Interface, httpTrigger *httptriggerapi.HTTPTrigger) error {
	data, err := json.Marshal(httpTrigger)
	if err != nil {
		return err
	}
	_, err = kubelessClient.KubelessV1beta1().HTTPTriggers(httpTrigger.Namespace).Patch(httpTrigger.Name, types.MergePatchType, data)
	return err
}

// DeleteHTTPTriggerCustomResource will delete  HTTP trigger custom resource object
func DeleteHTTPTriggerCustomResource(kubelessClient versioned.Interface, httpTriggerName, ns string) error {
	err := kubelessClient.KubelessV1beta1().HTTPTriggers(ns).Delete(httpTriggerName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetHTTPTriggerCustomResource will get  HTTP trigger custom resource object
func GetHTTPTriggerCustomResource(kubelessClient versioned.Interface, httpTriggerName, ns string) (*httptriggerapi.HTTPTrigger, error) {
	kafkaCRD, err := kubelessClient.KubelessV1beta1().HTTPTriggers(ns).Get(httpTriggerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return kafkaCRD, nil
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

// CreateIngress creates ingress rule for a specific function
func CreateIngress(client kubernetes.Interface, httpTriggerObj *kubelessApi.HTTPTrigger, or []metav1.OwnerReference) error {
	funcSvc, err := client.CoreV1().Services(httpTriggerObj.ObjectMeta.Namespace).Get(httpTriggerObj.Spec.FunctionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to find the function internal service: %v", funcSvc)
	}

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            httpTriggerObj.Name,
			Namespace:       httpTriggerObj.Namespace,
			OwnerReferences: or,
			Labels:          httpTriggerObj.ObjectMeta.Labels,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: httpTriggerObj.Spec.HostName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/" + httpTriggerObj.Spec.Path,
									Backend: v1beta1.IngressBackend{
										ServiceName: funcSvc.Name,
										ServicePort: funcSvc.Spec.Ports[0].TargetPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ingressAnnotations := make(map[string]string)

	// If exposed URL in the backend service differs from the specified path in the Ingress rule.
	// Without a rewrite any request will return 404. Set the annotation ingress.kubernetes.io/rewrite-target
	// to the path expected by the service
	ingressAnnotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/"

	if len(httpTriggerObj.Spec.BasicAuthSecret) > 0 {
		switch gateway := httpTriggerObj.Spec.Gateway; gateway {
		case "nginx":
			ingressAnnotations["kubernetes.io/ingress.class"] = "nginx"
			ingressAnnotations["nginx.ingress.kubernetes.io/auth-secret"] = httpTriggerObj.Spec.BasicAuthSecret
			ingressAnnotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
			break
		case "traefik":
			ingressAnnotations["kubernetes.io/ingress.class"] = "traefik"
			ingressAnnotations["ingress.kubernetes.io/auth-secret"] = httpTriggerObj.Spec.BasicAuthSecret
			ingressAnnotations["ingress.kubernetes.io/auth-type"] = "basic"
			break
		case "kong":
			return fmt.Errorf("Setting basic authentication with Kong is not yet supported")
		}
	}

	if len(httpTriggerObj.Spec.TLSSecret) > 0 && httpTriggerObj.Spec.TLSAcme {
		return fmt.Errorf("Can not create ingress object from HTTP trigger spec with both TLSSecret and IngressTLS specified")
	}

	//  secure an Ingress by specified secret that contains a TLS private key and certificate
	if len(httpTriggerObj.Spec.TLSSecret) > 0 {
		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				SecretName: httpTriggerObj.Spec.TLSSecret,
				Hosts:      []string{httpTriggerObj.Spec.HostName},
			},
		}
	}

	// add annotations and TLS configuration for kube-lego
	if httpTriggerObj.Spec.TLSAcme {
		ingressAnnotations["kubernetes.io/tls-acme"] = "true"
		ingressAnnotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "true"
		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{httpTriggerObj.Spec.HostName},
				SecretName: httpTriggerObj.Name + "-tls",
			},
		}
	}
	ingress.ObjectMeta.Annotations = ingressAnnotations
	_, err = client.ExtensionsV1beta1().Ingresses(httpTriggerObj.Namespace).Create(ingress)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		var newIngress *v1beta1.Ingress
		newIngress, err = client.ExtensionsV1beta1().Ingresses(httpTriggerObj.Namespace).Get(ingress.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if len(ingress.ObjectMeta.Labels) > 0 {
			newIngress.ObjectMeta.Labels = ingress.ObjectMeta.Labels
		}
		newIngress.ObjectMeta.OwnerReferences = or
		newIngress.Spec = ingress.Spec
		_, err = client.ExtensionsV1beta1().Ingresses(httpTriggerObj.Namespace).Update(newIngress)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The configmap may already exist and there is nothing to update
			return nil
		}
	}
	return err
}
