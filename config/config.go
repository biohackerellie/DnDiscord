package config

import (
	"os"
)

type Config struct {
	OpenaiAPIKey    string
	DiscordClientID string
	BotToken        string
}

func Load() *Config {
	return &Config{
		OpenaiAPIKey:    os.Getenv("OPENAI_API_KEY"),
		DiscordClientID: os.Getenv("DISCORD_CLIENT_ID"),
		BotToken:        os.Getenv("BOT_TOKEN"),
	}
}
