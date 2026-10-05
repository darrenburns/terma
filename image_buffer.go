package terma

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

type imageRecordKey struct {
	node *widgetNode
	slot int
}
type imageRecord struct {
	source  *ImageResource
	mapping imageMapping
	visible Rect
}
type imageCell struct {
	owner      *imageRecord
	preview    uv.Cell
	background color.NRGBA
}

// imageBuffer owns logical cells only. Every mutation revokes ownership even
// when it writes identical pixels; final-cell checks also catch pointer edits.
// Do not embed Buffer: its mutators must never bypass ownership tracking.
type imageBuffer struct {
	method  uv.WidthMethod
	buffer  *uv.Buffer
	cells   []imageCell
	records map[imageRecordKey]*imageRecord
}

var _ uv.Screen = (*imageBuffer)(nil)

func newImageBuffer(w, h int) *imageBuffer {
	b := &imageBuffer{buffer: uv.NewBuffer(w, h), method: ansi.WcWidth, records: make(map[imageRecordKey]*imageRecord)}
	b.cells = make([]imageCell, w*h)
	return b
}
func (b *imageBuffer) WidthMethod() uv.WidthMethod { return b.method }
func (b *imageBuffer) Bounds() uv.Rectangle        { return b.buffer.Bounds() }
func (b *imageBuffer) CellAt(x, y int) *uv.Cell    { return b.buffer.CellAt(x, y) }
func (b *imageBuffer) index(x, y int) int {
	if x < 0 || y < 0 || x >= b.buffer.Width() || y >= b.buffer.Height() {
		return -1
	}
	return y*b.buffer.Width() + x
}
func (b *imageBuffer) SetCell(x, y int, c *uv.Cell) {
	if b.index(x, y) < 0 {
		return
	}
	width := 1
	if c != nil {
		width = max(1, c.Width)
	}
	// A write may erase an old wide glyph, including one whose continuation
	// cell was targeted. Include all intersecting old glyphs in the invalidation.
	lo, hi := x, min(b.buffer.Width(), x+width)
	for start := max(0, x-4); start < hi; start++ {
		if old := b.CellAt(start, y); old != nil && start+max(1, old.Width) > x {
			lo = min(lo, start)
			hi = max(hi, min(b.buffer.Width(), start+max(1, old.Width)))
		}
	}
	for col := lo; col < hi; col++ {
		b.cells[b.index(col, y)] = imageCell{}
	}
	b.buffer.SetCell(x, y, c)
}
func (b *imageBuffer) Clear()                   { b.Fill(nil) }
func (b *imageBuffer) ClearArea(a uv.Rectangle) { b.FillArea(nil, a) }
func (b *imageBuffer) Fill(c *uv.Cell)          { b.FillArea(c, b.Bounds()) }
func (b *imageBuffer) FillArea(c *uv.Cell, a uv.Rectangle) {
	a = a.Intersect(b.Bounds())
	step := 1
	if c != nil {
		step = max(1, c.Width)
	}
	for y := a.Min.Y; y < a.Max.Y; y++ {
		for x := a.Min.X; x < a.Max.X; x += step {
			b.SetCell(x, y, c)
		}
	}
}
func (b *imageBuffer) Resize(w, h int) {
	b.buffer.Resize(max(0, w), max(0, h))
	b.cells = make([]imageCell, max(0, w)*max(0, h))
	clear(b.records)
}
func (b *imageBuffer) record(k imageRecordKey, s *ImageResource, m imageMapping, v Rect) *imageRecord {
	rec := b.records[k]
	if rec == nil {
		rec = &imageRecord{}
		b.records[k] = rec
	}
	rec.source, rec.mapping, rec.visible = s, m, v
	return rec
}
func (b *imageBuffer) claim(x, y int, r *imageRecord, c uv.Cell, bg color.NRGBA) {
	if idx := b.index(x, y); idx >= 0 {
		b.cells[idx] = imageCell{r, c, bg}
	}
}
func (b *imageBuffer) owns(x, y int, r *imageRecord) bool {
	idx := b.index(x, y)
	if idx < 0 {
		return false
	}
	c := b.cells[idx]
	return c.owner == r && r.visible.Contains(x, y) && c.preview.Equal(b.CellAt(x, y))
}
func (b *imageBuffer) prune() {
	active := make(map[*imageRecord]bool)
	for _, c := range b.cells {
		if c.owner != nil {
			active[c.owner] = true
		}
	}
	for k, r := range b.records {
		if !active[r] {
			delete(b.records, k)
		}
	}
}
func (b *imageBuffer) removeNode(n *widgetNode) {
	for k, r := range b.records {
		if k.node == n {
			for i, c := range b.cells {
				if c.owner == r {
					b.cells[i] = imageCell{}
				}
			}
			delete(b.records, k)
		}
	}
}
func copyCells(dst CellBuffer, src CellBuffer, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; {
			c := src.CellAt(x, y)
			dst.SetCell(x, y, c)
			step := 1
			if c != nil {
				step = max(1, c.Width)
			}
			x += step
		}
	}
}
func (r *Renderer) imageCellSize() (int, int) {
	if r.imageCellWidth > 0 && r.imageCellHeight > 0 {
		return r.imageCellWidth, r.imageCellHeight
	}
	return 8, 16
}

// finishImages switches only on first image use. A second, complete frame is
// essential: earlier siblings were painted into the original terminal.
func (r *Renderer) finishImages(root Widget) (switched bool) {
	if r.imageUsed && r.images == nil {
		r.images = newImageBuffer(r.width, r.height)
		if screen, ok := r.terminal.(uv.Screen); ok {
			r.images.method = screen.WidthMethod()
		}
		r.presentation = r.terminal
		r.terminal = r.images
		r.renderFull(root)
		switched = true
	}
	if r.images != nil {
		r.images.prune()
		copyCells(r.presentation, r.images, r.width, r.height)
	}
	return switched
}
