package config

import "os"

type Config struct {
	DDEnv     string
	DDService string
	DDVersion string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	RedisHost string
	RedisPort string

	KafkaBroker string
}

func Load() Config {
	return Config{
		DDEnv:     getEnv("DD_ENV", "hackathon"),
		DDService: getEnv("DD_SERVICE", "inventory-service"),
		DDVersion: getEnv("DD_VERSION", "1.0.0"),

		PostgresHost:     getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "payments"),

		RedisHost: getEnv("REDIS_HOST", "redis"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		KafkaBroker: getEnv("KAFKA_BROKER", "kafka:9092"),
	}
}

func (c Config) PostgresDSN() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword +
		"@" + c.PostgresHost + ":" + c.PostgresPort +
		"/" + c.PostgresDB + "?sslmode=disable"
}

func (c Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
