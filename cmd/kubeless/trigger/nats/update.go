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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	natsUtils "github.com/kubeless/nats-trigger/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var updateCmd = &cobra.Command{
	Use:   "update <nats_trigger_name> FLAG",
	Short: "Update a NATS trigger",
	Long:  `Update a NATS trigger`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - NATS trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		natsClient, err := natsUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		natsTrigger, err := natsUtils.GetNatsTriggerCustomResource(natsClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find NATS trigger %s in namespace %s. Error %s", triggerName, ns, err)
		}

		topic, err := cmd.Flags().GetString("trigger-topic")
		if err != nil {
			logrus.Fatal(err)
		}

		if topic != "" {
			natsTrigger.Spec.Topic = topic
		}

		functionSelector, err := cmd.Flags().GetString("function-selector")
		if err != nil {
			logrus.Fatal(err)
		}

		if functionSelector != "" {
			labelSelector, err := metav1.ParseToLabelSelector(functionSelector)
			if err != nil {
				logrus.Fatal("Invalid label selector specified " + err.Error())
			}
			natsTrigger.Spec.FunctionSelector.MatchLabels = labelSelector.MatchLabels
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		if dryrun == true {
			res, err := kubelessUtils.DryRunFmt(output, natsTrigger)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(res)
			return
		}

		err = natsUtils.UpdateNatsTriggerCustomResource(natsClient, natsTrigger)
		if err != nil {
			logrus.Fatalf("Failed to update NATS trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("NATS trigger %s updated in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	updateCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the NATS trigger")
	updateCmd.Flags().StringP("trigger-topic", "", "", "Specify topic to listen to in NATS")
	updateCmd.Flags().StringP("function-selector", "", "", "Selector (label query) to select function on (e.g. --function-selector key1=value1,key2=value2)")
	updateCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	updateCmd.Flags().StringP("output", "o", "yaml", "Output format")
}
