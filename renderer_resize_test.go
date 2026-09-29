package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

func TestRendererResizeGrowsHeadlessBuffer(t *testing.T) {
	for _, frame := range []struct {
		name string
		draw func(*Renderer, Widget) []FocusableEntry
	}{
		{"Render", (*Renderer).Render},
		{"Update", (*Renderer).Update},
	} {
		t.Run(frame.name, func(t *testing.T) {
			buffer := uv.NewBuffer(2, 1)
			renderer := newTestRenderer(buffer, 2, 1)
			root := Text{Content: "ABCDE\nFGHIJ\nKLMNO"}
			frame.draw(renderer, root)
			renderer.Resize(5, 3)
			frame.draw(renderer, root)
			require.Equal(t, "ABCDE\nFGHIJ\nKLMNO", renderer.ScreenText(), "new columns and rows must accept cells after growth")
		})
	}
}

func TestRendererResizeRepeatedGrowthAndShrink(t *testing.T) {
	for _, frame := range []struct {
		name string
		draw func(*Renderer, Widget) []FocusableEntry
	}{
		{"Render", (*Renderer).Render},
		{"Update", (*Renderer).Update},
	} {
		t.Run(frame.name, func(t *testing.T) {
			buffer := uv.NewBuffer(2, 1)
			renderer := newTestRenderer(buffer, 2, 1)
			rows := []string{"ABCDEFGH", "IJKLMNOP", "QRSTUVWX", "YZabcdef"}
			root := Text{ID: "resized-text", Content: strings.Join(rows, "\n")}
			for _, size := range []Size{{2, 1}, {5, 3}, {3, 2}, {8, 4}, {4, 1}, {0, 0}, {6, 3}} {
				renderer.Resize(size.Width, size.Height)
				frame.draw(renderer, root)
				require.Equal(t, size.Width, buffer.Width())
				require.Equal(t, size.Height, buffer.Height())
				var want []string
				for _, row := range rows[:size.Height] {
					want = append(want, row[:size.Width])
				}
				require.Equal(t, strings.Join(want, "\n"), renderer.ScreenText(), "viewport %dx%d", size.Width, size.Height)
				if size.Width > 0 && size.Height > 0 {
					entry := renderer.WidgetAt(size.Width-1, size.Height-1)
					require.NotNil(t, entry, "new bottom-right cell must be hit-testable")
					require.Equal(t, "resized-text", entry.ID)
				}
				require.Nil(t, renderer.WidgetAt(size.Width, 0), "outside right edge must not be hit-testable")
				require.Nil(t, renderer.WidgetAt(0, size.Height), "outside bottom edge must not be hit-testable")
			}
		})
	}
}

// fixedResizeScreen deliberately exposes only the original CellBuffer API.
type fixedResizeScreen struct{ buffer *uv.Buffer }

func (s *fixedResizeScreen) SetCell(x, y int, cell *uv.Cell) { s.buffer.SetCell(x, y, cell) }
func (s *fixedResizeScreen) CellAt(x, y int) *uv.Cell        { return s.buffer.CellAt(x, y) }

// terminalResizeScreen has the same Resize signature as *uv.Terminal. It owns
// its dimensions and must not be mistaken for the optional buffer capability.
type terminalResizeScreen struct {
	*uv.Buffer
	resizeCalls int
}

func (s *terminalResizeScreen) Resize(width, height int) error {
	s.resizeCalls++
	s.Buffer.Resize(width, height)
	return nil
}

type customResizeScreen struct {
	fixedResizeScreen
	resizeCalls []Size
}

func (s *customResizeScreen) Resize(width, height int) {
	s.resizeCalls = append(s.resizeCalls, Size{width, height})
	s.buffer.Resize(width, height)
}

func TestRendererResizeCellBufferCapabilities(t *testing.T) {
	t.Run("bare buffer matches initial viewport", func(t *testing.T) {
		buffer := uv.NewBuffer(2, 1)
		renderer := newTestRenderer(buffer, 5, 3)
		renderer.Render(Text{Content: "ABCDE\nFGHIJ\nKLMNO"})
		require.Equal(t, "ABCDE\nFGHIJ\nKLMNO", renderer.ScreenText())
		require.Equal(t, 5, buffer.Width())
		require.Equal(t, 3, buffer.Height())
	})
	t.Run("embedded buffer inherits resize capability", func(t *testing.T) {
		buffer := uv.NewBuffer(2, 1)
		renderer := newTestRenderer(reactivityScreen{buffer}, 5, 3)
		renderer.Resize(7, 4)
		require.Equal(t, 7, buffer.Width())
		require.Equal(t, 4, buffer.Height())
	})
	t.Run("custom buffer opts into dimensions", func(t *testing.T) {
		screen := &customResizeScreen{fixedResizeScreen: fixedResizeScreen{buffer: uv.NewBuffer(2, 1)}}
		renderer := newTestRenderer(screen, 5, 3)
		renderer.Resize(7, 4)
		require.Equal(t, []Size{{5, 3}, {7, 4}}, screen.resizeCalls)
		require.Equal(t, 7, screen.buffer.Width())
		require.Equal(t, 4, screen.buffer.Height())
	})
	t.Run("fixed buffer stays caller managed", func(t *testing.T) {
		screen := &fixedResizeScreen{buffer: uv.NewBuffer(2, 1)}
		renderer := newTestRenderer(screen, 5, 3)
		renderer.Resize(7, 4)
		require.Equal(t, 2, screen.buffer.Width())
		require.Equal(t, 1, screen.buffer.Height())
	})
	t.Run("terminal dimensions stay caller managed", func(t *testing.T) {
		screen := &terminalResizeScreen{Buffer: uv.NewBuffer(2, 1)}
		renderer := newTestRenderer(screen, 5, 3)
		renderer.Resize(7, 4)
		require.Zero(t, screen.resizeCalls)
		require.Equal(t, 2, screen.Width())
		require.Equal(t, 1, screen.Height())
	})
	t.Run("negative dimensions become zero", func(t *testing.T) {
		buffer := uv.NewBuffer(2, 1)
		renderer := newTestRenderer(buffer, -1, -2)
		require.Zero(t, buffer.Width())
		require.Zero(t, buffer.Height())
		renderer.Resize(4, 3)
		renderer.Resize(-3, -4)
		require.Zero(t, buffer.Width())
		require.Zero(t, buffer.Height())
	})
}
