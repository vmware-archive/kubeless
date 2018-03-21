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

package http

import (
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var updateCmd = &cobra.Command{
	Use:   "update <http_trigger_name> FLAG",
	Short: "Update a http trigger",
	Long:  `Update a http trigger`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - cronjob trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		httpTrigger, err := utils.GetHTTPTriggerCustomResource(kubelessClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find HTTP trigger %s in namespace %s. Error %s", triggerName, ns, err)
		}

		functionName, err := cmd.Flags().GetString("function-name")
		if err != nil {
			logrus.Fatal(err)
		}

		if functionName != "" {
			_, err = utils.GetFunctionCustomResource(kubelessClient, functionName, ns)
			if err != nil {
				logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
			}
			httpTrigger.Spec.FunctionName = functionName
		}

		headless, err := cmd.Flags().GetBool("headless")
		if err != nil {
			logrus.Fatal(err)
		}

		if headless {
			httpTrigger.Spec.ServiceSpec.ClusterIP = v1.ClusterIPNone
		}

		port, err := cmd.Flags().GetInt32("port")
		if err != nil {
			logrus.Fatal(err)
		}
		if port <= 0 || port > 65535 {
			logrus.Fatalf("Invalid port number %d specified", port)
		}
		if port != 0 {
			httpTrigger.Spec.ServiceSpec.Ports[0].Port = port
			httpTrigger.Spec.ServiceSpec.Ports[0].TargetPort = intstr.FromInt(int(port))
		}

		err = utils.UpdateHTTPTriggerCustomResource(kubelessClient, httpTrigger)
		if err != nil {
			logrus.Fatalf("Failed to deploy HTTP trigger %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("HTTP trigger %s updated in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	updateCmd.Flags().StringP("namespace", "", "", "Specify namespace for the function")
	updateCmd.Flags().StringP("function-name", "", "", "Name of the function to be associated with trigger")
	updateCmd.Flags().Bool("headless", false, "Deploy http-based function without a single service IP and load balancing support from Kubernetes. See: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services")
	updateCmd.Flags().Int32("port", 8080, "Deploy http-based function with a custom port")
}
