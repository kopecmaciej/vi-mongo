package modal

import (
	"context"
	"fmt"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"
)

const ServerInfoModalId = "ServerInfoModal"

type ServerInfoModal struct {
	*core.BaseElement
	*primitives.ViewModal

	dao *mongo.Dao
}

func NewServerInfoModal(dao *mongo.Dao) *ServerInfoModal {
	s := &ServerInfoModal{
		BaseElement: core.NewBaseElement(),
		ViewModal:   primitives.NewViewModal(),
		dao:         dao,
	}

	s.SetIdentifier(ServerInfoModalId)
	s.SetTitle("Server Info")
	return s
}

func (s *ServerInfoModal) Init(app *core.App) error {
	s.App = app
	s.setStyle()
	return nil
}

func (s *ServerInfoModal) setStyle() {
	s.ViewModal.SetBackgroundColor(s.App.GetStyles().Global.BackgroundColor.Color())
	s.ViewModal.SetTextColor(s.App.GetStyles().Global.TextColor.Color())
	s.ViewModal.SetButtonBackgroundColor(s.App.GetStyles().Global.BackgroundColor.Color())
	s.ViewModal.SetButtonTextColor(s.App.GetStyles().Global.TextColor.Color())
}

func (s *ServerInfoModal) Render(ctx context.Context) error {
	ss, err := s.dao.GetServerStatus(ctx)
	if err != nil {
		return err
	}

	info := map[string]string{
		"Host":                  s.dao.Config.Host,
		"Port":                  fmt.Sprintf("%d", s.dao.Config.Port),
		"Database":              s.dao.Config.Database,
		"Version":               ss.Version,
		"Uptime":                fmt.Sprintf("%d seconds", ss.Uptime),
		"Current Connections":   fmt.Sprintf("%d", ss.CurrentConns),
		"Available Connections": fmt.Sprintf("%d", ss.AvailableConns),
		"Resident Memory":       fmt.Sprintf("%d MB", ss.Mem.Resident),
		"Virtual Memory":        fmt.Sprintf("%d MB", ss.Mem.Virtual),
		"Is Master":             fmt.Sprintf("%v", ss.Repl.IsMaster),
	}

	content := ""
	for key, value := range info {
		content += fmt.Sprintf("[%s]%s[%s] %s\n", s.App.GetStyles().Others.ModalTextColor.Color(), key, s.App.GetStyles().Others.ModalSecondaryTextColor.Color(), value)
	}

	s.ViewModal.SetText(primitives.Text{
		Content: content,
		Align:   tview.AlignLeft,
	})
	s.ViewModal.AddButtons([]string{"Close"})
	s.ViewModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		s.App.Pages.RemovePage(ServerInfoModalId)
	})

	return nil
}
