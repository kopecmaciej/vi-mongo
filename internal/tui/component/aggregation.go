package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/view"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	AggregationId            = "Aggregation"
	AggregationResultsId     = "AggregationResults"
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
	peeker        *Peeker
	tableJson     *view.TableJson
	tableColumns  *view.TableColumns

	state       *mongo.CollectionState
	stateMap    *mongo.StateMap
	currentDB   string
	currentColl string

	editingIdx     int // -1 = adding new, >=0 = editing existing
	focusOnResults bool
	isPreview      bool
	currentView    ViewType
}

func NewAggregation() *Aggregation {
	a := &Aggregation{
		BaseElement:   core.NewBaseElement(),
		Flex:          core.NewFlex(),
		stagesTable:   core.NewTable(),
		stageBar:      NewInputBar(AggregationStageBarId, "Stage"),
		resultsView:   core.NewFlex(),
		resultsHeader: core.NewTextView(),
		resultsTable:  core.NewTable(),
		deleteModal:   modal.NewConfirm(AggregationDeleteModalId),
		peeker:        NewPeeker(),
		tableJson:     view.NewTableJson(),
		state:         &mongo.CollectionState{},
		stateMap:      mongo.NewStateMap(),
		editingIdx:    -1,
		currentView:   JsonView,
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
	if err := a.peeker.Init(a.App); err != nil {
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

	a.resultsTable.SetIdentifier(AggregationResultsId)

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

	a.tableColumns = view.NewTableColumns(&styles.Content)
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
		case k.Contains(k.Aggregation.Stages.AddStage, event.Name()):
			a.handleToggleStage(-1)
			return nil
		case k.Contains(k.Aggregation.Stages.EditStage, event.Name()):
			row, _ := a.stagesTable.GetSelection()
			if row >= 1 {
				a.handleToggleStage(row - 1)
			}
			return nil
		case k.Contains(k.Aggregation.Stages.DeleteStage, event.Name()):
			a.showDeleteStageModal()
			return nil
		case k.Contains(k.Aggregation.Stages.RunPipeline, event.Name()):
			ctx := context.Background()
			a.runPipeline(ctx, false)
			return nil
		case k.Contains(k.Aggregation.Stages.ClearPipeline, event.Name()):
			a.clearPipeline()
			return nil
		case k.Contains(k.Aggregation.Stages.MoveStageDown, event.Name()):
			a.moveStage(1)
			return nil
		case k.Contains(k.Aggregation.Stages.MoveStageUp, event.Name()):
			a.moveStage(-1)
			return nil
		case k.Contains(k.Aggregation.Stages.FocusResults, event.Name()):
			a.focusOnResults = true
			a.App.SetFocus(a.resultsTable)
			return nil
		}
		return event
	})

	a.resultsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Aggregation.Results.FocusStages, event.Name()):
			a.focusOnResults = false
			a.App.SetFocus(a.stagesTable)
			return nil
		case k.Contains(k.Aggregation.Results.ChangeView, event.Name()):
			a.toggleResultsView()
			return nil
		case k.Contains(k.Aggregation.Results.PeekDocument, event.Name()):
			a.handlePeekDocument(false)
			return nil
		case k.Contains(k.Aggregation.Results.FullPagePeek, event.Name()):
			a.handlePeekDocument(true)
			return nil
		case k.Contains(k.Aggregation.Results.CopyHighlight, event.Name()):
			row, col := a.resultsTable.GetSelection()
			a.handleCopyLine(row, col)
			return nil
		case k.Contains(k.Aggregation.Results.CopyDocument, event.Name()):
			row, col := a.resultsTable.GetSelection()
			a.handleCopyDocument(row, col)
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
		a.state = mongo.NewCollectionState(db, coll)
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

	if a.stageBar.IsEnabled() {
		a.Flex.AddItem(a.stageBar, 3, 0, false)
	}

	a.renderResultsView()
	a.Flex.AddItem(a.resultsView, 0, 2, false)

	if a.stageBar.IsEnabled() {
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
		operator := mongo.ExtractStageOperator(stage)
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
	title := fmt.Sprintf(" RESULTS  %d docs ", len(docs))
	if a.isPreview {
		title = fmt.Sprintf(" RESULTS  %d docs (preview) ", len(docs))
	}
	resultsHeaderFlex.SetTitle(title)
	resultsHeaderFlex.SetTitleAlign(tview.AlignLeft)

	a.resultsTable.SetSelectable(true, a.currentView == TableView)
	switch a.currentView {
	case TableView:
		a.tableColumns.Render(a.resultsTable, 0, docs)
	case JsonView:
		if err := a.tableJson.Render(a.resultsTable, 0, docs); err != nil {
			modal.ShowError(a.App.Pages, "Error rendering results", err)
		}
	}

	resultsHeaderFlex.AddItem(a.resultsTable, 0, 1, false)
	a.resultsView.AddItem(resultsHeaderFlex, 0, 1, false)
}

