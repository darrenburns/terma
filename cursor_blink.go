package terma

import (
	"math"
	"sync/atomic"
	"time"

	uv "github.com/darrenburns/ultraviolet"
)

const (
	cursorBlinkInterval = 530 * time.Millisecond
	cursorPulseInterval = time.Second / 30
	cursorPulseHold     = 400 * time.Millisecond
	cursorPulsePeriod   = 1600 * time.Millisecond
	cursorPulseSteps    = 32
	cursorHiddenLevel   = cursorPulseSteps + 1
)

// CursorBlinkStyle controls how an enabled text cursor blinks.
type CursorBlinkStyle int32

const (
	// CursorBlinkClassic alternates between a shown and hidden cursor (the default).
	CursorBlinkClassic CursorBlinkStyle = iota
	// CursorBlinkPulse smoothly dims the cursor using [CursorPulseConfig].
	CursorBlinkPulse
)

// CursorPulseConfig controls [CursorBlinkPulse] for TextInput and TextArea.
type CursorPulseConfig struct {
	// MinOpacity is the cursor's minimum blend into its background, from 0 to 1.
	// Zero allows it to fade completely; 1 keeps it at full brightness.
	// The app default is 0.35. Values are clamped to [0, 1]; NaN uses the default.
	MinOpacity float64
	// Easing applies separately to each half-cycle: full to minimum brightness,
	// then minimum to full. Nil uses EaseInOutSine, the app default.
	// Overshooting curves are clamped to keep opacity within the configured range.
	Easing EasingFunc
	// Period is the duration of a complete full → minimum → full cycle.
	// Zero or negative values use the default of 1600 ms. The 400 ms hold after
	// input is separate and does not scale with Period.
	Period time.Duration
}

var cursorBlinkEnabled atomic.Bool
var cursorBlinkStyle atomic.Int32
var cursorPulseConfig atomic.Pointer[CursorPulseConfig]

// textCursor caches the fully resolved cursor glyph during normal painting.
// The palette makes pulse ticks allocation-free and independent of document
// size. A wide grapheme occupies two terminal columns but is still one uv.Cell.
type textCursor struct {
	valid      bool
	prepared   bool
	minOpacity float64
	x, y       int
	focusID    string
	level      int
	frames     [cursorPulseSteps + 2]uv.Cell
}

// SetCursorBlink turns blinking of the text cursor in TextInput and
// TextArea on or off for the whole app. It is off by default (a steady cursor).
// The default style is classic on/off blinking; use [SetCursorBlinkStyle] to
// select a smooth pulse instead.
// Typing, pasting, clicking and changing focus restore full brightness briefly.
// Animation updates only the cached cursor cell, without building, measuring,
// painting widgets or running text highlighters.
func SetCursorBlink(enabled bool) {
	if cursorBlinkEnabled.Swap(enabled) != enabled {
		scheduleRender()
	}
}

// CursorBlink reports whether the text cursor blinks. See [SetCursorBlink].
func CursorBlink() bool {
	return cursorBlinkEnabled.Load()
}

// SetCursorBlinkStyle selects classic blinking or smooth pulsing for the whole
// app. It does not enable blinking; use [SetCursorBlink] for that. Changing the
// style while blinking restarts the cursor at full brightness.
func SetCursorBlinkStyle(style CursorBlinkStyle) {
	if cursorBlinkStyle.Swap(int32(style)) != int32(style) {
		scheduleRender()
	}
}

// CurrentCursorBlinkStyle returns the selected style, even if blinking is off.
func CurrentCursorBlinkStyle() CursorBlinkStyle {
	return CursorBlinkStyle(cursorBlinkStyle.Load())
}

// SetCursorPulseConfig configures the pulse for the whole app without changing
// the blink style or enabling blinking. It restarts the cursor at full brightness.
// The 400 ms hold after input is unchanged.
func SetCursorPulseConfig(config CursorPulseConfig) {
	if math.IsNaN(config.MinOpacity) {
		config.MinOpacity = 0.35
	}
	config.MinOpacity = clamp01(config.MinOpacity)
	if config.Easing == nil {
		config.Easing = EaseInOutSine
	}
	if config.Period <= 0 {
		config.Period = cursorPulsePeriod
	}
	cursorPulseConfig.Store(&config)
	scheduleRender()
}

// CurrentCursorPulseConfig returns the pulse settings, even if blinking is off.
// Modify the returned copy and call [SetCursorPulseConfig] to update one setting
// while preserving the others.
func CurrentCursorPulseConfig() CursorPulseConfig {
	if config := cursorPulseConfig.Load(); config != nil {
		return *config
	}
	return CursorPulseConfig{MinOpacity: 0.35, Easing: EaseInOutSine, Period: cursorPulsePeriod}
}

func (s *appSession) resetCursorPulse() {
	s.cursorStarted = s.now()
}

func (s *appSession) cursorPulseActive() bool {
	cursor := &s.renderer.textCursor
	return CursorBlink() && cursor.valid && cursor.focusID == s.focus.FocusedID()
}

