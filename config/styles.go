package config

import (
	"os"
	"regexp"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Styles is a struct that contains all the styles for the application
type (
	Style string

	Styles struct {
		Root      RootStyle      `yaml:"main"`
		Connector ConnectorStyle `yaml:"connector"`
		Header    HeaderStyle    `yaml:"header"`
		Sidebar   SidebarStyle   `yaml:"sidebar"`
		Content   ContentStyle   `yaml:"content"`
		DocPeeker DocPeekerStyle `yaml:"docPeeker"`
		InputBar  InputBarStyle  `yaml:"filterBar"`
		History   HistoryStyle   `yaml:"history"`
		Help      HelpStyle      `yaml:"help"`
		Others    OthersStyle    `yaml:"others"`
	}

	// RootStyle is a struct that contains all the root styles for the application
	RootStyle struct {
		BackgroundColor    Style `yaml:"backgroundColor"`
		TextColor          Style `yaml:"textColor"`
		SecondaryTextColor Style `yaml:"secondaryTextColor"`
		BorderColor        Style `yaml:"borderColor"`
		FocusColor         Style `yaml:"focusColor"`
		TitleColor         Style `yaml:"titleColor"`
		GraphicsColor      Style `yaml:"graphicsColor"`
	}

	// ConnectorStyle is a struct that contains all the styles for the connector
	ConnectorStyle struct {
		BackgroundColor              Style `yaml:"backgroundColor"`
		BorderColor                  Style `yaml:"borderColor"`
		TitleColor                   Style `yaml:"titleColor"`
		FormLabelColor               Style `yaml:"formLabelColor"`
		FormInputBackgroundColor     Style `yaml:"formInputBackgroundColor"`
		FormInputColor               Style `yaml:"formInputColor"`
		FormButtonColor              Style `yaml:"formButtonColor"`
		ListTextColor                Style `yaml:"listTextColor"`
		ListSelectedTextColor        Style `yaml:"listSelectedTextColor"`
		ListSelectedBackgroundColor  Style `yaml:"listSelectedBackgroundColor"`
		ListSecondaryTextColor       Style `yaml:"listSecondaryTextColor"`
		ListSecondaryBackgroundColor Style `yaml:"listSecondaryBackgroundColor"`
	}

	// HeaderStyle is a struct that contains all the styles for the header
	HeaderStyle struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		KeyColor        Style `yaml:"keyColor"`
		ValueColor      Style `yaml:"valueColor"`
		ActiveSymbol    Style `yaml:"activeSymbol"`
		InactiveSymbol  Style `yaml:"inactiveSymbol"`
	}

	// SidebarStyle is a struct that contains all the styles for the sidebar
	SidebarStyle struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		NodeColor       Style `yaml:"nodeColor"`
		NodeSymbol      Style `yaml:"nodeSymbol"`
		LeafColor       Style `yaml:"leafColor"`
		LeafSymbol      Style `yaml:"leafSymbol"`
		BranchColor     Style `yaml:"branchColor"`
	}

	// ContentStyle is a struct that contains all the styles for the content
	ContentStyle struct {
		BackgroundColor  Style `yaml:"backgroundColor"`
		BorderColor      Style `yaml:"borderColor"`
		TitleColor       Style `yaml:"titleColor"`
		CellTextColor    Style `yaml:"cellTextColor"`
		ActiveRowColor   Style `yaml:"activeRowColor"`
		SelectedRowColor Style `yaml:"selectedRowColor"`
	}

	// DocPeekerStyle is a struct that contains all the styles for the json peeker
	DocPeekerStyle struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		KeyColor        Style `yaml:"keyColor"`
		ValueColor      Style `yaml:"valueColor"`
		BracketColor    Style `yaml:"bracketColor"`
	}

	// InputBarStyle is a struct that contains all the styles for the filter bar
	InputBarStyle struct {
		BackgroundColor Style             `yaml:"backgroundColor"`
		BorderColor     Style             `yaml:"borderColor"`
		TitleColor      Style             `yaml:"titleColor"`
		LabelColor      Style             `yaml:"labelColor"`
		InputColor      Style             `yaml:"inputColor"`
		Autocomplete    AutocompleteStyle `yaml:"autocomplete"`
	}

	AutocompleteStyle struct {
		BackgroundColor       Style `yaml:"backgroundColor"`
		BorderColor           Style `yaml:"borderColor"`
		TextColor             Style `yaml:"textColor"`
		ActiveBackgroundColor Style `yaml:"activeBackgroundColor"`
		ActiveTextColor       Style `yaml:"activeTextColor"`
		SecondaryTextColor    Style `yaml:"secondaryTextColor"`
	}

	HistoryStyle struct {
		BackgroundColor         Style `yaml:"backgroundColor"`
		TextColor               Style `yaml:"textColor"`
		SelectedTextColor       Style `yaml:"selectedTextColor"`
		SelectedBackgroundColor Style `yaml:"selectedBackgroundColor"`
	}

	HelpStyle struct {
		BackgroundColor  Style `yaml:"backgroundColor"`
		BorderColor      Style `yaml:"borderColor"`
		TitleColor       Style `yaml:"titleColor"`
		KeyColor         Style `yaml:"keyColor"`
		DescriptionColor Style `yaml:"descriptionColor"`
	}

	OthersStyle struct {
		ButtonsTextColor       Style `yaml:"buttonsTextColor"`
		ButtonsBackgroundColor Style `yaml:"buttonsBackgroundColor"`
	}
)

