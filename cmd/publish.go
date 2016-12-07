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
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish FLAG",
	Short: "publish message to PubSub function",
	Long:  `publish message to PubSub function`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := cmd.Flags().GetString("data")
		if err != nil {
			logrus.Fatal(err)
		}

		topic, err := cmd.Flags().GetString("topic")
		if err != nil {
			logrus.Fatal(err)
		}

		body := fmt.Sprintf(`echo %s > msg.txt`, data)
		command := []string{"bash", "-c", body}
		execCommand(command)

		body = fmt.Sprintf(`/opt/kafka_2.11-0.10.1.0/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic %s < msg.txt`, topic)
		command = []string{"bash", "-c", body}
		execCommand(command)
	},
}

func init() {
	publishCmd.Flags().StringP("data", "", "", "Specify data for function")
	publishCmd.Flags().StringP("topic", "", "kubeless", "Specify topic name")
}
