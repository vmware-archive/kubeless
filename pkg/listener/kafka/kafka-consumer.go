package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/sirupsen/logrus"
)

func main() {
	brokers := os.Getenv("BROKER_LIST")
	if brokers == "" {
		logrus.Fatalf("The comma separated list of brokers can't be empty. Please set it in env BROKER_LIST.")
	}

	topics := os.Getenv("TOPICS")
	if topics == "" {
		logrus.Fatalf("The comma separated list of topics can't be empty. Please set env TOPICS.")
	}

	functionID := os.Getenv("FUNCTION_ID")
	if functionID == "" {
		logrus.Fatalf("Function_ID is string concatenation of `function_name + namespace`, and it can't be empty. Please set env FUNCTION_ID")
	}

	// Init config
	config := cluster.NewConfig()

	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Init consumer, consume errors & messages
	// consumer is grouped and labeled by functionID to receive load-balanced messages
	// More details: https://kafka.apache.org/documentation/#intro_consumers
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
				fmt.Printf("Partition:\t%d\n", msg.Partition)
				fmt.Printf("Offset:\t%d\n", msg.Offset)
				fmt.Printf("Key:\t%s\n", string(msg.Key))
				fmt.Printf("Value:\t%s\n", string(msg.Value))
				fmt.Println()
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
