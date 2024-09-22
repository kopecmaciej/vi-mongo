package config

import (
	"embed"
	"fmt"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/util"
)

//go:embed styles
var stylesFS embed.FS

// Styles is a struct that contains all the styles for the application
type (
	Style string

	Styles struct {
		Global      GlobalStyles     `yaml:"global"`
		Welcome     WelcomeStyle     `yaml:"welcome"`
		Connection  ConnectionStyle  `yaml:"connection"`
		Header      HeaderStyle      `yaml:"header"`
		Databases   DatabasesStyle   `yaml:"databases"`
		Content     ContentStyle     `yaml:"content"`
		DocPeeker   DocPeekerStyle   `yaml:"docPeeker"`
		InputBar    InputBarStyle    `yaml:"filterBar"`
		History     HistoryStyle     `yaml:"history"`
		Help        HelpStyle        `yaml:"help"`
		Others      OthersStyle      `yaml:"others"`
		StyleChange StyleChangeStyle `yaml:"styleChange"`
	}

	// GlobalStyles is a struct that contains all the global styles for the application
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

	// ConnectionStyle is a struct that contains all the styles for the connection
	ConnectionStyle struct {
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
		NodeColor        Style `yaml:"nodeColor"`
		OpenNodeSymbol   Style `yaml:"openNodeSymbol"`
		ClosedNodeSymbol Style `yaml:"closedNodeSymbol"`
		LeafColor        Style `yaml:"leafColor"`
		LeafSymbol       Style `yaml:"leafSymbol"`
		BranchColor      Style `yaml:"branchColor"`
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
		// buttons
		ButtonsTextColor       Style `yaml:"buttonsTextColor"`
		ButtonsSelectedColor   Style `yaml:"buttonsSelectedColor"`
		ButtonsBackgroundColor Style `yaml:"buttonsBackgroundColor"`
		// modals specials
		ModalTextColor          Style `yaml:"modalTextColor"`
		ModalSecondaryTextColor Style `yaml:"modalSecondaryTextColor"`
	}

	StyleChangeStyle struct {
		TextColor               Style `yaml:"textColor"`
		SelectedTextColor       Style `yaml:"selectedTextColor"`
		SelectedBackgroundColor Style `yaml:"selectedBackgroundColor"`
	}
)

