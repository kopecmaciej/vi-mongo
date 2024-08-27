package core

import (
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
)

type Pages struct {
	*tview.Pages

	manager *manager.ElementManager
	app     *App
}

func NewPages(manager *manager.ElementManager, app *App) *Pages {
	return &Pages{
		Pages:   tview.NewPages(),
		manager: manager,
		app:     app,
	}
}

// AddPage is a wrapper for tview.Pages.AddPage
func (r *Pages) AddPage(view manager.ElementId, page tview.Primitive, resize, visable bool) *tview.Pages {
	if r.Pages.HasPage(string(view)) && r.manager.CurrentElement() == view {
		return r.Pages
	}
	r.manager.PushElement(view)
	r.app.SetPreviousFocus()
	r.Pages.AddPage(string(view), page, resize, visable)
	return r.Pages
}

// RemovePage is a wrapper for tview.Pages.RemovePage
func (r *Pages) RemovePage(view manager.ElementId) *tview.Pages {
	r.manager.PopElement()
	r.Pages.RemovePage(string(view))
	r.app.GiveBackFocus()
	return r.Pages
}

// HasPage is a wrapper for tview.Pages.HasPage
func (r *Pages) HasPage(view manager.ElementId) bool {
	return r.Pages.HasPage(string(view))
}
