package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
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
	list := List[string]{ID: "list", State: state, ScrollState: scroll}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	scene.snapshot("TestCollectionWheelScroll_List_wheel_down", "After 6 wheel-downs over a 20-item List in a 5-row viewport")
}

func TestCollectionWheelScroll_Table(t *testing.T) {
	scroll := NewScrollState()
	state := NewTableState(wheelSceneItems())
	table := Table[string]{
		ID: "table", State: state, ScrollState: scroll,
		Columns: []TableColumn{{Width: Cells(10), Header: Text{Content: "Name"}}},
	}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: table}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	scene.snapshot("TestCollectionWheelScroll_Table_wheel_down", "After 6 wheel-downs over a 20-row Table with a header in a 5-row viewport")
}

func TestCollectionWheelScroll_Tree(t *testing.T) {
	roots := make([]TreeNode[string], 0, 20)
	for _, item := range wheelSceneItems() {
		roots = append(roots, TreeNode[string]{Data: item, Children: []TreeNode[string]{}})
	}
	scroll := NewScrollState()
	state := NewTreeState(roots)
	tree := Tree[string]{ID: "tree", State: state, ScrollState: scroll}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: tree}, 24, 5)

	scene.wheel(uv.MouseWheelDown, 6)
	scene.snapshot("TestCollectionWheelScroll_Tree_wheel_down", "After 6 wheel-downs over a 20-node Tree in a 5-row viewport")
}
