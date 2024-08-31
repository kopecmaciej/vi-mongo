package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"gopkg.in/yaml.v3"
)

// Styles is a struct that contains all the styles for the application
type (
	Style string

	Styles struct {
		Global    GlobalStyles   `yaml:"global"`
		Welcome   WelcomeStyle   `yaml:"welcome"`
		Connector ConnectorStyle `yaml:"connector"`
		Header    HeaderStyle    `yaml:"header"`
		Databases DatabasesStyle `yaml:"databases"`
		Content   ContentStyle   `yaml:"content"`
		DocPeeker DocPeekerStyle `yaml:"docPeeker"`
		InputBar  InputBarStyle  `yaml:"filterBar"`
		History   HistoryStyle   `yaml:"history"`
		Help      HelpStyle      `yaml:"help"`
		Others    OthersStyle    `yaml:"others"`
	}

	// GlobalStyles is a struct that contains all the root styles for the application
	GlobalStyles struct {
		BackgroundColor    Style `yaml:"backgroundColor"`
		TextColor          Style `yaml:"textColor"`
		SecondaryTextColor Style `yaml:"secondaryTextColor"`
		BorderColor        Style `yaml:"borderColor"`
		FocusColor         Style `yaml:"focusColor"`
		TitleColor         Style `yaml:"titleColor"`
		GraphicsColor      Style `yaml:"graphicsColor"`
	}

	// WelcomeStyle is a struct that contains all the styles for the welcome screen
	WelcomeStyle struct {
		TextColor                Style `yaml:"textColor"`
		FormLabelColor           Style `yaml:"formLabelColor"`
		FormInputColor           Style `yaml:"formInputColor"`
		FormInputBackgroundColor Style `yaml:"formInputBackgroundColor"`
	}

	// ConnectorStyle is a struct that contains all the styles for the connector
	ConnectorStyle struct {
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
		KeyColor       Style `yaml:"keyColor"`
		ValueColor     Style `yaml:"valueColor"`
		ActiveSymbol   Style `yaml:"activeSymbol"`
		InactiveSymbol Style `yaml:"inactiveSymbol"`
	}

	// DatabasesStyle is a struct that contains all the styles for the databases
	DatabasesStyle struct {
		NodeColor   Style `yaml:"nodeColor"`
		NodeSymbol  Style `yaml:"nodeSymbol"`
		LeafColor   Style `yaml:"leafColor"`
		LeafSymbol  Style `yaml:"leafSymbol"`
		BranchColor Style `yaml:"branchColor"`
	}

	// ContentStyle is a struct that contains all the styles for the content
	ContentStyle struct {
		StatusTextColor          Style `yaml:"docInfoTextColor"`
		HeaderRowBackgroundColor Style `yaml:"headerRowColor"`
		ColumnKeyColor           Style `yaml:"columnKeyColor"`
		ColumnTypeColor          Style `yaml:"columnTypeColor"`
		CellTextColor            Style `yaml:"cellTextColor"`
		ActiveRowColor           Style `yaml:"activeRowColor"`
		SelectedRowColor         Style `yaml:"selectedRowColor"`
		SeparatorSymbol          Style `yaml:"separatorSymbol"`
		SeparatorColor           Style `yaml:"separatorColor"`
	}

	// DocPeekerStyle is a struct that contains all the styles for the json peeker
	DocPeekerStyle struct {
		KeyColor       Style `yaml:"keyColor"`
		ValueColor     Style `yaml:"valueColor"`
		BracketColor   Style `yaml:"bracketColor"`
		ArrayColor     Style `yaml:"arrayColor"`
		HighlightColor Style `yaml:"highlightColor"`
	}

	// InputBarStyle is a struct that contains all the styles for the filter bar
	InputBarStyle struct {
		LabelColor   Style             `yaml:"labelColor"`
		InputColor   Style             `yaml:"inputColor"`
		Autocomplete AutocompleteStyle `yaml:"autocomplete"`
	}

	AutocompleteStyle struct {
		BackgroundColor       Style `yaml:"backgroundColor"`
		TextColor             Style `yaml:"textColor"`
		ActiveBackgroundColor Style `yaml:"activeBackgroundColor"`
		ActiveTextColor       Style `yaml:"activeTextColor"`
		SecondaryTextColor    Style `yaml:"secondaryTextColor"`
	}

	HistoryStyle struct {
		TextColor               Style `yaml:"textColor"`
		SelectedTextColor       Style `yaml:"selectedTextColor"`
		SelectedBackgroundColor Style `yaml:"selectedBackgroundColor"`
	}

	HelpStyle struct {
		HeaderColor      Style `yaml:"headerColor"`
		KeyColor         Style `yaml:"keyColor"`
		DescriptionColor Style `yaml:"descriptionColor"`
	}

	OthersStyle struct {
		ButtonsTextColor       Style `yaml:"buttonsTextColor"`
		ButtonsBackgroundColor Style `yaml:"buttonsBackgroundColor"`
	}
)

