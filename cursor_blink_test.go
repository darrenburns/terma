package terma

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	uv "github.com/darrenburns/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enableCursorBlinkForTest(t testing.TB) {
	t.Helper()
	SetCursorBlink(true)
	t.Cleanup(func() { SetCursorBlink(false) })
}

func enableCursorPulseForTest(t testing.TB) {
	t.Helper()
	style := CurrentCursorBlinkStyle()
	SetCursorBlinkStyle(CursorBlinkPulse)
	t.Cleanup(func() { SetCursorBlinkStyle(style) })
	enableCursorBlinkForTest(t)
}

func TestCursorBlink_OffByDefault(t *testing.T) {
	assert.False(t, CursorBlink())
	assert.Equal(t, CursorBlinkClassic, CurrentCursorBlinkStyle())
	assert.Equal(t, 0.35, CurrentCursorPulseConfig().MinOpacity)
	assert.Equal(t, 1600*time.Millisecond, CurrentCursorPulseConfig().Period)
}

func TestSnapshot_TextInput_CursorBlinkShown(t *testing.T) {
	enableCursorBlinkForTest(t)
	widget := TextInput{ID: "input", State: NewTextInputState("hello"), Style: Style{Width: Cells(10)}}
	AssertSnapshot(t, widget, 10, 1, "Focused input at full pulse brightness: reverse-video cursor after 'hello'.")
}

func TestCursorBlink_ClassicRestoresUnderlyingCells(t *testing.T) {
	for _, kind := range []string{"input", "area"} {
		for _, scenario := range []string{"end", "wide", "selection", "placeholder"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				enableCursorBlinkForTest(t)
				text := "ab界d"
				if scenario == "placeholder" {
					text = ""
				}
				style := Style{Width: Cells(12), Height: Cells(1), ForegroundColor: RGB(220, 170, 90), BackgroundColor: RGB(20, 40, 60)}
				highlighter := HighlighterFunc(func(string, []string) []TextHighlight {
					return []TextHighlight{{Start: 2, End: 3, Style: SpanStyle{
						Foreground: RGB(40, 200, 100), Background: RGB(100, 20, 180).WithAlpha(0.5),
					}}}
				})
				var widget Widget = TextInput{ID: "editor", State: NewTextInputState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
				if kind == "area" {
					widget = TextArea{ID: "editor", State: NewTextAreaState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
				}
				p := NewPilot(t, Column{Children: []Widget{widget, Button{ID: "button", Label: "OK"}}}, 12, 3)
				if scenario == "wide" {
					p.Press("left", "left")
				} else if scenario == "selection" {
					p.Press("shift+left", "shift+left")
				}
				row := func() []uv.Cell {
					cells := make([]uv.Cell, 12)
					for x := range cells {
						cells[x] = *p.Buffer().CellAt(x, 0)
					}
					return cells
				}
				shown := row()
				if scenario == "wide" || scenario == "selection" {
					assert.Equal(t, RGB(60, 30, 120), FromANSI(shown[2].Style.Bg), "shown cursor resolves the highlight once, without the selection background")
				}
				p.Press("tab")
				hidden := row() // The same widget with no cursor provides the expected text and styles.
				require.NotEqual(t, shown, hidden)
				p.Press("shift+tab")
				p.Advance(529 * time.Millisecond)
				assert.Equal(t, shown, row())
				p.Advance(time.Millisecond)
				assert.Equal(t, hidden, row(), "classic blink removes only the cursor, retaining highlighting and selection")
				assert.Zero(t, p.session.renderer.Stats().PaintCount)
				p.Advance(530 * time.Millisecond)
				assert.Equal(t, shown, row())
				p.Advance(530 * time.Millisecond)
				SetCursorBlink(false)
				p.ScreenText()
				assert.Equal(t, shown, row(), "disabling while hidden restores the steady cursor")
				p.Advance(2 * time.Second)
				assert.Equal(t, shown, row())
			})
		}
	}
}

