package terma

import "testing"

func selectRow(s Signal[int], row int) *widgetNode {
	node := newWidgetNode(nil)
	withTrackedRead(node, readPhasePaint, func() {
		Select(s, func(cursor int) bool { return cursor == row })
	})
	node.clearDirty()
	return node
}

func TestSelect_NotifiesOnlyWhenResultChanges(t *testing.T) {
	cursor := NewSignal(3)
	rows := make([]*widgetNode, 6)
	for i := range rows {
		rows[i] = selectRow(cursor, i)
	}

	cursor.Set(4)

	for i, node := range rows {
		want := i == 3 || i == 4
		if node.isDirty() != want {
			t.Errorf("row %d dirty = %v, want %v", i, node.isDirty(), want)
		}
		if want && node.dirtyLevel() != DirtyPaint {
			t.Errorf("row %d dirty level = %v, want paint (the phase it read in)", i, node.dirtyLevel())
		}
	}
}

func TestSelect_ReturnsProjectedValue(t *testing.T) {
	cursor := NewSignal(2)
	if got := Select(cursor, func(c int) bool { return c == 2 }); !got {
		t.Error("expected true for the matching row")
	}
}

func TestSelect_PlainGetStillNotifiesOnAnyChange(t *testing.T) {
	cursor := NewSignal(0)
	node := newWidgetNode(nil)
	withTrackedRead(node, readPhasePaint, func() {
		Select(cursor, func(c int) bool { return c == 5 })
		_ = cursor.Get()
	})
	node.clearDirty()

	cursor.Set(1)

	if !node.isDirty() {
		t.Error("a plain Get in the same phase must still notify on every change")
	}
}

func TestSelect_ClearingPhaseRemovesSelector(t *testing.T) {
	cursor := NewSignal(0)
	node := selectRow(cursor, 1)
	node.clearDependenciesForPhase(readPhasePaint)

	cursor.Set(1)

	if node.isDirty() {
		t.Error("expected no notification after the selector's phase was cleared")
	}
	cursor.core.mu.Lock()
	defer cursor.core.mu.Unlock()
	if len(cursor.core.selectors.byNode) != 0 {
		t.Errorf("expected selectors to be removed, got %d nodes", len(cursor.core.selectors.byNode))
	}
}

func TestSelect_UpdateNotifiesChangedSelectors(t *testing.T) {
	cursor := NewSignal(0)
	first, second := selectRow(cursor, 0), selectRow(cursor, 2)

	cursor.Update(func(c int) int { return c + 1 })

	if !first.isDirty() {
		t.Error("row 0 stopped being active and must be notified")
	}
	if second.isDirty() {
		t.Error("row 2 is unaffected and must not be notified")
	}
}

func TestSelectAny_NotifiesOnlyWhenResultChanges(t *testing.T) {
	selection := NewAnySignal(map[int]struct{}{1: {}})
	nodes := make([]*widgetNode, 3)
	for i := range nodes {
		row := i
		nodes[i] = newWidgetNode(nil)
		withTrackedRead(nodes[i], readPhasePaint, func() {
			SelectAny(selection, func(m map[int]struct{}) bool { _, ok := m[row]; return ok })
		})
		nodes[i].clearDirty()
	}

	selection.Set(map[int]struct{}{1: {}, 2: {}})

	for i, want := range []bool{false, false, true} {
		if nodes[i].isDirty() != want {
			t.Errorf("row %d dirty = %v, want %v", i, nodes[i].isDirty(), want)
		}
	}
}
