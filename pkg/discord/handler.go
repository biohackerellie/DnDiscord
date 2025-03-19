package discord

import (
	"log/slog"

	"github.com/biohackerellie/DnDiscord/ports"
)

type DiscordHandler struct {
	GptService ports.GPTService
	log        *slog.Logger
}

func NewDiscordHandler(gpt ports.GPTService, log *slog.Logger) *DiscordHandler {
	return &DiscordHandler{
		GptService: gpt,
		log:        log,
	}
}
