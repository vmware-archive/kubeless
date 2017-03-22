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
	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kubeless controller",
	Long: `This command helps to install the Kubeless controller and along with Apache Kafka to handle event-based functions.

By default we will install the latest release of bitnami/kubeless-controller image.
Use your own controller by specifying --controller-image flag.
`,
	Run: func(cmd *cobra.Command, args []string) {
		okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
		fmt.Println("We are going to install the controller into the kubeless namespace. [Y/n]?")
		var text string
		n, _ := fmt.Scanln(&text)
		if n < 1 {
			// If no value is provided, default to Yes
			text = "Y"
		}

		//getting versions
		ctlImage, err := cmd.Flags().GetString("controller-image")
		if err != nil {
			logrus.Fatal(err)
		}
		kafkaVer, err := cmd.Flags().GetString("kafka-version")
		if err != nil {
			logrus.Fatal(err)
		}

		if containsString(okayResponses, text) {
			cfg := controller.NewControllerConfig("", "")
			c := controller.New(cfg)
			c.Init()
			c.InstallKubeless(ctlImage)
			c.InstallMsgBroker(kafkaVer)
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
	installCmd.Flags().StringP("controller-image", "", "", "Install a specific image of Kubeless controller")
	installCmd.Flags().StringP("kafka-version", "", "", "Install a specific version of Kafka")
}
