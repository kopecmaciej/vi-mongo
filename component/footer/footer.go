package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Footer struct {
	*tview.Flex
	*tview.Table
}

func NewFooter() Footer {
	f := Footer{
		Flex: tview.NewFlex(),
	}
	return f
}

func (f *Footer) Init() {
	f.Flex.SetBackgroundColor(tcell.ColorDefault)
	f.Flex.SetDirection(tview.FlexColumn)

}