func (s *Styles) loadDefaultStyles() {
	s.Global = GlobalStyles{
		BackgroundColor:    "#0F172A",
		TextColor:          "#FFFFFF",
		SecondaryTextColor: "#F1FA8C",
		BorderColor:        "#387D44",
		FocusColor:         "#50FA7B",
		TitleColor:         "#387D44",
		GraphicsColor:      "#387D44",
	}

	s.Welcome = WelcomeStyle{
		TextColor:                "#F1FA8C",
		FormLabelColor:           "#F1FA8C",
		FormInputColor:           "#F1FA8C",
		FormInputBackgroundColor: "#163694",
	}

	s.Connector = ConnectorStyle{
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
		KeyColor:       "#F1FA8C",
		ValueColor:     "#387D44",
		ActiveSymbol:   "●",
		InactiveSymbol: "○",
	}

	s.Databases = DatabasesStyle{
		NodeColor:   "#387D44",
		LeafColor:   "#4368da",
		BranchColor: "#44bb58",
		NodeSymbol:  "📁",
		LeafSymbol:  "📄",
	}

	s.Content = ContentStyle{
		StatusTextColor:          "#F1FA8C",
		HeaderRowBackgroundColor: "#163694",
		ColumnKeyColor:           "#F1FA8C",
		ColumnTypeColor:          "#689e76",
		CellTextColor:            "#387D44",
		ActiveRowColor:           "#50FA7B",
		SelectedRowColor:         "#50FA7B",
		SeparatorSymbol:          "|",
		SeparatorColor:           "#6c6e6d",
	}

	s.DocPeeker = DocPeekerStyle{
		KeyColor:       "#387D44",
		ValueColor:     "#FFFFFF",
		ArrayColor:     "#387D44",
		HighlightColor: "#163694",
		BracketColor:   "#FF5555",
	}

	s.InputBar = InputBarStyle{
		LabelColor: "#F1FA8C",
		InputColor: "#FFFFFF",
		Autocomplete: AutocompleteStyle{
			BackgroundColor:       "#163694",
			TextColor:             "#F1FA8C",
			ActiveBackgroundColor: "#A4A4A4",
			ActiveTextColor:       "#FFFFFF",
			SecondaryTextColor:    "#50D78E",
		},
	}

	s.History = HistoryStyle{
		TextColor:               "#F1FA8C",
		SelectedTextColor:       "#50FA7B",
		SelectedBackgroundColor: "#163694",
	}

	s.Help = HelpStyle{
		HeaderColor:      "#F1FA8C",
		KeyColor:         "#F1FA8C",
		DescriptionColor: "#FFFFFF",
	}

	s.Others = OthersStyle{
		ButtonsTextColor:       "#F1FA8C",
		ButtonsBackgroundColor: "#0F172A",
	}
}

// LoadStyles creates a new Styles struct with default values
func LoadStyles() (*Styles, error) {
	styles := &Styles{}
	defaultStyles := &Styles{}
	defaultStyles.loadDefaultStyles()

	stylePath, err := getStylePath()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(stylePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Use default styles if file doesn't exist
			return defaultStyles, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(bytes, styles)
	if err != nil {
		return nil, err
	}

	// Merge loaded styles with default styles
	util.MergeConfigs(styles, defaultStyles)

	return styles, nil
}

func (s *Styles) LoadMainStyles() {
	tview.Styles.PrimitiveBackgroundColor = s.loadColor(s.Global.BackgroundColor)
	tview.Styles.ContrastBackgroundColor = s.loadColor(s.Global.BackgroundColor)
	tview.Styles.MoreContrastBackgroundColor = s.loadColor(s.Global.BackgroundColor)
	tview.Styles.PrimaryTextColor = s.loadColor(s.Global.TextColor)
	tview.Styles.SecondaryTextColor = s.loadColor(s.Global.SecondaryTextColor)
	tview.Styles.TertiaryTextColor = s.loadColor(s.Global.SecondaryTextColor)
	tview.Styles.InverseTextColor = s.loadColor(s.Global.SecondaryTextColor)
	tview.Styles.ContrastSecondaryTextColor = s.loadColor(s.Global.SecondaryTextColor)
	tview.Styles.BorderColor = s.loadColor(s.Global.BorderColor)
	tview.Styles.FocusColor = s.loadColor(s.Global.FocusColor)
	tview.Styles.TitleColor = s.loadColor(s.Global.TitleColor)
	tview.Styles.GraphicsColor = s.loadColor(s.Global.GraphicsColor)
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

// SetColor sets the color of the style
func (s *Style) GetWithColor(color tcell.Color) string {
	return fmt.Sprintf("[%s]%s[%s]", color.String(), s.String(), tcell.ColorReset.String())
}

// String returns the string value of the style
func (s *Style) String() string {
	return string(*s)
}

// Rune returns the rune value of the style
func (s *Style) Rune() rune {
	return rune(s.String()[0])
}

func isHexColor(s string) bool {
	re := regexp.MustCompile("^#(?:[0-9a-fA-F]{3}){1,2}$")
	return re.MatchString(s)
}

func getStylePath() (string, error) {
	configPath, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return configPath + "/styles.yaml", nil
}
