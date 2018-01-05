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

var topicCreateCmd = &cobra.Command{
	Use:   "create <topic_name> FLAG",
	Short: "create a topic to Kubeless",
	Long:  `create a topic to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - topic name")
		}
		ctlNamespace, err := cmd.Flags().GetString("kafka-namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		topicName := args[0]

		conf, err := utils.BuildOutOfClusterConfig()
		if err != nil {
			logrus.Fatal(err)
		}

		k8sClientSet := utils.GetClientOutOfCluster()

		err = createTopic(conf, k8sClientSet, ctlNamespace, topicName, cmd.OutOrStdout())
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func createTopic(conf *rest.Config, clientset kubernetes.Interface, ctlNamespace, topicName string, out io.Writer) error {
	command := []string{
		"bash", "/opt/bitnami/kafka/bin/kafka-topics.sh",
		"--zookeeper", "zookeeper." + ctlNamespace + ":2181",
		"--replication-factor", "1",
		"--partitions", "1",
		"--create",
		"--topic", topicName,
	}
	return execCommand(conf, clientset, ctlNamespace, command, out)
}

// wrapper of kubectl exec
// execCommand executes a command to kafka pod
func execCommand(conf *rest.Config, k8sClientSet kubernetes.Interface, ctlNamespace string, command []string, out io.Writer) error {
	pods, err := utils.GetPodsByLabel(k8sClientSet, ctlNamespace, "kubeless", "kafka")
	if err != nil {
		return fmt.Errorf("Can't find the kafka pod: %v", err)
	} else if len(pods.Items) == 0 {
		return fmt.Errorf("Can't find any kafka pod")
	}

	cmd := utils.Cmd{
		Stdout: out,
		Stderr: out,
	}
	rt, err := utils.ExecRoundTripper(conf, cmd.RoundTripCallback)
	if err != nil {
		return err
	}

	opts := v1.PodExecOptions{
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		Container: "broker",
		Command:   command,
	}

	req, err := utils.Exec(k8sClientSet.Core(), pods.Items[0].Name, ctlNamespace, opts)
	if err != nil {
		return err
	}

	_, err = rt.RoundTrip(req)
	return err
}
