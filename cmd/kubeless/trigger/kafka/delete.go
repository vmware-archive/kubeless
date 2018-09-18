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

package kafka

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kafkaUtils "github.com/kubeless/kafka-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
)

var deleteCmd = &cobra.Command{

	Use:   "delete <kafka_trigger_name>",
	Short: "Delete a Kafka trigger",
	Long:  `Delete a Kafka trigger`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - Kafka trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		kafkaClient, err := kafkaUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}
		err = kafkaUtils.DeleteKafkaTriggerCustomResource(kafkaClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Failed to delete Kafka trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Kafka trigger %s deleted from namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	deleteCmd.Flags().StringP("namespace", "n", "", "Specify namespace of the Kafka trigger")
}
