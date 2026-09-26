package terma

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
	"github.com/stretchr/testify/require"
)

// Each side owns its widget state, renderer, focus manager, and buffer. The
// reference always renders fully; it cannot repair the incremental side's state.
type reactivitySequence[T Widget] struct {
	t        *testing.T
	actual   *reactivitySurface[T]
	expected *reactivitySurface[T]
	frames   []SnapshotComparison
	output   string
}

// reactivityScreen lets the renderer clear the buffer on full frames, as it
// clears the real terminal. A bare buffer would keep stale cells.
type reactivityScreen struct{ *uv.Buffer }

func (reactivityScreen) WidthMethod() uv.WidthMethod { return ansi.GraphemeWidth }

type reactivitySurface[T Widget] struct {
	root       T
	buffer     *uv.Buffer
	renderer   *Renderer
	focus      *FocusManager
	focused    AnySignal[Focusable]
	focusables []FocusableEntry
}

func newReactivitySequence[T Widget](t *testing.T, width, height int, factory func() T) *reactivitySequence[T] {
	t.Helper()
	newSurface := func() *reactivitySurface[T] {
		root := factory()
		buffer := uv.NewBuffer(width, height)
		focus := NewFocusManager()
		focus.SetRootWidget(root)
		focused := NewAnySignal[Focusable](nil)
		return &reactivitySurface[T]{
			root: root, buffer: buffer, focus: focus, focused: focused,
			renderer: NewRenderer(reactivityScreen{buffer}, width, height, focus, focused, NewAnySignal[Widget](nil)),
		}
	}
	sequence := &reactivitySequence[T]{
		t: t, actual: newSurface(), expected: newSurface(),
		output: os.Getenv("TERMA_REACTIVITY_OUTPUT"),
	}
	previousPendingFocus := pendingFocusID
	t.Cleanup(func() {
		pendingFocusID = previousPendingFocus
		for _, side := range []*reactivitySurface[T]{sequence.actual, sequence.expected} {
			if side.renderer.rootNode != nil {
				side.renderer.rootNode.dispose()
			}
			for _, float := range side.renderer.retainedFloats {
				float.root.dispose()
			}
		}
		if sequence.output == "" && !t.Failed() {
			return
		}
		output := sequence.output
		if output == "" {
			output = filepath.Join("snapshot-output", "reactivity")
		}
		output = filepath.Join(output, t.Name())
		require.NoError(t, os.MkdirAll(output, 0755))
		for i, frame := range sequence.frames {
			for label, svg := range map[string]string{"actual": frame.Actual, "expected": frame.Expected, "diff": frame.DiffSVG} {
				path := filepath.Join(output, fmt.Sprintf("%02d-%s.svg", i, label))
				require.NoError(t, os.WriteFile(path, []byte(svg), 0644))
			}
		}
		gallery := filepath.Join(output, "index.html")
		require.NoError(t, GenerateGallery(sequence.frames, gallery))
		t.Logf("incremental/full frame comparison: %s", gallery)
	})
	return sequence
}

func (s *reactivitySurface[T]) draw(full bool) {
	render := s.renderer.Update
	if full {
		render = s.renderer.Render
	}
	// Settle focus changes just as the app does, but bound the loop so a
	// self-invalidating fixture cannot hang the suite.
	for pass := 0; pass < 4; pass++ {
		pendingFocusID = ""
		before := s.focus.FocusedID()
		s.focusables = render(s.root)
		s.focus.SetFocusables(s.focusables)
		if pendingFocusID != "" {
			s.focus.FocusByID(pendingFocusID)
			pendingFocusID = ""
		}
		if s.focus.FocusedID() == before {
			return
		}
		s.focused.Set(s.focus.Focused())
	}
	panic("reactivity fixture did not settle focus")
}

func (s *reactivitySequence[T]) frame(name string, change func(T)) RenderStats {
	s.t.Helper()
	if change != nil {
		change(s.actual.root)
		change(s.expected.root)
	}
	framesBefore := s.actual.renderer.fullRenderCount + s.actual.renderer.partialRenderCount
	s.actual.draw(false)
	s.expected.draw(true)
	width, height := s.actual.buffer.Width(), s.actual.buffer.Height()
	stats := CompareBuffers(s.expected.buffer, s.actual.buffer, width, height)
	opts := DefaultSVGOptions()
	work := s.actual.renderer.Stats()
	if s.actual.renderer.fullRenderCount+s.actual.renderer.partialRenderCount == framesBefore {
		// Nothing was dirty, so no frame ran; Stats still describes the last one.
		work = RenderStats{}
	}
	s.frames = append(s.frames, SnapshotComparison{
		Name: name, Passed: stats.MismatchedCells == 0, Stats: stats,
		Description: fmt.Sprintf("Incremental: %s, builds=%d, layouts=%d, paints=%d. Expected: forced full render.", work.FrameMode, work.BuildCount, work.LayoutCount, work.PaintCount),
		Expected:    BufferToSVG(s.expected.buffer, width, height, opts),
		Actual:      BufferToSVG(s.actual.buffer, width, height, opts),
		DiffSVG:     GenerateDiffSVG(s.expected.buffer, s.actual.buffer, width, height, opts),
	})
	require.Zero(s.t, stats.MismatchedCells, "%s: incremental output differs from full render:\n%s", name, describeMismatches(s.expected.buffer, s.actual.buffer, width, height))
	require.Equal(s.t, s.expected.focus.FocusedID(), s.actual.focus.FocusedID(), "%s: focused widget", name)
	// A node left dirty is invisible in the picture but makes later frames do
	// more work than needed, so every frame must leave the tree clean.
	var requireClean func(*widgetNode, string)
	requireClean = func(node *widgetNode, path string) {
		if node == nil {
			return
		}
		require.Equal(s.t, DirtyNone, node.dirtyLevel(), "%s: node %s left dirty", name, path)
		require.Equal(s.t, DirtyNone, node.subtreeDirtyLevel(), "%s: node %s subtree left dirty", name, path)
		for i, child := range node.children {
			requireClean(child, fmt.Sprintf("%s/%d", path, i))
		}
	}
	requireClean(s.actual.renderer.rootNode, "root")
	focusOrder := func(entries []FocusableEntry) []string {
		result := make([]string, len(entries))
		for i, entry := range entries {
			result[i] = entry.ID + "/" + entry.TrapID
		}
		return result
	}
	require.Equal(s.t, focusOrder(s.expected.focusables), focusOrder(s.actual.focusables), "%s: focus order and traps", name)
	// A correct picture must also have correct hit targets after reflow.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			actual, expected := s.actual.renderer.WidgetAt(x, y), s.expected.renderer.WidgetAt(x, y)
			if expected == nil {
				require.Nil(s.t, actual, "%s: hit target at %d,%d", name, x, y)
			} else {
				require.NotNil(s.t, actual, "%s: hit target at %d,%d", name, x, y)
				require.Equal(s.t, expected.ID, actual.ID, "%s: hit target at %d,%d", name, x, y)
				require.Equal(s.t, expected.Bounds, actual.Bounds, "%s: hit bounds at %d,%d", name, x, y)
			}
		}
	}
	return work
}

