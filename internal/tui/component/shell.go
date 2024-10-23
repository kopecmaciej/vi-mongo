package component

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/rs/zerolog/log"
)

const (
	ShellId = "Shell"
)

type Shell struct {
	*core.BaseElement
	*core.Flex

	output *tview.TextView
	input  *tview.InputField
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

func (s *Shell) Init() error {
	return s.init()
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
			go s.handleInput()
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
		s.output.Write([]byte("Error: " + err.Error() + "\n"))
		return
	}

	s.output.Write(output)
	s.output.Write([]byte("\n"))
}

func (s *Shell) executeCommand(command string) ([]byte, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "mongosh", "--eval", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("Error executing mongosh command")
		return nil, err
	}

	if stderr.Len() > 0 {
		return nil, fmt.Errorf(stderr.String())
	}

	return stdout.Bytes(), nil
}
