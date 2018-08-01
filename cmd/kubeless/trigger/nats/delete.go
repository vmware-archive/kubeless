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

	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	natsUtils "github.com/kubeless/nats-trigger/pkg/utils"
)

var deleteCmd = &cobra.Command{

	Use:   "delete <nats_trigger_name>",
	Short: "Delete a NATS trigger",
	Long:  `Delete a NATS trigger`,
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

		err = natsUtils.DeleteNatsTriggerCustomResource(natsClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Failed to delete NATS trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("NATS trigger %s deleted from namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	deleteCmd.Flags().StringP("namespace", "n", "", "Specify namespace of the NATS trigger")
}
