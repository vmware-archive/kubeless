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
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/kubeless/kinesis-trigger/pkg/client/clientset/versioned"
	kinesisUtils "github.com/kubeless/kinesis-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var listCmd = &cobra.Command{

	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all Kinesis triggers deployed to Kubeless",
	Long:    `list all Kinesis triggers deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		kinesisClient, err := kinesisUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		if err := doList(cmd.OutOrStdout(), kinesisClient, ns); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	listCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the NATS trigger")
}

func doList(w io.Writer, kubelessClient versioned.Interface, ns string) error {
	triggersList, err := kubelessClient.KubelessV1beta1().KinesisTriggers(ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAME", "NAMESPACE", "REGION", "STREAM", "SHARD", "FUNCTION NAME")
	for _, trigger := range triggersList.Items {
		table.AddRow(trigger.Name, trigger.Namespace, trigger.Spec.Region, trigger.Spec.Stream, trigger.Spec.ShardID, trigger.Spec.FunctionName)
	}
	fmt.Fprintln(w, table)
	return nil
}
