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
)

var routeDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "delete a route from Kubeless",
	Long:  `delete a route from Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - route name")
		}
		routeName := args[0]

		httpTriggerName, err := cmd.Flags().GetString("http-trigger")
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

		kubelessClient, err := utils.GetHTTPTriggerClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		httpTrigger, err := utils.GetHTTPTriggerCustomResource(kubelessClient, httpTriggerName, ns)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				logrus.Fatalf("http trigger %s doesn't exist in namespace %s", httpTriggerName, ns)
			} else {
				logrus.Fatalf("error validate input %v, object %v", err, httpTrigger)
			}
		}

		httpTrigger.Spec.RouteName = routeName
		httpTrigger.Spec.EnableIngress = false
		err = utils.UpdateHTTPTriggerCustomResource(kubelessClient, httpTrigger)
		if err != nil {
			logrus.Fatalf("Can't create route: %v", err)
		}
	},
}

func init() {
	routeDeleteCmd.Flags().StringP("http-trigger", "", "", "Name of the http-trigger")
	routeDeleteCmd.MarkFlagRequired("http-trigger")
}
