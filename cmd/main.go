package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/biohackerellie/DnDiscord/config"

	"github.com/biohackerellie/DnDiscord/internal/gpt"

	"github.com/biohackerellie/DnDiscord/internal/lib/logger"
	"github.com/biohackerellie/DnDiscord/internal/lib/ref"
	"github.com/biohackerellie/DnDiscord/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:         discord.ApplicationCommandChat,
			Description:  "Create a new thread for conversation.",
			DMPermission: ref.Of(true),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "messages",
					Description: "Message to send",
					Required:    true,
				},
			},
		},
	}
)

var log *slog.Logger

func main() {
	envPath := flag.String("env", "", "path to .env")
	flag.Parse()
	if enPath := *envPath; enPath != "" {
		err := godotenv.Load(*envPath)
		if err != nil {
			fmt.Println("Error loading .env file")
		}
	}

	cfg := config.Load()
	logLevel := cfg.LogLevel
	local := (cfg.Env == "development")
	channels := strings.Split(cfg.AllowedChannels, ",")
	logOptions := logger.LogOptions(strings.TrimSpace(strings.ToLower(logLevel)), true, local)
	if local {
		log = slog.New(slog.NewTextHandler(os.Stdout, logOptions))
	} else {
		log = slog.New(slog.NewJSONHandler(os.Stdout, logOptions))
	}
	chatGPTService := gpt.NewChatGPTService(cfg.OpenaiAPIKey, cfg.GPT_MODEL)
	discordHandler := discord.NewDiscordHandler(chatGPTService, log, channels)

	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Error("error creating Discord session", "error", err)
		return
	}

	app, err := dg.Application("@me")
	if err != nil {
		log.Error("error getting application", "error", err)
		return
	}

	log.Info("Adding commands...")

	for _, v := range commands {
		_, err = dg.ApplicationCommandCreate(app.ID, "", v)
		if err != nil {
			panic(err)
		}
	}

	dg.AddHandler(discordHandler.GetInteractionCreateHandler())
	dg.AddHandler(discordHandler.GetMessageCreateHandler())

	err = dg.Open()
	if err != nil {
		log.Error("error opening connection", "error", err)
		return
	}
	defer dg.Close()

	botInviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%s&scope=%s",
		cfg.DiscordClientID, "328565073920", "bot")
	log.Info("invite bot to your server: ", "url", botInviteURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	log.Info("Bot is running, press CTRL-C to exit.")
	<-stop

	log.Info("Gracefully shutting down.")
	registeredCommands, err := dg.ApplicationCommands(app.ID, "")
	if err != nil {
		panic("Could not fetch registered commands: " + err.Error())
	}
	for _, v := range registeredCommands {
		err = dg.ApplicationCommandDelete(app.ID, "", v.ID)
		if err != nil {
			panic(err)
		}
	}
	log.Info("Bot shutting down. Bye bye!")
}
