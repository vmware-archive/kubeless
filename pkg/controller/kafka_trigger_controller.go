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
	"os"
	"time"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kv1beta1 "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	triggerMaxRetries = 5
	objKind           = "Trigger"
	objAPI            = "kubeless.io"
)

// KafkaTriggerController object
type KafkaTriggerController struct {
	logger         *logrus.Entry
	clientset      kubernetes.Interface
	kubelessclient versioned.Interface
	smclient       *monitoringv1alpha1.MonitoringV1alpha1Client
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
func NewKafkaTriggerController(cfg KafkaTriggerConfig, smclient *monitoringv1alpha1.MonitoringV1alpha1Client) *KafkaTriggerController {
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
		logger:         logrus.WithField("controller", "trigger-controller"),
		clientset:      cfg.KubeCli,
		smclient:       smclient,
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

	// run one round of GC at startup to detect orphaned objects from the last time
	c.garbageCollect()

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
		err := c.deleteKafkaTriggerResources(ns, name)
		if err != nil {
			c.logger.Errorf("Can't delete function: %v", err)
			return err
		}
		c.logger.Infof("Deleted Function %s", key)
		return nil
	}

	triggerObj := obj.(*kubelessApi.KafkaTrigger)

	funcObj, err := utils.GetFunction(triggerObj.Spec.FunctionName, ns)
	if err != nil {
		c.logger.Errorf("Unable to find the function %s in the namespace %s. Received %s: ", triggerObj.Spec.FunctionName, ns, err)
		return err
	}

	err = c.ensureKafkaTriggerResources(triggerObj, &funcObj)
	if err != nil {
		c.logger.Errorf("Function can not be created/updated: %v", err)
		return err
	}

	c.logger.Infof("Processed change to Trigger: %s Namespace: %s", triggerObj.ObjectMeta.Name, ns)
	return nil
}

// ensureK8sResources creates/updates k8s objects (deploy, svc, configmap) for the function
func (c *KafkaTriggerController) ensureKafkaTriggerResources(triggerObj *kubelessApi.KafkaTrigger, funcObj *kubelessApi.Function) error {
	if len(funcObj.ObjectMeta.Labels) == 0 {
		funcObj.ObjectMeta.Labels = make(map[string]string)
	}
	funcObj.ObjectMeta.Labels["function"] = funcObj.ObjectMeta.Name

	or, err := getKafkaTriggerOwnerReference(triggerObj)
	if err != nil {
		return err
	}

	err = utils.EnsureFuncConfigMap(c.clientset, funcObj, or, c.langRuntime)
	if err != nil {
		return err
	}

	err = utils.EnsureFuncService(c.clientset, funcObj, or)
	if err != nil {
		return err
	}

	err = utils.EnsureFuncDeployment(c.clientset, funcObj, or, c.langRuntime)
	if err != nil {
		return err
	}

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

func (c *KafkaTriggerController) deleteAutoscale(ns, name string) error {
	if c.smclient != nil {
		// Delete Service monitor if the client is available
		err := utils.DeleteServiceMonitor(*c.smclient, name, ns)
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
	}
	// delete autoscale
	err := utils.DeleteAutoscale(c.clientset, name, ns)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

// deleteK8sResources removes k8s objects of the function
func (c *KafkaTriggerController) deleteKafkaTriggerResources(ns, name string) error {
	//check if func is scheduled or not
	_, err := c.clientset.BatchV2alpha1().CronJobs(ns).Get(fmt.Sprintf("trigger-%s", name), metav1.GetOptions{})
	if err == nil {
		err = c.clientset.BatchV2alpha1().CronJobs(ns).Delete(fmt.Sprintf("trigger-%s", name), &metav1.DeleteOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
	}

	// delete deployment
	deletePolicy := metav1.DeletePropagationBackground
	err = c.clientset.Extensions().Deployments(ns).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	// delete svc
	err = c.clientset.Core().Services(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	// delete cm
	err = c.clientset.Core().ConfigMaps(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (c *KafkaTriggerController) garbageCollect() error {
	err := c.collectServices()
	if err != nil {
		return err
	}
	err = c.collectDeployment()
	if err != nil {
		return err
	}
	err = c.collectConfigMap()
	if err != nil {
		return err
	}
	return nil
}

func (c *KafkaTriggerController) collectServices() error {
	srvs, err := c.clientset.CoreV1().Services(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, srv := range srvs.Items {
		if len(srv.OwnerReferences) == 0 {
			continue
		}
		// Include the derived key from existing svc owner reference to the workqueue
		// This will make sure the controller can detect the non-existing function and
		// react to delete its belonging objects
		// Assumption: a service has ownerref Kind = "Function" and APIVersion = "k8s.io" is assumed
		// to be created by kubeless controller
		if (srv.OwnerReferences[0].Kind == objKind) && (srv.OwnerReferences[0].APIVersion == objAPI) {
			//service and its function are deployed in the same namespace
			key := fmt.Sprintf("%s/%s", srv.Namespace, srv.OwnerReferences[0].Name)
			c.queue.Add(key)
		}
	}

	return nil
}

func (c *KafkaTriggerController) collectDeployment() error {
	ds, err := c.clientset.AppsV1beta1().Deployments(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, d := range ds.Items {
		if len(d.OwnerReferences) == 0 {
			continue
		}
		// Assumption: a deployment has ownerref Kind = "Function" and APIVersion = "k8s.io" is assumed
		// to be created by kubeless controller
		if (d.OwnerReferences[0].Kind == objKind) && (d.OwnerReferences[0].APIVersion == objAPI) {
			key := fmt.Sprintf("%s/%s", d.Namespace, d.OwnerReferences[0].Name)
			c.queue.Add(key)
		}
	}

	return nil
}

func (c *KafkaTriggerController) collectConfigMap() error {
	cm, err := c.clientset.CoreV1().ConfigMaps(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, m := range cm.Items {
		if len(m.OwnerReferences) == 0 {
			continue
		}
		// Assumption: a configmap has ownerref Kind = "Function" and APIVersion = "k8s.io" is assumed
		// to be created by kubeless controller
		if (m.OwnerReferences[0].Kind == objKind) && (m.OwnerReferences[0].APIVersion == objAPI) {
			key := fmt.Sprintf("%s/%s", m.Namespace, m.OwnerReferences[0].Name)
			c.queue.Add(key)
		}
	}

	return nil
}

// GetKafkaTriggerOwnerReference returns ownerRef for appending to objects's metadata for
// objects created by kafka-controller for the trigger deployment
func getKafkaTriggerOwnerReference(triggerObj *kubelessApi.KafkaTrigger) ([]metav1.OwnerReference, error) {
	if triggerObj.ObjectMeta.Name == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("Kafka trigger object name can't be empty")
	}
	if triggerObj.ObjectMeta.UID == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("uid of Kafka trigger %s can't be empty", triggerObj.ObjectMeta.Name)
	}
	return []metav1.OwnerReference{
		{
			Kind:       "KafkaTrigger",
			APIVersion: "kubeless.io",
			Name:       triggerObj.ObjectMeta.Name,
			UID:        triggerObj.ObjectMeta.UID,
		},
	}, nil
}
