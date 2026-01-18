package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTabState(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}

	state := NewTabState(tabs)

	assert.Equal(t, 2, state.TabCount())
	assert.Equal(t, "home", state.ActiveKey())
	assert.Equal(t, 0, state.ActiveIndex())
}

func TestNewTabState_Empty(t *testing.T) {
	state := NewTabState(nil)

	assert.Equal(t, 0, state.TabCount())
	assert.Equal(t, "", state.ActiveKey())
	assert.Equal(t, -1, state.ActiveIndex())
}

func TestNewTabStateWithActive(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}

	state := NewTabStateWithActive(tabs, "settings")

	assert.Equal(t, "settings", state.ActiveKey())
	assert.Equal(t, 1, state.ActiveIndex())
}

func TestTabState_SelectNext(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	assert.Equal(t, "a", state.ActiveKey())

	state.SelectNext()
	assert.Equal(t, "b", state.ActiveKey())

	state.SelectNext()
	assert.Equal(t, "c", state.ActiveKey())

	// Should wrap around
	state.SelectNext()
	assert.Equal(t, "a", state.ActiveKey())
}

func TestTabState_SelectPrevious(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	// Should wrap around to end
	state.SelectPrevious()
	assert.Equal(t, "c", state.ActiveKey())

	state.SelectPrevious()
	assert.Equal(t, "b", state.ActiveKey())

	state.SelectPrevious()
	assert.Equal(t, "a", state.ActiveKey())
}

func TestTabState_SelectIndex(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	state.SelectIndex(2)
	assert.Equal(t, "c", state.ActiveKey())

	state.SelectIndex(0)
	assert.Equal(t, "a", state.ActiveKey())

	// Out of bounds should do nothing
	state.SelectIndex(-1)
	assert.Equal(t, "a", state.ActiveKey())

	state.SelectIndex(100)
	assert.Equal(t, "a", state.ActiveKey())
}

func TestTabState_AddTab(t *testing.T) {
	state := NewTabState(nil)
	assert.Equal(t, 0, state.TabCount())

	state.AddTab(Tab{Key: "first", Label: "First"})
	assert.Equal(t, 1, state.TabCount())
	assert.Equal(t, "first", state.ActiveKey())

	state.AddTab(Tab{Key: "second", Label: "Second"})
	assert.Equal(t, 2, state.TabCount())
	assert.Equal(t, "first", state.ActiveKey()) // First tab should still be active

	tabs := state.TabsPeek()
	assert.Equal(t, "first", tabs[0].Key)
	assert.Equal(t, "second", tabs[1].Key)
}

func TestTabState_InsertTab(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	state.InsertTab(1, Tab{Key: "b", Label: "B"})
	assert.Equal(t, 3, state.TabCount())

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
	assert.Equal(t, "b", result[1].Key)
	assert.Equal(t, "c", result[2].Key)
}

func TestTabState_InsertTab_AtStart(t *testing.T) {
	tabs := []Tab{
		{Key: "b", Label: "B"},
	}
	state := NewTabState(tabs)

	state.InsertTab(0, Tab{Key: "a", Label: "A"})

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
	assert.Equal(t, "b", result[1].Key)
}

func TestTabState_InsertTab_AtEnd(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	state.InsertTab(10, Tab{Key: "z", Label: "Z"})

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
	assert.Equal(t, "z", result[1].Key)
}

func TestTabState_RemoveTab(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	removed := state.RemoveTab("b")
	assert.True(t, removed)
	assert.Equal(t, 2, state.TabCount())

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
	assert.Equal(t, "c", result[1].Key)
}

func TestTabState_RemoveTab_NotFound(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	removed := state.RemoveTab("nonexistent")
	assert.False(t, removed)
	assert.Equal(t, 1, state.TabCount())
}

func TestTabState_RemoveTab_Active(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabStateWithActive(tabs, "b")

	// Removing active tab should switch to next
	state.RemoveTab("b")
	assert.Equal(t, "c", state.ActiveKey())
}

func TestTabState_RemoveTab_LastActive(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	state := NewTabStateWithActive(tabs, "b")

	// Removing last active tab should switch to previous
	state.RemoveTab("b")
	assert.Equal(t, "a", state.ActiveKey())
}

func TestTabState_RemoveTab_OnlyTab(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	state.RemoveTab("a")
	assert.Equal(t, "", state.ActiveKey())
	assert.Equal(t, 0, state.TabCount())
}

func TestTabState_MoveTabLeft(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	moved := state.MoveTabLeft("b")
	assert.True(t, moved)

	result := state.TabsPeek()
	assert.Equal(t, "b", result[0].Key)
	assert.Equal(t, "a", result[1].Key)
	assert.Equal(t, "c", result[2].Key)
}

func TestTabState_MoveTabLeft_AtStart(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	state := NewTabState(tabs)

	moved := state.MoveTabLeft("a")
	assert.False(t, moved)

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
}

func TestTabState_MoveTabRight(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C"},
	}
	state := NewTabState(tabs)

	moved := state.MoveTabRight("b")
	assert.True(t, moved)

	result := state.TabsPeek()
	assert.Equal(t, "a", result[0].Key)
	assert.Equal(t, "c", result[1].Key)
	assert.Equal(t, "b", result[2].Key)
}

func TestTabState_MoveTabRight_AtEnd(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	state := NewTabState(tabs)

	moved := state.MoveTabRight("b")
	assert.False(t, moved)
}

