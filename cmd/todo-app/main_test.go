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

func TestBuildInputRow_NewTaskInputIncludesEscapeToTaskList(t *testing.T) {
	app := NewTodoApp()

	widget := app.buildInputRow(terma.ThemeData{})
	row, ok := widget.(terma.Row)
	require.True(t, ok)
	require.Len(t, row.Children, 2)

	autocomplete, ok := row.Children[1].(terma.Autocomplete)
	require.True(t, ok)

	input, ok := autocomplete.Child.(terma.TextInput)
	require.True(t, ok)

	keybind, ok := findKeybindByKey(input.ExtraKeybinds, "escape")
	require.True(t, ok)
	require.Equal(t, "Tasks", keybind.Name)
	require.NotNil(t, keybind.Action)
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