func TestCursorBlink_StyleAndEnabledAreIndependent(t *testing.T) {
	t.Cleanup(func() {
		SetCursorBlink(false)
		SetCursorBlinkStyle(CursorBlinkClassic)
	})
	p := NewPilot(t, TextInput{ID: "editor", State: NewTextInputState("hi"), Width: Cells(10)}, 10, 1)
	cell := func() uv.Cell { return *p.Buffer().CellAt(2, 0) }
	shown := cell()
	SetCursorBlinkStyle(CursorBlinkPulse)
	assert.False(t, CursorBlink(), "selecting a style must not enable blinking")
	p.Advance(1200 * time.Millisecond)
	assert.Equal(t, shown, cell())
	SetCursorBlink(true)
	p.Advance(1200 * time.Millisecond)
	dim := cell()
	assert.NotEqual(t, shown, dim)
	SetCursorBlinkStyle(CursorBlinkClassic)
	p.ScreenText()
	assert.True(t, CursorBlink(), "changing style must not disable blinking")
	assert.Equal(t, shown, cell(), "changing style restarts the visible phase")
	p.Advance(530 * time.Millisecond)
	assert.NotEqual(t, shown, cell())
	assert.NotEqual(t, dim, cell(), "classic hidden phase is not the pulse trough")
	SetCursorBlinkStyle(CursorBlinkPulse)
	p.ScreenText()
	assert.Equal(t, shown, cell())
	p.Advance(1200 * time.Millisecond)
	assert.Equal(t, dim, cell())
	SetCursorBlink(false)
	p.Advance(2 * time.Second)
	assert.Equal(t, shown, cell())
	assert.Equal(t, CursorBlinkPulse, CurrentCursorBlinkStyle(), "disabling preserves the chosen style")
}

func TestCursorPulse_Config(t *testing.T) {
	originalConfig := CurrentCursorPulseConfig()
	originalTheme := CurrentThemeName()
	RegisterTheme(t.Name(), ExtendTheme(ThemeNameRosePine, WithCursor(RGB(180, 120, 220))))
	SetTheme(t.Name())
	t.Cleanup(func() {
		SetCursorPulseConfig(originalConfig)
		SetTheme(originalTheme)
		delete(themeRegistry, t.Name())
	})
	for _, kind := range []string{"input", "area"} {
		t.Run(kind, func(t *testing.T) {
			SetCursorPulseConfig(CursorPulseConfig{MinOpacity: 0.1, Easing: EaseInQuad})
			assert.False(t, CursorBlink(), "configuring the pulse must not enable blinking")
			assert.Equal(t, CursorBlinkClassic, CurrentCursorBlinkStyle())
			enableCursorPulseForTest(t)
			style := Style{Width: Cells(10), Height: Cells(1), BackgroundColor: RGB(20, 40, 60)}
			var widget Widget = TextInput{ID: "editor", State: NewTextInputState("hi"), Style: style}
			if kind == "area" {
				widget = TextArea{ID: "editor", State: NewTextAreaState("hi"), Style: style}
			}
			p := NewPilot(t, widget, 10, 1)
			cell := func() uv.Cell { return *p.Buffer().CellAt(2, 0) }
			shown := cell()
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, shown, cell(), "hold at full brightness")
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, RGB(144, 102, 184), FromANSI(cell().Style.Bg), "halfway down: 77.5% opacity with quadratic easing")
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, RGB(36, 48, 76), FromANSI(cell().Style.Bg), "10% trough")
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, RGB(72, 66, 112), FromANSI(cell().Style.Bg), "halfway up: 32.5%, not the mirrored fade-out")
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, shown, cell(), "full brightness at the end of the cycle")
			p.Advance(800 * time.Millisecond)

			config := CurrentCursorPulseConfig()
			config.MinOpacity = 0.25
			assert.Equal(t, 0.1, CurrentCursorPulseConfig().MinOpacity, "the getter returns a copy")
			SetCursorPulseConfig(config)
			p.ScreenText()
			assert.Equal(t, shown, cell(), "a live config change restarts at full brightness")
			p.Advance(400 * time.Millisecond)
			assert.Equal(t, shown, cell(), "a live config change restarts the hold")
			p.Advance(800 * time.Millisecond)
			assert.Equal(t, RGB(60, 60, 100), FromANSI(cell().Style.Bg), "refresh the cached palette without changing the cursor cell")

			config.Easing = EaseLinear
			SetCursorPulseConfig(config)
			p.ScreenText()
			assert.Equal(t, shown, cell(), "an easing-only change also restarts the pulse")
			p.Advance(600 * time.Millisecond)
			assert.Equal(t, RGB(150, 105, 190), FromANSI(cell().Style.Bg), "one quarter down with linear easing: 81.25% opacity")
			assert.Zero(t, p.session.renderer.Stats().PaintCount, "configured pulses still use the cell-only path")

			for _, period := range []time.Duration{800 * time.Millisecond, 2400 * time.Millisecond} {
				config.Period = period
				SetCursorPulseConfig(config)
				p.ScreenText()
				assert.Equal(t, shown, cell(), "a period-only change restarts the pulse")
				p.Advance(400 * time.Millisecond)
				assert.Equal(t, shown, cell(), "the hold does not scale with the period")
				p.Advance(period / 4)
				assert.Equal(t, RGB(120, 90, 160), FromANSI(cell().Style.Bg), "halfway down at one quarter of the configured period")
				p.Advance(period / 4)
				assert.Equal(t, RGB(60, 60, 100), FromANSI(cell().Style.Bg), "minimum at half the configured period")
				p.Advance(period / 2)
				assert.Equal(t, shown, cell(), "full brightness after one complete period")
				p.Advance(period / 2)
				assert.Equal(t, RGB(60, 60, 100), FromANSI(cell().Style.Bg), "subsequent cycles have no hold")
				assert.Zero(t, p.session.renderer.Stats().PaintCount)
			}
		})
	}
}

