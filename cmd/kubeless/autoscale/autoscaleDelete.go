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

package autoscale

import (
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/autoscaling/v2beta1"
)

var autoscaleDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "delete an autoscale from Kubeless",
	Long:  `delete an autoscale from Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - autoscale name")
		}
		funcName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		function, err := utils.GetFunction(funcName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find the function %s. Received %s: ", funcName, err)
		}

		if function.Spec.HorizontalPodAutoscaler.Name != "" {
			function.Spec.HorizontalPodAutoscaler = v2beta1.HorizontalPodAutoscaler{}
			kubelessClient, err := utils.GetKubelessClientOutCluster()
			if err != nil {
				logrus.Fatal(err)
			}
			logrus.Infof("Removing autoscaling rule from the function...")
			err = utils.UpdateFunctionCustomResource(kubelessClient, &function)
			if err != nil {
				logrus.Fatal(err)
			}
			logrus.Infof("Remove Autoscaling rule from %s successfully", funcName)
		} else {
			logrus.Fatalf("Not found an autoscale definition for %s", funcName)
		}
	},
}
