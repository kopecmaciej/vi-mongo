package modal

import (
	"context"
	"strconv"

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

	state         *mongo.CollectionState
	applyCallback func(projection string, limit int64) error
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
	qo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			qo.App.Pages.RemovePage(QueryOptionsModalId)
			return nil
		}
		return event
	})
}

func (qo *QueryOptionsModal) SetState(state *mongo.CollectionState) {
	qo.state = state
}

func (qo *QueryOptionsModal) SetApplyCallback(callback func(projection string, limit int64) error) {
	qo.applyCallback = callback
}

func (qo *QueryOptionsModal) Render(ctx context.Context) error {
	qo.Form.Clear(true)

	qo.Form.AddInputField("Projection", "", 50, nil, nil)

	limitStr := strconv.FormatInt(qo.state.Limit, 10)
	qo.Form.AddInputField("Limit", limitStr, 20,
		func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil || textToCheck == ""
		}, nil)

	qo.Form.AddButton("Apply", func() {
		projText := qo.Form.GetFormItemByLabel("Projection").(*tview.InputField).GetText()
		limitText := qo.Form.GetFormItemByLabel("Limit").(*tview.InputField).GetText()

		var limitVal int64
		if limitText != "" {
			val, err := strconv.ParseInt(limitText, 10, 64)
			if err != nil {
				ShowError(qo.App.Pages, "Invalid limit value", err)
				return
			}
			limitVal = val
		} else {
			limitVal = qo.state.Limit
		}

		if qo.applyCallback != nil {
			err := qo.applyCallback(projText, limitVal)
			if err != nil {
				ShowError(qo.App.Pages, "Error applying query options", err)
				return
			}
		}

		qo.App.Pages.RemovePage(QueryOptionsModalId)
	})

	qo.Form.AddButton("Cancel", func() {
		qo.App.Pages.RemovePage(QueryOptionsModalId)
	})

	return nil
}

func (qo *QueryOptionsModal) Show() {
	qo.App.Pages.AddPage(QueryOptionsModalId, qo, true, true)
}
