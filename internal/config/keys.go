package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"gopkg.in/yaml.v3"
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
		Global       GlobalKeys       `yaml:"global"`
		Navigation   NavigationKeys   `yaml:"navigation"`
		Help         HelpKeys         `yaml:"help"`
		Connection   ConnectionKeys   `yaml:"connection"`
		Main         MainKeys         `yaml:"main"`
		Databases    DatabasesKeys    `yaml:"databases"`
		FilterBar    FilterBarKeys    `yaml:"filterBar"`
		Content      ContentKeys      `yaml:"content"`
		Peeker       PeekerKeys       `yaml:"peeker"`
		QueryBar     QueryBar         `yaml:"queryBar"`
		SortBar      SortBar          `yaml:"sortBar"`
		Index        IndexKeys        `yaml:"index"`
		IndexAddForm IndexAddFormKeys `yaml:"indexAddForm"`
		AIQuery      AIQueryKeys      `yaml:"aiQuery"`
		History      HistoryKeys      `yaml:"history"`
		Aggregation  AggregationKeys  `yaml:"aggregation"`
	}

	// NavigationKeys holds shared navigation keybindings used across all components
	NavigationKeys struct {
		MoveUp    Key `yaml:"moveUp"`
		MoveDown  Key `yaml:"moveDown"`
		MoveLeft  Key `yaml:"moveLeft"`
		MoveRight Key `yaml:"moveRight"`
	}

	// Key is a lowest level of keybindings
	// It holds the keys and runes that are used to trigger the action
	// and a description of the action that will be displayed in the help
	Key struct {
		Keys        []string `yaml:"keys,omitempty,flow"`
		Runes       []string `yaml:"runes,omitempty,flow"`
		Description string   `yaml:"description,omitempty"`
	}

	// GlobalKeys is a struct that holds the global keybindings
	// for the application, they can be triggered from any view
	// as keys are passed from top to bottom
	GlobalKeys struct {
		CloseApp             Key `yaml:"closeApp"`
		ToggleFullScreenHelp Key `yaml:"toggleFullScreenHelp"`
		OpenConnection       Key `yaml:"openConnection"`
		ShowStyleModal       Key `yaml:"showStyleModal"`
		ToggleHeader         Key `yaml:"toggleHeader"`
	}

	MainKeys struct {
		FocusNext      Key `yaml:"focusNext"`
		FocusPrevious  Key `yaml:"focusPrevious"`
		HideDatabases  Key `yaml:"hideDatabases"`
		ShowAIQuery    Key `yaml:"showAIQuery"`
		ShowServerInfo Key `yaml:"showServerInfo"`
	}

	DatabasesKeys struct {
		FilterBar        Key `yaml:"filterBar"`
		ClearFilter      Key `yaml:"clearFilter"`
		ExpandAll        Key `yaml:"expandAll"`
		CollapseAll      Key `yaml:"collapseAll"`
		AddCollection    Key `yaml:"addCollection"`
		DeleteCollection Key `yaml:"deleteCollection"`
		RenameCollection Key `yaml:"renameCollection"`
	}

	FilterBarKeys struct {
		CloseFilter Key `yaml:"closeFilter"`
		ClearFilter Key `yaml:"clearFilter"`
	}

	ContentKeys struct {
		SwitchView                 Key `yaml:"switchView"`
		PeekDocument               Key `yaml:"peekDocument"`
		FullPagePeek               Key `yaml:"fullPagePeek"`
		AddDocument                Key `yaml:"addDocument"`
		EditDocument               Key `yaml:"editDocument"`
		InlineEdit                 Key `yaml:"inlineEdit"`
		DuplicateDocument          Key `yaml:"duplicateDocument"`
		DuplicateDocumentNoConfirm Key `yaml:"duplicateDocumentNoConfirm"`
		DeleteDocument             Key `yaml:"deleteDocument"`
		DeleteDocumentNoConfirm    Key `yaml:"deleteDocumentNoConfirm"`
		CopyHighlight              Key `yaml:"copyValue"`
		CopyDocument               Key `yaml:"copyDocument"`
		Refresh                    Key `yaml:"refresh"`
		ToggleQueryBar             Key `yaml:"toggleQueryBar"`
		NextDocument               Key `yaml:"nextDocument"`
		PreviousDocument           Key `yaml:"previousDocument"`
		NextPage                   Key `yaml:"nextPage"`
		PreviousPage               Key `yaml:"previousPage"`
		ToggleSortBar              Key `yaml:"toggleSortBar"`
		SortByColumn               Key `yaml:"sortByColumn"`
		HideColumn                 Key `yaml:"hideColumn"`
		ResetHiddenColumns         Key `yaml:"resetHiddenColumns"`
		ToggleQueryOptions         Key `yaml:"toggleQueryOptions"`
		MultipleSelect             Key `yaml:"multipleSelect"`
		ClearSelection             Key `yaml:"clearSelection"`
	}

	QueryBar struct {
		ShowHistory Key `yaml:"showHistory"`
		ClearInput  Key `yaml:"clearInput"`
		Paste       Key `yaml:"paste"`
		NextMarker  Key `yaml:"nextMarker"`
	}

	SortBar struct {
		ClearInput Key `yaml:"clearInput"`
		Paste      Key `yaml:"paste"`
	}

	ConnectionKeys struct {
		ConnectionForm ConnectionFormKeys `yaml:"connectionForm"`
		ConnectionList ConnectionListKeys `yaml:"connectionList"`
	}

	ConnectionFormKeys struct {
		SaveConnection Key `yaml:"saveConnection"`
		FocusList      Key `yaml:"focusList"`
	}

	ConnectionListKeys struct {
		FocusForm        Key `yaml:"focusForm"`
		DeleteConnection Key `yaml:"deleteConnection"`
		EditConnection   Key `yaml:"editConnection"`
		SetConnection    Key `yaml:"setConnection"`
	}

	HelpKeys struct {
		Close   Key `yaml:"close"`
		Search  Key `yaml:"search"`
		EditKey Key `yaml:"editKey"`
	}

	PeekerKeys struct {
		MoveToTop        Key `yaml:"moveToTop"`
		MoveToBottom     Key `yaml:"moveToBottom"`
		CopyHighlight    Key `yaml:"copyHighlight"`
		CopyValue        Key `yaml:"copyValue"`
		ToggleFullScreen Key `yaml:"toggleFullScreen"`
		Exit             Key `yaml:"exit"`
	}

	HistoryKeys struct {
		ClearHistory Key `yaml:"clearHistory"`
		AcceptEntry  Key `yaml:"acceptEntry"`
		CloseHistory Key `yaml:"closeHistory"`
	}

	IndexKeys struct {
		AddIndex    Key `yaml:"addIndex"`
		DeleteIndex Key `yaml:"deleteIndex"`
	}

	IndexAddFormKeys struct {
		ExitForm    Key `yaml:"exitForm"`
		AddColumn   Key `yaml:"addColumn"`
		CreateIndex Key `yaml:"createIndex"`
	}

	AIQueryKeys struct {
		ExitAIQuery Key `yaml:"exitAIQuery"`
		ClearPrompt Key `yaml:"clearPrompt"`
	}

	AggregationKeys struct {
		Stages  AggregationStageKeys  `yaml:"stages"`
		Results AggregationResultKeys `yaml:"results"`
	}

	AggregationStageKeys struct {
		AddStage      Key `yaml:"addStage"`
		EditStage     Key `yaml:"editStage"`
		DeleteStage   Key `yaml:"deleteStage"`
		RunPipeline   Key `yaml:"runPipeline"`
		ClearPipeline Key `yaml:"clearPipeline"`
		MoveStageDown Key `yaml:"moveStageDown"`
		MoveStageUp   Key `yaml:"moveStageUp"`
		FocusResults  Key `yaml:"focusResults"`
	}

	AggregationResultKeys struct {
		FocusStages   Key `yaml:"focusStages"`
		ChangeView    Key `yaml:"changeView"`
		PeekDocument  Key `yaml:"peekDocument"`
		FullPagePeek  Key `yaml:"fullPagePeek"`
		CopyHighlight Key `yaml:"copyHighlight"`
		CopyDocument  Key `yaml:"copyDocument"`
	}
)