func (s *Styles) loadDefaults() {
	s.Global = GlobalStyles{
		BackgroundColor:    "#0F172A",
		TextColor:          "#E2E8F0",
		SecondaryTextColor: "#FDE68A",
		BorderColor:        "#387D44",
		FocusColor:         "#4ADE80",
		TitleColor:         "#387D44",
		GraphicsColor:      "#387D44",
	}

	s.Welcome = WelcomeStyle{
		TextColor:                "#FDE68A",
		FormLabelColor:           "#FDE68A",
		FormInputColor:           "#E2E8F0",
		FormInputBackgroundColor: "#1E293B",
	}

	s.Connection = ConnectionStyle{
		FormLabelColor:               "#F1FA8C",
		FormInputBackgroundColor:     "#163694",
		FormInputColor:               "#F1FA8C",
		FormButtonColor:              "#387D44",
		ListTextColor:                "#F1FA8C",
		ListSelectedTextColor:        "#F1FA8C",
		ListSelectedBackgroundColor:  "#163694",
		ListSecondaryTextColor:       "#387D44",
		ListSecondaryBackgroundColor: "#0F172A",
	}

	s.Header = HeaderStyle{
		KeyColor:       "#FDE68A",
		ValueColor:     "#387D44",
		ActiveSymbol:   "●",
		InactiveSymbol: "○",
	}

	s.Databases = DatabasesStyle{
		NodeColor:        "#387D44",
		LeafColor:        "#E2E8F0",
		BranchColor:      "#4ADE80",
		OpenNodeSymbol:   "[#FDE68A]🗁[-:-:-]",
		ClosedNodeSymbol: "[#FDE68A]🖿[-:-:-]",
		LeafSymbol:       "[#387D44]🗎[-:-:-]",
	}

	s.Content = ContentStyle{
		StatusTextColor:          "#FDE68A",
		HeaderRowBackgroundColor: "#1E293B",
		ColumnKeyColor:           "#FDE68A",
		ColumnTypeColor:          "#387D44",
		CellTextColor:            "#387D44",
		ActiveRowColor:           "#4ADE80",
		SelectedRowColor:         "#4ADE80",
		SeparatorSymbol:          "|",
		SeparatorColor:           "#334155",
	}

	s.DocPeeker = DocPeekerStyle{
		KeyColor:       "#387D44",
		ValueColor:     "#E2E8F0",
		ArrayColor:     "#387D44",
		HighlightColor: "#3a4963",
		BracketColor:   "#FDE68A",
	}

	s.InputBar = InputBarStyle{
		LabelColor: "#FDE68A",
		InputColor: "#E2E8F0",
		Autocomplete: AutocompleteStyle{
			BackgroundColor:       "#0F172A",
			TextColor:             "#E2E8F0",
			ActiveBackgroundColor: "#387D44",
			ActiveTextColor:       "#0F172A",
			SecondaryTextColor:    "#FDE68A",
		},
	}

	s.History = HistoryStyle{
		TextColor:               "#E2E8F0",
		SelectedTextColor:       "#0F172A",
		SelectedBackgroundColor: "#387D44",
	}

	s.Help = HelpStyle{
		HeaderColor:      "#387D44",
		KeyColor:         "#FDE68A",
		DescriptionColor: "#E2E8F0",
	}

	s.Others = OthersStyle{
		ButtonsTextColor:        "#0F172A",
		ButtonsSelectedColor:    "#387D44",
		ButtonsBackgroundColor:  "#1E293B",
		ModalTextColor:          "#FDE68A",
		ModalSecondaryTextColor: "#387D44",
	}

	s.StyleChange = StyleChangeStyle{
		TextColor:               "#E2E8F0",
		SelectedTextColor:       "#0F172A",
		SelectedBackgroundColor: "#387D44",
	}
}

// LoadStyles creates a new Styles struct with default values
func LoadStyles(styleName string) (*Styles, error) {
	defaultStyles := &Styles{}
	defaultStyles.loadDefaults()

	if os.Getenv("ENV") == "vi-dev" {
		return defaultStyles, nil
	}

	stylePath, err := getStylePath(styleName)
	if err != nil {
		return nil, err
	}

	if err := ExtractStyles(); err != nil {
		return nil, err
	}

	return util.LoadConfigFile(defaultStyles, stylePath)
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
	return util.IsHexColor(s)
}

func getStylePath(styleName string) (string, error) {
	configPath, err := util.GetConfigDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/styles/%s", configPath, styleName), nil
}

func GetAllStyles() ([]string, error) {
	configPath, err := util.GetConfigDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(fmt.Sprintf("%s/styles", configPath))
	if err != nil {
		return nil, err
	}

	styleNames := make([]string, 0, len(files))
	for _, file := range files {
		styleNames = append(styleNames, file.Name())
	}
	return styleNames, nil
}

func ExtractStyles() error {
	configDir, err := util.GetConfigDir()
	if err != nil {
		return err
	}

	stylesDir := fmt.Sprintf("%s/styles", configDir)

	// Check if styles directory exists
	if info, err := os.Stat(stylesDir); err == nil && info.IsDir() {
		// Check if any style files exist
		entries, err := os.ReadDir(stylesDir)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			// Styles already exist, return early
			return nil
		}
	} else if os.IsNotExist(err) {
		// Create styles directory if it doesn't exist
		if err := os.MkdirAll(stylesDir, 0755); err != nil {
			return err
		}
	} else {
		// Return any other error
		return err
	}

	// Populate styles directory
	entries, err := stylesFS.ReadDir("styles")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			content, err := stylesFS.ReadFile("styles/" + entry.Name())
			if err != nil {
				return err
			}

			err = os.WriteFile(stylesDir+"/"+entry.Name(), content, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
