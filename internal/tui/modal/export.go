package modal

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const ExportModalId = "ExportModal"

type ExportScope int

const (
	ExportCurrentPage ExportScope = iota
	ExportAllMatching
)

type ExportRequest struct {
	Path        string
	Scope       ExportScope
	PrettyPrint bool
}

type ExportModal struct {
	*core.BaseElement
	*core.FormModal

	exportCallback func(ExportRequest)
	exporting      bool
}

func NewExportModal() *ExportModal {
	em := &ExportModal{
		BaseElement: core.NewBaseElement(),
		FormModal:   core.NewFormModal(),
	}
	em.SetIdentifier(ExportModalId)
	em.SetAfterInitFunc(em.init)
	return em
}

func (em *ExportModal) init() error {
	em.SetTitle(" Export JSON ")
	em.SetBorder(true)
	em.SetTitleAlign(tview.AlignCenter)
	em.Form.SetBorderPadding(2, 2, 2, 2)

	styles := em.App.GetStyles()
	em.SetStyle(styles)
	em.Form.SetFieldTextColor(styles.Connection.FormInputColor.Color())
	em.Form.SetFieldBackgroundColor(styles.Connection.FormInputBackgroundColor.Color())
	em.Form.SetLabelColor(styles.Connection.FormLabelColor.Color())

	em.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			if em.exporting {
				return nil
			}
			em.Hide()
			return nil
		}
		return event
	})
	return nil
}

func (em *ExportModal) SetExportCallback(callback func(ExportRequest)) {
	em.exportCallback = callback
}

func (em *ExportModal) Render(defaultPath string) {
	em.Form.Clear(true)

	scopeLabels := []string{"Current page", "All pages"}
	scopes := []ExportScope{ExportCurrentPage, ExportAllMatching}

	em.Form.AddInputField("Path", defaultPath, 0, nil, nil)
	pathField := em.Form.GetFormItemByLabel("Path").(*tview.InputField)
	pathField.SetAutocompleteMaxHeight(10)
	pathField.SetAutocompleteFunc(pathAutocompleteEntries)
	pathField.SetAutocompletedFunc(func(text string, _ int, source int) bool {
		if source == 0 {
			return false
		}
		pathField.SetText(text)
		return !strings.HasSuffix(text, string(filepath.Separator))
	})
	em.Form.AddDropDown("Scope", scopeLabels, 0, nil)
	em.Form.AddCheckbox("Pretty print", true, nil)
	em.Form.AddButton("Export", func() {
		path := expandHomePath(em.Form.GetFormItemByLabel("Path").(*tview.InputField).GetText())
		scopeIndex, _ := em.Form.GetFormItemByLabel("Scope").(*tview.DropDown).GetCurrentOption()
		prettyPrint := em.Form.GetFormItemByLabel("Pretty print").(*tview.Checkbox).IsChecked()
		if em.exportCallback != nil {
			em.exportCallback(ExportRequest{Path: path, Scope: scopes[scopeIndex], PrettyPrint: prettyPrint})
		}
	})
	em.Form.AddButton("Cancel", em.Hide)

	em.App.Pages.AddPage(ExportModalId, em, true, true)
}

func (em *ExportModal) Hide() {
	em.SetExporting(false)
	em.App.Pages.RemovePage(ExportModalId)
}

func (em *ExportModal) SetExporting(exporting bool) {
	em.exporting = exporting
	if exporting {
		em.SetTitle(" Exporting, please wait... ")
	} else {
		em.SetTitle(" Export JSON ")
	}
	for index := 0; index < em.Form.GetFormItemCount(); index++ {
		em.Form.GetFormItem(index).SetDisabled(exporting)
	}
	for index := 0; index < em.Form.GetButtonCount(); index++ {
		em.Form.GetButton(index).SetDisabled(exporting)
	}
}

func pathAutocompleteEntries(currentText string) []tview.AutocompleteItem {
	expanded := expandHomePath(currentText)
	directory, prefix := filepath.Split(expanded)
	if directory == "" {
		directory = "."
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}

	displayDirectory, _ := filepath.Split(currentText)
	if currentText == "~" {
		displayDirectory = "~" + string(filepath.Separator)
	}
	suggestions := make([]tview.AutocompleteItem, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(prefix)) {
			continue
		}
		if !entry.IsDir() && !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(displayDirectory, entry.Name())
		secondary := "JSON file"
		if entry.IsDir() {
			path += string(filepath.Separator)
			secondary = "Directory"
		}
		suggestions = append(suggestions, tview.AutocompleteItem{Main: path, Secondary: secondary})
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		iDirectory := suggestions[i].Secondary == "Directory"
		jDirectory := suggestions[j].Secondary == "Directory"
		if iDirectory != jDirectory {
			return iDirectory
		}
		return strings.ToLower(suggestions[i].Main) < strings.ToLower(suggestions[j].Main)
	})
	return suggestions
}

func expandHomePath(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home + string(filepath.Separator)
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"+string(filepath.Separator)))
}
