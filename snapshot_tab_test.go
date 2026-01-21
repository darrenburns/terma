package terma

import (
	"testing"
)

// =============================================================================
// TabBar Widget Tests
// =============================================================================

func TestSnapshot_TabBar_Basic(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"Three tabs in a row. 'Home' is active (highlighted), 'Settings' and 'Profile' inactive. Each tab has padding.")
}

func TestSnapshot_TabBar_SecondActive(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "settings")

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"Three tabs with 'Settings' active (highlighted). 'Home' and 'Profile' inactive.")
}

func TestSnapshot_TabBar_LastActive(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "profile")

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"Three tabs with 'Profile' active (highlighted). 'Home' and 'Settings' inactive.")
}

func TestSnapshot_TabBar_SingleTab(t *testing.T) {
	tabs := []Tab{
		{Key: "only", Label: "Only Tab"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 3,
		"Single tab 'Only Tab' displayed as active.")
}

func TestSnapshot_TabBar_Closable(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:       "tabs",
		State:    state,
		Closable: true,
	}
	AssertSnapshot(t, widget, 40, 3,
		"Two tabs with close buttons (×). 'Home ×' active, 'Settings ×' inactive.")
}

func TestSnapshot_TabBar_CustomStyle(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "Tab A"},
		{Key: "b", Label: "Tab B"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:    "tabs",
		State: state,
		TabStyle: Style{
			ForegroundColor: RGB(150, 150, 150),
			BackgroundColor: RGB(40, 40, 40),
		},
		ActiveTabStyle: Style{
			ForegroundColor: RGB(255, 255, 255),
			BackgroundColor: RGB(80, 120, 200),
		},
	}
	AssertSnapshot(t, widget, 40, 3,
		"Two tabs with custom colors. Active 'Tab A' blue background, inactive 'Tab B' dark gray.")
}

func TestSnapshot_TabBar_WithContainerStyle(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:    "tabs",
		State: state,
		Style: Style{
			BackgroundColor: RGB(30, 30, 30),
		},
	}
	AssertSnapshot(t, widget, 40, 3,
		"Tab bar with dark background. Two tabs on dark gray container.")
}

func TestSnapshot_TabBar_ManyTabs(t *testing.T) {
	tabs := []Tab{
		{Key: "1", Label: "Tab 1"},
		{Key: "2", Label: "Tab 2"},
		{Key: "3", Label: "Tab 3"},
		{Key: "4", Label: "Tab 4"},
		{Key: "5", Label: "Tab 5"},
	}
	state := NewTabState(tabs)

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 60, 3,
		"Five tabs in a row. 'Tab 1' active, others inactive. Tabs extend horizontally.")
}

func TestSnapshot_TabBar_Empty(t *testing.T) {
	state := NewTabState(nil)

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 3,
		"Empty tab bar with no tabs rendered.")
}

func TestSnapshot_TabBar_NilState(t *testing.T) {
	widget := TabBar{
		ID:    "tabs",
		State: nil,
	}
	AssertSnapshot(t, widget, 30, 3,
		"Tab bar with nil state renders as empty row.")
}

// =============================================================================
// TabView Widget Tests
// =============================================================================

func TestSnapshot_TabView_Basic(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home", Content: Text{Content: "Home content goes here"}},
		{Key: "settings", Label: "Settings", Content: Text{Content: "Settings content"}},
	}
	state := NewTabState(tabs)

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 40, 10,
		"Tab view with 'Home' tab active. Tab bar at top, 'Home content goes here' below.")
}

func TestSnapshot_TabView_SecondTabActive(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home", Content: Text{Content: "Home content"}},
		{Key: "settings", Label: "Settings", Content: Text{Content: "Settings panel with options"}},
	}
	state := NewTabStateWithActive(tabs, "settings")

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 45, 10,
		"Tab view with 'Settings' active. Shows 'Settings panel with options' content.")
}