// describeMismatches lists the first differing cells so failures can be read
// without opening the generated images.
func describeMismatches(expected, actual *uv.Buffer, width, height int) string {
	var b strings.Builder
	count := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			e, a := expected.CellAt(x, y), actual.CellAt(x, y)
			if cellsEqual(e, a) {
				continue
			}
			if count++; count > 8 {
				fmt.Fprintf(&b, "  ...\n")
				return b.String()
			}
			fmt.Fprintf(&b, "  (%d,%d) expected %s, actual %s\n", x, y, describeCell(e), describeCell(a))
		}
	}
	return b.String()
}

func describeCell(c *uv.Cell) string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q fg=%v bg=%v attrs=%v", c.Content, c.Style.Fg, c.Style.Bg, c.Style.Attrs)
}

func (s *reactivitySequence[T]) resize(width, height int) {
	for _, side := range []*reactivitySurface[T]{s.actual, s.expected} {
		side.buffer.Resize(width, height)
		side.renderer.Resize(width, height)
	}
}

func (s *reactivitySequence[T]) focus(id string) {
	for _, side := range []*reactivitySurface[T]{s.actual, s.expected} {
		side.focus.FocusByID(id)
		side.focused.Set(side.focus.Focused())
	}
}

type reactivityBuildScene struct {
	value Signal[string]
}

func (s *reactivityBuildScene) Build(BuildContext) Widget {
	return Row{Children: []Widget{
		reactivityBuildText{value: s.value},
		Text{ID: "neighbor", Content: "unchanged", Style: Style{Width: Cells(12), Height: Cells(1)}},
	}}
}

func TestReactivityBuildIsolation(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 3, func() *reactivityBuildScene {
		return &reactivityBuildScene{value: NewSignal("before")}
	})
	sequence.frame("Initial", nil)
	work := sequence.frame("Only the left widget changes", func(s *reactivityBuildScene) { s.value.Set("after") })
	require.Equal(t, 1, work.BuildCount, "a leaf's build dependency must not rebuild its parent or sibling")
	require.Equal(t, "reflow", work.FrameMode)
	require.Equal(t, 2, work.LayoutCount, "the unchanged sibling reuses its cached layout")
	require.Equal(t, []Rect{{X: 0, Y: 0, Width: 12, Height: 1}}, work.DamagedRects, "only the rebuilt leaf is repainted")
	sequence.frame("Shorter text clears the old value", func(s *reactivityBuildScene) { s.value.Set("x") })
}

type reactivityFocusScene struct {
	checkbox *CheckboxState
	input    *TextInputState
}

func (s *reactivityFocusScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Button{ID: "first", Label: "First"},
		Button{ID: "second", Label: "Second"},
		&Checkbox{ID: "third", State: s.checkbox, Label: "Third"},
		TextInput{ID: "fourth", State: s.input},
	}}
}

func TestReactivityFocusChangeRestylesWidgets(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 6, func() *reactivityFocusScene {
		return &reactivityFocusScene{checkbox: NewCheckboxState(false), input: NewTextInputState("text")}
	})
	sequence.frame("Initial", nil)
	for _, id := range []string{"first", "second", "third", "fourth", "first"} {
		sequence.focus(id)
		sequence.frame("Focus "+id, nil)
	}
}

// reactivityBuilder subscribes only itself to whatever its build function reads.
type reactivityBuilder struct {
	ID    string
	build func(BuildContext) Widget
}

func (b reactivityBuilder) WidgetID() string              { return b.ID }
func (b reactivityBuilder) Build(ctx BuildContext) Widget { return b.build(ctx) }

type reactivityOverlayScene struct {
	open    Signal[bool]
	counter Signal[int]
}

func (s *reactivityOverlayScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "dialog-owner", build: func(BuildContext) Widget {
			// A widget returned directly from Build is never built itself, so
			// the Dialog must be a child for its Build to register the overlay.
			return Column{Children: []Widget{Dialog{
				ID: "dialog", Visible: s.open.Get(), Title: "Confirm",
				Content: Text{Content: "Proceed?"},
				Buttons: []Button{{Label: "No"}, {Label: "Yes"}},
			}}}
		}},
		reactivityBuilder{ID: "counter", build: func(BuildContext) Widget {
			return Text{Content: fmt.Sprintf("count %d", s.counter.Get())}
		}},
	}}
}

func TestReactivityOverlayOwnedByCleanNode(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 12, func() *reactivityOverlayScene {
		return &reactivityOverlayScene{open: NewSignal(false), counter: NewSignal(0)}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Open dialog", func(s *reactivityOverlayScene) { s.open.Set(true) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "Proceed?")
	work := sequence.frame("Unrelated change keeps dialog", func(s *reactivityOverlayScene) { s.counter.Set(1) })
	require.Equal(t, 1, work.BuildCount, "only the counter rebuilds; the dialog's owner and content are reused")
	require.Contains(t, sequence.actual.renderer.ScreenText(), "Proceed?")
	sequence.focus("dialog-btn-1")
	sequence.frame("Focus moves within dialog", nil)
	sequence.frame("Close dialog", func(s *reactivityOverlayScene) { s.open.Set(false) })
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Proceed?")
	sequence.frame("Unrelated change after close", func(s *reactivityOverlayScene) { s.counter.Set(2) })
}

type reactivityReflowScene struct {
	label Signal[string]
}

func (s *reactivityReflowScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Row{Spacing: 1, Children: []Widget{
			reactivityBuilder{ID: "label", build: func(BuildContext) Widget {
				return Text{Content: s.label.Get(), Style: Style{Padding: EdgeInsetsXY(1, 0)}}
			}},
			Text{ID: "neighbor", Content: "neighbor"},
		}},
		Row{Spacing: 1, Children: []Widget{
			// The Text is a child node rebuilt with new content by its parent;
			// it has no signal of its own, so nothing else marks it changed.
			reactivityBuilder{ID: "wrapped-label", build: func(BuildContext) Widget {
				return Column{Children: []Widget{Text{Content: s.label.Get()}}}
			}},
			Text{ID: "wrapped-neighbor", Content: "neighbor"},
		}},
		Text{Content: "below", Style: Style{Width: Flex(1)}},
	}}
}

