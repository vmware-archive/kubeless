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

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
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

// CronJobTriggerConfig contains config for CronJobTriggerController
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
				newObj := new.(*kubelessApi.CronJobTrigger)
				oldObj := old.(*kubelessApi.CronJobTrigger)
				if cronJobTriggerObjChanged(oldObj, newObj) {
					queue.Add(key)
				}
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

	if !c.WaitForCacheSync(stopCh) {
		return
	}

	c.logger.Info("Cron Job Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// WaitForCacheSync is required for caches to be synced
func (c *CronJobTriggerController) WaitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.cronJobInformer.Informer().HasSynced, c.functionInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for Cronjob triggers controller to sync;"))
		return false
	}
	c.logger.Info("Cronjob Trigger controller caches are synced and ready")
	return true
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

	err := c.syncCronJobTrigger(key.(string))
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

func (c *CronJobTriggerController) syncCronJobTrigger(key string) error {
	c.logger.Infof("Processing update to CronJob Trigger: %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.cronJobInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when CronJob trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("Cronjob Trigger %s not found, ignoring", key)
		return nil
	}

	cronJobtriggerObj := obj.(*kubelessApi.CronJobTrigger)

	// CronJob trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if cronJobtriggerObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !c.cronJobTriggerObjHasFinalizer(cronJobtriggerObj) {
			return nil
		}

		// CronJob Trigger object should be deleted, so remove associated cronjob and remove the finalizer
		_, err := c.clientset.BatchV2alpha1().CronJobs(ns).Get(fmt.Sprintf("trigger-%s", name), metav1.GetOptions{})
		if err == nil {
			err = c.clientset.BatchV2alpha1().CronJobs(ns).Delete(fmt.Sprintf("trigger-%s", name), &metav1.DeleteOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				c.logger.Errorf("Failed to remove CronJob created for CronJobTrigger Obj: %s due to: %v: ", key, err)
				return err
			}
		}

		// remove finalizer from the cronjob trigger object, so that we dont have to process any further and object can be deleted
		err = c.cronJobTriggerObjRemoveFinalizer(cronJobtriggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove CronJob trigger controller as finalizer to CronJob Obj: %s due to: %v: ", key, err)
			return err
		}

		c.logger.Infof("Cronjob trigger object %s has been successfully processed and marked for deletion", key)
		return nil
	}

	// If CronJob trigger API in not marked with self as finalizer, then add the finalizer
	if !c.cronJobTriggerObjHasFinalizer(cronJobtriggerObj) {
		err = c.cronJobTriggerObjAddFinalizer(cronJobtriggerObj)
		if err != nil {
			c.logger.Errorf("Error adding CronJob trigger controller as finalizer to  CronJobTrigger Obj: %s CRD object due to: %v: ", key, err)
			return err
		}
	}

	or, err := utils.GetCronJobTriggerOwnerReference(cronJobtriggerObj)
	if err != nil {
		return err
	}

	restIface := c.clientset.BatchV2alpha1().RESTClient()
	groupVersion, err := c.getResouceGroupVersion("cronjobs")
	if err != nil {
		return err
	}

	functionObj, err := c.functionInformer.Lister().Functions(ns).Get(cronJobtriggerObj.Spec.FunctionName)
	if err != nil {
		c.logger.Errorf("Unable to find the function %s in the namespace %s. Received %s: ", cronJobtriggerObj.Spec.FunctionName, ns, err)
		return err
	}
	err = utils.EnsureCronJob(restIface, functionObj, cronJobtriggerObj, or, groupVersion)
	if err != nil {
		return err
	}

	c.logger.Infof("Processed update to CronJobrigger: %s", key)
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

	c.logger.Infof("Processing update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	if deleted {
		//check if func is scheduled or not
		cronJobName := fmt.Sprintf("trigger-%s", functionObj.ObjectMeta.Name)
		_, err := c.clientset.BatchV2alpha1().CronJobs(functionObj.ObjectMeta.Namespace).Get(cronJobName, metav1.GetOptions{})
		if err == nil {
			err = c.clientset.BatchV2alpha1().CronJobs(functionObj.ObjectMeta.Namespace).Delete(cronJobName, &metav1.DeleteOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				c.logger.Errorf("Failed to delete cronjob %s created for the function %s in namespace %s", cronJobName, functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace)
			}
		}
	}
}

func (c *CronJobTriggerController) cronJobTriggerObjHasFinalizer(triggerObj *kubelessApi.CronJobTrigger) bool {
	currentFinalizers := triggerObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == cronJobTriggerFinalizer {
			return true
		}
	}
	return false
}

func (c *CronJobTriggerController) cronJobTriggerObjAddFinalizer(triggercObj *kubelessApi.CronJobTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	triggercObjClone.ObjectMeta.Finalizers = append(triggercObjClone.ObjectMeta.Finalizers, cronJobTriggerFinalizer)
	return utils.UpdateCronJobCustomResource(c.kubelessclient, triggercObjClone)
}

func (c *CronJobTriggerController) cronJobTriggerObjRemoveFinalizer(triggercObj *kubelessApi.CronJobTrigger) error {
	triggerObjClone := triggercObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range triggerObjClone.ObjectMeta.Finalizers {
		if item == cronJobTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	triggerObjClone.ObjectMeta.Finalizers = newSlice
	err := utils.UpdateCronJobCustomResource(c.kubelessclient, triggerObjClone)
	if err != nil {
		return err
	}
	return nil
}

func cronJobTriggerObjChanged(oldObj, newObj *kubelessApi.CronJobTrigger) bool {
	// If the CronJob trigger object's deletion timestamp is set, then process
	if oldObj.DeletionTimestamp != newObj.DeletionTimestamp {
		return true
	}
	// If the new and old CronJob trigger object's resource version is same
	if oldObj.ResourceVersion != newObj.ResourceVersion {
		return true
	}
	newSpec := &newObj.Spec
	oldSpec := &oldObj.Spec

	if newSpec.Schedule != oldSpec.Schedule {
		return true
	}

	return false
}
