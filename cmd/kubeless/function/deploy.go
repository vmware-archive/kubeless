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
	"io/ioutil"
	"os"
	"strings"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <function_name> FLAG",
	Short: "deploy a function to Kubeless",
	Long:  `deploy a function to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetClientOutOfCluster()
		controllerNamespace := os.Getenv("KUBELESS_NAMESPACE")
		kubelessConfig := os.Getenv("KUBELESS_CONFIG")

		if len(controllerNamespace) == 0 {
			controllerNamespace = "kubeless"
		}

		if len(kubelessConfig) == 0 {
			kubelessConfig = "kubeless-config"
		}
		config, err := cli.CoreV1().ConfigMaps(controllerNamespace).Get(kubelessConfig, metav1.GetOptions{})
		if err != nil {
			logrus.Fatalf("Unable to read the configmap: %v", err)
		}

		var lr = langruntime.New(config)
		lr.ReadConfigMap()

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		triggerHTTP, err := cmd.Flags().GetBool("trigger-http")
		if err != nil {
			logrus.Fatal(err)
		}

		schedule, err := cmd.Flags().GetString("schedule")
		if err != nil {
			logrus.Fatal(err)
		}

		if schedule != "" {
			if _, err := cron.ParseStandard(schedule); err != nil {
				logrus.Fatalf("Invalid value for --schedule. " + err.Error())
			}
		}

		topic, err := cmd.Flags().GetString("trigger-topic")
		if err != nil {
			logrus.Fatal(err)
		}

		labels, err := cmd.Flags().GetStringSlice("label")
		if err != nil {
			logrus.Fatal(err)
		}

		envs, err := cmd.Flags().GetStringArray("env")
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
			ns = utils.GetDefaultNamespace()
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

		mem, err := cmd.Flags().GetString("memory")
		if err != nil {
			logrus.Fatal(err)
		}

		timeout, err := cmd.Flags().GetString("timeout")
		if err != nil {
			logrus.Fatal(err)
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

		funcDeps := ""
		if deps != "" {
			bytes, err := ioutil.ReadFile(deps)
			if err != nil {
				logrus.Fatalf("Unable to read file %s: %v", deps, err)
			}
			funcDeps = string(bytes)
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
		f, err := getFunctionDescription(cli, funcName, ns, handler, file, funcDeps, runtime, runtimeImage, mem, timeout, envs, labels, secrets, defaultFunctionSpec)
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := utils.GetFunctionClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Deploying function...")
		err = utils.CreateFunctionCustomResource(kubelessClient, f)
		if err != nil {
			logrus.Fatalf("Failed to deploy %s. Received:\n%s", funcName, err)
		}
		logrus.Infof("Function %s submitted for deployment", funcName)
		logrus.Infof("Check the deployment status executing 'kubeless function ls %s'", funcName)

		triggers := []bool{triggerHTTP, topic != "", schedule != ""}
		triggerCount := 0
		for i := len(triggers) - 1; i >= 0; i-- {
			if triggers[i] {
				triggerCount++
			}
		}
		if triggerCount > 1 {
			logrus.Fatal("exactly one of --trigger-http, --trigger-topic, --schedule must be specified")
		}

		// Specifying trigger is not mandatory, if no trigger is specified then just return
		if triggerCount == 0 {
			return
		}

		switch {
		case triggerHTTP:
			httpTrigger := kubelessApi.HTTPTrigger{}
			httpTrigger.TypeMeta = metav1.TypeMeta{
				Kind:       "HTTPTrigger",
				APIVersion: "kubeless.io/v1beta1",
			}
			httpTrigger.ObjectMeta = metav1.ObjectMeta{
				Name:      funcName,
				Namespace: ns,
			}
			httpTrigger.ObjectMeta.Labels = map[string]string{
				"created-by": "kubeless",
			}
			httpTrigger.Spec.FunctionName = funcName

			svcSpec := v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:     "http-function-port",
						NodePort: 0,
						Protocol: v1.ProtocolTCP,
					},
				},
				Selector: f.ObjectMeta.Labels,
				Type:     v1.ServiceTypeClusterIP,
			}

			if headless {
				svcSpec.ClusterIP = v1.ClusterIPNone
			}

			if port != 0 {
				svcSpec.Ports[0].Port = port
				svcSpec.Ports[0].TargetPort = intstr.FromInt(int(port))
			}
			httpTrigger.Spec.ServiceSpec = svcSpec
			err = utils.CreateHTTPTriggerCustomResource(kubelessClient, &httpTrigger)
			if err != nil {
				logrus.Fatalf("Failed to deploy HTTP job trigger %s. Received:\n%s", funcName, err)
			}
			break
		case schedule != "":
			cronJobTrigger := kubelessApi.CronJobTrigger{}
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
			}
			cronJobTrigger.Spec.FunctionName = funcName
			cronJobTrigger.Spec.Schedule = schedule
			err = utils.CreateCronJobCustomResource(kubelessClient, &cronJobTrigger)
			if err != nil {
				logrus.Fatalf("Failed to deploy cron job trigger %s. Received:\n%s", funcName, err)
			}
			break
		case topic != "":
			kafkaTrigger := kubelessApi.KafkaTrigger{}
			kafkaTrigger.TypeMeta = metav1.TypeMeta{
				Kind:       "KafkaTrigger",
				APIVersion: "kubeless.io/v1beta1",
			}
			kafkaTrigger.ObjectMeta = metav1.ObjectMeta{
				Name:      funcName,
				Namespace: ns,
			}
			kafkaTrigger.ObjectMeta.Labels = map[string]string{
				"created-by": "kubeless",
				"function":   funcName,
			}
			kafkaTrigger.Spec.FunctionSelector.MatchLabels = f.ObjectMeta.Labels
			kafkaTrigger.Spec.Topic = topic
			err = utils.CreateKafkaTriggerCustomResource(kubelessClient, &kafkaTrigger)
			if err != nil {
				logrus.Fatalf("Failed to deploy Kafka trigger %s. Received:\n%s", funcName, err)
			}
			break
		}
	},
}

func init() {
	deployCmd.Flags().StringP("runtime", "", "", "Specify runtime")
	deployCmd.Flags().StringP("handler", "", "", "Specify handler")
	deployCmd.Flags().StringP("from-file", "", "", "Specify code file")
	deployCmd.Flags().StringSliceP("label", "", []string{}, "Specify labels of the function. Both separator ':' and '=' are allowed. For example: --label foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringSliceP("secrets", "", []string{}, "Specify Secrets to be mounted to the functions container. For example: --secrets mySecret")
	deployCmd.Flags().StringArrayP("env", "", []string{}, "Specify environment variable of the function. Both separator ':' and '=' are allowed. For example: --env foo1=bar1,foo2:bar2")
	deployCmd.Flags().StringP("namespace", "", "", "Specify namespace for the function")
	deployCmd.Flags().StringP("dependencies", "", "", "Specify a file containing list of dependencies for the function")
	deployCmd.Flags().StringP("trigger-topic", "", "", "Deploy a pubsub function to Kubeless")
	deployCmd.Flags().StringP("schedule", "", "", "Specify schedule in cron format for scheduled function")
	deployCmd.Flags().StringP("memory", "", "", "Request amount of memory, which is measured in bytes, for the function. It is expressed as a plain integer or a fixed-point interger with one of these suffies: E, P, T, G, M, K, Ei, Pi, Ti, Gi, Mi, Ki")
	deployCmd.Flags().Bool("trigger-http", false, "Deploy a http-based function to Kubeless")
	deployCmd.Flags().StringP("runtime-image", "", "", "Custom runtime image")
	deployCmd.Flags().StringP("timeout", "", "180", "Maximum timeout (in seconds) for the function to complete its execution")
	deployCmd.Flags().Bool("headless", false, "Deploy http-based function without a single service IP and load balancing support from Kubernetes. See: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services")
	deployCmd.Flags().Int32("port", 8080, "Deploy http-based function with a custom port")
}