func TestCursorPulse_ConfigBounds(t *testing.T) {
	original := CurrentCursorPulseConfig()
	t.Cleanup(func() { SetCursorPulseConfig(original) })
	enableCursorPulseForTest(t)
	for _, period := range []time.Duration{0, -time.Second} {
		SetCursorPulseConfig(CursorPulseConfig{Period: period})
		assert.Equal(t, 1600*time.Millisecond, CurrentCursorPulseConfig().Period, "non-positive periods restore the default")
	}
	for _, test := range []struct {
		name    string
		minimum float64
		want    float64
	}{
		{"below zero", -0.5, 0},
		{"zero", 0, 0},
		{"one", 1, 1},
		{"above one", 1.5, 1},
		{"NaN", math.NaN(), 0.35},
	} {
		t.Run(test.name, func(t *testing.T) {
			SetCursorPulseConfig(CursorPulseConfig{MinOpacity: test.minimum})
			config := CurrentCursorPulseConfig()
			assert.Equal(t, test.want, config.MinOpacity)
			require.NotNil(t, config.Easing)
			assert.InDelta(t, 0.1464466094, config.Easing(0.25), 1e-9, "nil easing restores sine easing")
			p := NewPilot(t, TextInput{State: NewTextInputState(""), Width: Cells(2)}, 2, 1)
			shown := *p.Buffer().CellAt(0, 0)
			p.Advance(1200 * time.Millisecond)
			cell := p.Buffer().CellAt(0, 0)
			bg, fg := FromANSI(shown.Style.Bg), FromANSI(shown.Style.Fg)
			assert.Equal(t, bg.Blend(fg, test.want), FromANSI(cell.Style.Bg))
		})
	}
	SetCursorPulseConfig(CursorPulseConfig{MinOpacity: 0.1, Easing: EaseInBack})
	p := NewPilot(t, TextInput{State: NewTextInputState(""), Width: Cells(2)}, 2, 1)
	shown := *p.Buffer().CellAt(0, 0)
	p.Advance(800 * time.Millisecond)
	assert.Equal(t, shown, *p.Buffer().CellAt(0, 0), "fade-out overshoot clamps to full brightness")
	p.Advance(800 * time.Millisecond)
	cell := p.Buffer().CellAt(0, 0)
	bg, fg := FromANSI(shown.Style.Bg), FromANSI(shown.Style.Fg)
	assert.Equal(t, bg.Blend(fg, 0.1), FromANSI(cell.Style.Bg), "fade-in undershoot clamps to the minimum")
}

