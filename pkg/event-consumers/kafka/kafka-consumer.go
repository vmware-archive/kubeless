package kafka

import (
	"bytes"
	"errors"
	"fmt"
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
func createConsumerProcess(brokers, topic, funcName, ns, consumerGroupID string, stopchan, stoppedchan chan struct{}) {
	// Init config
	config := cluster.NewConfig()

	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Init consumer, consume errors & messages
	// consumer is grouped and labeled by functionID to receive load-balanced messages
	// More details: https://kafka.apache.org/documentation/#intro_consumers
	consumer, err := cluster.NewConsumer(strings.Split(brokers, ","), consumerGroupID, strings.Split(topic, ","), config)
	if err != nil {
		logrus.Fatalf("Failed to start consumer: %v", err)
	}
	defer consumer.Close()

	logrus.Infof("Started Kakfa consumer Broker: %v, Topic: %v, Function: %v, consumerID: %v", brokers, topic, funcName, consumerGroupID)

	// Consume messages, wait for signal to stopchan to exit
	defer close(stoppedchan)
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				logrus.Infof("Received Kafka message Partition: %d Offset: %d Key: %s Value: %s ", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				//forward msg to function
				clientset := utils.GetClient()
				err = sendMessage(clientset, funcName, ns, string(msg.Value))
				if err != nil {
					logrus.Errorf("Failed to send message to function: %v", err)
				}
				consumer.MarkOffset(msg, "")
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				logrus.Debugf("Rebalanced: %+v\n", ntf)
			}
		case err, more := <-consumer.Errors():
			if more {
				logrus.Fatalf("Error: %s\n", err.Error())
			}
		case <-stopchan:
			return
		}
	}
}

func sendMessage(clientset kubernetes.Interface, funcName, ns, msg string) error {
	svc, err := clientset.CoreV1().Services(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to find the service for function %s", funcName)
	}
	logrus.Infof("Sending message %s to function %s", msg, funcName)
	funcPort := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		funcPort = svc.Spec.Ports[0].Name
	}

	jsonStr := []byte(msg)
	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(jsonStr)).SetHeader("Content-Type", "application/json")
	req = req.AbsPath(svc.ObjectMeta.SelfLink + ":" + funcPort + "/proxy/")

	timestamp := time.Now().UTC()
	req.SetHeader("event-id", fmt.Sprintf("kafka-consumer-%s-%s-%s", funcName, ns, timestamp.Format(time.RFC3339Nano)))
	req.SetHeader("event-type", "application/json")
	req.SetHeader("event-time", timestamp.String())
	req.SetHeader("event-namespace", "kafkatriggers.kubeless.io")

	_, err = req.Do().Raw()
	if err != nil {
		//detect the request timeout case
		if strings.Contains(err.Error(), "status code 408") {
			return errors.New("Request timeout exceeded")
		}
		return err
	}

	logrus.Infof("Message has sent to function %s successfully", funcName)
	return nil
}

// CreateKafkaConsumer creates a goroutine that subscribes to Kafka topic
func CreateKafkaConsumer(triggerObjName, funcName, ns, topic string) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if !consumerM[consumerID] {
		logrus.Infof("Creating Kafka consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
		stopM[consumerID] = make(chan struct{})
		stoppedM[consumerID] = make(chan struct{})
		go createConsumerProcess(brokers, topic, funcName, ns, consumerID, stopM[consumerID], stoppedM[consumerID])
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
