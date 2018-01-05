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

package topic

import (
	"fmt"
	"io"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var topicPublishCmd = &cobra.Command{
	Use:   "publish FLAG",
	Short: "publish message to a topic",
	Long:  `publish message to a topic`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := cmd.Flags().GetString("data")
		if err != nil {
			logrus.Fatal(err)
		}

		topic, err := cmd.Flags().GetString("topic")
		if err != nil {
			logrus.Fatal(err)
		}

		ctlNamespace, err := cmd.Flags().GetString("kafka-namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		conf, err := utils.BuildOutOfClusterConfig()
		if err != nil {
			logrus.Fatal(err)
		}

		k8sClientSet := utils.GetClientOutOfCluster()

		err = publishTopic(conf, k8sClientSet, ctlNamespace, topic, data, cmd.OutOrStdout())
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func publishTopic(conf *rest.Config, clientset kubernetes.Interface, namespace, topic, data string, out io.Writer) error {
	command := []string{
		"bash", "/opt/bitnami/kafka/bin/kafka-console-producer.sh",
		"--broker-list", "localhost:9092",
		"--topic", topic,
	}

	// Can't just use `execCommand` since we want to specify stdin
	// TODO(gus): refactor better.

	pods, err := utils.GetPodsByLabel(clientset, namespace, "kubeless", "kafka")
	if err != nil {
		return fmt.Errorf("Can't find the kafka pod: %v", err)
	} else if len(pods.Items) == 0 {
		return fmt.Errorf("Can't find any kafka pod")
	}

	pRead, pWrite := io.Pipe()
	cmd := utils.Cmd{
		Stdin:  pRead,
		Stdout: out,
		Stderr: out,
	}
	rt, err := utils.ExecRoundTripper(conf, cmd.RoundTripCallback)
	if err != nil {
		return err
	}

	go func() {
		io.WriteString(pWrite, data+"\n")
		pWrite.Close()
	}()

	opts := v1.PodExecOptions{
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: "broker",
		Command:   command,
	}

	req, err := utils.Exec(clientset.Core(), pods.Items[0].Name, namespace, opts)
	if err != nil {
		return err
	}

	_, err = rt.RoundTrip(req)
	return err
}

func init() {
	topicPublishCmd.Flags().StringP("data", "", "", "Specify data for function")
	topicPublishCmd.Flags().StringP("topic", "", "kubeless", "Specify topic name")
}
