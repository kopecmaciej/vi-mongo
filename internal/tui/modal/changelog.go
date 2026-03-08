package modal

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const ChangelogModalId = "Changelog"

// ChangelogEntry describes changes introduced in a specific version.
// Entries with Breaking=true present the user with a "Quit" option
// in addition to proceeding. MigrationFn is executed when the user proceeds.
type ChangelogEntry struct {
	Version     string
	Breaking    bool
	Title       string
	Changes     []string
	MigrationFn func() error
}

// Changelog is a one-time startup modal that informs the user about
// version changes. Breaking entries include a "Quit" button.
type Changelog struct {
	*core.BaseElement
	*core.Modal

	entries   []ChangelogEntry
	onProceed func()
	onQuit    func()
}

func NewChangelog(entries []ChangelogEntry) *Changelog {
	c := &Changelog{
		BaseElement: core.NewBaseElement(),
		Modal:       core.NewModal(),
		entries:     entries,
	}
	c.SetIdentifier(ChangelogModalId)
	c.SetAfterInitFunc(c.init)
	return c
}

func (c *Changelog) init() error {
	c.setLayout()
	c.setStyle()
	c.setKeybindings()
	c.setContent()
	c.handleEvents()
	return nil
}

func (c *Changelog) setLayout() {
	c.SetTitle(" Release Notes ")
	c.SetBorder(true)
	c.SetBorderPadding(0, 0, 1, 1)
}

func (c *Changelog) setContent() {
	hasBreaking := false
	for _, e := range c.entries {
		if e.Breaking {
			hasBreaking = true
			break
		}
	}

	c.SetText(c.buildText())

	if hasBreaking {
		c.AddButtons([]string{"Proceed", "Quit"})
	} else {
		c.AddButtons([]string{"Continue"})
	}

	c.SetDoneFunc(func(_ int, label string) {
		switch label {
		case "Quit":
			if c.onQuit != nil {
				c.onQuit()
			}
		default:
			if c.onProceed != nil {
				c.onProceed()
			}
		}
	})
}

func (c *Changelog) setKeybindings() {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		}
		return event
	})
}

func (c *Changelog) handleEvents() {
	go c.HandleEvents(c.GetIdentifier(), func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			c.setStyle()
		}
	})
}

func (c *Changelog) setStyle() {
	c.Modal.SetStyle(c.App.GetStyles())
	style := c.App.GetStyles()
	c.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(style.Global.FocusColor.Color()).
		Foreground(style.Global.BackgroundColor.Color()))
}

func (c *Changelog) buildText() string {
	var sb strings.Builder
	for i, e := range c.entries {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		if e.Breaking {
			sb.WriteString(fmt.Sprintf("[red::b]v%s — Breaking Change[::-]\n", e.Version))
		} else {
			sb.WriteString(fmt.Sprintf("[yellow::b]v%s[-]\n", e.Version))
		}
		sb.WriteString(fmt.Sprintf("[white::b]%s[::-]\n", e.Title))
		for _, ch := range e.Changes {
			sb.WriteString(fmt.Sprintf("  • %s\n", ch))
		}
	}
	return sb.String()
}

func (c *Changelog) SetOnProceed(fn func()) {
	c.onProceed = fn
}

func (c *Changelog) SetOnQuit(fn func()) {
	c.onQuit = fn
}

func (c *Changelog) Render() {
	c.App.Pages.AddPage(ChangelogModalId, c, true, true)
}
