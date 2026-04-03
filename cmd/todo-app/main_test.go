package main

import (
	"testing"

	terma "github.com/darrenburns/terma"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskList_BlurPreservesSelectionWhenMoveMenuOpen(t *testing.T) {
	app := NewTodoApp()
	app.activeList().Tasks.Select(1)
	app.showMoveMenu.Set(true)

	widget := app.buildTaskList(terma.BuildContext{})
	scrollable, ok := widget.(terma.Scrollable)
	require.True(t, ok)

	listWidget, ok := scrollable.Child.(terma.List[Task])
	require.True(t, ok)

	listWidget.Blur()
	require.Len(t, app.activeList().Tasks.SelectedItems(), 1)
}

func TestBuildTaskList_BlurClearsSelectionWhenMoveMenuClosed(t *testing.T) {
	app := NewTodoApp()
	app.activeList().Tasks.Select(1)

	widget := app.buildTaskList(terma.BuildContext{})
	scrollable, ok := widget.(terma.Scrollable)
	require.True(t, ok)

	listWidget, ok := scrollable.Child.(terma.List[Task])
	require.True(t, ok)

	listWidget.Blur()
	require.Empty(t, app.activeList().Tasks.SelectedItems())
}

func TestMoveTaskToList_MovesAllSelectedTasks(t *testing.T) {
	app := NewTodoApp()
	source := app.activeList()
	target := app.taskLists[1]

	sourceItems := source.Tasks.GetItems()
	require.GreaterOrEqual(t, len(sourceItems), 4)
	firstSelected := sourceItems[1]
	secondSelected := sourceItems[3]

	source.Tasks.Select(1)
	source.Tasks.Select(3)

	app.moveTaskToList(target)

	sourceIDs := collectTaskIDs(source.Tasks.GetItems())
	require.NotContains(t, sourceIDs, firstSelected.ID)
	require.NotContains(t, sourceIDs, secondSelected.ID)

	targetItems := target.Tasks.GetItems()
	require.GreaterOrEqual(t, len(targetItems), 2)
	require.Equal(t, firstSelected.ID, targetItems[0].ID)
	require.Equal(t, secondSelected.ID, targetItems[1].ID)
}

func TestMoveTaskToList_UsesFilteredSelectionInFilterMode(t *testing.T) {
	app := NewTodoApp()
	source := app.activeList()
	target := app.taskLists[1]

	app.filterMode.Set(true)
	app.filterInputState.SetText("#fun")
	app.refreshFilteredTasks()

	filteredItems := app.filteredListState.GetItems()
	require.GreaterOrEqual(t, len(filteredItems), 2)
	firstSelected := filteredItems[0]
	secondSelected := filteredItems[1]

	app.filteredListState.Select(0)
	app.filteredListState.Select(1)

	app.moveTaskToList(target)

	sourceIDs := collectTaskIDs(source.Tasks.GetItems())
	require.NotContains(t, sourceIDs, firstSelected.ID)
	require.NotContains(t, sourceIDs, secondSelected.ID)

	targetItems := target.Tasks.GetItems()
	require.GreaterOrEqual(t, len(targetItems), 2)
	require.Equal(t, firstSelected.ID, targetItems[0].ID)
	require.Equal(t, secondSelected.ID, targetItems[1].ID)

	require.Empty(t, app.filteredListState.SelectedItems())
}

func TestKeybinds_IncludeNewTaskShortcut(t *testing.T) {
	app := NewTodoApp()

	keybind, ok := findKeybindByKey(app.Keybinds(), "n")
	require.True(t, ok)
	require.Equal(t, "New", keybind.Name)
	require.NotNil(t, keybind.Action)
}

func TestNewTodoApp_IncludesArchiveList(t *testing.T) {
	app := NewTodoApp()

	require.Len(t, app.taskLists, 3)
	require.Equal(t, archiveListID, app.taskLists[2].ID)
	require.Equal(t, "Archive", app.taskLists[2].Name)
}

func TestBuildInputRow_NewTaskInputUsesTextAreaAndIncludesEscapeToTaskList(t *testing.T) {
	app := NewTodoApp()

	widget := app.buildInputRow(terma.BuildContext{}, terma.ThemeData{})
	row, ok := widget.(terma.Row)
	require.True(t, ok)
	require.Len(t, row.Children, 2)

	autocomplete, ok := row.Children[1].(terma.Autocomplete)
	require.True(t, ok)

	input, ok := autocomplete.Child.(terma.TextArea)
	require.True(t, ok)

	keybind, ok := findKeybindByKey(input.ExtraKeybinds, "escape")
	require.True(t, ok)
	require.Equal(t, "Tasks", keybind.Name)
	require.NotNil(t, keybind.Action)

	keybind, ok = findKeybindByKey(input.ExtraKeybinds, "shift+enter")
	require.True(t, ok)
	require.Equal(t, "Newline", keybind.Name)
}

func TestBuildInputRow_NewTaskInputUsesStrongerBackgroundWhenFocused(t *testing.T) {
	app := NewTodoApp()
	theme := terma.ThemeData{
		Background: terma.Hex("#101010"),
		Surface:    terma.Hex("#202020"),
	}
	ctx := terma.NewBuildContext(
		nil,
		terma.NewAnySignal[terma.Focusable](testFocusable{id: "new-task-input"}),
		terma.AnySignal[terma.Widget]{},
		nil,
	)

	widget := app.buildInputRow(ctx, theme)
	row, ok := widget.(terma.Row)
	require.True(t, ok)
	require.Equal(t, theme.Background.Blend(theme.Surface, 0.45), row.Style.BackgroundColor)

	autocomplete, ok := row.Children[1].(terma.Autocomplete)
	require.True(t, ok)

	input, ok := autocomplete.Child.(terma.TextArea)
	require.True(t, ok)
	require.Equal(t, theme.Background.Blend(theme.Surface, 0.45), input.Style.BackgroundColor)
}

func TestIncompleteTodoColors_UseTextColorAndDimmerCircle(t *testing.T) {
	theme := terma.ThemeData{
		Background: terma.Hex("#101010"),
		TextMuted:  terma.Hex("#666666"),
		Text:       terma.Hex("#eeeeee"),
	}

	textColor := theme.TextMuted.Blend(theme.Text, 0.35)
	require.Equal(t, textColor, incompleteTodoTextColor(theme))
	require.Equal(t, textColor.Blend(theme.Background, 0.2), incompleteTodoCircleColor(theme))
}

func TestInitialFocusTarget_TaskListWhenActiveListHasTasks(t *testing.T) {
	app := NewTodoApp()
	ctx := terma.NewBuildContext(nil, terma.AnySignal[terma.Focusable]{}, terma.AnySignal[terma.Widget]{}, nil)

	require.Equal(t, "task-list", app.initialFocusTarget(ctx))
}

func TestInitialFocusTarget_EmptyWhenActiveListHasNoTasks(t *testing.T) {
	app := NewTodoApp()
	app.activeListIdx.Set(1)
	ctx := terma.NewBuildContext(nil, terma.AnySignal[terma.Focusable]{}, terma.AnySignal[terma.Widget]{}, nil)

	require.Empty(t, app.initialFocusTarget(ctx))
}

func TestInitialFocusTarget_EmptyWhenSomethingIsAlreadyFocused(t *testing.T) {
	app := NewTodoApp()
	ctx := terma.NewBuildContext(
		nil,
		terma.NewAnySignal[terma.Focusable](testFocusable{id: "already-focused"}),
		terma.AnySignal[terma.Widget]{},
		nil,
	)

	require.Empty(t, app.initialFocusTarget(ctx))
}

func TestAddTask_SelectsNewTaskAndResetsListViewport(t *testing.T) {
	app := NewTodoApp()
	listState := app.activeList().Tasks
	scrollState := app.activeList().ScrollState

	listState.Select(2)
	scrollState.SetOffset(4)

	app.addTask("  New task  ")

	items := listState.GetItems()
	require.NotEmpty(t, items)
	require.Equal(t, "New task", items[0].Title)
	require.Equal(t, 0, listState.CursorIndex.Peek())
	require.Empty(t, listState.SelectedItems())
	require.Equal(t, 0, scrollState.GetOffset())
	require.Equal(t, "", app.inputState.GetText())
}

func TestBuildFooter_AddsSpacerAboveKeybindBar(t *testing.T) {
	app := NewTodoApp()
	theme := terma.ThemeData{
		Primary:   terma.Hex("#ff00aa"),
		TextMuted: terma.Hex("#778899"),
	}

	widget := app.buildFooter(theme)
	column, ok := widget.(terma.Column)
	require.True(t, ok)
	require.Len(t, column.Children, 3)
	require.Equal(t, 0, column.Spacing)

	switcher, ok := column.Children[0].(terma.Text)
	require.True(t, ok)
	require.Equal(t, terma.TextAlignCenter, switcher.TextAlign)
	require.Equal(t, terma.EdgeInsets{}, switcher.Style.Padding)
	require.Equal(t, terma.EdgeInsets{}, switcher.Style.Margin)
	require.True(t, switcher.Height.IsCells())
	require.Equal(t, 1, switcher.Height.CellsValue())

	spacer, ok := column.Children[1].(terma.Spacer)
	require.True(t, ok)
	require.True(t, spacer.Height.IsCells())
	require.Equal(t, 1, spacer.Height.CellsValue())

	_, ok = column.Children[2].(terma.KeybindBar)
	require.True(t, ok)
}

func TestBuildListSwitcher_CentersActiveListInSequence(t *testing.T) {
	app := NewTodoApp()
	theme := terma.ThemeData{
		Primary:   terma.Hex("#ff00aa"),
		TextMuted: terma.Hex("#778899"),
	}

	widget := app.buildListSwitcher(theme)
	text, ok := widget.(terma.Text)
	require.True(t, ok)
	require.True(t, text.Width.IsFlex())
	require.Equal(t, 1.0, text.Width.FlexValue())
	require.Equal(t, terma.TextAlignCenter, text.TextAlign)
	require.Len(t, text.Spans, 5)
	require.Equal(t, app.taskLists[0].Name, text.Spans[0].Text)
	require.Equal(t, theme.Primary, text.Spans[0].Style.Foreground)
	require.Equal(t, " · ", text.Spans[1].Text)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), text.Spans[1].Style.Foreground)
	require.Equal(t, app.taskLists[1].Name, text.Spans[2].Text)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), text.Spans[2].Style.Foreground)
	require.Equal(t, " · ", text.Spans[3].Text)
	require.Equal(t, app.taskLists[2].Name, text.Spans[4].Text)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), text.Spans[4].Style.Foreground)
}

