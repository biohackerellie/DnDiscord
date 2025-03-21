package discord

import (
	"log/slog"
	"sync"

	"github.com/biohackerellie/DnDiscord/internal/gpt"
	lru "github.com/hashicorp/golang-lru/v2"
)

type DiscordHandler struct {
	GptService      gpt.GPTService
	log             *slog.Logger
	AllowedChannels map[string]bool
	memoryMu        sync.Mutex
	StreamLocks     sync.Map
	ChannelMemory   *lru.Cache[any, any]
}

const (
	MaxChannelMemory = 100
)

func NewDiscordHandler(gpt gpt.GPTService, log *slog.Logger) *DiscordHandler {
	cache, _ := lru.New[any, any](MaxChannelMemory)
	return &DiscordHandler{
		GptService:    gpt,
		log:           log,
		ChannelMemory: cache,
	}
}
