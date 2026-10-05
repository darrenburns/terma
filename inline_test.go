package terma

import (
	"bytes"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newInlineScene renders root the way an inline app does: into a canvas
// width x maxHeight, for a screen screenRows tall.
func newInlineScene(t *testing.T, root Widget, width, screenRows int, opts InlineOptions) (*inlineMode, *Renderer, *FocusManager) {
	t.Helper()
	m := &inlineMode{opts: opts}
	focus := NewFocusManager()
	focus.SetRootWidget(root)
	canvas, w, h := m.canvas(nil, width, screenRows)
	r := NewRenderer(canvas, w, h, focus, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	focus.SetFocusables(r.Update(root))
	return m, r, focus
}

func rows(n int) Widget {
	children := make([]Widget, n)
	for i := range children {
		children[i] = Text{Content: "row"}
	}
	return Column{Children: children}
}

func TestInlineRegionHeight_FollowsRootHeight(t *testing.T) {
	m, r, _ := newInlineScene(t, rows(3), 20, 24, InlineOptions{})
	assert.Equal(t, 3, m.regionHeight(r.contentHeight()))
}

func TestInlineRegionHeight_GrowsForOverlayBelowRoot(t *testing.T) {
	root := Column{Children: []Widget{
		Text{ID: "anchor", Content: "prompt"},
		Floating{
			Visible: true,
			Config:  FloatConfig{AnchorID: "anchor", Anchor: AnchorBottomLeft},
			Child:   rows(4),
		},
	}}
	m, r, _ := newInlineScene(t, root, 20, 24, InlineOptions{})
	assert.Equal(t, 1, r.lastLayoutHeight, "the root is one row")
	assert.Equal(t, 5, m.regionHeight(r.contentHeight()), "the region reaches the overlay's bottom")
}

func TestInlineRegionHeight_OverlayInsideRootDoesNotGrowIt(t *testing.T) {
	root := Column{Children: []Widget{
		Text{ID: "anchor", Content: "prompt"},
		rows(5),
		Floating{
			Visible: true,
			Config:  FloatConfig{AnchorID: "anchor", Anchor: AnchorBottomLeft},
			Child:   rows(2),
		},
	}}
	m, r, _ := newInlineScene(t, root, 20, 24, InlineOptions{})
	assert.Equal(t, 6, m.regionHeight(r.contentHeight()))
}

func TestInlineRegionHeight_ClampsToMaxHeightAndOneRow(t *testing.T) {
	m, r, _ := newInlineScene(t, rows(10), 20, 24, InlineOptions{MaxHeight: 4})
	assert.Equal(t, 4, m.regionHeight(r.contentHeight()), "content taller than MaxHeight is cut to it")

	m, r, _ = newInlineScene(t, Column{}, 20, 24, InlineOptions{})
	assert.Equal(t, 1, m.regionHeight(r.contentHeight()), "an empty app still takes a row")
}

func TestInlineMaxHeight_FlexRootFillsIt(t *testing.T) {
	root := Column{Height: Flex(1), Children: []Widget{Text{Content: "a"}}}
	m, r, _ := newInlineScene(t, root, 20, 24, InlineOptions{MaxHeight: 6})
	assert.Equal(t, 6, m.regionHeight(r.contentHeight()))

	m, r, _ = newInlineScene(t, root, 20, 24, InlineOptions{MaxHeight: 100})
	assert.Equal(t, 24, m.regionHeight(r.contentHeight()), "MaxHeight is capped at the screen")

	m, r, _ = newInlineScene(t, root, 20, 24, InlineOptions{})
	assert.Equal(t, 24, m.regionHeight(r.contentHeight()), "0 means the screen's height")
}

type growingApp struct {
	extra  Signal[int]
	builds *int
}

type countedText struct{ builds *int }

func (c countedText) Build(BuildContext) Widget {
	*c.builds++
	return Text{Content: "unrelated"}
}

func (a growingApp) Build(BuildContext) Widget {
	return Column{Children: []Widget{countedText{a.builds}, growingRows{a.extra}}}
}

type growingRows struct{ extra Signal[int] }

func (g growingRows) Build(BuildContext) Widget { return rows(g.extra.Get()) }

func TestInlineRegionHeight_ChangeDoesNotRebuildTheTree(t *testing.T) {
	app := growingApp{extra: NewSignal(1), builds: new(int)}
	m, r, focus := newInlineScene(t, app, 20, 24, InlineOptions{})
	require.Equal(t, 2, m.regionHeight(r.contentHeight()))
	builds := *app.builds

	for _, extra := range []int{5, 2, 8} {
		app.extra.Set(extra)
		focus.SetFocusables(r.Update(app))
		assert.Equal(t, 1+extra, m.regionHeight(r.contentHeight()))
	}
	assert.Equal(t, builds, *app.builds, "widgets whose inputs didn't change are not built again")
	assert.NotEqual(t, string(rendererFrameFull), r.Stats().FrameMode, "a height change is not a forced full render")
	assert.Equal(t, 24, r.height, "the renderer keeps its canvas size")
}

func TestInlineOrigin_FollowsPrintedRowsAndScrolling(t *testing.T) {
	m := &inlineMode{screenRows: 24, origin: inlineOrigin{row: 5, known: true}}

	m.moved(0, 4)
	assert.Equal(t, 5, m.origin.row, "a frame that fits stays put")
	m.moved(3, 4)
	assert.Equal(t, 8, m.origin.row, "printed rows push the region down")
	m.moved(0, 20)
	assert.Equal(t, 4, m.origin.row, "a region reaching past the bottom scrolls the screen up")
	m.moved(10, 20)
	assert.Equal(t, 4, m.origin.row, "rows printed with the region at the bottom scroll it")
	m.moved(0, 2)
	assert.Equal(t, 4, m.origin.row, "shrinking leaves the top where it is")
}

func TestInlineOrigin_AnswerReplaysFramesSinceTheQuestion(t *testing.T) {
	m := &inlineMode{screenRows: 24, opts: InlineOptions{Mouse: true}}
	m.forget()
	require.True(t, m.origin.ask)

	// The question goes out with a frame whose cursor is on the region's
	// second row; two frames follow before the answer.
	m.origin.ask, m.origin.outstanding, m.origin.queryRow = false, 1, 1
	m.moved(2, 3)
	m.moved(0, 3)
	require.False(t, m.origin.known, "unknown until answered")

	assert.True(t, m.handle(uv.CursorPositionEvent{X: 0, Y: 11}))
	assert.True(t, m.origin.known)
	assert.Equal(t, 12, m.origin.row, "the region started on row 10, then two rows were printed")
}

func TestInlineOrigin_OnlyTheLatestAnswerCounts(t *testing.T) {
	m := &inlineMode{screenRows: 24}
	m.origin.outstanding = 2

	assert.True(t, m.handle(uv.CursorPositionEvent{Y: 3}), "the answer is consumed")
	assert.False(t, m.origin.known, "an answer to an older question is ignored")
	assert.True(t, m.handle(uv.CursorPositionEvent{Y: 7}))
	assert.True(t, m.origin.known)
	assert.Equal(t, 7, m.origin.row)

	assert.False(t, m.handle(uv.CursorPositionEvent{Y: 1}), "unasked reports are left for the app")
	assert.Equal(t, 7, m.origin.row)
}

func TestInlineLocate_MakesMouseRowsRelativeToTheRegion(t *testing.T) {
	m := &inlineMode{}
	_, placed := m.locate(uv.MouseClickEvent{X: 3, Y: 9})
	assert.False(t, placed, "mouse events wait for the region's place to be known")

	key := uv.KeyPressEvent{Code: 'a'}
	got, placed := m.locate(key)
	assert.True(t, placed)
	assert.Equal(t, uv.Event(key), got, "other events pass through")

	m.origin = inlineOrigin{row: 6, known: true}
	cases := []uv.Event{
		uv.MouseClickEvent{X: 3, Y: 9},
		uv.MouseReleaseEvent{X: 3, Y: 9},
		uv.MouseMotionEvent{X: 3, Y: 9},
		uv.MouseWheelEvent{X: 3, Y: 9},
	}
	for _, ev := range cases {
		got, placed := m.locate(ev)
		require.True(t, placed)
		assert.IsType(t, ev, got)
		var mouse uv.Mouse
		switch e := got.(type) {
		case uv.MouseClickEvent:
			mouse = uv.Mouse(e)
		case uv.MouseReleaseEvent:
			mouse = uv.Mouse(e)
		case uv.MouseMotionEvent:
			mouse = uv.Mouse(e)
		case uv.MouseWheelEvent:
			mouse = uv.Mouse(e)
		}
		assert.Equal(t, uv.Mouse{X: 3, Y: 3}, mouse, "%T", ev)
	}
}

func TestInlineLocate_ClicksReachTheRegionAndOutsideHitsNothing(t *testing.T) {
	presses := 0
	root := Column{Children: []Widget{
		Text{Content: "header"},
		Button{ID: "ok", Label: "OK", OnPress: func() { presses++ }},
	}}
	m, r, focus := newInlineScene(t, root, 20, 24, InlineOptions{Mouse: true})
	m.origin = inlineOrigin{row: 10, known: true}
	router := newMouseRouter(r, focus, NewAnySignal[Widget](nil))
	now := time.Now()
	click := func(x, y int) {
		now = now.Add(time.Second)
		press, _ := m.locate(uv.MouseClickEvent{X: x, Y: y, Button: uv.MouseLeft})
		release, _ := m.locate(uv.MouseReleaseEvent{X: x, Y: y, Button: uv.MouseLeft})
		router.press(press.(uv.MouseClickEvent), 0.5, 0.5, now)
		router.release(release.(uv.MouseReleaseEvent), 0.5, 0.5)
		focus.SetFocusables(r.Update(root))
	}

	click(1, 11)
	assert.Equal(t, 1, presses, "the button is on the region's second row, screen row 11")

	assert.NotPanics(t, func() {
		click(1, 1)  // above the region
		click(1, 23) // below it
		wheel, _ := m.locate(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelUp})
		router.wheelAt(wheel.(uv.MouseWheelEvent), 0.5, 0.5)
		motion, _ := m.locate(uv.MouseMotionEvent{X: 1, Y: 2})
		router.motion(motion.(uv.MouseMotionEvent), 0.5, 0.5)
	})
	assert.Equal(t, 1, presses, "clicks outside the region hit nothing")
}

// fakeInlineTerminal records what the live region draws.
type fakeInlineTerminal struct {
	*uv.Buffer
	x, y    int
	out     strings.Builder
	erased  bool
	resizes int
}

func newFakeInlineTerminal(width, height int) *fakeInlineTerminal {
	return &fakeInlineTerminal{Buffer: uv.NewBuffer(width, height)}
}

func (f *fakeInlineTerminal) Resize(width, height int) error {
	f.resizes++
	// Like uv.Terminal, a resize marks every row to be redrawn.
	f.Touched = nil
	f.Buffer.Resize(width, height)
	return nil
}
func (f *fakeInlineTerminal) Position() (int, int)              { return f.x, f.y }
func (f *fakeInlineTerminal) SetPosition(x, y int)              { f.x, f.y = x, y }
func (f *fakeInlineTerminal) WriteString(s string) (int, error) { return f.out.WriteString(s) }
func (f *fakeInlineTerminal) Erase()                            { f.erased = true; f.Clear() }
func (f *fakeInlineTerminal) ClearArea(area uv.Rectangle)       { f.Buffer.ClearArea(area) }
func (f *fakeInlineTerminal) CellAt(x, y int) *uv.Cell          { return f.Buffer.CellAt(x, y) }
func (f *fakeInlineTerminal) SetCell(x, y int, c *uv.Cell)      { f.Buffer.SetCell(x, y, c) }
func (f *fakeInlineTerminal) Bounds() uv.Rectangle              { return f.Buffer.Bounds() }

func bufferRow(b *uv.Buffer, y int) string {
	var sb strings.Builder
	for x := 0; x < b.Width(); x++ {
		if c := b.CellAt(x, y); c != nil && c.Content != "" {
			sb.WriteString(c.Content)
		} else if c != nil && c.Width == 0 {
			continue
		} else {
			sb.WriteByte(' ')
		}
	}
	return strings.TrimRight(sb.String(), " ")
}

func TestInlineShow_SizesTheTerminalToTheRegion(t *testing.T) {
	root := Column{Children: []Widget{Text{Content: "one"}, Text{Content: "two"}}}
	m, r, _ := newInlineScene(t, root, 10, 24, InlineOptions{})
	term := newFakeInlineTerminal(10, 1)
	term.SetCell(0, 0, &uv.Cell{Content: "x", Width: 1})

	m.show(term, m.regionHeight(r.contentHeight()), 0)
	assert.Equal(t, 2, term.Height())
	assert.Equal(t, []string{"one", "two"}, []string{bufferRow(term.Buffer, 0), bufferRow(term.Buffer, 1)})
	assert.Equal(t, 2, m.height)

	// On exit the frame is kept with a blank row below it.
	m.show(term, m.height, 1)
	assert.Equal(t, 3, term.Height())
	assert.Equal(t, "two", bufferRow(term.Buffer, 1))
	assert.Equal(t, "", bufferRow(term.Buffer, 2))

	// A cleared region is a single blank row.
	m.show(term, 0, 1)
	assert.Equal(t, 1, term.Height())
	assert.Equal(t, "", bufferRow(term.Buffer, 0))
}

func TestInlineShow_UnchangedHeightLeavesTheTerminalAlone(t *testing.T) {
	m, r, _ := newInlineScene(t, rows(3), 10, 24, InlineOptions{})
	term := newFakeInlineTerminal(10, 1)
	m.show(term, m.regionHeight(r.contentHeight()), 0)
	resizes := term.resizes
	term.Touched = make([]*uv.LineData, 3)

	m.show(term, m.regionHeight(r.contentHeight()), 0)
	assert.Equal(t, resizes, term.resizes)
	for y, line := range term.Touched {
		assert.Nil(t, line, "row %d was not changed, so it is not redrawn", y)
	}
}

func withPrintAboveOpen(t *testing.T) {
	t.Helper()
	openPrintAbove()
	t.Cleanup(func() { takePrintAbove(true) })
}

func TestPrintAbove_QueuesWhileInlineAndAsksForAFrame(t *testing.T) {
	withPrintAboveOpen(t)
	trigger := make(chan struct{}, 1)
	previous := swapRenderTrigger(trigger)
	t.Cleanup(func() { swapRenderTrigger(previous) })

	PrintAboveText("first\nsecond\n")
	PrintAbove(Text{Content: "widget"})

	select {
	case <-trigger:
	default:
		t.Fatal("printing above schedules a frame")
	}
	items := takePrintAbove(false)
	require.Len(t, items, 2)
	assert.Equal(t, []string{"first", "second"}, items[0].lines(20, 24), "one trailing newline is dropped")
	assert.Empty(t, takePrintAbove(false), "taking the queue empties it")
}

func TestPrintAboveItem_RendersWidgetsAtTheTerminalWidth(t *testing.T) {
	item := printAboveItem{widget: Text{Content: "a long line of words that wraps", Wrap: WrapSoft}}
	lines := item.lines(12, 24)
	plain := make([]string, len(lines))
	for i, line := range lines {
		plain[i] = strings.TrimRight(ansi.Strip(line), " ")
	}
	assert.Equal(t, []string{"a long line", "of words", "that wraps"}, plain)

	assert.Empty(t, printAboveItem{widget: Column{}}.lines(12, 24), "an empty widget prints nothing")
}

func TestPrintAboveItem_TallWidgetsAreNotCutAtTheScreenHeight(t *testing.T) {
	assert.Len(t, printAboveItem{widget: rows(40)}.lines(12, 10), 40)
}

func TestInlinePrintAbove_WritesLinesWhereTheRegionWas(t *testing.T) {
	m := &inlineMode{width: 10}
	term := newFakeInlineTerminal(10, 3)
	term.SetPosition(4, 2)

	printed := m.printAbove(term, []printAboveItem{
		{text: "exactly10!\n" + strings.Repeat("w", 15)},
		{text: "last"},
	})

	assert.Equal(t, 3, printed, "a line as wide as the terminal takes one row; wider ones are cut")
	want := "\r" + ansi.CursorUp(2) + ansi.EraseScreenBelow +
		"exactly10!" + ansi.ResetStyle + "\r\n" +
		strings.Repeat("w", 10) + ansi.ResetStyle + "\r\n" +
		"last" + ansi.ResetStyle + "\r\n"
	assert.Equal(t, want, term.out.String())
	assert.True(t, term.erased, "the region is drawn again below the lines")
	x, y := term.Position()
	assert.Equal(t, [2]int{-1, -1}, [2]int{x, y}, "the region starts again where the cursor is")
}

func TestInlinePrintAbove_NothingQueuedWritesNothing(t *testing.T) {
	m := &inlineMode{width: 10}
	term := newFakeInlineTerminal(10, 3)
	assert.Equal(t, 0, m.printAbove(term, nil))
	assert.Empty(t, term.out.String())
	assert.False(t, term.erased)
}

func TestInlineShow_GrowingByABlankRowKeepsTheRowAboveIt(t *testing.T) {
	root := Column{Children: []Widget{Text{Content: "one"}, Text{Content: "two"}}}
	m, _, _ := newInlineScene(t, root, 10, 24, InlineOptions{})
	term := newFakeInlineTerminal(10, 1)
	var out bytes.Buffer
	scr := uv.NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	scr.ExitAltScreen()
	display := func() string {
		out.Reset()
		scr.Render(term.Buffer)
		require.NoError(t, scr.Flush())
		return out.String()
	}
	m.show(term, 2, 0)
	display()

	// What exit does: keep the frame and add a blank row below it.
	m.show(term, 2, 1)
	got := display()
	if strings.Contains(got, ansi.EraseScreenBelow) {
		assert.Contains(t, got, "two", "a row uv erases while growing is drawn again")
	}
}
