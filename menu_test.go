package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMenu_NavigationWraps(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{Label: "First"},
		{Label: "Second"},
		{Label: "Third"},
	})
	menu := Menu{ID: "menu", State: state}

	state.SetCursorIndex(2)
	menu.moveNext()
	assert.Equal(t, 0, state.CursorIndex())

	menu.movePrevious()
	assert.Equal(t, 2, state.CursorIndex())
}

func TestMenu_NavigationSkipsDisabledAndDividers(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{Label: "Alpha"},
		{Divider: "Group"},
		{Label: "Beta", Disabled: true},
		{Label: "Gamma"},
	})
	menu := Menu{ID: "menu", State: state}

	state.SetCursorIndex(0)
	menu.moveNext()
	assert.Equal(t, 3, state.CursorIndex())

	menu.movePrevious()
	assert.Equal(t, 0, state.CursorIndex())
}

func TestMenu_SubmenuOpenClose(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{Label: "File", Children: []MenuItem{{Label: "New"}}},
		{Label: "Edit"},
	})
	menu := Menu{ID: "menu", State: state}

	state.SetCursorIndex(0)
	menu.openSubmenu()

	assert.True(t, state.HasOpenSubmenu())
	assert.Equal(t, 0, state.openSubmenu.Peek())
	assert.NotNil(t, state.submenuState)
	assert.Equal(t, 1, len(state.submenuState.Items()))

	menu.closeSubmenu()
	assert.False(t, state.HasOpenSubmenu())
}

func TestMenu_NestedSubmenuState(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{
			Label: "File",
			Children: []MenuItem{
				{Label: "Recent", Children: []MenuItem{{Label: "a.txt"}, {Label: "b.txt"}}},
			},
		},
	})
	menu := Menu{ID: "menu", State: state}

	state.SetCursorIndex(0)
	menu.openSubmenu()

	submenu := state.submenuState
	assert.NotNil(t, submenu)

	submenu.SetCursorIndex(0)
	Menu{ID: "menu-sub", State: submenu}.openSubmenu()

	assert.True(t, submenu.HasOpenSubmenu())
	assert.NotNil(t, submenu.submenuState)
}

func TestSnapshot_Menu_Basic(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{Label: "Open", Shortcut: "Ctrl+O"},
		{Label: "Save As", Shortcut: "Ctrl+Shift+S"},
		{Divider: "Export"},
		{Label: "Export", Disabled: true},
		{Label: "Recent", Children: []MenuItem{{Label: "one.txt"}, {Label: "two.txt"}}},
	})
	state.SetCursorIndex(0)

	widget := Menu{
		ID:    "menu-basic",
		State: state,
	}

	AssertSnapshot(t, widget, 40, 10,
		"Menu with active 'Open' item, shortcut alignment, titled divider, disabled 'Export', and 'Recent' submenu indicator.")
}

func TestSnapshot_Menu_Submenu(t *testing.T) {
	state := NewMenuState([]MenuItem{
		{Label: "File", Children: []MenuItem{{Label: "New"}, {Label: "Open"}}},
		{Label: "Edit", Children: []MenuItem{{Label: "Cut"}, {Label: "Copy"}, {Label: "Paste"}}},
		{Label: "Help"},
	})
	state.SetCursorIndex(1)
	state.OpenSubmenu(1)

	widget := Menu{
		ID:    "menu-sub",
		State: state,
	}

	AssertSnapshot(t, widget, 60, 10,
		"Menu with 'Edit' submenu open to the right. Parent menu shows active item and submenu arrow.")
}
