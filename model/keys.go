package model

type KeyAction string

const (
	// ConnectorForm
	FormFocusUp   KeyAction = "formFocusUp"
	FormFocusDown KeyAction = "formFocusDown"
)

var (
	MapKeyAction = map[string]KeyAction{
		"formFocusUp":   FormFocusUp,
		"formFocusDown": FormFocusDown,
	}
)

type Keys struct {
	Keys        []string `json:"keys,omitempty"`
	Runes       []string `json:"runes,omitempty"`
	Description string   `json:"description"`
}

type ConnectorForm struct {
	MoveFocusUp    Keys `json:"moveFocusUp"`
	MoveFocusDown  Keys `json:"moveFocusDown"`
	SaveConnection Keys `json:"saveConnection"`
	FocusList      Keys `json:"focusList"`
}

type ConnectorList struct {
	FocusForm        Keys `json:"focusForm"`
	DeleteConnection Keys `json:"deleteConnection"`
	SetConnection    Keys `json:"setConnection"`
}

type KeyBindings struct {
	ConnectorForm ConnectorForm `json:"connectorForm"`
	ConnectorList ConnectorList `json:"connectorList"`
}

type Global struct {
	Quit Keys `json:"quit"`
	Help Keys `json:"help"`
}
