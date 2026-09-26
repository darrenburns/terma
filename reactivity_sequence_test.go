package terma

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
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
	s.actual.draw(false)
	s.expected.draw(true)
	width, height := s.actual.buffer.Width(), s.actual.buffer.Height()
	stats := CompareBuffers(s.expected.buffer, s.actual.buffer, width, height)
	opts := DefaultSVGOptions()
	work := s.actual.renderer.Stats()
	s.frames = append(s.frames, SnapshotComparison{
		Name: name, Passed: stats.MismatchedCells == 0, Stats: stats,
		Description: fmt.Sprintf("Incremental: %s, builds=%d, layouts=%d, paints=%d. Expected: forced full render.", work.FrameMode, work.BuildCount, work.LayoutCount, work.PaintCount),
		Expected:    BufferToSVG(s.expected.buffer, width, height, opts),
		Actual:      BufferToSVG(s.actual.buffer, width, height, opts),
		DiffSVG:     GenerateDiffSVG(s.expected.buffer, s.actual.buffer, width, height, opts),
	})
	require.Zero(s.t, stats.MismatchedCells, "%s: incremental output differs from full render:\n%s", name, describeMismatches(s.expected.buffer, s.actual.buffer, width, height))
	require.Equal(s.t, s.expected.focus.FocusedID(), s.actual.focus.FocusedID(), "%s: focused widget", name)
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
	require.Equal(t, 1, work.BuildCount-reactivityFloatBuilds(sequence.actual.renderer), "only the counter rebuilds in the main tree")
	require.Contains(t, sequence.actual.renderer.ScreenText(), "Proceed?")
	sequence.focus("dialog-btn-1")
	sequence.frame("Focus moves within dialog", nil)
	sequence.frame("Close dialog", func(s *reactivityOverlayScene) { s.open.Set(false) })
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Proceed?")
	sequence.frame("Unrelated change after close", func(s *reactivityOverlayScene) { s.counter.Set(2) })
}

// Float roots are rebuilt every frame, so exclude them from main-tree counts.
func reactivityFloatBuilds(r *Renderer) int {
	count := 0
	var walk func(*widgetNode)
	walk = func(n *widgetNode) {
		if n == nil {
			return
		}
		count++
		for _, child := range n.children {
			walk(child)
		}
	}
	for _, float := range r.retainedFloats {
		walk(float.root)
	}
	return count
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
		Text{Content: "below", Style: Style{Width: Flex(1)}},
	}}
}

func TestReactivityAutoSizeReflowsSiblings(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 4, func() *reactivityReflowScene {
		return &reactivityReflowScene{label: NewSignal("short")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Grow pushes neighbor right", func(s *reactivityReflowScene) { s.label.Set("a much longer label") })
	sequence.frame("Shrink pulls neighbor left", func(s *reactivityReflowScene) { s.label.Set("x") })
	sequence.frame("Empty", func(s *reactivityReflowScene) { s.label.Set("") })
	sequence.resize(12, 4)
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
	header Signal[string]
}

func (s *reactivityListScene) list() List[string] {
	list := List[string]{ID: "list", State: s.state, ScrollState: s.scroll}
	if s.custom {
		list.RenderItem = func(item string, active, selected bool) Widget {
			prefix := "  "
			if active {
				prefix = "> "
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
				sequence.frame(fmt.Sprintf("Cursor down %d", i+1), func(s *reactivityListScene) { s.list().keyCursorDown() })
			}
			sequence.frame("Header change", func(s *reactivityListScene) { s.header.Set("changed") })
			sequence.frame("Replace items", func(s *reactivityListScene) { s.state.SetItems([]string{"alpha", "beta"}) })
		})
	}
}

// Known divergence: ListState cursor setters don't know the list's ScrollState,
// so a programmatic move past the viewport doesn't scroll until an unrelated
// full frame runs listContainer.OnLayout, which then snaps the list into view.
func TestReactivityListProgrammaticCursorScrolls(t *testing.T) {
	t.Skip("known issue: programmatic ListState cursor moves only scroll on the next full layout")
	sequence := newReactivitySequence(t, 20, 6, func() *reactivityListScene {
		items := []string{"one", "two", "three", "four", "five", "six", "seven"}
		return &reactivityListScene{state: NewListState(items), scroll: NewScrollState(), header: NewSignal("list")}
	})
	sequence.frame("Initial", nil)
	for i := 0; i < 5; i++ {
		sequence.frame(fmt.Sprintf("SelectNext %d", i+1), func(s *reactivityListScene) { s.state.SelectNext() })
	}
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
