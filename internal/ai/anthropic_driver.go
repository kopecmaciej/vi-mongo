package ai

import (
    "context"
    "fmt"
    "github.com/anthdm/anthropic-go"
)

type AnthropicDriver struct {
    client        *anthropic.Client
    systemMessage string
}

func NewAnthropicDriver(apiKey string) *AnthropicDriver {
    return &AnthropicDriver{
        client: anthropic.NewClient(apiKey),
    }
}

func (d *AnthropicDriver) SetSystemMessage(message string) {
    d.systemMessage = message
}

func (d *AnthropicDriver) GetResponse(prompt string) (string, error) {
    resp, err := d.client.Complete(
        context.Background(),
        &anthropic.CompletionRequest{
            Model: anthropic.ClaudeInstant,
            Prompt: fmt.Sprintf("%s\n\n%s", d.systemMessage, prompt),
            MaxTokensToSample: 300,
        },
    )

    if err != nil {
        return "", fmt.Errorf("completion failed: %w", err)
    }

    return resp.Completion, nil
}
