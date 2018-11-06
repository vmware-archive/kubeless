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

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/ghodss/yaml"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	clientsetAPIExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetFunctionPort returns the port for a function service
func GetFunctionPort(clientset kubernetes.Interface, namespace, functionName string) (string, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Unable to find the service for function %s", functionName)
	}
	return strconv.Itoa(int(svc.Spec.Ports[0].Port)), nil
}

// IsJSON returns true if the string is json
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

func appendToCommand(orig string, command ...string) string {
	if len(orig) > 0 {
		return fmt.Sprintf("%s && %s", orig, strings.Join(command, " && "))
	}
	return strings.Join(command, " && ")
}

func getProvisionContainer(function, checksum, fileName, handler, contentType, runtime, prepareImage string, runtimeVolume, depsVolume v1.VolumeMount, lr *langruntime.Langruntimes) (v1.Container, error) {
	prepareCommand := ""
	originFile := path.Join(depsVolume.MountPath, fileName)

	// Prepare Function file and dependencies
	if strings.Contains(contentType, "base64") {
		// File is encoded in base64
		decodedFile := "/tmp/func.decoded"
		prepareCommand = appendToCommand(prepareCommand, fmt.Sprintf("base64 -d < %s > %s", originFile, decodedFile))
		originFile = decodedFile
	} else if strings.Contains(contentType, "url") {
		fromURLFile := "/tmp/func.fromurl"
		prepareCommand = appendToCommand(prepareCommand, fmt.Sprintf("curl %s -L --silent --output %s", function, fromURLFile))
		originFile = fromURLFile
	} else if strings.Contains(contentType, "text") || contentType == "" {
		// Assumming that function is plain text
		// So we don't need to preprocess it
	} else {
		return v1.Container{}, fmt.Errorf("Unable to prepare function of type %s: Unknown format", contentType)
	}

	// Validate checksum
	if checksum == "" {
		// DEPRECATED: Checksum may be empty
	} else {
		checksumInfo := strings.Split(checksum, ":")
		switch checksumInfo[0] {
		case "sha256":
			shaFile := "/tmp/func.sha256"
			prepareCommand = appendToCommand(prepareCommand,
				fmt.Sprintf("echo '%s  %s' > %s", checksumInfo[1], originFile, shaFile),
				fmt.Sprintf("sha256sum -c %s", shaFile),
			)
			break
		default:
			return v1.Container{}, fmt.Errorf("Unable to verify checksum %s: Unknown format", checksum)
		}
	}

	// Extract content in case it is a Zip file
	if strings.Contains(contentType, "zip") {
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("unzip -o %s -d %s", originFile, runtimeVolume.MountPath),
		)
	} else {
		// Copy the target as a single file
		destFileName, err := getFileName(handler, contentType, runtime, lr)
		if err != nil {
			return v1.Container{}, err
		}
		dest := path.Join(runtimeVolume.MountPath, destFileName)
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("cp %s %s", originFile, dest),
		)
	}

	// Copy deps file to the installation path
	runtimeInf, err := lr.GetRuntimeInfo(runtime)
	if err == nil && runtimeInf.DepName != "" {
		depsFile := path.Join(depsVolume.MountPath, runtimeInf.DepName)
		prepareCommand = appendToCommand(prepareCommand,
			fmt.Sprintf("cp %s %s", depsFile, runtimeVolume.MountPath),
		)
	}

	return v1.Container{
		Name:            "prepare",
		Image:           prepareImage,
		Command:         []string{"sh", "-c"},
		Args:            []string{prepareCommand},
		VolumeMounts:    []v1.VolumeMount{runtimeVolume, depsVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
	}, nil
}

func addDefaultLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["created-by"] = "kubeless"
	return labels
}

func hasDefaultLabel(labels map[string]string) bool {
	if labels == nil || labels["created-by"] != "kubeless" {
		return false
	}
	return true
}

