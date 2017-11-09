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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"
	// "github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	// "k8s.io/client-go/rest"
)

var functionCmd = &cobra.Command{
	Use:   "function SUBCOMMAND",
	Short: "function specific operations",
	Long:  `function command allows user to list, deploy, edit, delete functions running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	functionCmd.AddCommand(deployCmd)
	functionCmd.AddCommand(deleteCmd)
	functionCmd.AddCommand(listCmd)
	functionCmd.AddCommand(callCmd)
	functionCmd.AddCommand(logsCmd)
	functionCmd.AddCommand(describeCmd)
	functionCmd.AddCommand(updateCmd)
	functionCmd.AddCommand(autoscaleCmd)
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

func readFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data[:]), nil
}

func parseLabel(labels []string) map[string]string {
	funcLabels := map[string]string{}
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

func parseMemory(mem string) (resource.Quantity, error) {
	quantity, err := resource.ParseQuantity(mem)
	if err != nil {
		return resource.Quantity{}, err
	}

	return quantity, nil
}

func getFunctionDescription(funcName, ns, handler, funcContent, deps, runtime, topic, schedule, runtimeImage, mem string, triggerHTTP bool, envs, labels []string, defaultFunction spec.Function) (f *spec.Function, err error) {

	if handler == "" {
		handler = defaultFunction.Spec.Handler
	}

	if funcContent == "" {
		funcContent = defaultFunction.Spec.Function
	}

	if deps == "" {
		deps = defaultFunction.Spec.Deps
	}

	if runtime == "" {
		runtime = defaultFunction.Spec.Runtime
	}

	funcType := ""
	switch {
	case triggerHTTP:
		funcType = "HTTP"
		topic = ""
		schedule = ""
		break
	case schedule != "":
		funcType = "Scheduled"
		topic = ""
		break
	case topic != "":
		funcType = "PubSub"
		schedule = ""
		break
	default:
		funcType = defaultFunction.Spec.Type
		topic = defaultFunction.Spec.Topic
		schedule = defaultFunction.Spec.Schedule
	}

	funcEnv := parseEnv(envs)
	if len(funcEnv) == 0 && len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
		funcEnv = defaultFunction.Spec.Template.Spec.Containers[0].Env
	}

	funcLabels := parseLabel(labels)
	if len(funcLabels) == 0 {
		funcLabels = defaultFunction.Metadata.Labels
	}

	funcMem := resource.Quantity{}
	resources := v1.ResourceRequirements{}
	if mem != "" {
		funcMem, err = parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("Wrong format of the memory value: %v", err)
			return
		}
		resource := map[v1.ResourceName]resource.Quantity{
			v1.ResourceMemory: funcMem,
		}
		resources = v1.ResourceRequirements{
			Limits:   resource,
			Requests: resource,
		}
	} else {
		if len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
			resources = defaultFunction.Spec.Template.Spec.Containers[0].Resources
		}
	}

	if len(runtimeImage) == 0 && len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
		runtimeImage = defaultFunction.Spec.Template.Spec.Containers[0].Image
	}

	f = &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler:  handler,
			Runtime:  runtime,
			Type:     funcType,
			Function: funcContent,
			Deps:     deps,
			Topic:    topic,
			Schedule: schedule,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env:       funcEnv,
							Resources: resources,
							Image:     runtimeImage,
						},
					},
				},
			},
		},
	}
	return
}

// Returns the maximum timeout that we will wait for a deployment
// Defaults: 10 min
func getDeploymentTimeout() (int, error) {
	envTimeout := os.Getenv("KUBELESS_DEPLOYMENT_TIMEOUT")
	if envTimeout == "" {
		envTimeout = "600"
	}
	timeout, err := strconv.Atoi(envTimeout)
	if err != nil {
		return 0, fmt.Errorf("Wrong value for KUBELESS_DEPLOYMENT_TIMEOUT: %s", err)
	}
	return timeout, nil
}

func getContainerStatus(c v1.ContainerStatus) string {
	if c.State.Terminated != nil {
		return c.State.Terminated.Reason
	} else if c.State.Waiting != nil {
		return c.State.Waiting.Reason
	} else if c.State.Running != nil {
		return "Running"
	}
	return ""
}

func getPodStatus(pod v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	if len(pod.Status.ContainerStatuses) > 0 {
		return getContainerStatus(pod.Status.ContainerStatuses[0])
	} else if len(pod.Status.InitContainerStatuses) > 0 {
		return "Init:" + getContainerStatus(pod.Status.InitContainerStatuses[0])
	}
	// If the pod doesn't have containers it is a temporary Pod
	return ""
}

func getPodList(cli kubernetes.Interface, funcName, ns string) ([]v1.Pod, error) {
	pods, err := cli.CoreV1().Pods(ns).List(metav1.ListOptions{
		LabelSelector: "function=" + funcName,
	})
	if err != nil {
		return []v1.Pod{}, err
	}
	return pods.Items, nil
}

func printDeploymentLogs(cli kubernetes.Interface, funcName, ns string) error {
	podList, err := getPodList(cli, funcName, ns)
	if err != nil {
		return err
	}
	for _, pod := range podList {
		podStatus := getPodStatus(pod)
		// Retrieving logs of useful pods
		if podStatus != "Unknown" && podStatus != "Terminating" {
			resBytes, err := cli.CoreV1().RESTClient().Get().Namespace(ns).Name(pod.Name).Resource("pods").SubResource("log").Do().Raw()
			if err == nil {
				logrus.Errorf("Logs from %s:\n%s", pod.Name, string(resBytes))
			}
		}
	}
	return nil
}

func waitForDeployment(cli kubernetes.Interface, funcName, ns string, timeout int) error {
	successCount := 0
	retriesCount := 0
	logrus.Info("Deployment status:")
	var previousStatus, currentStatus string
	err := wait.Poll(time.Duration(time.Second), time.Duration(timeout)*time.Second, func() (bool, error) {
		dpm, err := cli.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
		ready := false
		if err != nil {
			// Deployment not available
			retriesCount++
			if retriesCount >= 3 {
				// The deployment may not be ready yet if the controller
				// is still processing the function so we retry 3 times
				// (if it takes more than that we assume there has been an error)
				return false, fmt.Errorf("Unable to find a deployment for %s", funcName)
			}
		} else {
			currentStatus = fmt.Sprintf("Ready replicas: %d/%d", dpm.Status.ReadyReplicas, dpm.Status.Replicas)
			if currentStatus != previousStatus {
				previousStatus = currentStatus
				logrus.Info(currentStatus)
			}
			if dpm.Status.Replicas == dpm.Status.ReadyReplicas {
				// Pods may be running for a short amount of time
				// So we ensure that it keeps running for at least 5 seconds
				successCount++
			} else {
				successCount = 0
			}
			if successCount == 5 {
				logrus.Infof("Function %s is ready", funcName)
				ready = true
			}
		}
		return ready, nil
	})
	if err != nil {
		printDeploymentLogs(cli, funcName, ns)
	}
	return err
}
