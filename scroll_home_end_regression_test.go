package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

func TestTextAreaPointerInterruptsGlide(t *testing.T) {
	s := NewScrollState()
	w := TextArea{ID: "area", State: NewTextAreaState(strings.Join(numberedItems(20), "\n")), ScrollState: s}
	w.State.CursorIndex.Set(0)
	p := NewPilot(t, Scrollable{State: s, Height: Cells(5), Child: w}, 24, 5)
	w.cursorPageDown()
	w.cursorPageDown()
	p.Advance(testFrame)
	before := s.GetOffset()
	w.OnMouseDown(MouseEvent{LocalX: 0, LocalY: 3, Button: uv.MouseLeft})
	p.settle()
	require.Nil(t, s.animation)
	require.Equal(t, before, s.GetOffset(), "clicking visible text stops the glide without moving the text")
}

func TestEndGlideFollowsGrowingPinnedContent(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 100)
	s.PinToBottom = true
	w := Scrollable{State: s, Focusable: true}
	require.True(t, w.OnKey(makeKeyEvent(uv.KeyEnd, 0)))
	advance(testFrame)
	beforeGrowth := s.GetOffset()
	s.updateLayout(10, 110)
	require.Equal(t, beforeGrowth, s.GetOffset(), "growth should extend the glide without jumping")
	glide(t, s, advance)
	require.Equal(t, 100, s.GetOffset())
	s.updateLayout(10, 120)
	require.Equal(t, 110, s.GetOffset(), "the pin keeps following after the glide finishes")
	require.True(t, s.IsPinned())

	w.OnKey(makeKeyEvent(uv.KeyHome, 0))
	advance(testFrame)
	s.updateLayout(10, 130)
	glide(t, s, advance)
	require.Zero(t, s.GetOffset(), "Home releases the pin even when content grows during its glide")
	require.False(t, s.IsPinned())
}

func TestCollectionRangeSelectionInterruptsGlide(t *testing.T) {
	type collection struct {
		widget                        Widget
		last, selectFirst, selectLast func()
	}
	cases := map[string]func(*ScrollState) collection{
		"List": func(s *ScrollState) collection {
			w := List[string]{ID: "c", State: NewListState(numberedItems(20)), ScrollState: s, MultiSelect: true}
			return collection{w, w.keyCursorToLast, w.shiftCursorToFirst, w.shiftCursorToLast}
		},
		"Table rows": func(s *ScrollState) collection {
			rows := make([][]string, 20)
			for i, item := range numberedItems(20) {
				rows[i] = []string{item}
			}
			w := Table[[]string]{ID: "c", State: NewTableState(rows), ScrollState: s, Columns: []TableColumn{{}}, MultiSelect: true}
			return collection{w, w.keyCursorToLast, w.shiftRowToFirst, w.shiftRowToLast}
		},
		"Table cells": func(s *ScrollState) collection {
			rows := make([][]string, 20)
			for i, item := range numberedItems(20) {
				rows[i] = []string{item}
			}
			w := Table[[]string]{ID: "c", State: NewTableState(rows), ScrollState: s, Columns: []TableColumn{{}}, MultiSelect: true, SelectionMode: TableSelectionCursor}
			return collection{w, w.keyCursorToLast, w.shiftCellToFirst, w.shiftCellToLast}
		},
		"Tree": func(s *ScrollState) collection {
			nodes := make([]TreeNode[string], 20)
			for i, item := range numberedItems(20) {
				nodes[i] = TreeNode[string]{Data: item}
			}
			w := Tree[string]{ID: "c", State: NewTreeState(nodes), ScrollState: s, MultiSelect: true}
			return collection{w, w.keyCursorToLast, w.shiftCursorToFirst, w.shiftCursorToLast}
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			s := NewScrollState()
			c := build(s)
			p := NewPilot(t, Scrollable{ID: "scroll", State: s, Height: Cells(5), Child: c.widget}, 24, 5)
			c.last()
			p.Advance(testFrame)
			require.Positive(t, s.GetOffset())
			c.selectFirst()
			p.settle()
			require.Zero(t, s.GetOffset(), "Shift+Home immediately reveals the selection endpoint")
			require.Nil(t, s.animation)

			c.last()
			p.Advance(testFrame)
			c.selectLast()
			p.settle()
			require.Equal(t, s.maxOffset(), s.GetOffset(), "Shift+End interrupts even when the glide already targets the endpoint")
			require.Nil(t, s.animation)

			s.SetOffset(0)
			c.last()
			p.Advance(testFrame)
			beforeClick := s.GetOffset()
			c.widget.(interface{ OnMouseDown(MouseEvent) }).OnMouseDown(MouseEvent{LocalX: 1, LocalY: 6, Button: uv.MouseLeft})
			p.settle()
			require.Nil(t, s.animation, "clicking a visible row stops the glide")
			require.Equal(t, beforeClick, s.GetOffset(), "a visible click must not scroll against the old animation target")

			c.last()
			p.Advance(testFrame)
			c.widget.(interface{ OnMouseMove(MouseEvent) }).OnMouseMove(MouseEvent{LocalX: 1, LocalY: 100, Button: uv.MouseLeft})
			p.settle()
			require.Nil(t, s.animation, "drag selection stops the glide")
			require.Equal(t, s.maxOffset(), s.GetOffset(), "drag selection immediately reveals its endpoint")
		})
	}
}