func TestReactivityAutoSizeReflowsSiblings(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 5, func() *reactivityReflowScene {
		return &reactivityReflowScene{label: NewSignal("short")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Grow pushes neighbor right", func(s *reactivityReflowScene) { s.label.Set("a much longer label") })
	sequence.frame("Shrink pulls neighbor left", func(s *reactivityReflowScene) { s.label.Set("x") })
	sequence.frame("Empty", func(s *reactivityReflowScene) { s.label.Set("") })
	sequence.resize(12, 5)
	sequence.frame("Resize narrower", nil)
	sequence.frame("Grow after resize", func(s *reactivityReflowScene) { s.label.Set("wider than the screen") })
}

type reactivityKeyedScene struct {
	keys  AnySignal[[]string]
	title Signal[string]
}

func (s *reactivityKeyedScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "title", build: func(BuildContext) Widget { return Text{Content: s.title.Get()} }},
		reactivityBuilder{ID: "items", build: func(BuildContext) Widget {
			keys := s.keys.Get()
			children := make([]Widget, len(keys))
			for i, key := range keys {
				children[i] = Button{ID: "item-" + key, Label: key}
			}
			return Column{Children: children}
		}},
	}}
}

func TestReactivityKeyedChildrenReorderAndRemove(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 8, func() *reactivityKeyedScene {
		return &reactivityKeyedScene{keys: NewAnySignal([]string{"a", "b", "c"}), title: NewSignal("items")}
	})
	sequence.frame("Initial", nil)
	sequence.focus("item-b")
	sequence.frame("Focus b", nil)
	sequence.frame("Reverse", func(s *reactivityKeyedScene) { s.keys.Set([]string{"c", "b", "a"}) })
	sequence.frame("Title change keeps order", func(s *reactivityKeyedScene) { s.title.Set("renamed") })
	sequence.frame("Remove focused", func(s *reactivityKeyedScene) { s.keys.Set([]string{"c", "a"}) })
	sequence.frame("Add items", func(s *reactivityKeyedScene) { s.keys.Set([]string{"d", "c", "e", "a"}) })
}

type reactivityListScene struct {
	state  *ListState[string]
	scroll *ScrollState
	custom bool
	multi  bool
	header Signal[string]
}

func (s *reactivityListScene) list() List[string] {
	list := List[string]{ID: "list", State: s.state, ScrollState: s.scroll, MultiSelect: s.multi}
	if s.custom {
		list.RenderItem = func(item string, active, selected bool) Widget {
			prefix := "  "
			if active {
				prefix = "> "
			} else if selected {
				prefix = "* "
			}
			return Text{Content: prefix + item}
		}
	}
	return list
}

func (s *reactivityListScene) Build(BuildContext) Widget {
	list := s.list()
	return Column{Children: []Widget{
		reactivityBuilder{ID: "header", build: func(BuildContext) Widget { return Text{Content: s.header.Get()} }},
		Scrollable{ID: "scroll", State: s.scroll, Height: Cells(4), Child: list},
	}}
}

func TestReactivityListCursorAndScroll(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 20, 6, func() *reactivityListScene {
				items := []string{"one", "two", "three", "four", "five", "six", "seven"}
				return &reactivityListScene{state: NewListState(items), scroll: NewScrollState(), custom: custom, header: NewSignal("list")}
			})
			sequence.frame("Initial", nil)
			sequence.focus("list")
			sequence.frame("Focus list", nil)
			for i := 0; i < 5; i++ {
				// Drive the list through its own keybinding, as a user would.
				work := sequence.frame(fmt.Sprintf("Cursor down %d", i+1), func(s *reactivityListScene) { s.list().keyCursorDown() })
				if i == 0 {
					// Within the viewport, only the old and new cursor rows change.
					require.LessOrEqual(t, damagedRows(work), 2, "damage covers at most the two affected rows: %v", work.DamagedRects)
					if custom {
						require.LessOrEqual(t, work.BuildCount, 4, "only the two affected rows rebuild")
					} else {
						require.Zero(t, work.BuildCount)
					}
				}
			}
			sequence.frame("Header change", func(s *reactivityListScene) { s.header.Set("changed") })
			sequence.frame("Replace items", func(s *reactivityListScene) { s.state.SetItems([]string{"alpha", "beta"}) })
		})
	}
}

// Moving the cursor through ListState from app code keeps it visible, just as
// the list's own keybindings do.
func TestReactivityListProgrammaticCursorScrolls(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 6, func() *reactivityListScene {
		items := []string{"one", "two", "three", "four", "five", "six", "seven"}
		return &reactivityListScene{state: NewListState(items), scroll: NewScrollState(), header: NewSignal("list")}
	})
	sequence.frame("Initial", nil)
	for i := 0; i < 5; i++ {
		sequence.frame(fmt.Sprintf("SelectNext %d", i+1), func(s *reactivityListScene) { s.state.SelectNext() })
	}
	sequence.frame("SelectFirst jumps above the viewport", func(s *reactivityListScene) { s.state.SelectFirst() })
	sequence.frame("SelectLast jumps below it", func(s *reactivityListScene) { s.state.SelectLast() })
	for i := 0; i < 5; i++ {
		sequence.frame(fmt.Sprintf("SelectPrevious %d", i+1), func(s *reactivityListScene) { s.state.SelectPrevious() })
	}
	sequence.frame("SelectIndex below the viewport", func(s *reactivityListScene) { s.state.SelectIndex(5) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "six", "the selected item is on screen")
}

type reactivityKeybindScene struct {
	list *ListState[string]
}

func (s *reactivityKeybindScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Button{ID: "button", Label: "Button"},
		List[string]{ID: "list", State: s.list},
		KeybindBar{},
	}}
}