func TestListSwitcherSpans_UpdatesWhenActiveListChanges(t *testing.T) {
	app := NewTodoApp()
	theme := terma.ThemeData{
		Primary:   terma.Hex("#ff00aa"),
		TextMuted: terma.Hex("#778899"),
	}

	initial := app.listSwitcherSpans(theme)
	require.Equal(t, theme.Primary, initial[0].Style.Foreground)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), initial[2].Style.Foreground)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), initial[4].Style.Foreground)

	app.switchToNextList()

	updated := app.listSwitcherSpans(theme)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), updated[0].Style.Foreground)
	require.Equal(t, theme.Primary, updated[2].Style.Foreground)
	require.Equal(t, theme.TextMuted.WithAlpha(0.6), updated[4].Style.Foreground)
}

func TestBuildMainContainer_NormalBorderOmitsActiveListNameTitle(t *testing.T) {
	app := NewTodoApp()
	theme, ok := terma.GetTheme(terma.CurrentThemeName())
	require.True(t, ok)

	widget := app.buildMainContainer(terma.BuildContext{}, theme.Background)
	column, ok := widget.(terma.Column)
	require.True(t, ok)

	decorations := column.Style.Border.Decorations
	require.NotEmpty(t, decorations)
	for _, decoration := range decorations {
		require.NotEqual(t, app.activeList().Name, decoration.Text)
	}
}

