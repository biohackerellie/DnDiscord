package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenaiAPIKey    string
	DiscordClientID string
	BotToken        string
}

func Load() (*Config, error) {
	development := os.Getenv("APP_ENV") == "development"
	if development {
		err := godotenv.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}
	return &Config{
		OpenaiAPIKey:    os.Getenv("OPENAI_API_KEY"),
		DiscordClientID: os.Getenv("DISCORD_CLIENT_ID"),
		BotToken:        os.Getenv("BOT_TOKEN"),
	}, nil
}
