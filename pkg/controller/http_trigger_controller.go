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
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kv1beta1 "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
)

const (
	httpTriggerMaxRetries = 5
	httpTriggerObjKind    = "HttpTrigger"
	httpTriggerObjAPI     = "kubeless.io"
)

// HTTPTriggerController object
type HTTPTriggerController struct {
	logger         *logrus.Entry
	clientset      kubernetes.Interface
	kubelessclient versioned.Interface
	queue          workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
}

// HTTPTriggerConfig contains k8s client of a controller
type HTTPTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewHTTPTriggerController initializes a controller object
func NewHTTPTriggerController(cfg HTTPTriggerConfig) *HTTPTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := kv1beta1.NewHTTPTriggerInformer(cfg.TriggerClient, corev1.NamespaceAll, 0, cache.Indexers{})

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

	return &HTTPTriggerController{
		logger:         logrus.WithField("controller", "http-trigger-controller"),
		clientset:      cfg.KubeCli,
		kubelessclient: cfg.TriggerClient,
		informer:       informer,
		queue:          queue,
	}
}

// Run starts the Trigger controller
func (c *HTTPTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting HTTP Trigger controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("HTTP Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *HTTPTriggerController) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *HTTPTriggerController) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *HTTPTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *HTTPTriggerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < triggerMaxRetries {
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

func (c *HTTPTriggerController) processItem(key string) error {
	c.logger.Infof("Processing change to Trigger %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	c.logger.Infof("Processing update to Trigger: %s Namespace: %s", name, ns)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		if err != nil {
			c.logger.Errorf("Can't delete function: %v", err)
			return err
		}
		c.logger.Infof("Deleted Function %s", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.HTTPTrigger)
	c.logger.Infof("Processed change to Trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
	return nil
}