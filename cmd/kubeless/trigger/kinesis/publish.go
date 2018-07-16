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
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var publishCmd = &cobra.Command{
	Use:   "publish FLAG",
	Short: "publish message to a Kinesis stream",
	Long:  `publish message to a Kinesis stream`,
	Run: func(cmd *cobra.Command, args []string) {

		records, err := cmd.Flags().GetStringArray("records")
		if err != nil {
			logrus.Fatal(err)
		}

		streamName, err := cmd.Flags().GetString("stream")
		if err != nil {
			logrus.Fatal(err)
		}

		region, err := cmd.Flags().GetString("aws-region")
		if err != nil {
			logrus.Fatal(err)
		}

		key, err := cmd.Flags().GetString("partition-key")
		if err != nil {
			logrus.Fatal(err)
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
		customCreds := credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, "")
		var s *session.Session
		if len(endpointURL) > 0 {
			s = session.New(&aws.Config{Region: aws.String(region), Endpoint: aws.String(endpointURL), Credentials: customCreds})
		} else {
			s = session.New(&aws.Config{Region: aws.String(region), Credentials: customCreds})
		}
		kc := kinesis.New(s)
		entries := make([]*kinesis.PutRecordsRequestEntry, len(records))
		for i, record := range records {
			entries[i] = &kinesis.PutRecordsRequestEntry{
				Data:         []byte(record),
				PartitionKey: aws.String(key),
			}
		}
		_, err = kc.PutRecords(&kinesis.PutRecordsInput{
			Records:    entries,
			StreamName: aws.String(streamName),
		})
		if err != nil {
			panic("Failed to put to record in the stream. Error: " + err.Error())
		}
	},
}

func init() {
	var records []string
	publishCmd.Flags().StringP("stream", "", "", "Name of the AWS Kinesis stream")
	publishCmd.Flags().StringP("aws-region", "", "", "AWS region in which stream is available")
	publishCmd.Flags().StringP("partition-key", "", "", "partiion key to use put message in AWS kinesis stream")
	publishCmd.Flags().StringArray("records", records, "Specify list of records to be published to the stream")
	publishCmd.Flags().StringP("endpoint", "", "", "Override AWS's default service URL with the given URL")
	publishCmd.Flags().StringP("secret", "", "", "Kubernetes secret that has AWS access key and secret key")
	publishCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the Kinesis trigger")
	publishCmd.MarkFlagRequired("stream")
	publishCmd.MarkFlagRequired("aws-region")
	publishCmd.MarkFlagRequired("partition-key")
	publishCmd.MarkFlagRequired("message")
	publishCmd.MarkFlagRequired("aws_access_key_id")
	publishCmd.MarkFlagRequired("aws_secret_access_key")
	publishCmd.MarkFlagRequired("records")
	publishCmd.MarkFlagRequired("secret")
}