func splitHandler(handler string) (string, string, error) {
	str := strings.Split(handler, ".")
	if len(str) != 2 {
		return "", "", fmt.Errorf("failed: incorrect handler format. It should be module_name.handler_name")
	}

	return str[0], str[1], nil
}

// getFileName returns a file name based on a handler identifier
func getFileName(handler, funcContentType, runtime string, lr *langruntime.Langruntimes) (string, error) {
	modName, _, err := splitHandler(handler)
	if err != nil {
		return "", err
	}
	filename := modName
	if funcContentType == "text" || funcContentType == "" || funcContentType == "url" {
		// We can only guess the extension if the function is specified as plain text
		runtimeInf, err := lr.GetRuntimeInfo(runtime)
		if err == nil {
			filename = modName + runtimeInf.FileNameSuffix
		}
	}
	return filename, nil
}

// EnsureFuncConfigMap creates/updates a config map with a function specification
func EnsureFuncConfigMap(client kubernetes.Interface, funcObj *kubelessApi.Function, or []metav1.OwnerReference, lr *langruntime.Langruntimes) error {
	configMapData := map[string]string{}
	var err error
	if funcObj.Spec.Handler != "" {
		fileName, err := getFileName(funcObj.Spec.Handler, funcObj.Spec.FunctionContentType, funcObj.Spec.Runtime, lr)
		if err != nil {
			return err
		}
		configMapData = map[string]string{
			"handler": funcObj.Spec.Handler,
			fileName:  funcObj.Spec.Function,
		}
		runtimeInfo, err := lr.GetRuntimeInfo(funcObj.Spec.Runtime)
		if err == nil && runtimeInfo.DepName != "" {
			configMapData[runtimeInfo.DepName] = funcObj.Spec.Deps
		}
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.ObjectMeta.Name,
			Labels:          addDefaultLabel(funcObj.ObjectMeta.Labels),
			OwnerReferences: or,
		},
		Data: configMapData,
	}

	_, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Create(configMap)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the ConfigMap already exists we should update
		// just certain fields (to avoid race conditions)
		var newConfigMap *v1.ConfigMap
		newConfigMap, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !hasDefaultLabel(newConfigMap.ObjectMeta.Labels) {
			return fmt.Errorf("Found a conflicting configmap object %s/%s. Aborting", funcObj.ObjectMeta.Namespace, funcObj.ObjectMeta.Name)
		}
		newConfigMap.ObjectMeta.Labels = funcObj.ObjectMeta.Labels
		newConfigMap.ObjectMeta.OwnerReferences = or
		newConfigMap.Data = configMap.Data
		_, err = client.Core().ConfigMaps(funcObj.ObjectMeta.Namespace).Update(newConfigMap)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The configmap may already exist and there is nothing to update
			return nil
		}
	}

	return err
}

// this function resolves backward incompatibility in case user uses old client which doesn't include serviceSpec into funcSpec.
// if serviceSpec is empty, we will use the default serviceSpec whose port is 8080
func serviceSpec(funcObj *kubelessApi.Function) v1.ServiceSpec {
	if len(funcObj.Spec.ServiceSpec.Ports) == 0 {
		return v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					// Note: Prefix: "http-" is added to adapt to Istio so that it can discover the function services
					Name:       "http-function-port",
					Protocol:   v1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: funcObj.ObjectMeta.Labels,
			Type:     v1.ServiceTypeClusterIP,
		}
	}
	return funcObj.Spec.ServiceSpec
}

