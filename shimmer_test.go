package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	shimmerBlack = RGB(0, 0, 0)
	shimmerWhite = RGB(255, 255, 255)
)

func blackToWhite(state *ShimmerState, path ShimmerPath, bandWidth int) Shimmer {
	return Shimmer{State: state, Base: shimmerBlack, Highlight: shimmerWhite, BandWidth: bandWidth, Path: path}
}

type cellPos struct{ x, y int }

func fgAt(buf *uv.Buffer, c cellPos) Color {
	return FromANSI(buf.CellAt(c.x, c.y).Style.Fg)
}

func brightestFg(buf *uv.Buffer, cells []cellPos) cellPos {
	brightest := cells[0]
	for _, c := range cells {
		if fgAt(buf, c).r > fgAt(buf, brightest).r {
			brightest = c
		}
	}
	return brightest
}

func rowCells(width int) []cellPos {
	cells := make([]cellPos, width)
	for x := range cells {
		cells[x] = cellPos{x, 0}
	}
	return cells
}

func TestShimmer_SweepTravelsLeftToRight(t *testing.T) {
	state := NewShimmerState(time.Second)
	widget := Text{Content: "abcdefghijklmnopqrst", Style: Style{ForegroundColor: blackToWhite(state, ShimmerSweep, 4)}}

	var positions []int
	// The band travels 20+4 cells; these phases center it on cells 4, 10 and 15.
	for _, phase := range []float64{6.5 / 24, 12.5 / 24, 17.5 / 24} {
		state.phase.Set(phase)
		buf := RenderToBuffer(widget, 20, 1)
		require.Equal(t, shimmerBlack, fgAt(buf, cellPos{0, 0}), "cells outside the band keep the base color")
		require.Equal(t, shimmerBlack, fgAt(buf, cellPos{19, 0}), "cells outside the band keep the base color")
		positions = append(positions, brightestFg(buf, rowCells(20)).x)
	}
	assert.Equal(t, []int{4, 10, 15}, positions)
}

func TestShimmer_SweepCycleBoundariesAndStoppedShowBaseColor(t *testing.T) {
	state := NewShimmerState(time.Second)
	widget := Text{Content: "loading", Style: Style{ForegroundColor: blackToWhite(state, ShimmerSweep, 4)}}
	allBase := func(name string) {
		buf := RenderToBuffer(widget, 7, 1)
		for _, c := range rowCells(7) {
			require.Equal(t, shimmerBlack, fgAt(buf, c), "%s: cell %d", name, c.x)
		}
	}

	allBase("stopped")
	state.phase.Set(0)
	allBase("cycle start")
	state.phase.Set(0.999)
	allBase("cycle end")
}

// A 12x4 box has a perimeter of 11+6+11+6 = 34, counting each vertical step as two.
func perimeterBox(state *ShimmerState, bandWidth int) Widget {
	return Column{Style: Style{
		Border: SquareBorder(blackToWhite(state, ShimmerPerimeter, bandWidth)),
		Width:  Cells(12),
		Height: Cells(4),
	}}
}

func borderCells(width, height int) []cellPos {
	var cells []cellPos
	for y := range height {
		for x := range width {
			if y == 0 || y == height-1 || x == 0 || x == width-1 {
				cells = append(cells, cellPos{x, y})
			}
		}
	}
	return cells
}

func TestShimmer_PerimeterTravelsClockwise(t *testing.T) {
	state := NewShimmerState(time.Second)
	widget := perimeterBox(state, 4)

	var positions []cellPos
	for _, distance := range []float64{5, 15, 25, 32} {
		state.phase.Set(distance / 34)
		positions = append(positions, brightestFg(RenderToBuffer(widget, 12, 4), borderCells(12, 4)))
	}
	assert.Equal(t, []cellPos{{5, 0}, {11, 2}, {3, 3}, {0, 1}}, positions, "top, right, bottom, then left edge")
}

