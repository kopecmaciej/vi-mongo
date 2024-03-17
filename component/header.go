package component

import (
	"context"
	"strconv"
	"time"

	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

type (
	order int

	info struct {
		label string
		value string
	}

	BaseInfo map[order]info

	// Header is a component that displays information about the database
	// in the header of the application
	Header struct {
		*Component
		*tview.Table

		style    *config.HeaderStyle
		baseInfo BaseInfo
	}
)

// NewHeader creates a new header component
func NewHeader() *Header {
	h := Header{
		Component: NewComponent("Header"),
		Table:     tview.NewTable(),
	}

	h.SetAfterInitFunc(h.init)

	return &h
}

func (h *Header) init() error {
	ctx := context.Background()
	h.setStyle()

	if err := h.setBaseInfo(ctx); err != nil {
		return err
	}
	h.render()

	go h.refresh()

	return nil
}

func (h *Header) setStyle() {
	h.style = &h.app.Styles.Header
	h.Table.SetBackgroundColor(h.style.BackgroundColor.Color())
	h.Table.SetBorderColor(h.style.BorderColor.Color())
	h.Table.SetSelectable(false, false)
	h.Table.SetBorder(true)
	h.Table.SetBorderPadding(0, 0, 1, 1)
	h.Table.SetTitle(" Database Info ")
}

// setBaseInfo sets the base information about the database
// such as status, host, port, database, version, uptime, connections, memory etc.
func (h *Header) setBaseInfo(ctx context.Context) error {
	ss, err := h.dao.GetServerStatus(ctx)
	if err != nil {
		return err
	}

	port := strconv.Itoa(h.dao.Config.Port)

	status := h.style.InactiveSymbol.String()
	if ss.Ok == 1 {
		status = h.style.ActiveSymbol.String()
	}

	h.baseInfo = BaseInfo{
		0: {"Status", status},
		1: {"Host", h.dao.Config.Host},
		2: {"Port", port},
		3: {"Database", h.dao.Config.Database},
		4: {"Version", ss.Version},
		5: {"Uptime", strconv.Itoa(int(ss.Uptime))},
		6: {"Connections", strconv.Itoa(int(ss.CurrentConns))},
		7: {"Available Connections", strconv.Itoa(int(ss.AvailableConns))},
		8: {"Resident Memory", strconv.Itoa(int(ss.Mem.Resident))},
		9: {"Virtual Memory", strconv.Itoa(int(ss.Mem.Virtual))},
	}

	return nil
}

// refresh refreshes the header component every 10 seconds
// to display the most recent information about the database
func (h *Header) refresh() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := h.setBaseInfo(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error while refreshing header")
			h.baseInfo[0] = info{"Status", h.style.InactiveSymbol.String()}
		}
		h.app.QueueUpdateDraw(func() {
			h.render()
		})
		time.Sleep(10 * time.Second)
	}
}

// render renders the header component
func (h *Header) render() {
	b := h.baseInfo

	maxInRow := 2
	currCol := 0
	currRow := 0

	for i := 0; i < len(b); i++ {
		if i%maxInRow == 0 && i != 0 {
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
	cell.SetBackgroundColor(h.style.BackgroundColor.Color())
	cell.SetTextColor(h.style.KeyColor.Color())

	return cell
}

func (h *Header) valueCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text)
	cell.SetBackgroundColor(h.style.BackgroundColor.Color())
	cell.SetTextColor(h.style.ValueColor.Color())

	return cell
}