func (a *Aggregation) handleToggleStage(idx int) {
	a.editingIdx = idx
	if idx >= 0 {
		stages := a.state.GetPipelineStages()
		if idx < len(stages) {
			a.stageBar.Open(stages[idx])
		}
	} else {
		a.stageBar.Toggle("")
	}
	a.Render()
}

func (a *Aggregation) closeStageBar() {
	a.stageBar.Close()
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
	a.stageBar.SetText("")
	text = strings.TrimSpace(text)
	if text == "" || text == "{ <$0> }" {
		a.closeStageBar()
		return
	}

	// Validate it parses
	_, err := mongo.ParsePipeline([]string{text})
	if err != nil {
		modal.ShowError(a.App.Pages, "Invalid stage", err)
		return
	}

	// Validate that the top-level operator starts with '$'
	operator := mongo.ExtractStageOperator(text)
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
	a.runPipeline(context.Background(), true)
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

	operator := mongo.ExtractStageOperator(stages[idx])
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
	return a.stageBar.IsEnabled()
}

func (a *Aggregation) toggleResultsView() {
	if a.currentView == JsonView {
		a.currentView = TableView
	} else {
		a.currentView = JsonView
	}
	a.renderResultsView()
}

func (a *Aggregation) getAggDocId(row, col int) any {
	switch a.currentView {
	case JsonView:
		cell := a.resultsTable.GetCellAboveThatMatch(row, col, func(c *tview.TableCell) bool {
			return c.GetReference() != nil
		})
		if cell == nil {
			return nil
		}
		return cell.GetReference()
	case TableView:
		return a.resultsTable.GetCell(row, 0).GetReference()
	}
	return nil
}

func (a *Aggregation) handleCopyLine(row, col int) {
	cell := a.resultsTable.GetCell(row, col)
	if cell == nil {
		return
	}
	if err := clipboard.WriteAll(strings.TrimSpace(cell.Text)); err != nil {
		modal.ShowError(a.App.Pages, "Error copying cell", err)
	}
}

func (a *Aggregation) handleCopyDocument(row, col int) {
	_id := a.getAggDocId(row, col)
	if _id == nil {
		return
	}
	doc, err := a.state.GetJsonDocById(_id)
	if err != nil {
		modal.ShowError(a.App.Pages, "Error copying document", err)
		return
	}
	if err := clipboard.WriteAll(doc); err != nil {
		modal.ShowError(a.App.Pages, "Error copying document", err)
	}
}

func (a *Aggregation) handlePeekDocument(fullScreen bool) {
	row, col := a.resultsTable.GetSelection()
	_id := a.getAggDocId(row, col)
	if _id == nil {
		return
	}
	a.peeker.SetFullScreen(fullScreen)
	if err := a.peeker.Render(context.Background(), a.state, _id); err != nil {
		modal.ShowError(a.App.Pages, "Error peeking document", err)
	}
}

func (a *Aggregation) runPipeline(ctx context.Context, preview bool) {
	stages := a.state.GetPipelineStages()
	if len(stages) == 0 {
		modal.ShowError(a.App.Pages, "No stages", fmt.Errorf("add at least one stage before running"))
		return
	}

	pipeline, err := mongo.ParsePipeline(stages)
	if err != nil {
		modal.ShowError(a.App.Pages, "Pipeline parse error", err)
		return
	}

	if preview {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: 5}})
	}

	docs, err := a.Dao.AggregateDocuments(ctx, a.currentDB, a.currentColl, pipeline)
	if err != nil {
		modal.ShowError(a.App.Pages, "Aggregation error", err)
		return
	}

	a.isPreview = preview
	a.state.SetAggDocs(docs)
	a.Render()
}
