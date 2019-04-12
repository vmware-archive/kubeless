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

	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
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
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	httpTriggerMaxRetries = 5
	httpTriggerKind       = "HTTPTrigger"
	httpTriggerAPIVersion = "kubeless.io/v1beta1"
	httpTriggerFinalizer  = "kubeless.io/httptrigger"
)

// HTTPTriggerController object
type HTTPTriggerController struct {
	logger              *logrus.Entry
	clientset           kubernetes.Interface
	kubelessclient      versioned.Interface
	queue               workqueue.RateLimitingInterface
	httpTriggerInformer kubelessInformers.HTTPTriggerInformer
	functionInformer    kubelessInformers.FunctionInformer
}

// HTTPTriggerConfig contains k8s client of a controller
type HTTPTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewHTTPTriggerController initializes a controller object
func NewHTTPTriggerController(cfg HTTPTriggerConfig, sharedInformerFactory externalversions.SharedInformerFactory) *HTTPTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	httpTrigggerInformer := sharedInformerFactory.Kubeless().V1beta1().HTTPTriggers()
	functionInformer := sharedInformerFactory.Kubeless().V1beta1().Functions()

	httpTrigggerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newObj := new.(*kubelessApi.HTTPTrigger)
				oldObj := old.(*kubelessApi.HTTPTrigger)
				if httpTriggerObjChanged(oldObj, newObj) {
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

	controller := HTTPTriggerController{
		logger:              logrus.WithField("controller", "http-trigger-controller"),
		clientset:           cfg.KubeCli,
		kubelessclient:      cfg.TriggerClient,
		httpTriggerInformer: httpTrigggerInformer,
		functionInformer:    functionInformer,
		queue:               queue,
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
func (c *HTTPTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting HTTP Trigger controller")

	go c.httpTriggerInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !c.waitForCacheSync(stopCh) {
		return
	}

	c.logger.Info("HTTP Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *HTTPTriggerController) waitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.httpTriggerInformer.Informer().HasSynced, c.functionInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for HTTP triggers controller to sync;"))
		return false
	}
	c.logger.Info("HTTP Trigger controller caches are synced and ready")
	return true
}

// HasSynced is required for the cache.Controller interface.
func (c *HTTPTriggerController) HasSynced() bool {
	return c.httpTriggerInformer.Informer().HasSynced()
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

	err := c.syncHTTPTrigger(key.(string))
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

func (c *HTTPTriggerController) syncHTTPTrigger(key string) error {
	c.logger.Infof("Processing update to HTTPTrigger: %s", key)

	_, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.httpTriggerInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when HTTP trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("HTTP Trigger %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	httpTriggerObj := obj.(*kubelessApi.HTTPTrigger)

	// HTTP trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if httpTriggerObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !c.httpTriggerObjHasFinalizer(httpTriggerObj) {
			return nil
		}

		// remove ingress resource if any. Ignore any error, as ingress resource will be GC'ed
		_ = utils.DeleteIngress(c.clientset, httpTriggerObj.Name, httpTriggerObj.Namespace)

		err = c.httpTriggerObjRemoveFinalizer(httpTriggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove HTTP trigger controller as finalizer to http trigger Obj: %s due to: %v: ", key, err)
			return err
		}
		c.logger.Infof("HTTP trigger object %s has been successfully processed and marked for deleteion", key)
		return nil
	}

	if !c.httpTriggerObjHasFinalizer(httpTriggerObj) {
		err = c.httpTriggerObjAddFinalizer(httpTriggerObj)
		if err != nil {
			c.logger.Errorf("Error adding HTTP trigger controller as finalizer to  HTTPTrigger Obj: %s CRD object due to: %v: ", key, err)
			return err
		}
		return nil
	}

	// create ingress resource if required
	c.logger.Infof("Adding ingress resource for http trigger Obj: %s ", key)
	or, err := utils.GetOwnerReference(httpTriggerKind, httpTriggerAPIVersion, httpTriggerObj.Name, httpTriggerObj.UID)
	if err != nil {
		return err
	}
	err = utils.CreateIngress(c.clientset, httpTriggerObj, or)
	if err != nil && !k8sErrors.IsAlreadyExists(err) {
		c.logger.Errorf("Failed to create ingress rule %s corresponding to http trigger Obj: %s due to: %v: ", httpTriggerObj.Name, key, err)
	}

	// delete ingress resource if not required
	c.logger.Infof("Processed update to HTTPTrigger: %s", key)
	return nil
}

// FunctionAddedDeletedUpdated process the updates to Function objects
func (c *HTTPTriggerController) functionAddedDeletedUpdated(obj interface{}, deleted bool) error {
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
			err := fmt.Errorf("Tombstone contained object that is not a Pod %#v", obj)
			c.logger.Errorf(err.Error())
			return err
		}
	}

	if deleted {
		c.logger.Infof("Function %s deleted. Removing associated http trigger", functionObj.Name)
		httptList, err := c.kubelessclient.KubelessV1beta1().HTTPTriggers(functionObj.Namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, httpTrigger := range httptList.Items {
			if httpTrigger.Spec.FunctionName == functionObj.Name {
				err = c.kubelessclient.KubelessV1beta1().HTTPTriggers(functionObj.Namespace).Delete(httpTrigger.Name, &metav1.DeleteOptions{})
				if err != nil && !k8sErrors.IsNotFound(err) {
					c.logger.Errorf("Failed to delete httptrigger created for the function %s in namespace %s, Error: %s", functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace, err)
					return err
				}
			}
		}
	}

	c.logger.Infof("Successfully processed update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	return nil
}

func (c *HTTPTriggerController) httpTriggerObjHasFinalizer(triggerObj *kubelessApi.HTTPTrigger) bool {
	currentFinalizers := triggerObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == httpTriggerFinalizer {
			return true
		}
	}
	return false
}

func (c *HTTPTriggerController) httpTriggerObjAddFinalizer(triggercObj *kubelessApi.HTTPTrigger) error {
	triggercObjClone := triggercObj.DeepCopy()
	triggercObjClone.ObjectMeta.Finalizers = append(triggercObjClone.ObjectMeta.Finalizers, httpTriggerFinalizer)
	return utils.UpdateHTTPTriggerCustomResource(c.kubelessclient, triggercObjClone)
}

func (c *HTTPTriggerController) httpTriggerObjRemoveFinalizer(triggercObj *kubelessApi.HTTPTrigger) error {
	triggerObjClone := triggercObj.DeepCopy()
	newSlice := make([]string, 0)
	for _, item := range triggerObjClone.ObjectMeta.Finalizers {
		if item == httpTriggerFinalizer {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	triggerObjClone.ObjectMeta.Finalizers = newSlice
	err := utils.UpdateHTTPTriggerCustomResource(c.kubelessclient, triggerObjClone)
	if err != nil {
		return err
	}
	return nil
}

func httpTriggerObjChanged(oldObj, newObj *kubelessApi.HTTPTrigger) bool {
	// If the function object's deletion timestamp is set, then process
	if oldObj.DeletionTimestamp != newObj.DeletionTimestamp {
		return true
	}
	// If the new and old function object's resource version is same
	if oldObj.ResourceVersion != newObj.ResourceVersion {
		return true
	}
	newSpec := &newObj.Spec
	oldSpec := &oldObj.Spec

	if !apiequality.Semantic.DeepEqual(newSpec, oldSpec) {
		return true
	}
	return false
}
