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
	apimachineryHelpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	cronJobTriggerMaxRetries = 5
	cronJobObjKind           = "Trigger"
	cronJobObjAPI            = "kubeless.io"
	cronJobTriggerFinalizer  = "kubeless.io/cronjobtrigger"
)

// CronJobTriggerController object
type CronJobTriggerController struct {
	logger           *logrus.Entry
	clientset        kubernetes.Interface
	kubelessclient   versioned.Interface
	queue            workqueue.RateLimitingInterface
	cronJobInformer  kubelessInformers.CronJobTriggerInformer
	functionInformer kubelessInformers.FunctionInformer
}

// CronJobTriggerConfig contains k8s client of a controller
type CronJobTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewCronJobTriggerController initializes a controller object
func NewCronJobTriggerController(cfg CronJobTriggerConfig) *CronJobTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	sharedInformers := externalversions.NewSharedInformerFactory(cfg.TriggerClient, 0)
	cronJobInformer := sharedInformers.Kubeless().V1beta1().CronJobTriggers()
	functionInformer := sharedInformers.Kubeless().V1beta1().Functions()

	cronJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
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

	controller := CronJobTriggerController{
		logger:           logrus.WithField("controller", "cronjob-trigger-controller"),
		clientset:        cfg.KubeCli,
		kubelessclient:   cfg.TriggerClient,
		cronJobInformer:  cronJobInformer,
		functionInformer: functionInformer,
		queue:            queue,
	}

	functionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.functionAddedDeletedUpdated(obj, false)
		},
		DeleteFunc: func(obj interface{}) {
			controller.functionAddedDeletedUpdated(obj, true)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.functionAddedDeletedUpdated(new, false)
		},
	})

	return &controller
}

// Run starts the Trigger controller
func (c *CronJobTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Cron Job Trigger controller")

	go c.cronJobInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Cron Job Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *CronJobTriggerController) HasSynced() bool {
	return c.cronJobInformer.Informer().HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *CronJobTriggerController) LastSyncResourceVersion() string {
	return c.cronJobInformer.Informer().LastSyncResourceVersion()
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

	obj, exists, err := c.cronJobInformer.Informer().GetIndexer().GetByKey(key)
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

	restIface := c.clientset.BatchV2alpha1().RESTClient()
	groupVersion, err := c.getResouceGroupVersion("cronjobs")
	if err != nil {
		return err
	}

	funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&cronJobtriggerObj.Spec.FunctionSelector)
	if err != nil {
		c.logger.Errorf("Failed to convert LabelSelector to Selector due to %s: ", err)
	}
	functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
	if err != nil {
		c.logger.Errorf("Failed to list function by Selector due to %s: ", err)
	}

	if len(functions) == 0 {
		c.logger.Infof("No matching functions with selector %v found in namespace %s", funcSelector, ns)
	}

	for _, function := range functions {
		if err != nil {
			c.logger.Errorf("Unable to find the function %s in the namespace %s. Received %s: ", function.ObjectMeta.Name, ns, err)
			return err
		}
		if needToAddFinalizer(function) {
			funcObjClone := function.DeepCopy()
			funcObjClone.ObjectMeta.Finalizers = append(funcObjClone.ObjectMeta.Finalizers, cronJobTriggerFinalizer)
			err = utils.UpdateFunctionCustomResource(c.kubelessclient, funcObjClone)
			if err != nil {
				c.logger.Errorf("Error adding CronJob trigger controller as finalizer to Function: %s CRD object due to: %s: ", function.ObjectMeta.Name, err)
			}
		}
		err = utils.EnsureCronJob(restIface, function, cronJobtriggerObj, or, groupVersion)
		if err != nil {
			return err
		}
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

func (c *CronJobTriggerController) functionAddedDeletedUpdated(obj interface{}, deleted bool) {
	functionObj, ok := obj.(*kubelessApi.Function)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			c.logger.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		functionObj, ok = tombstone.Obj.(*kubelessApi.Function)
		if !ok {
			c.logger.Errorf("Tombstone contained object that is not a Pod %#v", obj)
			return
		}
	}
	if deleted || functionObj.DeletionTimestamp == nil {
		return
	}

	//check if func is scheduled or not
	cronJobName := fmt.Sprintf("trigger-%s", functionObj.ObjectMeta.Name)
	_, err := c.clientset.BatchV2alpha1().CronJobs(functionObj.ObjectMeta.Namespace).Get(cronJobName, metav1.GetOptions{})
	if err == nil {
		err = c.clientset.BatchV2alpha1().CronJobs(functionObj.ObjectMeta.Namespace).Delete(cronJobName, &metav1.DeleteOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			c.logger.Errorf("Failed to delete cronjob %s created for the function %s in namespace %s", cronJobName, functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace)
		}
	}

	funcObjClone := functionObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range functionObj.ObjectMeta.Finalizers {
		if item == cronJobTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	funcObjClone.ObjectMeta.Finalizers = newSlice
	err = utils.UpdateFunctionCustomResource(c.kubelessclient, funcObjClone)
	if err != nil {
		c.logger.Errorf("Error removing CronJob trigger controller as finalizer to Function: %s CRD object due to: %s: ", functionObj.ObjectMeta.Name, err)
		return
	}
	c.logger.Infof("Successfully removed CronJob trigger controller as finalizer to Function: %s CRD object", functionObj.ObjectMeta.Name)
}

func needToAddFinalizer(funcObj *kubelessApi.Function) bool {
	currentFinalizers := funcObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == cronJobTriggerFinalizer {
			return false
		}
	}
	return funcObj.ObjectMeta.DeletionTimestamp == nil
}
