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
		"model":      model,
		"max_tokens": 1024,
		"system":     d.systemMessage,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
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
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(result.Content) == 0 || result.Content[0].Text == "" {
		return "", fmt.Errorf("unexpected response format")
	}

	return result.Content[0].Text, nil
}

func GetAnthropicModels() ([]string, int) {
	models := []string{"claude-3-opus-20240229", "claude-3-haiku-20240307", "claude-3-sonnet-20240229"}
	defaultModelIndex := 2
	return models, defaultModelIndex
}
