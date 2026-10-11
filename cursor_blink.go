package terma

import (
	"math"
	"sync/atomic"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

const (
	cursorPulseInterval = time.Second / 30
	cursorPulseHold     = 400 * time.Millisecond
	cursorPulsePeriod   = 1600 * time.Millisecond
	cursorPulseSteps    = 32
)

var cursorBlinkEnabled atomic.Bool

// textCursor caches the fully resolved cursor glyph during normal painting.
// The palette makes pulse ticks allocation-free and independent of document
// size. A wide grapheme occupies two terminal columns but is still one uv.Cell.
type textCursor struct {
	valid    bool
	prepared bool
	x, y     int
	focusID  string
	level    int
	frames   [cursorPulseSteps + 1]uv.Cell
}

// SetCursorBlink turns gentle pulsing of the text cursor in TextInput and
// TextArea on or off for the whole app. It is off by default (a steady cursor).
// The cursor fades between full and reduced brightness without disappearing.
// Typing, pasting, clicking and changing focus restore full brightness briefly.
// Animation updates only the cached cursor cell, without building, measuring,
// painting widgets or running text highlighters.
func SetCursorBlink(enabled bool) {
	if cursorBlinkEnabled.Swap(enabled) != enabled {
		scheduleRender()
	}
}

// CursorBlink reports whether the text cursor pulses. See [SetCursorBlink].
func CursorBlink() bool {
	return cursorBlinkEnabled.Load()
}

func (s *appSession) resetCursorPulse() {
	s.cursorStarted = s.now()
}

func (s *appSession) cursorPulseActive() bool {
	cursor := &s.renderer.textCursor
	return CursorBlink() && cursor.valid && cursor.focusID == s.focus.FocusedID()
}

func (s *appSession) cursorPulseLevel() int {
	elapsed := s.now().Sub(s.cursorStarted) - cursorPulseHold
	if elapsed <= 0 {
		return cursorPulseSteps
	}
	phase := float64(elapsed%cursorPulsePeriod) / float64(cursorPulsePeriod)
	// A cosine eases both turnarounds. Keep a visible block at the trough.
	return int(math.Round(float64(cursorPulseSteps) * (0.5 + 0.5*math.Cos(2*math.Pi*phase))))
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
func drawTextCursor(ctx *RenderContext, x, y int, text string, style Style) {
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
	if CursorBlink() {
		cursor.prepare()
	}
}

func (cursor *textCursor) prepare() {
	if cursor.prepared {
		return
	}
	cell := &cursor.frames[cursorPulseSteps]
	fg, bg := FromANSI(cell.Style.Fg), FromANSI(cell.Style.Bg)
	for i := range cursorPulseSteps {
		painted := *cell
		painted.Style.Attrs &^= uv.AttrReverse
		painted.Style.Fg = bg.toANSI()
		painted.Style.Bg = bg.Blend(fg, 0.35+0.65*float64(i)/cursorPulseSteps).toANSI()
		cursor.frames[i] = painted
	}
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
	cursor.prepare()
	cursor.level = level
	r.terminal.SetCell(cursor.x, cursor.y, &cursor.frames[level])
	if r.presentation != nil {
		r.presentation.SetCell(cursor.x, cursor.y, &cursor.frames[level])
	}
	return true
}
