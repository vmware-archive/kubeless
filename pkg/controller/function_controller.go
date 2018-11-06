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
	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/ghodss/yaml"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kv1beta1 "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/registry"
	"github.com/kubeless/kubeless/pkg/utils"
)

const (
	maxRetries        = 5
	funcKind          = "Function"
	funcAPIVersion    = "kubeless.io/v1beta1"
	functionFinalizer = "kubeless.io/function"
)

// FunctionController object
type FunctionController struct {
	logger           *logrus.Entry
	clientset        kubernetes.Interface
	kubelessclient   versioned.Interface
	smclient         *monitoringv1alpha1.MonitoringV1alpha1Client
	Functions        map[string]*kubelessApi.Function
	queue            workqueue.RateLimitingInterface
	informer         cache.SharedIndexInformer
	config           *corev1.ConfigMap
	langRuntime      *langruntime.Langruntimes
	imagePullSecrets []corev1.LocalObjectReference
}

// Config contains k8s client of a controller
type Config struct {
	KubeCli        kubernetes.Interface
	FunctionClient versioned.Interface
}

// NewFunctionController returns a new *FunctionController
func NewFunctionController(cfg Config, smclient *monitoringv1alpha1.MonitoringV1alpha1Client) *FunctionController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	apiExtensionsClientset := utils.GetAPIExtensionsClientInCluster()
	config, err := utils.GetKubelessConfig(cfg.KubeCli, apiExtensionsClientset)
	if err != nil {
		logrus.Fatalf("Unable to read the configmap: %s", err)
	}

	informer := kv1beta1.NewFunctionInformer(cfg.FunctionClient, config.Data["functions-namespace"], 0, cache.Indexers{})

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
				newFunctionObj := new.(*kubelessApi.Function)
				oldFunctionObj := old.(*kubelessApi.Function)
				if functionObjChanged(oldFunctionObj, newFunctionObj) {
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

	var lr = langruntime.New(config)
	lr.ReadConfigMap()

	imagePullSecrets := utils.GetSecretsAsLocalObjectReference(config.Data["provision-image-secret"], config.Data["builder-image-secret"])
	if config.Data["enable-build-step"] == "true" {
		imagePullSecrets = append(imagePullSecrets, utils.GetSecretsAsLocalObjectReference("kubeless-registry-credentials")...)
	}
	return &FunctionController{
		logger:           logrus.WithField("pkg", "function-controller"),
		clientset:        cfg.KubeCli,
		smclient:         smclient,
		kubelessclient:   cfg.FunctionClient,
		informer:         informer,
		queue:            queue,
		config:           config,
		langRuntime:      lr,
		imagePullSecrets: imagePullSecrets,
	}
}

// Run starts the kubeless controller
func (c *FunctionController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting Function controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Function controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *FunctionController) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *FunctionController) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *FunctionController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *FunctionController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
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

func (c *FunctionController) processItem(key string) error {
	c.logger.Infof("Processing change to Function %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// this is an update when Function API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("Function object %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	funcObj := obj.(*kubelessApi.Function)

	// Function API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	if funcObj.ObjectMeta.DeletionTimestamp != nil {

		// If finalizer is removed, then we already processed the delete update, so just return
		if !utils.FunctionObjHasFinalizer(funcObj, functionFinalizer) {
			return nil
		}

		// Function object should be deleted, so cleanup the associated resources and remove the finalizer
		err := c.deleteK8sResources(ns, name)
		if err != nil {
			c.logger.Errorf("Can't delete function: %v", err)
			return err
		}

		// remove finalizer from the function object, so that we dont have to process any further and object can be deleted
		err = utils.FunctionObjRemoveFinalizer(c.kubelessclient, funcObj, functionFinalizer)
		if err != nil {
			c.logger.Errorf("Failed to remove function controller as finalizer to Function Obj: %s object due to: %v: ", key, err)
			return err
		}
		c.logger.Infof("Function object %s has been successfully processed and marked for deletion", key)
		return nil
	}

	// If function object in not marked with self as finalizer, then add the finalizer
	if !utils.FunctionObjHasFinalizer(funcObj, functionFinalizer) {
		err = utils.FunctionObjAddFinalizer(c.kubelessclient, funcObj, functionFinalizer)
		if err != nil {
			c.logger.Errorf("Error adding Function controller as finalizer to Function Obj: %s CRD due to: %v: ", key, err)
			return err
		}
	}

	err = c.ensureK8sResources(funcObj)
	if err != nil {
		c.logger.Errorf("Function can not be created/updated: %v", err)
		return err
	}

	c.logger.Infof("Processed change to function: %s", key)
	return nil
}

