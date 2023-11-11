package component

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Content struct {
	*tview.Table

  app *tview.Application
	dao *mongo.Dao
}

func NewContent(dao *mongo.Dao) *Content {
	return &Content{
		Table: tview.NewTable(),

		dao: dao,
	}
}

func (c *Content) Init(ctx context.Context) error {
  c.app = ctx.Value("app").(*tview.Application)
	c.setStyle()
	return nil
}

func (c *Content) RenderDocuments(db, coll string) error {
	c.Clear()
  c.app.SetFocus(c)
	documents, count, err := c.dao.ListDocuments(db, coll)
	if err != nil {
		return err
	}

	if len(documents) == 0 {
		c.SetCell(1, 1, tview.NewTableCell("No documents found").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	}

	// first row is the header
	// show count of documents
	c.SetCell(0, 1, tview.NewTableCell("Documents: "+fmt.Sprint(count)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

	for i, d := range documents {
		if i > 50 {
			break
		}
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}

		c.SetCell(i+1, 1, tview.NewTableCell(string(jsonBytes)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

	}
	return nil
}

func (c *Content) setStyle() {
	c.SetBackgroundColor(tcell.ColorDefault)
	c.SetBorder(true)
	c.SetTitle("Content")
	c.SetTitleAlign(tview.AlignLeft)
	c.SetTitleColor(tcell.ColorSteelBlue)
	c.SetBorderColor(tcell.ColorSteelBlue)
	c.SetBorderPadding(0, 0, 3, 3)
	c.SetFixed(1, 1)
	c.SetSelectable(true, false)
}
