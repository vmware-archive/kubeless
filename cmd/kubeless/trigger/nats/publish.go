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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nats-io/go-nats"
)

var publishCmd = &cobra.Command{
	Use:   "publish FLAG",
	Short: "publish message to a topic",
	Long:  `publish message to a topic`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := cmd.Flags().GetString("message")
		if err != nil {
			logrus.Fatal(err)
		}

		topic, err := cmd.Flags().GetString("topic")
		if err != nil {
			logrus.Fatal(err)
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			logrus.Fatal(err)
		}

		err = publishTopic(topic, data, url)
		if err != nil {
			logrus.Fatal("Failed to publish message to topic: ", err)
		}
	},
}

func publishTopic(topic, message, url string) error {
	nc, err := nats.Connect(url)
	if err != nil {
		logrus.Fatal(err)
	}
	defer nc.Close()
	nc.Publish(topic, []byte(message))
	nc.Flush()
	if err := nc.LastError(); err != nil {
		return err
	}
	logrus.Infof("Published [%s] : '%s'\n", topic, message)
	return nil
}

func init() {
	publishCmd.Flags().StringP("message", "", "", "Specify message to be published")
	publishCmd.Flags().StringP("topic", "", "kubeless", "Specify topic name")
	publishCmd.Flags().StringP("url", "", "", "Specify NATS server details for e.g nats://localhost:4222)")
	publishCmd.MarkFlagRequired("url")
	publishCmd.MarkFlagRequired("topic")
	publishCmd.MarkFlagRequired("message")
}
