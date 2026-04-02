package terma

import "testing"

func TestListStateGetItems_SubscribesDuringTrackedRead(t *testing.T) {
	state := NewListState([]string{"a", "b"})
	node := newWidgetNode(nil)

	withTrackedRead(node, readPhaseBuild, func() {
		_ = state.GetItems()
	})
	node.clearDirty()

	state.SetItems([]string{"a", "b", "c"})

	if got := node.dirtyLevel(); got != DirtyBuild {
		t.Fatalf("expected DirtyBuild, got %v", got)
	}
}

func TestListStateItemCount_SubscribesDuringTrackedRead(t *testing.T) {
	state := NewListState([]string{"a", "b"})
	node := newWidgetNode(nil)

	withTrackedRead(node, readPhaseBuild, func() {
		_ = state.ItemCount()
	})
	node.clearDirty()

	state.SetItems([]string{"a", "b", "c"})

	if got := node.dirtyLevel(); got != DirtyBuild {
		t.Fatalf("expected DirtyBuild, got %v", got)
	}
}

func TestListStateSelectedItem_SubscribesDuringTrackedRead(t *testing.T) {
	state := NewListState([]string{"a", "b"})
	node := newWidgetNode(nil)

	withTrackedRead(node, readPhaseBuild, func() {
		_, _ = state.SelectedItem()
	})
	node.clearDirty()

	state.SelectIndex(1)

	if got := node.dirtyLevel(); got != DirtyBuild {
		t.Fatalf("expected DirtyBuild, got %v", got)
	}
}

func TestListStateSelectedItems_SubscribesDuringTrackedRead(t *testing.T) {
	state := NewListState([]string{"a", "b"})
	node := newWidgetNode(nil)

	withTrackedRead(node, readPhaseBuild, func() {
		_ = state.SelectedItems()
	})
	node.clearDirty()

	state.Select(1)

	if got := node.dirtyLevel(); got != DirtyBuild {
		t.Fatalf("expected DirtyBuild, got %v", got)
	}
}
