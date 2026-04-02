package terma

import (
	"strconv"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

type buildCountWidget struct {
	builds *int
	child  Widget
}

func (w buildCountWidget) Build(ctx BuildContext) Widget {
	if w.builds != nil {
		(*w.builds)++
	}
	if w.child != nil {
		return w.child
	}
	return Text{Content: "x"}
}

type trackingScreen struct {
	*uv.Buffer
	touched map[[2]int]struct{}
}

func newTrackingScreen(width, height int) *trackingScreen {
	return &trackingScreen{
		Buffer:  uv.NewBuffer(width, height),
		touched: make(map[[2]int]struct{}),
	}
}

func (s *trackingScreen) SetCell(x, y int, c *uv.Cell) {
	s.touched[[2]int{x, y}] = struct{}{}
	s.Buffer.SetCell(x, y, c)
}

func (s *trackingScreen) ClearArea(area uv.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			s.touched[[2]int{x, y}] = struct{}{}
		}
	}
	s.Buffer.ClearArea(area)
}

func (s *trackingScreen) resetTouched() {
	s.touched = make(map[[2]int]struct{})
}

func newTestRenderer(screen CellBuffer, width, height int) *Renderer {
	return NewRenderer(
		screen,
		width,
		height,
		NewFocusManager(),
		NewAnySignal[Focusable](nil),
		NewAnySignal[Widget](nil),
	)
}

func TestSignal_PhaseSubscriptionsDriveDirtyLevels(t *testing.T) {
	t.Run("build", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhaseBuild, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyBuild {
			t.Fatalf("expected DirtyBuild, got %v", got)
		}
	})

	t.Run("layout", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhaseLayout, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyLayout {
			t.Fatalf("expected DirtyLayout, got %v", got)
		}
	})

	t.Run("paint", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyPaint {
			t.Fatalf("expected DirtyPaint, got %v", got)
		}
	})

	t.Run("escalates to highest phase", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
		withTrackedRead(node, readPhaseLayout, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyLayout {
			t.Fatalf("expected DirtyLayout, got %v", got)
		}
	})
}

func TestSignal_DisposeClearsSubscriptions(t *testing.T) {
	s := NewSignal(1)
	node := newWidgetNode(nil)
	withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
	node.clearDirty()

	node.dispose()
	s.Set(2)

	if node.isDirty() {
		t.Fatal("disposed node should not be dirtied by future signal updates")
	}
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	if len(s.core.listeners) != 0 {
		t.Fatalf("expected listeners to be cleared, got %d", len(s.core.listeners))
	}
}

func TestRenderer_RowBuildsChildrenOncePerFrame(t *testing.T) {
	builds := 0
	widget := Row{
		Children: []Widget{
			buildCountWidget{builds: &builds},
		},
	}

	renderer := newTestRenderer(uv.NewBuffer(10, 1), 10, 1)
	renderer.Render(widget)

	if builds != 1 {
		t.Fatalf("expected child Build to run once, got %d", builds)
	}
}

func TestRenderer_SpinnerTickUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewSpinnerState(SpinnerLine)
	builds := 0
	root := buildCountWidget{
		builds: &builds,
		child: Row{
			Children: []Widget{
				Spinner{State: state},
				Text{Content: " loading"},
			},
		},
	}

	renderer.Update(root)
	if renderer.lastFrameMode != rendererFrameFull {
		t.Fatalf("expected initial full render, got %q", renderer.lastFrameMode)
	}
	if builds != 1 {
		t.Fatalf("expected one build on initial render, got %d", builds)
	}

	screen.resetTouched()
	state.animation.Value().Set("|")
	renderer.Update(root)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint on spinner tick, got %q", renderer.lastFrameMode)
	}
	if builds != 1 {
		t.Fatalf("spinner tick should not rebuild root, got %d builds", builds)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("spinner tick should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
	for pos := range screen.touched {
		if pos[0] != 0 || pos[1] != 0 {
			t.Fatalf("expected only spinner cell to be touched, got (%d,%d)", pos[0], pos[1])
		}
	}
}

func TestRenderer_MixedWidthSpinnerFramesStayPaintOnly(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewSpinnerState(SpinnerStyle{
		Frames:    []string{"-", ">>"},
		FrameTime: 50 * time.Millisecond,
	})

	renderer.Update(Spinner{State: state})
	screen.resetTouched()

	state.animation.Value().Set(">>")
	renderer.Update(Spinner{State: state})

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for mixed-width spinner tick, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("mixed-width spinner tick should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
}

func TestRenderer_TextInputAutoWidthPromotesLayoutOnlyWhenIntrinsicWidthChanges(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewTextInputState("ab")
	widget := TextInput{State: state, Width: Auto}

	renderer.Update(widget)

	state.CursorLeft()
	renderer.Update(widget)
	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("cursor move should stay paint-only, got %q", renderer.lastFrameMode)
	}

	state.Insert("c")
	renderer.Update(widget)
	if renderer.lastFrameMode != rendererFrameFull {
		t.Fatalf("content width growth should force full render, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount == 0 {
		t.Fatal("expected layout work after intrinsic width changed")
	}
}

func TestRenderer_CleanUpdateDoesNotFallbackToFullRender(t *testing.T) {
	renderer := newTestRenderer(uv.NewBuffer(10, 1), 10, 1)
	root := Text{Content: "steady"}

	renderer.Update(root)
	fullCount := renderer.fullRenderCount
	partialCount := renderer.partialRenderCount

	renderer.Update(root)

	if renderer.fullRenderCount != fullCount {
		t.Fatalf("clean update should not trigger a full render: got %d, want %d", renderer.fullRenderCount, fullCount)
	}
	if renderer.partialRenderCount != partialCount {
		t.Fatalf("clean update should not trigger a partial render: got %d, want %d", renderer.partialRenderCount, partialCount)
	}
}

func TestRenderer_TextInputBlurRepaintsVirtualCursor(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		20,
		1,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewTextInputState("hello")
	widget := TextInput{ID: "input", State: state, Width: Cells(10)}

	focusManager.focusedID = "input"
	focusedSignal.Set(widget)
	renderer.Update(widget)

	screen.resetTouched()
	focusManager.focusedID = ""
	focusedSignal.Set(nil)
	renderer.Update(widget)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint on blur, got %q", renderer.lastFrameMode)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected blur to repaint the input and clear the virtual cursor")
	}
}

