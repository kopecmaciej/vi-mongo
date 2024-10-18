package component

import (
	"github.com/kopecmaciej/tview"
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
	// TODO: Implement submission logic
}

func (a *AIPrompt) Render() {
	// This method is called by TabBar to render the component
	// For now, we don't need to do anything here as the component
	// is already set up in the init method
}