// startImageBuildJob creates (if necessary) a job that will build an image for the given function
// returns the name of the image, a boolean indicating if the build job has been created and an error
func (c *FunctionController) startImageBuildJob(funcObj *kubelessApi.Function, or []metav1.OwnerReference) (string, bool, error) {
	imagePullSecret, err := c.clientset.CoreV1().Secrets(funcObj.ObjectMeta.Namespace).Get("kubeless-registry-credentials", metav1.GetOptions{})
	if err != nil {
		return "", false, fmt.Errorf("Unable to locate registry credentials to build function image: %v", err)
	}
	reg, err := registry.New(*imagePullSecret)
	if err != nil {
		return "", false, fmt.Errorf("Unable to retrieve registry information: %v", err)
	}
	// Use function content and deps as tag (digested)
	tag := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v%v", funcObj.Spec.Function, funcObj.Spec.Deps))))
	imageName := fmt.Sprintf("%s/%s", reg.Creds.Username, funcObj.ObjectMeta.Name)
	// Check if image already exists
	exists, err := reg.ImageExists(imageName, tag)
	if err != nil {
		return "", false, fmt.Errorf("Unable to check is target image exists: %v", err)
	}
	regURL, err := url.Parse(reg.Endpoint)
	if err != nil {
		return "", false, fmt.Errorf("Unable to parse registry URL: %v", err)
	}
	image := fmt.Sprintf("%s/%s:%s", regURL.Host, imageName, tag)
	if !exists {
		tlsVerify := true
		if c.config.Data["function-registry-tls-verify"] == "false" {
			tlsVerify = false
		}
		err = utils.EnsureFuncImage(c.clientset, funcObj, c.langRuntime, or, imageName, tag, c.config.Data["builder-image"], regURL.Host, imagePullSecret.Name, c.config.Data["provision-image"], tlsVerify, c.imagePullSecrets)
		if err != nil {
			return "", false, fmt.Errorf("Unable to create image build job: %v", err)
		}
	} else {
		// Image already exists
		return image, false, nil
	}
	return image, true, nil
}

