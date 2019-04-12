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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createCmd = &cobra.Command{

	Use:   "create <kinesis_trigger_name> FLAG",
	Short: "Create a Kinesis trigger",
	Long:  `Create a Kinesis trigger`,
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
			ns = utils.GetDefaultNamespace()
		}

		functionName, err := cmd.Flags().GetString("function-name")
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		_, err = utils.GetFunctionCustomResource(kubelessClient, functionName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
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
		endpointURL, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			logrus.Fatal(err)
		}
		if len(endpointURL) > 0 {
			_, err = url.ParseRequestURI(endpointURL)
			if err != nil {
				panic(err)
			}
		}

		cli := utils.GetClientOutOfCluster()
		_, err = cli.Core().Secrets(ns).Get(secretName, metav1.GetOptions{})
		if err != nil {
			logrus.Fatal(err)
		}

		kinesisTrigger := kubelessApi.KinesisTrigger{}
		kinesisTrigger.TypeMeta = metav1.TypeMeta{
			Kind:       "KinesisTrigger",
			APIVersion: "kubeless.io/v1beta1",
		}
		kinesisTrigger.ObjectMeta = metav1.ObjectMeta{
			Name:      triggerName,
			Namespace: ns,
		}
		kinesisTrigger.ObjectMeta.Labels = map[string]string{
			"created-by": "kubeless",
		}
		kinesisTrigger.Spec.FunctionName = functionName
		kinesisTrigger.Spec.Region = regionName
		kinesisTrigger.Spec.Stream = streamName
		kinesisTrigger.Spec.ShardID = shardID
		kinesisTrigger.Spec.Secret = secretName
		kinesisTrigger.Spec.Endpoint = endpointURL
		err = utils.CreateKinesisTriggerCustomResource(kubelessClient, &kinesisTrigger)
		if err != nil {
			logrus.Fatalf("Failed to create Kinesis trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Kinesis trigger %s created in namespace %s successfully!", triggerName, ns)

	},
}

func init() {
	createCmd.Flags().StringP("namespace", "", "", "Specify namespace for the Kinesis trigger")
	createCmd.Flags().StringP("stream", "", "", "Name of the AWS Kinesis stream")
	createCmd.Flags().StringP("aws-region", "", "", "AWS region in which stream is available")
	createCmd.Flags().StringP("shard-id", "", "", "Shard-ID of the AWS kinesis stream")
	createCmd.Flags().StringP("function-name", "", "", "Name of the Kubeless function to be associated with AWS Kinesis stream")
	createCmd.Flags().StringP("secret", "", "", "Kubernetes secret that has AWS access key and secret key")
	createCmd.Flags().StringP("endpoint", "", "", "Override AWS's default service URL with the given URL")
	createCmd.MarkFlagRequired("stream")
	createCmd.MarkFlagRequired("aws-region")
	createCmd.MarkFlagRequired("shard-id")
	createCmd.MarkFlagRequired("function-name")
	createCmd.MarkFlagRequired("secret")
}
