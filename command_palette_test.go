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

func TestCommandPaletteState_CloseResetsRootCursorPosition(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Divider: "Group"},
		{Label: "Open"},
		{Label: "Save"},
	})

	root := state.CurrentLevel()
	if root == nil {
		t.Fatal("expected root level")
	}

	root.InputState.SetText("sa")
	root.InputState.SetSelectionAnchor(0)
	root.ScrollState.SetOffset(4)
	root.FilterState.Query.Set("sa")
	root.ListState.SelectIndex(2)

	state.PushLevel("Nested", []CommandPaletteItem{
		{Label: "Nested action"},
	})
	state.Close()
	state.Open()

	root = state.CurrentLevel()
	if root == nil {
		t.Fatal("expected root level after reopen")
	}
	if state.IsNested() {
		t.Fatal("expected palette to reset to root level")
	}
	if got := root.InputState.GetText(); got != "" {
		t.Fatalf("expected cleared input text, got %q", got)
	}
	if got := root.InputState.CursorIndex.Peek(); got != 0 {
		t.Fatalf("expected input cursor at start, got %d", got)
	}
	if got := root.InputState.SelectionAnchor.Peek(); got != -1 {
		t.Fatalf("expected cleared input selection, got %d", got)
	}
	if got := root.FilterState.Query.Peek(); got != "" {
		t.Fatalf("expected cleared filter query, got %q", got)
	}
	if got := root.ScrollState.GetOffset(); got != 0 {
		t.Fatalf("expected reset scroll offset, got %d", got)
	}
	if got := root.ListState.CursorIndex.Peek(); got != 1 {
		t.Fatalf("expected list cursor to reset to first selectable item, got %d", got)
	}
}

func TestCommandPaletteState_CloseKeepsPositionWhenRequested(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Open"},
		{Label: "Save"},
	})

	root := state.CurrentLevel()
	if root == nil {
		t.Fatal("expected root level")
	}

	root.InputState.SetText("sa")
	root.FilterState.Query.Set("sa")
	root.ListState.SelectIndex(1)

	state.PushLevel("Nested", []CommandPaletteItem{
		{Label: "Nested action"},
	})
	state.Close(true)
	state.Open()

	if !state.IsNested() {
		t.Fatal("expected palette level stack to be preserved")
	}

	level := state.CurrentLevel()
	if level == nil {
		t.Fatal("expected current level after reopen")
	}
	if got := level.Title; got != "Nested" {
		t.Fatalf("expected nested level to remain active, got %q", got)
	}
}

func TestCommandPalette_FloatOffset_DefaultTopInset(t *testing.T) {
	tests := []struct {
		name     string
		palette  CommandPalette
		expected Offset
	}{
		{
			name:     "default top center when position unset",
			palette:  CommandPalette{},
			expected: Offset{X: 0, Y: 2},
		},
		{
			name:     "top left with no explicit y",
			palette:  CommandPalette{Position: FloatPositionTopLeft},
			expected: Offset{X: 0, Y: 2},
		},
		{
			name:     "top right preserves x while applying default y",
			palette:  CommandPalette{Position: FloatPositionTopRight, Offset: Offset{X: 3}},
			expected: Offset{X: 3, Y: 2},
		},
		{
			name:     "absolute without offset still uses top default",
			palette:  CommandPalette{Position: FloatPositionAbsolute},
			expected: Offset{X: 0, Y: 2},
		},
	}

	for _, tt := range tests {
		got := tt.palette.floatOffset()
		if got != tt.expected {
			t.Fatalf("%s: got %+v, want %+v", tt.name, got, tt.expected)
		}
	}
}

