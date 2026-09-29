package terma

import (
	"math"

	uv "github.com/charmbracelet/ultraviolet"
)

// hoverItem is implemented by the widget that shows one item of a collection:
// a List or Tree row, a Table cell and a tab. While the pointer rests on it,
// or on anything inside it, the item shows a hover highlight.
//
// The highlight is a hoverTint, which describes how it is applied.
type hoverItem interface {
	// hoverKey identifies the item being shown. It must be comparable. Moving
	// between widgets with the same key (the cells of a row, in a Table that
	// highlights rows) doesn't change the hover.
	hoverKey() any
	// setHovered records that the pointer entered or left the item.
	setHovered(hovered bool)
}

// hoverItemKey is a hoverItem's key: the state of the collection showing the
// item, and the item within it.
type hoverItemKey struct {
	owner any
	item  any
}

// hoveredKey is the item under the pointer, if any.
type hoveredKey[K comparable] struct {
	key K
	ok  bool
}

// itemHover holds which item of a collection the pointer rests on. Items read
// it with is while painting, so moving the pointer between two items repaints
// just those two. A zero itemHover (from a state not made by its constructor)
// never reports a hovered item.
type itemHover[K comparable] struct {
	signal Signal[hoveredKey[K]]
}

func newItemHover[K comparable]() itemHover[K] {
	return itemHover[K]{signal: NewSignal(hoveredKey[K]{})}
}

// is reports whether the item with this key is hovered, subscribing only to
// changes in that answer.
func (h itemHover[K]) is(key K) bool {
	if !h.signal.IsValid() {
		return false
	}
	return Select(h.signal, func(hovered hoveredKey[K]) bool {
		return hovered.ok && hovered.key == key
	})
}

// set records that the pointer entered or left the item with this key.
// Leaving an item that is no longer the hovered one changes nothing, so the
// pointer can enter the next item before it leaves the last.
func (h itemHover[K]) set(key K, hovered bool) {
	if !h.signal.IsValid() {
		return
	}
	if hovered {
		h.signal.Set(hoveredKey[K]{key: key, ok: true})
		return
	}
	if current := h.signal.Peek(); current.ok && current.key == key {
		h.signal.Set(hoveredKey[K]{})
	}
}

// itemHoverTracker remembers the hoverItem under the pointer and moves the
// hover when a different item comes under it.
type itemHoverTracker struct {
	item hoverItem
	key  any
}

// update makes item (nil for none) the hovered item, reporting whether the
// hover moved.
func (t *itemHoverTracker) update(item hoverItem) bool {
	var key any
	if item != nil {
		key = item.hoverKey()
	}
	if key == t.key {
		t.item = item
		return false
	}
	// Enter before leaving: within one collection the entered item replaces
	// the left one directly, so an item hovered both before and after (such
	// as a Table row, when moving between its cells) isn't repainted.
	if item != nil {
		item.setHovered(true)
	}
	if t.item != nil {
		t.item.setHovered(false)
	}
	t.item, t.key = item, key
	return true
}

// underlayPainter is implemented by widgets that paint beneath their own
// background, such as a collection item's hover highlight. The renderer calls
// it with the widget's border box before drawing the background, border and
// content, and its signal reads are paint dependencies.
type underlayPainter interface {
	hasUnderlay() bool
	paintUnderlay(ctx *RenderContext)
}

// hoverTint is the hover highlight: the theme's translucent Hover color laid
// over whatever the hovered widget sits on. Everything that shows hover uses
// it, in one of two ways:
//
//   - As an underlay (paintUnderlay), beneath content with a transparent
//     background: List, Table and Tree items. Opaque parts of the content,
//     such as the cursor highlight or a badge, cover it unchanged.
//   - Blended into the widget's own opaque background (background): tabs and
//     buttons, whose background would hide an underlay.
//
// hovered is read when the tint is applied, so the widget is repainted (as an
// underlay) or rebuilt (as a background) when it changes.
type hoverTint struct {
	hovered func() bool
	color   Color
}

