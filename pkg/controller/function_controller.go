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

package controller

import (
	"fmt"
	"os"
	"time"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kv1beta1 "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	maxRetries = 5
	funcKind   = "Function"
	funcAPI    = "kubeless.io"
)

// FunctionController object
type FunctionController struct {
	logger         *logrus.Entry
	clientset      kubernetes.Interface
	kubelessclient versioned.Interface
	smclient       *monitoringv1alpha1.MonitoringV1alpha1Client
	Functions      map[string]*kubelessApi.Function
	queue          workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
	config         *corev1.ConfigMap
	langRuntime    *langruntime.Langruntimes
}

// Config contains k8s client of a controller
type Config struct {
	KubeCli        kubernetes.Interface
	FunctionClient versioned.Interface
}

// NewFunctionController initializes a controller object
func NewFunctionController(cfg Config, smclient *monitoringv1alpha1.MonitoringV1alpha1Client) *FunctionController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := kv1beta1.NewFunctionInformer(cfg.FunctionClient, corev1.NamespaceAll, 0, cache.Indexers{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	kubelessConfig := os.Getenv("KUBELESS_CONFIG")
	if len(controllerNamespace) == 0 {
		controllerNamespace = "kubeless"
	}
	if len(kubelessConfig) == 0 {
		kubelessConfig = "kubeless-config"
	}
	config, err := cfg.KubeCli.CoreV1().ConfigMaps(controllerNamespace).Get(kubelessConfig, metav1.GetOptions{})
	if err != nil {
		logrus.Fatalf("Unable to read the configmap: %s", err)
	}

	var lr = langruntime.New(config)
	lr.ReadConfigMap()

	return &FunctionController{
		logger:         logrus.WithField("pkg", "controller"),
		clientset:      cfg.KubeCli,
		smclient:       smclient,
		kubelessclient: cfg.FunctionClient,
		informer:       informer,
		queue:          queue,
		config:         config,
		langRuntime:    lr,
	}
}

// Run starts the kubeless controller
func (c *FunctionController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting kubeless Functions controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Kubeless Functions controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *FunctionController) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *FunctionController) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *FunctionController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *FunctionController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *FunctionController) processItem(key string) error {
	c.logger.Infof("Processing change to Function %s", key)

	_, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		c.logger.Infof("Deleted Function %s", key)
		return nil
	}

	funcObj := obj.(*kubelessApi.Function)
	err = utils.UpdateFunctionDeployments(funcObj)
	if err != nil {
		return fmt.Errorf("Error updating deployments for the function with key %s due to %v", key, err)
	}
	c.logger.Infof("Updated Function %s", key)
	return nil
}