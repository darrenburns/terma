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
