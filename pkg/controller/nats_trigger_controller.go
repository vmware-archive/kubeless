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
	"github.com/kubeless/kubeless/pkg/event-consumers/nats"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	natsTriggerMaxRetries = 5
	natsTriggerFinalizer  = "kubeless.io/natstrigger"
)

// NatsTriggerController object
type NatsTriggerController struct {
	logger           *logrus.Entry
	kubelessclient   versioned.Interface
	kubernetesClient kubernetes.Interface
	queue            workqueue.RateLimitingInterface
	natsInformer     kubelessInformers.NATSTriggerInformer
	functionInformer kubelessInformers.FunctionInformer
}

// NatsTriggerConfig contains config for NatsTriggerController
type NatsTriggerConfig struct {
	TriggerClient versioned.Interface
}

// NewNatsTriggerController returns a new *NatsTriggerController.
func NewNatsTriggerController(cfg NatsTriggerConfig) *NatsTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	sharedInformers := externalversions.NewSharedInformerFactory(cfg.TriggerClient, 0)
	natsInformer := sharedInformers.Kubeless().V1beta1().NATSTriggers()
	functionInformer := sharedInformers.Kubeless().V1beta1().Functions()

	natsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newNatsTriggerObj := new.(*kubelessApi.NATSTrigger)
				oldNatsTriggerObj := old.(*kubelessApi.NATSTrigger)
				if natsTriggerObjChanged(oldNatsTriggerObj, newNatsTriggerObj) {
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

	controller := NatsTriggerController{
		logger:           logrus.WithField("controller", "nats-trigger-controller"),
		kubelessclient:   cfg.TriggerClient,
		kubernetesClient: utils.GetClient(),
		natsInformer:     natsInformer,
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

// Run starts the nats trigger controller
func (c *NatsTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting NATS Trigger controller")
	defer c.logger.Info("Shutting down NATS Trigger controller")

	go c.natsInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !c.waitForCacheSync(stopCh) {
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *NatsTriggerController) waitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.natsInformer.Informer().HasSynced, c.functionInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for NATS triggers controller to sync;"))
		return false
	}
	c.logger.Info("NATS Trigger controller caches are synced and ready")
	return true
}

func (c *NatsTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *NatsTriggerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncNatsTrigger(key.(string))
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

func (c *NatsTriggerController) syncNatsTrigger(key string) error {
	c.logger.Infof("Processing update to NATS Trigger: %s", key)
	ns, triggerObjName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.natsInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when NATS trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("NATS Trigger %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.NATSTrigger)
	topic := triggerObj.Spec.Topic
	if topic == "" {
		return errors.New("NATS Trigger Topic can't be empty. Please check the trigger object %s" + key)
	}

	// NATS trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if triggerObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !c.natsTriggerHasFinalizer(triggerObj) {
			return nil
		}

		// NATS Trigger object should be deleted, so process the associated functions and remove the finalizer
		funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
		if err != nil {
			c.logger.Errorf("Failed to convert LabelSelector %v in NATS Trigger object %s to Selector due to %v: ", triggerObj.Spec.FunctionSelector, key, err)
			return err
		}
		functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
		if err != nil {
			c.logger.Errorf("Failed to list associated functions with NATS trigger %s by selector due to %s: ", key, err)
			return err
		}
		if len(functions) == 0 {
			c.logger.Infof("No matching functions found for NATS trigger %s so marking CRD object for deleteion", key)
		}

		for _, function := range functions {
			funcName := function.ObjectMeta.Name
			err = nats.DeleteNATSConsumer(triggerObjName, funcName, ns, topic)
			if err != nil {
				c.logger.Errorf("Failed to delete the NATS consumer for the function %s associated with the NATS trigger %s due to %v: ", funcName, key, err)
			}
		}

		// remove finalizer from the NATS trigger object, so that we dont have to process any further and object can be deleted
		err = c.natsTriggerObjRemoveFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove NATS trigger controller as finalizer to NATS Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
		c.logger.Infof("NATS trigger %s has been successfully processed and marked for deletion", key)
		return nil
	}

	// If NATS trigger API in not marked with self as finalizer, then add the finalizer
	if !c.natsTriggerHasFinalizer(triggerObj) {
		err = c.natsTriggerObjAddFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Error adding NATS trigger controller as finalizer to NATS Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
	}

	funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
	if err != nil {
		c.logger.Errorf("Failed to convert LabelSelector %v in NATS Trigger object %s to Selector due to %v: ", triggerObj.Spec.FunctionSelector, key, err)
		return err
	}
	functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
	if err != nil {
		c.logger.Errorf("Failed to list associated functions with NATS trigger %s by Selector due to %s: ", key, err)
	}

	if len(functions) == 0 {
		c.logger.Infof("No matching functions with selector %v found in namespace %s", funcSelector, ns)
	}

	for _, function := range functions {
		funcName := function.ObjectMeta.Name
		err = nats.CreateNATSConsumer(triggerObjName, funcName, ns, topic, c.kubernetesClient)
		if err != nil {
			c.logger.Errorf("Failed to create the NATS consumer for the function %s associated with the NATS trigger %s due to %v: ", funcName, key, err)
		}
	}

	c.logger.Infof("Processed change to NATS trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
	return nil
}

// FunctionAddedDeletedUpdated process the updates to Function objects
func (c *NatsTriggerController) FunctionAddedDeletedUpdated(obj interface{}, deleted bool) {
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
	natsTriggers, err := c.natsInformer.Lister().NATSTriggers(functionObj.ObjectMeta.Namespace).List(labels.Everything())
	if err != nil {
		c.logger.Errorf("Failed to get list of NATS trigger in namespace %s due to %s: ", functionObj.ObjectMeta.Namespace, err)
		return
	}

	for _, triggerObj := range natsTriggers {
		funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
		if err != nil {
			c.logger.Errorf("Failed to convert LabelSelector to Selector due to %s: ", err)
		}
		if !funcSelector.Matches(labels.Set(functionObj.Labels)) {
			continue
		}
		if deleted {
			c.logger.Infof("We got a NATS trigger  %s that is associated with deleted function %s so cleanup NATS consumer", triggerObj.Name, functionObj.Name)
			nats.DeleteNATSConsumer(triggerObj.ObjectMeta.Name, functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace, triggerObj.Spec.Topic)
			c.logger.Infof("Successfully removed NATS consumer for Function: %s", functionObj.Name)
		} else {
			c.logger.Infof("We got a NATS trigger  %s that function %s need to be associated so create NATS consumer", triggerObj.Name, functionObj.Name)
			nats.CreateNATSConsumer(triggerObj.Name, functionObj.Name, functionObj.Namespace, triggerObj.Spec.Topic, c.kubernetesClient)
			c.logger.Infof("Successfully created NATS consumer for Function: %s", functionObj.Name)
		}
	}
	c.logger.Infof("Successfully processed update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
}

func (c *NatsTriggerController) natsTriggerObjNeedFinalizer(triggercObj *kubelessApi.NATSTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == natsTriggerFinalizer {
			return false
		}
	}
	return triggercObj.ObjectMeta.DeletionTimestamp == nil
}

func (c *NatsTriggerController) natsTriggerHasFinalizer(triggercObj *kubelessApi.NATSTrigger) bool {
	currentFinalizers := triggercObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == natsTriggerFinalizer {
			return true
		}
	}
	return false
}

func (c *NatsTriggerController) natsTriggerObjAddFinalizer(triggercObj *kubelessApi.NATSTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	triggercObjClone.ObjectMeta.Finalizers = append(triggercObjClone.ObjectMeta.Finalizers, natsTriggerFinalizer)
	return utils.UpdateNatsTriggerCustomResource(c.kubelessclient, triggercObjClone)
}

func (c *NatsTriggerController) natsTriggerObjRemoveFinalizer(triggercObj *kubelessApi.NATSTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range triggercObjClone.ObjectMeta.Finalizers {
		if item == natsTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	triggercObjClone.ObjectMeta.Finalizers = newSlice
	err := utils.UpdateNatsTriggerCustomResource(c.kubelessclient, triggercObjClone)
	if err != nil {
		return err
	}
	return nil
}

func natsTriggerObjChanged(oldNatsTriggerObj, newNatsTriggerObj *kubelessApi.NATSTrigger) bool {
	// If the NATS trigger object's deletion timestamp is set, then process the update
	if oldNatsTriggerObj.DeletionTimestamp != newNatsTriggerObj.DeletionTimestamp {
		return true
	}
	// If the new and old NATS trigger object's resource version is same, then skip processing the update
	if oldNatsTriggerObj.ResourceVersion == newNatsTriggerObj.ResourceVersion {
		return false
	}
	if !apiequality.Semantic.DeepEqual(oldNatsTriggerObj.Spec, newNatsTriggerObj.Spec) {
		return true
	}
	return false
}