// Test through the public widgets and clock: only the cursor's cell may change,
// even when its widget is auto-sized, contains a wide glyph, or has no content.
func TestCursorPulse_CellOnly(t *testing.T) {
	originalTheme := CurrentThemeName()
	RegisterTheme(t.Name(), ExtendTheme(ThemeNameRosePine, WithCursor(RGB(160, 100, 240))))
	SetTheme(t.Name())
	t.Cleanup(func() {
		SetTheme(originalTheme)
		delete(themeRegistry, t.Name())
	})
	for _, kind := range []string{"input", "area"} {
		for _, sizing := range []string{"fixed", "auto"} {
			for _, text := range []string{"", "ab界d"} {
				t.Run(kind+"/"+sizing+"/"+text, func(t *testing.T) {
					enableCursorPulseForTest(t)
					style := Style{ForegroundColor: RGB(220, 200, 160), BackgroundColor: RGB(20, 40, 60)}
					if sizing == "fixed" {
						style.Width = Cells(12)
						style.Height = Cells(2)
					}
					highlights := 0
					highlighter := HighlighterFunc(func(string, []string) []TextHighlight {
						highlights++
						return nil
					})
					var widget Widget
					if kind == "input" {
						widget = TextInput{ID: "editor", State: NewTextInputState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
					} else {
						widget = TextArea{ID: "editor", State: NewTextAreaState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
					}
					p := NewPilot(t, Column{Children: []Widget{widget, Text{Content: "unchanged sibling"}}}, 24, 5)
					positions := []int{0}
					if text != "" {
						positions = []int{5, 2} // End of text, then the two-cell glyph.
					}
					for _, x := range positions {
						if x == 2 {
							p.Press("left", "left")
						}
						if kind == "input" && sizing == "auto" && text != "" {
							// Auto input is five cells wide; it scrolls one cell
							// to make room for the cursor after the five-cell text.
							x--
						}
						before := make([]uv.Cell, 24*5)
						for y := range 5 {
							for col := range 24 {
								before[y*24+col] = *p.Buffer().CellAt(col, y)
							}
						}
						assert.Equal(t, RGB(160, 100, 240), FromANSI(before[x].Style.Fg), "reverse-video cursor uses the theme color, not the text color")
						fullFrames := p.session.renderer.Stats().FullRenderCount
						highlightsBefore := highlights
						p.Advance(1200 * time.Millisecond) // 400 ms hold + half a cycle.
						stats := p.session.renderer.Stats()
						assert.Equal(t, "cursor", stats.FrameMode)
						assert.Zero(t, stats.BuildCount)
						assert.Zero(t, stats.LayoutCount)
						assert.Zero(t, stats.PaintCount)
						assert.Equal(t, highlightsBefore, highlights, "pulses must not rerun highlighting")
						assert.Equal(t, fullFrames, stats.FullRenderCount)
						cell := p.Buffer().CellAt(x, 0)
						require.Equal(t, before[x].Content, cell.Content)
						assert.Equal(t, before[x].Width, cell.Width)
						assert.Equal(t, RGB(20, 40, 60), FromANSI(cell.Style.Fg))
						assert.Equal(t, RGB(69, 61, 123), FromANSI(cell.Style.Bg))
						for y := range 5 {
							for col := range 24 {
								if y == 0 && col >= x && col < x+cell.Width {
									continue
								}
								assert.Equal(t, before[y*24+col], *p.Buffer().CellAt(col, y), "non-cursor cell (%d,%d)", col, y)
							}
						}
					}
				})
			}
		}
	}
}

func TestCursorPulse_ScrollingAndOverlays(t *testing.T) {
	enableCursorPulseForTest(t)
	state := NewTextAreaState(strings.Repeat("row of text\n", 1000))
	state.CursorIndex.Set(0)
	scroll := NewScrollState()
	covered := NewSignal(false)
	style := Style{ForegroundColor: RGB(220, 200, 160), BackgroundColor: RGB(20, 40, 60)}
	highlights := 0
	root := pilotBuilder(func(ctx BuildContext) Widget {
		coverStyle := style
		coverStyle.ForegroundColor = ctx.Theme().Cursor
		coverStyle.Reverse = true
		return Column{Children: []Widget{
			Scrollable{State: scroll, Width: Cells(20), Height: Cells(4), Child: TextArea{
				ID: "editor", State: state, ScrollState: scroll, Width: Flex(1), Style: style,
				Highlighter: HighlighterFunc(func(string, []string) []TextHighlight { highlights++; return nil }),
			}},
			Floating{Visible: covered.Get(), Config: FloatConfig{Position: FloatPositionTopLeft, PointerPassthrough: true},
				Child: Text{Content: "r", Style: coverStyle}},
		}}
	})
	p := NewPilot(t, root, 24, 6)
	buffer := p.Buffer()
	bright := *buffer.CellAt(0, 0)
	initialHighlights := highlights
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(0, 0).Style)
	assert.Equal(t, initialHighlights, highlights, "a large scrolled document is not highlighted on pulse ticks")
	assert.Zero(t, p.session.renderer.Stats().PaintCount)

	scroll.Offset.Set(10)
	p.ScreenText()
	offscreen := *buffer.CellAt(0, 0)
	initialHighlights = highlights
	p.Advance(2 * time.Second)
	assert.Equal(t, offscreen, *buffer.CellAt(0, 0), "no stale cursor after scrolling it out of view")
	assert.Equal(t, initialHighlights, highlights)
	scroll.Offset.Set(0)
	p.ScreenText()
	assert.Equal(t, bright, *buffer.CellAt(0, 0))
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(0, 0).Style, "scrolling back resumes the pulse")

	covered.Set(true)
	p.ScreenText()
	require.Equal(t, bright, *buffer.CellAt(0, 0), "the covering cell deliberately looks identical to the full-brightness cursor")
	p.Advance(2 * time.Second)
	assert.Equal(t, bright, *buffer.CellAt(0, 0), "a passive overlay must not inherit the underlying cursor pulse")
	covered.Set(false)
	p.ScreenText()
	p.Press("right")
	oldPosition, newPosition := *buffer.CellAt(0, 0), *buffer.CellAt(1, 0)
	p.Advance(1200 * time.Millisecond)
	assert.Equal(t, oldPosition, *buffer.CellAt(0, 0), "movement discards the old cached position")
	assert.NotEqual(t, newPosition.Style, buffer.CellAt(1, 0).Style)
	p.Resize(18, 5)
	newPosition = *buffer.CellAt(1, 0)
	p.Advance(400 * time.Millisecond)
	assert.NotEqual(t, newPosition.Style, buffer.CellAt(1, 0).Style, "resize refreshes the cached cell")
}

