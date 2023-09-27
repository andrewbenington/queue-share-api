package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	postgresHost string
	postgresPort string
	postgresUser string
	postgresPass string
	postgresDB   string

	spotifyClientID     string
	spotifyClientSecret string
	spotifyRedirectURL  string

	encryptionKey string
	signingSecret []byte
	env           string
}

var (
	config Config
)

func init() {
	signingSecret, err := base64.StdEncoding.DecodeString(os.Getenv("SIGNING_SECRET"))
	if err != nil {
		panic("can't decode signing secret")
	}
	config = Config{
		postgresHost: os.Getenv("POSTGRES_HOST"),
		postgresPort: os.Getenv("POSTGRES_PORT"),
		postgresUser: os.Getenv("POSTGRES_USER"),
		postgresPass: os.Getenv("POSTGRES_PASS"),
		postgresDB:   os.Getenv("POSTGRES_DB"),

		spotifyClientID:     os.Getenv("SPOTIFY_ID"),
		spotifyClientSecret: os.Getenv("SPOTIFY_SECRET"),
		spotifyRedirectURL:  os.Getenv("SPOTIFY_REDIRECT"),

		encryptionKey: os.Getenv("ENCRYPTION_KEY"),
		signingSecret: signingSecret,
		env:           os.Getenv("ENV"),
	}
	if config.env == "" {
		config.env = "LOCAL"
	}
}

func GetEncryptionKey() string {
	return config.encryptionKey
}

func GetSigningSecret() []byte {
	return config.signingSecret
}

func GetSpotifyRedirect() string {
	return config.spotifyRedirectURL
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
