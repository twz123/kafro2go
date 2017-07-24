package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/twz123/kafro2go/pkg/rdkafka"
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

	messages, closeTopic, err := rdkafka.ReadFromTopic(brokerList, *topic, partitionList, initialOffset, *bufferSize)
	if err != nil {
		return xGeneralError, err.Error()
	}

	go func() {
		<-osSignals
		closeTopic()
	}()

	for msg := range messages {
		fmt.Printf("%v\n", msg)
		fmt.Printf("Partition:\t%d\n", msg.TopicPartition.Partition)
		fmt.Printf("Offset:\t%d\n", msg.TopicPartition.Offset)
		fmt.Printf("Key:\t%s\n", string(msg.Key))
		fmt.Printf("Value:\t%s\n", string(msg.Value))
		fmt.Println()
	}

	fmt.Println("Exiting main loop")
	return xOK, ""
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
