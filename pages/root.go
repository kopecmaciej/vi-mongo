package pages

import (
	"context"
	"mongo-ui/component"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Root struct {
	*tview.Pages

	mongoDao *mongo.Dao
	header   *component.Header
	body     *component.Body
}

func NewRoot(mongoDao *mongo.Dao) *Root {
	return &Root{
		Pages:    tview.NewPages(),
		header:   component.NewHeader(mongoDao),
		body:     component.NewBody(mongoDao),
		mongoDao: mongoDao,
	}
}

func (r *Root) Init(ctx context.Context) *tview.Pages {
	rootPage := tview.NewFlex()
	rootPage.SetDirection(tview.FlexRow)
	rootPage.SetBackgroundColor(tcell.ColorDefault)

	rootPage.AddItem(r.header.Init(), 0, 1, false)
	rootPage.AddItem(r.body.Init(ctx), 0, 5, false)

	r.AddPage("main", rootPage, true, true)

	return r.Pages
}