func TestCursorPulse_ImageCanvas(t *testing.T) {
	enableCursorPulseForTest(t)
	p := NewPilot(t, Row{Children: []Widget{
		TextInput{ID: "input", State: NewTextInputState("hi"), Width: Cells(10)},
		Image{Source: testImage(t, 2, 2), Style: Style{Width: Cells(2), Height: Cells(2)}},
	}}, 12, 2)
	buffer := p.Buffer()
	bright, imageCell := *buffer.CellAt(2, 0), *buffer.CellAt(10, 0)
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(2, 0).Style, "the cursor reaches the presentation buffer without copying the image canvas")
	assert.Equal(t, imageCell, *buffer.CellAt(10, 0))
	assert.Zero(t, p.session.renderer.Stats().PaintCount)
}

func TestCursorPulse_ThemeChanges(t *testing.T) {
	for _, kind := range []string{"input", "area"} {
		t.Run(kind, func(t *testing.T) {
			originalTheme := CurrentThemeName()
			t.Cleanup(func() { SetTheme(originalTheme) })
			SetTheme(ThemeNameAurora)
			style := Style{BackgroundColor: RGB(20, 40, 60)}
			var widget Widget = TextInput{ID: "editor", State: NewTextInputState("hi"), Width: Cells(10), Style: style}
			if kind == "area" {
				widget = TextArea{ID: "editor", State: NewTextAreaState("hi"), Width: Cells(10), Style: style}
			}
			p := NewPilot(t, widget, 10, 2)
			assert.Equal(t, Hex("#45FFB3"), FromANSI(p.Buffer().CellAt(2, 0).Style.Fg), "steady cursor uses Aurora's caret color")
			enableCursorPulseForTest(t)
			p.Advance(1200 * time.Millisecond)
			SetTheme(ThemeNameMidnightEmber)
			p.ScreenText()  // Repaint after the theme changes, refreshing the palette.
			p.Press("left") // Restart at full brightness over a glyph.
			assert.Equal(t, Hex("#FF9A5C"), FromANSI(p.Buffer().CellAt(1, 0).Style.Fg))
			p.Advance(1200 * time.Millisecond)
			assert.Equal(t, RGB(102, 80, 71), FromANSI(p.Buffer().CellAt(1, 0).Style.Bg), "35% Midnight Ember cursor over the explicit widget background")
			assert.Zero(t, p.session.renderer.Stats().PaintCount)
		})
	}
}

func BenchmarkCursorPulse(b *testing.B) {
	for _, lines := range []int{10, 10000} {
		b.Run(fmt.Sprintf("%d_lines", lines), func(b *testing.B) {
			enableCursorPulseForTest(b)
			state := NewTextAreaState(strings.Repeat("a line of text\n", lines))
			state.CursorIndex.Set(0)
			scroll := NewScrollState()
			p := NewPilot(b, Scrollable{State: scroll, Width: Cells(80), Height: Cells(24), Child: TextArea{
				ID: "editor", State: state, ScrollState: scroll, Width: Flex(1),
			}}, 80, 24)
			p.Advance(time.Second) // Warm the cursor and damage bookkeeping.
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.now = p.now.Add(cursorPulseInterval)
				p.session.advanceCursorPulse()
			}
		})
	}
}
