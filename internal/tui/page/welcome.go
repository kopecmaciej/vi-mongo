package page

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/widget"
)

const (
	WelcomePageId = "Welcome"
)

type Welcome struct {
	*core.BaseElement
	*core.Flex

	form    *core.Form
	hintBar *widget.HintBar

	style *config.WelcomeStyle

	onSubmit func()
}

func NewWelcome() *Welcome {
	w := &Welcome{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		form:        core.NewForm(),
		hintBar:     widget.NewHintBar(),
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

	w.form.AddButton("Save", func() {
		err := w.saveConfig()
		if err != nil {
			modal.ShowError(w.App.Pages, "Error while saving config", err)
			return
		}
		if w.onSubmit != nil {
			w.onSubmit()
		}
	})

	w.form.AddButton("Exit", func() {
		w.App.Stop()
	})

	w.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := w.App.GetKeys()
		if k.Contains(k.Connection.ConnectionForm.SaveConnection, event.Name()) {
			err := w.saveConfig()
			if err != nil {
				modal.ShowError(w.App.Pages, "Error while saving config", err)
				return nil
			}
			if w.onSubmit != nil {
				w.onSubmit()
			}
			return nil
		}
		return event
	})
}

func (w *Welcome) setStyle() {
	w.style = &w.App.GetStyles().Welcome
	w.Flex.SetStyle(w.App.GetStyles())
	w.form.SetStyle(w.App.GetStyles())
	w.hintBar.SetStyle(w.App.GetStyles())

	w.form.SetFieldTextColor(w.style.FormInputColor.Color())
	w.form.SetFieldBackgroundColor(w.style.FormInputBackgroundColor.Color())
	w.form.SetLabelColor(w.style.FormLabelColor.Color())
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
	w.Clear()
	w.SetDirection(tview.FlexRow)

	centerFlex := tview.NewFlex()
	centerFlex.AddItem(tview.NewBox(), 0, 1, false)
	w.renderForm()
	centerFlex.AddItem(w.form, 0, 3, true)
	centerFlex.AddItem(tview.NewBox(), 0, 1, false)

	w.AddItem(centerFlex, 0, 1, true)
	w.renderHints()
	w.AddItem(w.hintBar, 1, 0, false)
	w.AddItem(tview.NewBox(), 1, 0, false)

	if page, _ := w.App.Pages.GetFrontPage(); page == WelcomePageId {
		w.App.SetFocus(w)
	}
}

func (w *Welcome) renderHints() {
	k := w.App.GetKeys()
	w.hintBar.SetHints([]widget.Hint{
		{Key: "Tab", Desc: "form down"},
		{Key: "Backtab", Desc: "form up"},
		{Key: k.Connection.ConnectionForm.SaveConnection.String(), Desc: "save"},
	})
}

func (w *Welcome) SetOnSubmitFunc(onSubmit func()) {
	w.onSubmit = onSubmit
}

func (w *Welcome) renderForm() {
	w.form.Clear(false)

	cfg := w.App.GetConfig()
	gKeys := w.App.GetKeys().Global

	configFile, err := cfg.GetCurrentConfigPath()
	if err != nil {
		modal.ShowError(w.App.Pages, "Error while getting config path", err)
		return
	}

	welcomeText := "All configuration can be set in " + configFile + " file. You can also set it here."
	w.form.AddTextView("Welcome info", welcomeText, 0, 2, true, false)
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
	w.form.AddTextView("Help", fmt.Sprintf("'%s' for full page, '%s' to expand keys in header", gKeys.ToggleFullScreenHelp.String(), gKeys.ToggleHeader.String()), 60, 1, true, false)
	w.form.AddTextView("Motions", "Use basic vim motions or normal arrow keys to move around", 60, 2, true, false)
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
