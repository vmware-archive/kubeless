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
	"github.com/bitnami/kubeless/pkg/controller"
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
		nokayResponses := []string{"n", "N", "no", "No", "NO"}
		fmt.Println("We are going to install the controller in the kubeless namespace. Are you OK with this: [Y/N]")
		var text string
		_, err := fmt.Scanln(&text)
		if err != nil {
			logrus.Fatal(err)
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
		} else if containsString(nokayResponses, text) {
			return
		} else {
			fmt.Println("Please type yes or no and then press enter:")
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
