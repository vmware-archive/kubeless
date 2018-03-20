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
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	apimachineryHelpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/event-consumers/kafka"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	triggerMaxRetries     = 5
	kafkaTriggerFinalizer = "kubeless.io/kafkatrigger"
)

// KafkaTriggerController object
type KafkaTriggerController struct {
	logger           *logrus.Entry
	kubelessclient   versioned.Interface
	kubernetesClient kubernetes.Interface
	queue            workqueue.RateLimitingInterface
	kafkaInformer    kubelessInformers.KafkaTriggerInformer
	functionInformer kubelessInformers.FunctionInformer
}

// KafkaTriggerConfig contains config for KafkaTriggerController
type KafkaTriggerConfig struct {
	TriggerClient versioned.Interface
}

// NewKafkaTriggerController returns a new *KafkaTriggerController.
func NewKafkaTriggerController(cfg KafkaTriggerConfig) *KafkaTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	sharedInformers := externalversions.NewSharedInformerFactory(cfg.TriggerClient, 0)
	kafkaInformer := sharedInformers.Kubeless().V1beta1().KafkaTriggers()
	functionInformer := sharedInformers.Kubeless().V1beta1().Functions()

	kafkaInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newKafkaTriggerObj := new.(*kubelessApi.KafkaTrigger)
				oldKafkaTriggerObj := old.(*kubelessApi.KafkaTrigger)
				if kafkaTriggerObjChanged(oldKafkaTriggerObj, newKafkaTriggerObj) {
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

	controller := KafkaTriggerController{
		logger:           logrus.WithField("controller", "kafka-trigger-controller"),
		kubelessclient:   cfg.TriggerClient,
		kubernetesClient: utils.GetClient(),
		kafkaInformer:    kafkaInformer,
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

// Run starts the Kafka trigger controller
func (c *KafkaTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Kafka Trigger controller")
	defer c.logger.Info("Shutting down Kafka Trigger controller")

	go c.kafkaInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !c.waitForCacheSync(stopCh) {
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *KafkaTriggerController) waitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.kafkaInformer.Informer().HasSynced, c.functionInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for Kafka triggers controller to sync;"))
		return false
	}
	c.logger.Info("Kafka Trigger controller caches are synced and ready")
	return true
}

func (c *KafkaTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *KafkaTriggerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncKafkaTrigger(key.(string))
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

func (c *KafkaTriggerController) syncKafkaTrigger(key string) error {
	c.logger.Infof("Processing update to Kafka Trigger: %s", key)
	ns, triggerObjName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.kafkaInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when Kafka trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("Kafka Trigger %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.KafkaTrigger)
	topic := triggerObj.Spec.Topic
	if topic == "" {
		return errors.New("Kafka Trigger Topic can't be empty. Please check the trigger object %s" + key)
	}

	// Kafka trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if triggerObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !c.kafkaTriggerHasFinalizer(triggerObj) {
			return nil
		}

		// Kafka Trigger object should be deleted, so process the associated functions and remove the finalizer
		funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
		if err != nil {
			c.logger.Errorf("Failed to convert LabelSelector %v in Kafka Trigger object %s to Selector due to %v: ", triggerObj.Spec.FunctionSelector, key, err)
			return err
		}
		functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
		if err != nil {
			c.logger.Errorf("Failed to list associated functions with Kafka trigger %s by selector due to %s: ", key, err)
			return err
		}
		if len(functions) == 0 {
			c.logger.Infof("No matching functions found for Kafka trigger %s so marking CRD object for deleteion", key)
		}

		for _, function := range functions {
			funcName := function.ObjectMeta.Name
			err = kafka.DeleteKafkaConsumer(triggerObjName, funcName, ns, topic)
			if err != nil {
				c.logger.Errorf("Failed to delete the Kafka consumer for the function %s associated with the Kafka trigger %s due to %v: ", funcName, key, err)
			}
		}

		// remove finalizer from the kafka trigger object, so that we dont have to process any further and object can be deleted
		err = c.kafkaTriggerObjRemoveFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove Kafka trigger controller as finalizer to Kafka Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
		c.logger.Infof("Kafka trigger %s has been successfully processed and marked for deletion", key)
		return nil
	}

	// If Kafka trigger API in not marked with self as finalizer, then add the finalizer
	if !c.kafkaTriggerHasFinalizer(triggerObj) {
		err = c.kafkaTriggerObjAddFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Error adding Kafka trigger controller as finalizer to Kafka Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
	}

	funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
	if err != nil {
		c.logger.Errorf("Failed to convert LabelSelector %v in Kafka Trigger object %s to Selector due to %v: ", triggerObj.Spec.FunctionSelector, key, err)
		return err
	}
	functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
	if err != nil {
		c.logger.Errorf("Failed to list associated functions with Kafka trigger %s by Selector due to %s: ", key, err)
	}

	if len(functions) == 0 {
		c.logger.Infof("No matching functions with selector %v found in namespace %s", funcSelector, ns)
	}

	for _, function := range functions {
		funcName := function.ObjectMeta.Name
		err = kafka.CreateKafkaConsumer(triggerObjName, funcName, ns, topic, c.kubernetesClient)
		if err != nil {
			c.logger.Errorf("Failed to create the Kafka consumer for the function %s associated with the Kafka trigger %s due to %v: ", funcName, key, err)
		}
	}

	c.logger.Infof("Processed change to Kafka trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
	return nil
}

// FunctionAddedDeletedUpdated process the updates to Function objects
func (c *KafkaTriggerController) FunctionAddedDeletedUpdated(obj interface{}, deleted bool) {
	functionObj, ok := obj.(*kubelessApi.Function)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			c.logger.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		functionObj, ok = tombstone.Obj.(*kubelessApi.Function)
		if !ok {
			c.logger.Errorf("Tombstone contained object that is not a Function object %#v", obj)
			return
		}
	}

	c.logger.Infof("Processing update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	kafkaTriggers, err := c.kafkaInformer.Lister().KafkaTriggers(functionObj.ObjectMeta.Namespace).List(labels.Everything())
	if err != nil {
		c.logger.Errorf("Failed to get list of Kafka trigger in namespace %s due to %s: ", functionObj.ObjectMeta.Namespace, err)
		return
	}

	for _, triggerObj := range kafkaTriggers {
		funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
		if err != nil {
			c.logger.Errorf("Failed to convert LabelSelector to Selector due to %s: ", err)
		}
		if !funcSelector.Matches(labels.Set(functionObj.Labels)) {
			continue
		}
		if deleted {
			c.logger.Infof("We got a Kafka trigger  %s that is associated with deleted function %s so cleanup Kafka consumer", triggerObj.Name, functionObj.Name)
			kafka.DeleteKafkaConsumer(triggerObj.ObjectMeta.Name, functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace, triggerObj.Spec.Topic)
			c.logger.Infof("Successfully removed Kafka consumer for Function: %s", functionObj.Name)
		} else {
			c.logger.Infof("We got a Kafka trigger  %s that function %s need to be associated so create Kafka consumer", triggerObj.Name, functionObj.Name)
			kafka.CreateKafkaConsumer(triggerObj.Name, functionObj.Name, functionObj.Namespace, triggerObj.Spec.Topic, c.kubernetesClient)
			c.logger.Infof("Successfully created Kafka consumer for Function: %s", functionObj.Name)
		}
	}
	c.logger.Infof("Successfully processed update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
}

func (c *KafkaTriggerController) kafkaTriggerObjNeedFinalizer(triggercObj *kubelessApi.KafkaTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == kafkaTriggerFinalizer {
			return false
		}
	}
	return triggercObj.ObjectMeta.DeletionTimestamp == nil
}

func (c *KafkaTriggerController) kafkaTriggerHasFinalizer(triggercObj *kubelessApi.KafkaTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == kafkaTriggerFinalizer {
			return true
		}
	}
	return false
}

func (c *KafkaTriggerController) kafkaTriggerObjAddFinalizer(triggercObj *kubelessApi.KafkaTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	triggercObjClone.ObjectMeta.Finalizers = append(triggercObjClone.ObjectMeta.Finalizers, kafkaTriggerFinalizer)
	return utils.UpdateKafkaTriggerCustomResource(c.kubelessclient, triggercObjClone)
}

func (c *KafkaTriggerController) kafkaTriggerObjRemoveFinalizer(triggercObj *kubelessApi.KafkaTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range triggercObjClone.ObjectMeta.Finalizers {
		if item == kafkaTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	triggercObjClone.ObjectMeta.Finalizers = newSlice
	err := utils.UpdateKafkaTriggerCustomResource(c.kubelessclient, triggercObjClone)
	if err != nil {
		return err
	}
	return nil
}

func kafkaTriggerObjChanged(oldKafkaTriggerObj, newKafkaTriggerObj *kubelessApi.KafkaTrigger) bool {
	// If the kafka trigger object's deletion timestamp is set, then process
	if oldKafkaTriggerObj.DeletionTimestamp != newKafkaTriggerObj.DeletionTimestamp {
		return true
	}
	// If the new and old kafka trigger object's resource version is same
	if oldKafkaTriggerObj.ResourceVersion == newKafkaTriggerObj.ResourceVersion {
		return false
	}
	if !reflect.DeepEqual(oldKafkaTriggerObj.Spec.FunctionSelector, newKafkaTriggerObj.Spec.FunctionSelector) {
		return true
	}
	if oldKafkaTriggerObj.Spec.Topic != newKafkaTriggerObj.Spec.Topic {
		return true
	}
	return false
}
