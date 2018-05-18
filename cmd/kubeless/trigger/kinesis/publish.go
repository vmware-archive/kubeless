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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var publishCmd = &cobra.Command{
	Use:   "publish FLAG",
	Short: "publish message to a Kinesis stream",
	Long:  `publish message to a Kinesis stream`,
	Run: func(cmd *cobra.Command, args []string) {
		message, err := cmd.Flags().GetString("message")
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
		accessKey, err := cmd.Flags().GetString("aws_access_key_id")
		if err != nil {
			logrus.Fatal(err)
		}
		secretKey, err := cmd.Flags().GetString("aws_secret_access_key")
		if err != nil {
			logrus.Fatal(err)
		}

		customCreds := credentials.NewStaticCredentials(accessKey, secretKey, "")
		s := session.New(&aws.Config{Region: aws.String(region), Credentials: customCreds})
		kc := kinesis.New(s)
		_, err = kc.PutRecord(&kinesis.PutRecordInput{
			Data:         []byte(message),
			StreamName:   aws.String(streamName),
			PartitionKey: aws.String(key)})
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	publishCmd.Flags().StringP("stream", "", "", "Name of the AWS Kinesis stream")
	publishCmd.Flags().StringP("aws-region", "", "", "AWS region in which stream is available")
	publishCmd.Flags().StringP("partition-key", "", "", "partiion key to use put message in AWS kinesis stream")
	publishCmd.Flags().StringP("message", "", "", "Specify message to be published")
	publishCmd.Flags().StringP("aws_access_key_id", "", "", "AWS access key")
	publishCmd.Flags().StringP("aws_secret_access_key", "", "", "AWS secret key")
	publishCmd.MarkFlagRequired("stream")
	publishCmd.MarkFlagRequired("aws-region")
	publishCmd.MarkFlagRequired("partition-key")
	publishCmd.MarkFlagRequired("message")
	publishCmd.MarkFlagRequired("aws_access_key_id")
	publishCmd.MarkFlagRequired("aws_secret_access_key")
}
