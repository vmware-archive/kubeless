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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/kubeless/kubeless/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createStreamCmd = &cobra.Command{

	Use:   "create-stream <kinesis_streamr_name> FLAG",
	Short: "Create a Kinesis stream",
	Long:  `Create a Kinesis stream. Provide only for convenience/quick testing in Kubeless cli`,
	Run: func(cmd *cobra.Command, args []string) {

		regionName, err := cmd.Flags().GetString("aws-region")
		if err != nil {
			logrus.Fatal(err)
		}

		shardCount, err := cmd.Flags().GetInt64("shard-count")
		if err != nil {
			logrus.Fatal(err)
		}
		endpointURL, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			logrus.Fatal(err)
		}
		_, err = url.ParseRequestURI(endpointURL)
		if err != nil {
			panic(err)
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}
		secretName, err := cmd.Flags().GetString("secret")
		if err != nil {
			logrus.Fatal(err)
		}
		client := utils.GetClientOutOfCluster()
		secret, err := client.Core().Secrets(ns).Get(secretName, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("Error getting secret: %s necessary to connect to AWS services. Erro: %v", secretName, err)
		}
		if _, ok := secret.Data["aws_access_key_id"]; !ok {
			logrus.Fatalf("Error getting aws_access_key_id from the secret: %s necessary to connect to AWS Kinesis service. Error: %v", secretName, err)
		}
		if _, ok := secret.Data["aws_secret_access_key"]; !ok {
			logrus.Fatalf("Error getting aws_aaws_secret_access_keyccess_key_id from the secret: %s necessary to connect to AWS Kinesis service. Error: %v", secretName, err)
		}
		awsAccessKey := string(secret.Data["aws_access_key_id"][:])
		awsSecretAccessKey := string(secret.Data["aws_secret_access_key"][:])

		streamName, err := cmd.Flags().GetString("stream-name")
		customCreds := credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, "")
		var s *session.Session
		if len(endpointURL) > 0 {
			s = session.New(&aws.Config{Region: aws.String(regionName), Endpoint: aws.String(endpointURL), Credentials: customCreds})
		} else {
			s = session.New(&aws.Config{Region: aws.String(regionName), Credentials: customCreds})
		}

		kc := kinesis.New(s)
		_, err = kc.CreateStream(&kinesis.CreateStreamInput{ShardCount: &shardCount, StreamName: &streamName})
		if err != nil {
			logrus.Fatalf("Failed to create Kinesis stream. Error: %v", err)
		}
	},
}

func init() {
	createStreamCmd.Flags().StringP("stream-name", "", "", "A  name to identify the stream.")
	createStreamCmd.Flags().StringP("aws-region", "", "", "AWS region in which stream is to be created.")
	createStreamCmd.Flags().Int64("shard-count", 1, "The number of shards that the stream will use.")
	createStreamCmd.Flags().StringP("endpoint", "", "", "Override AWS's default service URL with the given URL")
	createStreamCmd.Flags().StringP("secret", "", "", "Kubernetes secret that has AWS access key and secret key")
	createStreamCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the Kinesis trigger")
	createStreamCmd.MarkFlagRequired("stream-name")
	createStreamCmd.MarkFlagRequired("aws-region")
	createStreamCmd.MarkFlagRequired("aws_access_key_id")
	createStreamCmd.MarkFlagRequired("aws_secret_access_key")
	createStreamCmd.MarkFlagRequired("secret")
}
