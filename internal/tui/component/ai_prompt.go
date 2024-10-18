package component

import (
	"fmt"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/ai"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptTabID = "AIPromptTab"
)

type AIPrompt struct {
	*core.BaseElement
	*core.Flex

	modelDropdown *tview.DropDown
	promptInput   *tview.TextArea
	submitButton  *tview.Button
}

func NewAIPrompt() *AIPrompt {
	a := &AIPrompt{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
	}

	a.SetIdentifier(AIPromptTabID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPrompt) init() error {
	a.setupComponents()
	a.setLayout()
	a.setStyle()

	return nil
}

func (a *AIPrompt) setupComponents() {
	a.modelDropdown = tview.NewDropDown().
		SetLabel("Model: ").
		SetOptions([]string{"OpenAI", "Anthropic"}, nil)

	a.promptInput = tview.NewTextArea().
		SetLabel("Prompt: ").
		SetPlaceholder("Enter your prompt here...")

	a.submitButton = tview.NewButton("Submit").
		SetSelectedFunc(a.onSubmit)
}

func (a *AIPrompt) setLayout() {
	// Create a vertical Flex layout
	verticalFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add the modelDropdown and promptInput to a horizontal Flex layout
	inputFlex := tview.NewFlex().
		AddItem(a.modelDropdown, 0, 1, false).
		AddItem(a.promptInput, 0, 2, false)

	// Add the inputFlex and submitButton to the vertical Flex layout
	verticalFlex.
		AddItem(inputFlex, 0, 1, false).
		AddItem(a.submitButton, 1, 0, false)

	// Set the verticalFlex as the main layout
	a.AddItem(verticalFlex, 0, 1, false)
}

func (a *AIPrompt) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)

	a.modelDropdown.SetLabelColor(styles.AIPrompt.LabelColor.Color())
	a.modelDropdown.SetFieldBackgroundColor(styles.AIPrompt.DropdownBackgroundColor.Color())
	a.modelDropdown.SetFieldTextColor(styles.AIPrompt.DropdownTextColor.Color())

	a.promptInput.SetBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())

	a.submitButton.SetBackgroundColor(styles.AIPrompt.ButtonBackgroundColor.Color())
	a.submitButton.SetLabelColor(styles.AIPrompt.ButtonTextColor.Color())
}

func (a *AIPrompt) onSubmit() {
	var driver ai.AIDriver

	_, options := a.modelDropdown.GetCurrentOption()
	switch options {
	case "OpenAI":
		driver = ai.NewOpenAIDriver("your-openai-api-key") // Replace with actual API key
	case "Anthropic":
		driver = ai.NewAnthropicDriver("your-anthropic-api-key") // Replace with actual API key
	default:
		return
	}

	systemMessage := `This prompt is for a query to MongoDB using the Query Bar. Example: { name: { $regex: "^catelyn", "$options": "i" } }`
	driver.SetSystemMessage(systemMessage)

	prompt := a.promptInput.GetText()
	response, err := driver.GetResponse(prompt)
	if err != nil {
		a.App.Error(fmt.Sprintf("Error getting response: %v", err))
		return
	}

	// TODO: Display the response in the UI
	fmt.Println("Response:", response)
}

func (a *AIPrompt) Render() {
	// This method is called by TabBar to render the component
	// For now, we don't need to do anything here as the component
	// is already set up in the init method
}
package component

import (
	"fmt"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/ai"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptTabID = "AIPromptTab"
)

type AIPrompt struct {
	*core.BaseElement
	*tview.Form

	modelDropdown *tview.DropDown
	promptInput   *tview.InputField
}

func NewAIPrompt() *AIPrompt {
	a := &AIPrompt{
		BaseElement: core.NewBaseElement(),
		Form:        tview.NewForm(),
	}

	a.SetIdentifier(AIPromptTabID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPrompt) init() error {
	a.setupComponents()
	a.setLayout()
	a.setStyle()

	return nil
}

func (a *AIPrompt) setupComponents() {
	a.modelDropdown = tview.NewDropDown().
		SetLabel("Model: ").
		SetOptions([]string{"OpenAI", "Anthropic"}, nil)

	a.promptInput = tview.NewInputField().
		SetLabel("Prompt: ").
		SetPlaceholder("Enter your prompt here...")

	a.AddFormItem(a.modelDropdown)
	a.AddFormItem(a.promptInput)

	a.AddButton("Submit", a.onSubmit)
	a.AddButton("Cancel", func() {
		// Handle cancel action
	})
}

func (a *AIPrompt) setLayout() {
	a.SetBorder(true).SetTitle("AI Prompt").SetTitleAlign(tview.AlignCenter)
}

func (a *AIPrompt) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)

	a.modelDropdown.SetLabelColor(styles.AIPrompt.LabelColor.Color())
	a.modelDropdown.SetFieldBackgroundColor(styles.AIPrompt.DropdownBackgroundColor.Color())
	a.modelDropdown.SetFieldTextColor(styles.AIPrompt.DropdownTextColor.Color())

	a.promptInput.SetFieldBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())

	a.SetButtonBackgroundColor(styles.AIPrompt.ButtonBackgroundColor.Color())
	a.SetButtonTextColor(styles.AIPrompt.ButtonTextColor.Color())
}

func (a *AIPrompt) onSubmit() {
	var driver ai.AIDriver

	_, options := a.modelDropdown.GetCurrentOption()
	switch options {
	case "OpenAI":
		driver = ai.NewOpenAIDriver("your-openai-api-key") // Replace with actual API key
	case "Anthropic":
		driver = ai.NewAnthropicDriver("your-anthropic-api-key") // Replace with actual API key
	default:
		return
	}

	systemMessage := `This prompt is for a query to MongoDB using the Query Bar. Example: { name: { $regex: "^catelyn", "$options": "i" } }`
	driver.SetSystemMessage(systemMessage)

	prompt := a.promptInput.GetText()
	response, err := driver.GetResponse(prompt)
	if err != nil {
		a.App.Error(fmt.Sprintf("Error getting response: %v", err))
		return
	}

	// TODO: Display the response in the UI
	fmt.Println("Response:", response)
}

func (a *AIPrompt) Render() {
	// This method is called by TabBar to render the component
	// For now, we don't need to do anything here as the component
	// is already set up in the init method
}
