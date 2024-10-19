package component

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	// a.setKeybindings()

	a.handleEvents()

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

	a.responseArea.SetBackgroundColor(styles.AIPrompt.InputBackgroundColor.Color())
	a.responseArea.SetTextColor(styles.AIPrompt.InputTextColor.Color())
}

func (a *AIPrompt) setKeybindings() {
	k := a.App.GetKeys()
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.AIPrompt.NextItem, event.Name()):
			curItem, curButton := a.Form.GetFocusedItemIndex()
			totalItems := a.Form.GetFormItemCount()
			totalButtons := a.Form.GetButtonCount()

			if curItem >= 0 {
				if curItem < totalItems-1 {
					a.Form.SetFocus(curItem + 1)
				} else if totalButtons > 0 {
					a.Form.SetFocus(totalItems)
				} else {
					a.Form.SetFocus(0)
				}
			} else if curButton >= 0 {
				if curButton < totalButtons-1 {
					a.Form.SetFocus(totalItems + curButton + 1)
				} else {
					a.Form.SetFocus(0)
				}
			}
			return nil
		case k.Contains(k.AIPrompt.PrevItem, event.Name()):
			curItem, curButton := a.Form.GetFocusedItemIndex()
			totalItems := a.Form.GetFormItemCount()
			totalButtons := a.Form.GetButtonCount()

			if curItem >= 0 {
				if curItem > 0 {
					a.Form.SetFocus(curItem - 1)
				} else if totalButtons > 0 {
					a.Form.SetFocus(totalItems + totalButtons - 1)
				} else {
					a.Form.SetFocus(totalItems - 1)
				}
			} else if curButton >= 0 {
				if curButton > 0 {
					a.Form.SetFocus(totalItems + curButton - 1)
				} else {
					a.Form.SetFocus(totalItems - 1)
				}
			}
			return nil
		}
		return event
	})
}

func (a *AIPrompt) IsAIPromptFocused() bool {
	if a.App.GetFocus() == a.Form {
		return true
	}
	if a.App.GetFocus().GetIdentifier() == a.GetIdentifier() {
		return true
	}
	return false
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

	switch {
	case slices.Contains(ai.GetOpenAiModels(), model):
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			a.showError("OpenAI API key not found in environment variables")
			return
		}
		driver = ai.NewOpenAIDriver(apiKey)
	case slices.Contains(ai.GetAnthropicModels(), model):
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			a.showError("Anthropic API key not found in environment variables")
			return
		}
		driver = ai.NewAnthropicDriver(apiKey)
	default:
		a.showError(fmt.Sprintf("Invalid AI model selected: %s", model))
		return
	}

	systemMessage := fmt.Sprintf(`You are an assistant helping to create MongoDB queries. 
	Respond with valid MongoDB query syntax that can be directly used in a query.
	
	Rules:
	1. Always use proper MongoDB operators (e.g., $regex, $exists, $gt, $lt, $in).
	2. Quote values that are not numbers or booleans.
	3. Use proper formatting for regex patterns (e.g., "^pattern").
	
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
	a.responseArea.SetText(fmt.Sprintf("Error: %s[-]", message))
}

func (a *AIPrompt) showResponse(response string) {
	a.responseArea.SetText(fmt.Sprintf("Response:[-]\n%s", response))
}