func (k *KeyBindings) loadDefaults() {
	k.Global = GlobalKeys{
		CloseApp: Key{
			Keys:        []string{"Ctrl+c"},
			Runes:       []string{"q"},
			Description: "Close application",
		},
		ToggleFullScreenHelp: Key{
			Runes:       []string{"?"},
			Description: "Toggle full screen help",
		},
		OpenConnection: Key{
			Keys:        []string{"Ctrl+o"},
			Description: "Open connection page",
		},
		ShowStyleModal: Key{
			Keys:        []string{"Ctrl+t"},
			Description: "Toggle style change modal",
		},
		ToggleHeader: Key{
			Runes:       []string{"t"},
			Description: "Expand/collapse header",
		},
	}

	k.Main = MainKeys{
		FocusNext: Key{
			Keys:        []string{"Ctrl+l", "Tab"},
			Description: "Focus next component",
		},
		FocusPrevious: Key{
			Keys:        []string{"Ctrl+h", "Backtab"},
			Description: "Focus prev component",
		},
		HideDatabases: Key{
			Keys:        []string{"Ctrl+n"},
			Description: "Hide databases",
		},
		ShowServerInfo: Key{
			Keys:        []string{"Alt+s"},
			Description: "Show server info",
		},
		ShowAIQuery: Key{
			Keys:        []string{"Alt+a"},
			Description: "Show AI prompt",
		},
	}

	k.Navigation = NavigationKeys{
		MoveUp: Key{
			Keys:        []string{"Up"},
			Runes:       []string{"k"},
			Description: "Move up",
		},
		MoveDown: Key{
			Keys:        []string{"Down"},
			Runes:       []string{"j"},
			Description: "Move down",
		},
		MoveLeft: Key{
			Keys:        []string{"Left"},
			Runes:       []string{"h"},
			Description: "Move left",
		},
		MoveRight: Key{
			Keys:        []string{"Right"},
			Runes:       []string{"l"},
			Description: "Move right",
		},
	}

	k.Databases = DatabasesKeys{
		FilterBar: Key{
			Runes:       []string{"/"},
			Description: "Focus filter bar",
		},
		ClearFilter: Key{
			Keys:        []string{"Ctrl+u"},
			Description: "Clear filter",
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
			Keys:        []string{"Ctrl+d"},
			Description: "Delete collection",
		},
		RenameCollection: Key{
			Runes:       []string{"R"},
			Description: "Rename collection",
		},
	}

	k.FilterBar = FilterBarKeys{
		CloseFilter: Key{
			Keys:        []string{"Escape"},
			Description: "Close filter bar",
		},
		ClearFilter: Key{
			Keys:        []string{"Ctrl+u"},
			Description: "Clear filter",
		},
	}

	k.Content = ContentKeys{
		SwitchView: Key{
			Runes:       []string{"v"},
			Description: "Change view",
		},
		PeekDocument: Key{
			Runes:       []string{"o"},
			Keys:        []string{"Enter"},
			Description: "Open peek",
		},
		FullPagePeek: Key{
			Runes:       []string{"O"},
			Description: "Full peek",
		},
		AddDocument: Key{
			Runes:       []string{"A"},
			Description: "Add new",
		},
		EditDocument: Key{
			Runes:       []string{"E"},
			Description: "Full Edit",
		},
		InlineEdit: Key{
			Runes:       []string{"e"},
			Description: "Inline edit",
		},
		DuplicateDocument: Key{
			Runes:       []string{"D"},
			Description: "Duplicate",
		},
		DuplicateDocumentNoConfirm: Key{
			Keys:        []string{"Alt+D"},
			Description: "Duplicate no confirm",
		},
		DeleteDocument: Key{
			Keys:        []string{"Ctrl+d"},
			Description: "Delete",
		},
		DeleteDocumentNoConfirm: Key{
			Keys:        []string{"Alt+d"},
			Description: "Delete no confirm",
		},
		MultipleSelect: Key{
			Runes:       []string{"V"},
			Description: "Multiple select",
		},
		ClearSelection: Key{
			Keys:        []string{"Esc"},
			Description: "Clear selection",
		},
		CopyHighlight: Key{
			Runes:       []string{"c"},
			Description: "Copy highlighted",
		},
		CopyDocument: Key{
			Runes:       []string{"C"},
			Description: "Copy document",
		},
		Refresh: Key{
			Keys:        []string{"Ctrl+r"},
			Description: "Refresh",
		},
		ToggleQueryBar: Key{
			Runes:       []string{"/"},
			Description: "Query bar",
		},
		ToggleSortBar: Key{
			Runes:       []string{"s"},
			Description: "Sort bar",
		},
		SortByColumn: Key{
			Runes:       []string{"S"},
			Description: "Sort by column",
		},
		HideColumn: Key{
			Runes:       []string{"H"},
			Description: "Hide column",
		},
		ResetHiddenColumns: Key{
			Runes:       []string{"r"},
			Description: "Reset hidden columns",
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
		ToggleQueryOptions: Key{
			Keys:        []string{"Alt+o"},
			Description: "Toggle query options",
		},
	}

	k.QueryBar = QueryBar{
		ShowHistory: Key{
			Keys:        []string{"Ctrl+y"},
			Description: "Show history",
		},
		ClearInput: Key{
			Keys:        []string{"Ctrl+u"},
			Description: "Clear input",
		},
		Paste: Key{
			Keys:        []string{"Ctrl+v"},
			Description: "Paste from clipboard",
		},
		NextMarker: Key{
			Keys:        []string{"Ctrl+l"},
			Description: "Next marker",
		},
	}

	k.SortBar = SortBar{
		ClearInput: Key{
			Keys:        []string{"Ctrl+u"},
			Description: "Clear input",
		},
		Paste: Key{
			Keys:        []string{"Ctrl+v"},
			Description: "Paste from clipboard",
		},
	}
	k.Connection.ConnectionForm = ConnectionFormKeys{
		SaveConnection: Key{
			Keys:        []string{"Ctrl+s"},
			Description: "Save connection",
		},
		FocusList: Key{
			Keys:        []string{"Ctrl+h", "Ctrl+Left"},
			Description: "Focus Connection List",
		},
	}

	k.Connection.ConnectionList = ConnectionListKeys{
		FocusForm: Key{
			Keys:        []string{"Ctrl+l", "Ctrl+Right"},
			Description: "Focus form",
		},
		DeleteConnection: Key{
			Keys:        []string{"Ctrl+d"},
			Description: "Delete connection",
		},
		EditConnection: Key{
			Runes:       []string{"E"},
			Description: "Edit connection",
		},
		SetConnection: Key{
			Keys:        []string{"Enter", "Space"},
			Description: "Set connection",
		},
	}

	k.Help = HelpKeys{
		Close: Key{
			Keys:        []string{"Esc"},
			Description: "Close help",
		},
		Search: Key{
			Runes:       []string{"/"},
			Description: "Search keybindings",
		},
		EditKey: Key{
			Runes:       []string{"e"},
			Description: "Edit keybinding",
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
			Runes:       []string{"o", "O"},
			Description: "Exit",
		},
	}

	k.History = HistoryKeys{
		ClearHistory: Key{
			Keys:        []string{"Ctrl+d"},
			Description: "Clear history",
		},
		AcceptEntry: Key{
			Keys:        []string{"Enter", "Space"},
			Description: "Accept entry",
		},
		CloseHistory: Key{
			Keys:        []string{"Esc", "Ctrl+y"},
			Description: "Close history",
		},
	}

	k.Index = IndexKeys{
		AddIndex: Key{
			Runes:       []string{"A"},
			Description: "Add index",
		},
		DeleteIndex: Key{
			Keys:        []string{"Ctrl+d"},
			Description: "Delete index",
		},
	}

	k.IndexAddForm = IndexAddFormKeys{
		ExitForm: Key{
			Keys:        []string{"Esc"},
			Description: "Exit form",
		},
		AddColumn: Key{
			Keys:        []string{"Ctrl+a"},
			Description: "Add column",
		},
		CreateIndex: Key{
			Keys:        []string{"Ctrl+s"},
			Description: "Create index",
		},
	}

	k.AIQuery = AIQueryKeys{
		ExitAIQuery: Key{
			Keys:        []string{"Esc"},
			Description: "Exit AI query",
		},
		ClearPrompt: Key{
			Keys:        []string{"Ctrl+u"},
			Description: "Clear prompt",
		},
	}

	k.Aggregation = AggregationKeys{
		Stages: AggregationStageKeys{
			AddStage: Key{
				Runes:       []string{"a"},
				Description: "Add new stage",
			},
			EditStage: Key{
				Runes:       []string{"e"},
				Description: "Edit selected stage",
			},
			DeleteStage: Key{
				Keys:        []string{"Ctrl+d"},
				Description: "Delete selected stage",
			},
			RunPipeline: Key{
				Runes:       []string{"R"},
				Description: "Run pipeline",
			},
			ClearPipeline: Key{
				Runes:       []string{"C"},
				Description: "Clear all stages",
			},
			MoveStageDown: Key{
				Runes:       []string{"J"},
				Description: "Move stage down",
			},
			MoveStageUp: Key{
				Runes:       []string{"K"},
				Description: "Move stage up",
			},
			FocusResults: Key{
				Keys:        []string{"Ctrl+j"},
				Description: "Focus results",
			},
		},
		Results: AggregationResultKeys{
			FocusStages: Key{
				Keys:        []string{"Ctrl+j"},
				Description: "Focus stages",
			},
			ChangeView: Key{
				Runes:       []string{"v"},
				Description: "Switch view",
			},
			PeekDocument: Key{
				Runes:       []string{"o"},
				Keys:        []string{"Enter"},
				Description: "Peek document",
			},
			FullPagePeek: Key{
				Runes:       []string{"O"},
				Description: "Full page peek",
			},
			CopyHighlight: Key{
				Runes:       []string{"c"},
				Description: "Copy cell",
			},
			CopyDocument: Key{
				Runes:       []string{"C"},
				Description: "Copy document",
			},
		},
	}
}

const keybindingsFileHeader = `# runes: literal characters, case-sensitive (e.g. [a], [A])
# keys:  named/combo keys (e.g. [Enter], [Escape], [Tab], [Space])
#        Ctrl+<letter>: case-insensitive in config, but no Ctrl+Shift — in config Ctrl+L is the same as Ctrl+l
#        Alt+<char>:    case-sensitive, both upper and lower work (e.g. Alt+a, Alt+A)

`

// LoadKeybindings loads keybindings from the config file.
// If the file does not exist it creates a new one with default keybindings.
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

	if _, err := os.Stat(keybindingsPath); os.IsNotExist(err) {
		if err := writeKeybindingsWithHeader(defaultKeybindings, keybindingsPath); err != nil {
			return nil, err
		}
		return defaultKeybindings, nil
	}

	return util.LoadConfigFile(defaultKeybindings, keybindingsPath)
}

