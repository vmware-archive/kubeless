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
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var routeDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "delete a route from Kubeless",
	Long:  `delete a route from Kubeless`,
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

		crdClient, err := utils.GetCRDClientOutOfCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		function, err := utils.GetFunction(crdClient, funcName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find the function %s. Received %s: ", funcName, err)
		}

		function.Spec.Route = v1beta1.Ingress{}
		logrus.Infof("Removing routing rule to the function...")
		err = utils.UpdateK8sCustomResource(crdClient, &function)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Removed routing rule for %s", funcName)
	},
}