func TestTabState_SetLabel(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
	}
	state := NewTabState(tabs)

	state.SetLabel("home", "Dashboard")

	result := state.TabsPeek()
	assert.Equal(t, "Dashboard", result[0].Label)
}

func TestTabState_Editing(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
	}
	state := NewTabState(tabs)

	assert.False(t, state.IsEditing("home"))
	assert.Equal(t, "", state.EditingKey())

	state.StartEditing("home")
	assert.True(t, state.IsEditing("home"))
	assert.Equal(t, "home", state.EditingKey())

	state.StopEditing()
	assert.False(t, state.IsEditing("home"))
	assert.Equal(t, "", state.EditingKey())
}

func TestTabState_ActiveTab(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	state := NewTabState(tabs)

	tab := state.ActiveTab()
	assert.NotNil(t, tab)
	assert.Equal(t, "a", tab.Key)

	state.SetActiveKey("b")
	tab = state.ActiveTab()
	assert.NotNil(t, tab)
	assert.Equal(t, "b", tab.Key)
}

func TestTabState_ActiveTab_NotFound(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabStateWithActive(tabs, "nonexistent")

	tab := state.ActiveTab()
	assert.Nil(t, tab)
}

func TestTabBar_Keybinds(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	state := NewTabState(tabs)

	tabBar := TabBar{
		ID:             "tabs",
		State:          state,
		KeybindPattern: TabKeybindNumbers,
	}

	keybinds := tabBar.Keybinds()

	// Should have left, right, and number keybinds
	assert.True(t, len(keybinds) >= 4)

	// Check that navigation keys are present
	var hasLeft, hasRight, hasH, hasL bool
	for _, kb := range keybinds {
		if kb.Key == "left" {
			hasLeft = true
		}
		if kb.Key == "right" {
			hasRight = true
		}
		if kb.Key == "h" {
			hasH = true
		}
		if kb.Key == "l" {
			hasL = true
		}
	}
	assert.True(t, hasLeft)
	assert.True(t, hasRight)
	assert.True(t, hasH)
	assert.True(t, hasL)

	// Check that number keybinds are present
	var has1, has2 bool
	for _, kb := range keybinds {
		if kb.Key == "1" {
			has1 = true
		}
		if kb.Key == "2" {
			has2 = true
		}
	}
	assert.True(t, has1)
	assert.True(t, has2)
}

func TestTabBar_Keybinds_AltNumbers(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	tabBar := TabBar{
		ID:             "tabs",
		State:          state,
		KeybindPattern: TabKeybindAltNumbers,
	}

	keybinds := tabBar.Keybinds()

	var hasAlt1 bool
	for _, kb := range keybinds {
		if kb.Key == "alt+1" {
			hasAlt1 = true
		}
	}
	assert.True(t, hasAlt1)
}

func TestTabBar_Keybinds_CtrlNumbers(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	tabBar := TabBar{
		ID:             "tabs",
		State:          state,
		KeybindPattern: TabKeybindCtrlNumbers,
	}

	keybinds := tabBar.Keybinds()

	var hasCtrl1 bool
	for _, kb := range keybinds {
		if kb.Key == "ctrl+1" {
			hasCtrl1 = true
		}
	}
	assert.True(t, hasCtrl1)
}

func TestTabBar_Keybinds_Reorder(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	tabBar := TabBar{
		ID:           "tabs",
		State:        state,
		AllowReorder: true,
	}

	keybinds := tabBar.Keybinds()

	var hasCtrlH, hasCtrlL bool
	for _, kb := range keybinds {
		if kb.Key == "ctrl+h" {
			hasCtrlH = true
		}
		if kb.Key == "ctrl+l" {
			hasCtrlL = true
		}
	}
	assert.True(t, hasCtrlH)
	assert.True(t, hasCtrlL)
}

func TestTabBar_Keybinds_NoReorder(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "A"},
	}
	state := NewTabState(tabs)

	tabBar := TabBar{
		ID:           "tabs",
		State:        state,
		AllowReorder: false,
	}

	keybinds := tabBar.Keybinds()

	var hasCtrlH, hasCtrlL bool
	for _, kb := range keybinds {
		if kb.Key == "ctrl+h" {
			hasCtrlH = true
		}
		if kb.Key == "ctrl+l" {
			hasCtrlL = true
		}
	}
	assert.False(t, hasCtrlH)
	assert.False(t, hasCtrlL)
}

func TestTabBar_NilState(t *testing.T) {
	tabBar := TabBar{
		ID:    "tabs",
		State: nil,
	}

	// Should not panic
	keybinds := tabBar.Keybinds()
	assert.Nil(t, keybinds)
}

func TestTabBar_IsFocusable(t *testing.T) {
	tabBar := TabBar{ID: "tabs"}
	assert.True(t, tabBar.IsFocusable())
}

func TestTabKeybindPattern_Values(t *testing.T) {
	assert.Equal(t, TabKeybindPattern(0), TabKeybindNone)
	assert.Equal(t, TabKeybindPattern(1), TabKeybindNumbers)
	assert.Equal(t, TabKeybindPattern(2), TabKeybindAltNumbers)
	assert.Equal(t, TabKeybindPattern(3), TabKeybindCtrlNumbers)
}

func TestTabState_SelectNext_Empty(t *testing.T) {
	state := NewTabState(nil)
	// Should not panic
	state.SelectNext()
	assert.Equal(t, "", state.ActiveKey())
}

func TestTabState_SelectPrevious_Empty(t *testing.T) {
	state := NewTabState(nil)
	// Should not panic
	state.SelectPrevious()
	assert.Equal(t, "", state.ActiveKey())
}
