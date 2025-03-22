package discord

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/biohackerellie/DnDiscord/internal/models"
	"github.com/bwmarrin/discordgo"
)

const (
	ApplicationCommandChat string = "chat"
)

const (
	MaxMessageLength int = 2000
)

func (h *DiscordHandler) GetInteractionCreateHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		switch data.Name {
		case ApplicationCommandChat:
			h.handleChatCommand(s, i)
		}
	}
}

func (h *DiscordHandler) handleChannelMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Prevent concurrent processing
	if _, loaded := h.StreamLocks.LoadOrStore(m.ChannelID, true); loaded {
		s.ChannelMessageSend(m.ChannelID, "⌛ I'm already thinking here!")
		return
	}
	defer h.StreamLocks.Delete(m.ChannelID)

	history := h.getChannelHistory(m.ChannelID)
	history = append(history, models.ChatCompletionMessage{
		Role:    models.ChatMessageRoleUser,
		Content: fmt.Sprintf("%s: %s", m.Author.GlobalName, m.Content),
	})

	recv := make(chan models.ChatCompletionStreamResponse)
	go func() {
		err := h.GptService.CreateChatCompletionStream(context.Background(),
			models.ChatCompletionRequest{Messages: history},
			recv,
		)
		if err != nil {
			h.log.Error("failed to create chat completion stream", "error", err)
			close(recv)
		}
	}()

	channelTyping := func() {
		err := s.ChannelTyping(m.ChannelID)
		if err != nil {
			h.log.Error("failed to send typing", "error", err)
		}
	}

	channelTyping()
	var msgID string
	var chunkToRead string
	var readChunkLength int
	const (
		maxContentLengthPerChunk = 2000
		intervalOfCharacters     = 100
	)
	for resp := range recv {
		if len(resp.Choices) <= 0 {
			continue
		}
		incomingContent := resp.Choices[0].Delta.Content
		if incomingContent == "" {
			continue
		}

		chunkToRead += incomingContent
		prevWordDividerIndex := -1
		for i, w := readChunkLength, 0; i < len(chunkToRead); i += w {
			char, width := utf8.DecodeRuneInString(chunkToRead[i:])
			w = width
			readChunkLength += width
			if readChunkLength > maxContentLengthPerChunk {
				chunkEndingIdx := h.getChunkEndingIndex(prevWordDividerIndex, i, char)
				h.sendMessage(s, m.ChannelID, msgID, chunkToRead[:chunkEndingIdx])
				chunkToRead = chunkToRead[chunkEndingIdx:]
				msgID = ""
				readChunkLength = 0
				channelTyping()
				break
			}
			if IsWordDivider(char) {
				prevWordDividerIndex = i
			}
			if readChunkLength%intervalOfCharacters == 0 {
				msgID = h.sendMessage(s, m.ChannelID, msgID, chunkToRead[:i+w])
				channelTyping()
			}
		}
	}
	if chunkToRead != "" {
		h.sendMessage(s, m.ChannelID, msgID, chunkToRead)
	}

	history = append(history, models.ChatCompletionMessage{
		Role:    models.ChatMessageRoleAssistant,
		Content: chunkToRead,
	})
	h.updateChannelHistory(m.ChannelID, history)
}

func (h *DiscordHandler) handleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if data.Name != ApplicationCommandChat {
		return
	}

	if len(data.Options) <= 0 {
		h.log.Warn("empty options")
		return
	}
	content, ok := data.Options[0].Value.(string)
	if !ok {
		_, err := s.ChannelMessageSend(i.ChannelID, "Invalid input loser")
		if err != nil {
			h.log.Error("failed to send message", "error", err)
		}
		return
	}

	var author string
	if i.User != nil {
		author = i.User.ID
	}
	if i.Member != nil {
		if i.Member.User != nil {
			author = i.Member.User.ID
		}
	}
	response := getUserMessage(content, author)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		h.log.Error("failed to respond to interaction", "error", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		})
		return
	}

	req := models.ChatCompletionRequest{Messages: []models.ChatCompletionMessage{
		{
			Role:    models.ChatMessageRoleUser,
			Content: content,
		},
	}}
	resp, err := h.GptService.CreateChatCompletion(context.Background(), req)
	if err != nil {
		h.log.Error("failed to create chat completion", "error", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		})
		return
	}

	response += resp.Content

	chunks := make(chan string)
	go SendMessageByChunk(response, MaxMessageLength, chunks)
	for chunk := range chunks {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: chunk,
		})
		if err != nil {
			h.log.Error("failed to send message", "error", err)
		}
	}
}

