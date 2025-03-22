package config

import (
	"os"
)

type Config struct {
	OpenaiAPIKey    string
	DiscordClientID string
	BotToken        string
	LogLevel        string
	Env             string
	AllowedChannels string
	GPT_MODEL       string
}

func Load() *Config {
	return &Config{
		OpenaiAPIKey:    os.Getenv("OPENAI_API_KEY"),
		DiscordClientID: os.Getenv("DISCORD_CLIENT_ID"),
		BotToken:        os.Getenv("BOT_TOKEN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		Env:             os.Getenv("APP_ENV"),
		AllowedChannels: os.Getenv("ALLOWED_CHANNELS"),
		GPT_MODEL:       os.Getenv("GPT_MODEL"),
	}
}
