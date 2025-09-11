package config

import (
	"os"
	"subscriptions-service/internal/logger"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ServerAddress string
	LogLevel      string
}

func LoadConfig() Config {
	cfg := Config{
		DBHost:        getEnv("DB_HOST"),
		DBPort:        getEnv("DB_PORT"),
		DBUser:        getEnv("DB_USER"),
		DBPassword:    getEnv("DB_PASSWORD"),
		DBName:        getEnv("DB_NAME"),
		ServerAddress: getEnv("SERVER_ADDRESS"),
		LogLevel:      getEnv("LOG_LEVEL"),
	}

	logger.Log.Info("Загружена конфигурация", "config", cfg)
	return cfg
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	logger.Log.Error("Не установлена переменная окружения", "var", key)
	panic("Do not set enviroment variable")
}