func (h *DiscordHandler) GetMessageCreateHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Interaction != nil && m.Interaction.ID != "" {
			return
		}
		if m.GuildID != "" && !h.AllowedChannels[m.ChannelID] {
			h.log.Debug("channel not allowed", "channel", m.ChannelID)
			return
		}
		if m.GuildID == "" {

			h.handleDirectMessage(s, m)
			return
		}
		h.handleChannelMessage(s, m)
	}
}
func (h *DiscordHandler) handleDirectMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		h.log.Error("failed to create user channel", "error", err)

		s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the message! OUCH!!",
		)

		return
	}
	h.log.Info("message recieved", "message", m.Content, "author", m.Author.ID)
	req := models.ChatCompletionRequest{Messages: []models.ChatCompletionMessage{
		{
			Role:    models.ChatMessageRoleUser,
			Content: m.Content,
		},
	}}
	recv := make(chan models.ChatCompletionStreamResponse)

	go func() {
		err = h.GptService.CreateChatCompletionStream(context.Background(), req, recv)
		if err != nil {
			h.log.Error("failed to create chat completion stream", "error", err)
		}
	}()

	channelTyping := func() {
		err = s.ChannelTyping(channel.ID)
		if err != nil {
			h.log.Error("failed to send typing", "error", err)
		}
	}

	channelTyping()

	var msgID string
	var chunkToRead string
	var readChunkLength int
	const (
		maxContentLengthPerChunk = 2000
		intervalOfCharacters     = 100
	)
	for resp := range recv {
		if len(resp.Choices) <= 0 {
			continue
		}
		incomingContent := resp.Choices[0].Delta.Content
		if incomingContent == "" {
			continue
		}

		chunkToRead += incomingContent
		prevWordDividerIndex := -1
		for i, w := readChunkLength, 0; i < len(chunkToRead); i += w {
			char, width := utf8.DecodeRuneInString(chunkToRead[i:])
			w = width
			readChunkLength += width
			if readChunkLength > maxContentLengthPerChunk {
				chunkEndingIdx := h.getChunkEndingIndex(prevWordDividerIndex, i, char)
				h.sendMessage(s, m.ChannelID, msgID, chunkToRead[:chunkEndingIdx])
				chunkToRead = chunkToRead[chunkEndingIdx:]
				msgID = ""
				readChunkLength = 0
				channelTyping()
				break
			}
			if IsWordDivider(char) {
				prevWordDividerIndex = i
			}
			if readChunkLength%intervalOfCharacters == 0 {
				msgID = h.sendMessage(s, m.ChannelID, msgID, chunkToRead[:i+w])
				channelTyping()
			}
		}
	}
	if chunkToRead != "" {
		h.sendMessage(s, m.ChannelID, msgID, chunkToRead)
	}
}

func (h *DiscordHandler) getChunkEndingIndex(prevWordDividerIdx, currentIdx int, currentChar rune) int {
	if prevWordDividerIdx != -1 && prevWordDividerIdx != currentIdx-1 && !IsWordDivider(currentChar) {
		return prevWordDividerIdx + 1
	}
	return currentIdx
}

func (h *DiscordHandler) getChannelHistory(channelID string) []models.ChatCompletionMessage {
	h.memoryMu.Lock()
	defer h.memoryMu.Unlock()

	if history, ok := h.ChannelMemory.Get(channelID); ok {
		return history
	}
	return []models.ChatCompletionMessage{}
}

func (h *DiscordHandler) updateChannelHistory(channelID string, history []models.ChatCompletionMessage) {
	h.memoryMu.Lock()
	defer h.memoryMu.Unlock()
	h.ChannelMemory.Add(channelID, history)
}

func (h *DiscordHandler) clearChannelHistory(channelID string) {
	h.memoryMu.Lock()
	defer h.memoryMu.Unlock()

	h.ChannelMemory.Remove(channelID)
}

func (h *DiscordHandler) sendMessage(s *discordgo.Session, channelId, messageId, content string) string {
	formatted := fmt.Sprintf("🤖 **Assistant**\n%s", content)
	if messageId == "" {
		_msg, err := s.ChannelMessageSend(channelId, formatted)
		if err != nil {
			h.log.Error("failed to send message", "error", err)
		}
		return _msg.ID
	}
	_msg, err := s.ChannelMessageEdit(channelId, messageId, content)
	if err != nil {
		h.log.Error("failed to edit message", "error", err)
	}
	return _msg.ID
}
func formatBotResponse(originalContent string, currentUser string) string {
	if strings.Contains(originalContent, "(To ") {
		return originalContent
	}
	return fmt.Sprintf("**%s**\n%s", currentUser, originalContent)
}
