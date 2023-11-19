package custom

import (
	"github.com/rivo/tview"
)

type Flex struct {
	*tview.Flex
}

func NewFlex() *Flex {
	return &Flex{
		Flex: tview.NewFlex(),
	}
}

func AddItemAt(flex *tview.Flex, item tview.Primitive, fixedSize, proportion int, position int) {
	count := flex.GetItemCount()
	if position > count {
		return
	}
	if position == count {
		flex.AddItem(item, fixedSize, proportion, false)
	}
	items := make([]tview.Primitive, count)
	for i := 0; i < count; i++ {
		items[i] = flex.GetItem(i)
	}
	flex.Clear()
	newItems := make([]tview.Primitive, count+1)
	for i := 0; i < count+1; i++ {
		if i < position {
			newItems[i] = items[i]
		} else if i == position {
			newItems[i] = item
		} else {
			newItems[i] = items[i-1]
		}
	}
  for i := 0; i < count+1; i++ {
    flex.AddItem(newItems[i], fixedSize, proportion, false)
  }
}
