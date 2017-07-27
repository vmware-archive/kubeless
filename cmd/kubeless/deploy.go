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
	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <function_name> FLAG",
	Short: "deploy a function to Kubeless",
	Long:  `deploy a function to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		triggerHTTP, err := cmd.Flags().GetBool("trigger-http")
		if err != nil {
			logrus.Fatal(err)
		}

		topic, err := cmd.Flags().GetString("trigger-topic")
		if err != nil {
			logrus.Fatal(err)
		}

		description, err := cmd.Flags().GetString("description")
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

		runtime, err := cmd.Flags().GetString("runtime")
		if err != nil {
			logrus.Fatal(err)
		}

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

		deps, err := cmd.Flags().GetString("dependencies")
		if err != nil {
			logrus.Fatal(err)
		}

		mem, err := cmd.Flags().GetString("memory")
		if err != nil {
			logrus.Fatal(err)
		}
		funcMem := resource.Quantity{}
		if mem != "" {
			funcMem = parseMemory(mem)
		}

		funcType := "PubSub"
		if triggerHTTP {
			funcType = "HTTP"
			topic = ""
		}

		f := constructFunction(runtime, handler, file, funcName, funcType, topic, ns, deps, description, funcMem, labels, envs)
		err = utils.CreateK8sCustomResource(f)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	deployCmd.Flags().StringP("runtime", "", "", "Specify runtime")
	deployCmd.Flags().StringP("handler", "", "", "Specify handler")
	deployCmd.Flags().StringP("from-file", "", "", "Specify code file")
	deployCmd.Flags().StringP("description", "", "", "Specify description of the function")
	deployCmd.Flags().StringSliceP("label", "", []string{}, "Specify labels of the function. Both separator ':' and '=' are allowed. For example: --label foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringSliceP("env", "", []string{}, "Specify environment variable of the function. Both separator ':' and '=' are allowed. For example: --env foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the function")
	deployCmd.Flags().StringP("dependencies", "", "", "Specify a file containing list of dependencies for the function")
	deployCmd.Flags().StringP("trigger-topic", "", "kubeless", "Deploy a pubsub function to Kubeless")
	deployCmd.Flags().StringP("memory", "", "", "Request amount of memory, which is measured in bytes, for the function. It is expressed as a plain interger or a fixed-point interger with one of these suffies: E, P, T, G, M, K, Ei, Pi, Ti, Gi, Mi, Ki")
	deployCmd.Flags().Bool("trigger-http", false, "Deploy a http-based function to Kubeless")
}

func parseMemory(mem string) resource.Quantity {
	quantity, err := resource.ParseQuantity(mem)
	if err != nil {
		logrus.Fatalf("Wrong format of the memory value: %v", err)
	}

	return quantity
}
