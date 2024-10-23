package component

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	ShellId = "Shell"
)

type Shell struct {
	*core.BaseElement
	*core.Flex

	output *tview.TextView
	input  *tview.InputField
	db     string
}

func NewShell() *Shell {
	s := &Shell{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		output:      tview.NewTextView(),
		input:       tview.NewInputField(),
	}

	s.SetIdentifier(ShellId)
	s.SetAfterInitFunc(s.init)

	return s
}

func (s *Shell) init() error {
	s.setLayout()
	s.setStyle()
	s.setKeybindings()

	s.handleEvents()

	return nil
}

func (s *Shell) setLayout() {
	s.Flex.SetDirection(tview.FlexRow)
	s.output.SetDynamicColors(true)
	s.output.SetScrollable(true)
	s.output.SetBorder(true)
	s.output.SetTitle(" Shell Output ")
	s.output.SetBorderPadding(0, 0, 1, 1)

	s.input.SetLabel(" > ")
	s.input.SetBorder(true)
	s.input.SetTitle(" Command Input ")
	s.input.SetBorderPadding(0, 0, 1, 1)
}

func (s *Shell) setStyle() {
	s.SetStyle(s.App.GetStyles())
	s.output.SetTextColor(s.App.GetStyles().Global.TextColor.Color())
	s.input.SetFieldTextColor(s.App.GetStyles().Global.TextColor.Color())
}

func (s *Shell) setKeybindings() {
	s.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			go s.App.QueueUpdateDraw(func() {
				s.handleInput()
			})
			return nil
		}
		return event
	})
}

func (s *Shell) handleEvents() {
	go s.HandleEvents(ShellId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			s.setStyle()
			s.Render()
		}
	})
}

func (s *Shell) SetDb(db string) {
	s.db = db
}

func (s *Shell) Render() {
	s.Flex.Clear()
	s.Flex.AddItem(s.output, 0, 1, false)
	s.Flex.AddItem(s.input, 3, 0, true)
}

func (s *Shell) handleInput() {
	command := s.input.GetText()
	if command == "" {
		return
	}

	s.input.SetText("")
	s.output.Write([]byte(" > " + command + "\n"))

	output, err := s.executeCommand(command)
	if err != nil {
		s.output.SetText("Error: " + err.Error())
		return
	}

	s.output.SetText(string(output))
}

func (s *Shell) executeCommand(command string) ([]byte, error) {
	return s.Dao.ExecuteCommand(context.Background(), s.db, command)
}
