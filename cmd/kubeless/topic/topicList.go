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
	"io"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var topicListCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all topics created in Kubeless",
	Long:    `list all topics created in Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		ctlNamespace, err := cmd.Flags().GetString("kafka-namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		conf, err := utils.BuildOutOfClusterConfig()
		if err != nil {
			logrus.Fatal(err)
		}

		k8sClientSet := utils.GetClientOutOfCluster()

		err = listTopic(conf, k8sClientSet, ctlNamespace, cmd.OutOrStdout())
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func listTopic(conf *rest.Config, clientset kubernetes.Interface, ctlNamespace string, out io.Writer) error {
	command := []string{
		"bash", "/opt/bitnami/kafka/bin/kafka-topics.sh",
		"--zookeeper", "zookeeper." + ctlNamespace + ":2181",
		"--list",
	}
	return execCommand(conf, clientset, ctlNamespace, command, out)
}
