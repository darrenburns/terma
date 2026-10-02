package terma

import (
	"image/color"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type terminalHostWidget struct {
	capture bool
	keys    []KeyEvent
	x, y    int
	visible bool
}

func (w *terminalHostWidget) Build(BuildContext) Widget { return w }
func (w *terminalHostWidget) IsFocusable() bool         { return true }
func (w *terminalHostWidget) CapturesKey(string) bool   { return w.capture }
func (w *terminalHostWidget) OnKey(event KeyEvent) bool {
	w.keys = append(w.keys, event)
	return false
}
func (w *terminalHostWidget) CursorPosition() (int, int, bool) {
	return w.x, w.y, w.visible
}

func TestTerminalHostCapturedKeysStayInFocusedWidget(t *testing.T) {
	widget := &terminalHostWidget{capture: true}
	ancestor := &terminalHostWidget{}
	focus := NewFocusManager()
	focus.SetFocusables([]FocusableEntry{
		{ID: "terminal", Focusable: widget, Ancestors: []Widget{ancestor}},
		{ID: "button", Focusable: newTestFocusable("button")},
	})

	keys := []KeyEvent{
		makeKeyEvent(uv.KeyTab, 0),
		makeKeyEvent(uv.KeyTab, uv.ModShift),
		makeKeyEvent('c', uv.ModCtrl),
		makeKeyEvent('z', uv.ModCtrl),
		makeKeyEvent('s', uv.ModCtrl|uv.ModShift),
	}
	for _, key := range keys {
		require.True(t, focus.capturesKey(key), key.Key())
		require.True(t, focus.HandleKey(key), key.Key())
		require.Equal(t, "terminal", focus.FocusedID())
	}
	require.Equal(t, keys, widget.keys)
	require.Empty(t, ancestor.keys, "captured keys cannot invoke ancestor actions")

	widget.capture = false
	require.False(t, focus.capturesKey(makeKeyEvent('c', uv.ModCtrl)))
	focus.HandleKey(makeKeyEvent(uv.KeyTab, 0))
	require.Equal(t, "button", focus.FocusedID())
	focus.HandleKey(makeKeyEvent(uv.KeyTab, uv.ModShift))
	require.Equal(t, "terminal", focus.FocusedID())
}

func TestTerminalHostRawKeyPreservesModifiersAndText(t *testing.T) {
	raw := uv.KeyPressEvent(uv.Key{Code: 'é', Text: "é", Mod: uv.ModAlt | uv.ModShift})
	require.Equal(t, raw, (KeyEvent{event: raw}).Raw())
}

func TestDrawCellsPreservesWideGraphemesAndStyles(t *testing.T) {
	buffer := uv.NewBuffer(8, 3)
	ctx := NewRenderContext(buffer, 8, 3, nil, nil, BuildContext{}, nil).SubContext(2, 1, 5, 1)
	style := uv.Style{Fg: color.RGBA{R: 240, A: 255}, Bg: color.RGBA{B: 80, A: 255}, Attrs: uv.AttrBold | uv.AttrItalic, Underline: uv.UnderlineCurly}
	cells := []uv.Cell{
		{Content: "é", Width: 1, Style: style},
		{Content: "界", Width: 2, Style: style},
		{},
		{Content: "x", Width: 1},
	}
	ctx.DrawCells(0, 0, cells)
	require.Equal(t, cells[0], *buffer.CellAt(2, 1))
	require.Equal(t, cells[1], *buffer.CellAt(3, 1))
	require.True(t, buffer.CellAt(4, 1).IsZero())
	require.Equal(t, cells[3], *buffer.CellAt(5, 1))
	ctx.DrawCells(0, 0, []uv.Cell{{Content: "a", Width: 1}, {Content: "b", Width: 1}, {Content: "c", Width: 1}})
	require.Equal(t, "c", buffer.CellAt(4, 1).Content, "overwriting a wide cell clears its continuation")
}

func TestDrawCellsClipsWideCellsAndLeavesNeighbours(t *testing.T) {
	buffer := uv.NewBuffer(6, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			buffer.SetCell(x, y, &uv.Cell{Content: ".", Width: 1})
		}
	}
	ctx := NewRenderContext(buffer, 6, 3, nil, nil, BuildContext{}, nil).SubContext(2, 1, 2, 1)
	style := uv.Style{Attrs: uv.AttrReverse}
	cells := []uv.Cell{{Content: "界", Width: 2, Style: style}, {}, {Content: "界", Width: 2, Style: style}, {}}
	ctx.DrawCells(-1, 0, cells)
	for _, x := range []int{2, 3} {
		require.Equal(t, uv.Cell{Content: " ", Width: 1, Style: style}, *buffer.CellAt(x, 1))
	}
	ctx.DrawCells(0, -1, cells)
	ctx.DrawCells(0, 1, cells)
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			if y != 1 || x < 2 || x > 3 {
				require.Equal(t, ".", buffer.CellAt(x, y).Content, "neighbour %d,%d", x, y)
			}
		}
	}
}

type cursorRecorder struct {
	x, y    int
	visible bool
}

func (c *cursorRecorder) MoveTo(x, y int) { c.x, c.y = x, y }
func (c *cursorRecorder) ShowCursor()     { c.visible = true }
func (c *cursorRecorder) HideCursor()     { c.visible = false }

func TestCursorProviderPositionVisibilityAndClipping(t *testing.T) {
	widget := &terminalHostWidget{x: 3, y: 2, visible: true}
	entry := &WidgetEntry{Widget: widget, Bounds: Rect{X: 10, Y: 5, Width: 20, Height: 8}, Visible: Rect{X: 11, Y: 6, Width: 18, Height: 6}}
	terminal := &cursorRecorder{}
	positionCursor(terminal, entry)
	require.Equal(t, cursorRecorder{x: 13, y: 7, visible: true}, *terminal)

	widget.visible = false
	widget.x = 4
	positionCursor(terminal, entry)
	require.Equal(t, cursorRecorder{x: 14, y: 7, visible: false}, *terminal, "hidden cursor still positions IME")

	widget.visible = true
	widget.x = 0
	positionCursor(terminal, entry)
	require.False(t, terminal.visible, "clipped cursor cannot appear over a neighbour")
	widget.x = 4
	positionCursor(terminal, entry)
	require.True(t, terminal.visible)
	positionCursor(terminal, nil)
	require.False(t, terminal.visible, "moving focus away hides the hardware cursor")
}

func TestTextInputKeepsHiddenIMECursor(t *testing.T) {
	input := TextInput{State: NewTextInputState("hello")}
	terminal := &cursorRecorder{}
	positionCursor(terminal, &WidgetEntry{Widget: input, Bounds: Rect{X: 4, Y: 2, Width: 10, Height: 1}, Visible: Rect{X: 4, Y: 2, Width: 10, Height: 1}})
	require.Equal(t, cursorRecorder{x: 9, y: 2}, *terminal)
}