func TestCommandPalette_FloatOffset_ExplicitAndNonTopOffsetsUnchanged(t *testing.T) {
	tests := []struct {
		name     string
		palette  CommandPalette
		expected Offset
	}{
		{
			name:     "explicit top y preserved",
			palette:  CommandPalette{Position: FloatPositionTopCenter, Offset: Offset{Y: 1}},
			expected: Offset{X: 0, Y: 1},
		},
		{
			name:     "explicit large top y preserved",
			palette:  CommandPalette{Position: FloatPositionTopCenter, Offset: Offset{Y: 5}},
			expected: Offset{X: 0, Y: 5},
		},
		{
			name:     "center position unchanged",
			palette:  CommandPalette{Position: FloatPositionCenter},
			expected: Offset{X: 0, Y: 0},
		},
		{
			name:     "bottom position unchanged",
			palette:  CommandPalette{Position: FloatPositionBottomLeft},
			expected: Offset{X: 0, Y: 0},
		},
		{
			name:     "absolute with explicit coordinates unchanged",
			palette:  CommandPalette{Position: FloatPositionAbsolute, Offset: Offset{X: 4, Y: 1}},
			expected: Offset{X: 4, Y: 1},
		},
	}

	for _, tt := range tests {
		got := tt.palette.floatOffset()
		if got != tt.expected {
			t.Fatalf("%s: got %+v, want %+v", tt.name, got, tt.expected)
		}
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

func TestSnapshot_CommandPalette_DividerHeaders(t *testing.T) {
	items := []CommandPaletteItem{
		{Divider: "File"},
		{Label: "New File", Hint: "Ctrl+N", Description: "Create an empty buffer"},
		{Label: "Open File", Hint: "Ctrl+O", Description: "Open from disk"},
		{Divider: "Edit"},
		{Label: "Cut", Hint: "Ctrl+X", Description: "Move selection to clipboard"},
		{Divider: ""},
		{Label: "Preferences"},
	}

	state := NewCommandPaletteState("Commands", items)
	state.Visible.Set(true)

	widget := CommandPalette{
		ID:       "palette-divider-headers",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, widget, 80, 24, "Titled dividers 'File' and 'Edit' render bold in the primary text colour, distinct from the muted descriptions and hints; an untitled divider renders as a plain line")
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
	// Keep the cursor at the bottom while testing an offset past the list.
	level.ListState.SelectIndex(len(items) - 1)
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

func TestCommandPalette_MoveCursorScrollsIntoView(t *testing.T) {
	items := make([]CommandPaletteItem, 0, 8)
	for i := 0; i < 8; i++ {
		items = append(items, CommandPaletteItem{Label: fmt.Sprintf("Item %d", i+1)})
	}

	state := NewCommandPaletteState("Commands", items)
	level := state.CurrentLevel()
	if level == nil {
		t.Fatal("expected current level")
	}

	viewIndices := make([]int, len(items))
	layouts := make([]listItemLayout, len(items))
	for i := range items {
		viewIndices[i] = i
		layouts[i] = listItemLayout{y: i, height: 1}
	}
	level.ListState.setViewIndices(viewIndices)
	level.ListState.itemLayouts = layouts
	level.ScrollState.updateLayout(3, len(items))

	palette := CommandPalette{
		ID:    "palette",
		State: state,
	}

	palette.moveCursor(1)
	palette.moveCursor(1)
	if got := level.ScrollState.GetOffset(); got != 0 {
		t.Fatalf("expected offset 0 while cursor remains visible, got %d", got)
	}

	palette.moveCursor(1)
	if got := level.ScrollState.GetOffset(); got != 1 {
		t.Fatalf("expected offset 1 after cursor moved past viewport, got %d", got)
	}

	if got := level.ListState.CursorIndex.Peek(); got != 3 {
		t.Fatalf("expected cursor index 3, got %d", got)
	}
}

// typeInPalette sets the current level's query as if typed into the input.
func typeInPalette(p CommandPalette, text string) {
	level := p.State.CurrentLevel()
	input := p.buildInput(level, ThemeData{}).(TextInput)
	level.InputState.SetText(text)
	input.OnChange(text)
}

func currentPaletteLabel(t *testing.T, state *CommandPaletteState) string {
	t.Helper()
	item, ok := state.CurrentItem()
	if !ok {
		t.Fatal("expected a current item")
	}
	return item.Label
}

func TestCommandPalette_TypingMovesCursorToTopResult(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Profile Settings"},
		{Label: "Save"},
		{Divider: "Files"},
		{Label: "New File"},
	})
	palette := CommandPalette{ID: "palette", State: state}

	// The cursor starts on an item that still matches, but isn't the best match.
	typeInPalette(palette, "fil")
	if got := currentPaletteLabel(t, state); got != "New File" {
		t.Fatalf("expected cursor on top result, got %q", got)
	}

	palette.moveCursor(1)
	if got := currentPaletteLabel(t, state); got != "Profile Settings" {
		t.Fatalf("expected cursor on second result, got %q", got)
	}

	// Typing more re-ranks, so the cursor returns to the top.
	typeInPalette(palette, "file")
	if got := currentPaletteLabel(t, state); got != "New File" {
		t.Fatalf("expected cursor back on top result, got %q", got)
	}

	// Clearing the query returns to the first item of the full list.
	typeInPalette(palette, "")
	if got := currentPaletteLabel(t, state); got != "Profile Settings" {
		t.Fatalf("expected cursor on first item, got %q", got)
	}
}

func TestCommandPalette_TypingScrollsBackToTop(t *testing.T) {
	items := make([]CommandPaletteItem, 0, 20)
	for i := 0; i < 20; i++ {
		items = append(items, CommandPaletteItem{Label: fmt.Sprintf("Item %02d", i)})
	}
	state := NewCommandPaletteState("Commands", items)
	palette := CommandPalette{ID: "palette", State: state}
	level := state.CurrentLevel()
	level.ScrollState.updateLayout(5, len(items))
	level.ListState.SelectIndex(15)
	level.ScrollState.SetOffset(11)

	typeInPalette(palette, "item")

	if got := level.ScrollState.GetOffset(); got != 0 {
		t.Fatalf("expected scroll offset reset to 0, got %d", got)
	}
	if got := currentPaletteLabel(t, state); got != "Item 00" {
		t.Fatalf("expected cursor on top result, got %q", got)
	}
}

func TestCommandPaletteFilteredView_HidesDividersWhileSearching(t *testing.T) {
	items := []CommandPaletteItem{
		{Divider: "Files"},
		{Label: "New File"},
		{Divider: "Edit"},
		{Label: "Copy"},
	}
	filter := NewFilterState()
	filter.Mode.Set(FilterFuzzy)

	if view := commandPaletteFilteredView(items, filter); len(view.Indices) != 4 {
		t.Fatalf("expected every item with no query, got %v", view.Indices)
	}

	filter.Query.Set("file")
	view := commandPaletteFilteredView(items, filter)
	if want := []int{1}; !reflect.DeepEqual(view.Indices, want) {
		t.Fatalf("expected only matching items while searching: got %v, want %v", view.Indices, want)
	}
}

func TestCommandPaletteFilteredView_KeywordMatchesRankBelowLabelMatches(t *testing.T) {
	items := []CommandPaletteItem{
		{Label: "Toggle Sidebar", FilterText: "Toggle Sidebar layout panel"},
		{Label: "Reset Layout"},
	}
	filter := NewFilterState()
	filter.Mode.Set(FilterFuzzy)
	filter.Query.Set("layout")

	view := commandPaletteFilteredView(items, filter)

	if want := []int{1, 0}; !reflect.DeepEqual(view.Indices, want) {
		t.Fatalf("expected label match first: got %v, want %v", view.Indices, want)
	}
	if ranges := view.Matches[1].Ranges; len(ranges) != 0 {
		t.Fatalf("expected no label highlight for a keyword-only match, got %v", ranges)
	}
}

func TestCommandPalette_LevelOpensOnCurrentItem(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{{Label: "Theme"}})
	palette := CommandPalette{ID: "palette", State: state}

	state.PushLevel("Theme", []CommandPaletteItem{
		{Divider: "Dark"},
		{Label: "Dracula"},
		{Label: "Nord"},
		{Divider: "Light"},
		{Label: "Solarized Light", Current: true},
	})
	if got := currentPaletteLabel(t, state); got != "Solarized Light" {
		t.Fatalf("expected level to open on its current item, got %q", got)
	}

	typeInPalette(palette, "d")
	if got := currentPaletteLabel(t, state); got != "Dracula" {
		t.Fatalf("expected cursor on top result while searching, got %q", got)
	}

	typeInPalette(palette, "")
	if got := currentPaletteLabel(t, state); got != "Solarized Light" {
		t.Fatalf("expected cursor back on current item after clearing, got %q", got)
	}
}