func newHoverTint(theme ThemeData, hovered func() bool) hoverTint {
	return hoverTint{hovered: hovered, color: theme.Hover}
}

// active reports whether the tint shows.
func (h hoverTint) active() bool {
	return h.hovered != nil && h.hovered()
}

// underlay returns the color to lay beneath the content: the tint when
// hovered, else none. It suits PresentedText.Underlay.
func (h hoverTint) underlay(*RenderContext) Color {
	if h.active() {
		return h.color
	}
	return Color{}
}

func (h hoverTint) hasUnderlay() bool { return h.hovered != nil }

func (h hoverTint) paintUnderlay(ctx *RenderContext) {
	ctx.tintRect(0, 0, ctx.Width, ctx.Height, h.underlay(ctx))
}

// background returns bg with the tint blended in when hovered, else bg.
func (h hoverTint) background(bg ColorProvider) ColorProvider {
	if bg == nil || !bg.IsSet() || !h.active() {
		return bg
	}
	if color, ok := bg.(Color); ok {
		return tintOver(h.color, color)
	}
	return tintedColors{base: bg, tint: h.color}
}

// tintOver composites a translucent tint over a color. The result is opaque
// over an opaque color, and translucent over a translucent one, so it still
// blends with whatever is beneath.
func tintOver(tint, base Color) Color {
	if base.IsOpaque() {
		return tint.BlendOver(base)
	}
	ta, ba := tint.Alpha(), base.Alpha()
	alpha := ta + ba*(1-ta)
	if alpha <= 0 {
		return base
	}
	channel := func(t, b uint8) uint8 {
		return uint8(math.Round((float64(t)*ta + float64(b)*ba*(1-ta)) / alpha))
	}
	return RGBA(channel(tint.r, base.r), channel(tint.g, base.g), channel(tint.b, base.b), alpha)
}

// tintedColors is a ColorProvider (such as a gradient) with a tint laid over
// every color it gives.
type tintedColors struct {
	base ColorProvider
	tint Color
}

func (t tintedColors) ColorAt(width, height, x, y int) Color {
	return tintOver(t.tint, t.base.ColorAt(width, height, x, y))
}

func (t tintedColors) IsSet() bool { return t.base.IsSet() }

// hoverUnderlay holds one item of a collection rendered by app code, and
// paints the hover highlight beneath it. Like passThrough it leaves layout to
// the child.
type hoverUnderlay struct {
	passThrough
	hoverTint
}

func (h hoverUnderlay) Build(BuildContext) Widget { return h }

// withHoverUnderlay gives a PresentedText the hover highlight as its underlay.
func withHoverUnderlay(text PresentedText, tint hoverTint) PresentedText {
	text.Underlay = tint.underlay
	return text
}

// tintRect lays a translucent color over the background already drawn in the
// rectangle, clearing any content so what is painted next sits on top of it.
func (ctx *RenderContext) tintRect(x, y, width, height int, tint Color) {
	if !tint.IsSet() {
		return
	}
	for row := 0; row < height; row++ {
		absY := ctx.Y + y + row
		if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
			continue
		}
		for col := 0; col < width; col++ {
			absX := ctx.X + x + col
			if absX < ctx.clip.X || absX >= ctx.clip.X+ctx.clip.Width {
				continue
			}
			var bg Color
			if existing := ctx.terminal.CellAt(absX, absY); existing != nil {
				bg = FromANSI(existing.Style.Bg)
			}
			if !bg.IsSet() && ctx.inheritedBgAt != nil {
				bg = ctx.inheritedBgAt(absX, absY)
			}
			if !bg.IsSet() {
				bg = ctx.buildContext.Theme().Background
			}
			ctx.terminal.SetCell(absX, absY, &uv.Cell{
				Content: " ",
				Width:   1,
				Style:   uv.Style{Bg: tint.BlendOver(bg).toANSI()},
			})
		}
	}
}
