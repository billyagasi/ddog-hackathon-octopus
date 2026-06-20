package infra

import (
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/config"
)

func NewKafkaWriterProducer(cfg config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBroker),
		Topic:    "order-created",
		Balancer: &kafka.LeastBytes{},
	}
}

func NewKafkaReaderPaymentCompleted(cfg config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    "payment-completed",
		GroupID:  "order-service",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
}

func EnsureTopics(cfg config.Config) {
	log.Println("[infra] ensuring Kafka topics (auto-create enabled)")
	topics := []string{"order-created", "inventory-reserved", "payment-completed"}
	for _, t := range topics {
		log.Printf("[infra] topic expected: %s", t)
	}
}
