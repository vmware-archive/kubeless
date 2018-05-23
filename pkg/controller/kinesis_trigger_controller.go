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
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/event-consumers/kinesis"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	kenisisTriggerMaxRetries = 5
	kenisisTriggerFinalizer  = "kubeless.io/kenisistrigger"
)

// KinesisTriggerController object
type KinesisTriggerController struct {
	logger           *logrus.Entry
	kubelessclient   versioned.Interface
	kubernetesClient kubernetes.Interface
	queue            workqueue.RateLimitingInterface
	kinesisInformer  kubelessInformers.KinesisTriggerInformer
	functionInformer kubelessInformers.FunctionInformer
}

// KinesisTriggerConfig contains config for KinesisTriggerController
type KinesisTriggerConfig struct {
	TriggerClient versioned.Interface
}

// NewKinesisTriggerController returns a new *KinesisTriggerController.
func NewKinesisTriggerController(cfg KinesisTriggerConfig) *KinesisTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	sharedInformers := externalversions.NewSharedInformerFactory(cfg.TriggerClient, 0)
	kinesisInformer := sharedInformers.Kubeless().V1beta1().KinesisTriggers()
	functionInformer := sharedInformers.Kubeless().V1beta1().Functions()

	kinesisInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newKinesisTriggerObj := new.(*kubelessApi.KinesisTrigger)
				oldKinesisTriggerObj := old.(*kubelessApi.KinesisTrigger)
				if kinesisTriggerObjChanged(oldKinesisTriggerObj, newKinesisTriggerObj) {
					queue.Add(key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := KinesisTriggerController{
		logger:           logrus.WithField("controller", "kinesis-trigger-controller"),
		kubelessclient:   cfg.TriggerClient,
		kubernetesClient: utils.GetClient(),
		kinesisInformer:  kinesisInformer,
		functionInformer: functionInformer,
		queue:            queue,
	}

	functionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.FunctionAddedDeletedUpdated(obj, false)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.FunctionAddedDeletedUpdated(new, false)
		},
		DeleteFunc: func(obj interface{}) {
			controller.FunctionAddedDeletedUpdated(obj, true)
		},
	})

	return &controller
}

// Run starts the Kinesis trigger controller
func (c *KinesisTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting AWS Kinesis Trigger controller.")
	defer c.logger.Info("Shutting down AWS Kinesis Trigger controller.")

	go c.kinesisInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !c.waitForCacheSync(stopCh) {
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *KinesisTriggerController) waitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.kinesisInformer.Informer().HasSynced, c.functionInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for AWS Kinesis triggers controller to sync;"))
		return false
	}
	c.logger.Info(" AWS Kinesis Trigger controller caches are synced and ready")
	return true
}

func (c *KinesisTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *KinesisTriggerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncKinesisTrigger(key.(string))
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

func (c *KinesisTriggerController) syncKinesisTrigger(key string) error {
	c.logger.Infof("Processing update to AWS Kinesis Trigger: %s", key)
	ns, triggerObjName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.kinesisInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when Kinesis trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("Kinesis Trigger %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.KinesisTrigger)

	if triggerObj.Spec.Stream == "" {
		return errors.New("Kinesis Trigger Stream can't be empty. Please check the Kinesis trigger object %s" + key)
	}
	if triggerObj.Spec.Region == "" {
		return errors.New("Kinesis Trigger region can't be empty. Please check the Kinesis trigger object %s" + key)
	}
	if triggerObj.Spec.ShardID == "" {
		return errors.New("Kinesis Trigger shard-id can't be empty. Please check the Kinesis trigger object %s" + key)
	}
	if triggerObj.Spec.Secret == "" {
		return errors.New("Kinesis Trigger secret can't be empty. Please check the Kinesis trigger object %s" + key)
	}

	// Kinesis trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if triggerObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !c.kinesisTriggerHasFinalizer(triggerObj) {
			return nil
		}
		err = kinesis.DeleteKinesisConsumer(triggerObj, triggerObj.Spec.FunctionName, triggerObj.Namespace)
		if err != nil {
			c.logger.Errorf("Error deleting Kinesis stream processor. Error: %v", err)
		}
		// remove finalizer from the Kinesis trigger object, so that we dont have to process any further and object can be deleted
		err = c.kinesisTriggerObjRemoveFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove Kinesis trigger controller as finalizer to Kinesis Trigger Obj: %s CRD object due to: %v: ", triggerObjName, err)
			return err
		}
		c.logger.Infof("Kinesis trigger %s has been successfully processed and marked for deletion", key)
		return nil
	}

	// If Kinesis trigger API in not marked with self as finalizer, then add the finalizer
	if !c.kinesisTriggerHasFinalizer(triggerObj) {
		err = c.kinesisTriggerObjAddFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Error adding Kinesis trigger controller as finalizer to Kinesis Trigger Obj: %s CRD object due to: %v: ", triggerObjName, err)
			return err
		}
	}

	_, err = utils.GetFunctionCustomResource(c.kubelessclient, triggerObj.Spec.FunctionName, ns)
	if err != nil {
		return fmt.Errorf("Unable to find Function %s in namespace %s. Error %v", triggerObj.Spec.FunctionName, ns, err)
	}
	_, err = c.kubernetesClient.Core().Secrets(ns).Get(triggerObj.Spec.Secret, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to find secret %s in namespace %s. Error %v", triggerObj.Spec.Secret, ns, err)
	}

	err = kinesis.CreateKinesisStreamConsumer(triggerObj, triggerObj.Spec.FunctionName, triggerObj.Namespace, c.kubernetesClient)
	if err != nil {
		c.logger.Errorf("Error creating Kinesis stream processor. Error: %v", err)
	}

	c.logger.Infof("Processed change to Kinesis trigger: %s", key)
	return nil
}

