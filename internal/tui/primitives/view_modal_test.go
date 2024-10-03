package primitives

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestCalculateNextLinesToHighlight(t *testing.T) {
	m := NewViewModal()
	tests := []struct {
		name                 string
		lines                []string
		selectedLine         int
		nextLinesToHighlight int
	}{
		{
			name: "two line value",
			lines: []string{
				"{",
				`  "key": "value",`,
				`  "key2": "value2`,
				`verylongvaluethatdoesnotfitoneline`,
				`  "key3": "value3",`,
				"}",
			},
			selectedLine:         2,
			nextLinesToHighlight: 1,
		},
		{
			name: "Simple object",
			lines: []string{
				"{",
				`  "key": "value",`,
				`  "object": {`,
				`    "nested": "value"`,
				`  }`,
				"}",
			},
			selectedLine:         2,
			nextLinesToHighlight: 2,
		},
		{
			name: "Nested object",
			lines: []string{
				"{",
				`  "object": {`,
				`    "nested1": "value1",`,
				`    "nested2": {`,
				`      "nested3": "value3"`,
				`    }`,
				`  },`,
				`  "key": "value"`,
				"}",
			},
			selectedLine:         1,
			nextLinesToHighlight: 5,
		},
		{
			name: "Array",
			lines: []string{
				"{",
				`  "array": [`,
				`    "item1",`,
				`    "item2",`,
				`    "item3"`,
				`  ]`,
				`}`,
			},
			selectedLine:         1,
			nextLinesToHighlight: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.selectedLine = tt.selectedLine
			m.scrollPosition = 0
			result := m.calculateNextLinesToHighlight(tt.lines)
			assert.Equal(t, tt.nextLinesToHighlight, result)
		})
	}
}

func TestFormatLine(t *testing.T) {
	m := NewViewModal()
	m.SetDocumentColors(tcell.ColorRed, tcell.ColorGreen, tcell.ColorBlue)

	tests := []struct {
		name     string
		input    string
		isFirst  bool
		expected string
	}{
		{
			name:     "First line with bracket",
			input:    "{",
			isFirst:  true,
			expected: "[#0000FF]{[#008000]",
		},
		{
			name:     "Key-value pair",
			input:    `  "key": "value",`,
			isFirst:  false,
			expected: `  [#FF0000]"key"[:]: [#008000]"value",[-]`,
		},
		{
			name:     "Nested object",
			input:    `  "object": {`,
			isFirst:  false,
			expected: `  [#FF0000]"object"[:]: [#008000][#0000FF]{[-]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.formatAndColorizeLine(tt.input, tt.isFirst)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHighlightLine(t *testing.T) {
	m := NewViewModal()
	m.SetHighlightColor(tcell.ColorYellow)

	tests := []struct {
		name     string
		input    string
		withMark bool
		expected string
	}{
		{
			name:     "With mark",
			input:    "Test line",
			withMark: true,
			expected: "[-:#FFFF00:b]>Test line[-:-:-]",
		},
		{
			name:     "Without mark",
			input:    "Another test",
			withMark: false,
			expected: "[-:#FFFF00:b]Another test[-:-:-]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.highlightLine(tt.input, tt.withMark)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCopySelectedLine(t *testing.T) {
	m := NewViewModal()
	m.SetRect(0, 0, 50, 10) // Set a fixed size for testing

	tests := []struct {
		name     string
		content  string
		copyType string
		expected string
	}{
		{
			name: "Copy full single line",
			content: `{
  "key": "value",
  "object": {
    "nested": "test"
  }
}`,
			copyType: "full",
			expected: `"key": "value"`,
		},
		{
			name: "Copy full multiline line",
			content: `{
  "object": {
    "nested": "test"
  }
}`,
			copyType: "full",
			expected: `"object": { "nested": "test" }`,
		},
		{
			name: "Copy value only",
			content: `{
  "key": "value",
  "object": {
    "nested": "test"
  }
}`,
			copyType: "value",
			expected: `"value"`,
		},
		{
			name: "Copy value that's object",
			content: `{
  "object": {
    "nested": "test"
  }
}`,
			copyType: "value",
			expected: `{ "nested": "test" }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.SetText(Text{Content: tt.content})
			m.selectedLine = 1
			m.scrollPosition = 0

			var copiedText string
			copyFunc := func(text string) error {
				copiedText = text
				return nil
			}

			err := m.CopySelectedLine(copyFunc, tt.copyType)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, copiedText)
		})
	}
}
