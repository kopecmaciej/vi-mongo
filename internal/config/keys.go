package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/util"
)

type (
	OrderedKeys struct {
		Element string
		Keys    []Key
	}
	// KeyBindings is a way to define keybindings for the application
	// There are views that have only keybindings and some have
	// nested keybindings of their children views
	KeyBindings struct {
		Global     GlobalKeys     `json:"global"`
		Help       HelpKeys       `json:"help"`
		Welcome    WelcomeKeys    `json:"welcome"`
		Connection ConnectionKeys `json:"connection"`
		Main       MainKeys       `json:"main"`
		Database   DatabaseKeys   `json:"databases"`
		Content    ContentKeys    `json:"content"`
		Peeker     PeekerKeys     `json:"peeker"`
		QueryBar   QueryBar       `json:"queryBar"`
		SortBar    SortBar        `json:"sortBar"`
		Index      IndexKeys      `json:"index"`
		AIQuery    AIQuery        `json:"aiPrompt"`
		History    HistoryKeys    `json:"history"`
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
	// for the application, they can be triggered from any view
	// as keys are passed from top to bottom
	GlobalKeys struct {
		ToggleFullScreenHelp Key `json:"toggleFullScreenHelp"`
		OpenConnection       Key `json:"openConnection"`
		ShowStyleModal       Key `json:"showStyleModal"`
	}

	MainKeys struct {
		FocusNext      Key `json:"focusNext"`
		FocusPrevious  Key `json:"focusPrevious"`
		HideDatabase   Key `json:"hideDatabases"`
		ShowAIQuery    Key `json:"showAIQuery"`
		ShowServerInfo Key `json:"showServerInfo"`
		ShowShell      Key `json:"showShell"`
	}

	DatabaseKeys struct {
		FilterBar        Key `json:"filterBar"`
		ExpandAll        Key `json:"expandAll"`
		CollapseAll      Key `json:"collapseAll"`
		AddCollection    Key `json:"addCollection"`
		DeleteCollection Key `json:"deleteCollection"`
	}

	ContentKeys struct {
		ChangeView        Key `json:"switchView"`
		PeekDocument      Key `json:"peekDocument"`
		FullPagePeek      Key `json:"fullPagePeek"`
		AddDocument       Key `json:"addDocument"`
		EditDocument      Key `json:"editDocument"`
		DuplicateDocument Key `json:"duplicateDocument"`
		DeleteDocument    Key `json:"deleteDocument"`
		CopyHighlight     Key `json:"copyValue"`
		CopyDocument      Key `json:"copyDocument"`
		Refresh           Key `json:"refresh"`
		ToggleQueryBar    Key `json:"toggleQueryBar"`
		NextDocument      Key `json:"nextDocument"`
		PreviousDocument  Key `json:"previousDocument"`
		NextPage          Key `json:"nextPage"`
		PreviousPage      Key `json:"previousPage"`
		ToggleSortBar     Key `json:"toggleSortBar"`

		// MultipleSelect    Key      `json:"multipleSelect"`
		// ClearSelection   Key      `json:"clearSelection"`
	}

	QueryBar struct {
		ShowHistory Key `json:"showHistory"`
		ClearInput  Key `json:"clearInput"`
		Paste       Key `json:"paste"`
	}

	SortBar struct {
		ClearInput Key `json:"clearInput"`
		Paste      Key `json:"paste"`
	}

	ConnectionKeys struct {
		ToggleFocus    Key                `json:"toggleFocus"`
		ConnectionForm ConnectionFormKeys `json:"connectionForm"`
		ConnectionList ConnectionListKeys `json:"connectionList"`
	}

	ConnectionFormKeys struct {
		SaveConnection Key `json:"saveConnection"`
		FocusList      Key `json:"focusList"`
	}

	ConnectionListKeys struct {
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

	PeekerKeys struct {
		MoveToTop        Key `json:"moveToTop"`
		MoveToBottom     Key `json:"moveToBottom"`
		CopyHighlight    Key `json:"copyHighlight"`
		CopyValue        Key `json:"copyValue"`
		ToggleFullScreen Key `json:"toggleFullScreen"`
		Exit             Key `json:"exit"`
	}

	HistoryKeys struct {
		ClearHistory Key `json:"clearHistory"`
		AcceptEntry  Key `json:"acceptEntry"`
		CloseHistory Key `json:"closeHistory"`
	}

	IndexKeys struct {
		ExitAddIndex Key `json:"exitModal"`
		AddIndex     Key `json:"addIndex"`
		DeleteIndex  Key `json:"deleteIndex"`
	}

	AIQuery struct {
		ExitAIQuery Key `json:"exitAIQuery"`
		ClearPrompt Key `json:"clearPrompt"`
	}
)

func (k *KeyBindings) loadDefaults() {
	k.Global = GlobalKeys{
		ToggleFullScreenHelp: Key{
			Runes:       []string{"?"},
			Description: "Toggle full screen help",
		},
		OpenConnection: Key{
			Keys:        []string{"Ctrl+O"},
			Description: "Open connection page",
		},
		ShowStyleModal: Key{
			Keys:        []string{"Ctrl+T"},
			Description: "Toggle style change modal",
		},
	}

	k.Main = MainKeys{
		FocusNext: Key{
			Keys:        []string{"Ctrl+L", "Tab"},
			Description: "Focus next component",
		},
		FocusPrevious: Key{
			Keys:        []string{"Ctrl+H", "Backtab"},
			Description: "Focus previous component",
		},
		HideDatabase: Key{
			Keys:        []string{"Ctrl+N"},
			Description: "Hide databases",
		},
		ShowServerInfo: Key{
			Keys:        []string{"Ctrl+I"},
			Description: "Show server info",
		},
		ShowAIQuery: Key{
			Keys:        []string{"Ctrl+A"},
			Description: "Show AI prompt",
		},
		ShowShell: Key{
			Keys:        []string{"Ctrl+S"},
			Description: "Show shell",
		},
	}

	k.Database = DatabaseKeys{
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
		AddCollection: Key{
			Runes:       []string{"A"},
			Description: "Add collection",
		},
		DeleteCollection: Key{
			Runes:       []string{"D"},
			Description: "Delete collection",
		},
	}

	k.Content = ContentKeys{
		ChangeView: Key{
			Runes:       []string{"f"},
			Description: "Change view",
		},
		PeekDocument: Key{
			Runes:       []string{"p"},
			Keys:        []string{"Enter"},
			Description: "Quick peek",
		},
		FullPagePeek: Key{
			Runes:       []string{"P"},
			Description: "Full page peek",
		},
		AddDocument: Key{
			Runes:       []string{"A"},
			Description: "Add new",
		},
		EditDocument: Key{
			Runes:       []string{"E"},
			Description: "Edit",
		},
		DuplicateDocument: Key{
			Runes:       []string{"D"},
			Description: "Duplicate",
		},
		DeleteDocument: Key{
			Runes:       []string{"d"},
			Description: "Delete",
		},
		// MultipleSelect: Key{
		// 	Runes:       []string{"v"},
		// 	Description: "Multiple select",
		// },
		// ClearSelection: Key{
		// 	Runes:       []string{"C"},
		// 	Description: "Clear selection",
		// },
		CopyHighlight: Key{
			Runes:       []string{"c"},
			Description: "Copy highlighted",
		},
		CopyDocument: Key{
			Runes:       []string{"C"},
			Description: "Copy document",
		},
		Refresh: Key{
			Runes:       []string{"R"},
			Description: "Refresh",
		},
		ToggleQueryBar: Key{
			Runes:       []string{"/"},
			Description: "Toggle query bar",
		},
		ToggleSortBar: Key{
			Runes:       []string{"s"},
			Description: "Toggle sort bar",
		},
		NextDocument: Key{
			Runes:       []string{"]"},
			Description: "Next document",
		},
		PreviousDocument: Key{
			Runes:       []string{"["},
			Description: "Previous document",
		},
		NextPage: Key{
			Runes:       []string{"n"},
			Description: "Next page",
		},
		PreviousPage: Key{
			Runes:       []string{"b"},
			Description: "Previous page",
		},
	}

	k.QueryBar = QueryBar{
		ShowHistory: Key{
			Keys:        []string{"Ctrl+Y"},
			Description: "Show history",
		},
		ClearInput: Key{
			Keys:        []string{"Ctrl+D"},
			Description: "Clear input",
		},
		Paste: Key{
			Keys:        []string{"Ctrl+V"},
			Description: "Paste from clipboard",
		},
	}

	k.SortBar = SortBar{
		ClearInput: Key{
			Keys:        []string{"Ctrl+D"},
			Description: "Clear input",
		},
		Paste: Key{
			Keys:        []string{"Ctrl+V"},
			Description: "Paste from clipboard",
		},
	}

	k.Connection.ToggleFocus = Key{
		Keys:        []string{"Tab", "Backtab"},
		Description: "Toggle focus",
	}

	k.Connection.ConnectionForm = ConnectionFormKeys{
		SaveConnection: Key{
			Keys:        []string{"Ctrl+S", "Enter"},
			Description: "Save connection",
		},
		FocusList: Key{
			Keys:        []string{"Ctrl+H", "Ctrl+Left"},
			Description: "Focus Connection List",
		},
	}

	k.Connection.ConnectionList = ConnectionListKeys{
		FocusForm: Key{
			Keys:        []string{"Ctrl+L", "Ctrl+Right"},
			Description: "Move focus to form",
		},
		DeleteConnection: Key{
			Runes:       []string{"D"},
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

	k.Peeker = PeekerKeys{
		MoveToTop: Key{
			Runes:       []string{"g"},
			Description: "Move to top",
		},
		MoveToBottom: Key{
			Runes:       []string{"G"},
			Description: "Move to bottom",
		},
		CopyHighlight: Key{
			Runes:       []string{"c"},
			Description: "Copy highlighted",
		},
		CopyValue: Key{
			Runes:       []string{"C"},
			Description: "Copy only value",
		},
		ToggleFullScreen: Key{
			Runes:       []string{"F"},
			Description: "Toggle full screen",
		},
		Exit: Key{
			Runes:       []string{"p", "P"},
			Description: "Exit",
		},
	}

	k.History = HistoryKeys{
		ClearHistory: Key{
			Runes:       []string{"C"},
			Description: "Clear history",
		},
		AcceptEntry: Key{
			Keys:        []string{"Enter", "Space"},
			Description: "Accept entry",
		},
		CloseHistory: Key{
			Keys:        []string{"Esc", "Ctrl+Y"},
			Description: "Close history",
		},
	}

	k.Index = IndexKeys{
		ExitAddIndex: Key{
			Keys:        []string{"Esc"},
			Description: "Exit modal",
		},
		AddIndex: Key{
			Runes:       []string{"A"},
			Description: "Add index",
		},
		DeleteIndex: Key{
			Runes:       []string{"D"},
			Description: "Delete index",
		},
	}

	k.AIQuery = AIQuery{
		ExitAIQuery: Key{
			Keys:        []string{"Esc"},
			Description: "Exit AI query",
		},
		ClearPrompt: Key{
			Keys:        []string{"Ctrl+D"},
			Description: "Clear prompt",
		},
	}
}

// LoadKeybindings loads keybindings from the config file
// if the file does not exist it creates a new one with default keybindings
func LoadKeybindings() (*KeyBindings, error) {
	defaultKeybindings := &KeyBindings{}
	defaultKeybindings.loadDefaults()

	if os.Getenv("ENV") == "vi-dev" {
		return defaultKeybindings, nil
	}

	keybindingsPath, err := getKeybindingsPath()
	if err != nil {
		return nil, err
	}

	return util.LoadConfigFile(defaultKeybindings, keybindingsPath)
}

// extractKeysFromStruct extracts all Key structs from a reflect.Value
func extractKeysFromStruct(val reflect.Value) []Key {
	var keys []Key

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Type() == reflect.TypeOf(Key{}) {
			keys = append(keys, field.Interface().(Key))
		} else if field.Kind() == reflect.Struct {
			keys = append(keys, extractKeysFromStruct(field)...)
		}
	}

	return keys
}

// GetAvaliableKeys returns all keys
func (kb KeyBindings) GetAvaliableKeys() []OrderedKeys {
	var keys []OrderedKeys

	v := reflect.ValueOf(kb)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		orderedKeys := OrderedKeys{
			Element: fieldName,
			Keys:    extractKeysFromStruct(field),
		}

		keys = append(keys, orderedKeys)
	}

	return keys
}

// GetKeysForElement returns keys for element
func (kb KeyBindings) GetKeysForElement(elementId string) ([]OrderedKeys, error) {
	if elementId == "" {
		return nil, fmt.Errorf("element is empty")
	}

	v := reflect.ValueOf(kb)
	field := v.FieldByName(elementId)

	if !field.IsValid() || field.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s not found", elementId)
	}

	keys := []OrderedKeys{{
		Element: elementId,
		Keys:    extractKeysFromStruct(field),
	}}

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
	// some hacks for couple of keys
	if namedKey == "Rune[ ]" {
		namedKey = "Space"
	}
	// in some terminals ctrl+H often is seen as backspace
	if namedKey == "Backspace" {
		namedKey = "Ctrl+H"
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
	configDir, err := util.GetConfigDir()
	if err != nil {
		return "", err
	}

	return configDir + "/keybindings.json", nil
}
