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
)

var routeCreateCmd = &cobra.Command{
	Use:   "create <name> FLAG",
	Short: "create a route to function",
	Long:  `create a route to function`,
	Run: func(cmd *cobra.Command, args []string) {
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

		hostName, err := cmd.Flags().GetString("hostname")
		if err != nil {
			logrus.Fatal(err)
		}
		if hostName == "" {
			config, err := utils.BuildOutOfClusterConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			hostName, err = utils.GetLocalHostname(config, funcName)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		enableTLSAcme, err := cmd.Flags().GetBool("enableTLSAcme")
		if err != nil {
			logrus.Fatal(err)
		}

		crdClient, err := utils.GetCRDClientOutOfCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		function, err := utils.GetFunction(crdClient, funcName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find the function %s. Received %s: ", funcName, err)
		}

		route, err := getRouteDefinition(function, funcName, hostName, ns, enableTLSAcme)
		if err != nil {
			logrus.Fatalf("Unable to construct routing rule for the function %s: %v", funcName, err)
		}
		function.Spec.Route = route

		logrus.Infof("Adding routing rule to the function...")
		err = utils.UpdateK8sCustomResource(crdClient, &function)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Routing rule for %s submitted for deployment", funcName)
	},
}

func init() {
	routeCreateCmd.Flags().StringP("hostname", "", "", "Specify a valid hostname for the function")
	routeCreateCmd.Flags().BoolP("enableTLSAcme", "", false, "If true, routing rule will be configured for use with kube-lego")
}
