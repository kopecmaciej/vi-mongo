package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type (
	// KeyBindings is a way to define keybindings for the application
	// There are components that have only keybindings and some have
	// nested keybindings of their children components
	KeyBindings struct {
		Global    GlobalKeys    `json:"global"`
		Root      RootKeys      `json:"rootKeys"`
		Sidebar   SidebarKeys   `json:"sidebarKeys"`
		Connector ConnectorKeys `json:"connector"`
		Help      HelpKeys      `json:"helpKeys"`
	}

	// Key is a lowest level of keybindings
	// It holds the keys and runes that are used to trigger the action
	// and a description of the action that will be displayed in the help
	Key struct {
		Keys        []string `json:"keys,omitempty"`
		Runes       []string `json:"runes,omitempty"`
		Description string   `json:"description"`
	}

	// GlobalKeys is a struct that holds the global keybindings
	// for the application, they can be triggered from any component
	// as keys are passed from top to bottom
	GlobalKeys struct {
		ToggleHelp Key `json:"toggleHelp"`
	}

	RootKeys struct {
		FocusNext     Key         `json:"focusNext"`
		HideSidebar   Key         `json:"hideSidebar"`
		OpenConnector Key         `json:"openConnector"`
		SidebarKeys   SidebarKeys `json:"sidebar"`
		Content       ContentKeys `json:"content"`
	}

	SidebarKeys struct {
		FilterBar Key        `json:"filterBar"`
		DBTree    DBTreeKeys `json:"dbTree"`
	}

	DBTreeKeys struct {
		ExpandAll        Key `json:"expandAll"`
		CollapseAll      Key `json:"collapseAll"`
		ToggleExpand     Key `json:"toggleExpand"`
		AddCollection    Key `json:"addCollection"`
		DeleteCollection Key `json:"deleteCollection"`
	}

	ContentKeys struct {
		PeekDocument      Key          `json:"peekDocument"`
		ViewDocument      Key          `json:"viewDocument"`
		AddDocument       Key          `json:"addDocument"`
		EditDocument      Key          `json:"editDocument"`
		DuplicateDocument Key          `json:"duplicateDocument"`
		DeleteDocument    Key          `json:"deleteDocument"`
		Refresh           Key          `json:"refresh"`
		ToggleQuery       Key          `json:"toggleQuery"`
		NextPage          Key          `json:"nextPage"`
		PreviousPage      Key          `json:"previousPage"`
		InputBar          InputBarKeys `json:"inputBar"`
	}

	InputBarKeys struct {
		ShowHistory Key `json:"showHistory"`
		ClearInput  Key `json:"clearInput"`
	}

	ConnectorKeys struct {
		ConnectorForm ConnectorFormKeys `json:"connectorForm"`
		ConnectorList ConnectorListKeys `json:"connectorList"`
	}

	ConnectorFormKeys struct {
		FormFocusUp    Key `json:"formFocusUp"`
		FormFocusDown  Key `json:"formFocusDown"`
		SaveConnection Key `json:"saveConnection"`
		FocusList      Key `json:"focusList"`
	}

	ConnectorListKeys struct {
		FocusForm        Key `json:"focusForm"`
		DeleteConnection Key `json:"deleteConnection"`
		SetConnection    Key `json:"setConnection"`
	}

	HelpKeys struct {
		Close Key `json:"close"`
	}
)