func TestRenderer_SignalMarkupUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	cursor := NewSignal(1)
	widget := SignalMarkup(cursor, func(i int) string {
		return "Cursor: [b]" + strconv.Itoa(i) + "[/]"
	})

	renderer.Update(widget)
	screen.resetTouched()

	cursor.Set(2)
	renderer.Update(widget)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for SignalMarkup, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("SignalMarkup should not relayout when width is stable, got %d layout ops", renderer.lastLayoutCount)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected SignalMarkup update to repaint text")
	}
}

func TestRenderer_SignalTextAutoWidthPromotesLayoutWhenWidthChanges(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	content := NewSignal("ab")
	widget := SignalText(content, func(s string) string { return s })
	widget.LayoutStyle.Width = Auto

	renderer.Update(widget)

	content.Set("wider")
	renderer.Update(widget)

	if renderer.lastFrameMode != rendererFrameFull {
		t.Fatalf("expected full render after SignalText width growth, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount == 0 {
		t.Fatal("expected layout work after SignalText intrinsic width changed")
	}
}

func TestRenderer_ListCursorMoveUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(20, 3)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		20,
		3,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewListState([]string{"One", "Two", "Three"})
	list := List[string]{ID: "list", State: state}

	focusManager.focusedID = "list"
	focusedSignal.Set(list)
	renderer.Update(list)

	screen.resetTouched()
	state.SelectNext()
	renderer.Update(list)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for list cursor move, got %q", renderer.lastFrameMode)
	}
	if renderer.lastBuildCount != 0 {
		t.Fatalf("list cursor move should not rebuild, got %d build ops", renderer.lastBuildCount)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("list cursor move should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected list cursor move to repaint affected rows")
	}
}

type listSelectionSummaryRoot struct {
	list  List[string]
	state *ListState[string]
}

func (w listSelectionSummaryRoot) Build(ctx BuildContext) Widget {
	selection := w.state.Selection.Get()
	return Column{
		Children: []Widget{
			w.list,
			Text{Content: "selected: " + strconv.Itoa(len(selection))},
		},
	}
}

func TestRenderer_ListCursorMoveDoesNotRebuildWhenSelectionAlreadyEmpty(t *testing.T) {
	screen := newTrackingScreen(30, 6)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		30,
		6,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewListState([]string{"One", "Two", "Three"})
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	root := listSelectionSummaryRoot{
		list:  list,
		state: state,
	}

	focusManager.focusedID = "list"
	focusedSignal.Set(list)
	renderer.Update(root)

	screen.resetTouched()
	list.keyCursorDown()
	renderer.Update(root)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for cursor move with empty selection summary, got %q", renderer.lastFrameMode)
	}
	if renderer.lastBuildCount != 0 {
		t.Fatalf("cursor move with unchanged empty selection should not rebuild, got %d build ops", renderer.lastBuildCount)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("cursor move with unchanged empty selection should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
}

func TestRenderer_AutocompleteDoesNotSelfInvalidateAfterBuild(t *testing.T) {
	screen := newTrackingScreen(40, 8)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		40,
		8,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	inputState := NewTextInputState("he")
	inputState.CursorIndex.Set(2)
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "hello"},
		{Label: "help"},
	})

	widget := Autocomplete{
		ID:    "ac",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(20)},
	}

	focusManager.focusedID = "input"
	focusedSignal.Set(widget.Child.(TextInput))
	renderer.Update(widget)

	screenText := renderer.ScreenText()
	if !strings.Contains(screenText, "hello") {
		t.Fatalf("expected autocomplete popup to render on first update, got screen:\n%s", screenText)
	}

	fullCount := renderer.fullRenderCount
	partialCount := renderer.partialRenderCount
	renderer.Update(widget)

	if renderer.fullRenderCount != fullCount || renderer.partialRenderCount != partialCount {
		t.Fatalf("expected clean follow-up update to be a no-op, got full=%d->%d partial=%d->%d",
			fullCount, renderer.fullRenderCount, partialCount, renderer.partialRenderCount)
	}
}

