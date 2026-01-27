package terma

import "testing"

// testFocusable is a minimal Focusable implementation for testing.
type testFocusable struct {
	id        string
	focusable bool
}

func (f *testFocusable) WidgetID() string         { return f.id }
func (f *testFocusable) OnKey(event KeyEvent) bool { return false }
func (f *testFocusable) IsFocusable() bool         { return f.focusable }
func (f *testFocusable) Build(ctx BuildContext) Widget {
	return EmptyWidget{}
}

func newTestFocusable(id string) *testFocusable {
	return &testFocusable{id: id, focusable: true}
}

func TestFocusTrap_NextWrapsWithinTrap(t *testing.T) {
	// Setup: A (no trap), B (trap "t"), C (trap "t")
	// Focus B, Tab should cycle B -> C -> B
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A"), TrapID: ""},
		{ID: "B", Focusable: newTestFocusable("B"), TrapID: "t"},
		{ID: "C", Focusable: newTestFocusable("C"), TrapID: "t"},
	})
	fm.FocusByID("B")

	// Tab from B → C
	fm.FocusNext()
	if fm.FocusedID() != "C" {
		t.Errorf("expected focus on C, got %q", fm.FocusedID())
	}

	// Tab from C → wraps back to B (within trap)
	fm.FocusNext()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus to wrap to B, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_PreviousWrapsWithinTrap(t *testing.T) {
	// Setup: A (no trap), B (trap "t"), C (trap "t")
	// Focus C, Shift+Tab should cycle C -> B -> C
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A"), TrapID: ""},
		{ID: "B", Focusable: newTestFocusable("B"), TrapID: "t"},
		{ID: "C", Focusable: newTestFocusable("C"), TrapID: "t"},
	})
	fm.FocusByID("C")

	// Shift+Tab from C → B
	fm.FocusPrevious()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus on B, got %q", fm.FocusedID())
	}

	// Shift+Tab from B → wraps to C (within trap)
	fm.FocusPrevious()
	if fm.FocusedID() != "C" {
		t.Errorf("expected focus to wrap to C, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_InactiveTrapIsTransparent(t *testing.T) {
	// All focusables have empty TrapID (inactive trap / no trap)
	// Tab should cycle through all focusables globally
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A"), TrapID: ""},
		{ID: "B", Focusable: newTestFocusable("B"), TrapID: ""},
		{ID: "C", Focusable: newTestFocusable("C"), TrapID: ""},
	})
	fm.FocusByID("A")

	fm.FocusNext()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus on B, got %q", fm.FocusedID())
	}

	fm.FocusNext()
	if fm.FocusedID() != "C" {
		t.Errorf("expected focus on C, got %q", fm.FocusedID())
	}

	fm.FocusNext()
	if fm.FocusedID() != "A" {
		t.Errorf("expected focus to wrap to A, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_NestedTraps(t *testing.T) {
	// Inner trap "inner" should win for its subtree
	// A (trap "outer"), B (trap "inner"), C (trap "inner"), D (trap "outer")
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A"), TrapID: "outer"},
		{ID: "B", Focusable: newTestFocusable("B"), TrapID: "inner"},
		{ID: "C", Focusable: newTestFocusable("C"), TrapID: "inner"},
		{ID: "D", Focusable: newTestFocusable("D"), TrapID: "outer"},
	})

	// Focus B (in inner trap) → Tab cycles B -> C -> B
	fm.FocusByID("B")
	fm.FocusNext()
	if fm.FocusedID() != "C" {
		t.Errorf("expected focus on C, got %q", fm.FocusedID())
	}
	fm.FocusNext()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus to wrap to B, got %q", fm.FocusedID())
	}

	// Focus A (in outer trap) → Tab cycles A -> D -> A
	fm.FocusByID("A")
	fm.FocusNext()
	if fm.FocusedID() != "D" {
		t.Errorf("expected focus on D, got %q", fm.FocusedID())
	}
	fm.FocusNext()
	if fm.FocusedID() != "A" {
		t.Errorf("expected focus to wrap to A, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_NoTrapCyclesGlobally(t *testing.T) {
	// Baseline: no traps, all focusables cycle globally
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A")},
		{ID: "B", Focusable: newTestFocusable("B")},
		{ID: "C", Focusable: newTestFocusable("C")},
	})
	fm.FocusByID("A")

	// Forward cycle: A -> B -> C -> A
	fm.FocusNext()
	if fm.FocusedID() != "B" {
		t.Errorf("expected B, got %q", fm.FocusedID())
	}
	fm.FocusNext()
	if fm.FocusedID() != "C" {
		t.Errorf("expected C, got %q", fm.FocusedID())
	}
	fm.FocusNext()
	if fm.FocusedID() != "A" {
		t.Errorf("expected A, got %q", fm.FocusedID())
	}

	// Backward cycle: A -> C -> B -> A
	fm.FocusPrevious()
	if fm.FocusedID() != "C" {
		t.Errorf("expected C, got %q", fm.FocusedID())
	}
	fm.FocusPrevious()
	if fm.FocusedID() != "B" {
		t.Errorf("expected B, got %q", fm.FocusedID())
	}
	fm.FocusPrevious()
	if fm.FocusedID() != "A" {
		t.Errorf("expected A, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_SingleFocusableInTrap(t *testing.T) {
	// Only one focusable in trap → Tab stays on it
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{
		{ID: "A", Focusable: newTestFocusable("A"), TrapID: ""},
		{ID: "B", Focusable: newTestFocusable("B"), TrapID: "t"},
	})
	fm.FocusByID("B")

	fm.FocusNext()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus to stay on B, got %q", fm.FocusedID())
	}

	fm.FocusPrevious()
	if fm.FocusedID() != "B" {
		t.Errorf("expected focus to stay on B, got %q", fm.FocusedID())
	}
}

