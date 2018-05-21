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

package kinesis

import (
	"github.com/spf13/cobra"
)

// KinesisTriggerCmd command for Kinesis trigger commands
var KinesisTriggerCmd = &cobra.Command{
	Use:   "kinesis SUBCOMMAND",
	Short: "kinesis trigger specific operations",
	Long:  `kinesis trigger command allows users to create, list, update, delete Kinesis triggers running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	KinesisTriggerCmd.AddCommand(createCmd)
	KinesisTriggerCmd.AddCommand(deleteCmd)
	KinesisTriggerCmd.AddCommand(listCmd)
	KinesisTriggerCmd.AddCommand(updateCmd)
	KinesisTriggerCmd.AddCommand(publishCmd)
	KinesisTriggerCmd.AddCommand(createStreamCmd)
}
