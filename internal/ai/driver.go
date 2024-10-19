package ai

type AIDriver interface {
	SetSystemMessage(message string)
	GetResponse(prompt string, model string) (string, error)
}
