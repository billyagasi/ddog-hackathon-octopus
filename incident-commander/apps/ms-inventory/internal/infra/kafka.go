package infra

import (
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/config"
)

func NewKafkaReaderOrderCreated(cfg config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    "order-created",
		GroupID:  "inventory-service",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

func NewKafkaWriterProducer(cfg config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBroker),
		Topic:    "inventory-reserved",
		Balancer: &kafka.LeastBytes{},
	}
}

func EnsureTopics(cfg config.Config) {
	log.Println("[infra] ensuring Kafka topics (auto-create enabled)")
	topics := []string{"order-created", "inventory-reserved", "payment-completed"}
	for _, t := range topics {
		log.Printf("[infra] topic expected: %s", t)
	}
}
