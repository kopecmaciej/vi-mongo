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
		Root      Root      `yaml:"main"`
		Header    Header    `yaml:"header"`
		Sidebar   Sidebar   `yaml:"sidebar"`
		Content   Content   `yaml:"content"`
		DocPeeker DocPeeker `yaml:"docPeeker"`
		InputBar  InputBar  `yaml:"filterBar"`
		Others    Others    `yaml:"others"`
	}

	// Root is a struct that contains all the root styles for the application
	Root struct {
		BackgroundColor    Style `yaml:"backgroundColor"`
		TextColor          Style `yaml:"textColor"`
		SecondaryTextColor Style `yaml:"secondaryTextColor"`
		BorderColor        Style `yaml:"borderColor"`
		FocusColor         Style `yaml:"focusColor"`
		TitleColor         Style `yaml:"titleColor"`
		GraphicsColor      Style `yaml:"graphicsColor"`
	}

	// Header is a struct that contains all the styles for the header
	Header struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		KeyColor        Style `yaml:"keyColor"`
		ValueColor      Style `yaml:"valueColor"`
		ActiveSymbol    Style `yaml:"activeSymbol"`
		InactiveSymbol  Style `yaml:"inactiveSymbol"`
	}

	// Sidebar is a struct that contains all the styles for the sidebar
	Sidebar struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		NodeColor       Style `yaml:"nodeColor"`
		LeafColor       Style `yaml:"leafColor"`
		BranchColor     Style `yaml:"branchColor"`
	}

	// Content is a struct that contains all the styles for the content
	Content struct {
		BackgroundColor  Style `yaml:"backgroundColor"`
		BorderColor      Style `yaml:"borderColor"`
		TitleColor       Style `yaml:"titleColor"`
		CellTextColor    Style `yaml:"cellTextColor"`
		ActiveRowColor   Style `yaml:"activeRowColor"`
		SelectedRowColor Style `yaml:"selectedRowColor"`
	}

	// DocPeeker is a struct that contains all the styles for the json peeker
	DocPeeker struct {
		BackgroundColor Style `yaml:"backgroundColor"`
		BorderColor     Style `yaml:"borderColor"`
		TitleColor      Style `yaml:"titleColor"`
		KeyColor        Style `yaml:"keyColor"`
		ValueColor      Style `yaml:"valueColor"`
		BracketColor    Style `yaml:"bracketColor"`
	}

	// InputBar is a struct that contains all the styles for the filter bar
	InputBar struct {
		BackgroundColor Style        `yaml:"backgroundColor"`
		BorderColor     Style        `yaml:"borderColor"`
		TitleColor      Style        `yaml:"titleColor"`
		LabelColor      Style        `yaml:"labelColor"`
		InputColor      Style        `yaml:"inputColor"`
		Autocomplete    Autocomplete `yaml:"autocomplete"`
	}

	Autocomplete struct {
		BackgroundColor       Style `yaml:"backgroundColor"`
		BorderColor           Style `yaml:"borderColor"`
		TextColor             Style `yaml:"textColor"`
		ActiveBackgroundColor Style `yaml:"activeBackgroundColor"`
		ActiveTextColor       Style `yaml:"activeTextColor"`
	}

	Others struct {
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
	s.Root = Root{
		BackgroundColor:    "#0F172A",
		TextColor:          "#FFFFFF",
		SecondaryTextColor: "#F1FA8C",
		BorderColor:        "#387D44",
		FocusColor:         "#50FA7B",
		TitleColor:         "#387D44",
		GraphicsColor:      "#387D44",
	}

	s.Header = Header{
		BackgroundColor: "#0F172A",
		BorderColor:     "#387D44",
		KeyColor:        "#F1FA8C",
		ValueColor:      "#387D44",
		ActiveSymbol:    "●",
		InactiveSymbol:  "○",
	}

	s.Sidebar = Sidebar{
		BackgroundColor: "#0F172A",
		BorderColor:     "#387D44",
		NodeColor:       "#387D44",
		LeafColor:       "#163694",
		BranchColor:     "#387D44",
	}

	s.Content = Content{
		BackgroundColor:  "#0F172A",
		BorderColor:      "#387D44",
		TitleColor:       "#163694",
		CellTextColor:    "#387D44",
		ActiveRowColor:   "#50FA7B",
		SelectedRowColor: "#50FA7B",
	}

	s.DocPeeker = DocPeeker{
		BackgroundColor: "#0F172A",
		BorderColor:     "#50FA7B",
		TitleColor:      "#F1FA8C",
		KeyColor:        "#F1FA8C",
		ValueColor:      "#FFFFFF",
		BracketColor:    "#FFFFFF",
	}

	s.InputBar = InputBar{
		BackgroundColor: "#0F172A",
		BorderColor:     "#50FA7B",
		LabelColor:      "#F1FA8C",
		InputColor:      "#FFFFFF",
		Autocomplete: Autocomplete{
			BackgroundColor:       "#0F172A",
			TextColor:             "#F1FA8C",
			BorderColor:           "#50FA7B",
			ActiveBackgroundColor: "#163694",
			ActiveTextColor:       "#FFFFFF",
		},
	}

	s.Others = Others{
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
