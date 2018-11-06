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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kafkaApi "github.com/kubeless/kafka-trigger/pkg/apis/kubeless/v1beta1"
	kafkaUtils "github.com/kubeless/kafka-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createCmd = &cobra.Command{

	Use:   "create <kafka_trigger_name> FLAG",
	Short: "Create a Kafka trigger",
	Long:  `Create a Kafka trigger`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - kafka trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		topic, err := cmd.Flags().GetString("trigger-topic")
		if err != nil {
			logrus.Fatal(err)
		}

		functionSelector, err := cmd.Flags().GetString("function-selector")
		if err != nil {
			logrus.Fatal(err)
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		labelSelector, err := metav1.ParseToLabelSelector(functionSelector)
		if err != nil {
			logrus.Fatal("Invalid lable selector specified " + err.Error())
		}

		kafkaTrigger := kafkaApi.KafkaTrigger{}
		kafkaTrigger.TypeMeta = metav1.TypeMeta{
			Kind:       "KafkaTrigger",
			APIVersion: "kubeless.io/v1beta1",
		}
		kafkaTrigger.ObjectMeta = metav1.ObjectMeta{
			Name:      triggerName,
			Namespace: ns,
		}
		kafkaTrigger.ObjectMeta.Labels = map[string]string{
			"created-by": "kubeless",
		}
		kafkaTrigger.Spec.FunctionSelector.MatchLabels = labelSelector.MatchLabels
		kafkaTrigger.Spec.Topic = topic

		if dryrun == true {
			res, err := kubelessUtils.DryRunFmt(output, kafkaTrigger)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(res)
			return
		}

		kafkaClient, err := kafkaUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}
		err = kafkaUtils.CreateKafkaTriggerCustomResource(kafkaClient, &kafkaTrigger)
		if err != nil {
			logrus.Fatalf("Failed to create Kafka trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Kafka trigger %s created in namespace %s successfully!", triggerName, ns)

	},
}

func init() {
	createCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the kafka trigger")
	createCmd.Flags().StringP("trigger-topic", "", "", "Specify topic to listen to in Kafka broker")
	createCmd.Flags().StringP("function-selector", "", "", "Selector (label query) to select function on (e.g. --function-selector key1=value1,key2=value2)")
	createCmd.MarkFlagRequired("trigger-topic")
	createCmd.MarkFlagRequired("function-selector")
	createCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	createCmd.Flags().StringP("output", "o", "yaml", "Output format")
}
