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
	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

var updateCmd = &cobra.Command{
	Use:   "update <function_name> FLAG",
	Short: "update a function on Kubeless",
	Long:  `update a function on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		handler, err := cmd.Flags().GetString("handler")
		if err != nil {
			logrus.Fatal(err)
		}

		file, err := cmd.Flags().GetString("from-file")
		if err != nil {
			logrus.Fatal(err)
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		runtime, err := cmd.Flags().GetString("runtime")
		if err != nil {
			logrus.Fatal(err)
		}

		labels, err := cmd.Flags().GetStringSlice("label")
		if err != nil {
			logrus.Fatal(err)
		}

		envs, err := cmd.Flags().GetStringSlice("env")
		if err != nil {
			logrus.Fatal(err)
		}

		runtimeImage, err := cmd.Flags().GetString("runtime-image")
		if err != nil {
			logrus.Fatal(err)
		}

		mem, err := cmd.Flags().GetString("memory")
		if err != nil {
			logrus.Fatal(err)
		}

		funcType := "HTTP"
		funcLabels := parseLabel(labels)
		funcEnv := parseEnv(envs)
		funcMem := resource.Quantity{}
		if mem != "" {
			funcMem, err = parseMemory(mem)
			if err != nil {
				logrus.Fatalf("Wrong format of the memory value: %v", err)
			}
		}
		funcContent := ""
		if len(file) != 0 {
			funcContent, err = readFile(file)
			if err != nil {
				logrus.Fatalf("Unable to read file %s: %v", file, err)
			}
		} else {
			if len(runtimeImage) == 0 {
				logrus.Fatalf("You should specify either a file containing your function or a runtime image")
			}
		}

		resource := map[v1.ResourceName]resource.Quantity{
			v1.ResourceMemory: funcMem,
		}

		f := &spec.Function{
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
				Topic:    "",
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Env: funcEnv,
								Resources: v1.ResourceRequirements{
									Limits:   resource,
									Requests: resource,
								},
							},
						},
					},
				},
			},
		}

		if len(runtimeImage) != 0 {
			f.Spec.Template.Spec.Containers[0].Image = runtimeImage
		}

		err = utils.UpdateK8sCustomResource(f)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	updateCmd.Flags().StringP("runtime", "", "", "Specify runtime")
	updateCmd.Flags().StringP("handler", "", "", "Specify handler")
	updateCmd.Flags().StringP("from-file", "", "", "Specify code file")
	updateCmd.Flags().StringP("memory", "", "", "Request amount of memory for the function")
	updateCmd.Flags().StringSliceP("label", "", []string{}, "Specify labels of the function")
	updateCmd.Flags().StringSliceP("env", "", []string{}, "Specify environment variable of the function")
	updateCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the function")
	updateCmd.Flags().StringP("runtime-image", "", "", "Custom runtime image")
}
