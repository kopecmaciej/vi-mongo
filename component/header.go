package component

import (
	"context"
	"mongo-ui/mongo"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type order int

type BaseInfo map[order]struct {
	label string
	value string
}

type Header struct {
	*tview.Table

	label    string
	dao      *mongo.Dao
	baseInfo BaseInfo
}

func NewHeader(dao *mongo.Dao) *Header {
	h := Header{
		Table: tview.NewTable(),
		dao:   dao,
		label: "header",
	}

	return &h
}

func (h *Header) Init(ctx context.Context) error {
	ss, err := h.dao.GetServerStatus(ctx)
	if err != nil {
		panic(err)
	}

	h.setStyle()

	port := strconv.Itoa(h.dao.Config.Port)
	b := BaseInfo{
		0: {"Host", h.dao.Config.Host},
		1: {"Port", port},
		2: {"Database", h.dao.Config.Database},
		3: {"Collection", "-"},
		4: {"Version", ss.Version},
	}

	h.SetBaseInfo(b)

	return nil
}

func (h *Header) setStyle() {
	h.Table.SetBackgroundColor(tcell.ColorDefault)
	h.Table.SetSelectable(false, false)
	h.Table.SetBorder(true)
	h.Table.SetBorderColor(tcell.ColorGreen)
	h.Table.SetBorderPadding(0, 0, 0, 0)
  h.Table.SetTitle(" Database Info ")
}

// set base information about database
func (h *Header) SetBaseInfo(b BaseInfo) {
	maxInRow := 3
	currCol := 0
	currRow := 0

	for i := 0; i < len(b); i++ {
		if i%maxInRow == 0 {
			currCol += 2
			currRow = 0
		}
		order := order(i)
		h.Table.SetCell(currRow, currCol, h.keyCell(b[order].label))
		h.Table.SetCell(currRow, currCol+1, h.valueCell(b[order].value))
		currRow++
	}

}

func (h *Header) keyCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text + ":")
	cell.SetBackgroundColor(tcell.ColorDefault)

	return cell
}

func (h *Header) valueCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text)
	cell.SetTextColor(tcell.ColorGreen)
	cell.SetBackgroundColor(tcell.ColorDefault)

	return cell
}
