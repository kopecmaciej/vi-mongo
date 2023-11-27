package component

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event string

type InputBar struct {
	*tview.InputField

	EventChan chan interface{}
	mutex     sync.Mutex
	label     string
	enabled   bool
}

func NewInputBar(label string) *InputBar {
	f := &InputBar{
		InputField: tview.NewInputField(),
		mutex:      sync.Mutex{},
		label:      label,
		EventChan:  make(chan interface{}),
		enabled:    false,
	}

	return f
}

func (i *InputBar) Init(ctx context.Context) error {
	i.setStyle()
	i.SetLabel(" " + i.label + ": ")

	i.SetEventFunc()

	i.AutocompleteHistory()

	return nil
}

func (i *InputBar) SetEventFunc() {
	i.SetDoneFunc(func(key tcell.Key) {
		i.EventChan <- key
	})
}

func (i *InputBar) setStyle() {
	i.SetBackgroundColor(tcell.ColorDefault)
	i.SetFieldBackgroundColor(tcell.ColorDefault)
	i.SetFieldTextColor(tcell.ColorDefault)
	i.SetPlaceholderTextColor(tcell.ColorDefault)
}

const (
	maxHistory = 20
)

func (i *InputBar) SaveToHistory(text string) error {
	file, err := os.OpenFile("history.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
    return err
	}
	defer file.Close()

	if _, err := file.WriteString(text + "\n"); err != nil {
    return err
	}

  return nil
}

func (i *InputBar) LoadHistory() ([]string, error) {
	file, err := os.Open("history.txt")
	if err != nil {
    return nil, err
	}
	defer file.Close()

	var history []string
	var line string
	for {
		_, err := fmt.Fscanf(file, "%s\n", &line)
		if err != nil {
			break
		}
		history = append(history, line)
	}

  return history, nil
}

func (i *InputBar) AutocompleteHistory() {
	i.SetAutocompleteFunc(func(currentText string) (entries []string) {
		return []string{"test", "test2"}
	})
}

func (i *InputBar) IsEnabled() bool {
	return i.enabled
}

func (i *InputBar) Enable() {
	i.enabled = true
}

func (i *InputBar) Disable() {
	i.enabled = false
}

func (i *InputBar) Toggle() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.IsEnabled() {
		i.Disable()
	} else {
		i.Enable()
	}
}