func TestBuild_UsesReducedBottomPaddingForLowerFooter(t *testing.T) {
	app := NewTodoApp()

	widget := app.Build(terma.BuildContext{})
	column, ok := widget.(terma.Column)
	require.True(t, ok)
	require.Equal(t, 2, column.Style.Padding.Top)
	require.Equal(t, 1, column.Style.Padding.Bottom)
	require.Equal(t, 6, column.Style.Padding.Left)
	require.Equal(t, 6, column.Style.Padding.Right)
}

func TestEditInput_BlurSavesAndLeavesEditMode(t *testing.T) {
	app := NewTodoApp()
	app.editingIndex.Set(0)
	app.editInputState.SetText("  Edited title  ")

	task := app.activeList().Tasks.GetItems()[0]
	renderItem := app.renderTaskItem(terma.BuildContext{}, false)

	widget := renderItem(task, true, false)
	row, ok := widget.(terma.Row)
	require.True(t, ok)
	require.Len(t, row.Children, 2)

	autocomplete, ok := row.Children[1].(terma.Autocomplete)
	require.True(t, ok)

	textArea, ok := autocomplete.Child.(terma.TextArea)
	require.True(t, ok)
	require.NotNil(t, textArea.Blur)

	textArea.Blur()
	require.Equal(t, -1, app.editingIndex.Get())
	require.Equal(t, "Edited title", app.activeList().Tasks.GetItems()[0].Title)
}

func TestKeybinds_EscapeClearsSelectionWhenPresent(t *testing.T) {
	app := NewTodoApp()
	app.activeList().Tasks.Select(1)
	app.activeList().Tasks.Select(2)

	keybind, ok := findKeybindByKey(app.Keybinds(), "escape")
	require.True(t, ok)
	require.Equal(t, "Clear", keybind.Name)
	require.NotNil(t, keybind.Action)

	keybind.Action()
	require.Empty(t, app.activeList().Tasks.SelectedItems())
	require.False(t, app.activeList().Tasks.HasAnchor())
}

func TestKeybinds_EscapeOmittedWhenNoSelection(t *testing.T) {
	app := NewTodoApp()

	_, ok := findKeybindByKey(app.Keybinds(), "escape")
	require.False(t, ok)
}

