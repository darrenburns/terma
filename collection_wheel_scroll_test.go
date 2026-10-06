package terma

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectionWheelScroll_List(t *testing.T) {
	scroll := NewScrollState()
	state := NewListState(numberedItems(20))
	cursorChanges := 0
	list := List[string]{ID: "list", State: state, ScrollState: scroll, OnCursorChange: func(string) { cursorChanges++ }}
	p := NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5)

	p.Scroll(1, 1, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	p.AssertSnapshot("wheel_down", "After 6 wheel-downs over a 20-item List in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	p.Scroll(1, 1, -2)
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())

	// A full render must not snap the viewport back to the cursor either.
	p.session.renderer.Render(p.session.root)
	assert.Equal(t, 4, scroll.GetOffset())

	list.keyCursorDown()
	p.settle()
	assert.Equal(t, 1, state.CursorIndex.Peek())
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	p.AssertSnapshot("key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_Table(t *testing.T) {
	scroll := NewScrollState()
	state := NewTableState(numberedItems(20))
	cursorChanges := 0
	table := Table[string]{
		ID: "table", State: state, ScrollState: scroll,
		Columns:        []TableColumn{{Width: Cells(10), Header: Text{Content: "Name"}}},
		OnCursorChange: func(string) { cursorChanges++ },
	}
	p := NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: table}, 24, 5)

	p.Scroll(1, 1, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	p.AssertSnapshot("wheel_down", "After 6 wheel-downs over a 20-row Table with a header in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	p.Scroll(1, 1, -6)
	assert.Equal(t, 0, scroll.GetOffset(), "wheel scrolls back up to reveal the header")

	p.Scroll(1, 1, 8)
	table.keyCursorDown()
	p.settle()
	assert.Equal(t, 1, state.CursorIndex.Peek())
	assert.Equal(t, 2, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	p.AssertSnapshot("key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_Tree(t *testing.T) {
	roots := make([]TreeNode[string], 0, 20)
	for _, item := range numberedItems(20) {
		roots = append(roots, TreeNode[string]{Data: item, Children: []TreeNode[string]{}})
	}
	scroll := NewScrollState()
	state := NewTreeState(roots)
	cursorChanges := 0
	tree := Tree[string]{ID: "tree", State: state, ScrollState: scroll, OnCursorChange: func(string) { cursorChanges++ }}
	p := NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: tree}, 24, 5)

	p.Scroll(1, 1, 6)
	assert.Equal(t, []int{0}, state.CursorPath.Peek(), "wheel must not move the cursor")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")
	assert.Zero(t, cursorChanges, "wheel must not report cursor changes")
	p.AssertSnapshot("wheel_down", "After 6 wheel-downs over a 20-node Tree in a 5-row viewport: the view scrolls, the cursor stays on Item 00")

	tree.keyCursorDown()
	p.settle()
	assert.Equal(t, []int{1}, state.CursorPath.Peek())
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
	p.AssertSnapshot("key_after_wheel", "Pressing down after wheel scrolling moves the cursor to Item 01 and scrolls it back into view")
}

func TestCollectionWheelScroll_InitialCursorRevealedOnFirstLayout(t *testing.T) {
	newPilot := func(t *testing.T) (*Pilot, *ScrollState) {
		scroll := NewScrollState()
		state := NewListState(numberedItems(20))
		state.SelectIndex(15)
		list := List[string]{ID: "list", State: state, ScrollState: scroll}
		return NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5), scroll
	}

	t.Run("first layout", func(t *testing.T) {
		p, scroll := newPilot(t)
		assert.Equal(t, 11, scroll.GetOffset())
		assert.Contains(t, p.ScreenText(), "Item 15")
		p.session.renderer.Render(p.session.root)
		assert.Equal(t, 11, scroll.GetOffset())
	})

	t.Run("after wheel", func(t *testing.T) {
		p, scroll := newPilot(t)
		p.Scroll(1, 1, 2)
		p.session.renderer.Render(p.session.root)
		assert.Equal(t, 13, scroll.GetOffset(), "the user's scroll stays relative to the initially revealed cursor")
	})
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
	p := NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: area}, 24, 5)

	// A click leaves an anchor at the cursor, ready for a drag.
	area.OnMouseDown(MouseEvent{ClickCount: 1})
	p.settle()

	p.Scroll(1, 1, 6)
	assert.Equal(t, 0, state.CursorIndex.Peek(), "wheel must not move the cursor")
	assert.Equal(t, "", state.GetSelectedText(), "wheel must not select text")
	assert.Equal(t, 6, scroll.GetOffset(), "wheel scrolls the viewport")

	p.Scroll(1, 1, -2)
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())

	area.cursorDown()
	p.settle()
	assert.Equal(t, 1, scroll.GetOffset(), "moving the cursor with the keyboard brings it back into view")
}