func (s *appSession) cursorPulseLevel() int {
	if CurrentCursorBlinkStyle() != CursorBlinkPulse {
		if s.now().Sub(s.cursorStarted)/cursorBlinkInterval%2 != 0 {
			return cursorHiddenLevel
		}
		return cursorPulseSteps
	}
	elapsed := s.now().Sub(s.cursorStarted) - cursorPulseHold
	if elapsed <= 0 {
		return cursorPulseSteps
	}
	config := CurrentCursorPulseConfig()
	phase := float64(elapsed%config.Period) / float64(config.Period)
	var level float64
	if phase < 0.5 {
		level = 1 - config.Easing(phase*2)
	} else {
		level = config.Easing(phase*2 - 1)
	}
	if math.IsNaN(level) {
		return cursorPulseSteps
	}
	return int(math.Round(float64(cursorPulseSteps) * clamp01(level)))
}

// advanceCursorPulse is the entire pulse frame: no signal notifications,
// session.frame, Renderer.Update, tree traversal, or widget Render calls.
func (s *appSession) advanceCursorPulse() bool {
	r := s.renderer
	if !s.cursorPulseActive() || !r.paintTextCursor(s.cursorPulseLevel()) {
		return false
	}
	r.lastFrameMode = rendererFrameCursor
	r.lastBuildCount, r.lastLayoutCount, r.lastPaintCount = 0, 0, 0
	r.lastAssignCount, r.lastMeasureCount, r.lastClearCount, r.lastScanCount = 0, 0, 0, 0
	r.lastDamagedRects = append(r.lastDamagedRects[:0], r.textCursor.bounds())
	return true
}

// drawTextCursor uses the theme's caret color, resolving alpha against the
// widget background. Keeping the glyph color fixed avoids losing it halfway
// through a foreground/background swap.
// underlying optionally supplies the text style to restore when hidden; nil
// preserves the existing cell (a placeholder or the space after the text).
func drawTextCursor(ctx *RenderContext, x, y int, text string, style Style, underlying *Style) {
	var hidden uv.Cell
	if cell := ctx.terminal.CellAt(ctx.X+x, ctx.Y+y); cell != nil {
		hidden = *cell
		if underlying != nil {
			background := hidden
			ctx.DrawStyledText(x, y, text, *underlying)
			hidden = *ctx.terminal.CellAt(ctx.X+x, ctx.Y+y)
			// Resolve both styles against the original background, so translucent
			// highlights aren't blended twice or inherited by the shown cursor.
			ctx.terminal.SetCell(ctx.X+x, ctx.Y+y, &background)
		}
	}
	style.ForegroundColor = ctx.buildContext.Theme().Cursor
	style.Reverse = true
	ctx.DrawStyledText(x, y, text, style)
	r := ctx.buildContext.renderer
	x, y = ctx.X+x, ctx.Y+y
	if r == nil || !ctx.clip.Contains(x, y) {
		return
	}
	cell := ctx.terminal.CellAt(x, y)
	if cell == nil || !ctx.clip.Contains(x+cell.Width-1, y) {
		return
	}
	cursor := &r.textCursor
	cursor.valid, cursor.x, cursor.y = true, x, y
	cursor.focusID, cursor.level = ctx.currentEventID, cursorPulseSteps
	cursor.prepared = cursor.prepared && cell.Equal(&cursor.frames[cursorPulseSteps])
	cursor.frames[cursorPulseSteps] = *cell
	cursor.frames[cursorHiddenLevel] = hidden
	if CursorBlink() && CurrentCursorBlinkStyle() == CursorBlinkPulse {
		cursor.prepare()
	}
}

func (cursor *textCursor) prepare() {
	minOpacity := CurrentCursorPulseConfig().MinOpacity
	if cursor.prepared && cursor.minOpacity == minOpacity {
		return
	}
	cell := &cursor.frames[cursorPulseSteps]
	fg, bg := FromANSI(cell.Style.Fg), FromANSI(cell.Style.Bg)
	for i := range cursorPulseSteps {
		painted := *cell
		painted.Style.Attrs &^= uv.AttrReverse
		painted.Style.Fg = bg.toANSI()
		painted.Style.Bg = bg.Blend(fg, minOpacity+(1-minOpacity)*float64(i)/cursorPulseSteps).toANSI()
		cursor.frames[i] = painted
	}
	cursor.minOpacity = minOpacity
	cursor.prepared = true
}

func (c *textCursor) bounds() Rect {
	return Rect{X: c.x, Y: c.y, Width: c.frames[cursorPulseSteps].Width, Height: 1}
}

// invalidateTextCursor runs only during regular painting. Invalidate before
// each intersecting painter, so an overlay cannot leave a live cursor cache
// beneath it, even if it paints a cell identical to the cursor's base cell.
func (r *Renderer) invalidateTextCursor(area Rect) {
	if r.textCursor.valid && r.textCursor.bounds().Intersects(area) {
		r.textCursor.valid = false
	}
}

func (r *Renderer) paintTextCursor(level int) bool {
	cursor := &r.textCursor
	if !cursor.valid || level == cursor.level {
		return false
	}
	if !cursor.frames[cursor.level].Equal(r.terminal.CellAt(cursor.x, cursor.y)) {
		// Also catch writes outside the widget painter, such as debug overlays.
		cursor.valid = false
		return false
	}
	// Usually prepared by drawTextCursor; enabling animation at runtime can
	// reuse an already-painted steady cursor without repainting its widget.
	if level < cursorPulseSteps {
		cursor.prepare()
	}
	cursor.level = level
	r.terminal.SetCell(cursor.x, cursor.y, &cursor.frames[level])
	if r.presentation != nil {
		r.presentation.SetCell(cursor.x, cursor.y, &cursor.frames[level])
	}
	return true
}
