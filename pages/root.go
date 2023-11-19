package pages

import (
	"context"
	"mongo-ui/component/body"
	"mongo-ui/component/header"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Root struct {
	*tview.Pages

	mongoDao *mongo.Dao
	header   *header.Header
	body     *body.Body
}

func NewRoot(mongoDao *mongo.Dao) *Root {
	return &Root{
		Pages:    tview.NewPages(),
		header:   header.NewHeader(mongoDao),
		body:     body.NewBody(mongoDao),
		mongoDao: mongoDao,
	}
}

func (r *Root) Init(ctx context.Context) *tview.Pages {
	r.SetBackgroundColor(tcell.ColorDefault)
	rootPage := tview.NewFlex()
	rootPage.SetDirection(tview.FlexRow)
	rootPage.SetBackgroundColor(tcell.ColorDefault)

	r.header.Init()

	rootPage.AddItem(r.header.Table, 0, 1, false)
	rootPage.AddItem(r.body.Init(ctx), 0, 7, false)

	r.AddPage("main", rootPage, true, true)

	return r.Pages
}
