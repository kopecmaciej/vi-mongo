package component

import (
    "github.com/gdamore/tcell/v2"
    "github.com/kopecmaciej/tview"
)

const (
    SortBarComponent = "SortBar"
)

type SortBar struct {
    *Component
    *tview.InputField

    enabled bool
}

func NewSortBar() *SortBar {
    s := &SortBar{
        Component:  NewComponent(SortBarComponent),
        InputField: tview.NewInputField().SetLabel(" Sort: "),
        enabled:    false,
    }

    s.SetAfterInitFunc(s.init)

    return s
}

func (s *SortBar) init() error {
    s.setStyle()
    s.setKeybindings()

    return nil
}

func (s *SortBar) setStyle() {
    s.SetBorder(true)
    s.SetFieldTextColor(tcell.ColorWhite)
}

func (s *SortBar) setKeybindings() {
    s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyEsc:
            s.Toggle()
        }
        return event
    })
}

func (s *SortBar) Toggle() {
    s.enabled = !s.enabled
}

func (s *SortBar) IsEnabled() bool {
    return s.enabled
}
