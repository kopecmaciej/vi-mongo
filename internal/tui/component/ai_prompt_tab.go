package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptTabID = "AIPromptTab"
)

type AIPromptTab struct {
	*core.BaseElement
	*tview.Flex

	modelDropdown *tview.DropDown
	promptInput   *tview.TextArea
	submitButton  *tview.Button
}

func NewAIPromptTab() *AIPromptTab {
	a := &AIPromptTab{
		BaseElement: core.NewBaseElement(),
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
	}

	a.SetIdentifier(AIPromptTabID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPromptTab) init() error {
	a.setupComponents()
	a.setLayout()
	a.setStyle()

	return nil
}

func (a *AIPromptTab) setupComponents() {
	a.modelDropdown = tview.NewDropDown().
		SetLabel("Model: ").
		SetOptions([]string{"OpenAI", "Anthropic"}, nil)

	a.promptInput = tview.NewTextArea().
		SetLabel("Prompt: ").
		SetPlaceholder("Enter your prompt here...")

	a.submitButton = tview.NewButton("Submit").
		SetSelectedFunc(a.onSubmit)
}

func (a *AIPromptTab) setLayout() {
	a.AddItem(a.modelDropdown, 1, 0, false)
	a.AddItem(a.promptInput, 0, 1, false)
	a.AddItem(a.submitButton, 1, 0, false)
}

func (a *AIPromptTab) setStyle() {
	styles := a.App.GetStyles()
	a.SetBackgroundColor(styles.Content.BackgroundColor.Color())

	a.modelDropdown.SetBackgroundColor(styles.Content.BackgroundColor.Color())
	a.modelDropdown.SetLabelColor(styles.Content.TextColor.Color())
	a.modelDropdown.SetFieldBackgroundColor(styles.Content.BackgroundColor.Color())
	a.modelDropdown.SetFieldTextColor(styles.Content.TextColor.Color())

	a.promptInput.SetBackgroundColor(styles.Content.BackgroundColor.Color())
	a.promptInput.SetLabelColor(styles.Content.TextColor.Color())
	a.promptInput.SetTextColor(styles.Content.TextColor.Color())

	a.submitButton.SetBackgroundColor(styles.Content.BackgroundColor.Color())
	a.submitButton.SetLabelColor(styles.Content.TextColor.Color())
}

func (a *AIPromptTab) onSubmit() {
	// TODO: Implement submission logic
}

func (a *AIPromptTab) Render() {
	// This method is called by TabBar to render the component
	// For now, we don't need to do anything here as the component
	// is already set up in the init method
}
