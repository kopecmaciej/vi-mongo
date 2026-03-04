package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	int_mongo "github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/view"
)

const (
	AggregationId            = "Aggregation"
	AggregationStageBarId    = "AggregationStageBar"
	AggregationDeleteModalId = "AggregationDeleteModal"
)

// Aggregation is a TUI component for building and running MongoDB aggregation pipelines.
type Aggregation struct {
	*core.BaseElement
	*core.Flex

	stagesTable   *core.Table
	stageBar      *InputBar
	resultsView   *core.Flex
	resultsHeader *core.TextView
	resultsTable  *core.Table
	deleteModal   *modal.Confirm

	state       *int_mongo.CollectionState
	stateMap    *int_mongo.StateMap
	currentDB   string
	currentColl string

	isStageBarVisible bool
	editingIdx        int // -1 = adding new, >=0 = editing existing
	focusOnResults    bool
}

func NewAggregation() *Aggregation {
	a := &Aggregation{
		BaseElement:       core.NewBaseElement(),
		Flex:              core.NewFlex(),
		stagesTable:       core.NewTable(),
		stageBar:          NewInputBar(AggregationStageBarId, "Stage"),
		resultsView:       core.NewFlex(),
		resultsHeader:     core.NewTextView(),
		resultsTable:      core.NewTable(),
		deleteModal:       modal.NewConfirm(AggregationDeleteModalId),
		state:             &int_mongo.CollectionState{},
		stateMap:          int_mongo.NewStateMap(),
		isStageBarVisible: false,
		editingIdx:        -1,
	}

	a.SetIdentifier(AggregationId)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *Aggregation) init() error {
	a.setLayout()
	a.setStyle()
	a.setKeybindings()

	if err := a.deleteModal.Init(a.App); err != nil {
		return err
	}
	if err := a.stageBar.Init(a.App); err != nil {
		return err
	}

	a.stageBar.EnableAggregationAutocomplete()
	a.stageBar.EnableHistory()
	a.stageBar.SetDefaultText("{ <$0> }")

	a.stageBarHandler()
	a.handleEvents()

	return nil
}

func (a *Aggregation) setLayout() {
	a.SetDirection(tview.FlexRow)
	a.SetBorder(true)
	a.SetTitle(" Aggregation ")
	a.SetTitleAlign(tview.AlignCenter)
	a.SetBorderPadding(0, 0, 1, 1)

	a.stagesTable.SetIdentifier(AggregationId)
	a.stagesTable.SetSelectable(true, false)
	a.stagesTable.SetFixed(1, 0)

	a.resultsView.SetDirection(tview.FlexRow)
	a.resultsHeader.SetDynamicColors(true)

	a.resultsTable.SetSelectable(true, false)
}

func (a *Aggregation) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)
	a.stagesTable.SetStyle(styles)
	a.resultsTable.SetStyle(styles)
	a.resultsView.SetStyle(styles)

	a.stagesTable.SetSeparator(styles.Others.SeparatorSymbol.Rune())
	a.stagesTable.SetBordersColor(styles.Others.SeparatorColor.Color())
	a.resultsTable.SetSeparator(styles.Others.SeparatorSymbol.Rune())
	a.resultsTable.SetBordersColor(styles.Others.SeparatorColor.Color())
	a.resultsHeader.SetTextColor(styles.Content.StatusTextColor.Color())
}

func (a *Aggregation) handleEvents() {
	go a.HandleEvents(AggregationId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			a.setStyle()
			a.Render()
		case manager.UpdateAutocompleteKeys:
			if keys, ok := event.Message.Data.([]string); ok {
				a.stageBar.LoadAutocomleteKeys(keys)
			}
		}
	})
}

func (a *Aggregation) setKeybindings() {
	k := a.App.GetKeys()

	a.stagesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Aggregation.AddStage, event.Name()):
			a.openStageBar(-1)
			return nil
		case k.Contains(k.Aggregation.EditStage, event.Name()):
			row, _ := a.stagesTable.GetSelection()
			if row >= 1 {
				a.openStageBar(row - 1)
			}
			return nil
		case k.Contains(k.Aggregation.DeleteStage, event.Name()):
			a.showDeleteStageModal()
			return nil
		case k.Contains(k.Aggregation.RunPipeline, event.Name()):
			ctx := context.Background()
			a.runPipeline(ctx)
			return nil
		case k.Contains(k.Aggregation.ClearPipeline, event.Name()):
			a.clearPipeline()
			return nil
		case k.Contains(k.Aggregation.MoveStageDown, event.Name()):
			a.moveStage(1)
			return nil
		case k.Contains(k.Aggregation.MoveStageUp, event.Name()):
			a.moveStage(-1)
			return nil
		case k.Contains(k.Aggregation.FocusResults, event.Name()):
			a.focusOnResults = true
			a.App.SetFocus(a.resultsTable)
			return nil
		}
		return event
	})

	a.resultsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Aggregation.FocusResults, event.Name()):
			a.focusOnResults = false
			a.App.SetFocus(a.stagesTable)
			return nil
		}
		return event
	})
}

