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

package trigger

import (
	"github.com/kubeless/kubeless/cmd/kubeless/trigger/cronjob"
	"github.com/kubeless/kubeless/cmd/kubeless/trigger/http"
	"github.com/kubeless/kubeless/cmd/kubeless/trigger/kafka"
	"github.com/kubeless/kubeless/cmd/kubeless/trigger/kinesis"
	"github.com/kubeless/kubeless/cmd/kubeless/trigger/nats"
	"github.com/spf13/cobra"
)

//TriggerCmd contains first-class command for trigger
var TriggerCmd = &cobra.Command{
	Use:   "trigger SUBCOMMAND",
	Short: "trigger specific operations",
	Long:  `trigger command allows user to create, list, update, delete triggers running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	TriggerCmd.AddCommand(cronjob.CronjobTriggerCmd)
	TriggerCmd.AddCommand(kafka.KafkaTriggerCmd)
	TriggerCmd.AddCommand(http.HTTPTriggerCmd)
	TriggerCmd.AddCommand(nats.NATSTriggerCmd)
	TriggerCmd.AddCommand(kinesis.KinesisTriggerCmd)
}
