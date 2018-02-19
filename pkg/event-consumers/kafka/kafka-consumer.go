package kafka

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createConsumerProcess(brokers, topics, funcName, ns, funcPort string, stopchan, stoppedchan chan struct{}) {
	// Init config
	config := cluster.NewConfig()

	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Init consumer, consume errors & messages
	// consumer is grouped and labeled by functionID to receive load-balanced messages
	// More details: https://kafka.apache.org/documentation/#intro_consumers
	functionID := funcName + "+" + ns
	consumer, err := cluster.NewConsumer(strings.Split(brokers, ","), functionID, strings.Split(topics, ","), config)
	if err != nil {
		logrus.Fatalf("Failed to start consumer: %v", err)
	}
	defer consumer.Close()

	// Consume messages, wait for signal to stopchan to exit
	defer close(stoppedchan)
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				//print to stdout
				//TODO: should be logrus.Debugf and enable verbosity
				fmt.Printf("Partition:\t%d\n", msg.Partition)
				fmt.Printf("Offset:\t%d\n", msg.Offset)
				fmt.Printf("Key:\t%s\n", string(msg.Key))
				fmt.Printf("Value:\t%s\n", string(msg.Value))
				fmt.Println()

				//forward msg to function
				clientset := utils.GetClient()
				err = sendMessage(clientset, funcName, ns, funcPort, string(msg.Value))
				if err != nil {
					logrus.Errorf("Failed to send message to function: %v", err)
				}
				consumer.MarkOffset(msg, "")
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				logrus.Infof("Rebalanced: %+v\n", ntf)
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

func sendMessage(clientset kubernetes.Interface, funcName, ns, funcPort, msg string) error {
	svc, err := clientset.CoreV1().Services(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to find the service for function %s", funcName)
	}

	jsonStr := []byte(msg)
	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(jsonStr)).SetHeader("Content-Type", "application/json")
	req = req.AbsPath(svc.ObjectMeta.SelfLink + ":" + funcPort + "/proxy/")

	_, err = req.Do().Raw()
	if err != nil {
		return err
	}

	logrus.Infof("Message has sent to function %s successfully", funcName)
	return nil
}

func CreateKafkaConsumer(stopM map[string](chan struct{}), stoppedM map[string](chan struct{}), brokers, topics, funcName, ns, funcPort string) {
	logrus.Infof("Creating consumer: broker %s - topic %s - function %s - namespace %s", brokers, topics, funcName, ns)
	funcID := funcName + "+" + ns
	stopM[funcID] = make(chan struct{})
	stoppedM[funcID] = make(chan struct{})

	// create consumer process
	go createConsumerProcess(brokers, topics, funcName, ns, funcPort, stopM[funcID], stoppedM[funcID])
	logrus.Infof("Created successfully.")
}

func DeleteKafkaConsumer(stopM map[string](chan struct{}), stoppedM map[string](chan struct{}), brokers, topics, funcName, ns string) {
	logrus.Infof("Stopping consumer: broker %s - topic %s - function %s - namespace %s", brokers, topics, funcName, ns)
	funcID := funcName + "+" + ns + "+"

	// delete consumer process
	close(stopM[funcID])
	<-stoppedM[funcID]
	logrus.Infof("Stopped successfully.")
}
