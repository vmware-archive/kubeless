package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateKafkaConsumer(brokers, topics []string, funcName, ns, funcPort string) {
	// Init config
	config := cluster.NewConfig()

	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Init consumer, consume errors & messages
	// consumer is grouped and labeled by functionID to receive load-balanced messages
	// More details: https://kafka.apache.org/documentation/#intro_consumers
	functionID := funcName + "+" + ns + "+" + funcPort
	consumer, err := cluster.NewConsumer(strings.Split(brokers, ","), functionID, strings.Split(topics, ","), config)
	if err != nil {
		logrus.Fatalf("Failed to start consumer: %v", err)
	}
	defer consumer.Close()

	// Create signal channel
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// Consume all channels, wait for signal to exit
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
				err = sendMessage(clientset, functionID, string(msg.Value))
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
		case <-sigchan:
			return
		}
	}
}

func sendMessage(clientset kubernetes.Interface, funcID, msg string) error {
	s := strings.Split(funcID, "+")
	if len(s) != 3 {
		return fmt.Errorf("FunctionID isn't set correctly. It should be `funcName+namespace+funcPort`")
	}
	svc, err := clientset.CoreV1().Services(s[1]).Get(s[0], metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to find the service for function %s", s[0])
	}

	jsonStr := []byte(msg)
	req := clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(jsonStr)).SetHeader("Content-Type", "application/json")
	req = req.AbsPath(svc.ObjectMeta.SelfLink + ":" + s[2] + "/proxy/")

	_, err = req.Do().Raw()
	if err != nil {
		return err
	}

	logrus.Infof("Message has sent to function %s successfully", s[0])
	return nil
}

func main() {
	brokers := os.Getenv("BROKER_LIST")
	if brokers == "" {
		logrus.Fatalf("The comma separated list of brokers can't be empty. Please set it in env BROKER_LIST.")
	}

	topics := os.Getenv("TOPICS")
	if topics == "" {
		logrus.Fatalf("The comma separated list of topics can't be empty. Please set env TOPICS.")
	}

	funcName := os.Getenv("FUNCTION_NAME")
	if funcName == "" {
		logrus.Fatalf("FuncName can't be empty. Please set env FUNCTION_NAME")
	}

	ns := os.Getenv("FUNCTION_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	funcPort := os.Getenv("FUNCTION_PORT")
	if funcPort == "" {
		funcPort = "8080"
	}

	CreateKafkaConsumer(brokers, topics, funcName, ns, funcPort)
}