func TestReactivityKeybindBarFollowsFocus(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 5, func() *reactivityKeybindScene {
		return &reactivityKeybindScene{list: NewListState([]string{"one", "two"})}
	})
	sequence.frame("Initial", nil)
	sequence.focus("list")
	sequence.frame("Focus list shows list keybinds", nil)
	sequence.focus("button")
	sequence.frame("Focus button hides list keybinds", nil)
}

type reactivityScrollGrowthScene struct {
	count  Signal[int]
	scroll *ScrollState
}

func (s *reactivityScrollGrowthScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Text{Content: "header"},
		Scrollable{ID: "scroll", State: s.scroll, Height: Cells(4), Child: reactivityBuilder{ID: "rows", build: func(BuildContext) Widget {
			rows := make([]Widget, s.count.Get())
			for i := range rows {
				// Narrow rows leave the scrollbar column outside their damage.
				rows[i] = Text{Content: fmt.Sprintf("row %d", i), Style: Style{Width: Cells(6)}}
			}
			return Column{Children: rows}
		}}},
	}}
}

// The scrollbar is painted by the Scrollable, which is neither rebuilt nor
// moved when its content grows; only its layout box changes.
func TestReactivityScrollbarFollowsContentGrowth(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 6, func() *reactivityScrollGrowthScene {
		return &reactivityScrollGrowthScene{count: NewSignal(2), scroll: NewScrollState()}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Overflow shows scrollbar", func(s *reactivityScrollGrowthScene) { s.count.Set(8) })
	sequence.frame("More content shrinks thumb", func(s *reactivityScrollGrowthScene) { s.count.Set(20) })
	sequence.frame("Scroll down", func(s *reactivityScrollGrowthScene) { s.scroll.ScrollDown(3) })
	sequence.frame("Shrink removes scrollbar", func(s *reactivityScrollGrowthScene) { s.count.Set(1) })
}

type reactivityOverflowScene struct {
	badge Signal[string]
	body  Signal[string]
}

func (s *reactivityOverflowScene) Build(BuildContext) Widget {
	// The Stack is the root so no ancestor clips the badge overflowing its right edge.
	return Stack{Children: []Widget{
		Text{Content: "card", Style: Style{Width: Cells(10), Height: Cells(3)}},
		Positioned{Top: IntPtr(0), Right: IntPtr(-6), Child: reactivityBench(s.badge)},
		Positioned{Top: IntPtr(1), Left: IntPtr(0), Child: reactivityBench(s.body)},
	}}
}

func reactivityBench(value Signal[string]) Widget {
	return SignalText(value, func(v string) string { return v })
}

// Positioned children can overflow the Stack. Repainting one child must not
// drop the overflowing sibling from the Stack's recorded painted area.
func TestReactivityStackOverflowAfterPartialRepaint(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 6, func() *reactivityOverflowScene {
		return &reactivityOverflowScene{badge: NewSignal("[1]"), body: NewSignal("aa")}
	})
	sequence.frame("Initial", nil)
	require.Contains(t, sequence.actual.renderer.ScreenText(), "[1]", "badge overflows the Stack but stays visible")
	sequence.frame("Repaint body only", func(s *reactivityOverflowScene) { s.body.Set("bb") })
	sequence.frame("Repaint overflowing badge", func(s *reactivityOverflowScene) { s.badge.Set("[2]") })
}

// reactivityRootScene changes its own identity without any parent rebuilding,
// so the renderer must notice and replace the retained node.
type reactivityRootScene struct {
	id   string
	text Signal[string]
}

func (s *reactivityRootScene) WidgetID() string { return s.id }

func (s *reactivityRootScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "text", build: func(BuildContext) Widget { return Text{Content: s.text.Get()} }},
	}}
}

func TestReactivityRootReplacementClearsOldOutput(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 3, func() *reactivityRootScene {
		return &reactivityRootScene{id: "first", text: NewSignal("a much longer line")}
	})
	sequence.frame("Initial", nil)
	work := sequence.frame("Replace root with shorter content", func(s *reactivityRootScene) {
		s.id = "second"
		s.text.Set("short")
	})
	require.Equal(t, 2, work.BuildCount, "the replaced root and its child are built afresh")
}

// reactivityLayoutReader reads its width while layout is computed (not when
// its layout node is built, which cache hits can trigger incidentally).
type reactivityLayoutReader struct {
	width Signal[int]
}

func (w reactivityLayoutReader) Build(BuildContext) Widget { return w }

func (w reactivityLayoutReader) BuildLayoutNode(BuildContext) layout.LayoutNode {
	return &layout.BoxNode{MeasureFunc: func(layout.Constraints) (int, int) { return w.width.Get(), 1 }}
}

func (w reactivityLayoutReader) Render(ctx *RenderContext) {
	ctx.DrawText(0, 0, strings.Repeat("#", ctx.Width))
}

type reactivityLayoutReadScene struct {
	width Signal[int]
	other Signal[string]
}

func (s *reactivityLayoutReadScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		// A direct child of the root, which recomputes its layout when "other"
		// changes, so the reader's cached result is consulted.
		reactivityLayoutReader{width: s.width},
		reactivityBuilder{ID: "other", build: func(BuildContext) Widget { return Text{Content: s.other.Get()} }},
	}}
}

// A layout-phase subscription must survive frames where the reader's layout
// is reused from the cache and its reads don't run.
func TestReactivityLayoutReadSurvivesCachedFrames(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 3, func() *reactivityLayoutReadScene {
		return &reactivityLayoutReadScene{width: NewSignal(3), other: NewSignal("one")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Unrelated change reuses reader layout", func(s *reactivityLayoutReadScene) { s.other.Set("two") })
	sequence.frame("Reader width grows", func(s *reactivityLayoutReadScene) { s.width.Set(8) })
}

func TestReactivityListMultiSelect(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 20, 6, func() *reactivityListScene {
				items := []string{"one", "two", "three", "four", "five", "six", "seven"}
				return &reactivityListScene{state: NewListState(items), scroll: NewScrollState(), custom: custom, multi: true, header: NewSignal("list")}
			})
			sequence.frame("Initial", nil)
			sequence.focus("list")
			sequence.frame("Focus list", nil)
			sequence.frame("Extend selection", func(s *reactivityListScene) { s.list().shiftCursorDown() })
			sequence.frame("Extend again", func(s *reactivityListScene) { s.list().shiftCursorDown() })
			sequence.frame("Shrink selection", func(s *reactivityListScene) { s.list().shiftCursorUp() })
			sequence.frame("Plain move clears selection", func(s *reactivityListScene) { s.list().keyCursorDown() })
			sequence.frame("Select past viewport", func(s *reactivityListScene) {
				for i := 0; i < 4; i++ {
					s.list().shiftCursorDown()
				}
			})
		})
	}
}

