package modal

import (
	"context"
	"fmt"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const ServerInfoModalView = "ServerInfoModal"

type ServerInfoModal struct {
	*core.BaseElement
	*tview.Modal

	dao *mongo.Dao
}

func NewServerInfoModal(dao *mongo.Dao) *ServerInfoModal {
	s := &ServerInfoModal{
		BaseElement: core.NewBaseElement(),
		Modal:       tview.NewModal(),
		dao:         dao,
	}

	s.SetIdentifier(ServerInfoModalView)
	return s
}

func (s *ServerInfoModal) Init(app *core.App) error {
	s.App = app
	s.setStyle()
	return nil
}

func (s *ServerInfoModal) setStyle() {
	s.Modal.SetBackgroundColor(s.App.GetStyles().Global.BackgroundColor.Color())
	s.Modal.SetTextColor(s.App.GetStyles().Global.TextColor.Color())
	s.Modal.SetButtonBackgroundColor(s.App.GetStyles().Global.BackgroundColor.Color())
	s.Modal.SetButtonTextColor(s.App.GetStyles().Global.TextColor.Color())
}

func (s *ServerInfoModal) Render(ctx context.Context) error {
	ss, err := s.dao.GetServerStatus(ctx)
	if err != nil {
		return err
	}

	info := fmt.Sprintf(`Server Information:
Host: %s
Port: %d
Database: %s
Version: %s
Uptime: %d seconds
Current Connections: %d
Available Connections: %d
Resident Memory: %d MB
Virtual Memory: %d MB
Is Master: %v`,
		s.dao.Config.Host,
		s.dao.Config.Port,
		s.dao.Config.Database,
		ss.Version,
		ss.Uptime,
		ss.CurrentConns,
		ss.AvailableConns,
		ss.Mem.Resident,
		ss.Mem.Virtual,
		ss.Repl.IsMaster,
	)

	s.Modal.SetText(info)
	s.Modal.AddButtons([]string{"Close"})
	s.Modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		s.App.Pages.RemovePage(ServerInfoModalView)
	})

	return nil
}
