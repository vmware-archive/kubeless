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
	"os"

	"sync"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	mutex     = &sync.Mutex{}
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
		req, err := utils.GetHTTPReq(clientset, funcName, ns, "natstriggers.kubeless.io", "POST", string(msg.Data))
		if err != nil {
			logrus.Errorf("Unable to elaborate request: %v", err)
		} else {
			//forward msg to function
			err = utils.SendMessage(req)
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

// CreateNATSConsumer creates a goroutine that subscribes to NATS topic
func CreateNATSConsumer(triggerObjName, funcName, ns, topic string, clientset kubernetes.Interface) error {
	mutex.Lock()
	defer mutex.Unlock()
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
	mutex.Lock()
	defer mutex.Unlock()
	queueGroupID := generateUniqueQueueGroupName(triggerObjName, funcName, ns, topic)
	if consumerM[queueGroupID] {
		logrus.Infof("Stopping consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		// delete consumer process
		close(stopM[queueGroupID])
		<-stoppedM[queueGroupID]
		delete(consumerM, queueGroupID)
		logrus.Infof("Stopped consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	} else {
		logrus.Infof("Consumer for function %s associated with trigger does n't exists. Good enough to skip the stop", funcName, triggerObjName)
	}
	return nil
}

func generateUniqueQueueGroupName(triggerObjName, funcName, ns, topic string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + topic
}
