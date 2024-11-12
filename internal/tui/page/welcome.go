package page

import (
	"fmt"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
)

const (
	WelcomePageId = "Welcome"
)

type Welcome struct {
	*core.BaseElement
	*core.Flex

	// Form
	form *core.Form

	// Callbacks
	onSubmit func()
}

func NewWelcome() *Welcome {
	w := &Welcome{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		form:        core.NewForm(),
	}

	w.SetIdentifier(WelcomePageId)

	return w
}

func (w *Welcome) Init(app *core.App) error {
	w.App = app

	w.setLayout()
	w.setStyle()

	w.handleEvents()

	return nil
}

func (w *Welcome) setLayout() {
	w.form.SetBorder(true)
	w.form.SetTitle(" Welcome to Vi Mongo ")
	w.form.SetTitleAlign(tview.AlignCenter)
	w.form.SetButtonsAlign(tview.AlignCenter)

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

func (w *Welcome) setStyle() {
	w.Flex.SetStyle(w.App.GetStyles())
	w.form.SetStyle(w.App.GetStyles())
	style := w.App.GetStyles().Welcome

	w.form.SetFieldTextColor(style.FormInputColor.Color())
	w.form.SetFieldBackgroundColor(style.FormInputBackgroundColor.Color())
	w.form.SetLabelColor(style.FormLabelColor.Color())
}

func (w *Welcome) handleEvents() {
	go w.HandleEvents(WelcomePageId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			w.setStyle()
			go w.App.QueueUpdateDraw(func() {
				w.Render()
			})
		}
	})
}
func (w *Welcome) Render() {
	w.Flex.Clear()

	// easy way to center the form
	w.AddItem(tview.NewBox(), 0, 1, false)

	w.renderForm()
	w.Flex.AddItem(w.form, 0, 3, true)

	w.AddItem(tview.NewBox(), 0, 1, false)

	if page, _ := w.App.Pages.GetFrontPage(); page == WelcomePageId {
		w.App.SetFocus(w)
	}
}

func (w *Welcome) SetOnSubmitFunc(onSubmit func()) {
	w.onSubmit = onSubmit
}

func (w *Welcome) renderForm() {
	w.form.Clear(false)

	configFile, err := config.GetConfigPath()
	if err != nil {
		modal.ShowError(w.App.Pages, "Error while getting config path", err)
		return
	}

	cfg := w.App.GetConfig()

	welcomeText := "All configuration can be set in " + configFile + " file. You can also set it here."
	w.form.AddTextView("Welcome info", welcomeText, 0, 2, true, false)
	w.form.AddTextView(" ", "----------------------------------------------------------", 0, 1, true, false)
	w.form.AddTextView("Editor", "Set command (vim, nano etc) or env ($ENV)", 0, 1, true, false)
	editorCmd, err := cfg.GetEditorCmd()
	if err != nil {
		editorCmd = ""
	}
	w.form.AddInputField("Set editor", editorCmd, 30, nil, nil)
	w.form.AddTextView("Logs", "Requires restart if changed", 0, 1, true, false)
	w.form.AddInputField("Log File", cfg.Log.Path, 30, nil, nil)
	logLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	w.form.AddDropDown("Log Level", logLevels, getLogLevelIndex(cfg.Log.Level, logLevels), nil)
	w.form.AddCheckbox("Use symbols 🗁 🖿 🗎", cfg.Styles.BetterSymbols, nil)
	w.form.AddTextView("Show on start", "Set pages to show on every start", 60, 1, true, false)
	w.form.AddCheckbox("Connection page", cfg.ShowConnectionPage, nil)
	w.form.AddCheckbox("Welcome page", cfg.ShowWelcomePage, nil)
	w.form.AddTextView("Show help", fmt.Sprintf("Press %s to show key help", w.App.GetKeys().Global.ToggleFullScreenHelp.String()), 60, 1, true, false)
	w.form.AddTextView("Motions", "Use basic vim motions or normal arrow keys to move around", 60, 1, true, false)
}

func (w *Welcome) saveConfig() error {
	editorCmd := w.form.GetFormItemByLabel("Set editor").(*tview.InputField).GetText()
	logFile := w.form.GetFormItemByLabel("Log File").(*tview.InputField).GetText()
	// Get the selected log level from the dropdown
	_, logLevel := w.form.GetFormItemByLabel("Log Level").(*tview.DropDown).GetCurrentOption()

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
	c.ShowConnectionPage = w.form.GetFormItemByLabel("Connection page").(*tview.Checkbox).IsChecked()
	c.ShowWelcomePage = w.form.GetFormItemByLabel("Welcome page").(*tview.Checkbox).IsChecked()

	betterSymbols := w.form.GetFormItemByLabel("Use symbols 🗁 🖿 🗎").(*tview.Checkbox).IsChecked()
	if betterSymbols != c.Styles.BetterSymbols {
		c.Styles.BetterSymbols = betterSymbols
		w.App.SetStyle(c.Styles.CurrentStyle)
	}

	err := w.App.GetConfig().UpdateConfig()
	if err != nil {
		return err
	}

	return nil
}

// Add this helper function at the end of the file
func getLogLevelIndex(currentLevel string, levels []string) int {
	for i, level := range levels {
		if level == currentLevel {
			return i
		}
	}
	return 0
}