func TestSnapshot_TabView_WithComplexContent(t *testing.T) {
	tabs := []Tab{
		{
			Key:   "list",
			Label: "List",
			Content: Column{
				Children: []Widget{
					Text{Content: "Item 1", Style: Style{ForegroundColor: RGB(100, 200, 100)}},
					Text{Content: "Item 2", Style: Style{ForegroundColor: RGB(100, 200, 100)}},
					Text{Content: "Item 3", Style: Style{ForegroundColor: RGB(100, 200, 100)}},
				},
			},
		},
		{Key: "empty", Label: "Empty"},
	}
	state := NewTabState(tabs)

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 40, 10,
		"Tab view with 'List' tab showing green items stacked vertically in content area.")
}

func TestSnapshot_TabView_Closable(t *testing.T) {
	tabs := []Tab{
		{Key: "file1", Label: "file.go", Content: Text{Content: "package main"}},
		{Key: "file2", Label: "test.go", Content: Text{Content: "package main_test"}},
	}
	state := NewTabState(tabs)

	widget := TabView{
		ID:       "tabview",
		State:    state,
		Closable: true,
		Height:   Cells(8),
	}
	AssertSnapshot(t, widget, 45, 8,
		"Tab view with closable tabs. 'file.go ×' active with code content, 'test.go ×' inactive.")
}

func TestSnapshot_TabView_CustomStyles(t *testing.T) {
	tabs := []Tab{
		{Key: "a", Label: "Tab A", Content: Text{Content: "Content A"}},
		{Key: "b", Label: "Tab B", Content: Text{Content: "Content B"}},
	}
	state := NewTabState(tabs)

	widget := TabView{
		ID:    "tabview",
		State: state,
		Style: Style{
			BackgroundColor: RGB(25, 25, 25),
		},
		TabBarStyle: Style{
			BackgroundColor: RGB(40, 40, 40),
		},
		ContentStyle: Style{
			BackgroundColor: RGB(30, 30, 30),
		},
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 40, 10,
		"Tab view with custom dark theme. Tab bar slightly lighter, content area dark.")
}

func TestSnapshot_TabView_Empty(t *testing.T) {
	state := NewTabState(nil)

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(8),
	}
	AssertSnapshot(t, widget, 30, 8,
		"Empty tab view with no tabs. Just an empty column.")
}

func TestSnapshot_TabView_NilState(t *testing.T) {
	widget := TabView{
		ID:     "tabview",
		State:  nil,
		Height: Cells(8),
	}
	AssertSnapshot(t, widget, 30, 8,
		"Tab view with nil state renders as empty column.")
}

func TestSnapshot_TabView_NilContent(t *testing.T) {
	tabs := []Tab{
		{Key: "no-content", Label: "No Content", Content: nil},
	}
	state := NewTabState(tabs)

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(8),
	}
	AssertSnapshot(t, widget, 35, 8,
		"Tab view where active tab has nil content. Shows tab bar but empty content area.")
}

// =============================================================================
// TabBar in Layout Context
// =============================================================================

func TestSnapshot_TabBar_InDock(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "about", Label: "About"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:    "tabs",
				State: state,
				Style: Style{BackgroundColor: RGB(40, 40, 40)},
			},
		},
		Body: Text{Content: "Main content area", Style: Style{BackgroundColor: RGB(30, 30, 30)}},
	}
	AssertSnapshot(t, widget, 40, 10,
		"Dock layout with TabBar docked at top. Tab bar dark gray, body darker below.")
}

func TestSnapshot_TabBar_WithKeybindBar(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:             "tabs",
				State:          state,
				KeybindPattern: TabKeybindNumbers,
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Text{Content: "Content"},
	}
	AssertSnapshot(t, widget, 50, 10,
		"Dock with TabBar at top, KeybindBar at bottom. Shows tab navigation keybinds in footer.")
}

