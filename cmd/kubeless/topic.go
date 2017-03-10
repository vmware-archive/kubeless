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

var topicCmd = &cobra.Command{
	Use:   "topic SUBCOMMAND",
	Short: "manage message topics in Kubeless",
	Long:  `topic command allows user to list, create, delete topics on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	topicCmd.AddCommand(topicCreateCmd)
	topicCmd.AddCommand(topicDeleteCmd)
	topicCmd.AddCommand(topicListCmd)
}
