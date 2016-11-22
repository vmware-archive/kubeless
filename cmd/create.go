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

var createCmd = &cobra.Command{
	Use:   "create <function_name> FLAG",
	Short: "create a function to Kubeless",
	Long: `create a function to Kubeless`,
	Run: func(cmd *cobra.Command, args []string){
		runtime, err := cmd.Flags().GetString("runtime")
		if err != nil {
			logrus.Fatal(err)
		}

		handler, err := cmd.Flags().GetString("handler")
		if err != nil {
			logrus.Fatal(err)
		}

		file, err := cmd.Flags().GetString("from-file")
		if err != nil {
			logrus.Fatal(err)
		}

		master, err := cmd.Flags().GetString("master")
		if err != nil {
			logrus.Fatal(err)
		}

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		utils.CreateK8sCustomResource(runtime, handler, file, funcName, master)
	},
}

func init() {
	createCmd.Flags().StringP("runtime", "", "", "Specify runtime")
	createCmd.Flags().StringP("handler", "", "", "Specify handler")
	createCmd.Flags().StringP("from-file", "", "", "Specify code file")
}