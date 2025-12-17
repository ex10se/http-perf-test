package config

import (
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	RabbitMQURL string
	SocketPath  string
}

// Load читает конфигурацию из переменных окружения
// Паникует если обязательные переменные не заданы
func Load() *Config {
	return &Config{
		RabbitMQURL: getEnvRequired("DSN__RABBITMQ"),
		SocketPath:  getEnvRequired("SOCKET_PATH"),
	}
}

// getEnvRequired читает обязательную переменную окружения
// Паникует если переменная не задана
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s is required but not set", key))
	}
	return value
}
