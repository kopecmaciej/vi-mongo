package ai

type AIDriver interface {
	SetSystemMessage(message string)
	GetResponse(prompt string, model string) (string, error)
}

func GetAiModels() ([]string, int) {
	openaiModels, openaiDefaultModel := GetGptModels()
	anthropicModels, _ := GetAnthropicModels()
	return append(openaiModels, anthropicModels...), openaiDefaultModel
}