// reactivitySwappable is a retained pointer widget that can change its own ID.
type reactivitySwappable struct {
	id   string
	text Signal[string]
}

func (w *reactivitySwappable) WidgetID() string { return w.id }

func (w *reactivitySwappable) Build(BuildContext) Widget {
	return Text{Content: w.text.Get()}
}

type reactivitySwapScene struct {
	child *reactivitySwappable
}

func (s *reactivitySwapScene) Build(BuildContext) Widget {
	return Row{Spacing: 1, Children: []Widget{s.child, Text{Content: "neighbor"}}}
}

// Replacing a child beneath a clean parent must keep it connected: its later
// signal changes have to reach the renderer, and the parent can't reuse a
// layout computed for the old child.
func TestReactivityReplacedChildStaysConnected(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 2, func() *reactivitySwapScene {
		return &reactivitySwapScene{child: &reactivitySwappable{id: "a", text: NewSignal("short")}}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Child replaces itself with wider content", func(s *reactivitySwapScene) {
		s.child.id = "b"
		s.child.text.Set("much wider text")
	})
	sequence.frame("Replaced child's own signal changes", func(s *reactivitySwapScene) {
		s.child.text.Set("x")
	})
}

// reactivityPlainSwappable changes its ID and content without any signal, so
// only the renderer noticing the new identity can pick the change up.
type reactivityPlainSwappable struct {
	id      string
	content string
}

func (w *reactivityPlainSwappable) WidgetID() string          { return w.id }
func (w *reactivityPlainSwappable) Build(BuildContext) Widget { return Text{Content: w.content} }

type reactivityPlainSwapScene struct {
	child *reactivityPlainSwappable
	other Signal[string]
}

func (s *reactivityPlainSwapScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Row{Spacing: 1, Children: []Widget{s.child, Text{Content: "neighbor"}}},
		reactivityBuilder{ID: "other", build: func(BuildContext) Widget { return Text{Content: s.other.Get()} }},
	}}
}

func TestReactivityReplacementDuringUnrelatedUpdate(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 3, func() *reactivityPlainSwapScene {
		return &reactivityPlainSwapScene{child: &reactivityPlainSwappable{id: "a", content: "short"}, other: NewSignal("one")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Unrelated update while the child replaces itself wider", func(s *reactivityPlainSwapScene) {
		s.child.id, s.child.content = "b", "much wider text"
		s.other.Set("two")
	})
}

type reactivityWideScene struct {
	value Signal[string]
	paint Signal[string]
}

// Ten rows of ten leaves, one build-read leaf and one paint-read leaf.
func (s *reactivityWideScene) Build(BuildContext) Widget {
	rows := make([]Widget, 10)
	for r := range rows {
		cells := make([]Widget, 10)
		for c := range cells {
			cells[c] = Text{Content: "steady", Style: Style{Width: Cells(7)}}
		}
		if r == 4 {
			cells[3] = reactivityBuildText{value: s.value}
			cells[6] = reactivityBench(s.paint)
		}
		rows[r] = Row{Children: cells}
	}
	return Column{Children: rows}
}

// Work after a one-leaf change must not scale with the rest of the tree.
func TestReactivityPassesSkipCleanSubtrees(t *testing.T) {
	sequence := newReactivitySequence(t, 80, 12, func() *reactivityWideScene {
		return &reactivityWideScene{value: NewSignal("before"), paint: NewSignal("aa")}
	})
	sequence.frame("Initial", nil)
	r := sequence.actual.renderer

	sequence.frame("Build-read leaf changes", func(s *reactivityWideScene) { s.value.Set("after") })
	require.LessOrEqual(t, r.lastAssignCount, 4, "layouts assigned: root, its row, the leaf")
	require.LessOrEqual(t, r.lastMeasureCount, 15, "nodes measured: the changed row's cells and its ancestors")
	require.LessOrEqual(t, r.lastClearCount, 4, "dirty flags cleared along the changed path only")

	sequence.frame("Paint-read leaf changes", func(s *reactivityWideScene) { s.paint.Set("bb") })
	require.LessOrEqual(t, r.lastScanCount, 4, "damage scan follows the dirty path only")
	require.LessOrEqual(t, r.lastClearCount, 4)
}

type reactivityOverlapScene struct {
	under Signal[string]
	over  Signal[string]
}

// Buttons overlap inside a Stack: click targets depend on paint order, which
// must survive one of them being skipped while the other changes.
func (s *reactivityOverlapScene) Build(BuildContext) Widget {
	return Stack{Children: []Widget{
		reactivityBuilder{ID: "under", build: func(BuildContext) Widget {
			return Column{Children: []Widget{Button{ID: "under-btn", Label: s.under.Get()}}}
		}},
		Positioned{Top: IntPtr(0), Left: IntPtr(3), Child: reactivityBuilder{ID: "over", build: func(BuildContext) Widget {
			return Column{Children: []Widget{Button{ID: "over-btn", Label: s.over.Get()}}}
		}}},
	}}
}

