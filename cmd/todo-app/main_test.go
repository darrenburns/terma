package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	terma "github.com/darrenburns/terma"
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

func collectTaskIDs(tasks []Task) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}