// EnsureFuncService creates/updates a function service
func EnsureFuncService(client kubernetes.Interface, funcObj *kubelessApi.Function, or []metav1.OwnerReference) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            funcObj.ObjectMeta.Name,
			Labels:          addDefaultLabel(funcObj.ObjectMeta.Labels),
			OwnerReferences: or,
		},
		Spec: serviceSpec(funcObj),
	}

	_, err := client.Core().Services(funcObj.ObjectMeta.Namespace).Create(svc)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the SVC already exists we should update
		// just certain fields (to avoid race conditions)
		var newSvc *v1.Service
		newSvc, err = client.Core().Services(funcObj.ObjectMeta.Namespace).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !hasDefaultLabel(newSvc.ObjectMeta.Labels) {
			return fmt.Errorf("Found a conflicting service object %s/%s. Aborting", funcObj.ObjectMeta.Namespace, funcObj.ObjectMeta.Name)
		}
		newSvc.ObjectMeta.Labels = funcObj.ObjectMeta.Labels
		newSvc.ObjectMeta.OwnerReferences = or
		newSvc.Spec.Ports = svc.Spec.Ports
		_, err = client.Core().Services(funcObj.ObjectMeta.Namespace).Update(newSvc)
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			// The service may already exist and there is nothing to update
			return nil
		}
	}
	return err
}

func getRuntimeVolumeMount(name string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      name,
		MountPath: "/kubeless",
	}
}

func getChecksum(content string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(content))
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// populatePodSpec populates a basic Pod Spec that uses init containers to populate
// the runtime container with the function content and its dependencies.
// The caller should define the runtime container(s).
// It accepts a prepopulated podSpec with default information and volume that the
// runtime container should mount
func populatePodSpec(funcObj *kubelessApi.Function, lr *langruntime.Langruntimes, podSpec *v1.PodSpec, runtimeVolumeMount v1.VolumeMount, provisionImage string, imagePullSecrets []v1.LocalObjectReference) error {
	depsVolumeName := funcObj.ObjectMeta.Name + "-deps"
	result := podSpec
	if len(imagePullSecrets) > 0 {
		result.ImagePullSecrets = imagePullSecrets
	}
	result.Volumes = append(podSpec.Volumes,
		v1.Volume{
			Name: runtimeVolumeMount.Name,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		v1.Volume{
			Name: depsVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: funcObj.ObjectMeta.Name,
					},
				},
			},
		},
	)
	// prepare init-containers if some function is specified
	if funcObj.Spec.Function != "" {
		fileName, err := getFileName(funcObj.Spec.Handler, funcObj.Spec.FunctionContentType, funcObj.Spec.Runtime, lr)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		srcVolumeMount := v1.VolumeMount{
			Name:      depsVolumeName,
			MountPath: "/src",
		}
		provisionContainer, err := getProvisionContainer(
			funcObj.Spec.Function,
			funcObj.Spec.Checksum,
			fileName,
			funcObj.Spec.Handler,
			funcObj.Spec.FunctionContentType,
			funcObj.Spec.Runtime,
			provisionImage,
			runtimeVolumeMount,
			srcVolumeMount,
			lr,
		)
		if err != nil {
			return err
		}
		result.InitContainers = []v1.Container{provisionContainer}
	}

	// Add the imagesecrets if present to pull images from private docker registry
	if funcObj.Spec.Runtime != "" {
		imageSecrets, err := lr.GetImageSecrets(funcObj.Spec.Runtime)
		if err != nil {
			return fmt.Errorf("Unable to fetch ImagePullSecrets, %v", err)
		}
		result.ImagePullSecrets = append(result.ImagePullSecrets, imageSecrets...)
	}

	// ensure that the runtime is supported for installing dependencies
	_, err := lr.GetRuntimeInfo(funcObj.Spec.Runtime)
	if funcObj.Spec.Deps != "" && err != nil {
		return fmt.Errorf("Unable to install dependencies for the runtime %s", funcObj.Spec.Runtime)
	} else if funcObj.Spec.Deps != "" {
		envVars := []v1.EnvVar{}
		if len(result.Containers) > 0 {
			envVars = result.Containers[0].Env
		}
		depsChecksum, err := getChecksum(funcObj.Spec.Deps)
		if err != nil {
			return fmt.Errorf("Unable to obtain dependencies checksum: %v", err)
		}
		depsInstallContainer, err := lr.GetBuildContainer(funcObj.Spec.Runtime, depsChecksum, envVars, runtimeVolumeMount)
		if err != nil {
			return err
		}
		if depsInstallContainer.Name != "" {
			result.InitContainers = append(
				result.InitContainers,
				depsInstallContainer,
			)
		}
	}

	// add compilation init container if needed
	_, funcName, _ := splitHandler(funcObj.Spec.Handler)
	compContainer, err := lr.GetCompilationContainer(funcObj.Spec.Runtime, funcName, runtimeVolumeMount)
	if err != nil {
		return err
	}
	if compContainer != nil {
		result.InitContainers = append(
			result.InitContainers,
			*compContainer,
		)
	}
	return nil
}

