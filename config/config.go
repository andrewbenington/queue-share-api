package config

import (
	"fmt"
	"os"
)

type Config struct {
	PostgresHost string
	PostgresPort string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
}

func GetConfig() Config {
	return Config{
		PostgresHost: os.Getenv("POSTGRES_HOST"),
		PostgresPort: os.Getenv("POSTGRES_PORT"),
		PostgresUser: os.Getenv("POSTGRES_USER"),
		PostgresPass: os.Getenv("POSTGRES_PASS"),
		PostgresDB:   os.Getenv("POSTGRES_DB"),
	}
}

func (c *Config) GetDBString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s connect_timeout=5",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPass, c.PostgresDB)
}