func TestFocusTrap_CollectorSetsCorrectTrapID(t *testing.T) {
	fc := NewFocusCollector()

	// Simulate collecting focusables with a trap scope
	fc.PushTrap("trap1")
	widget1 := newTestFocusable("w1")
	fm := NewFocusManager()
	ctx := NewBuildContext(fm, AnySignal[Focusable]{}, AnySignal[Widget]{}, nil)
	fc.Collect(widget1, "w1", ctx)

	fc.PushTrap("trap2") // Nested trap
	widget2 := newTestFocusable("w2")
	fc.Collect(widget2, "w2", ctx)
	fc.PopTrap() // Back to trap1

	widget3 := newTestFocusable("w3")
	fc.Collect(widget3, "w3", ctx)
	fc.PopTrap() // No trap

	widget4 := newTestFocusable("w4")
	fc.Collect(widget4, "w4", ctx)

	focusables := fc.Focusables()
	if len(focusables) != 4 {
		t.Fatalf("expected 4 focusables, got %d", len(focusables))
	}

	if focusables[0].TrapID != "trap1" {
		t.Errorf("w1: expected TrapID 'trap1', got %q", focusables[0].TrapID)
	}
	if focusables[1].TrapID != "trap2" {
		t.Errorf("w2: expected TrapID 'trap2', got %q", focusables[1].TrapID)
	}
	if focusables[2].TrapID != "trap1" {
		t.Errorf("w3: expected TrapID 'trap1', got %q", focusables[2].TrapID)
	}
	if focusables[3].TrapID != "" {
		t.Errorf("w4: expected empty TrapID, got %q", focusables[3].TrapID)
	}
}

func TestFocusTrap_CollectorResetClearsTrapStack(t *testing.T) {
	fc := NewFocusCollector()
	fc.PushTrap("trap1")
	fc.Reset()

	if fc.CurrentTrapID() != "" {
		t.Errorf("expected empty TrapID after Reset, got %q", fc.CurrentTrapID())
	}
}

func TestFocusTrap_WidgetInterface(t *testing.T) {
	ft := FocusTrap{
		ID:     "my-trap",
		Active: true,
		Child:  Text{Content: "hello"},
	}

	if ft.WidgetID() != "my-trap" {
		t.Errorf("expected WidgetID 'my-trap', got %q", ft.WidgetID())
	}

	if !ft.TrapsFocus() {
		t.Error("expected TrapsFocus() to return true when Active is true")
	}

	ft.Active = false
	if ft.TrapsFocus() {
		t.Error("expected TrapsFocus() to return false when Active is false")
	}
}
