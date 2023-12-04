package component

import (
	"context"

	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
)

type InputModal struct {
	*tview.Modal

	app   *App
	dao   *mongo.Dao
	label string
}

func NewInputModal(dao *mongo.Dao, label string) *InputModal {
	return &InputModal{
		Modal: tview.NewModal(),
		dao:   dao,
		label: label,
	}
}

func (i *InputModal) Init(ctx context.Context) {
	i.app = GetApp(ctx)
	i.SetTitle(i.label)
	i.SetBorder(true)
}
