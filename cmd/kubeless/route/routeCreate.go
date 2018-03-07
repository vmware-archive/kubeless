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

package route

import (
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var routeCreateCmd = &cobra.Command{
	Use:   "create <name> FLAG",
	Short: "create a route to function",
	Long:  `create a route to function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - route name")
		}
		ingressName := args[0]
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		httpTriggerName, err := cmd.Flags().GetString("http-trigger")
		if err != nil {
			logrus.Fatal(err)
		}

		hostName, err := cmd.Flags().GetString("hostname")
		if err != nil {
			logrus.Fatal(err)
		}
		if hostName == "" {
			config, err := utils.BuildOutOfClusterConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			hostName, err = utils.GetLocalHostname(config, httpTriggerName)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		enableTLSAcme, err := cmd.Flags().GetBool("enableTLSAcme")
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := utils.GetHTTPTriggerClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		httpTrigger, err := kubelessClient.KubelessV1beta1().HTTPTriggers(ns).Get(httpTriggerName, metav1.GetOptions{})
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				logrus.Fatalf("http trigger %s doesn't exist in namespace %s", httpTriggerName, ns)
			} else {
				logrus.Fatalf("error validate input %v", err)
			}
		}

		httpTrigger.Spec.HostName = hostName
		httpTrigger.Spec.TLSAcme = enableTLSAcme
		httpTrigger.Spec.RouteName = ingressName
		httpTrigger.Spec.EnableIngress = true

		err = utils.UpdateHTTPTriggerCustomResource(kubelessClient, httpTrigger)
		if err != nil {
			logrus.Fatalf("Can't create route: %v", err)
		}
	},
}

func init() {
	routeCreateCmd.Flags().StringP("hostname", "", "", "Specify a valid hostname for the function")
	routeCreateCmd.Flags().StringP("http-trigger", "", "", "Name of the http-trigger")
	routeCreateCmd.Flags().BoolP("enableTLSAcme", "", false, "If true, routing rule will be configured for use with kube-lego")
}