func (a *Aggregation) HandleDatabaseSelection(ctx context.Context, db, coll string) error {
	a.currentDB = db
	a.currentColl = coll

	stateKey := a.stateMap.Key(db, coll)
	state, ok := a.stateMap.Get(stateKey)
	if ok {
		a.state = state
	} else {
		a.state = int_mongo.NewCollectionState(db, coll)
		a.stateMap.Set(stateKey, a.state)
	}

	a.Render()
	return nil
}

func (a *Aggregation) Render() {
	a.Flex.Clear()

	a.renderStagesTable()

	stagesFlex := core.NewFlex()
	stagesFlex.SetDirection(tview.FlexRow)
	stagesFlex.SetBorder(true)
	stagesFlex.SetTitle(fmt.Sprintf(" PIPELINE STAGES  %d stgs ", len(a.state.GetPipelineStages())))
	stagesFlex.SetTitleAlign(tview.AlignLeft)
	stagesFlex.AddItem(a.stagesTable, 0, 1, true)

	a.Flex.AddItem(stagesFlex, 0, 1, true)

	if a.isStageBarVisible {
		a.Flex.AddItem(a.stageBar, 3, 0, false)
	}

	a.renderResultsView()
	a.Flex.AddItem(a.resultsView, 0, 2, false)

	if a.isStageBarVisible {
		a.App.SetFocus(a.stageBar)
	} else if a.focusOnResults {
		a.App.SetFocus(a.resultsTable)
	} else {
		a.App.SetFocus(a.stagesTable)
	}
}

func (a *Aggregation) renderStagesTable() {
	styles := a.App.GetStyles()
	a.stagesTable.Clear()
	a.stagesTable.SetFixed(1, 0)

	headers := []string{"#", "Operator", "Preview"}
	for col, header := range headers {
		cell := tview.NewTableCell(" " + header + " ").
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetTextColor(styles.Content.ColumnKeyColor.Color()).
			SetBackgroundColor(styles.Content.HeaderRowBackgroundColor.Color())
		a.stagesTable.SetCell(0, col, cell)
	}

	stages := a.state.GetPipelineStages()
	for row, stage := range stages {
		operator := int_mongo.ExtractStageOperator(stage)
		preview := stage

		a.stagesTable.SetCell(row+1, 0, tview.NewTableCell(fmt.Sprintf(" [%d] ", row)).
			SetAlign(tview.AlignCenter).
			SetReference(row))
		a.stagesTable.SetCell(row+1, 1, tview.NewTableCell(" "+operator+" ").
			SetAlign(tview.AlignLeft).
			SetTextColor(styles.Content.ColumnKeyColor.Color()))
		a.stagesTable.SetCell(row+1, 2, tview.NewTableCell(" "+preview+" ").
			SetAlign(tview.AlignLeft))
	}

	if len(stages) > 0 {
		a.stagesTable.Select(1, 0)
	}
}

func (a *Aggregation) renderResultsView() {
	a.resultsView.Clear()
	a.resultsTable.Clear()

	docs := a.state.GetAggDocs()

	resultsHeaderFlex := core.NewFlex()
	resultsHeaderFlex.SetDirection(tview.FlexRow)
	resultsHeaderFlex.SetBorder(true)
	resultsHeaderFlex.SetTitle(fmt.Sprintf(" RESULTS  %d docs ", len(docs)))
	resultsHeaderFlex.SetTitleAlign(tview.AlignLeft)

	tableJson := view.NewTableJson()
	if err := tableJson.Render(a.resultsTable, 0, docs); err != nil {
		modal.ShowError(a.App.Pages, "Error rendering results", err)
	}

	resultsHeaderFlex.AddItem(a.resultsTable, 0, 1, false)
	a.resultsView.AddItem(resultsHeaderFlex, 0, 1, false)
}

