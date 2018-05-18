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
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	mutex            = &sync.Mutex{}
	stopM            map[string](chan struct{})
	stoppedM         map[string](chan struct{})
	streamProcessors map[string]bool
	url              string
)

func init() {
	stopM = make(map[string](chan struct{}))
	stoppedM = make(map[string](chan struct{}))
	streamProcessors = make(map[string]bool)
}

// createStreamProcessor polls and gets messages from given AWS kinesis stream and send the stream records to function service
func createStreamProcessor(triggerObj *kubelessApi.KinesisTrigger, funcName, ns string, clientset kubernetes.Interface, stopchan, stoppedchan chan struct{}) {

	// TODO: for now use 1 sec as polling period to poll Kinesis stream. But need to figure right value as per best practice, as let user specify
	// polling period as its very subjective to application requirements.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	defer close(stoppedchan)

	client := utils.GetClient()
	secret, err := client.Core().Secrets(ns).Get(triggerObj.Spec.Secret, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error getting secret: %s necessary to connect to AWS Kinesis service. Error: %v", triggerObj.Spec.Secret, err)
	}
	if _, ok := secret.Data["aws_access_key_id"]; !ok {
		logrus.Errorf("Error getting aws_access_key_id from the secret: %s necessary to connect to AWS Kinesis service. Error: %v", triggerObj.Spec.Secret, err)
	}
	if _, ok := secret.Data["aws_secret_access_key"]; !ok {
		logrus.Errorf("Error getting aws_secret_access_key from the secret: %s necessary to connect to AWS Kinesis service. Error: %v", triggerObj.Spec.Secret, err)
	}
	awsAccessKey := string(secret.Data["aws_access_key_id"][:])
	awsSecretAccessKey := string(secret.Data["aws_secret_access_key"][:])

	customCreds := credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, "")
	s := session.New(&aws.Config{Region: aws.String(triggerObj.Spec.Region), Credentials: customCreds})

	kc := kinesis.New(s)
	shardIterator, err := getShardIterator(kc, triggerObj.Spec.ShardID, triggerObj.Spec.Stream)
	if err != nil {
		logrus.Errorf("Error getting shard iterator necessary to read records from the Kinesis stream %s in region %s. Error: %v", triggerObj.Spec.Stream, triggerObj.Spec.Region, err)
	}

	for {
		// get records using shard iterator for making request
		records, err := kc.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIterator,
		})
		if err != nil {
			// Kinesis shard iterator is only valid for fixed duration of time, so refresh it if we run into ErrCodeExpiredIteratorException exception
			if strings.HasPrefix(err.Error(), kinesis.ErrCodeExpiredIteratorException) {
				shardIterator, err = getShardIterator(kc, triggerObj.Spec.ShardID, triggerObj.Spec.Stream)
				if err != nil {
					logrus.Errorf("Error getting shard iterator. Error: %v", err)
				}
			} else {
				logrus.Errorf("Error getting record from Kinesis stream %s in region %s. Error: %v", triggerObj.Spec.Stream, triggerObj.Spec.Region, err)
			}
		}
		if len(records.Records) > 0 {
			for _, record := range records.Records {
				data := string(record.Data[:])
				req, err := utils.GetHTTPReq(clientset, funcName, ns, "kinesistriggers.kubeless.io", "POST", data)
				if err != nil {
					logrus.Errorf("Unable to elaborate request: %v", err)
				} else {
					//forward msg to function
					err = utils.SendMessage(req)
					if err != nil {
						logrus.Errorf("Failed to send message to function: %v", err)
					} else {
						logrus.Infof("Record from stream: %s in region: %s has sent to function %s successfully", triggerObj.Spec.Stream, triggerObj.Spec.Region, funcName)
					}
				}
			}
			// fetch record carries the iterator to fetch subsequnet record, so refresh the iterator
			shardIterator = records.NextShardIterator
		}
		select {
		case <-stopchan:
			return
		case <-ticker.C:
		}
	}
}

func getShardIterator(kc *kinesis.Kinesis, shardID, streamName string) (*string, error) {
	iteratorOutput, err := kc.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           &shardID,
		StreamName:        &streamName,
		ShardIteratorType: aws.String("LATEST"),
	})
	if err != nil {
		return nil, fmt.Errorf("Error getting shard iterator. Error: %v", err)
	}
	return iteratorOutput.ShardIterator, nil
}

// CreateKinesisStreamConsumer creates a goroutine that polls the Kinesis stream for new records and forwards the data to function
func CreateKinesisStreamConsumer(triggerObj *kubelessApi.KinesisTrigger, funcName, ns string, clientset kubernetes.Interface) error {
	mutex.Lock()
	defer mutex.Unlock()
	uniqueID := generateUniqueStreamProcessorID(triggerObj.Name, funcName, ns, triggerObj.Spec.Stream)
	if !streamProcessors[uniqueID] {
		logrus.Infof("Creating Kinesis stream processor for the function %s associated with Kinesis trigger %s", funcName, triggerObj.Name)
		stopM[uniqueID] = make(chan struct{})
		stoppedM[uniqueID] = make(chan struct{})
		go createStreamProcessor(triggerObj, funcName, ns, clientset, stopM[uniqueID], stoppedM[uniqueID])
		streamProcessors[uniqueID] = true
		logrus.Infof("Created Kinesis stream processor for the function %s associated with Kinesis trigger %s", funcName, triggerObj.Name)
	} else {
		logrus.Infof("Kinesis stream processor for function %s associated with trigger %s already exists, so just returning", funcName, triggerObj.Name)
	}
	return nil
}

// DeleteKinesisConsumer deletes goroutine created by CreateNATSConsumer
func DeleteKinesisConsumer(triggerObj *kubelessApi.KinesisTrigger, funcName, ns string) error {
	mutex.Lock()
	defer mutex.Unlock()
	uniqueID := generateUniqueStreamProcessorID(triggerObj.Name, funcName, ns, triggerObj.Spec.Stream)
	if streamProcessors[uniqueID] {
		logrus.Infof("Stopping Kinesis stream processor for the function %s associated with Kinesis trigger %s", funcName, triggerObj.Name)
		// delete consumer process
		close(stopM[uniqueID])
		<-stoppedM[uniqueID]
		delete(streamProcessors, uniqueID)
		logrus.Infof("Stopped  Kinesis stream processor for the function %s associated with Kinesis trigger %s", funcName, triggerObj.Name)
	} else {
		logrus.Infof(" Kinesis stream processor for function %s associated with trigger doesn't exists. Good enough to skip the stop", funcName, triggerObj.Name)
	}
	return nil
}

// generates unique id for internal book keeping of stream processors associated with active Kinesis triggers
func generateUniqueStreamProcessorID(triggerObjName, funcName, ns, streamName string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + streamName
}
