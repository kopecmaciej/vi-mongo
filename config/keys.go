package config

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type (
	KeyAction string

	KeyBindings struct {
		ConnectorForm ConnectorForm `json:"connectorForm"`
		ConnectorList ConnectorList `json:"connectorList"`
	}

	Key struct {
		Keys        []string `json:"keys,omitempty"`
		Runes       []string `json:"runes,omitempty"`
		Description string   `json:"description"`
	}

	ConnectorForm struct {
		MoveFocusUp    Key `json:"moveFocusUp"`
		MoveFocusDown  Key `json:"moveFocusDown"`
		SaveConnection Key `json:"saveConnection"`
		FocusList      Key `json:"focusList"`
	}

	ConnectorList struct {
		FocusForm        Key `json:"focusForm"`
		DeleteConnection Key `json:"deleteConnection"`
		SetConnection    Key `json:"setConnection"`
	}
)

func NewKeyBindings() KeyBindings {
	return KeyBindings{
		ConnectorForm: ConnectorForm{
			MoveFocusUp: Key{
				Keys:        []string{"Up"},
				Description: "Move form focus up",
			},
			MoveFocusDown: Key{
				Keys:        []string{"Down"},
				Description: "Move form focus down",
			},
			SaveConnection: Key{
				Keys:        []string{"Ctrl+S"},
				Description: "Save connection",
			},
			FocusList: Key{
				Keys:        []string{"Escape"},
				Description: "Focus Connection List",
			},
		},
		ConnectorList: ConnectorList{
			FocusForm: Key{
				Keys:        []string{"Ctrl+A"},
				Description: "Move focus to form",
			},
			DeleteConnection: Key{
				Keys:        []string{"Ctrl+D"},
				Description: "Delete selected connection",
			},
			SetConnection: Key{
				Keys:        []string{"Enter", "Space"},
				Description: "Set selected connection",
			},
		},
	}
}

// ConvertStrKeyToTcellKey converts string key to tcell key
func (kb *KeyBindings) ConvertStrKeyToTcellKey(key string) (tcell.Key, bool) {
	for k, v := range tcell.KeyNames {
		if v == key {
			return k, true
		}
	}
	return -1, false
}

func (kb *KeyBindings) Contains(key Key, val string) bool {
	if val == "Rune[ ]" {
		val = "Space"
	}

	if strings.HasPrefix(val, "Rune") {
		val = strings.TrimPrefix(val, "Rune")
		for _, k := range key.Runes {
			if k == val[1:2] {
				return true
			}
		}
	}

	for _, k := range key.Keys {
		if k == val {
			return true
		}
	}

	return false
}
