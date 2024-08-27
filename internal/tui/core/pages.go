package core

import (
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
)

type Pages struct {
	*tview.Pages

	manager *manager.ViewManager
	app     *App
}

func NewPages(manager *manager.ViewManager, app *App) *Pages {
	return &Pages{
		Pages:   tview.NewPages(),
		manager: manager,
		app:     app,
	}
}

// AddPage is a wrapper for tview.Pages.AddPage
func (r *Pages) AddPage(view manager.ViewIdentifier, page tview.Primitive, resize, visable bool) *tview.Pages {
	if r.Pages.HasPage(string(view)) && r.manager.CurrentView() == view {
		return r.Pages
	}
	r.manager.PushView(view)
	r.app.SetPreviousFocus()
	r.Pages.AddPage(string(view), page, resize, visable)
	return r.Pages
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Pages) RemovePage(view manager.ViewIdentifier) *tview.Pages {
	r.manager.PopView()
	r.Pages.RemovePage(string(view))
	r.app.GiveBackFocus()
	return r.Pages
}

// HasPage is a wrapper for tview.Pages.HasPage
func (r *Pages) HasPage(view manager.ViewIdentifier) bool {
	return r.Pages.HasPage(string(view))
}
