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
	"github.com/spf13/cobra"
)

var functionCmd = &cobra.Command{
	Use:   "function SUBCOMMAND",
	Short: "function specific operations",
	Long:  `function command allows user to list, create, edit, delete functions running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	functionCmd.AddCommand(createCmd)
	functionCmd.AddCommand(deleteCmd)
	functionCmd.AddCommand(listCmd)
	functionCmd.AddCommand(callCmd)
	functionCmd.AddCommand(publishCmd)
	functionCmd.AddCommand(logsCmd)
	//TODO: reserve edit cmd later
	//functionCmd.AddCommand(editCmd)
}
