package config

import (
	"fmt"
	"os"
)

type Config struct {
	// Server
	HTTPPort string
	HTTPHost string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Kafka
	KafkaBrokers string

	// Logging
	Environment string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		HTTPHost:     getEnv("HTTP_HOST", "0.0.0.0"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "orderuser"),
		DBPassword:   getEnv("DB_PASSWORD", "orderpass123"),
		DBName:       getEnv("DB_NAME", "order_db"),
		DBSSLMode:    getEnv("DB_SSL_MODE", "disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092,localhost:9094"),
		Environment:  getEnv("ENV", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "INFO"),
	}
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func (c *Config) KafkaAddress() string {
	return c.KafkaBrokers
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
