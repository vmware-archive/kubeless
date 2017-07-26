/*
Copyright 2016 Skippbox, Ltd.

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
	"io/ioutil"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func constructFunction(runtime, handler, file, funcName, funcType, topic, ns, deps, description string, mem resource.Quantity, labels, envs []string) *spec.Function {
	funcLabels := map[string]string{}
	for _, label := range labels {
		k, v := getKV(label)
		funcLabels[k] = v
	}

	funcEnv := map[string]string{}
	for _, env := range envs {
		k, v := getKV(env)
		funcEnv[k] = v
	}

	funcAnnotation := map[string]string{
		"Description": description,
	}

	f := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:        funcName,
			Namespace:   ns,
			Labels:      funcLabels,
			Annotations: funcAnnotation,
		},
		Spec: spec.FunctionSpec{
			Handler:  handler,
			Runtime:  runtime,
			Type:     funcType,
			Function: readFile(file),
			Topic:    topic,
			Env:      funcEnv,
			Memory:   mem,
		},
	}
	// add dependencies file to func spec
	if deps != "" {
		f.Spec.Deps = readFile(deps)
	}

	return f
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

func readFile(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Fatalf("Can not read file: %s. The file may not exist", file)
	}
	return string(data[:])
}
