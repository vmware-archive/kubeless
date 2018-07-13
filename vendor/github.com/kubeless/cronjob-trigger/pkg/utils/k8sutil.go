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
	"os"
	"path/filepath"

	//kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	cronjobtriggerApi "github.com/kubeless/cronjob-trigger/pkg/apis/kubeless/v1beta1"
	"github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	clientsetAPIExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/kubeless/cronjob-trigger/pkg/client/clientset/versioned"
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
	kafkaClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kafkaClient, nil
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
	config, err := rest.InClusterConfig()
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

// CreateCronJobCustomResource will create a custom function object
func CreateCronJobCustomResource(kubelessClient versioned.Interface, cronJob *cronjobtriggerApi.CronJobTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().CronJobTriggers(cronJob.Namespace).Create(cronJob)
	if err != nil {
		return err
	}
	return nil
}

// UpdateCronJobCustomResource applies changes to the function custom object
func UpdateCronJobCustomResource(kubelessClient versioned.Interface, cronJob *cronjobtriggerApi.CronJobTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().CronJobTriggers(cronJob.Namespace).Update(cronJob)
	return err
}

// DeleteCronJobCustomResource will delete custom function object
func DeleteCronJobCustomResource(kubelessClient versioned.Interface, cronJobName, ns string) error {
	err := kubelessClient.KubelessV1beta1().CronJobTriggers(ns).Delete(cronJobName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetCronJobCustomResource will get CronJobTrigger custom resource object
func GetCronJobCustomResource(kubelessClient versioned.Interface, cronJobName, ns string) (*cronjobtriggerApi.CronJobTrigger, error) {
	cronJobCRD, err := kubelessClient.KubelessV1beta1().CronJobTriggers(ns).Get(cronJobName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return cronJobCRD, nil
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
