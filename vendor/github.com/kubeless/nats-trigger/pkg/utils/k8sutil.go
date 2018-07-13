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
	"os"
	"path/filepath"

	natsApi "github.com/kubeless/nats-trigger/pkg/apis/kubeless/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubeless/nats-trigger/pkg/client/clientset/versioned"
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

// CreateNatsTriggerCustomResource will create a custom function object
func CreateNatsTriggerCustomResource(kubelessClient versioned.Interface, natsTrigger *natsApi.NATSTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().NATSTriggers(natsTrigger.Namespace).Create(natsTrigger)
	if err != nil {
		return err
	}
	return nil
}

// UpdateNatsTriggerCustomResource applies changes to the function custom object
func UpdateNatsTriggerCustomResource(kubelessClient versioned.Interface, natsTrigger *natsApi.NATSTrigger) error {
	_, err := kubelessClient.KubelessV1beta1().NATSTriggers(natsTrigger.Namespace).Update(natsTrigger)
	return err
}

// DeleteNatsTriggerCustomResource will delete custom function object
func DeleteNatsTriggerCustomResource(kubelessClient versioned.Interface, natsTriggerName, ns string) error {
	err := kubelessClient.KubelessV1beta1().NATSTriggers(ns).Delete(natsTriggerName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetNatsTriggerCustomResource will get CronJobTrigger custom resource object
func GetNatsTriggerCustomResource(kubelessClient versioned.Interface, natsTriggerName, ns string) (*natsApi.NATSTrigger, error) {
	kafkaCRD, err := kubelessClient.KubelessV1beta1().NATSTriggers(ns).Get(natsTriggerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return kafkaCRD, nil
}
