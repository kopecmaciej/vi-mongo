package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type (
	KeyBindings struct {
		Global        Global        `json:"global"`
		RootKeys      RootKeys      `json:"rootKeys"`
		SidebarKeys   SidebarKeys   `json:"sidebarKeys"`
		Contents      Contents      `json:"contents"`
		DBTree        DBTree        `json:"dbTree"`
		ConnectorForm ConnectorForm `json:"connectorForm"`
		ConnectorList ConnectorList `json:"connectorList"`
		HelpKeys      HelpKeys      `json:"helpKeys"`
	}

	Key struct {
		Keys        []string `json:"keys,omitempty"`
		Runes       []string `json:"runes,omitempty"`
		Description string   `json:"description"`
	}

	Global struct {
		ToggleHelp Key `json:"toggleHelp"`
	}

	RootKeys struct {
		FocusNext     Key `json:"focusNext"`
		HideSidebar   Key `json:"hideSidebar"`
		OpenConnector Key `json:"openConnector"`
	}

	SidebarKeys struct {
		FilterBar Key `json:"filterBar"`
	}

	Contents struct {
		PeekDocument      Key `json:"peekDocument"`
		ViewDocument      Key `json:"viewDocument"`
		AddDocument       Key `json:"addDocument"`
		EditDocument      Key `json:"editDocument"`
		DuplicateDocument Key `json:"duplicateDocument"`
		DeleteDocument    Key `json:"deleteDocument"`
		Refresh           Key `json:"refresh"`
		ToggleQuery       Key `json:"toggleQuery"`
		NextPage          Key `json:"nextPage"`
		PreviousPage      Key `json:"previousPage"`
	}

	DBTree struct {
		ExpandAll        Key `json:"expandAll"`
		CollapseAll      Key `json:"collapseAll"`
		ToggleExpand     Key `json:"toggleExpand"`
		AddCollection    Key `json:"addCollection"`
		DeleteCollection Key `json:"deleteCollection"`
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

	HelpKeys struct {
		Close Key `json:"close"`
	}
)

func NewKeyBindings() KeyBindings {
	return KeyBindings{
		Global: Global{
			ToggleHelp: Key{
				Runes:       []string{"?"},
				Description: "Toggle help",
			},
		},
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
				Keys:        []string{"Esc"},
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

// GetKeysForComponent returns keys for component
func (kb KeyBindings) GetKeysForComponent(component string) ([]Key, error) {
	var keys []Key

	v := reflect.ValueOf(kb)
	field := v.FieldByName(component)

	if !field.IsValid() || field.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s not found", component) // Return nil if the field doesn't exist or isn't a struct.
	}

	for i := 0; i < field.NumField(); i++ {
		keyField := field.Field(i)
		if keyField.Type() == reflect.TypeOf(Key{}) {
			keys = append(keys, keyField.Interface().(Key))
		}
	}

	return keys, nil
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
