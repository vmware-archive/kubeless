package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

func main() {
	brokers := os.Getenv("BROKER_LIST")
	if brokers == "" {
		logrus.Fatalf("broker can't be empty. Please set env BROKER_LIST.")
	}
	brokerList := strings.Split(brokers, ",")
	topic := os.Getenv("TOPIC")
	if topic == "" {
		logrus.Fatalf("topic can't be empty. Please set env TOPIC.")
	}
	partition := os.Getenv("PARTITION")
	if partition == "" {
		partition = "0"
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	master, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := master.Close(); err != nil {
			logrus.Fatal(err)
		}
	}()
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		logrus.Fatal(err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				logrus.Fatal(err)
			case msg := <-consumer.Messages():
				logrus.Fatalf("Received messages", string(msg.Key), string(msg.Value))
			case <-signals:
				logrus.Fatalf("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
}