// =============================================================================
// Navigation Wrap-Around Tests
// =============================================================================

func TestSnapshot_TabBar_NavigationWrapToFirst(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "profile") // Start on last tab
	state.SelectNext()                              // Should wrap to first

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After SelectNext on last tab, first tab 'Home' is now active. Demonstrates wrap-around navigation.")
}

func TestSnapshot_TabBar_NavigationWrapToLast(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabState(tabs)  // Start on first tab
	state.SelectPrevious()      // Should wrap to last

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After SelectPrevious on first tab, last tab 'Profile' is now active. Demonstrates wrap-around navigation.")
}

// =============================================================================
// Active Tab Removal Tests
// =============================================================================

func TestSnapshot_TabBar_RemoveActiveTab_ShiftsToNext(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "settings") // Middle tab active
	state.RemoveTab("settings")                       // Remove active tab

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After removing middle active tab 'Settings', 'Profile' (next tab) becomes active. Two tabs remain.")
}

func TestSnapshot_TabBar_RemoveActiveTab_ShiftsToPrevious(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "profile") // Last tab active
	state.RemoveTab("profile")                       // Remove active tab

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After removing last active tab 'Profile', 'Settings' (previous tab) becomes active. Two tabs remain.")
}

func TestSnapshot_TabBar_RemoveOnlyTab(t *testing.T) {
	tabs := []Tab{
		{Key: "only", Label: "Only Tab"},
	}
	state := NewTabState(tabs)
	state.RemoveTab("only")

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 3,
		"After removing the only tab, tab bar is empty with no tabs rendered.")
}

// =============================================================================
// Tab Reordering Tests
// =============================================================================

func TestSnapshot_TabBar_AfterMoveTabLeft(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "settings")
	state.MoveTabLeft("settings") // Move Settings before Home

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After MoveTabLeft, order is 'Settings' (active), 'Home', 'Profile'. Settings moved from middle to first.")
}

func TestSnapshot_TabBar_AfterMoveTabRight(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
		{Key: "profile", Label: "Profile"},
	}
	state := NewTabStateWithActive(tabs, "settings")
	state.MoveTabRight("settings") // Move Settings after Profile

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After MoveTabRight, order is 'Home', 'Profile', 'Settings' (active). Settings moved from middle to last.")
}

// =============================================================================
// Dynamic Tab Management Tests
// =============================================================================

func TestSnapshot_TabBar_AfterAddTab(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)
	state.AddTab(Tab{Key: "new", Label: "New Tab"})

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 50, 3,
		"After AddTab, three tabs shown: 'Home' (active), 'Settings', 'New Tab'. New tab appended at end.")
}

func TestSnapshot_TabBar_AfterInsertTabAtStart(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)
	state.InsertTab(0, Tab{Key: "first", Label: "First"})

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 50, 3,
		"After InsertTab at index 0, order is 'First', 'Home' (active), 'Settings'. New tab inserted at start.")
}

func TestSnapshot_TabBar_AfterInsertTabInMiddle(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)
	state.InsertTab(1, Tab{Key: "middle", Label: "Middle"})

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 50, 3,
		"After InsertTab at index 1, order is 'Home' (active), 'Middle', 'Settings'. New tab inserted in middle.")
}

func TestSnapshot_TabBar_AddTabToEmpty(t *testing.T) {
	state := NewTabState(nil)
	state.AddTab(Tab{Key: "first", Label: "First Tab"})

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 3,
		"After adding tab to empty state, 'First Tab' is shown and automatically becomes active.")
}

// =============================================================================
// Label Update Test
// =============================================================================

func TestSnapshot_TabBar_AfterSetLabel(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)
	state.SetLabel("home", "Dashboard")

	widget := TabBar{
		ID:    "tabs",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 3,
		"After SetLabel, first tab shows 'Dashboard' (active) instead of 'Home'. Second tab 'Settings' unchanged.")
}