func (a *Aggregation) openStageBar(idx int) {
	a.editingIdx = idx
	a.isStageBarVisible = true

	if idx >= 0 {
		stages := a.state.GetPipelineStages()
		if idx < len(stages) {
			a.stageBar.SetText(stages[idx])
		}
	} else {
		a.stageBar.Toggle("")
	}

	a.Render()
}

func (a *Aggregation) closeStageBar() {
	a.isStageBarVisible = false
	a.editingIdx = -1
	a.Render()
}

func (a *Aggregation) stageBarHandler() {
	a.stageBar.DoneFuncHandler(
		func(text string) {
			a.applyStage(text)
		},
		func() {
			a.closeStageBar()
		},
	)
}

func (a *Aggregation) applyStage(text string) {
	text = strings.TrimSpace(text)
	if text == "" || text == "{ <$0> }" {
		a.closeStageBar()
		return
	}

	// Validate it parses
	_, err := int_mongo.ParsePipeline([]string{text})
	if err != nil {
		modal.ShowError(a.App.Pages, "Invalid stage", err)
		return
	}

	// Validate that the top-level operator starts with '$'
	operator := int_mongo.ExtractStageOperator(text)
	if !strings.HasPrefix(operator, "$") {
		modal.ShowError(a.App.Pages, "Invalid stage", fmt.Errorf("stage operator %q must start with '$' (e.g. $match, $group)", operator))
		return
	}

	stages := a.state.GetPipelineStages()
	if a.editingIdx >= 0 && a.editingIdx < len(stages) {
		stages[a.editingIdx] = text
	} else {
		stages = append(stages, text)
	}
	a.state.SetPipelineStages(stages)

	a.closeStageBar()
}

func (a *Aggregation) showDeleteStageModal() {
	row, _ := a.stagesTable.GetSelection()
	if row < 1 {
		return
	}
	idx := row - 1
	stages := a.state.GetPipelineStages()
	if idx >= len(stages) {
		return
	}

	operator := int_mongo.ExtractStageOperator(stages[idx])
	a.deleteModal.SetText(fmt.Sprintf("Delete stage [%d] %s?", idx, operator))
	a.deleteModal.SetDoneFunc(func(buttonIndex int, _ string) {
		defer a.App.Pages.RemovePage(AggregationDeleteModalId)
		if buttonIndex == 0 {
			a.deleteStage(idx)
		}
	})
	a.App.Pages.AddPage(AggregationDeleteModalId, a.deleteModal, true, true)
}

func (a *Aggregation) deleteStage(idx int) {
	stages := a.state.GetPipelineStages()
	if idx < 0 || idx >= len(stages) {
		return
	}
	stages = append(stages[:idx], stages[idx+1:]...)
	a.state.SetPipelineStages(stages)
	a.Render()
}

func (a *Aggregation) clearPipeline() {
	a.state.SetPipelineStages(nil)
	a.state.SetAggDocs(nil)
	a.Render()
}

func (a *Aggregation) moveStage(direction int) {
	row, _ := a.stagesTable.GetSelection()
	if row < 1 {
		return
	}
	idx := row - 1
	stages := a.state.GetPipelineStages()
	newIdx := idx + direction
	if newIdx < 0 || newIdx >= len(stages) {
		return
	}
	stages[idx], stages[newIdx] = stages[newIdx], stages[idx]
	a.state.SetPipelineStages(stages)
	a.Render()
	// Re-select the moved item
	a.stagesTable.Select(newIdx+1, 0)
}

func (a *Aggregation) IsStageBarVisible() bool {
	return a.isStageBarVisible
}

func (a *Aggregation) runPipeline(ctx context.Context) {
	stages := a.state.GetPipelineStages()
	if len(stages) == 0 {
		modal.ShowError(a.App.Pages, "No stages", fmt.Errorf("add at least one stage before running"))
		return
	}

	pipeline, err := int_mongo.ParsePipeline(stages)
	if err != nil {
		modal.ShowError(a.App.Pages, "Pipeline parse error", err)
		return
	}

	docs, err := a.Dao.AggregateDocuments(ctx, a.currentDB, a.currentColl, pipeline)
	if err != nil {
		modal.ShowError(a.App.Pages, "Aggregation error", err)
		return
	}

	a.state.SetAggDocs(docs)
	a.Render()
}