func TestReactivityClickTargetsKeepPaintOrder(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 2, func() *reactivityOverlapScene {
		return &reactivityOverlapScene{under: NewSignal("underneath"), over: NewSignal("top")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Lower button changes, upper skipped", func(s *reactivityOverlapScene) { s.under.Set("below it") })
	sequence.frame("Upper button changes, lower skipped", func(s *reactivityOverlapScene) { s.over.Set("upper") })
}

type reactivityShiftScene struct {
	lead  Signal[string]
	share Signal[int]
}

// A clean nested subtree is moved by a growing sibling, then resized by a
// change in flexible space, without any change inside it.
func (s *reactivityShiftScene) Build(BuildContext) Widget {
	cleanBlock := Column{Children: []Widget{
		Button{ID: "inner-a", Label: "A"},
		Row{Children: []Widget{Text{Content: "deep", Style: Style{Width: Flex(1)}}, Button{ID: "inner-b", Label: "B"}}},
	}}
	return Column{Children: []Widget{
		reactivityBuilder{ID: "lead", build: func(BuildContext) Widget {
			return Text{Content: s.lead.Get(), Wrap: WrapSoft}
		}},
		Row{Children: []Widget{
			reactivityBuilder{ID: "grower", build: func(BuildContext) Widget {
				return Text{Content: strings.Repeat("=", s.share.Get()), Style: Style{Width: Cells(s.share.Get())}}
			}},
			Column{Style: Style{Width: Flex(1)}, Children: []Widget{cleanBlock}},
		}},
	}}
}

func TestReactivityCleanSubtreeMovesAndResizes(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 6, func() *reactivityShiftScene {
		return &reactivityShiftScene{lead: NewSignal("one line"), share: NewSignal(4)}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Lead wraps to two lines, pushing the block down", func(s *reactivityShiftScene) {
		s.lead.Set("this lead text now wraps")
	})
	sequence.frame("Sibling widens, shrinking the block", func(s *reactivityShiftScene) { s.share.Set(9) })
	sequence.frame("Lead shrinks back, pulling the block up", func(s *reactivityShiftScene) { s.lead.Set("short") })
}

type reactivityShiftedTargetsScene struct {
	count Signal[int]
}

// A clean block keeps its place while a sibling before it in paint order
// records a changing number of click targets, shifting where the block's
// entries land in the hit-test registry.
func (s *reactivityShiftedTargetsScene) Build(BuildContext) Widget {
	return Row{Children: []Widget{
		reactivityBuilder{ID: "growing", build: func(BuildContext) Widget {
			buttons := make([]Widget, s.count.Get())
			for i := range buttons {
				buttons[i] = Button{ID: fmt.Sprintf("grow-%d", i), Label: "g"}
			}
			return Column{Style: Style{Width: Cells(5)}, Children: buttons}
		}},
		Column{Children: []Widget{Button{ID: "stay-a", Label: "A"}, Button{ID: "stay-b", Label: "B"}}},
	}}
}

func TestReactivitySkippedClickTargetsAfterShift(t *testing.T) {
	sequence := newReactivitySequence(t, 20, 5, func() *reactivityShiftedTargetsScene {
		return &reactivityShiftedTargetsScene{count: NewSignal(1)}
	})
	sequence.frame("Initial", nil)
	sequence.frame("More targets before the block", func(s *reactivityShiftedTargetsScene) { s.count.Set(4) })
	sequence.frame("Fewer targets before the block", func(s *reactivityShiftedTargetsScene) { s.count.Set(2) })
	sequence.frame("None before the block", func(s *reactivityShiftedTargetsScene) { s.count.Set(0) })
}

type reactivityCollapseScene struct {
	state *TreeState[string]
}

func (s *reactivityCollapseScene) tree() Tree[string] {
	return Tree[string]{ID: "tree", State: s.state}
}

// The tree is a child: a widget returned directly from Build isn't built.
func (s *reactivityCollapseScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{s.tree()}}
}

func TestReactivityTreeCollapseAndExpand(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 8, func() *reactivityCollapseScene {
		roots := []TreeNode[string]{
			{Data: "alpha", Children: []TreeNode[string]{{Data: "a1"}, {Data: "a2"}}},
			{Data: "beta", Children: []TreeNode[string]{{Data: "b1"}}},
			{Data: "gamma"},
		}
		return &reactivityCollapseScene{state: NewTreeState(roots)}
	})
	sequence.frame("Initial", nil)
	sequence.focus("tree")
	sequence.frame("Focus tree", nil)
	sequence.frame("Collapse alpha", func(s *reactivityCollapseScene) { s.tree().collapseOrMoveToParent() })
	sequence.frame("Expand alpha", func(s *reactivityCollapseScene) { s.tree().expandOrMoveToChild() })
	sequence.frame("Collapse alpha again", func(s *reactivityCollapseScene) { s.tree().collapseOrMoveToParent() })
	sequence.frame("Down to beta", func(s *reactivityCollapseScene) { s.tree().keyCursorDown() })
	sequence.frame("Collapse beta", func(s *reactivityCollapseScene) { s.tree().collapseOrMoveToParent() })
	sequence.frame("Toggle beta", func(s *reactivityCollapseScene) { s.tree().toggleExpansion() })
}

// damagedRows counts distinct screen rows touched by a frame's damage. A full
// frame repaints everything and records no rectangles.
func damagedRows(work RenderStats) int {
	if work.FrameMode == string(rendererFrameFull) {
		return math.MaxInt
	}
	rows := map[int]bool{}
	for _, rect := range work.DamagedRects {
		for y := rect.Y; y < rect.Y+rect.Height; y++ {
			rows[y] = true
		}
	}
	return len(rows)
}

type reactivityTableScene struct {
	state  *TableState[[]string]
	mode   TableSelectionMode
	custom bool
	header Signal[string]
}

func (s *reactivityTableScene) table() Table[[]string] {
	cols := []TableColumn{{Width: Cells(6)}, {Width: Cells(6)}, {Width: Cells(6)}}
	table := Table[[]string]{ID: "table", State: s.state, Columns: cols, SelectionMode: s.mode, MultiSelect: true}
	if s.custom {
		table.RenderCell = func(row []string, rowIndex, colIndex int, active, selected bool) Widget {
			marker := " "
			if active {
				marker = ">"
			} else if selected {
				marker = "*"
			}
			return Text{Content: marker + row[colIndex]}
		}
	}
	return table
}

func (s *reactivityTableScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "header", build: func(BuildContext) Widget { return Text{Content: s.header.Get()} }},
		s.table(),
	}}
}