func TestRenderer_AutocompleteTracksTextInputTyping(t *testing.T) {
	screen := newTrackingScreen(40, 8)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		40,
		8,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	input := TextInput{ID: "input", State: NewTextInputState(""), Width: Cells(20)}
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{{Label: "hello"}, {Label: "help"}})
	widget := Autocomplete{
		ID:    "ac",
		State: acState,
		Child: input,
	}

	focusManager.focusedID = "input"
	focusedSignal.Set(input)
	renderer.Update(widget)

	input.State.SetText("he")
	input.State.CursorIndex.Set(2)
	renderer.Update(widget)

	if !strings.Contains(renderer.ScreenText(), "hello") {
		t.Fatalf("expected always-on autocomplete to rerender suggestions after typing, got screen:\n%s", renderer.ScreenText())
	}
}

func TestRenderer_AutocompleteTracksTriggerTyping(t *testing.T) {
	screen := newTrackingScreen(40, 8)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		40,
		8,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	input := TextInput{ID: "tag-input", State: NewTextInputState(""), Width: Cells(20)}
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{{Label: "bug", Value: "#bug"}, {Label: "feature", Value: "#feature"}})
	widget := Autocomplete{
		ID:           "tag-ac",
		State:        acState,
		TriggerChars: []rune{'#'},
		MinChars:     0,
		Child:        input,
	}

	focusManager.focusedID = "tag-input"
	focusedSignal.Set(input)
	renderer.Update(widget)

	input.State.SetText("#")
	input.State.CursorIndex.Set(1)
	renderer.Update(widget)

	if !strings.Contains(renderer.ScreenText(), "bug") {
		t.Fatalf("expected trigger autocomplete to show popup after typing trigger, got screen:\n%s", renderer.ScreenText())
	}
}

func TestAutocomplete_SelectUsesFilteredSuggestionAfterQueryNarrowing(t *testing.T) {
	input := NewTextInputState("")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "apple", Value: "apple"},
		{Label: "banana", Value: "banana"},
		{Label: "cherry", Value: "cherry"},
	})

	selected := ""
	widget := Autocomplete{
		State:     acState,
		MatchMode: FilterContains,
		Insert:    InsertReplace,
		Child: TextInput{
			ID:    "input",
			State: input,
		},
		OnSelect: func(s Suggestion) {
			selected = s.Value
		},
	}

	input.SetText("a")
	input.CursorIndex.Set(1)
	widget.handleTextChange("a", 1)
	widget.filteredSuggestionCount()
	widget.onDown() // move to banana

	input.SetText("ap")
	input.CursorIndex.Set(2)
	widget.handleTextChange("ap", 2)
	widget.filteredSuggestionCount()
	widget.selectCurrentSuggestion()

	if selected != "apple" {
		t.Fatalf("expected filtered selection to resolve to apple, got %q", selected)
	}
	if got := input.GetText(); got != "apple" {
		t.Fatalf("expected input text to be replaced with apple, got %q", got)
	}
}

func TestRenderer_TableCursorMoveUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(30, 3)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		30,
		3,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewTableState([][]string{
		{"Alice", "Engineer"},
		{"Bob", "Designer"},
	})
	table := Table[[]string]{
		ID:    "table",
		State: state,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(10)},
		},
	}

	focusManager.focusedID = "table"
	focusedSignal.Set(table)
	renderer.Update(table)

	screen.resetTouched()
	state.SelectColumn(1)
	renderer.Update(table)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for table cursor move, got %q", renderer.lastFrameMode)
	}
	if renderer.lastBuildCount != 0 {
		t.Fatalf("table cursor move should not rebuild, got %d build ops", renderer.lastBuildCount)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("table cursor move should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected table cursor move to repaint affected cells")
	}
}

func TestRenderer_TreeCursorMoveUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(30, 6)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		30,
		6,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewTreeState(sampleTreeSnapshotNodes())
	tree := Tree[string]{ID: "tree", State: state}

	focusManager.focusedID = "tree"
	focusedSignal.Set(tree)
	renderer.Update(tree)

	screen.resetTouched()
	state.CursorDown()
	renderer.Update(tree)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for tree cursor move, got %q", renderer.lastFrameMode)
	}
	if renderer.lastBuildCount != 0 {
		t.Fatalf("tree cursor move should not rebuild, got %d build ops", renderer.lastBuildCount)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("tree cursor move should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected tree cursor move to repaint affected rows")
	}
}
