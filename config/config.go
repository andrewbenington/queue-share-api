package config

import (
	"fmt"
	"os"
)

type Config struct {
	PostgresHost   string
	PostgresPort   string
	PostgresUser   string
	PostgresPass   string
	PostgresDB     string
	PostgresSocket string
	IsGAE          bool
}

func GetConfig() Config {
	return Config{
		PostgresHost:   os.Getenv("POSTGRES_HOST"),
		PostgresPort:   os.Getenv("POSTGRES_PORT"),
		PostgresUser:   os.Getenv("POSTGRES_USER"),
		PostgresPass:   os.Getenv("POSTGRES_PASS"),
		PostgresDB:     os.Getenv("POSTGRES_DB"),
		PostgresSocket: os.Getenv("POSTGRES_SOCKET"),
		IsGAE:          os.Getenv("GAE_APPLICATION") != "",
	}
}

func (c *Config) GetDBString() string {
	if c.IsGAE {
		return fmt.Sprintf("%s:%s@unix(%s)/%s",
			c.PostgresUser, c.PostgresPass, c.PostgresSocket, c.PostgresDB)
	} else {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
			c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPass, c.PostgresDB)
	}
}
