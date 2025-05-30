package ai

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

type OpenAIDriver struct {
	client        *openai.Client
	systemMessage string
}

func NewOpenAIDriver(apiKey string, apiUrl string) *OpenAIDriver {
	openAiClientCfg := openai.DefaultConfig(apiKey)
	if apiUrl != "" {
		openAiClientCfg.BaseURL = apiUrl
	}
	openAiClient := openai.NewClientWithConfig(openAiClientCfg)
	return &OpenAIDriver{
		client: openAiClient,
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
		log.Error().Err(err).Msg("Failed to create chat completion")
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func GetGptModels() ([]string, int) {
	models := []string{"gpt-3.5-turbo", "gpt-4o", "gpt-4o-mini"}
	defaultModelIndex := 2
	return models, defaultModelIndex
}
