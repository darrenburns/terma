package terma

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
)

// wheelScene renders a widget on a live renderer and re-renders after each
// wheel event, as the app loop does.
type wheelScene struct {
	t        *testing.T
	root     Widget
	buf      *uv.Buffer
	renderer *Renderer
	focus    *FocusManager
	focused  AnySignal[Focusable]
	width    int
	height   int
}

func newWheelScene(t *testing.T, root Widget, width, height int) *wheelScene {
	t.Helper()
	buf := uv.NewBuffer(width, height)
	focus := NewFocusManager()
	focus.SetRootWidget(root)
	focused := NewAnySignal[Focusable](nil)
	s := &wheelScene{
		t: t, root: root, buf: buf, focus: focus, focused: focused, width: width, height: height,
		renderer: NewRenderer(buf, width, height, focus, focused, NewAnySignal[Widget](nil)),
	}
	focus.SetFocusables(s.renderer.Render(root))
	focused.Set(focus.Focused())
	s.draw()
	return s
}

func (s *wheelScene) draw() {
	s.focus.SetFocusables(s.renderer.Update(s.root))
}

func (s *wheelScene) wheel(button uv.MouseButton, times int) {
	s.t.Helper()
	for i := 0; i < times; i++ {
		dispatchMouseWheel(s.renderer, 1, 1, button)
		s.draw()
	}
}

func (s *wheelScene) snapshot(name, description string) {
	s.t.Helper()
	assertBufferSnapshot(s.t, name, s.buf, s.width, s.height, DefaultSVGOptions(), description)
}

func wheelSceneItems() []string {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %02d", i)
	}
	return items
}

func TestCollectionWheelScroll_List(t *testing.T) {
	scroll := NewScrollState()
	state := NewListState(wheelSceneItems())
	cursorChanges := 0
	list := List[string]{ID: "list", State: state, ScrollState: scroll, OnCursorChange: func(string) { cursorChanges++ }}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	scene.snapshot("TestCollectionWheelScroll_List_wheel_down", "After 6 wheel-downs over a 20-item List in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	scene.wheel(uv.MouseWheelUp, 2)
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())

	// A full render must not snap the viewport back to the cursor either.
	scene.renderer.Render(scene.root)
	assert.Equal(t, 4, scroll.GetOffset())

	list.keyCursorDown()
	scene.draw()
	assert.Equal(t, 1, state.CursorIndex.Peek())
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	scene.snapshot("TestCollectionWheelScroll_List_key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_Table(t *testing.T) {
	scroll := NewScrollState()
	state := NewTableState(wheelSceneItems())
	cursorChanges := 0
	table := Table[string]{
		ID: "table", State: state, ScrollState: scroll,
		Columns:        []TableColumn{{Width: Cells(10), Header: Text{Content: "Name"}}},
		OnCursorChange: func(string) { cursorChanges++ },
	}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: table}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	scene.snapshot("TestCollectionWheelScroll_Table_wheel_down", "After 6 wheel-downs over a 20-row Table with a header in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	scene.wheel(uv.MouseWheelUp, 6)
	assert.Equal(t, 0, scroll.GetOffset(), "wheel scrolls back up to reveal the header")

	scene.wheel(uv.MouseWheelDown, 8)
	table.keyCursorDown()
	scene.draw()
	assert.Equal(t, 1, state.CursorIndex.Peek())
	assert.Equal(t, 2, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	scene.snapshot("TestCollectionWheelScroll_Table_key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_Tree(t *testing.T) {
	roots := make([]TreeNode[string], 0, 20)
	for _, item := range wheelSceneItems() {
		roots = append(roots, TreeNode[string]{Data: item, Children: []TreeNode[string]{}})
	}
	scroll := NewScrollState()
	state := NewTreeState(roots)
	cursorChanges := 0
	tree := Tree[string]{ID: "tree", State: state, ScrollState: scroll, OnCursorChange: func(string) { cursorChanges++ }}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: tree}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	assert.Equal(t, []int{0}, state.CursorPath.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	scene.snapshot("TestCollectionWheelScroll_Tree_wheel_down", "After 6 wheel-downs over a 20-node Tree in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	tree.keyCursorDown()
	scene.draw()
	assert.Equal(t, []int{1}, state.CursorPath.Peek())
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	scene.snapshot("TestCollectionWheelScroll_Tree_key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_InitialCursorRevealedOnFirstLayout(t *testing.T) {
	newScene := func() (*wheelScene, *ScrollState) {
		scroll := NewScrollState()
		state := NewListState(wheelSceneItems())
		state.SelectIndex(15)
		list := List[string]{ID: "list", State: state, ScrollState: scroll}
		return newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5), scroll
	}

	scene, scroll := newScene()
	assert.Equal(t, 11, scroll.GetOffset())
	assert.Contains(t, scene.renderer.ScreenText(), "Item 15")
	scene.renderer.Render(scene.root)
	assert.Equal(t, 11, scroll.GetOffset())

	scene, scroll = newScene()
	scene.wheel(uv.MouseWheelDown, 2)
	scene.renderer.Render(scene.root)
	assert.Equal(t, 13, scroll.GetOffset(), "the user's scroll stays relative to the initially revealed cursor")
}

func TestCollectionWheelScroll_TextArea(t *testing.T) {
	scroll := NewScrollState()
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line %02d", i)
	}
	state := NewTextAreaState(strings.Join(lines, "\n"))
	state.CursorIndex.Set(0)
	area := TextArea{ID: "area", State: state, ScrollState: scroll}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: area}, 24, 5)

	// A click leaves an anchor at the cursor, ready for a drag.
	area.OnMouseDown(MouseEvent{ClickCount: 1})
	scene.draw()

	scene.wheel(uv.MouseWheelDown, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, "", state.GetSelectedText(), "wheel must not select text")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")

	scene.wheel(uv.MouseWheelUp, 2)
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())

	area.cursorDown()
	scene.draw()
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
}
