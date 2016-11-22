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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/utils"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <function_name>",
	Short: "delete a function from Kubeless",
	Long: `delete a function from Kubeless`,
	Run: func(cmd *cobra.Command, args []string){
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		master, err := cmd.Flags().GetString("master")
		if err != nil {
			logrus.Fatal(err)
		}

		utils.DeleteK8sCustomResource(funcName, master)
	},
}