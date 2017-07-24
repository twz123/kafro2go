package rdkafka

import (
	"fmt"
	"os"

	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

func ReadFromTopic(brokers []string, topic string, partitions []int32, position int64, bufferSize int) (messages chan *kafka.Message, closeTopic func(), err error) {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(brokers, ","),
		"group.id":                        "kafro2go",
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create consumer")
	}

	messages = make(chan *kafka.Message, bufferSize)
	quit := make(chan bool, 1)

	fmt.Fprintf(os.Stderr, "Created Consumer %v\n", consumer)

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to subscribe")
	}

	closeTopic = func() {
		quit <- true
	}

	go func() {
		defer func() {
			fmt.Fprintf(os.Stderr, "Closing consumer\n")
			err := consumer.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to close consumer: %v\n", err)
			}
			fmt.Fprintf(os.Stderr, "Closing messages\n")
			close(messages)
		}()

		for {
			select {
			case ev := <-consumer.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					fmt.Fprintf(os.Stderr, "%% %v\n", e)
					consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					fmt.Fprintf(os.Stderr, "%% %v\n", e)
					consumer.Unassign()
				case *kafka.Message:
					messages <- e
				case kafka.PartitionEOF:
					fmt.Fprintf(os.Stderr, "%% Reached %v\n", e)
				case kafka.Error:
					fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
					return
				}
			case <-quit:
				return
			}
		}
	}()

	return
}
