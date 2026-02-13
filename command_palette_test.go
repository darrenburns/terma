package terma

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCommandPaletteState_PushPop(t *testing.T) {
	state := NewCommandPaletteState("Root", []CommandPaletteItem{
		{Label: "Open"},
	})

	if state.IsNested() {
		t.Fatalf("expected root to be non-nested")
	}
	if got := state.BreadcrumbPath(); !reflect.DeepEqual(got, []string{"Root"}) {
		t.Fatalf("unexpected breadcrumb path: %#v", got)
	}

	state.PushLevel("Child", []CommandPaletteItem{
		{Label: "Alpha"},
	})

	if !state.IsNested() {
		t.Fatalf("expected nested state after push")
	}
	if got := state.BreadcrumbPath(); !reflect.DeepEqual(got, []string{"Root", "Child"}) {
		t.Fatalf("unexpected breadcrumb path after push: %#v", got)
	}

	if !state.PopLevel() {
		t.Fatalf("expected pop to succeed")
	}
	if state.PopLevel() {
		t.Fatalf("expected pop to fail at root")
	}
}

func TestCommandPaletteState_CurrentItemSkipsDividers(t *testing.T) {
	state := NewCommandPaletteState("Root", []CommandPaletteItem{
		{Divider: "Group"},
		{Label: "Disabled", Disabled: true},
		{Label: "Enabled"},
	})

	item, ok := state.CurrentItem()
	if !ok {
		t.Fatalf("expected selectable current item")
	}
	if item.Label != "Enabled" {
		t.Fatalf("unexpected item selected: %q", item.Label)
	}
}

func TestCommandPaletteState_CloseUsesNextFocusOverride(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Open"},
	})
	palette := CommandPalette{
		ID:    "palette",
		State: state,
	}

	// Simulate a previously-visible palette closing.
	state.wasVisible = true
	state.lastFocusID = "last-focus"
	state.Visible.Set(false)
	state.SetNextFocusIDOnClose("override-focus")

	oldPending := pendingFocusID
	defer func() { pendingFocusID = oldPending }()
	pendingFocusID = ""

	ctx := NewBuildContext(
		NewFocusManager(),
		NewAnySignal[Focusable](nil),
		NewAnySignal[Widget](nil),
		NewFloatCollector(),
	)

	_ = palette.Build(ctx)

	if pendingFocusID != "override-focus" {
		t.Fatalf("expected pending focus override, got %q", pendingFocusID)
	}
	if state.nextFocusID != "" {
		t.Fatalf("expected nextFocusID to be cleared, got %q", state.nextFocusID)
	}
}

func TestSnapshot_CommandPalette_Basic(t *testing.T) {
	items := []CommandPaletteItem{
		{Label: "New File", Hint: "Ctrl+N"},
		{Label: "Open File", Hint: "Ctrl+O", Description: "Open from disk"},
		{Divider: "Edit"},
		{Label: "Cut", Hint: "Ctrl+X"},
		{Label: "Copy", Hint: "Ctrl+C", Disabled: true},
		{
			Label: "Primary",
			HintWidget: func() Widget {
				return Text{
					Content: "  ",
					Style: Style{
						BackgroundColor: RGB(100, 149, 237),
					},
				}
			},
		},
	}

	state := NewCommandPaletteState("Commands", items)
	state.Visible.Set(true)

	level := state.CurrentLevel()
	level.InputState.SetText("op")
	level.FilterState.Query.Set("op")
	level.ListState.SelectIndex(1)

	widget := CommandPalette{
		ID:       "palette-basic",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, widget, 80, 24, "Command palette with filter text, divider, disabled item, and hint widget")
}

func TestSnapshot_CommandPalette_Nested(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Theme"},
		{Label: "Settings"},
	})
	state.PushLevel("Theme", []CommandPaletteItem{
		{Label: "Rose Pine"},
		{Label: "Dracula"},
	})
	state.Visible.Set(true)

	widget := CommandPalette{
		ID:       "palette-nested",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, widget, 80, 20, "Nested command palette showing breadcrumbs and theme options")
}

func TestSnapshot_CommandPalette_NoResults(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Open File"},
		{Label: "Save All"},
	})
	state.Visible.Set(true)

	level := state.CurrentLevel()
	level.InputState.SetText("zzz")
	level.FilterState.Query.Set("zzz")

	widget := CommandPalette{
		ID:       "palette-empty",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, widget, 80, 20, "Command palette showing empty state when no items match the filter")
}

func TestSnapshot_CommandPalette_ScrollOverflow(t *testing.T) {
	items := make([]CommandPaletteItem, 0, 30)
	for i := 0; i < 30; i++ {
		items = append(items, CommandPaletteItem{
			Label: fmt.Sprintf("File %02d", i+1),
			Hint:  "txt",
		})
	}

	state := NewCommandPaletteState("Files", items)
	state.Visible.Set(true)

	level := state.CurrentLevel()
	level.InputState.SetText("")
	level.FilterState.Query.Set("")
	// Force a large offset so the layout clamps to the bottom of the list.
	level.ScrollState.Offset.Set(999)

	widget := CommandPalette{
		ID:       "palette-scroll-overflow",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
		Style: Style{
			MaxHeight: Cells(8),
		},
	}

	AssertSnapshot(t, widget, 60, 16, "Command palette with constrained height and enough items to require scrolling; scrollbar should remain visible within the palette")
}
