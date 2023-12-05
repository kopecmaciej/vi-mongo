package component

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/mongo"
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
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	h.app = app

	h.setStyle()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err = h.setBaseInfo(ctxWithTimeout); err != nil {
		return err
	}
	h.render()

	go h.Refresh(ctxWithTimeout)

	return nil
}

func (h *Header) setBaseInfo(ctx context.Context) error {
	ss, err := h.dao.GetServerStatus(ctx)
	if err != nil {
		return err
	}

	port := strconv.Itoa(h.dao.Config.Port)

	status := "○"
	if ss.Ok == 1 {
		status = "●"
	}

	h.baseInfo = BaseInfo{
		0:  {"Status", status},
		1:  {"Host", h.dao.Config.Host},
		2:  {"Port", port},
		3:  {"Database", h.dao.Config.Database},
		4:  {"Collection", "-"},
		5:  {"Version", ss.Version},
		6:  {"Uptime", strconv.Itoa(int(ss.Uptime))},
		7:  {"Connections", strconv.Itoa(int(ss.CurrentConns))},
		8:  {"Available Connections", strconv.Itoa(int(ss.AvailableConns))},
		9:  {"Resident Memory", strconv.Itoa(int(ss.Mem.Resident))},
		10: {"Virtual Memory", strconv.Itoa(int(ss.Mem.Virtual))},
	}

	return nil
}

func (h *Header) Refresh(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := h.setBaseInfo(ctx)
			if err != nil {
				log.Println(err)
				ctx.Done()
			}
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
