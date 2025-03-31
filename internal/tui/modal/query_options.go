package modal

import (
	"context"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	QueryOptionsModalId = "QueryOptionsModal"
)

type QueryOptionsModal struct {
	*core.BaseElement
	*core.FormModal

	applyCallback func()
}

func NewQueryOptionsModal() *QueryOptionsModal {
	qo := &QueryOptionsModal{
		BaseElement: core.NewBaseElement(),
		FormModal:   core.NewFormModal(),
	}

	qo.SetIdentifier(StyleChangeModal)
	qo.SetAfterInitFunc(qo.init)
	return qo
}

func (qo *QueryOptionsModal) init() error {
	qo.setLayout()
	qo.setStyle()
	qo.setKeybindings()

	return nil
}

func (qo *QueryOptionsModal) setLayout() {
	qo.SetTitle(" Query Options ")
	qo.SetBorder(true)
	qo.SetTitleAlign(tview.AlignCenter)
	qo.Form.SetBorderPadding(2, 2, 2, 2)
}

func (qo *QueryOptionsModal) setStyle() {
	styles := qo.App.GetStyles()
	qo.SetStyle(styles)

	qo.Form.SetFieldTextColor(styles.Connection.FormInputColor.Color())
	qo.Form.SetFieldBackgroundColor(styles.Connection.FormInputBackgroundColor.Color())
	qo.Form.SetLabelColor(styles.Connection.FormLabelColor.Color())
}

func (qo *QueryOptionsModal) setKeybindings() {
	k := qo.App.GetKeys()
	qo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			qo.Hide()
			return nil
		}
		switch {
		case k.Contains(k.Content.ToggleQueryOptions, event.Name()):
			qo.Hide()
			return nil
		}

		return event
	})
}

func (qo *QueryOptionsModal) SetApplyCallback(callback func()) {
	qo.applyCallback = callback
}

func (qo *QueryOptionsModal) Render(ctx context.Context, state *mongo.CollectionState) error {
	qo.Form.Clear(true)

	qo.Form.AddInputField("Projection", state.Projection, 40, nil, nil)

	limitStr := strconv.FormatInt(state.Limit, 10)
	qo.Form.AddInputField("Limit", limitStr, 20,
		func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil || textToCheck == ""
		}, nil)

	skipStr := strconv.FormatInt(state.Skip, 10)
	qo.Form.AddInputField("Skip", skipStr, 20,
		func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil || textToCheck == ""
		}, nil)

	qo.Form.AddButton("Apply", func() {
		limitText := qo.Form.GetFormItemByLabel("Limit").(*tview.InputField).GetText()
		skipText := qo.Form.GetFormItemByLabel("Skip").(*tview.InputField).GetText()
		projText := qo.Form.GetFormItemByLabel("Projection").(*tview.InputField).GetText()

		if strings.Trim(limitText, " ") != "" {
			val, err := strconv.ParseInt(limitText, 10, 64)
			if err != nil {
				ShowError(qo.App.Pages, "Invalid limit value", err)
				return
			}
			state.Limit = val
		}

		if strings.Trim(skipText, " ") != "" {
			val, err := strconv.ParseInt(skipText, 10, 64)
			if err != nil {
				ShowError(qo.App.Pages, "Invalid skip value", err)
				return
			}
			state.Skip = val
		}

		state.Projection = projText

		if qo.applyCallback != nil {
			qo.applyCallback()
		}
	})

	qo.Form.AddButton("Cancel", func() {
		qo.Hide()
	})

	qo.Show()

	return nil
}

func (qo *QueryOptionsModal) Show() {
	qo.App.Pages.AddPage(QueryOptionsModalId, qo, true, true)
}

func (qo *QueryOptionsModal) Hide() {
	qo.App.Pages.RemovePage(QueryOptionsModalId)
}