func TestReactivityTableCursorAndSelection(t *testing.T) {
	modes := map[string]TableSelectionMode{"cursor": TableSelectionCursor, "row": TableSelectionRow, "column": TableSelectionColumn}
	for name, mode := range modes {
		for _, custom := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/custom=%v", name, custom), func(t *testing.T) {
				sequence := newReactivitySequence(t, 24, 9, func() *reactivityTableScene {
					rows := make([][]string, 6)
					for i := range rows {
						rows[i] = []string{fmt.Sprintf("a%d", i), fmt.Sprintf("b%d", i), fmt.Sprintf("c%d", i)}
					}
					return &reactivityTableScene{state: NewTableState(rows), mode: mode, custom: custom, header: NewSignal("table")}
				})
				sequence.frame("Initial", nil)
				sequence.focus("table")
				sequence.frame("Focus table", nil)

				down := sequence.frame("Cursor down", func(s *reactivityTableScene) { s.table().keyCursorDown() })
				right := sequence.frame("Cursor right", func(s *reactivityTableScene) { s.table().keyCursorRight() })
				switch mode {
				case TableSelectionCursor, TableSelectionRow:
					require.LessOrEqual(t, damagedRows(down), 2, "a vertical move repaints the two affected rows")
				}
				if mode == TableSelectionCursor {
					require.LessOrEqual(t, damagedRows(right), 1, "a horizontal cell move repaints one row")
				}
				if custom {
					// Cells watch their row first, so a vertical move in cursor mode
					// rebuilds the old and new rows' cells (each a wrapper and a
					// child), and a horizontal move just the two cells.
					limits := map[TableSelectionMode]int{TableSelectionCursor: 12, TableSelectionRow: 12, TableSelectionColumn: 0}
					require.LessOrEqual(t, down.BuildCount, limits[mode], "only affected cells rebuild on a vertical move")
					if mode == TableSelectionCursor {
						require.LessOrEqual(t, right.BuildCount, 4, "only the two affected cells rebuild on a horizontal move")
					}
				} else {
					require.Zero(t, down.BuildCount)
					require.Zero(t, right.BuildCount)
				}

				sequence.frame("Extend selection", func(s *reactivityTableScene) {
					switch s.mode {
					case TableSelectionRow:
						s.table().shiftRowDown()
					case TableSelectionColumn:
						s.table().shiftColumnRight()
					default:
						s.table().shiftCellDown()
					}
				})
				// App code clears the selection without moving the cursor.
				sequence.frame("Selection cleared directly", func(s *reactivityTableScene) { s.state.ClearSelection() })
				sequence.frame("Plain move", func(s *reactivityTableScene) { s.table().keyCursorUp() })
				sequence.frame("Unrelated header change", func(s *reactivityTableScene) { s.header.Set("changed") })
				sequence.frame("Cursor to last", func(s *reactivityTableScene) { s.table().keyCursorToLast() })
			})
		}
	}
}

type reactivityTreeScene struct {
	state  *TreeState[string]
	custom bool
}

func (s *reactivityTreeScene) tree() Tree[string] {
	tree := Tree[string]{ID: "tree", State: s.state, MultiSelect: true}
	if s.custom {
		tree.RenderNode = func(node string, ctx TreeNodeContext) Widget {
			marker := " "
			if ctx.Active {
				marker = ">"
			} else if ctx.Selected {
				marker = "*"
			}
			return Text{Content: marker + node}
		}
	}
	return tree
}

// The tree is a child: a widget returned directly from Build isn't built.
func (s *reactivityTreeScene) Build(BuildContext) Widget { return Column{Children: []Widget{s.tree()}} }

func TestReactivityTreeCursorAndSelection(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 24, 9, func() *reactivityTreeScene {
				roots := []TreeNode[string]{
					{Data: "alpha", Children: []TreeNode[string]{{Data: "a1"}, {Data: "a2"}}},
					{Data: "beta", Children: []TreeNode[string]{{Data: "b1"}}},
					{Data: "gamma"},
					{Data: "delta"},
				}
				return &reactivityTreeScene{state: NewTreeState(roots), custom: custom}
			})
			sequence.frame("Initial", nil)
			sequence.focus("tree")
			sequence.frame("Focus tree", nil)
			down := sequence.frame("Cursor down", func(s *reactivityTreeScene) { s.tree().keyCursorDown() })
			require.LessOrEqual(t, damagedRows(down), 2, "a move repaints the two affected rows")
			if custom {
				require.LessOrEqual(t, down.BuildCount, 6, "only the two affected rows rebuild (row, prefix, node)")
			} else {
				require.Zero(t, down.BuildCount)
			}
			sequence.frame("Cursor up", func(s *reactivityTreeScene) { s.tree().keyCursorUp() })
			sequence.frame("Expand alpha", func(s *reactivityTreeScene) { s.tree().expandOrMoveToChild() })
			sequence.frame("Into child", func(s *reactivityTreeScene) { s.tree().expandOrMoveToChild() })
			sequence.frame("Extend selection", func(s *reactivityTreeScene) { s.tree().shiftCursorDown() })
			sequence.frame("Extend again", func(s *reactivityTreeScene) { s.tree().shiftCursorDown() })
			sequence.frame("Back to parent", func(s *reactivityTreeScene) { s.tree().collapseOrMoveToParent() })
			sequence.frame("Collapse", func(s *reactivityTreeScene) { s.tree().collapseOrMoveToParent() })
			sequence.frame("Cursor to last", func(s *reactivityTreeScene) { s.tree().keyCursorToLast() })
			// App code points the cursor at a hidden node; a custom-rendered tree
			// corrects it to the first visible row.
			// beta is collapsed by now, so its child is hidden.
			sequence.frame("Cursor set to hidden node", func(s *reactivityTreeScene) { s.state.CursorPath.Set([]int{1, 0}) })
		})
	}
}

// A custom cell in a row made taller by a wrapping neighbour must still fill
// the whole row height, so a background (such as a highlight) covers it.
func TestTableCustomCellFillsTallRow(t *testing.T) {
	state := NewTableState([][]string{{"short", "this note wraps onto a second line"}})
	table := Table[[]string]{
		ID: "table", State: state,
		Columns: []TableColumn{{Width: Cells(8)}, {Width: Cells(16)}},
		RenderCell: func(row []string, _, col int, _, _ bool) Widget {
			if col == 1 {
				return Text{Content: row[col], Wrap: WrapSoft}
			}
			return Text{Content: row[col], Style: Style{BackgroundColor: RGB(200, 0, 0)}}
		},
	}
	buf := RenderToBuffer(Column{Children: []Widget{table}}, 30, 4)
	for y := 0; y < 2; y++ {
		cell := buf.CellAt(2, y)
		require.NotNil(t, cell.Style.Bg, "first column, line %d: the highlighted cell must fill the row", y)
	}
}

type reactivityFloatsScene struct {
	bigMenu    Signal[bool]
	nestedOpen Signal[bool]
	lead       Signal[string]
	under      Signal[string]
	menuOpen   Signal[bool]
	menuLabel  Signal[string]
	dialogOpen Signal[bool]
	inside     Signal[string]
}

