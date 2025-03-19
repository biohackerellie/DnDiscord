package main

import (
	"flag"
	"fmt"
	"github.com/biohackerellie/DnDiscord/config"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	envPath := flag.String("env", "", "path to .env")
	flag.Parse()
	if enPath := *envPath; enPath != "" {
		err := godotenv.Load(*envPath)
		if err != nil {
			fmt.Println("Error loading .env file")
		}
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config")
	}

	dg, err := discordgo.New("Bot " + cfg.BotToken)

}
