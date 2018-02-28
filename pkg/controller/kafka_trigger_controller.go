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
	"os"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apimachineryHelpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	triggerMaxRetries     = 5
	kafkaTriggerFinalizer = "kubeless.io/kafkatrigger"
)

var (
	stopM    map[string](chan struct{})
	stoppedM map[string](chan struct{})
)

func init() {
	stopM = make(map[string](chan struct{}))
	stoppedM = make(map[string](chan struct{}))
}

// KafkaTriggerController object
type KafkaTriggerController struct {
	logger           *logrus.Entry
	clientset        kubernetes.Interface
	kubelessclient   versioned.Interface
	queue            workqueue.RateLimitingInterface
	kafkaInformer    kubelessInformers.KafkaTriggerInformer
	functionInformer kubelessInformers.FunctionInformer
	config           *corev1.ConfigMap
	langRuntime      *langruntime.Langruntimes
}

// KafkaTriggerConfig contains k8s client of a controller
type KafkaTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewKafkaTriggerController initializes a controller object
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

	controller := KafkaTriggerController{
		logger:           logrus.WithField("controller", "kafka-trigger-controller"),
		clientset:        cfg.KubeCli,
		kubelessclient:   cfg.TriggerClient,
		kafkaInformer:    kafkaInformer,
		functionInformer: functionInformer,
		queue:            queue,
		config:           config,
		langRuntime:      lr,
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
func (c *KafkaTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Kafka Trigger controller")

	go c.kafkaInformer.Informer().Run(stopCh)
	go c.functionInformer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Kafka Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *KafkaTriggerController) HasSynced() bool {
	return c.kafkaInformer.Informer().HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *KafkaTriggerController) LastSyncResourceVersion() string {
	return c.kafkaInformer.Informer().LastSyncResourceVersion()
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

func (c *KafkaTriggerController) processItem(key string) error {
	c.logger.Infof("Processing update to Kafka Trigger: %s", key)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.kafkaInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		c.logger.Infof("Kafka Trigger %s not found, ignoring", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.KafkaTrigger)
	topics := triggerObj.Spec.Topic
	if topics == "" {
		return errors.New("Kafka Trigger Topic can't be empty. Please check the trigger object %s" + key)
	}

	if triggerObj.ObjectMeta.DeletionTimestamp != nil && c.kafkaTriggerHasFinalizer(triggerObj) {
		// Kafka Trigger object should be deleted, so proecess the associated functions and remove the finalizer
		funcSelector, err := apimachineryHelpers.LabelSelectorAsSelector(&triggerObj.Spec.FunctionSelector)
		if err != nil {
			c.logger.Errorf("Failed to convert LabelSelector %v in Kafka Trigger object %s to Selector due to %v: ", triggerObj.Spec.FunctionSelector, key, err)
			return err
		}
		functions, err := c.functionInformer.Lister().Functions(ns).List(funcSelector)
		if err != nil {
			c.logger.Errorf("Failed to list associated functions with Kafka trigger %s by Selector due to %s: ", key, err)
			return err
		}

		if len(functions) == 0 {
			c.logger.Infof("No matching functions found for Kafka trigger %s so marking CRD object for deleteion", key)
		}

		for _, function := range functions {
			if err != nil {
				c.logger.Errorf("Unable to find the function %s in the namespace %s due to %v: ", function.ObjectMeta.Name, ns, err)
				return err
			}
			funcName := function.ObjectMeta.Name
			c.logger.Infof("Stopping consumer for trigger %s", name)
			kafka.DeleteKafkaConsumer(stopM, stoppedM, funcName, ns)
			c.logger.Infof("Stopped consumer for trigger %s", name)
		}

		err = c.kafkaTriggerObjRemoveFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Failed to remove Kafka trigger controller as finalizer to Kafka Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
		c.logger.Infof("Kafka trigger %s has been successfully processed and marked for deleteion", key)
		return nil
	}

	if c.kafkaTriggerObjNeedFinalizer(triggerObj) {
		err = c.kafkaTriggerObjAddFinalizer(triggerObj)
		if err != nil {
			c.logger.Errorf("Error adding Kafka trigger controller as finalizer to Kafka Trigger Obj: %s CRD object due to: %v: ", triggerObj.ObjectMeta.Name, err)
			return err
		}
	}

	// taking brokers from env var
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka.kubeless:9092"
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
		if err != nil {
			c.logger.Errorf("Unable to find the function %s in the namespace %s. Received %s: ", function.ObjectMeta.Name, ns, err)
			return err
		}
		if !utils.FunctionObjHasFinalizer(function, kafkaTriggerFinalizer) {
			err = utils.FunctionObjAddFinalizer(c.kubelessclient, function, kafkaTriggerFinalizer)
			if err != nil {
				c.logger.Errorf("Error adding Kafka trigger controller as finalizer to Function: %s CRD object due to: %s: ", function.ObjectMeta.Name, err)
			}
		}
		funcName := function.ObjectMeta.Name
		c.logger.Infof("Creating consumer: broker %s - topic %s - function %s - namespace %s", brokers, topics, funcName, ns)
		kafka.CreateKafkaConsumer(stopM, stoppedM, brokers, topics, funcName, ns)
		c.logger.Infof("Created consumer successfully")
	}

	c.logger.Infof("Processed change to Kafka trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
	return nil
}

func (c *KafkaTriggerController) getResouceGroupVersion(target string) (string, error) {
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

func (c *KafkaTriggerController) functionAddedDeletedUpdated(obj interface{}, deleted bool) {
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

	if !utils.FunctionObjHasFinalizer(functionObj, kafkaTriggerFinalizer) {
		return
	}

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
		functions, err := c.functionInformer.Lister().Functions(functionObj.ObjectMeta.Namespace).List(funcSelector)
		if err != nil {
			c.logger.Errorf("Failed to list function by Selector due to %s: ", err)
		}
		for _, matchedFunction := range functions {
			if matchedFunction.ObjectMeta.Name == functionObj.ObjectMeta.Name {
				c.logger.Infof("We got a Kafka trigger Name %s that is associated with deleted function %s so disassociate with the function", triggerObj.ObjectMeta.Name, functionObj.ObjectMeta.Name)
				kafka.DeleteKafkaConsumer(stopM, stoppedM, functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace)
			}
		}
	}

	err = utils.FunctionObjRemoveFinalizer(c.kubelessclient, functionObj, kafkaTriggerFinalizer)
	if err != nil {
		c.logger.Errorf("Error removing Kakfa trigger controller as finalizer to Function: %s CRD object due to: %s: ", functionObj.ObjectMeta.Name, err)
		return
	}

	c.logger.Infof("Successfully removed Kafka trigger controller as finalizer to Function: %s CRD object", functionObj.ObjectMeta.Name)
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
