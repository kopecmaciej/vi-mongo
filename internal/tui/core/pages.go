package core

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
)

type Pages struct {
	*tview.Pages

	manager *manager.ElementManager
	app     *App
}

func (p *Pages) SetStyle(style *config.Styles) {
	p.Pages.SetBackgroundColor(style.Global.BackgroundColor.Color())
	p.Pages.SetBorderColor(style.Global.BorderColor.Color())
	p.Pages.SetTitleColor(style.Global.TitleColor.Color())
	p.Pages.SetFocusStyle(tcell.StyleDefault.Foreground(style.Global.FocusColor.Color()).Background(style.Global.BackgroundColor.Color()))
}

func NewPages(manager *manager.ElementManager, app *App) *Pages {
	return &Pages{
		Pages:   tview.NewPages(),
		manager: manager,
		app:     app,
	}
}

// AddPage is a wrapper for tview.Pages.AddPage
func (r *Pages) AddPage(view tview.Identifier, page tview.Primitive, resize, visable bool) *tview.Pages {
	r.app.SetPreviousFocus()
	r.Pages.AddPage(string(view), page, resize, visable)
	if visable && page.HasFocus() {
		r.app.FocusChanged(page)
	}
	return r.Pages
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Pages) RemovePage(view tview.Identifier) *tview.Pages {
	r.Pages.RemovePage(string(view))
	r.app.GiveBackFocus()
	return r.Pages
}

// HasPage is a wrapper for tview.Pages.HasPage
func (r *Pages) HasPage(view tview.Identifier) bool {
	return r.Pages.HasPage(string(view))
}