// EnsureFuncImage creates a Job to build a function image
func EnsureFuncImage(client kubernetes.Interface, funcObj *kubelessApi.Function, lr *langruntime.Langruntimes, or []metav1.OwnerReference, imageName, tag, builderImage, registryHost, dockerSecretName, provisionImage string, registryTLSEnabled bool, imagePullSecrets []v1.LocalObjectReference) error {
	if len(tag) < 64 {
		return fmt.Errorf("Expecting sha256 as image tag")
	}
	jobName := fmt.Sprintf("build-%s-%s", funcObj.ObjectMeta.Name, tag[0:10])
	_, err := client.BatchV1().Jobs(funcObj.ObjectMeta.Namespace).Get(jobName, metav1.GetOptions{})
	if err == nil {
		// The job already exists
		logrus.Infof("Found a previous job for building %s:%s", imageName, tag)
		return nil
	}
	podSpec := v1.PodSpec{
		RestartPolicy: v1.RestartPolicyOnFailure,
	}
	runtimeVolumeMount := getRuntimeVolumeMount(funcObj.ObjectMeta.Name)
	err = populatePodSpec(funcObj, lr, &podSpec, runtimeVolumeMount, provisionImage, imagePullSecrets)
	if err != nil {
		return err
	}

	// Add a final initContainer to create the function bundle.tar
	prepareContainer := v1.Container{}
	for _, c := range podSpec.InitContainers {
		if c.Name == "prepare" {
			prepareContainer = c
		}
	}
	podSpec.InitContainers = append(podSpec.InitContainers, v1.Container{
		Name:         "bundle",
		Command:      []string{"sh", "-c"},
		Args:         []string{fmt.Sprintf("tar cvf %s/bundle.tar %s/*", runtimeVolumeMount.MountPath, runtimeVolumeMount.MountPath)},
		VolumeMounts: prepareContainer.VolumeMounts,
		Image:        provisionImage,
	})

	buildJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       funcObj.ObjectMeta.Namespace,
			OwnerReferences: or,
			Labels: addDefaultLabel(map[string]string{
				"function": funcObj.ObjectMeta.Name,
			}),
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: podSpec,
			},
		},
	}

	baseImage, err := lr.GetFunctionImage(funcObj.Spec.Runtime)
	if err != nil {
		return err
	}

	// Registry volume
	dockerCredsVol := dockerSecretName
	dockerCredsVolMountPath := "/docker"
	registryCredsVolume := v1.Volume{
		Name: dockerCredsVol,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: dockerSecretName,
			},
		},
	}
	buildJob.Spec.Template.Spec.Volumes = append(buildJob.Spec.Template.Spec.Volumes, registryCredsVolume)

	args := []string{
		"/imbuilder",
		"add-layer",
	}
	if !registryTLSEnabled {
		args = append(args, "--insecure")
	}
	args = append(args,
		"--src", fmt.Sprintf("docker://%s", baseImage),
		"--dst", fmt.Sprintf("docker://%s/%s:%s", registryHost, imageName, tag),
		fmt.Sprintf("%s/bundle.tar", podSpec.InitContainers[0].VolumeMounts[0].MountPath),
	)
	// Add main container
	buildJob.Spec.Template.Spec.Containers = []v1.Container{
		{
			Name:  "build",
			Image: builderImage,
			VolumeMounts: append(prepareContainer.VolumeMounts,
				v1.VolumeMount{
					Name:      dockerCredsVol,
					MountPath: dockerCredsVolMountPath,
				},
			),
			Env: []v1.EnvVar{
				{
					Name:  "DOCKER_CONFIG_FOLDER",
					Value: dockerCredsVolMountPath,
				},
			},
			Args: args,
		},
	}

	// Create the job if doesn't exists yet
	_, err = client.BatchV1().Jobs(funcObj.ObjectMeta.Namespace).Create(&buildJob)
	if err == nil {
		logrus.Infof("Started function build job %s", jobName)
	}
	return err
}

