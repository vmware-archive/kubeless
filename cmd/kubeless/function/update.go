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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <function_name> FLAG",
	Short: "update a function on Kubeless",
	Long:  `update a function on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetClientOutOfCluster()
		apiExtensionsClientset := utils.GetAPIExtensionsClientOutOfCluster()
		config, err := utils.GetKubelessConfig(cli, apiExtensionsClientset)
		if err != nil {
			logrus.Fatalf("Unable to read the configmap: %v", err)
		}

		var lr = langruntime.New(config)
		lr.ReadConfigMap()

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		handler, err := cmd.Flags().GetString("handler")
		if err != nil {
			logrus.Fatal(err)
		}

		file, err := cmd.Flags().GetString("from-file")
		if err != nil {
			logrus.Fatal(err)
		}

		secrets, err := cmd.Flags().GetStringSlice("secrets")
		if err != nil {
			logrus.Fatal(err)
		}

		runtime, err := cmd.Flags().GetString("runtime")
		if err != nil {
			logrus.Fatal(err)
		}

		if runtime != "" && !lr.IsValidRuntime(runtime) {
			logrus.Fatalf("Invalid runtime: %s. Supported runtimes are: %s",
				runtime, strings.Join(lr.GetRuntimes(), ", "))
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

		imagePullPolicy, err := cmd.Flags().GetString("image-pull-policy")
		if err != nil {
			logrus.Fatal(err)
		}

		if imagePullPolicy != "IfNotPresent" && imagePullPolicy != "Always" && imagePullPolicy != "Never" {
			err := fmt.Errorf("image-pull-policy must be {IfNotPresent|Always|Never}")
			logrus.Fatal(err)
		}

		mem, err := cmd.Flags().GetString("memory")
		if err != nil {
			logrus.Fatal(err)
		}

		cpu, err := cmd.Flags().GetString("cpu")
		if err != nil {
			logrus.Fatal(err)
		}

		timeout, err := cmd.Flags().GetString("timeout")
		if err != nil {
			logrus.Fatal(err)
		}

		deps, err := cmd.Flags().GetString("dependencies")
		if err != nil {
			logrus.Fatal(err)
		}
		funcDeps := ""
		if deps != "" {
			contentType, err := getContentType(deps)
			if err != nil {
				logrus.Fatal(err)
			}
			funcDeps, _, err = parseContent(deps, contentType)
			if err != nil {
				logrus.Fatal(err)
			}
		}
		headless, err := cmd.Flags().GetBool("headless")
		if err != nil {
			logrus.Fatal(err)
		}
		port, err := cmd.Flags().GetInt32("port")
		if err != nil {
			logrus.Fatal(err)
		}
		if port <= 0 || port > 65535 {
			logrus.Fatalf("Invalid port number %d specified", port)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		previousFunction, err := utils.GetFunction(funcName, ns)
		if err != nil {
			logrus.Fatal(err)
		}

		f, err := getFunctionDescription(funcName, ns, handler, file, funcDeps, runtime, runtimeImage, mem, cpu, timeout, imagePullPolicy, port, headless, envs, labels, secrets, previousFunction)
		if err != nil {
			logrus.Fatal(err)
		}

		if dryrun == true {
			if output == "json" {
				j, err := json.MarshalIndent(f, "", "    ")
				if err != nil {
					logrus.Fatal(err)
				}
				fmt.Println(string(j[:]))
				return
			} else if output == "yaml" {
				y, err := yaml.Marshal(f)
				if err != nil {
					logrus.Fatal(err)
				}
				fmt.Println(string(y[:]))
				return
			} else {
				logrus.Infof("Output format needs to be yaml or json")
				return
			}
		}

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Redeploying function...")
		err = utils.PatchFunctionCustomResource(kubelessClient, f)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Function %s submitted for deployment", funcName)
		logrus.Infof("Check the deployment status executing 'kubeless function ls %s'", funcName)
	},
}

func init() {
	updateCmd.Flags().StringP("runtime", "r", "", "Specify runtime")
	updateCmd.Flags().StringP("handler", "", "", "Specify handler")
	updateCmd.Flags().StringP("from-file", "f", "", "Specify code file or a URL to the code file")
	updateCmd.Flags().StringP("memory", "", "", "Request amount of memory for the function")
	updateCmd.Flags().StringP("cpu", "", "", "Request amount of cpu for the function.")
	updateCmd.Flags().StringSliceP("label", "l", []string{}, "Specify labels of the function")
	updateCmd.Flags().StringSliceP("secrets", "", []string{}, "Specify Secrets to be mounted to the functions container. For example: --secrets mySecret")
	updateCmd.Flags().StringSliceP("env", "e", []string{}, "Specify environment variable of the function")
	updateCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	updateCmd.Flags().StringP("dependencies", "d", "", "Specify a file containing list of dependencies for the function")
	updateCmd.Flags().StringP("runtime-image", "", "", "Custom runtime image")
	updateCmd.Flags().StringP("image-pull-policy", "", "Always", "Image pull policy")
	updateCmd.Flags().StringP("timeout", "", "180", "Maximum timeout (in seconds) for the function to complete its execution")
	updateCmd.Flags().Bool("headless", false, "Deploy http-based function without a single service IP and load balancing support from Kubernetes. See: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services")
	updateCmd.Flags().Int32("port", 8080, "Deploy http-based function with a custom port")
	updateCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	updateCmd.Flags().StringP("output", "o", "yaml", "Output format")

}
