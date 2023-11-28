package component

import (
	"context"
	"log"
	"mongo-ui/mongo"
	"strconv"
	"time"

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

	app      *App
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
	h.app = GetApp(ctx)

	h.setStyle()

	h.setBaseInfo(ctx)
	h.render()

	go h.Refresh(ctx)

	return nil
}

func (h *Header) setBaseInfo(ctx context.Context) {
	ss, err := h.dao.GetServerStatus(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	port := strconv.Itoa(h.dao.Config.Port)

	h.baseInfo = BaseInfo{
		0: {"Host", h.dao.Config.Host},
		1: {"Port", port},
		2: {"Database", h.dao.Config.Database},
		3: {"Collection", "-"},
		4: {"Version", ss.Version},
		5: {"Uptime", strconv.Itoa(int(ss.Uptime))},
		6: {"Connections", strconv.Itoa(int(ss.CurrentConns))},
		7: {"Available Connections", strconv.Itoa(int(ss.AvailableConns))},
		8: {"Resident Memory", strconv.Itoa(int(ss.Mem.Resident))},
		9: {"Virtual Memory", strconv.Itoa(int(ss.Mem.Virtual))},
	}
}

func (h *Header) Refresh(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			h.setBaseInfo(ctx)
			h.app.QueueUpdateDraw(func() {
				h.render()
			})
			time.Sleep(10 * time.Second)
		}
	}
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
func (h *Header) render() {
	b := h.baseInfo

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
