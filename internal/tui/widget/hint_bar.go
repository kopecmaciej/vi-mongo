package widget

import (
	"fmt"
	"strings"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
)

// Hint is a key-description pair displayed in a HintBar.
type Hint struct {
	Key  string
	Desc string
}

// HintBar is a single-line, centered, read-only text view that renders a row
// of key-description hints with accent/dim coloring.
type HintBar struct {
	*tview.TextView
	styles *config.Styles
}

func NewHintBar() *HintBar {
	h := &HintBar{TextView: tview.NewTextView()}
	h.SetTextAlign(tview.AlignCenter)
	h.SetDynamicColors(true)
	return h
}

func (h *HintBar) SetStyle(styles *config.Styles) {
	h.styles = styles
	h.SetBackgroundColor(styles.Global.BackgroundColor.Color())
	h.SetTextColor(styles.Global.TextColor.Color())
}

// SetHints renders the provided hints into the bar. Must be called after SetStyle.
func (h *HintBar) SetHints(hints []Hint) {
	if h.styles == nil {
		return
	}
	dimHex := fmt.Sprintf("#%06x", h.styles.Global.TextColor.Color().Hex())
	accentHex := fmt.Sprintf("#%06x", h.styles.Global.FocusColor.Color().Hex())

	parts := make([]string, len(hints))
	for i, hint := range hints {
		parts[i] = fmt.Sprintf("[%s]%s[-] [%s]%s[-]", accentHex, hint.Key, dimHex, hint.Desc)
	}
	h.SetText(strings.Join(parts, "  "))
}