func svcPort(funcObj *kubelessApi.Function) int32 {
	if len(funcObj.Spec.ServiceSpec.Ports) == 0 {
		return int32(8080)
	}
	return funcObj.Spec.ServiceSpec.Ports[0].Port
}

func mergeMap(dst, src map[string]string) map[string]string {
	if len(dst) == 0 {
		dst = make(map[string]string)
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// EnsureFuncDeployment creates/updates a function deployment
func EnsureFuncDeployment(client kubernetes.Interface, funcObj *kubelessApi.Function, or []metav1.OwnerReference, lr *langruntime.Langruntimes, prebuiltRuntimeImage, provisionImage string, imagePullSecrets []v1.LocalObjectReference) error {

	var err error

	podAnnotations := map[string]string{
		// Attempt to attract the attention of prometheus.
		// For runtimes that don't support /metrics,
		// prometheus will get a 404 and mostly silently
		// ignore the pod (still displayed in the list of
		// "targets")
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   strconv.Itoa(int(svcPort(funcObj))),
	}
	maxUnavailable := intstr.FromInt(0)

	//add deployment and copy all func's Spec.Deployment to the deployment
	dpm := funcObj.Spec.Deployment.DeepCopy()
	dpm.OwnerReferences = or
	dpm.ObjectMeta.Name = funcObj.ObjectMeta.Name
	dpm.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: funcObj.ObjectMeta.Labels,
	}

	dpm.Spec.Strategy = v1beta1.DeploymentStrategy{
		RollingUpdate: &v1beta1.RollingUpdateDeployment{
			MaxUnavailable: &maxUnavailable,
		},
	}

	//append data to dpm deployment
	dpm.Labels = addDefaultLabel(mergeMap(dpm.Labels, funcObj.Labels))
	dpm.Spec.Template.Labels = mergeMap(dpm.Spec.Template.Labels, funcObj.Labels)
	dpm.Annotations = mergeMap(dpm.Annotations, funcObj.Annotations)
	dpm.Spec.Template.Annotations = mergeMap(dpm.Spec.Template.Annotations, funcObj.Annotations)
	dpm.Spec.Template.Annotations = mergeMap(dpm.Spec.Template.Annotations, podAnnotations)

	if len(dpm.Spec.Template.Spec.Containers) == 0 {
		dpm.Spec.Template.Spec.Containers = append(dpm.Spec.Template.Spec.Containers, v1.Container{})
	}

	runtimeVolumeMount := getRuntimeVolumeMount(funcObj.ObjectMeta.Name)
	if funcObj.Spec.Handler != "" && funcObj.Spec.Function != "" {
		modName, handlerName, err := splitHandler(funcObj.Spec.Handler)
		if err != nil {
			return err
		}
		//only resolve the image name and build the function if it has not been built already
		if dpm.Spec.Template.Spec.Containers[0].Image == "" && prebuiltRuntimeImage == "" {
			err := populatePodSpec(funcObj, lr, &dpm.Spec.Template.Spec, runtimeVolumeMount, provisionImage, imagePullSecrets)
			if err != nil {
				return err
			}

			imageName, err := lr.GetFunctionImage(funcObj.Spec.Runtime)
			if err != nil {
				return err
			}
			dpm.Spec.Template.Spec.Containers[0].Image = imageName

			dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, runtimeVolumeMount)

		} else {
			if dpm.Spec.Template.Spec.Containers[0].Image == "" {
				dpm.Spec.Template.Spec.Containers[0].Image = prebuiltRuntimeImage
			}
			dpm.Spec.Template.Spec.ImagePullSecrets = imagePullSecrets
		}
		timeout := funcObj.Spec.Timeout
		if timeout == "" {
			// Set default timeout to 180 seconds
			timeout = defaultTimeout
		}
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  "FUNC_HANDLER",
				Value: handlerName,
			},
			v1.EnvVar{
				Name:  "MOD_NAME",
				Value: modName,
			},
			v1.EnvVar{
				Name:  "FUNC_TIMEOUT",
				Value: timeout,
			},
			v1.EnvVar{
				Name:  "FUNC_RUNTIME",
				Value: funcObj.Spec.Runtime,
			},
			v1.EnvVar{
				Name:  "FUNC_MEMORY_LIMIT",
				Value: dpm.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String(),
			},
		)
	} else {
		logrus.Warn("Expected non-empty handler and non-empty function content")
	}

	dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  "FUNC_PORT",
			Value: strconv.Itoa(int(svcPort(funcObj))),
		},
	)

	dpm.Spec.Template.Spec.Containers[0].Name = funcObj.ObjectMeta.Name
	dpm.Spec.Template.Spec.Containers[0].Ports = append(dpm.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{
		ContainerPort: svcPort(funcObj),
	})

	// update deployment for loading dependencies
	lr.UpdateDeployment(dpm, runtimeVolumeMount.MountPath, funcObj.Spec.Runtime)

	livenessProbeInfo := lr.GetLivenessProbeInfo(funcObj.Spec.Runtime, int(svcPort(funcObj)))

	if dpm.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
		dpm.Spec.Template.Spec.Containers[0].LivenessProbe = livenessProbeInfo
	}

	// Add security context
	runtimeUser := int64(1000)
	if dpm.Spec.Template.Spec.SecurityContext == nil {
		dpm.Spec.Template.Spec.SecurityContext = &v1.PodSecurityContext{
			RunAsUser: &runtimeUser,
			FSGroup:   &runtimeUser,
		}
	}

	_, err = client.ExtensionsV1beta1().Deployments(funcObj.ObjectMeta.Namespace).Create(dpm)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		// In case the Deployment already exists we should update
		// just certain fields (to avoid race conditions)
		var newDpm *v1beta1.Deployment
		newDpm, err = client.ExtensionsV1beta1().Deployments(funcObj.ObjectMeta.Namespace).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !hasDefaultLabel(newDpm.ObjectMeta.Labels) {
			return fmt.Errorf("Found a conflicting deployment object %s/%s. Aborting", funcObj.ObjectMeta.Namespace, funcObj.ObjectMeta.Name)
		}
		newDpm.ObjectMeta.Labels = funcObj.ObjectMeta.Labels
		newDpm.ObjectMeta.Annotations = funcObj.Spec.Deployment.ObjectMeta.Annotations
		newDpm.ObjectMeta.OwnerReferences = or
		// We should maintain previous selector to avoid duplicated ReplicaSets
		selector := newDpm.Spec.Selector
		newDpm.Spec = dpm.Spec
		newDpm.Spec.Selector = selector
		data, err := json.Marshal(newDpm)
		if err != nil {
			return err
		}
		// Use `Patch` to do a rolling update
		_, err = client.ExtensionsV1beta1().Deployments(funcObj.ObjectMeta.Namespace).Patch(newDpm.Name, types.MergePatchType, data)
		if err != nil {
			return err
		}
	}

	return err
}