// =============================================================================
// KeybindBar Integration Tests
// =============================================================================

func TestSnapshot_TabBar_KeybindBar_WithClosable(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:       "tabs",
				State:    state,
				Closable: true,
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Text{Content: "Content"},
	}
	AssertSnapshot(t, widget, 50, 10,
		"TabBar with Closable=true. KeybindBar shows h/l navigation keybinds. Tabs have close buttons (×).")
}

func TestSnapshot_TabBar_KeybindBar_WithAllowReorder(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:           "tabs",
				State:        state,
				AllowReorder: true,
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Text{Content: "Content"},
	}
	AssertSnapshot(t, widget, 60, 10,
		"TabBar with AllowReorder=true. KeybindBar shows ctrl+h 'Move Left', ctrl+l 'Move Right' in addition to navigation.")
}

func TestSnapshot_TabBar_KeybindBar_WithAltNumbers(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:             "tabs",
				State:          state,
				KeybindPattern: TabKeybindAltNumbers,
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Text{Content: "Content"},
	}
	AssertSnapshot(t, widget, 50, 10,
		"TabBar with Alt+Numbers pattern. KeybindBar shows standard h/l navigation (position keybinds are hidden).")
}

func TestSnapshot_TabBar_KeybindBar_WithCtrlNumbers(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home"},
		{Key: "settings", Label: "Settings"},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabBar{
				ID:             "tabs",
				State:          state,
				KeybindPattern: TabKeybindCtrlNumbers,
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Text{Content: "Content"},
	}
	AssertSnapshot(t, widget, 50, 10,
		"TabBar with Ctrl+Numbers pattern. KeybindBar shows standard h/l navigation (position keybinds are hidden).")
}

// =============================================================================
// TabView Specific Tests
// =============================================================================

func TestSnapshot_TabView_AfterTabSwitch(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home", Content: Text{Content: "Home content here"}},
		{Key: "settings", Label: "Settings", Content: Text{Content: "Settings content here"}},
	}
	state := NewTabState(tabs)
	state.SelectNext() // Switch from Home to Settings

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 45, 10,
		"After SelectNext, TabView shows 'Settings' tab active with 'Settings content here' displayed below.")
}

func TestSnapshot_TabView_ContentPreservedAcrossSwitch(t *testing.T) {
	tabs := []Tab{
		{Key: "home", Label: "Home", Content: Column{
			Children: []Widget{
				Text{Content: "Line 1"},
				Text{Content: "Line 2"},
				Text{Content: "Line 3"},
			},
		}},
		{Key: "other", Label: "Other", Content: Text{Content: "Other content"}},
	}
	state := NewTabState(tabs)
	// Switch away and back
	state.SelectNext() // To "other"
	state.SelectPrevious() // Back to "home"

	widget := TabView{
		ID:     "tabview",
		State:  state,
		Height: Cells(10),
	}
	AssertSnapshot(t, widget, 40, 10,
		"After switching away and back, 'Home' tab content (3 lines) is still displayed correctly.")
}

func TestSnapshot_TabView_WithClosableAndReorder(t *testing.T) {
	tabs := []Tab{
		{Key: "file1", Label: "main.go", Content: Text{Content: "package main"}},
		{Key: "file2", Label: "test.go", Content: Text{Content: "package main_test"}},
		{Key: "file3", Label: "util.go", Content: Text{Content: "package util"}},
	}
	state := NewTabState(tabs)

	widget := Dock{
		Top: []Widget{
			TabView{
				ID:           "tabview",
				State:        state,
				Closable:     true,
				AllowReorder: true,
				Height:       Cells(6),
			},
		},
		Bottom: []Widget{
			KeybindBar{},
		},
		Body: Spacer{},
	}
	AssertSnapshot(t, widget, 60, 12,
		"TabView with Closable and AllowReorder. Shows tabs with × buttons and reorder keybinds in KeybindBar.")
}
