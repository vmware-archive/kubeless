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
	cronjobApi "github.com/kubeless/cronjob-trigger/pkg/apis/kubeless/v1beta1"
	cronjobUtils "github.com/kubeless/cronjob-trigger/pkg/utils"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <function_name> FLAG",
	Short: "deploy a function to Kubeless",
	Long:  `deploy a function to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := kubelessUtils.GetClientOutOfCluster()
		apiExtensionsClientset := kubelessUtils.GetAPIExtensionsClientOutOfCluster()
		config, err := kubelessUtils.GetKubelessConfig(cli, apiExtensionsClientset)
		if err != nil {
			logrus.Fatalf("Unable to read the configmap: %v", err)
		}

		var lr = langruntime.New(config)
		lr.ReadConfigMap()

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		schedule, err := cmd.Flags().GetString("schedule")
		if err != nil {
			logrus.Fatal(err)
		}

		if schedule != "" {
			if _, err := cron.ParseStandard(schedule); err != nil {
				logrus.Fatalf("Invalid value for --schedule. " + err.Error())
			}
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

		if runtime != "" && !lr.IsValidRuntime(runtime) {
			logrus.Fatalf("Invalid runtime: %s. Supported runtimes are: %s",
				runtime, strings.Join(lr.GetRuntimes(), ", "))
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
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		deps, err := cmd.Flags().GetString("dependencies")
		if err != nil {
			logrus.Fatal(err)
		}

		secrets, err := cmd.Flags().GetStringSlice("secrets")
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

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		headless, err := cmd.Flags().GetBool("headless")
		if err != nil {
			logrus.Fatal(err)
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
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

		if runtime == "" && runtimeImage == "" {
			logrus.Fatal("Either `--runtime` or `--runtime-image` flag must be specified.")
		}

		if runtime != "" && handler == "" {
			logrus.Fatal("You must specify handler for the runtime.")
		}

		defaultFunctionSpec := kubelessApi.Function{}
		defaultFunctionSpec.ObjectMeta.Labels = map[string]string{
			"created-by": "kubeless",
			"function":   funcName,
		}

		f, err := getFunctionDescription(funcName, ns, handler, file, funcDeps, runtime, runtimeImage, mem, cpu, timeout, imagePullPolicy, port, headless, envs, labels, secrets, defaultFunctionSpec)
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

		kubelessClient, err := kubelessUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Deploying function...")
		err = kubelessUtils.CreateFunctionCustomResource(kubelessClient, f)
		if err != nil {
			logrus.Fatalf("Failed to deploy %s. Received:\n%s", funcName, err)
		}
		logrus.Infof("Function %s submitted for deployment", funcName)
		logrus.Infof("Check the deployment status executing 'kubeless function ls %s'", funcName)

		if schedule != "" {
			cronJobTrigger := cronjobApi.CronJobTrigger{}
			cronJobTrigger.TypeMeta = metav1.TypeMeta{
				Kind:       "CronJobTrigger",
				APIVersion: "kubeless.io/v1beta1",
			}
			cronJobTrigger.ObjectMeta = metav1.ObjectMeta{
				Name:      funcName,
				Namespace: ns,
			}
			cronJobTrigger.ObjectMeta.Labels = map[string]string{
				"created-by": "kubeless",
				"function":   funcName,
			}
			cronJobTrigger.Spec.FunctionName = funcName
			cronJobTrigger.Spec.Schedule = schedule
			cronjobClient, err := cronjobUtils.GetKubelessClientOutCluster()
			if err != nil {
				logrus.Fatal(err)
			}
			err = cronjobUtils.CreateCronJobCustomResource(cronjobClient, &cronJobTrigger)
			if err != nil {
				logrus.Fatalf("Failed to deploy cron job trigger %s. Received:\n%s", funcName, err)
			}
		}
	},
}

func init() {
	deployCmd.Flags().StringP("runtime", "r", "", "Specify runtime")
	deployCmd.Flags().StringP("handler", "", "", "Specify handler")
	deployCmd.Flags().StringP("from-file", "f", "", "Specify code file or a URL to the code file")
	deployCmd.Flags().StringSliceP("label", "l", []string{}, "Specify labels of the function. Both separator ':' and '=' are allowed. For example: --label foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringSliceP("secrets", "", []string{}, "Specify Secrets to be mounted to the functions container. For example: --secrets mySecret")
	deployCmd.Flags().StringSliceP("env", "e", []string{}, "Specify environment variable of the function. Both separator ':' and '=' are allowed. For example: --env foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	deployCmd.Flags().StringP("dependencies", "d", "", "Specify a file containing list of dependencies for the function")
	deployCmd.Flags().StringP("schedule", "", "", "Specify schedule in cron format for scheduled function")
	deployCmd.Flags().StringP("memory", "", "", "Request amount of memory, which is measured in bytes, for the function. It is expressed as a plain integer or a fixed-point interger with one of these suffies: E, P, T, G, M, K, Ei, Pi, Ti, Gi, Mi, Ki")
	deployCmd.Flags().StringP("cpu", "", "", "Request amount of cpu for the function, which is measured in units of cores. Please see the following link for more information: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu")
	deployCmd.Flags().StringP("runtime-image", "", "", "Custom runtime image")
	deployCmd.Flags().StringP("image-pull-policy", "", "Always", "Image pull policy")
	deployCmd.Flags().StringP("timeout", "", "180", "Maximum timeout (in seconds) for the function to complete its execution")
	deployCmd.Flags().StringP("output", "o", "yaml", "Output format")
	deployCmd.Flags().Bool("headless", false, "Deploy http-based function without a single service IP and load balancing support from Kubernetes. See: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services")
	deployCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	deployCmd.Flags().Int32("port", 8080, "Deploy http-based function with a custom port")
}