// ensureK8sResources creates/updates k8s objects (deploy, svc, configmap) for the function
func (c *FunctionController) ensureK8sResources(funcObj *kubelessApi.Function) error {
	if len(funcObj.ObjectMeta.Labels) == 0 {
		funcObj.ObjectMeta.Labels = make(map[string]string)
	}
	funcObj.ObjectMeta.Labels["function"] = funcObj.ObjectMeta.Name

	deployment := v1beta1.Deployment{}
	if deploymentConfigData, ok := c.config.Data["deployment"]; ok {
		err := yaml.Unmarshal([]byte(deploymentConfigData), &deployment)
		if err != nil {
			logrus.Errorf("Error parsing Deployment data in ConfigMap kubeless-function-deployment-config: %v", err)
			return err
		}
		err = utils.MergeDeployments(&funcObj.Spec.Deployment, &deployment)
		if err != nil {
			logrus.Errorf(" Error while merging function.Spec.Deployment and Deployment from ConfigMap: %v", err)
			return err
		}
	}

	or, err := utils.GetOwnerReference(funcKind, funcAPIVersion, funcObj.Name, funcObj.UID)
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

	prebuiltImage := ""
	if len(funcObj.Spec.Deployment.Spec.Template.Spec.Containers) > 0 && funcObj.Spec.Deployment.Spec.Template.Spec.Containers[0].Image != "" {
		prebuiltImage = funcObj.Spec.Deployment.Spec.Template.Spec.Containers[0].Image
	}
	// Skip image build step if using a custom runtime
	if prebuiltImage == "" {
		if c.config.Data["enable-build-step"] == "true" {
			var isBuilding bool
			prebuiltImage, isBuilding, err = c.startImageBuildJob(funcObj, or)
			if err != nil {
				logrus.Errorf("Unable to build function: %v", err)
			} else {
				if isBuilding {
					logrus.Infof("Started build process for function %s", funcObj.ObjectMeta.Name)
				} else {
					logrus.Infof("Found existing image %s", prebuiltImage)
				}
			}
		}
	} else {
		logrus.Infof("Skipping image-build step for %s", funcObj.ObjectMeta.Name)
	}

	err = utils.EnsureFuncDeployment(c.clientset, funcObj, or, c.langRuntime, prebuiltImage, c.config.Data["provision-image"], c.imagePullSecrets)
	if err != nil {
		return err
	}

	if funcObj.Spec.HorizontalPodAutoscaler.Name != "" && funcObj.Spec.HorizontalPodAutoscaler.Spec.ScaleTargetRef.Name != "" {
		funcObj.Spec.HorizontalPodAutoscaler.OwnerReferences = or
		if funcObj.Spec.HorizontalPodAutoscaler.Spec.Metrics[0].Type == v2beta1.ObjectMetricSourceType {
			// A service monitor is needed when the metric is an object
			err = utils.CreateServiceMonitor(*c.smclient, funcObj, funcObj.ObjectMeta.Namespace, or)
			if err != nil {
				return err
			}
		}
		err = utils.CreateAutoscale(c.clientset, funcObj.Spec.HorizontalPodAutoscaler)
		if err != nil {
			return err
		}
	} else {
		// HorizontalPodAutoscaler doesn't exists, try to delete if it already existed
		err = c.deleteAutoscale(funcObj.ObjectMeta.Namespace, funcObj.ObjectMeta.Name)
		if err != nil && !k8sErrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *FunctionController) deleteAutoscale(ns, name string) error {
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
func (c *FunctionController) deleteK8sResources(ns, name string) error {

	// delete deployment
	deletePolicy := metav1.DeletePropagationBackground
	err := c.clientset.Extensions().Deployments(ns).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
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

	// delete service monitor
	err = c.deleteAutoscale(ns, name)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	// delete build job
	err = c.clientset.BatchV1().Jobs(ns).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("created-by=kubeless,function=%s", name),
	})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	return nil
}

func functionObjChanged(oldFunctionObj, newFunctionObj *kubelessApi.Function) bool {
	// If the function object's deletion timestamp is set, then process
	if oldFunctionObj.DeletionTimestamp != newFunctionObj.DeletionTimestamp {
		return true
	}
	// If the new and old function object's resource version is same
	if oldFunctionObj.ResourceVersion == newFunctionObj.ResourceVersion {
		return false
	}
	newSpec := &oldFunctionObj.Spec
	oldSpec := &newFunctionObj.Spec

	if newSpec.Function != oldSpec.Function ||
		// compare checksum since the url content type uses Function field to pass the URL for the function
		// comparing the checksum ensures that if the function code has changed but the URL remains the same, the function will get redeployed
		newSpec.Checksum != oldSpec.Checksum ||
		newSpec.Handler != oldSpec.Handler ||
		newSpec.FunctionContentType != oldSpec.FunctionContentType ||
		newSpec.Deps != oldSpec.Deps ||
		newSpec.Timeout != oldSpec.Timeout {
		return true
	}

	if !apiequality.Semantic.DeepEqual(newSpec.Deployment, oldSpec.Deployment) ||
		!apiequality.Semantic.DeepEqual(newSpec.HorizontalPodAutoscaler, oldSpec.HorizontalPodAutoscaler) ||
		!apiequality.Semantic.DeepEqual(newSpec.ServiceSpec, oldSpec.ServiceSpec) {
		return true
	}
	return false
}
