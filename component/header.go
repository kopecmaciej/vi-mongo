package component

import (
	"context"
	"mongo-ui/mongo"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type order int

type BaseInfo map[order]struct {
	label string
	value string
}

type Header struct {
	*tview.Flex
	*tview.Table

	dao      *mongo.Dao
	baseInfo BaseInfo
}

func NewHeader(dao *mongo.Dao) *Header {
	h := Header{
		Flex:  tview.NewFlex(),
		Table: tview.NewTable(),
		dao:   dao,
	}

	return &h
}

func (h *Header) Init() *tview.Flex {
  ctx := context.Background()
	ss, err := h.dao.GetServerStatus(ctx)
	if err != nil {
		panic(err)
	}

	h.Flex.SetBackgroundColor(tcell.ColorDefault)
	h.Flex.SetDirection(tview.FlexColumn)

	// add database information
	h.Table.SetBackgroundColor(tcell.ColorDefault)
	h.Table.SetBorders(false)

	b := BaseInfo{
		0: {"Host", "localhost"},
		1: {"Port", "27017"},
		2: {"Database", "test"},
		3: {"Collection", "restaurants"},
		4: {"Version", ss.Version},
	}

	h.SetBaseInfo(b)

	h.Flex.AddItem(h.Table, 0, 1, false)

	return h.Flex
}

// set base information about database
func (h *Header) SetBaseInfo(b BaseInfo) {

	for i := 0; i < len(b); i++ {
		order := order(i)
		h.Table.SetCell(i, 0, h.infoCell(b[order].label))
		h.Table.SetCell(i, 1, h.valueCell(b[order].value))
	}
}

func (h *Header) infoCell(text string) *tview.TableCell {
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
