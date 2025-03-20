package gpt

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/biohackerellie/DnDiscord/internal/models"
	"github.com/sashabaranov/go-openai"
)

type GPTService interface {
	CreateChatCompletion(ctx context.Context, req models.ChatCompletionRequest) (*models.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req models.ChatCompletionRequest, send chan<- models.ChatCompletionStreamResponse) error
}
type GPTImpl struct {
	apiKey string
}

func NewChatGPTService(apiKey string) GPTService {
	return &GPTImpl{
		apiKey: apiKey,
	}
}

func (s *GPTImpl) CreateChatCompletionStream(ctx context.Context,
	req models.ChatCompletionRequest, send chan<- models.ChatCompletionStreamResponse) error {
	defer close(send)

	c := openai.NewClient(s.apiKey)

	chatReq := openai.ChatCompletionRequest{
		Model:    openai.O3Mini,
		Stream:   true,
		Messages: getChatCompletionMessages(req.Messages),
	}
	stream, err := c.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return errors.New(fmt.Sprintf("chat completion error: %v\n", err))
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.New(fmt.Sprintf("stream error: %v\n", err))
		}
		choices := make([]models.ChatCompletionStreamChoice, 0, len(response.Choices))
		for _, choice := range response.Choices {
			choices = append(choices, models.ChatCompletionStreamChoice{
				Index: choice.Index,
				Delta: models.ChatCompletionStreamChoiceDelta{
					Content: choice.Delta.Content,
				},
				FinishReason: string(choice.FinishReason),
			})
		}
		if len(choices) > 0 {
			send <- models.ChatCompletionStreamResponse{
				ID:      response.ID,
				Object:  response.Object,
				Created: response.Created,
				Model:   response.Model,
				Choices: choices,
			}
		}
	}
}

func (s *GPTImpl) CreateChatCompletion(ctx context.Context, req models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	c := openai.NewClient(s.apiKey)
	resp, err := c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.O3Mini,
			Messages: getChatCompletionMessages(req.Messages),
		},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("chat completion error: %v\n", err))
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New(fmt.Sprintf("empty choices for chat completion"))
	}

	return &models.ChatCompletionResponse{Content: resp.Choices[0].Message.Content}, nil
}

func getChatCompletionMessages(messages []models.ChatCompletionMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range messages {
		result = append(result, openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}
	return result
}