func NewKeyBindings() KeyBindings {
	defaultKeyBindings := KeyBindings{
		Global: GlobalKeys{
			ToggleHelp: Key{
				Runes:       []string{"?"},
				Description: "Toggle help",
			},
		},
		Root: RootKeys{
			FocusNext: Key{
				Keys:        []string{"Tab"},
				Description: "Focus next component",
			},
			HideSidebar: Key{
				Keys:        []string{"Ctrl+B"},
				Description: "Hide sidebar",
			},
			OpenConnector: Key{
				Keys:        []string{"Ctrl+O"},
				Description: "Open connector",
			},
			SidebarKeys: SidebarKeys{
				FilterBar: Key{
					Runes:       []string{"/"},
					Description: "Focus filter bar",
				},
				DBTree: DBTreeKeys{
					ExpandAll: Key{
						Runes:       []string{"E"},
						Description: "Expand all",
					},
					CollapseAll: Key{
						Runes:       []string{"W"},
						Description: "Collapse all",
					},
					ToggleExpand: Key{
						Runes:       []string{"T"},
						Description: "Toggle expand",
					},
					AddCollection: Key{
						Runes:       []string{"A"},
						Description: "Add collection",
					},
					DeleteCollection: Key{
						Runes:       []string{"D"},
						Description: "Delete collection",
					},
				},
			},
			Content: ContentKeys{
				PeekDocument: Key{
					Runes:       []string{"P"},
					Keys:        []string{"Enter"},
					Description: "Peek document",
				},
				ViewDocument: Key{
					Runes:       []string{"V"},
					Description: "View document",
				},
				AddDocument: Key{
					Runes:       []string{"A"},
					Description: "Add document",
				},
				EditDocument: Key{
					Runes:       []string{"E"},
					Description: "Edit document",
				},
				DuplicateDocument: Key{
					Runes:       []string{"D"},
					Description: "Duplicate document",
				},
				DeleteDocument: Key{
					Keys:        []string{"Ctrl+D"},
					Description: "Delete document",
				},
				Refresh: Key{
					Keys:        []string{"Ctrl+R"},
					Description: "Refresh",
				},
				ToggleQuery: Key{
					Runes:       []string{"/"},
					Description: "Toggle query",
				},
				NextPage: Key{
					Keys:        []string{"Ctrl+N"},
					Description: "Next page",
				},
				PreviousPage: Key{
					Keys:        []string{"Ctrl+B"},
					Description: "Previous page",
				},
				InputBar: InputBarKeys{
					ShowHistory: Key{
						Keys:        []string{"Ctrl+H"},
						Description: "Show history",
					},
					ClearInput: Key{
						Keys:        []string{"Ctrl+L"},
						Description: "Clear input",
					},
				},
			},
		},
		Connector: ConnectorKeys{
			ConnectorForm: ConnectorFormKeys{
				FormFocusUp: Key{
					Keys:        []string{"Up"},
					Description: "Move form focus up",
				},
				FormFocusDown: Key{
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
			ConnectorList: ConnectorListKeys{
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
		},
		Help: HelpKeys{
			Close: Key{
				Keys:        []string{"Esc"},
				Description: "Close help",
			},
		},
	}

	customKeyBindings, err := defaultKeyBindings.LoadCustomKeyBindings("keybindings.json")
	if err != nil {
		return defaultKeyBindings
	}

	v := reflect.ValueOf(&defaultKeyBindings).Elem()
	cv := reflect.ValueOf(customKeyBindings).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		cfield := cv.Field(i)
		if cfield.Kind() == reflect.Struct {
			for j := 0; j < field.NumField(); j++ {
				keyField := field.Field(j)
				ckeyField := cfield.Field(j)
				if ckeyField.Kind() == reflect.Struct {
					for k := 0; k < keyField.NumField(); k++ {
						key := keyField.Field(k)
						ckey := ckeyField.Field(k)
						if ckey.Kind() == reflect.Slice {
							if ckey.Len() > 0 {
								key.Set(ckey)
							}
						}
					}
				}
			}
		}
	}

	return defaultKeyBindings
}

// LoadCustomKeyBindings loads custom keybindings from the config file
func (kb *KeyBindings) LoadCustomKeyBindings(path string) (keyBindings *KeyBindings, err error) {
	customKeyBindings := &KeyBindings{}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return keyBindings, err
	}
	err = json.Unmarshal(bytes, customKeyBindings)
	if err != nil {
		return keyBindings, err
	}

	return customKeyBindings, nil
}

type OrderedKeys struct {
	Component string
	Keys      []Key
}

// GetKeysForComponent returns keys for component
func (kb KeyBindings) GetKeysForComponent(component string) ([]OrderedKeys, error) {
	keys := make([]OrderedKeys, 0)
	if component == "" {
		return nil, fmt.Errorf("component is empty")
	}

	v := reflect.ValueOf(kb)
	field := v.FieldByName(component)

	if !field.IsValid() || field.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s not found", component) // Return nil if the field doesn't exist or isn't a struct.
	}

	var iterateOverField func(field reflect.Value, c string)
	iterateOverField = func(field reflect.Value, c string) {
		key := OrderedKeys{Component: c, Keys: make([]Key, 0)}
		keys = append(keys, key)
		for i := 0; i < field.NumField(); i++ {
			keyField := field.Field(i)
			if keyField.Type() == reflect.TypeOf(Key{}) {
				keys[len(keys)-1].Keys = append(keys[len(keys)-1].Keys, keyField.Interface().(Key))
			} else {
				iterateOverField(keyField, field.Type().Field(i).Name)
			}
		}
	}

	iterateOverField(field, component)

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

// Contains checks if the keybindings contains the key
func (kb *KeyBindings) Contains(configKey Key, namedKey string) bool {
	if namedKey == "Rune[ ]" {
		namedKey = "Space"
	}

	if strings.HasPrefix(namedKey, "Rune") {
		namedKey = strings.TrimPrefix(namedKey, "Rune")
		for _, k := range configKey.Runes {
			if k == namedKey[1:2] {
				return true
			}
		}
	}

	for _, k := range configKey.Keys {
		if k == namedKey {
			return true
		}
	}

	return false
}
