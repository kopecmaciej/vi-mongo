package modal

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const ChangelogModalId = "Changelog"

// ChangelogEntry describes changes introduced in a specific version.
type ChangelogEntry struct {
	Version     string
	Breaking    bool
	Title       string
	Changes     []string
	MigrationFn func() error
}

// Changelog is a startup modal showing release notes.
// Breaking entries show "Proceed" and "Quit" buttons; non-breaking show "Continue".
type Changelog struct {
	*core.BaseElement
	*tview.Box

	textView *tview.TextView
	form     *tview.Form

	entries     []ChangelogEntry
	onProceed   func()
	onQuit      func()
	hasBreaking bool

	lastContentWidth int
}

func NewChangelog(entries []ChangelogEntry) *Changelog {
	hasBreaking := false
	for _, e := range entries {
		if e.Breaking {
			hasBreaking = true
			break
		}
	}

	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	form := tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor)
	form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	form.SetBorderPadding(0, 0, 0, 0)

	c := &Changelog{
		BaseElement: core.NewBaseElement(),
		Box:         tview.NewBox(),
		textView:    tv,
		form:        form,
		entries:     entries,
		hasBreaking: hasBreaking,
	}
	c.SetIdentifier(ChangelogModalId)
	c.SetAfterInitFunc(c.init)
	return c
}

func (c *Changelog) init() error {
	c.setButtons()
	c.setStyle()
	go c.HandleEvents(c.GetIdentifier(), func(event manager.EventMsg) {
		if event.Message.Type == manager.StyleChanged {
			c.setStyle()
		}
	})
	return nil
}

func (c *Changelog) setButtons() {
	if c.hasBreaking {
		c.form.AddButton("Proceed", func() {
			if c.onProceed != nil {
				c.onProceed()
			}
		})
		c.form.AddButton("Quit", func() {
			if c.onQuit != nil {
				c.onQuit()
			}
		})
	} else {
		c.form.AddButton("Continue", func() {
			if c.onProceed != nil {
				c.onProceed()
			}
		})
	}

	c.form.SetCancelFunc(func() {
		if c.hasBreaking {
			if c.onQuit != nil {
				c.onQuit()
			}
		} else {
			if c.onProceed != nil {
				c.onProceed()
			}
		}
	})
}

func (c *Changelog) setStyle() {
	style := c.App.GetStyles()
	c.Box.SetBorder(true)
	c.Box.SetTitle(" Release Notes ")
	c.Box.SetBackgroundColor(style.Global.BackgroundColor.Color())
	c.Box.SetBorderColor(style.Global.BorderColor.Color())
	c.Box.SetTitleColor(style.Global.TitleColor.Color())

	c.textView.SetBackgroundColor(style.Global.BackgroundColor.Color())
	c.textView.SetTextColor(style.Global.TextColor.Color())

	c.form.SetBackgroundColor(style.Global.BackgroundColor.Color())
	c.form.SetButtonBackgroundColor(style.Others.ButtonsBackgroundColor.Color())
	c.form.SetButtonTextColor(style.Others.ButtonsTextColor.Color())

	activatedStyle := tcell.StyleDefault.
		Background(style.Global.FocusColor.Color()).
		Foreground(style.Global.BackgroundColor.Color())
	for i := 0; i < c.form.GetButtonCount(); i++ {
		c.form.GetButton(i).SetActivatedStyle(activatedStyle)
	}
}

func (c *Changelog) Draw(screen tcell.Screen) {
	screenW, screenH := screen.Size()

	const marginV = 4
	const buttonH = 3

	modalW := screenW * 3 / 5
	if modalW < 40 {
		modalW = 40
	}
	if modalW > 120 {
		modalW = 120
	}
	x := (screenW - modalW) / 2
	contentW := modalW - 2

	if contentW != c.lastContentWidth {
		c.textView.SetText(c.buildText(contentW))
		c.lastContentWidth = contentW
	}

	contentH := screenH - 2*marginV - buttonH - 2
	if contentH < 3 {
		contentH = 3
	}

	modalH := contentH + buttonH + 2
	y := (screenH - modalH) / 2
	if y < marginV {
		y = marginV
	}

	c.Box.SetRect(x, y, modalW, modalH)
	c.Box.DrawForSubclass(screen, c)

	c.textView.SetRect(x+1, y+1, contentW, contentH)
	c.textView.Draw(screen)

	c.form.SetRect(x+1, y+1+contentH, contentW, buttonH)
	c.form.Draw(screen)
}