// NewStyles creates a new Styles struct with default values
func NewStyles() *Styles {
	styles := &Styles{}

	customStyles, err := styles.LoadCustomConfig()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to load custom styles, loading default styles")
		styles.loadDefaultStyles()
	} else {
		log.Debug().Msg("Loaded custom styles")
		styles = customStyles
	}

	styles.loadMainStyles()

	return styles
}

func (s *Styles) LoadCustomConfig() (*Styles, error) {
	bytes, err := os.ReadFile("styles.yaml")
	if err != nil {
		return nil, err
	}

	styles := &Styles{}
	err = yaml.Unmarshal(bytes, styles)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Loaded styles from styles.yaml")

	return styles, nil
}

func (s *Styles) loadMainStyles() {
	tview.Styles.PrimitiveBackgroundColor = s.loadColor(s.Root.BackgroundColor)
	tview.Styles.ContrastBackgroundColor = s.loadColor(s.Root.BackgroundColor)
	tview.Styles.MoreContrastBackgroundColor = s.loadColor(s.Root.BackgroundColor)
	tview.Styles.PrimaryTextColor = s.loadColor(s.Root.TextColor)
	tview.Styles.SecondaryTextColor = s.loadColor(s.Root.SecondaryTextColor)
	tview.Styles.TertiaryTextColor = s.loadColor(s.Root.SecondaryTextColor)
	tview.Styles.InverseTextColor = s.loadColor(s.Root.SecondaryTextColor)
	tview.Styles.ContrastSecondaryTextColor = s.loadColor(s.Root.SecondaryTextColor)
	tview.Styles.BorderColor = s.loadColor(s.Root.BorderColor)
	tview.Styles.FocusColor = s.loadColor(s.Root.FocusColor)
	tview.Styles.TitleColor = s.loadColor(s.Root.TitleColor)
	tview.Styles.GraphicsColor = s.loadColor(s.Root.GraphicsColor)
}

