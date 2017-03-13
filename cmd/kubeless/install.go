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
	Short: "install Kubeless controller",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
		nokayResponses := []string{"n", "N", "no", "No", "NO"}
		fmt.Println("We are going to install the controller in the default namespace. Are you OK with this: [Y/N]")
		var text string
		_, err := fmt.Scanln(&text)
		if err != nil {
			logrus.Fatal(err)
		}
		if containsString(okayResponses, text) {
			cfg := controller.NewControllerConfig("", "")
			c := controller.New(cfg)
			c.Init()
			c.InstallKubeless()
			c.InstallMsgBroker()
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