func TestShimmer_PerimeterWrapsAcrossTheTopLeftCorner(t *testing.T) {
	state := NewShimmerState(time.Second)
	state.phase.Set(0)
	buf := RenderToBuffer(perimeterBox(state, 6), 12, 4)

	assert.Equal(t, shimmerWhite, fgAt(buf, cellPos{0, 0}), "the band is centered on the corner")
	assert.NotEqual(t, shimmerBlack, fgAt(buf, cellPos{1, 0}), "the band reaches along the top edge")
	assert.NotEqual(t, shimmerBlack, fgAt(buf, cellPos{0, 1}), "the band wraps onto the end of the left edge")
	assert.Equal(t, shimmerBlack, fgAt(buf, cellPos{11, 0}), "the far corner stays base")
}

func TestShimmerState_AdvanceWrapsAtPeriod(t *testing.T) {
	state := NewShimmerState(time.Second)
	state.running = true

	assert.True(t, state.Advance(250*time.Millisecond))
	assert.InDelta(t, 0.25, state.phase.Peek(), 1e-9)
	state.Advance(time.Second)
	assert.InDelta(t, 0.25, state.phase.Peek(), 1e-9)

	state.Stop()
	assert.False(t, state.Advance(time.Millisecond))
	assert.Equal(t, shimmerIdle, state.phase.Peek())
}

func TestRenderer_ShimmerTickUsesPartialPaint(t *testing.T) {
	cases := map[string]func(*ShimmerState) Widget{
		"text": func(s *ShimmerState) Widget {
			return Text{Content: "loading", Style: Style{ForegroundColor: blackToWhite(s, ShimmerSweep, 4)}}
		},
		"border": func(s *ShimmerState) Widget { return perimeterBox(s, 4) },
		"background": func(s *ShimmerState) Widget {
			return Column{Style: Style{BackgroundColor: blackToWhite(s, ShimmerSweep, 4), Width: Cells(12), Height: Cells(4)}}
		},
	}
	for name, widget := range cases {
		t.Run(name, func(t *testing.T) {
			screen := newTrackingScreen(12, 4)
			renderer := newTestRenderer(screen, 12, 4)
			state := NewShimmerState(time.Second)
			builds := 0
			root := buildCountWidget{builds: &builds, child: widget(state)}
			renderer.Update(root)

			before := styledCells(screen.Buffer)
			state.phase.Set(0.5)
			renderer.Update(root)

			assert.Equal(t, rendererFramePartial, renderer.lastFrameMode)
			assert.Equal(t, 1, builds, "a shimmer tick should not rebuild")
			assert.Equal(t, 0, renderer.lastLayoutCount, "a shimmer tick should not relayout")
			assert.NotEqual(t, before, styledCells(screen.Buffer), "the highlight should be painted")
		})
	}
}

func styledCells(buf *uv.Buffer) []uv.Style {
	var styles []uv.Style
	for y := range buf.Height() {
		for x := range buf.Width() {
			styles = append(styles, buf.CellAt(x, y).Style)
		}
	}
	return styles
}

func TestShimmer_Snapshots(t *testing.T) {
	theme := getTheme()
	state := NewShimmerState(time.Second)
	state.phase.Set(0.5)
	AssertSnapshotNamed(t, "shimmer_sweep_text", Text{
		Content: "Thinking about your request...",
		Style:   Style{ForegroundColor: Shimmer{State: state, Base: theme.TextMuted, Highlight: theme.Text}},
	}, 32, 1, "Muted text with a brighter band centered on the middle of the line")

	state.phase.Set(0.3)
	AssertSnapshotNamed(t, "shimmer_perimeter_border", Column{
		Style: Style{
			Border:  RoundedBorder(Shimmer{State: state, Base: theme.Border, Highlight: theme.Accent, BandWidth: 10, Path: ShimmerPerimeter}),
			Padding: EdgeInsetsXY(1, 0),
		},
		Children: []Widget{Text{Content: "Indexing workspace..."}},
	}, 30, 5, "Rounded border in the border color, with an accent band toward the right end of the top edge")
}
