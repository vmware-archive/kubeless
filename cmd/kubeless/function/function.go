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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

//FunctionCmd contains first-class command for function
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

func parseMemory(mem string) (resource.Quantity, error) {
	quantity, err := resource.ParseQuantity(mem)
	if err != nil {
		return resource.Quantity{}, err
	}

	return quantity, nil
}

func getFileSha256(file string) (string, error) {
	h := sha256.New()
	ff, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer ff.Close()
	_, err = io.Copy(h, ff)
	if err != nil {
		return "", err
	}
	checksum := hex.EncodeToString(h.Sum(nil))
	return "sha256:" + checksum, err
}

func getContentType(filename string, fbytes []byte) string {
	var contentType string
	isText := utf8.ValidString(string(fbytes))
	if isText {
		contentType = "text"
	} else {
		contentType = "base64"
		if path.Ext(filename) == ".zip" {
			contentType += "+zip"
		}
	}
	return contentType
}

<<<<<<< HEAD
func getFunctionDescription(cli kubernetes.Interface, funcName, ns, handler, file, deps, runtime, topic, schedule, runtimeImage, mem, timeout string, triggerHTTP bool, headlessFlag *bool, portFlag *int32, envs, labels []string, defaultFunction api.Function) (*api.Function, error) {
	function := defaultFunction
	function.TypeMeta = metav1.TypeMeta{
		Kind:       "Function",
		APIVersion: "k8s.io/v1",
	}
	if handler != "" {
		function.Spec.Handler = handler
=======
func getFunctionDescription(cli kubernetes.Interface, funcName, ns, handler, file, deps, runtime, topic, schedule, runtimeImage, mem, timeout string, triggerHTTP bool, headlessFlag *bool, portFlag *int32, envs, labels []string, defaultFunction kubelessApi.Function) (*kubelessApi.Function, error) {

	if handler == "" {
		handler = defaultFunction.Spec.Handler
>>>>>>> rename package alias from api to kubelessApi
	}

	if file != "" {
		functionBytes, err := ioutil.ReadFile(file)
		if err != nil {
<<<<<<< HEAD
			return nil, err
=======
			return &kubelessApi.Function{}, err
>>>>>>> rename package alias from api to kubelessApi
		}
		function.Spec.FunctionContentType = getContentType(file, functionBytes)
		if function.Spec.FunctionContentType == "text" {
			function.Spec.Function = string(functionBytes)
		} else {
			function.Spec.Function = base64.StdEncoding.EncodeToString(functionBytes)
		}
		function.Spec.Checksum, err = getFileSha256(file)
		if err != nil {
<<<<<<< HEAD
			return nil, err
=======
			return &kubelessApi.Function{}, err
>>>>>>> rename package alias from api to kubelessApi
		}
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

	triggers := []bool{triggerHTTP, topic != "", schedule != ""}
	triggerCount := 0
	for i := len(triggers) - 1; i >= 0; i-- {
		if triggers[i] {
			triggerCount++
		}
	}
	if triggerCount > 1 {
		return nil, errors.New("exactly one of --trigger-http, --trigger-topic, --schedule must be specified")
	}

	switch {
	case triggerHTTP:
		function.Spec.Type = "HTTP"
		function.Spec.Topic = ""
		function.Spec.Schedule = ""
		break
	case schedule != "":
		function.Spec.Type = "Scheduled"
		function.Spec.Schedule = schedule
		function.Spec.Topic = ""
		break
	case topic != "":
		function.Spec.Type = "PubSub"
		function.Spec.Topic = topic
		function.Spec.Schedule = ""
		break
	}

	funcEnv := parseEnv(envs)
	if len(funcEnv) == 0 && len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
		funcEnv = defaultFunction.Spec.Template.Spec.Containers[0].Env
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
	if mem != "" {
		funcMem, err := parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("Wrong format of the memory value: %v", err)
			return &kubelessApi.Function{}, err
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
	function.Spec.Template.Spec.Containers = []v1.Container{
		{
			Env:       funcEnv,
			Resources: resources,
			Image:     runtimeImage,
		},
	}

	selectorLabels := map[string]string{}
	for k, v := range funcLabels {
		selectorLabels[k] = v
	}
	selectorLabels["function"] = funcName

	svcSpec := v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:     "function-port",
				NodePort: 0,
				Protocol: v1.ProtocolTCP,
			},
		},
		Selector: selectorLabels,
		Type:     v1.ServiceTypeClusterIP,
	}

	if headlessFlag != nil {
		if *headlessFlag == true {
			svcSpec.ClusterIP = v1.ClusterIPNone
		}
	} else {
		svcSpec.ClusterIP = defaultFunction.Spec.ServiceSpec.ClusterIP
	}

	if portFlag != nil {
		svcSpec.Ports[0].Port = *portFlag
		svcSpec.Ports[0].TargetPort = intstr.FromInt(int(*portFlag))
	} else {
		svcSpec.Ports[0].Port = defaultFunction.Spec.ServiceSpec.Ports[0].Port
		svcSpec.Ports[0].TargetPort = defaultFunction.Spec.ServiceSpec.Ports[0].TargetPort
	}
	function.Spec.ServiceSpec = svcSpec

	return &function, nil
}

func getDeploymentStatus(cli kubernetes.Interface, funcName, ns string) (string, error) {
	dpm, err := cli.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
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