// CreateServiceMonitor creates a Service Monitor for the given function
func CreateServiceMonitor(smclient monitoringv1alpha1.MonitoringV1alpha1Client, funcObj *kubelessApi.Function, ns string, or []metav1.OwnerReference) error {
	_, err := smclient.ServiceMonitors(ns).Get(funcObj.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			s := &monitoringv1alpha1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      funcObj.ObjectMeta.Name,
					Namespace: ns,
					Labels: addDefaultLabel(map[string]string{
						"service-monitor": "function",
					}),
					OwnerReferences: or,
				},
				Spec: monitoringv1alpha1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"function": funcObj.ObjectMeta.Name,
						},
					},
					Endpoints: []monitoringv1alpha1.Endpoint{
						{
							Port: "http-function-port",
						},
					},
				},
			}
			_, err = smclient.ServiceMonitors(ns).Create(s)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("service monitor has already existed")
}

// GetOwnerReference returns ownerRef for appending to objects's metadata
func GetOwnerReference(kind, apiVersion, name string, uid types.UID) ([]metav1.OwnerReference, error) {
	if name == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("name can't be empty")
	}
	if uid == "" {
		return []metav1.OwnerReference{}, fmt.Errorf("uid can't be empty")
	}
	return []metav1.OwnerReference{
		{
			Kind:       kind,
			APIVersion: apiVersion,
			Name:       name,
			UID:        uid,
		},
	}, nil
}

