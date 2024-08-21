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
		Root      RootKeys      `json:"root"`
		Connector ConnectorKeys `json:"connector"`
		Welcome   WelcomeKeys   `json:"welcome"`
		Help      HelpKeys      `json:"help"`
		DocPeeker DocPeekerKeys `json:"docPeeker"`
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
		ToggleFullScreenHelp Key `json:"toggleFullScreenHelp"`
		ToggleHelpBarFooter  Key `json:"toggleHelpBarFooter"`
	}

	RootKeys struct {
		FocusNext      Key           `json:"focusNext"`
		HideDatabases  Key           `json:"hideDatabases"`
		OpenConnector  Key           `json:"openConnector"`
		HideFooterHelp Key           `json:"closeFooterHelp"`
		Databases      DatabasesKeys `json:"databases"`
		Content        ContentKeys   `json:"content"`
	}

	DatabasesKeys struct {
		FilterBar        Key `json:"filterBar"`
		ExpandAll        Key `json:"expandAll"`
		CollapseAll      Key `json:"collapseAll"`
		ToggleExpand     Key `json:"toggleExpand"`
		AddCollection    Key `json:"addCollection"`
		DeleteCollection Key `json:"deleteCollection"`
	}

	ContentKeys struct {
		ExpandDocument    Key      `json:"expandDocument"`
		SwitchView        Key      `json:"switchView"`
		PeekDocument      Key      `json:"peekDocument"`
		ViewDocument      Key      `json:"viewDocument"`
		AddDocument       Key      `json:"addDocument"`
		EditDocument      Key      `json:"editDocument"`
		DuplicateDocument Key      `json:"duplicateDocument"`
		DeleteDocument    Key      `json:"deleteDocument"`
		CopyValue         Key      `json:"copyValue"`
		Refresh           Key      `json:"refresh"`
		ToggleQuery       Key      `json:"toggleQuery"`
		NextPage          Key      `json:"nextPage"`
		PreviousPage      Key      `json:"previousPage"`
		QueryBar          QueryBar `json:"queryBar"`
		SortBar           Key      `json:"sortBar"`
		CommandBar        Key      `json:"commandBar"`
	}

	QueryBar struct {
		ShowHistory Key `json:"showHistory"`
		ClearInput  Key `json:"clearInput"`
	}

	ConnectorKeys struct {
		ToggleFocus   Key               `json:"toggleFocus"`
		ConnectorForm ConnectorFormKeys `json:"connectorForm"`
		ConnectorList ConnectorListKeys `json:"connectorList"`
	}

	ConnectorFormKeys struct {
		SaveConnection Key `json:"saveConnection"`
		FocusList      Key `json:"focusList"`
	}

	ConnectorListKeys struct {
		FocusForm        Key `json:"focusForm"`
		DeleteConnection Key `json:"deleteConnection"`
		SetConnection    Key `json:"setConnection"`
	}

	WelcomeKeys struct {
		MoveFocusUp   Key `json:"moveFocusUp"`
		MoveFocusDown Key `json:"moveFocusDown"`
	}

	HelpKeys struct {
		Close Key `json:"close"`
	}

	DocPeekerKeys struct {
		MoveToTop    Key `json:"moveToTop"`
		MoveToBottom Key `json:"moveToBottom"`
		CopyFullLine Key `json:"copyFullLine"`
		CopyValue    Key `json:"copyValue"`
		Refresh      Key `json:"refresh"`
	}
)

// LoadKeybindings loads keybindings from the config file
// if the file does not exist it creates a new one with default keybindings
func LoadKeybindings() (*KeyBindings, error) {
	keybindings := &KeyBindings{}
	defaultKeybindings := &KeyBindings{}
	defaultKeybindings.loadDefaultKeybindings()

	keybindingsPath, err := getKeybindingsPath()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(keybindingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := ensureConfigDirExist()
			if err != nil {
				return nil, err
			}
			bytes, err = json.Marshal(defaultKeybindings)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(keybindingsPath, bytes, 0644)
			if err != nil {
				return nil, err
			}
			return defaultKeybindings, nil
		}
		return nil, err
	}

	err = json.Unmarshal(bytes, keybindings)
	if err != nil {
		return nil, err
	}

	MergeConfigs(keybindings, defaultKeybindings)

	return keybindings, nil
}

