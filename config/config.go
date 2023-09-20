package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	postgresHost  string
	postgresPort  string
	postgresUser  string
	postgresPass  string
	postgresDB    string
	encryptionKey string
	env           string
}

var (
	config Config
)

func init() {
	config = Config{
		postgresHost:  os.Getenv("POSTGRES_HOST"),
		postgresPort:  os.Getenv("POSTGRES_PORT"),
		postgresUser:  os.Getenv("POSTGRES_USER"),
		postgresPass:  os.Getenv("POSTGRES_PASS"),
		postgresDB:    os.Getenv("POSTGRES_DB"),
		encryptionKey: os.Getenv("ENCRYPTION_KEY"),
		env:           os.Getenv("ENV"),
	}
	if config.env == "" {
		config.env = "LOCAL"
	}
}

func GetEncryptionKey() string {
	return config.encryptionKey
}

func GetDBString() string {
	switch strings.ToUpper(config.env) {
	case "LOCAL":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s connect_timeout=5 sslmode=disable",
			config.postgresHost, config.postgresPort, config.postgresUser, config.postgresPass, config.postgresDB)
	default:
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s connect_timeout=5",
			config.postgresHost, config.postgresPort, config.postgresUser, config.postgresPass, config.postgresDB)
	}
}
