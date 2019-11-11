package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	kafka "github.com/segmentio/kafka-go"
)

var (
	brokers = flag.String("brokers", "172.16.0.26:9092", "list of address of brokers")
	topic   = flag.String("topic", "mytopic", "kafka topic")
	groupID = flag.String("group", "", "id of consumer group")
)

func getKafkaReader(brokerlist, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(brokerlist, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func main() {
	flag.Parse()

	reader := getKafkaReader(*brokers, *topic, *groupID)
	defer reader.Close()

	fmt.Println("start consuming!...")
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("message at topic: %v, partition: %v, offset: %v, %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}
