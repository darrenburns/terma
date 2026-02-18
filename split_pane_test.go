package terma

import "testing"

func TestSplitPane_Horizontal(t *testing.T) {
	state := NewSplitPaneState(0.5)
	widget := SplitPane{
		State:       state,
		First:       Text{Content: "Left pane", Width: Flex(1)},
		Second:      Text{Content: "Right pane", Width: Flex(1)},
		Orientation: SplitHorizontal,
		DividerSize: 1,
		MinPaneSize: 1,
	}

	AssertSnapshot(t, widget, 40, 10, "Horizontal split at 50%")
}

func TestSplitPane_Vertical(t *testing.T) {
	state := NewSplitPaneState(0.3)
	widget := SplitPane{
		State:       state,
		First:       Text{Content: "Top pane", Width: Flex(1)},
		Second:      Text{Content: "Bottom pane", Width: Flex(1)},
		Orientation: SplitVertical,
		DividerSize: 1,
		MinPaneSize: 1,
	}

	AssertSnapshot(t, widget, 40, 10, "Vertical split at 30%")
}

type splitPaneFocusable struct {
	id string
}

func (f splitPaneFocusable) WidgetID() string {
	return f.id
}

func (f splitPaneFocusable) IsFocusable() bool {
	return true
}

func (f splitPaneFocusable) OnKey(event KeyEvent) bool {
	return false
}

func (f splitPaneFocusable) Build(ctx BuildContext) Widget {
	return Text{Content: "Child"}
}

func TestSplitPane_DisableFocus(t *testing.T) {
	state := NewSplitPaneState(0.5)
	widget := SplitPane{
		ID:                     "split",
		State:                  state,
		DisableFocus:           true,
		First:                  splitPaneFocusable{id: "child"},
		Second:                 Text{Content: "Right"},
		Orientation:            SplitHorizontal,
		DividerForeground:      RGB(255, 0, 0),
		DividerFocusForeground: NewGradient(RGB(0, 255, 0), RGB(0, 0, 255)).WithAngle(0),
	}

	svg := snapshotWithFocus(widget, 20, 5, "split")
	assertSnapshotFromSVG(t, svg, "Attempting to focus SplitPane by ID should fail when DisableFocus=true; divider remains in unfocused color (red), not the focus gradient")
}

func TestSplitPane_DraggingUsesFocusDividerColors(t *testing.T) {
	state := NewSplitPaneState(0.5)
	state.dragging = true

	unfocusedColor := RGB(255, 0, 0)
	focusedColor := RGB(0, 255, 0)

	widget := SplitPane{
		State:                  state,
		DisableFocus:           true,
		First:                  EmptyWidget{},
		Second:                 EmptyWidget{},
		Orientation:            SplitHorizontal,
		DividerForeground:      unfocusedColor,
		DividerFocusForeground: focusedColor,
	}

	width, height := 12, 4
	buf := renderToBufferWithFocus(widget, width, height, "")
	dividerX := computeSplitPaneMetrics(width, widget.dividerSize(), widget.minPaneSize(), state.GetPosition()).offset

	for y := 0; y < height; y++ {
		cell := buf.CellAt(dividerX, y)
		if cell == nil {
			t.Fatalf("expected divider cell at x=%d y=%d", dividerX, y)
		}
		got := FromANSI(cell.Style.Fg)
		if got.Hex() != focusedColor.Hex() {
			t.Fatalf("expected divider focus color %s while dragging, got %s", focusedColor.Hex(), got.Hex())
		}
		if got.Hex() == unfocusedColor.Hex() {
			t.Fatalf("expected divider not to use unfocused color while dragging")
		}
	}
}

func TestSplitPane_KeybindsHorizontalIncludesVimAliases(t *testing.T) {
	state := NewSplitPaneState(0.5)
	pane := SplitPane{
		State:       state,
		Orientation: SplitHorizontal,
	}

	keybinds := pane.Keybinds()
	if _, ok := splitPaneKeybindByKey(keybinds, "left"); !ok {
		t.Fatalf("expected left keybind")
	}
	if _, ok := splitPaneKeybindByKey(keybinds, "right"); !ok {
		t.Fatalf("expected right keybind")
	}
	if _, ok := splitPaneKeybindByKey(keybinds, "h"); !ok {
		t.Fatalf("expected h keybind")
	}
	if _, ok := splitPaneKeybindByKey(keybinds, "l"); !ok {
		t.Fatalf("expected l keybind")
	}
}

func TestSplitPane_KeybindsEscapeUsesOnExitFocus(t *testing.T) {
	state := NewSplitPaneState(0.5)
	calls := 0
	pane := SplitPane{
		State:       state,
		Orientation: SplitHorizontal,
		OnExitFocus: func() {
			calls++
		},
	}

	keybind, ok := splitPaneKeybindByKey(pane.Keybinds(), "escape")
	if !ok {
		t.Fatalf("expected escape keybind")
	}
	if keybind.Action == nil {
		t.Fatalf("expected escape keybind action")
	}
	keybind.Action()
	if calls != 1 {
		t.Fatalf("expected OnExitFocus to be called once, got %d", calls)
	}
}

func splitPaneKeybindByKey(keybinds []Keybind, key string) (Keybind, bool) {
	for _, keybind := range keybinds {
		if keybind.Key == key {
			return keybind, true
		}
	}
	return Keybind{}, false
}
