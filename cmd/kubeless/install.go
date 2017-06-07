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

package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/controller"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kubeless controller",
	Long:  `This command helps to install the Kubeless controller and along with Apache Kafka to handle event-based functions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default namespace
		ctlNamespace, err := cmd.Flags().GetString("controller-namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
		// ToDo martin: this fmt doesn't work
		fmt.Printf("We are going to install the controller into the '%s' namespace. [Y/n]?\n", ctlNamespace)
		var text string
		n, _ := fmt.Scanln(&text)
		if n < 1 {
			// If no value is provided, default to Yes
			text = "Y"
		}

		if containsString(okayResponses, text) {
			cfg := controller.Config{
				KubeCli: utils.GetClientOutOfCluster(),
			}
			c := controller.New(cfg)
			c.Init()
			c.InstallKubeless(ctlNamespace)
			c.InstallMsgBroker(ctlNamespace)
		} else {
			fmt.Println("Kubeless wasn't installed, exiting.")
			return
		}
	},
}

func containsString(slice []string, element string) bool {
	return posString(slice, element) != -1
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func init() {
	installCmd.Flags().StringP("controller-namespace", "", "kubeless", "Install Kubeless to a specific namespace. It will default to 'kubeless'")
}
