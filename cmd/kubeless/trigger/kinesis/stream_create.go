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
		accessKey, err := cmd.Flags().GetString("aws_access_key_id")
		if err != nil {
			logrus.Fatal(err)
		}
		secretKey, err := cmd.Flags().GetString("aws_secret_access_key")
		if err != nil {
			logrus.Fatal(err)
		}
		streamName, err := cmd.Flags().GetString("stream-name")
		if err != nil {
			logrus.Fatal(err)
		}

		customCreds := credentials.NewStaticCredentials(accessKey, secretKey, "")
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
	createStreamCmd.Flags().StringP("aws_access_key_id", "", "", "AWS access key")
	createStreamCmd.Flags().StringP("aws_secret_access_key", "", "", "AWS secret key")
	createStreamCmd.MarkFlagRequired("stream-name")
	createStreamCmd.MarkFlagRequired("aws-region")
	createStreamCmd.MarkFlagRequired("aws_access_key_id")
	createStreamCmd.MarkFlagRequired("aws_secret_access_key")
}
