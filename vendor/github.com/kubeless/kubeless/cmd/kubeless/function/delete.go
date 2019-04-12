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
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <function_name>",
	Short: "delete a function from Kubeless",
	Long:  `delete a function from Kubeless`,
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

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}

		// For convenience `kubeless function create` command also creates http/cronjob/kafka trigger object with same
		// name as that of function implicitly depending on which flag specified. Delete if there is any trigger obj with
		// function name.
		err = utils.DeleteHTTPTriggerCustomResource(kubelessClient, funcName, ns)
		if err != nil && !k8sErrors.IsNotFound(err) {
			logrus.Errorf("Failed to delete HTTP trigger associated with the function %s", funcName)
		}

		err = utils.DeleteCronJobCustomResource(kubelessClient, funcName, ns)
		if err != nil && !k8sErrors.IsNotFound(err) {
			logrus.Errorf("Failed to delete Cronjob trigger associated with the function %s", funcName)
		}

		err = utils.DeleteKafkaTriggerCustomResource(kubelessClient, funcName, ns)
		if err != nil && !k8sErrors.IsNotFound(err) {
			logrus.Errorf("Failed to delete Kafka trigger associated with the function %s", funcName)
		}

		err = utils.DeleteFunctionCustomResource(kubelessClient, funcName, ns)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
}
