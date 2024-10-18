package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

type AnthropicDriver struct {
	apiKey        string
	systemMessage string
}

func NewAnthropicDriver(apiKey string) *AnthropicDriver {
	return &AnthropicDriver{
		apiKey: apiKey,
	}
}

func (d *AnthropicDriver) SetSystemMessage(message string) {
	d.systemMessage = message
}

func (d *AnthropicDriver) GetResponse(prompt string, model string) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":                model,
		"prompt":               fmt.Sprintf("Human: %s\n\nAssistant: %s\n\nHuman: %s\n\nAssistant:", d.systemMessage, d.systemMessage, prompt),
		"max_tokens_to_sample": 300,
		"stop_sequences":       []string{"\n\nHuman:"},
		"temperature":          0.7,
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", d.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	completion, ok := result["completion"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return completion, nil
}

func GetAnthropicModels() []string {
	return []string{"claude-2", "claude-3-opus", "claude-3-sonnet"}
}
