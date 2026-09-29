package terma

import (
	"math/rand"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

func TestTextAreaGeometryMatchesUncachedLayout(t *testing.T) {
	texts := []string{"", "hello world", "hello  world ", "\n\nabc\n", "界👩‍👩‍👧‍👦é abc 界\nwide🙂 word", "a\tbc\rdef"}
	rng := rand.New(rand.NewSource(1))
	alphabet := []string{"a", " ", "\n", "界", "🙂", "é"}
	for range 50 {
		var text strings.Builder
		for range rng.Intn(50) {
			text.WriteString(alphabet[rng.Intn(len(alphabet))])
		}
		texts = append(texts, text.String())
	}
	for _, text := range texts {
		state := NewTextAreaState(text)
		graphemes, revision := state.Content.peekWithRevision()
		for _, wrap := range []WrapMode{WrapNone, WrapSoft, WrapHard} {
			for _, width := range []int{-1, 0, 1, 2, 5, 20} {
				for cursor := -1; cursor <= len(graphemes)+1; cursor++ {
					want := buildTextAreaLayout(graphemes, wrap, width, cursor)
					got := state.layoutFor(graphemes, revision, wrap, width, cursor)
					require.Equal(t, want, got, "text=%q wrap=%v width=%d cursor=%d", text, wrap, width, cursor)
				}
			}
		}
	}
}

func TestTextAreaGeometryTracksDirectContentMutations(t *testing.T) {
	state := NewTextAreaState("abc def")
	assertCurrent := func() {
		t.Helper()
		graphemes, revision := state.Content.peekWithRevision()
		for _, wrap := range []WrapMode{WrapNone, WrapSoft, WrapHard} {
			for _, width := range []int{2, 5, 10} {
				for cursor := 0; cursor <= len(graphemes); cursor++ {
					require.Equal(t, buildTextAreaLayout(graphemes, wrap, width, cursor), state.layoutFor(graphemes, revision, wrap, width, cursor))
				}
			}
		}
	}
	assertCurrent()
	graphemes := state.Content.Peek()
	graphemes[0] = "界"
	state.Content.Set(graphemes) // Same backing slice and length.
	assertCurrent()
	state.Content.Update(func(content []string) []string {
		content[3] = "\n"
		return content
	})
	assertCurrent()
	state.Content.Set(splitGraphemes("🙂\nnew"))
	assertCurrent()
	state.Content = NewAnySignal(splitGraphemes("replacement signal"))
	assertCurrent()
}

func TestTextAreaGeometryReusesLinesForCursorAndSelection(t *testing.T) {
	state := NewTextAreaState("first line\n界 second line with wraps")
	state.lastWidth = 11
	state.CursorScreenPosition(0, 0)
	lines := state.geometry.layout.lines
	columns := state.geometry.columns
	for cursor := 0; cursor <= len(state.Content.Peek()); cursor++ {
		state.CursorIndex.Set(cursor)
		state.SelectionAnchor.Set(0)
		state.CursorScreenPosition(0, 0)
		require.Same(t, &lines[0], &state.geometry.layout.lines[0])
		require.Same(t, &columns[0], &state.geometry.columns[0])
	}
}

func TestTextAreaCachedFramesMatchFreshGeometry(t *testing.T) {
	state := NewTextAreaState("abc 界 é🙂\nsecond line wraps here\nlast line")
	widget := TextArea{ID: "geometry-cache", State: state, Width: Cells(10), Height: Cells(3)}
	for _, wrap := range []WrapMode{WrapNone, WrapSoft, WrapHard} {
		state.WrapMode.Set(wrap)
		for _, width := range []int{10, 6, 14} {
			widget.Width = Cells(width)
			for cursor := 0; cursor <= len(state.Content.Peek()); cursor++ {
				state.CursorIndex.Set(cursor)
				state.SelectionAnchor.Set(cursor / 2)
				cached := RenderToBuffer(widget, width, 3)
				freshState := *state
				freshState.geometry = textAreaGeometry{}
				freshWidget := widget
				freshWidget.State = &freshState
				fresh := RenderToBuffer(freshWidget, width, 3)
				for y := 0; y < 3; y++ {
					for x := 0; x < width; x++ {
						require.Equal(t, fresh.CellAt(x, y), cached.CellAt(x, y), "wrap=%v width=%d cursor=%d cell=(%d,%d)", wrap, width, cursor, x, y)
					}
				}
			}
		}
	}
}

func TestTextAreaCachedGeometryKeepsContentSubscriptions(t *testing.T) {
	state := NewTextAreaState("initial text")
	widget := TextArea{ID: "geometry-subscriptions", State: state, Width: Cells(12), Height: Cells(3)}
	buffer := uv.NewBuffer(12, 3)
	renderer := newTestRenderer(buffer, 12, 3)
	t.Cleanup(func() { renderer.rootNode.dispose() })
	renderer.Update(widget)
	state.Content.Set(splitGraphemes("changed text"))
	renderer.Update(widget)
	require.Equal(t, "c", buffer.CellAt(0, 0).Content)
	graphemes := state.Content.Peek()
	graphemes[0] = "界"
	state.Content.Set(graphemes)
	renderer.Update(widget)
	require.Equal(t, "界", buffer.CellAt(0, 0).Content)
}

func TestTextAreaGeometryDoesNotCacheReactiveHighlighter(t *testing.T) {
	color := NewSignal(Red)
	calls := 0
	state := NewTextAreaState("text")
	widget := TextArea{ID: "reactive-highlighter", State: state, Width: Cells(12), Height: Cells(3),
		Highlighter: HighlighterFunc(func(_ string, _ []string) []TextHighlight {
			calls++
			return []TextHighlight{{Start: 0, End: 1, Style: SpanStyle{Foreground: color.Get()}}}
		})}
	buffer := uv.NewBuffer(12, 3)
	renderer := newTestRenderer(buffer, 12, 3)
	t.Cleanup(func() { renderer.rootNode.dispose() })
	renderer.Update(widget)
	before := *buffer.CellAt(0, 0)
	color.Set(Blue)
	renderer.Update(widget)
	require.Equal(t, 2, calls)
	require.NotEqual(t, before.Style, buffer.CellAt(0, 0).Style)
}
