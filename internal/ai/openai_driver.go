package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIDriver struct {
	client        *openai.Client
	systemMessage string
}

func NewOpenAIDriver(apiKey string) *OpenAIDriver {
	return &OpenAIDriver{
		client: openai.NewClient(apiKey),
	}
}

func (d *OpenAIDriver) SetSystemMessage(message string) {
	d.systemMessage = message
}

func (d *OpenAIDriver) GetResponse(prompt string, model string) (string, error) {
	resp, err := d.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: d.systemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func GetOpenAiModels() []string {
	return []string{"gpt-3.5-turbo", "gpt-4o", "gpt-4o-mini"}
}