// FunctionAddedDeletedUpdated process the updates to Function objects
func (c *KinesisTriggerController) FunctionAddedDeletedUpdated(obj interface{}, deleted bool) error {
	functionObj, ok := obj.(*kubelessApi.Function)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			err := fmt.Errorf("Couldn't get object from tombstone %#v", obj)
			c.logger.Errorf(err.Error())
			return err
		}
		functionObj, ok = tombstone.Obj.(*kubelessApi.Function)
		if !ok {
			err := fmt.Errorf("Tombstone contained object that is not a Function object %#v", obj)
			c.logger.Errorf(err.Error())
			return err
		}
	}
	c.logger.Infof("Processing update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	if deleted {
		c.logger.Infof("Function %s deleted. Removing associated Kinesis trigger", functionObj.Name)
		kinesisTriggersList, err := c.kubelessclient.KubelessV1beta1().KinesisTriggers(functionObj.Namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, kinesisTrigger := range kinesisTriggersList.Items {
			if kinesisTrigger.Spec.FunctionName == functionObj.Name {
				err = c.kubelessclient.KubelessV1beta1().KinesisTriggers(functionObj.Namespace).Delete(kinesisTrigger.Name, &metav1.DeleteOptions{})
				if err != nil && !k8sErrors.IsNotFound(err) {
					c.logger.Errorf("Failed to delete Kinesis trigger created for the function %s in namespace %s, Error: %s", functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace, err)
					return err
				}
			}
		}
	}
	c.logger.Infof("Successfully processed update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	return nil
}

func (c *KinesisTriggerController) kinesisTriggerObjNeedFinalizer(triggercObj *kubelessApi.KinesisTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == kenisisTriggerFinalizer {
			return false
		}
	}
	return triggercObj.ObjectMeta.DeletionTimestamp == nil
}

func (c *KinesisTriggerController) kinesisTriggerHasFinalizer(triggercObj *kubelessApi.KinesisTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == kenisisTriggerFinalizer {
			return true
		}
	}
	return false
}

func (c *KinesisTriggerController) kinesisTriggerObjAddFinalizer(triggercObj *kubelessApi.KinesisTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	triggercObjClone.ObjectMeta.Finalizers = append(triggercObjClone.ObjectMeta.Finalizers, kenisisTriggerFinalizer)
	return utils.UpdateKinesisTriggerCustomResource(c.kubelessclient, triggercObjClone)
}

func (c *KinesisTriggerController) kinesisTriggerObjRemoveFinalizer(triggercObj *kubelessApi.KinesisTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range triggercObjClone.ObjectMeta.Finalizers {
		if item == kenisisTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	triggercObjClone.ObjectMeta.Finalizers = newSlice
	err := utils.UpdateKinesisTriggerCustomResource(c.kubelessclient, triggercObjClone)
	if err != nil {
		return err
	}
	return nil
}

func kinesisTriggerObjChanged(oldKinesisTriggerObj, newKenisisTriggerObj *kubelessApi.KinesisTrigger) bool {
	// If the Kinesis trigger object's deletion timestamp is set, then process the update
	if oldKinesisTriggerObj.DeletionTimestamp != newKenisisTriggerObj.DeletionTimestamp {
		return true
	}
	// If the new and old Kinesis trigger object's resource version is same, then skip processing the update
	if oldKinesisTriggerObj.ResourceVersion == newKenisisTriggerObj.ResourceVersion {
		return false
	}
	if !apiequality.Semantic.DeepEqual(oldKinesisTriggerObj.Spec, newKenisisTriggerObj.Spec) {
		return true
	}
	return false
}
