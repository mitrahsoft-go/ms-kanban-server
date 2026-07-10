package configs

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	HTTP     HTTPConfig
}

type DatabaseConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Name        string
	SSLMode     string
	AutoMigrate string
}

type HTTPConfig struct {
	Port string
}

func LoadEnv() *Config {

	// Load environment variables and populate the Config struct

	return &Config{
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST",""),
			Port:        getEnv("DB_PORT","5432"),
			Username:    getEnv("DB_USERNAME",""),
			Password:    getEnv("DB_PASSWORD",""),
			Name:        getEnv("DB_NAME",""),
			SSLMode:     getEnv("DB_SSL_MODE",""),
			AutoMigrate: getEnv("DB_AUTOMIGRATE","false"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT","6369"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
