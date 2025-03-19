package gpt

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"

	"io"
)

type GPTService interface {
	CreateChatCompletion(ctx context.Context, req models.ChatCompletionRequest) (*models.ChatCompletionResponse, error)
	// CreateChatCompletionStream(ctx context.Context, req models.ChatCompletionRequest, send chan<- models.ChatCompletionStreamResponse) error
}
type GPTImpl struct {
	apiKey string
}

func NewChatGPTService(apiKey string) GPTService {
	return &GPTImpl{
		apiKey: apiKey,
	}
}
