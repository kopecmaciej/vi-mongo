package modal

import (
	"context"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/util"
)

const (
	InlineEditModalId = "InlineEditModal"
)

type InlineEditModal struct {
	*core.BaseElement
	*core.FormModal

	fieldName      string
	originalValue  string
	applyCallback  func(fieldName, newValue string) error
	cancelCallback func()
}

func NewInlineEditModal() *InlineEditModal {
	iem := &InlineEditModal{
		BaseElement: core.NewBaseElement(),
		FormModal:   core.NewFormModal(),
	}

	iem.SetIdentifier(InlineEditModalId)
	iem.SetAfterInitFunc(iem.init)
	return iem
}

func (iem *InlineEditModal) init() error {
	iem.setLayout()
	iem.setStyle()
	iem.setKeybindings()
	iem.handleEvents()

	return nil
}

func (iem *InlineEditModal) setLayout() {
	iem.SetTitle(" Inline Edit ")
	iem.SetBorder(true)
	iem.SetTitleAlign(tview.AlignCenter)
	iem.Form.SetBorderPadding(2, 2, 2, 2)
}

func (iem *InlineEditModal) setStyle() {
	styles := iem.App.GetStyles()
	iem.SetStyle(styles)

	iem.Form.SetFieldTextColor(styles.Connection.FormInputColor.Color())
	iem.Form.SetFieldBackgroundColor(styles.Connection.FormInputBackgroundColor.Color())
	iem.Form.SetLabelColor(styles.Connection.FormLabelColor.Color())
}

func (iem *InlineEditModal) setKeybindings() {
	iem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			if iem.cancelCallback != nil {
				iem.cancelCallback()
			}
			return nil
		case tcell.KeyEnter:
			formItemIndex, _ := iem.Form.GetFocusedItemIndex()
			if formItemIndex == 0 {
				iem.handleApply()
				return nil
			}
		}

		return event
	})
}

func (iem *InlineEditModal) handleEvents() {
	go iem.HandleEvents(iem.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			iem.setStyle()
		}
	})
}

func (iem *InlineEditModal) SetApplyCallback(callback func(fieldName, newValue string) error) {
	iem.applyCallback = callback
}

func (iem *InlineEditModal) SetCancelCallback(callback func()) {
	iem.cancelCallback = callback
}

func (iem *InlineEditModal) handleApply() {
	if iem.Form.GetFormItemCount() == 0 {
		return
	}

	var newValue string
	formItem := iem.Form.GetFormItem(0)

	switch field := formItem.(type) {
	case *tview.InputField:
		newValue = field.GetText()
	case *tview.TextArea:
		newValue = field.GetText()
	default:
		return
	}

	if iem.applyCallback != nil {
		err := iem.applyCallback(iem.fieldName, newValue)
		if err != nil {
			ShowError(iem.App.Pages, "Error applying edit", err)
			return
		}
	}
}

func (iem *InlineEditModal) Render(ctx context.Context, fieldName, currentValue string) error {
	iem.Form.Clear(true)
	iem.fieldName = fieldName
	iem.originalValue = currentValue

	// Clean the current value for display
	displayValue := util.CleanJsonWhitespaces(currentValue)

	// Use text area for multiline values or very long values
	if strings.Contains(displayValue, "\n") || len(displayValue) > 100 {
		textArea := tview.NewTextArea().
			SetText(displayValue, true).
			SetWrap(true).
			SetSize(8, 0)

		iem.Form.AddFormItem(textArea)
	} else {
		inputField := tview.NewInputField().
			SetText(displayValue).
			SetFieldWidth(0)

		iem.Form.AddFormItem(inputField)
	}

	iem.Form.AddButton("Apply", func() {
		iem.handleApply()
	})

	iem.Form.AddButton("Cancel", func() {
		if iem.cancelCallback != nil {
			iem.cancelCallback()
		}
	})

	iem.Show()
	return nil
}

func (iem *InlineEditModal) Show() {
	iem.App.Pages.AddPage(InlineEditModalId, iem, true, true)
}

func (iem *InlineEditModal) Hide() {
	iem.App.Pages.RemovePage(InlineEditModalId)
}
