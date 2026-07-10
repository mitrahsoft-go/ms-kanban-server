package configs

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	HTTP     HTTPConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	SSLMode  string
}

type HTTPConfig struct {
	Port string
}

func LoadEnv() *Config {

	// Load environment variables and populate the Config struct

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			Username: getEnv("DB_USERNAME"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSL_MODE"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT"),
		},
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}
