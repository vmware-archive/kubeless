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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	stopM     map[string](chan struct{})
	stoppedM  map[string](chan struct{})
	consumerM map[string]bool
	brokers   string
)

func init() {
	stopM = make(map[string](chan struct{}))
	stoppedM = make(map[string](chan struct{}))
	consumerM = make(map[string]bool)

	// taking brokers from env var
	brokers = os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka.kubeless:9092"
	}
}

// createConsumerProcess gets messages to a Kafka topic from the broker and send the payload to function service
func createConsumerProcess(broker, topic, funcName, ns, consumerGroupID string, clientset kubernetes.Interface, stopchan, stoppedchan chan struct{}) {
	// Init config
	config := cluster.NewConfig()

	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Init consumer
	brokersSlice := []string{broker}
	topicsSlice := []string{topic}

	consumer, err := cluster.NewConsumer(brokersSlice, consumerGroupID, topicsSlice, config)
	if err != nil {
		logrus.Fatalf("Failed to start consumer: %v", err)
	}
	defer consumer.Close()

	logrus.Infof("Started Kakfa consumer Broker: %v, Topic: %v, Function: %v, consumerID: %v", broker, topic, funcName, consumerGroupID)

	// Consume messages, wait for signal to stopchan to exit
	defer close(stoppedchan)
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				logrus.Infof("Received Kafka message Partition: %d Offset: %d Key: %s Value: %s ", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				logrus.Infof("Sending message %s to function %s", msg, funcName)
				req, err := getHTTPReq(clientset, funcName, ns, "POST", string(msg.Value))
				if err != nil {
					logrus.Errorf("Unable to elaborate request: %v", err)
				} else {
					//forward msg to function
					err = sendMessage(req)
					if err != nil {
						logrus.Errorf("Failed to send message to function: %v", err)
					} else {
						logrus.Infof("Message has sent to function %s successfully", funcName)
					}
					consumer.MarkOffset(msg, "")
				}
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				logrus.Infof("Rebalanced: %+v\n", ntf)
			}
		case err, more := <-consumer.Errors():
			if more {
				logrus.Errorf("Error: %s\n", err.Error())
			}
		case <-stopchan:
			return
		}
	}
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

func getHTTPReq(clientset kubernetes.Interface, funcName, namespace, method, body string) (*http.Request, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to find the service for function %s", funcName)
	}
	funcPort := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", funcName, namespace, funcPort), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Unable to create request %v", err)
	}
	timestamp := time.Now().UTC()
	eventID, err := utils.GetRandString(11)
	if err != nil {
		return nil, fmt.Errorf("Failed to create a event-ID %v", err)
	}
	req.Header.Add("event-id", eventID)
	req.Header.Add("event-time", timestamp.String())
	req.Header.Add("event-namespace", "kafkatriggers.kubeless.io")
	if isJSON(body) {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("event-type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("event-type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

func sendMessage(req *http.Request) error {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Error: received error code %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

// CreateKafkaConsumer creates a goroutine that subscribes to Kafka topic
func CreateKafkaConsumer(triggerObjName, funcName, ns, topic string, clientset kubernetes.Interface) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if !consumerM[consumerID] {
		logrus.Infof("Creating Kafka consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		stopM[consumerID] = make(chan struct{})
		stoppedM[consumerID] = make(chan struct{})
		go createConsumerProcess(brokers, topic, funcName, ns, consumerID, clientset, stopM[consumerID], stoppedM[consumerID])
		consumerM[consumerID] = true
		logrus.Infof("Created Kafka consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	} else {
		logrus.Infof("Consumer for function %s associated with trigger %s already exists, so just returning", funcName, triggerObjName)
	}
	return nil
}

// DeleteKafkaConsumer deletes goroutine created by CreateKafkaConsumer
func DeleteKafkaConsumer(triggerObjName, funcName, ns, topic string) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if consumerM[consumerID] {
		logrus.Infof("Stopping consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		// delete consumer process
		close(stopM[consumerID])
		<-stoppedM[consumerID]
		consumerM[consumerID] = false
		logrus.Infof("Stopped consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	} else {
		logrus.Infof("Consumer for function %s associated with trigger does n't exists. Good enough to skip the stop", funcName, triggerObjName)
	}
	return nil
}

func generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + topic
}
