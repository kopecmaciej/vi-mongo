package component

import (
	"fmt"
	"os"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/ai"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptID = "AIPrompt"
)

type AIPrompt struct {
	*core.BaseElement
	*core.Form

	responseArea *core.TextView
	docKeys      []string
}

func NewAIPrompt() *AIPrompt {
	a := &AIPrompt{
		BaseElement:  core.NewBaseElement(),
		Form:         core.NewForm(),
		responseArea: core.NewTextView(),
	}

	a.SetIdentifier(AIPromptID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPrompt) init() error {
	a.setLayout()
	a.setStyle()

	return nil
}

func (a *AIPrompt) setLayout() {
	a.SetBorder(true)
	a.SetTitle("AI Prompt")
	a.SetTitleAlign(tview.AlignCenter)
	a.SetBorderPadding(0, 0, 1, 1)
}

func (a *AIPrompt) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)

	a.SetButtonBackgroundColor(styles.AIPrompt.ButtonBackgroundColor.Color())
	a.SetButtonTextColor(styles.AIPrompt.ButtonTextColor.Color())
	a.SetLabelColor(styles.AIPrompt.LabelColor.Color())
	a.SetFieldBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())
	a.SetFieldTextColor(styles.AIPrompt.InputTextColor.Color())

	a.responseArea.SetBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())
	a.responseArea.SetTextColor(styles.AIPrompt.InputTextColor.Color())
}

func (a *AIPrompt) handleEvents() {
	go a.HandleEvents(AIPromptID, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			a.setStyle()
			a.Render()
		case manager.UpdateAutocompleteKeys:
			a.docKeys = event.Message.Data.([]string)
		}
	})
}

func (a *AIPrompt) Render() {
	a.Form.Clear(true)

	openaiModels := ai.GetOpenAiModels()
	anthropicModels := ai.GetAnthropicModels()

	a.AddDropDown("Model:", append(openaiModels, anthropicModels...), 0, nil).
		AddTextArea("Prompt:", "", 0, 3, 0, nil).
		AddButton("Submit", a.onSubmit)

	a.AddFormItem(a.responseArea)
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

	systemMessage := fmt.Sprintf(`You are an assistant helping to create MongoDB queries. 
	Respond with valid MongoDB query syntax that can be directly used in a query.
	
	Rules:
	1. Always use proper MongoDB operators (e.g., $regex, $exists, $gt, $lt, $in).
	2. Keys should always be quoted, but values should not be quoted unless they are strings.
	3. Numbers and booleans should not be in quotes.
	4. Use proper formatting for regex patterns (e.g., "^pattern").
	
	Available document keys: %s
	
	If the user makes a mistake with a key name, correct it based on the available keys.
	
	Example query: { name: { $regex: "^john", $options: "i" }, age: { $gt: 30 }, isActive: true }
	
	Respond only with the query, without any additional explanation.`, strings.Join(a.docKeys, ", "))

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
