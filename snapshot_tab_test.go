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
