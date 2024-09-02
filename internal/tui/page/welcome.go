package page

import (
	"fmt"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
)

const (
	WelcomeView = "Welcome"
)

type Welcome struct {
	*core.BaseElement
	*tview.Flex

	// Form
	form *tview.Form

	// Callbacks
	onSubmit func()
}

func NewWelcome() *Welcome {
	w := &Welcome{
		BaseElement: core.NewBaseElement(),
		Flex:        tview.NewFlex(),
		form:        tview.NewForm(),
	}

	w.SetIdentifier(WelcomeView)

	return w
}

func (w *Welcome) Init(app *core.App) error {
	w.App = app

	w.setStyle()

	return nil
}

func (w *Welcome) setStyle() {
	style := w.App.GetStyles().Welcome

	w.form.SetBorder(true)
	w.form.SetFieldTextColor(style.FormInputColor.Color())
	w.form.SetFieldBackgroundColor(style.FormInputBackgroundColor.Color())
	w.SetBorder(false)
}

func (w *Welcome) Render() {
	w.Flex.Clear()

	// easy way to center the form
	w.AddItem(tview.NewBox(), 0, 1, false)

	w.renderForm()
	w.Flex.AddItem(w.form, 0, 3, true)

	w.AddItem(tview.NewBox(), 0, 1, false)
}

func (w *Welcome) SetOnSubmitFunc(onSubmit func()) {
	w.onSubmit = onSubmit
}

func (w *Welcome) renderForm() {
	w.form.Clear(true)

	w.form.SetBorder(true)
	w.form.SetTitle(" Welcome to Vi Mongo ")
	w.form.SetTitleAlign(tview.AlignCenter)

	configFile, err := config.GetConfigPath()
	if err != nil {
		modal.ShowError(w.App.Pages, "Error while getting config path", err)
		return
	}

	cfg := w.App.GetConfig()

	welcomeText := "All configuration can be set in " + configFile + " file. You can also set it here."
	w.form.AddTextView("Welcome info", welcomeText, 0, 2, true, false)
	w.form.AddTextView(" ", "-------------------------------------------", 0, 1, true, false)
	w.form.AddTextView("Editor", "Set command (vim, nano etc) or env variable ($ENV) to open editor", 0, 2, true, false)
	editorCmd, err := cfg.GetEditorCmd()
	if err != nil {
		editorCmd = ""
	}
	w.form.AddInputField("Set editor", editorCmd, 30, nil, nil)
	w.form.AddTextView("Logs", "Requires restart if changed", 0, 1, true, false)
	w.form.AddInputField("Log File", cfg.Log.Path, 30, nil, nil)
	w.form.AddInputField("Log Level", cfg.Log.Level, 30, nil, nil)
	w.form.AddTextView("Show on start", "Set pages to show on every start", 60, 1, true, false)
	w.form.AddCheckbox("Connection page", cfg.ShowConnectionPage, nil)
	w.form.AddCheckbox("Welcome page", cfg.ShowWelcomePage, nil)
	w.form.AddTextView("Show help", fmt.Sprintf("Press %s to show help", w.App.GetKeys().Global.ToggleFullScreenHelp.String()), 60, 1, true, false)

	w.form.AddButton(" Save and Connect ", func() {
		err := w.saveConfig()
		if err != nil {
			modal.ShowError(w.App.Pages, "Error while saving config", err)
			return
		}
		if w.onSubmit != nil {
			w.onSubmit()
		}
	})

	w.form.AddButton(" Exit ", func() {
		w.App.Stop()
	})
}

func (w *Welcome) saveConfig() error {
	editorCmd := w.form.GetFormItemByLabel("Set editor").(*tview.InputField).GetText()
	logFile := w.form.GetFormItemByLabel("Log File").(*tview.InputField).GetText()
	logLevel := w.form.GetFormItemByLabel("Log Level").(*tview.InputField).GetText()
	connPage := w.form.GetFormItemByLabel("Connection page").(*tview.Checkbox).IsChecked()
	welcomePage := w.form.GetFormItemByLabel("Welcome page").(*tview.Checkbox).IsChecked()

	c := w.App.GetConfig()

	splitedEditorCmd := strings.Split(editorCmd, "$")
	if len(splitedEditorCmd) > 1 {
		c.Editor.Command = ""
		c.Editor.Env = splitedEditorCmd[1]
	} else {
		c.Editor.Env = ""
		c.Editor.Command = editorCmd
	}
	c.Log.Path = logFile
	c.Log.Level = logLevel
	c.ShowConnectionPage = connPage
	c.ShowWelcomePage = welcomePage

	err := w.App.GetConfig().UpdateConfig()
	if err != nil {
		return err
	}

	return nil
}
