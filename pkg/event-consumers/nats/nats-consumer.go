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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	stopM     map[string](chan struct{})
	stoppedM  map[string](chan struct{})
	consumerM map[string]bool
	url       string
)

func init() {
	stopM = make(map[string](chan struct{}))
	stoppedM = make(map[string](chan struct{}))
	consumerM = make(map[string]bool)

	// get the NATS server url from env var
	url = os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://nats.nats-io.svc.cluster.local:4222"
	}
}

// createConsumerProcess gets messages for a topic from the NATS server and send the payload to function service
func createConsumerProcess(topic, funcName, ns, queueGroupID string, clientset kubernetes.Interface, stopchan, stoppedchan chan struct{}) {

	nc, err := nats.Connect(url)
	if err != nil {
		logrus.Fatalf("Failed to initialize nats client to server %v due to %v", url, err)
	}
	defer nc.Close()

	logrus.Infof("Joining to the queue group %v and subscribing to the topic %v", queueGroupID, topic)
	subscription, err := nc.QueueSubscribe(topic, queueGroupID, func(msg *nats.Msg) {
		logrus.Debugf("Received Message %v on Topic: %v Queue: %v", string(msg.Data), msg.Subject, msg.Sub.Queue)
		logrus.Infof("Sending message %s to function %s", string(msg.Data), funcName)
		req, err := getHTTPReq(clientset, funcName, ns, "POST", string(msg.Data))
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
		}
	})
	if err != nil {
		logrus.Fatalf("Failed to create queue group %v and subscribing to the topic %v", queueGroupID, topic)
	}
	defer close(stoppedchan)
	for {
		select {
		case <-stopchan:
			err = subscription.Unsubscribe()
			if err != nil {
				logrus.Fatalf("Failed to unsubscribe from the queue group %v and topic %v", queueGroupID, topic)
			}
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
	req.Header.Add("event-namespace", "natstriggers.kubeless.io")
	if isJSON(body) {
		logrus.Infof("TRUE %v", body)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("event-type", "application/json")
	} else {
		logrus.Infof("FLASE %v", body)
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

// CreateNATSConsumer creates a goroutine that subscribes to NATS topic
func CreateNATSConsumer(triggerObjName, funcName, ns, topic string, clientset kubernetes.Interface) error {
	queueGroupID := generateUniqueQueueGroupName(triggerObjName, funcName, ns, topic)
	if !consumerM[queueGroupID] {
		logrus.Infof("Creating NATS consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		stopM[queueGroupID] = make(chan struct{})
		stoppedM[queueGroupID] = make(chan struct{})
		go createConsumerProcess(topic, funcName, ns, queueGroupID, clientset, stopM[queueGroupID], stoppedM[queueGroupID])
		consumerM[queueGroupID] = true
		logrus.Infof("Created NATS consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	} else {
		logrus.Infof("Consumer for function %s associated with trigger %s already exists, so just returning", funcName, triggerObjName)
	}
	return nil
}

// DeleteNATSConsumer deletes goroutine created by CreateNATSConsumer
func DeleteNATSConsumer(triggerObjName, funcName, ns, topic string) error {
	queueGroupID := generateUniqueQueueGroupName(triggerObjName, funcName, ns, topic)
	if consumerM[queueGroupID] {
		logrus.Infof("Stopping consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		// delete consumer process
		close(stopM[queueGroupID])
		<-stoppedM[queueGroupID]
		consumerM[queueGroupID] = false
		logrus.Infof("Stopped consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	} else {
		logrus.Infof("Consumer for function %s associated with trigger does n't exists. Good enough to skip the stop", funcName, triggerObjName)
	}
	return nil
}

func generateUniqueQueueGroupName(triggerObjName, funcName, ns, topic string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + topic
}