func TestCommandPalette_BackKeepsParentSearchAndCursor(t *testing.T) {
	var pushed []string
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "New File"},
		{
			Label: "Themes",
			Children: func() []CommandPaletteItem {
				pushed = append(pushed, "Themes")
				return []CommandPaletteItem{{Label: "Dracula"}}
			},
		},
		{Label: "Toggle Sidebar"},
	})
	palette := CommandPalette{ID: "palette", State: state}

	typeInPalette(palette, "th")
	if got := currentPaletteLabel(t, state); got != "Themes" {
		t.Fatalf("expected Themes as top result, got %q", got)
	}

	palette.selectCurrent()
	if !state.IsNested() || len(pushed) != 1 {
		t.Fatalf("expected Enter to open the nested level, pushed %v", pushed)
	}
	if got := state.CurrentLevel().FilterState.PeekQuery(); got != "" {
		t.Fatalf("expected nested level to start with an empty query, got %q", got)
	}

	palette.handleEscape()
	if state.IsNested() {
		t.Fatal("expected Escape to go back to the root level")
	}
	root := state.CurrentLevel()
	if got := root.InputState.GetText(); got != "th" {
		t.Fatalf("expected parent query to be kept, got %q", got)
	}
	if got := currentPaletteLabel(t, state); got != "Themes" {
		t.Fatalf("expected cursor on the item that opened the level, got %q", got)
	}

	palette.handleEscape()
	if state.Visible.Peek() {
		t.Fatal("expected Escape at the root to close the palette")
	}
	state.Open()
	if got := root.InputState.GetText(); got != "" {
		t.Fatalf("expected query cleared on reopen, got %q", got)
	}
	if got := currentPaletteLabel(t, state); got != "New File" {
		t.Fatalf("expected cursor on first item on reopen, got %q", got)
	}
}

