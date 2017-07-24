package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/twz123/kafro2go/pkg/rdkafka"
	"github.com/twz123/kafro2go/pkg/schemaregistry"
)

const (
	xOK = iota
	xGeneralError
	xCLIUsage
)

func main() {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGTERM, os.Interrupt)

	code, msg := run(osSignals)
	if code != xOK {
		fmt.Fprintln(os.Stderr, msg)
		if code == xCLIUsage {
			flag.Usage()
		}
	}
	os.Exit(code)
}

func run(osSignals <-chan os.Signal) (int, string) {
	brokers := flag.String("brokers", "", "A comma separated list of Kafka brokers.")
	topic := flag.String("topic", "", "The name of the topic to be consumed.")
	partitions := flag.String("partitions", "", "A comma separated list of partitions to consume. Defaults to all if omitted.")
	offset := flag.String("position", "newest", "The start position inside the partitions. Either \"oldest\" or \"newest\".")
	bufferSize := flag.Int("buffer-size", 256, "The buffer size of the message channel.")
	registry := flag.String("registry", "", "Host and port of the Schema Registry.")
	flag.Parse()

	if *brokers == "" {
		return xCLIUsage, "No kafka brokers specified."
	}
	brokerList := strings.Split(*brokers, ",")

	if *topic == "" {
		return xCLIUsage, "No topic specified."
	}

	var partitionList []int32
	if *partitions != "" {
		p, err := toInt32Slice(*partitions)
		if err != nil {
			return xCLIUsage, fmt.Sprintf("Invalid partition list: %v", err)
		}
		partitionList = p
	}

	var initialOffset int64
	switch *offset {
	case "oldest":
		initialOffset = -2
	case "newest":
		initialOffset = -1
	default:
		return xCLIUsage, fmt.Sprintf("Invalid offset: %s.", *offset)
	}

	if *registry == "" {
		return xCLIUsage, "No Schema Registry specified."
	}

	messages, closeTopic, err := rdkafka.ReadFromTopic(brokerList, *topic, partitionList, initialOffset, *bufferSize)
	if err != nil {
		return xGeneralError, err.Error()
	}

	go func() {
		<-osSignals
		closeTopic()
	}()

	if err := printMessages(messages, schemaregistry.NewRegistry(*registry)); err != nil {
		closeTopic()
		for _ = range messages {
			// wait until channel is closed
		}
		return xGeneralError, fmt.Sprintf("Error during consumption: %v", err)
	}

	fmt.Println("Exiting main loop")
	return xOK, ""
}

func printMessages(messages <-chan *kafka.Message, registry schemaregistry.API) error {
	for msg := range messages {
		if len(msg.Value) < 5 {
			return fmt.Errorf("message is too short to extract a schema ID: %d bytes on partition %d at offset %d",
				len(msg.Value), msg.TopicPartition.Partition, msg.TopicPartition.Offset)
		}

		if msg.Value[0] != 0 {
			return fmt.Errorf("illegal magic byte on partition %d at offset %d: %d",
				int(msg.Value[0]), msg.TopicPartition.Partition, msg.TopicPartition.Offset)
		}

		// 4 bytes int32 big endian
		schemaID := ((int32(msg.Value[1]) << 24) + (int32(msg.Value[2]) << 16) + (int32(msg.Value[3]) << 8) + (int32(msg.Value[4]) << 0))

		schema, err := registry.GetSchemaByID(int64(schemaID))
		if err != nil {
			return errors.Wrapf(err, "failed to fetch schema with ID %d on partition %d at offset %d",
				schemaID, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
		}

		fmt.Printf("yep: %d: %s", schemaID, schema)
	}

	return nil
}

func toInt32Slice(csv string) (result []int32, err error) {
	for _, value := range strings.Split(csv, ",") {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, int32(parsed))
	}

	return
}