func (k *KeyBindings) loadDefaultKeybindings() {
	k.Global = GlobalKeys{
		ToggleFullScreenHelp: Key{
			Runes:       []string{"?"},
			Description: "Toggle full screen help",
		},
		ToggleHelpBarFooter: Key{
			Keys:        []string{"Ctrl+?"},
			Description: "Toggle help in footer",
		},
	}

	k.Root = RootKeys{
		FocusNext: Key{
			Keys:        []string{"Tab"},
			Description: "Focus next component",
		},
		HideDatabases: Key{
			Keys:        []string{"Ctrl+S"},
			Description: "Hide databases",
		},
		OpenConnector: Key{
			Keys:        []string{"Ctrl+O"},
			Description: "Open connector",
		},
		HideFooterHelp: Key{
			Keys:        []string{"Esc"},
			Description: "Hide footer help",
		},
	}

	k.Root.Databases = DatabasesKeys{
		FilterBar: Key{
			Runes:       []string{"/"},
			Description: "Focus filter bar",
		},
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
			Keys:        []string{"Ctrl+D"},
			Description: "Delete collection",
		},
	}

	k.Root.Content = ContentKeys{
		ExpandDocument: Key{
			Runes:       []string{"f"},
			Description: "Expand document",
		},
		SwitchView: Key{
			Runes:       []string{"g"},
			Description: "Switch view",
		},
		PeekDocument: Key{
			Runes:       []string{"p"},
			Keys:        []string{"Enter"},
			Description: "Peek document",
		},
		ViewDocument: Key{
			Runes:       []string{"v"},
			Description: "View document",
		},
		AddDocument: Key{
			Runes:       []string{"a"},
			Description: "Add document",
		},
		EditDocument: Key{
			Runes:       []string{"e"},
			Description: "Edit document",
		},
		DuplicateDocument: Key{
			Runes:       []string{"d"},
			Description: "Duplicate document",
		},
		DeleteDocument: Key{
			Keys:        []string{"Ctrl+D"},
			Description: "Delete document",
		},
		CopyValue: Key{
			Runes:       []string{"c"},
			Description: "Copy value",
		},
		Refresh: Key{
			Keys:        []string{"Ctrl+R"},
			Description: "Refresh",
		},
		ToggleQuery: Key{
			Runes:       []string{"/"},
			Description: "Toggle query",
		},
		SortBar: Key{
			Runes:       []string{"s"},
			Description: "Toggle sort",
		},
		CommandBar: Key{
			Runes:       []string{":"},
			Description: "Toggle command bar",
		},
		NextPage: Key{
			Keys:        []string{"Ctrl+N"},
			Description: "Next page",
		},
		PreviousPage: Key{
			Keys:        []string{"Ctrl+B"},
			Description: "Previous page",
		},
	}

	k.Root.Content.QueryBar = QueryBar{
		ShowHistory: Key{
			Keys:        []string{"Ctrl+Y"},
			Description: "Show history",
		},
		ClearInput: Key{
			Keys:        []string{"Ctrl+D"},
			Description: "Clear input",
		},
	}

	k.Connector.ToggleFocus = Key{
		Keys:        []string{"Tab", "Backtab"},
		Description: "Toggle focus",
	}

	k.Connector.ConnectorForm = ConnectorFormKeys{
		SaveConnection: Key{
			Keys:        []string{"Ctrl+S", "Enter"},
			Description: "Save connection",
		},
		FocusList: Key{
			Keys:        []string{"Esc", "Ctrl+Left"},
			Description: "Focus Connection List",
		},
	}

	k.Connector.ConnectorList = ConnectorListKeys{
		FocusForm: Key{
			Keys:        []string{"Ctrl+A", "Ctrl+Right"},
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
	}

	k.Welcome = WelcomeKeys{
		MoveFocusUp: Key{
			Keys:        []string{"Backtab"},
			Description: "Move focus up",
		},
		MoveFocusDown: Key{
			Keys:        []string{"Tab"},
			Description: "Move focus down",
		},
	}

	k.Help = HelpKeys{
		Close: Key{
			Keys:        []string{"Esc"},
			Description: "Close help",
		},
	}

	k.DocPeeker = DocPeekerKeys{
		MoveToTop: Key{
			Runes:       []string{"g"},
			Description: "Move to top",
		},
		MoveToBottom: Key{
			Runes:       []string{"G"},
			Description: "Move to bottom",
		},
		CopyFullLine: Key{
			Runes:       []string{"c"},
			Description: "Copy full object",
		},
		CopyValue: Key{
			Runes:       []string{"v"},
			Description: "Copy value",
		},
		Refresh: Key{
			Keys:        []string{"Ctrl+R"},
			Description: "Refresh document",
		},
	}
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
		return nil, fmt.Errorf("field %s not found", component)
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

func (k *Key) String() string {
	var keyString string
	var iter []string
	if len(k.Keys) > 0 {
		iter = k.Keys
	} else {
		iter = k.Runes
	}
	for i, k := range iter {
		if i == 0 {
			keyString = k
		} else {
			keyString = fmt.Sprintf("%s, %s", keyString, k)
		}
	}

	return keyString
}

func getKeybindingsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return configDir + "/keybindings.json", nil
}
