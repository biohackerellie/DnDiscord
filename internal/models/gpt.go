package models

type ChatMessageRole string

const (
	ChatMessageRoleSystem ChatMessageRole = "system"
	ChatMessageRoleUser   ChatMessageRole = "user"
	ChatMessageRoleBot    ChatMessageRole = "bot"
)

type ChatCompletionMessage struct {
	Role    ChatMessageRole
	Content string
}

type ChatCompletionRequest struct {
	Messages []ChatCompletionMessage
}

type ChatCompletionResponse struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason"`
}

type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}
