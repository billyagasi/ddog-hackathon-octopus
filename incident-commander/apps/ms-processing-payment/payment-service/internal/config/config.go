package config

import "os"

type Config struct {
	// Datadog
	DDEnv     string
	DDService string
	DDVersion string

	// PostgreSQL
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Kafka
	KafkaBroker string

	// Elasticsearch
	ElasticsearchURL      string
	ElasticsearchUser      string
	ElasticsearchPassword string

	// Payment Gateway
	PaymentGatewayURL string
}

func Load() Config {
	return Config{
		DDEnv:     getEnv("DD_ENV", "hackathon"),
		DDService: getEnv("DD_SERVICE", "payment-service"),
		DDVersion: getEnv("DD_VERSION", "1.0.0"),

		PostgresHost:     getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "payments"),

		KafkaBroker: getEnv("KAFKA_BROKER", "kafka:9092"),

		ElasticsearchURL:      getEnv("ELASTICSEARCH_URL", "http://elasticsearch:9200"),
		ElasticsearchUser:      getEnv("ELASTICSEARCH_USER", "elastic"),
		ElasticsearchPassword: getEnv("ELASTICSEARCH_PASSWORD", "elastic"),

		PaymentGatewayURL: getEnv("PAYMENT_GATEWAY_URL", "http://payment-gateway-mock:8081"),
	}
}

func (c Config) PostgresDSN() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword +
		"@" + c.PostgresHost + ":" + c.PostgresPort +
		"/" + c.PostgresDB + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
