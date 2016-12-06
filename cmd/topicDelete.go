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
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api"
	k8scmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	"k8s.io/kubernetes/pkg/util/term"
)

var topicDeleteCmd = &cobra.Command{
	Use:   "delete <topic_name>",
	Short: "delete a topic from Kubeless",
	Long:  `delete a topic from Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - topic name")
		}
		topicName := args[0]

		f := utils.GetFactory()
		ns := "kubeless"
		kClient, err := f.Client()
		if err != nil {
			logrus.Fatalf("Deletion failed: %v", err)
		}
		kClientConfig, err := f.ClientConfig()
		if err != nil {
			logrus.Fatalf("Deletion failed: %v", err)
		}

		command := []string{"bash", "/opt/kafka_2.11-0.10.1.0/bin/kafka-topics.sh", "--zookeeper", "zookeeper:2181", "--delete", "--topic", topicName}

		podName, err := utils.GetPodName(kClient, ns, "kafka-controller")
		params := &k8scmd.ExecOptions{
			StreamOptions: k8scmd.StreamOptions{
				Namespace:     ns,
				PodName:       podName,
				ContainerName: "kafka",
				In:            nil,
				Out:           os.Stdout,
				Err:           os.Stderr,
				TTY:           false,
			},
			Executor: &k8scmd.DefaultRemoteExecutor{},
			Command:  command,
			Client:   kClient,
			Config:   kClientConfig,
		}

		t := setupTTY(params)
		var sizeQueue term.TerminalSizeQueue

		fn := func() error {
			req := params.Client.RESTClient.Post().
				Resource("pods").
				Name(podName).
				Namespace(ns).
				SubResource("exec").
				Param("container", "kafka")
			req.VersionedParams(&api.PodExecOptions{
				Container: "kafka",
				Command:   params.Command,
				Stdin:     params.Stdin,
				Stdout:    params.Out != nil,
				Stderr:    params.Err != nil,
				TTY:       false,
			}, api.ParameterCodec)

			return params.Executor.Execute("POST", req.URL(), params.Config, params.In, params.Out, params.Err, t.Raw, sizeQueue)
		}

		if err := t.Safe(fn); err != nil {
			logrus.Fatalf("Topic deletion failed: %v", err)
		}
	},
}
