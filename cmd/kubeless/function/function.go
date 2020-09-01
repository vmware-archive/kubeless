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

package function

import (
	"fmt"
	"strings"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessutil "github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// FunctionCmd contains first-class command for function
var FunctionCmd = &cobra.Command{
	Use:   "function SUBCOMMAND",
	Short: "function specific operations",
	Long:  `function command allows user to list, deploy, edit, delete functions running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	FunctionCmd.AddCommand(deployCmd)
	FunctionCmd.AddCommand(deleteCmd)
	FunctionCmd.AddCommand(listCmd)
	FunctionCmd.AddCommand(callCmd)
	FunctionCmd.AddCommand(logsCmd)
	FunctionCmd.AddCommand(describeCmd)
	FunctionCmd.AddCommand(updateCmd)
	FunctionCmd.AddCommand(topCmd)
}

func getKV(input string) (string, string) {
	var key, value string
	if pos := strings.IndexAny(input, "=:"); pos != -1 {
		key = input[:pos]
		value = input[pos+1:]
	} else {
		// no separator found
		key = input
		value = ""
	}

	return key, value
}

func parseLabel(labels []string) map[string]string {
	funcLabels := make(map[string]string)
	for _, label := range labels {
		k, v := getKV(label)
		funcLabels[k] = v
	}
	return funcLabels
}

func parseEnv(envs []string) []v1.EnvVar {
	funcEnv := []v1.EnvVar{}
	for _, env := range envs {
		k, v := getKV(env)
		funcEnv = append(funcEnv, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return funcEnv
}

func parseResource(in string) (resource.Quantity, error) {
	if in == "" {
		return resource.Quantity{}, nil
	}

	quantity, err := resource.ParseQuantity(in)
	if err != nil {
		return resource.Quantity{}, err
	}

	return quantity, nil
}

func parseNodeSelectors(nodeSelectors []string) map[string]string {
	funcNodeSelectors := make(map[string]string)
	for _, nodeSelector := range nodeSelectors {
		k, v := getKV(nodeSelector)
		funcNodeSelectors[k] = v
	}
	return funcNodeSelectors
}

func getFunctionDescription(funcName, ns, handler, file, deps, runtime, runtimeImage, mem, cpu, timeout string, imagePullPolicy string, serviceAccount string, port int32, servicePort int32, headless bool, envs, labels, secrets, nodeSelectors []string, defaultFunction kubelessApi.Function) (*kubelessApi.Function, error) {
	function := defaultFunction
	function.TypeMeta = metav1.TypeMeta{
		Kind:       "Function",
		APIVersion: "kubeless.io/v1beta1",
	}
	if handler != "" {
		function.Spec.Handler = handler
	}

	if file != "" {
		contentType, err := kubelessutil.GetContentType(file)
		if err != nil {
			return nil, err
		}
		functionContent, checksum, err := kubelessutil.ParseContent(file, contentType)
		if err != nil {
			return nil, err
		}
		if strings.Contains(contentType, "url") {
			// set the function to be the URL provided on the command line
			function.Spec.Function = file
		} else {
			function.Spec.Function = functionContent
		}
		function.Spec.Checksum = checksum
		function.Spec.FunctionContentType = contentType
	}

	if deps != "" {
		function.Spec.Deps = deps
	}

	if runtime != "" {
		function.Spec.Runtime = runtime
	}

	if timeout != "" {
		function.Spec.Timeout = timeout
	}

	funcEnv := parseEnv(envs)
	if len(funcEnv) == 0 && len(defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers) != 0 {
		funcEnv = defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers[0].Env
	}

	funcLabels := defaultFunction.ObjectMeta.Labels
	if len(funcLabels) == 0 {
		funcLabels = make(map[string]string)
	}
	ls := parseLabel(labels)
	for k, v := range ls {
		funcLabels[k] = v
	}
	function.ObjectMeta = metav1.ObjectMeta{
		Name:      funcName,
		Namespace: ns,
		Labels:    funcLabels,
	}

	resources := v1.ResourceRequirements{}
	if mem != "" || cpu != "" {
		funcMem, err := parseResource(mem)
		if err != nil {
			err = fmt.Errorf("Wrong format of the memory value: %v", err)
			return &kubelessApi.Function{}, err
		}
		funcCPU, err := parseResource(cpu)
		if err != nil {
			err = fmt.Errorf("Wrong format for cpu value: %v", err)
			return &kubelessApi.Function{}, err
		}
		resource := map[v1.ResourceName]resource.Quantity{
			v1.ResourceMemory: funcMem,
			v1.ResourceCPU:    funcCPU,
		}

		resources = v1.ResourceRequirements{
			Limits:   resource,
			Requests: resource,
		}
	} else {
		if len(defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers) != 0 {
			resources = defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers[0].Resources
		}
	}

	if len(runtimeImage) == 0 && len(defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers) != 0 {
		runtimeImage = defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers[0].Image
	}
	function.Spec.Deployment.Spec.Template.Spec.Containers = []v1.Container{
		{
			ImagePullPolicy: v1.PullPolicy(imagePullPolicy),
			Env:             funcEnv,
			Resources:       resources,
			Image:           runtimeImage,
		},
	}

	if serviceAccount != "" {
		function.Spec.Deployment.Spec.Template.Spec.ServiceAccountName = serviceAccount
	}

	if len(defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers) != 0 {
		function.Spec.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts = defaultFunction.Spec.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts
	}

	svcSpec := v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:     "http-function-port",
				NodePort: 0,
				Protocol: v1.ProtocolTCP,
			},
		},
		Selector: funcLabels,
		Type:     v1.ServiceTypeClusterIP,
	}

	if headless {
		svcSpec.ClusterIP = v1.ClusterIPNone
	}

	if port != 0 {
		svcSpec.Ports[0].Port = port
		svcSpec.Ports[0].TargetPort = intstr.FromInt(int(port))
	}
	if servicePort != 0 {
		svcSpec.Ports[0].Port = servicePort
	}
	function.Spec.ServiceSpec = svcSpec

	for _, secret := range secrets {
		function.Spec.Deployment.Spec.Template.Spec.Volumes = append(function.Spec.Deployment.Spec.Template.Spec.Volumes, v1.Volume{
			Name: secret + "-vol",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
				},
			},
		})
		function.Spec.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(function.Spec.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      secret + "-vol",
			MountPath: "/" + secret,
		})

	}

	funcNodeSelectors := parseNodeSelectors(nodeSelectors)
	if len(funcNodeSelectors) == 0 && len(defaultFunction.Spec.Deployment.Spec.Template.Spec.NodeSelector) != 0 {
		funcNodeSelectors = defaultFunction.Spec.Deployment.Spec.Template.Spec.NodeSelector
	}
	function.Spec.Deployment.Spec.Template.Spec.NodeSelector = funcNodeSelectors

	return &function, nil
}

func getDeploymentStatus(cli kubernetes.Interface, funcName, ns string) (string, error) {
	dpm, err := cli.AppsV1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	status := fmt.Sprintf("%d/%d", dpm.Status.ReadyReplicas, dpm.Status.Replicas)
	if dpm.Status.ReadyReplicas > 0 {
		status += " READY"
	} else {
		status += " NOT READY"
	}
	return status, nil
}

func getFunctions(kubelessClient versioned.Interface, namespace, functionName string) ([]*kubelessApi.Function, error) {
	if functionName == "" {
		f, err := kubelessClient.KubelessV1beta1().Functions(namespace).List(metav1.ListOptions{})
		if err != nil {
			return []*kubelessApi.Function{}, err
		}
		return f.Items, nil
	}

	f, err := kubelessClient.KubelessV1beta1().Functions(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		return []*kubelessApi.Function{}, err
	}
	return []*kubelessApi.Function{
		f,
	}, nil
}
