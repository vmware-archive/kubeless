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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kinesisUtils "github.com/kubeless/kinesis-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
)

var updateCmd = &cobra.Command{
	Use:   "update <kinesis_trigger_name> FLAG",
	Short: "Update a Kinesis trigger",
	Long:  `Update a Kinesis trigger`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - Kinesis trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		kubelessClient, err := kubelessUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}
		kinesisClient, err := kinesisUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		kinesisTrigger, err := kinesisUtils.GetKinesisTriggerCustomResource(kinesisClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Kinesis trigger %s in namespace %s. Error %s", triggerName, ns, err)
		}

		streamName, err := cmd.Flags().GetString("stream")
		if err != nil {
			logrus.Fatal(err)
		}

		regionName, err := cmd.Flags().GetString("aws-region")
		if err != nil {
			logrus.Fatal(err)
		}

		shardID, err := cmd.Flags().GetString("shard-id")
		if err != nil {
			logrus.Fatal(err)
		}
		secretName, err := cmd.Flags().GetString("secret")
		if err != nil {
			logrus.Fatal(err)
		}
		functionName, err := cmd.Flags().GetString("function-name")
		if err != nil {
			logrus.Fatal(err)
		}
		if functionName != "" {
			_, err = kubelessUtils.GetFunctionCustomResource(kubelessClient, functionName, ns)
			if err != nil {
				logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
			}
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		if regionName != "" {
			kinesisTrigger.Spec.Region = regionName
		}
		if secretName != "" {
			kinesisTrigger.Spec.Secret = secretName
		}
		if shardID != "" {
			kinesisTrigger.Spec.ShardID = shardID
		}
		if streamName != "" {
			kinesisTrigger.Spec.Stream = streamName
		}

		if dryrun == true {
			res, err := kubelessUtils.DryRunFmt(output, kinesisTrigger)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(res)
			return
		}

		err = kinesisUtils.UpdateKinesisTriggerCustomResource(kinesisClient, kinesisTrigger)
		if err != nil {
			logrus.Fatalf("Failed to update Kinesis trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Kinesis trigger %s updated in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	updateCmd.Flags().StringP("stream", "", "", "Name of the AWS Kinesis stream")
	updateCmd.Flags().StringP("aws-region", "", "", "AWS region in which stream is available")
	updateCmd.Flags().StringP("shard-id", "", "", "Shard-ID of the AWS kinesis stream")
	updateCmd.Flags().StringP("function-name", "", "", "Name of the Kubeless function to be associated with AWS Kinesis stream")
	updateCmd.Flags().StringP("secret", "", "", "Kubernetes secret that has AWS access key and secret key")
	updateCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	updateCmd.Flags().StringP("output", "o", "yaml", "Output format")
}