func TestCommandPalette_EnterWithNoResultsDoesNothing(t *testing.T) {
	selected := false
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Save", Action: func() { selected = true }},
	})
	palette := CommandPalette{ID: "palette", State: state}

	typeInPalette(palette, "zzz")
	palette.selectCurrent()

	if selected {
		t.Fatal("expected Enter with no results not to run an action")
	}
}

func TestSnapshot_CommandPalette_RankedSearch(t *testing.T) {
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{
		{Label: "Profile Settings"},
		{Divider: "File"},
		{Label: "Open Recent", Hint: "Ctrl+R"},
		{Label: "Find in Files", Hint: "Ctrl+Shift+F"},
		{Label: "New File", Hint: "Ctrl+N"},
		{Divider: "View"},
		{Label: "Toggle Sidebar", Hint: "Ctrl+B"},
	})
	state.Visible.Set(true)
	palette := CommandPalette{
		ID:       "palette-ranked",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}
	typeInPalette(palette, "file")

	AssertSnapshot(t, palette, 70, 14, "Query 'file' lists 'New File' (cursor), 'Find in Files', then mid-word 'Profile Settings'; no divider rows")
}

func TestSnapshot_CommandPalette_NestedOpensOnCurrent(t *testing.T) {
	items := []CommandPaletteItem{{Divider: "Themes"}}
	for i := 0; i < 20; i++ {
		items = append(items, CommandPaletteItem{
			Label:   fmt.Sprintf("Theme %02d", i+1),
			Current: i == 15,
		})
	}
	state := NewCommandPaletteState("Commands", []CommandPaletteItem{{Label: "Theme"}})
	state.PushLevel("Theme", items)
	state.Visible.Set(true)

	palette := CommandPalette{
		ID:       "palette-current",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, palette, 70, 20, "Nested Theme level scrolled so the cursor sits on the current item, 'Theme 16'")
}
