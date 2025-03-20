package discord

import (
	"log/slog"

	"github.com/biohackerellie/DnDiscord/internal/gpt"
)

type DiscordHandler struct {
	GptService gpt.GPTService
	log        *slog.Logger
}

func NewDiscordHandler(gpt gpt.GPTService, log *slog.Logger) *DiscordHandler {
	return &DiscordHandler{
		GptService: gpt,
		log:        log,
	}
}