func (s *Styles) loadDefaultStyles() {
	s.Root = RootStyle{
		BackgroundColor:    "#0F172A",
		TextColor:          "#FFFFFF",
		SecondaryTextColor: "#F1FA8C",
		BorderColor:        "#387D44",
		FocusColor:         "#50FA7B",
		TitleColor:         "#387D44",
		GraphicsColor:      "#387D44",
	}

	s.Connector = ConnectorStyle{
		BackgroundColor:             "#0F172A",
		BorderColor:                 "#387D44",
		TitleColor:                  "#F1FA8C",
		FormLabelColor:              "#F1FA8C",
		FormInputBackgroundColor:    "#163694",
		FormInputColor:              "#F1FA8C",
		FormButtonColor:             "#387D44",
		ListTextColor:               "#F1FA8C",
		ListSelectedTextColor:       "#50FA7B",
		ListSelectedBackgroundColor: "#163694",
		ListSecondaryTextColor:      "#50FA7B",
	}

	s.Header = HeaderStyle{
		BackgroundColor: "#0F172A",
		BorderColor:     "#387D44",
		KeyColor:        "#F1FA8C",
		ValueColor:      "#387D44",
		ActiveSymbol:    "●",
		InactiveSymbol:  "○",
	}

	s.Sidebar = SidebarStyle{
		BackgroundColor: "#0F172A",
		BorderColor:     "#387D44",
		NodeColor:       "#387D44",
		LeafColor:       "#163694",
		BranchColor:     "#387D44",
		NodeSymbol:      "📁",
		LeafSymbol:      "📄",
	}

	s.Content = ContentStyle{
		BackgroundColor:  "#0F172A",
		BorderColor:      "#387D44",
		TitleColor:       "#163694",
		CellTextColor:    "#387D44",
		ActiveRowColor:   "#50FA7B",
		SelectedRowColor: "#50FA7B",
	}

	s.DocPeeker = DocPeekerStyle{
		BackgroundColor: "#0F172A",
		BorderColor:     "#50FA7B",
		TitleColor:      "#F1FA8C",
		KeyColor:        "#F1FA8C",
		ValueColor:      "#FFFFFF",
		BracketColor:    "#FFFFFF",
	}

	s.InputBar = InputBarStyle{
		BackgroundColor: "#0F172A",
		BorderColor:     "#50FA7B",
		LabelColor:      "#F1FA8C",
		InputColor:      "#FFFFFF",
		Autocomplete: AutocompleteStyle{
			BackgroundColor:       "#163694",
			TextColor:             "#F1FA8C",
			BorderColor:           "#50FA7B",
			ActiveBackgroundColor: "#A4A4A4",
			ActiveTextColor:       "#FFFFFF",
			SecondaryTextColor:    "#50D78E",
		},
	}

	s.History = HistoryStyle{
		BackgroundColor:         "#0F172A",
		TextColor:               "#F1FA8C",
		SelectedTextColor:       "#50FA7B",
		SelectedBackgroundColor: "#163694",
	}

	s.Help = HelpStyle{
		BackgroundColor:  "#0F172A",
		BorderColor:      "#50FA7B",
		TitleColor:       "#163694",
		KeyColor:         "#F1FA8C",
		DescriptionColor: "#FFFFFF",
	}

	s.Others = OthersStyle{
		ButtonsTextColor:       "#F1FA8C",
		ButtonsBackgroundColor: "#0F172A",
	}
}

// LoadColor loads a color from a string
// It will check if the color is a hex color or a color name
func (s *Styles) loadColor(color Style) tcell.Color {
	strColor := string(color)
	if isHexColor(strColor) {
		intColor, _ := strconv.ParseInt(strColor[1:], 16, 32)
		return tcell.NewHexColor(int32(intColor))
	}

	c := tcell.GetColor(strColor)
	return c
}

// Color returns the tcell.Color of the style
func (s *Style) Color() tcell.Color {
	return tcell.GetColor(string(*s))
}

// String returns the string value of the style
func (s *Style) String() string {
	return string(*s)
}

func isHexColor(s string) bool {
	re := regexp.MustCompile("^#(?:[0-9a-fA-F]{3}){1,2}$")
	return re.MatchString(s)
}