// GetInClusterConfig returns necessary Config object to authenticate k8s clients if env variable is set
func GetInClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()

	tokenFile := os.Getenv("KUBELESS_TOKEN_FILE_PATH")
	if len(tokenFile) == 0 {
		return config, err
	}
	tokenBytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read file containing oauth token: %s", err)
	}
	config.BearerToken = string(tokenBytes)

	return config, nil
}

func getConfigLocation(apiExtensionsClientset clientsetAPIExtensions.Interface) (ConfigLocation, error) {
	configLocation := ConfigLocation{}
	controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
	kubelessConfig := os.Getenv("KUBELESS_CONFIG")

	annotationsCRD, err := GetAnnotationsFromCRD(apiExtensionsClientset, "functions.kubeless.io")
	if err != nil {
		return configLocation, err
	}
	if len(controllerNamespace) == 0 {
		if ns, ok := annotationsCRD["kubeless.io/namespace"]; ok {
			controllerNamespace = ns
		} else {
			controllerNamespace = "kubeless"
		}
	}
	configLocation.Namespace = controllerNamespace
	if len(kubelessConfig) == 0 {
		if config, ok := annotationsCRD["kubeless.io/config"]; ok {
			kubelessConfig = config
		} else {
			kubelessConfig = "kubeless-config"
		}
	}
	configLocation.Name = kubelessConfig
	return configLocation, nil
}

// GetKubelessConfig Returns Kubeless ConfigMap
func GetKubelessConfig(cli kubernetes.Interface, cliAPIExtensions clientsetAPIExtensions.Interface) (*v1.ConfigMap, error) {
	configLocation, err := getConfigLocation(cliAPIExtensions)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching config location: %v", err)
	}
	controllerNamespace := configLocation.Namespace
	kubelessConfig := configLocation.Name
	config, err := cli.CoreV1().ConfigMaps(controllerNamespace).Get(kubelessConfig, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to read the configmap: %s", err)
	}
	return config, nil
}

// DryRunFmt stringify the given interface in a specific format
func DryRunFmt(format string, trigger interface{}) (string, error) {
	switch format {
	case "json":
		j, err := json.MarshalIndent(trigger, "", "    ")
		if err != nil {
			return "", err
		}
		return string(j[:]), nil
	case "yaml":
		y, err := yaml.Marshal(trigger)
		if err != nil {
			return "", err
		}
		return string(y[:]), nil
	default:
		return "", fmt.Errorf("Output format needs to be yaml or json")
	}
}