func writeKeybindingsWithHeader(kb *KeyBindings, path string) error {
	data, err := yaml.Marshal(kb)
	if err != nil {
		return fmt.Errorf("failed to marshal keybindings: %w", err)
	}
	content := append([]byte(keybindingsFileHeader), data...)
	return os.WriteFile(path, content, FileMode)
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

// GetKeysForElement returns keys for element.
// elementId supports dot-notation to reach nested structs, e.g. "Aggregation.Stages".
func (kb KeyBindings) GetKeysForElement(elementId string) ([]OrderedKeys, error) {
	if elementId == "" {
		return nil, fmt.Errorf("element is empty")
	}

	v := reflect.ValueOf(kb)
	parts := strings.SplitN(elementId, ".", 2)
	field := v.FieldByName(parts[0])

	if !field.IsValid() || field.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s not found", elementId)
	}

	if len(parts) == 2 {
		field = field.FieldByName(parts[1])
		if !field.IsValid() || field.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %s not found", elementId)
		}
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
	// Normalize Ctrl+letter to uppercase since tcell always reports uppercase,
	// allowing config to use lowercase (e.g. "Ctrl+l") for user clarity
	if strings.HasPrefix(namedKey, "Ctrl+") && len(namedKey) == 6 {
		namedKey = "Ctrl+" + strings.ToUpper(string(namedKey[5]))
	}
	// Handle Alt+Rune combinations
	if strings.HasPrefix(namedKey, "Alt+Rune[") && len(namedKey) >= 10 {
		runeChar := namedKey[9:10]
		altCombo := "Alt+" + runeChar

		for _, k := range configKey.Keys {
			if k == altCombo {
				return true
			}
		}
		return false
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
		// Normalize Ctrl+letter to uppercase to match tcell's key naming
		if strings.HasPrefix(k, "Ctrl+") && len(k) == 6 {
			k = "Ctrl+" + strings.ToUpper(string(k[5]))
		}
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

	return configDir + "/keybindings.yaml", nil
}

// SaveKeybindings writes the current keybindings to the config file.
func (kb *KeyBindings) SaveKeybindings() error {
	path, err := getKeybindingsPath()
	if err != nil {
		return err
	}
	return writeKeybindingsWithHeader(kb, path)
}

// SetKeyAt updates the key at keyIndex within the named element's Key fields.
// The index matches the order returned by GetAvaliableKeys / extractKeysFromStruct.
func (kb *KeyBindings) SetKeyAt(element string, keyIndex int, newKey Key) error {
	v := reflect.ValueOf(kb).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Name == element {
			idx := 0
			if setKeyAtIndex(v.Field(i), keyIndex, &idx, newKey) {
				return nil
			}
			return fmt.Errorf("key index %d not found in element %s", keyIndex, element)
		}
	}
	return fmt.Errorf("element %s not found", element)
}

func setKeyAtIndex(val reflect.Value, targetIdx int, currentIdx *int, newKey Key) bool {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Type() == reflect.TypeOf(Key{}) {
			if *currentIdx == targetIdx {
				field.Set(reflect.ValueOf(newKey))
				return true
			}
			*currentIdx++
		} else if field.Kind() == reflect.Struct {
			if setKeyAtIndex(field, targetIdx, currentIdx, newKey) {
				return true
			}
		}
	}
	return false
}