func (s *reactivityFloatsScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Row{Spacing: 1, Children: []Widget{
			reactivityBuilder{ID: "lead", build: func(BuildContext) Widget { return Text{Content: s.lead.Get()} }},
			Button{ID: "anchor", Label: "Menu"},
		}},
		reactivityBuilder{ID: "menu-owner", build: func(BuildContext) Widget {
			return Column{Children: []Widget{Floating{
				Visible: s.menuOpen.Get(),
				Config:  FloatConfig{AnchorID: "anchor", Anchor: AnchorBottomLeft},
				Child:   s.menuContent(),
			}}}
		}},
		reactivityBuilder{ID: "dialog-owner", build: func(BuildContext) Widget {
			return Column{Children: []Widget{Dialog{
				ID: "dialog", Visible: s.dialogOpen.Get(), Title: "Dialog",
				Content: reactivityBuilder{ID: "dialog-inside", build: func(BuildContext) Widget { return Text{Content: "inside: " + s.inside.Get()} }},
				Buttons: []Button{{Label: "No"}, {Label: "Yes"}},
			}}}
		}},
		reactivityBuilder{ID: "under", build: func(BuildContext) Widget {
			return Text{Content: s.under.Get(), Style: Style{Width: Flex(1), BackgroundColor: RGB(20, 60, 20)}}
		}},
	}}
}

// menuContent is either a large or a small menu with different IDs, so
// switching replaces the overlay's root node. The large one can open a nested
// overlay of its own.
func (s *reactivityFloatsScene) menuContent() Widget {
	if !s.bigMenu.Get() {
		return Column{ID: "small-menu", Style: Style{BackgroundColor: RGB(90, 40, 40)}, Children: []Widget{Text{Content: "small"}}}
	}
	return Column{ID: "big-menu", Style: Style{BackgroundColor: RGB(40, 40, 90)}, Children: []Widget{
		Button{ID: "menu-item", Label: s.menuLabel.Get()},
		reactivityBuilder{ID: "menu-inside", build: func(BuildContext) Widget { return Text{Content: s.inside.Get()} }},
		reactivityBuilder{ID: "nested-owner", build: func(BuildContext) Widget {
			return Column{Children: []Widget{Floating{
				Visible: s.nestedOpen.Get(),
				// Modal, so opening it dims the whole screen: only discovered
				// while placing overlays, as it is registered by overlay content.
				Config: FloatConfig{AnchorID: "menu-item", Anchor: AnchorBottomLeft, Offset: Offset{X: 12}, Modal: true},
				// Nothing focusable inside, so focus doesn't move and trigger
				// a second, full render that would hide a missed repaint.
				Child: Text{Content: "nested overlay", Style: Style{BackgroundColor: RGB(90, 90, 40)}},
			}}}
		}},
	}}
}

// Frames with overlays repaint only damage while the set of overlays is unchanged.
func TestReactivityOverlaysRepaintOnlyDamage(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 14, func() *reactivityFloatsScene {
		return &reactivityFloatsScene{
			bigMenu: NewSignal(true), nestedOpen: NewSignal(false),
			lead: NewSignal("x"), under: NewSignal("under one"), menuOpen: NewSignal(false),
			menuLabel: NewSignal("Item"), dialogOpen: NewSignal(false), inside: NewSignal("a"),
		}
	})
	screen := 40 * 14
	requirePartial := func(work RenderStats, what string) {
		t.Helper()
		require.Equal(t, string(rendererFrameReflow), work.FrameMode, "%s: overlays alone shouldn't force a full repaint", what)
		area := 0
		for _, rect := range work.DamagedRects {
			area += rect.Width * rect.Height
		}
		require.Less(t, area, screen, "%s: damage %v", what, work.DamagedRects)
	}

	sequence.frame("Initial", nil)
	sequence.frame("Open menu", func(s *reactivityFloatsScene) { s.menuOpen.Set(true) })
	underMenu := sequence.frame("Change under the menu", func(s *reactivityFloatsScene) { s.under.Set("under two") })
	requirePartial(underMenu, "under menu")
	require.Equal(t, 1, underMenu.BuildCount, "the menu's owner reused its build, so the menu isn't rebuilt")
	requirePartial(sequence.frame("Change inside the menu", func(s *reactivityFloatsScene) { s.inside.Set("bb") }), "inside menu")
	requirePartial(sequence.frame("Anchor moves, menu follows", func(s *reactivityFloatsScene) { s.lead.Set("a longer lead") }), "anchor moves")
	requirePartial(sequence.frame("Menu owner rebuilds its content", func(s *reactivityFloatsScene) { s.menuLabel.Set("Renamed item") }), "menu content")
	sequence.frame("Anchor moves back", func(s *reactivityFloatsScene) { s.lead.Set("x") })
	sequence.frame("Nested overlay opens inside the menu", func(s *reactivityFloatsScene) { s.nestedOpen.Set(true) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "nested overlay")
	sequence.frame("Nested overlay closes", func(s *reactivityFloatsScene) { s.nestedOpen.Set(false) })
	requirePartial(sequence.frame("Menu content replaced by a smaller one", func(s *reactivityFloatsScene) { s.bigMenu.Set(false) }), "replaced root")
	sequence.frame("Menu content restored", func(s *reactivityFloatsScene) { s.bigMenu.Set(true) })

	sequence.frame("Open dialog", func(s *reactivityFloatsScene) { s.dialogOpen.Set(true) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "inside: bb")
	requirePartial(sequence.frame("Change inside both overlays", func(s *reactivityFloatsScene) { s.inside.Set("ccc") }), "inside dialog")
	requirePartial(sequence.frame("Change under the backdrop", func(s *reactivityFloatsScene) { s.under.Set("under three") }), "under backdrop")
	sequence.focus("dialog-btn-1")
	sequence.frame("Focus within dialog", nil)
	sequence.frame("Close dialog", func(s *reactivityFloatsScene) { s.dialogOpen.Set(false) })
	sequence.frame("Close menu", func(s *reactivityFloatsScene) { s.menuOpen.Set(false) })
	requirePartial(sequence.frame("Change with no overlays", func(s *reactivityFloatsScene) { s.under.Set("under four") }), "no overlays")
}
