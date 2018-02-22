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
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	cronJobTriggerMaxRetries = 5
	cronJobObjKind           = "Trigger"
	cronJobObjAPI            = "kubeless.io"
)

// CronJobTriggerController object
type CronJobTriggerController struct {
	logger         *logrus.Entry
	clientset      kubernetes.Interface
	kubelessclient versioned.Interface
	queue          workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
}

// CronJobTriggerConfig contains k8s client of a controller
type CronJobTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewCronJobTriggerController initializes a controller object
func NewCronJobTriggerController(cfg CronJobTriggerConfig) *CronJobTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := kv1beta1.NewCronJobTriggerInformer(cfg.TriggerClient, corev1.NamespaceAll, 0, cache.Indexers{})

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

	return &CronJobTriggerController{
		logger:         logrus.WithField("controller", "cronjob-trigger-controller"),
		clientset:      cfg.KubeCli,
		kubelessclient: cfg.TriggerClient,
		informer:       informer,
		queue:          queue,
	}
}

// Run starts the Trigger controller
func (c *CronJobTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Cron Job Trigger controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Cron Job Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *CronJobTriggerController) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *CronJobTriggerController) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *CronJobTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *CronJobTriggerController) processNextItem() bool {
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

func (c *CronJobTriggerController) processItem(key string) error {
	c.logger.Infof("Processing change to Trigger %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	c.logger.Infof("Processing update to Cron Job Trigger: %s Namespace: %s", name, ns)

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

	cronJobtriggerObj := obj.(*kubelessApi.CronJobTrigger)

	or, err := utils.GetCronJobTriggerOwnerReference(cronJobtriggerObj)
	if err != nil {
		return err
	}

	client, err := utils.GetFunctionClientInCluster()
	if err != nil {
		c.logger.Errorf("Unable to get the in-cluster client due to %s: ", err)
		return err
	}

	funcObj, err := utils.GetFunctionCustomResource(client, cronJobtriggerObj.Spec.FunctionName, ns)
	if err != nil {
		c.logger.Errorf("Unable to find the function %s in the namespace %s. Received %s: ", cronJobtriggerObj.Spec.FunctionName, ns, err)
		return err
	}

	restIface := c.clientset.BatchV2alpha1().RESTClient()
	groupVersion, err := c.getResouceGroupVersion("cronjobs")
	if err != nil {
		return err
	}
	err = utils.EnsureCronJob(restIface, funcObj, cronJobtriggerObj, or, groupVersion)
	if err != nil {
		return err
	}
	c.logger.Infof("Processed change cron job to Trigger: %s Namespace: %s", cronJobtriggerObj.ObjectMeta.Name, ns)
	return nil
}

func (c *CronJobTriggerController) getResouceGroupVersion(target string) (string, error) {
	resources, err := c.clientset.Discovery().ServerResources()
	if err != nil {
		return "", err
	}
	groupVersion := ""
	for _, resource := range resources {
		for _, apiResource := range resource.APIResources {
			if apiResource.Name == target {
				groupVersion = resource.GroupVersion
				break
			}
		}
	}
	if groupVersion == "" {
		return "", fmt.Errorf("Resource %s not found in any group", target)
	}
	return groupVersion, nil
}