func (c *Changelog) Focus(delegate func(tview.Primitive)) {
	delegate(c.form)
}

func (c *Changelog) HasFocus() bool {
	return c.form.HasFocus() || c.textView.HasFocus()
}

// activateButton triggers the callback for the currently focused button.
// This bypasses the form's internal Enter-as-Tab behavior in the tview fork.
func (c *Changelog) activateButton() {
	for i := 0; i < c.form.GetButtonCount(); i++ {
		btn := c.form.GetButton(i)
		if btn.HasFocus() {
			switch btn.GetLabel() {
			case "Continue", "Proceed":
				if c.onProceed != nil {
					c.onProceed()
				}
			case "Quit":
				if c.onQuit != nil {
					c.onQuit()
				}
			}
			return
		}
	}
	// Fallback: no button has focus, treat as proceed for non-breaking.
	if !c.hasBreaking && c.onProceed != nil {
		c.onProceed()
	}
}

func (c *Changelog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.Box.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		k := c.App.GetKeys()
		switch {
		case event.Key() == tcell.KeyEnter:
			c.activateButton()
		case k.Contains(k.Navigation.MoveDown, event.Name()):
			c.textView.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), setFocus)
		case k.Contains(k.Navigation.MoveUp, event.Name()):
			c.textView.InputHandler()(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), setFocus)
		case k.Contains(k.Navigation.MoveLeft, event.Name()):
			c.form.InputHandler()(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone), setFocus)
		case k.Contains(k.Navigation.MoveRight, event.Name()):
			c.form.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), setFocus)
		default:
			c.form.InputHandler()(event, setFocus)
		}
	})
}

func (c *Changelog) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return c.Box.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if consumed, capture = c.form.MouseHandler()(action, event, setFocus); consumed {
			return
		}
		consumed, capture = c.textView.MouseHandler()(action, event, setFocus)
		return
	})
}

// changeGroup holds a named set of changelog items that share the same category.
type changeGroup struct {
	label string
	items []string
}

var knownCategories = []struct {
	prefix string
	label  string
}{
	{"New", "Features"},
	{"Improvement", "Improvements"},
	{"Fix", "Bug Fixes"},
	{"Breaking", "Breaking Changes"},
}

func groupChanges(changes []string) []changeGroup {
	indexMap := map[string]int{}
	var groups []changeGroup

	for _, ch := range changes {
		matched := false
		for _, cat := range knownCategories {
			pfx := cat.prefix + ": "
			if strings.HasPrefix(ch, pfx) {
				text := strings.TrimPrefix(ch, pfx)
				if idx, ok := indexMap[cat.prefix]; ok {
					groups[idx].items = append(groups[idx].items, text)
				} else {
					indexMap[cat.prefix] = len(groups)
					groups = append(groups, changeGroup{label: cat.label, items: []string{text}})
				}
				matched = true
				break
			}
		}
		if !matched {
			const other = "__other__"
			if idx, ok := indexMap[other]; ok {
				groups[idx].items = append(groups[idx].items, ch)
			} else {
				indexMap[other] = len(groups)
				groups = append(groups, changeGroup{label: "Other", items: []string{ch}})
			}
		}
	}
	return groups
}

func (c *Changelog) buildText(contentWidth int) string {
	var sb strings.Builder

	for i, e := range c.entries {
		if i > 0 {
			sb.WriteString("\n[gray]" + strings.Repeat("─", contentWidth) + "[-]\n\n")
		}

		if e.Breaking {
			fmt.Fprintf(&sb, "[red::b]v%s — Breaking Change[::-]\n", e.Version)
		} else {
			fmt.Fprintf(&sb, "[yellow::b]v%s[-]\n", e.Version)
		}
		fmt.Fprintf(&sb, "[white::b]%s[::-]\n", e.Title)

		groups := groupChanges(e.Changes)
		for j, g := range groups {
			if j == 0 {
				sb.WriteString("\n")
			}
			fmt.Fprintf(&sb, "[blue::b]%s[::-]\n", g.label)
			for _, item := range g.items {
				fmt.Fprintf(&sb, "  [gray]•[-] %s\n", item)
			}
			if j < len(groups)-1 {
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func (c *Changelog) SetOnProceed(fn func()) { c.onProceed = fn }
func (c *Changelog) SetOnQuit(fn func())    { c.onQuit = fn }

func (c *Changelog) Render() {
	c.App.Pages.AddPage(ChangelogModalId, c, true, true)
}
