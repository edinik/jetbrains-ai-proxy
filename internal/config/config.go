package config

import (
	"os"

	"github.com/joho/godotenv"
)

var JetbrainsAiConfig *Config

type Config struct {
	JetbrainsToken string
	BearerToken    string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	JetbrainsAiConfig = &Config{
		JetbrainsToken: os.Getenv("JWT_TOKEN"),
		BearerToken:    os.Getenv("BEARER_TOKEN"),
	}
	return JetbrainsAiConfig
}
