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

package nats

import (
	"github.com/spf13/cobra"
)

// NATSTriggerCmd command for NATS trigger commands
var NATSTriggerCmd = &cobra.Command{
	Use:   "nats SUBCOMMAND",
	Short: "nats trigger specific operations",
	Long:  `nats trigger command allows user to create, list, update, delete NATS triggers running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	NATSTriggerCmd.AddCommand(createCmd)
	NATSTriggerCmd.AddCommand(deleteCmd)
	NATSTriggerCmd.AddCommand(listCmd)
	NATSTriggerCmd.AddCommand(updateCmd)
	NATSTriggerCmd.AddCommand(publishCmd)
}
