package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

var (
	brokers = flag.String("broker", "172.16.0.26:9092", "list of address of brokers")
	topic   = flag.String("topic", "mytopic", "kafka topic")
)

func newKafkaWriter(brokerlist, topic string) *kafka.Writer {
	brokers := strings.Split(brokerlist, ",")
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func main() {
	flag.Parse()

	writer := newKafkaWriter(*brokers, *topic)
	defer writer.Close()

	fmt.Println("start producing!...")
	for i := 0; ; i++ {
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("Key-%d", i)),
			Value: []byte(fmt.Sprintf("Value-%d", i)),
		}
		err := writer.WriteMessages(context.Background(), msg)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}
