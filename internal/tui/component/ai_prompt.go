package component

import (
	"fmt"
	"os"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/ai"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptTabID = "AIPromptTab"
)

type AIPrompt struct {
	*core.BaseElement
	*core.Form

	responseArea    *tview.TextView
	openaiDriver    *ai.OpenAIDriver
	anthropicDriver *ai.AnthropicDriver
}

func NewAIPrompt() *AIPrompt {
	a := &AIPrompt{
		BaseElement:     core.NewBaseElement(),
		Form:            core.NewForm(),
		openaiDriver:    ai.NewOpenAIDriver(os.Getenv("OPENAI_API_KEY")),
		anthropicDriver: ai.NewAnthropicDriver(os.Getenv("ANTHROPIC_API_KEY")),
	}

	a.SetIdentifier(AIPromptTabID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPrompt) init() error {
	a.setupComponents()
	a.setStyle()

	return nil
}

func (a *AIPrompt) setupComponents() {
	openaiModels := a.openaiDriver.GetModels()
	anthropicModels := a.anthropicDriver.GetModels()

	a.Form.
		AddDropDown("Model:", append(openaiModels, anthropicModels...), 0, nil).
		AddTextArea("Prompt:", "", 0, 3, 0, nil).
		AddButton("Submit", a.onSubmit)

	a.responseArea = tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			a.App.Draw()
		})

	a.Form.AddFormItem(a.responseArea)
}

func (a *AIPrompt) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)

	a.Form.SetButtonBackgroundColor(styles.AIPrompt.ButtonBackgroundColor.Color())
	a.Form.SetButtonTextColor(styles.AIPrompt.ButtonTextColor.Color())
	a.Form.SetLabelColor(styles.AIPrompt.LabelColor.Color())
	a.Form.SetFieldBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())
	a.Form.SetFieldTextColor(styles.AIPrompt.InputTextColor.Color())

	a.responseArea.SetBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())
	a.responseArea.SetTextColor(styles.AIPrompt.InputTextColor.Color())
}

func (a *AIPrompt) onSubmit() {
	var driver ai.AIDriver

	_, model := a.Form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
	prompt := a.Form.GetFormItem(1).(*tview.TextArea).GetText()

	switch model {
	case "OpenAI":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			a.showError("OpenAI API key not found in environment variables")
			return
		}
		driver = ai.NewOpenAIDriver(apiKey)
	case "Anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			a.showError("Anthropic API key not found in environment variables")
			return
		}
		driver = ai.NewAnthropicDriver(apiKey)
	default:
		a.showError("Invalid AI model selected")
		return
	}

	systemMessage := `This prompt is for a query to MongoDB using the Query Bar. Example: { name: { $regex: "^catelyn", "$options": "i" } }`
	driver.SetSystemMessage(systemMessage)

	response, err := driver.GetResponse(prompt, model)
	if err != nil {
		a.showError(fmt.Sprintf("Error getting response: %v", err))
		return
	}

	a.showResponse(response)
}

func (a *AIPrompt) showError(message string) {
	a.responseArea.SetText(fmt.Sprintf("[red]Error: %s[-]", message))
}

func (a *AIPrompt) showResponse(response string) {
	a.responseArea.SetText(fmt.Sprintf("[green]Response:[-]\n%s", response))
}

func (a *AIPrompt) Render() {
	// This method can remain empty as before
}
