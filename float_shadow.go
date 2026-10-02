package terma

import (
	"math"

	uv "github.com/charmbracelet/ultraviolet"
)

// FloatShadow tints the cells behind a floating widget without changing its size
// or pointer area. Dimensions are terminal cells. A zero Offset makes a glow.
type FloatShadow struct {
	Color      Color
	Offset     Offset
	BlurRadius int
	Spread     int
}

func (s FloatShadow) bounds(box Rect) Rect {
	if s.Color.Alpha() <= 0 || box.IsEmpty() {
		return box
	}
	radius := max(0, s.BlurRadius) + max(0, s.Spread)
	return box.Union(Rect{X: box.X + s.Offset.X - radius, Y: box.Y + s.Offset.Y - radius,
		Width: box.Width + 2*radius, Height: box.Height + 2*radius})
}

func floatShadow(config FloatConfig) FloatShadow {
	if config.Shadow == nil {
		return FloatShadow{}
	}
	return *config.Shadow
}

// Overlays can obscure one half of a wide glyph. Include its other half when
// restoring the underlay, then keep the resulting clip on whole glyphs too.
func (r *Renderer) overlayDamage(rect Rect) Rect {
	if rect.IsEmpty() {
		return rect
	}
	left, right := max(0, rect.X-1), min(r.width, rect.X+rect.Width+1)
	for {
		beforeLeft, beforeRight := left, right
		for y := max(0, rect.Y); y < min(r.height, rect.Y+rect.Height); y++ {
			for left > 0 {
				cell := r.terminal.CellAt(left, y)
				if cell == nil || cell.Width != 0 || cell.Content != "" {
					break
				}
				left--
			}
			if right > 0 {
				if cell := r.terminal.CellAt(right-1, y); cell != nil {
					right = min(r.width, max(right, right-1+cell.Width))
				}
			}
		}
		if left == beforeLeft && right == beforeRight {
			break
		}
	}
	return Rect{X: left, Y: rect.Y, Width: right - left, Height: rect.Height}
}

func paintFloatShadow(ctx *RenderContext, box Rect, shadow FloatShadow) {
	if shadow.Color.Alpha() <= 0 || box.IsEmpty() {
		return
	}
	spread, blur := max(0, shadow.Spread), max(0, shadow.BlurRadius)
	core := Rect{X: box.X + shadow.Offset.X - spread, Y: box.Y + shadow.Offset.Y - spread,
		Width: box.Width + 2*spread, Height: box.Height + 2*spread}
	area := shadow.bounds(box).Intersect(ctx.clip)
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			// A wide glyph is recolored through its leading cell. Writing its
			// continuation separately would erase the glyph in terminal buffers.
			existing := ctx.terminal.CellAt(x, y)
			if existing != nil && existing.Width == 0 && existing.Content == "" {
				continue
			}
			dx := max(core.X-x, 0, x-(core.X+core.Width-1))
			dy := max(core.Y-y, 0, y-(core.Y+core.Height-1))
			distance := math.Hypot(float64(dx), float64(dy))
			if distance > float64(blur) {
				continue
			}
			color := shadow.Color
			color.a *= 1 - distance/float64(blur+1)
			cell := uv.Cell{Content: " ", Width: 1}
			if existing != nil {
				cell = *existing
			}
			cell.Style.Fg = color.BlendOver(FromANSI(cell.Style.Fg)).toANSI()
			cell.Style.Bg = color.BlendOver(FromANSI(cell.Style.Bg)).toANSI()
			ctx.terminal.SetCell(x, y, &cell)
		}
	}
}
