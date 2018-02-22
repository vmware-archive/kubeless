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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kv1beta1 "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/event-consumers/kafka"
	"github.com/kubeless/kubeless/pkg/langruntime"
)

const (
	triggerMaxRetries = 5
	objKind           = "Trigger"
	objAPI            = "kubeless.io"
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
	logger         *logrus.Entry
	clientset      kubernetes.Interface
	kubelessclient versioned.Interface
	queue          workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
	config         *corev1.ConfigMap
	langRuntime    *langruntime.Langruntimes
}

// KafkaTriggerConfig contains k8s client of a controller
type KafkaTriggerConfig struct {
	KubeCli       kubernetes.Interface
	TriggerClient versioned.Interface
}

// NewKafkaTriggerController initializes a controller object
func NewKafkaTriggerController(cfg KafkaTriggerConfig) *KafkaTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := kv1beta1.NewKafkaTriggerInformer(cfg.TriggerClient, corev1.NamespaceAll, 0, cache.Indexers{})

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

	return &KafkaTriggerController{
		logger:         logrus.WithField("controller", "kafka-trigger-controller"),
		clientset:      cfg.KubeCli,
		kubelessclient: cfg.TriggerClient,
		informer:       informer,
		queue:          queue,
		config:         config,
		langRuntime:    lr,
	}
}

// Run starts the Trigger controller
func (c *KafkaTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Kafka Trigger controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Kafka Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *KafkaTriggerController) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *KafkaTriggerController) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
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
	c.logger.Infof("Processing change to Trigger %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	c.logger.Infof("Processing update to Kafka Trigger: %s Namespace: %s", name, ns)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		c.logger.Infof("Stopping consumer for trigger %s", name)
		// FIXME: should be funcName
		kafka.DeleteKafkaConsumer(stopM, stoppedM, name, ns)
		c.logger.Infof("Stopped consumer for trigger %s", name)
		return nil
	}

	triggerObj := obj.(*kubelessApi.KafkaTrigger)
	topics := triggerObj.Spec.Topic
	if topics == "" {
		return errors.New("Topic can't be empty. Please check the trigger object")
	}
	funcName := triggerObj.Spec.FunctionName
	if funcName == "" {
		return errors.New("Function name can't be empty. Please check the trigger object")
	}

	// taking brokers from env var
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka.kubeless:9092"
	}

	c.logger.Infof("Creating consumer: broker %s - topic %s - function %s - namespace %s", brokers, topics, funcName, ns)
	kafka.CreateKafkaConsumer(stopM, stoppedM, brokers, topics, funcName, ns)
	c.logger.Infof("Created consumer successfully")

	c.logger.Infof("Processed change to Trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
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