func TestKeybinds_IncludeListAliasesAndJumpKeys(t *testing.T) {
	app := NewTodoApp()

	keybind, ok := findKeybindByKey(app.Keybinds(), "h")
	require.True(t, ok)
	require.NotNil(t, keybind.Action)

	keybind, ok = findKeybindByKey(app.Keybinds(), "l")
	require.True(t, ok)
	require.NotNil(t, keybind.Action)

	keybind, ok = findKeybindByKey(app.Keybinds(), "3")
	require.True(t, ok)
	require.NotNil(t, keybind.Action)

	keybind.Action()
	require.Equal(t, archiveListID, app.activeList().ID)

	keybind, ok = findKeybindByKey(app.Keybinds(), "y")
	require.True(t, ok)
	require.Equal(t, "Copy", keybind.Name)

	app.activeListIdx.Set(2)
	keybind, ok = findKeybindByKey(app.Keybinds(), "D")
	require.True(t, ok)
	require.Equal(t, "Delete", keybind.Name)
}

func TestMoveSelectedTasksRightAndLeft_UsesAdjacentLists(t *testing.T) {
	app := NewTodoApp()
	task := app.activeList().Tasks.GetItems()[0]
	app.activeList().Tasks.SelectIndex(0)

	app.moveSelectedTasksRight()
	require.Equal(t, task.ID, app.taskLists[1].Tasks.GetItems()[0].ID)

	app.activeListIdx.Set(1)
	app.activeList().Tasks.SelectIndex(0)
	app.moveSelectedTasksRight()
	require.Equal(t, task.ID, app.archiveList().Tasks.GetItems()[0].ID)

	app.activeListIdx.Set(2)
	app.activeList().Tasks.SelectIndex(0)
	app.moveSelectedTasksLeft()
	require.Equal(t, task.ID, app.taskLists[1].Tasks.GetItems()[0].ID)
}

func TestArchiveCurrentTask_MovesTaskToArchive(t *testing.T) {
	app := NewTodoApp()
	task := app.activeList().Tasks.GetItems()[0]
	app.activeList().Tasks.SelectIndex(0)

	app.archiveCurrentTask()

	require.NotContains(t, collectTaskIDs(app.taskLists[0].Tasks.GetItems()), task.ID)
	require.Equal(t, task.ID, app.archiveList().Tasks.GetItems()[0].ID)
}

func TestPermanentlyDeleteCurrentTask_RemovesTaskFromArchive(t *testing.T) {
	app := NewTodoApp()
	task := app.activeList().Tasks.GetItems()[0]
	app.activeList().Tasks.SelectIndex(0)
	app.archiveCurrentTask()

	app.activeListIdx.Set(2)
	app.activeList().Tasks.SelectIndex(0)
	app.permanentlyDeleteCurrentTask()

	require.NotContains(t, collectTaskIDs(app.archiveList().Tasks.GetItems()), task.ID)
}

func TestTasksToMarkdownChecklist_FormatsSelection(t *testing.T) {
	app := NewTodoApp()
	tasks := []Task{
		app.activeList().Tasks.GetItems()[0],
		app.activeList().Tasks.GetItems()[2],
	}

	require.Equal(t, "- [ ] Invent a new color #creative #fun\n- [x] Find out who let the dogs out #pets #mystery", tasksToMarkdownChecklist(tasks))
}

func TestBuildHelpModal_UsesAutoWidthForPopupAndColumns(t *testing.T) {
	app := NewTodoApp()
	widget := app.buildHelpModal(terma.ThemeData{})

	floating, ok := widget.(terma.Floating)
	require.True(t, ok)

	column, ok := floating.Child.(terma.Column)
	require.True(t, ok)
	require.True(t, column.Width.IsUnset())

	contentRow, ok := column.Children[1].(terma.Row)
	require.True(t, ok)
	require.Len(t, contentRow.Children, 3)

	for _, child := range contentRow.Children {
		helpColumn, ok := child.(terma.Column)
		require.True(t, ok)
		require.True(t, helpColumn.Width.IsUnset())
	}
}

func collectTaskIDs(tasks []Task) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func findKeybindByKey(keybinds []terma.Keybind, key string) (terma.Keybind, bool) {
	for _, keybind := range keybinds {
		if keybind.Key == key {
			return keybind, true
		}
	}
	return terma.Keybind{}, false
}

type testFocusable struct {
	id string
}

func (f testFocusable) WidgetID() string                          { return f.id }
func (f testFocusable) Build(ctx terma.BuildContext) terma.Widget { return terma.Text{} }
func (f testFocusable) OnKey(event terma.KeyEvent) bool           { return false }
func (f testFocusable) IsFocusable() bool                         { return true }
