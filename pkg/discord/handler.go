package discord

import (
	"log/slog"
	"sync"

	"github.com/biohackerellie/DnDiscord/internal/gpt"
	"github.com/biohackerellie/DnDiscord/internal/models"
	lru "github.com/hashicorp/golang-lru/v2"
)

type DiscordHandler struct {
	GptService      gpt.GPTService
	log             *slog.Logger
	AllowedChannels map[string]bool
	memoryMu        sync.Mutex
	StreamLocks     sync.Map
	ChannelMemory   *lru.Cache[string, []models.ChatCompletionMessage]
}

const (
	MaxChannelMemory = 100
)

func NewDiscordHandler(gpt gpt.GPTService, log *slog.Logger, channelIds []string) *DiscordHandler {
	cache, _ := lru.New[string, []models.ChatCompletionMessage](MaxChannelMemory)
	allowedChannels := make(map[string]bool)
	for _, id := range channelIds {
		allowedChannels[id] = true
	}
	return &DiscordHandler{
		GptService:      gpt,
		log:             log,
		ChannelMemory:   cache,
		AllowedChannels: allowedChannels,
	}
}
