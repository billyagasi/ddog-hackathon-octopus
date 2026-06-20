package infra

import (
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/config"
)

func NewKafkaReader(cfg config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    "inventory-reserved",
		GroupID:  "payment-service",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

func NewKafkaWriter(cfg config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBroker),
		Topic:    "payment-completed",
		Balancer: &kafka.LeastBytes{},
	}
}

func EnsureTopics(cfg config.Config) error {
	log.Println("[infra] ensuring Kafka topics exist (auto-create enabled)")

	topics := []string{"order-created", "inventory-reserved", "payment-completed"}
	for _, topic := range topics {
		log.Printf("[infra] topic expected: %s", topic)
	}

	return nil
}
