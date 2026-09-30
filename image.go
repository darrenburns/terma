package terma

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"reflect"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/darrenburns/terma/layout"
)

// ImageResource is an immutable, decoded image. Construct it once and share it
// between widgets and frames. The source's pixels are copied, never retained.
type ImageResource struct{ pixels *image.NRGBA }

// NewImageResource copies a nonempty image into zero-origin, straight-alpha storage.
func NewImageResource(source image.Image) (*ImageResource, error) {
	if source == nil || (reflect.ValueOf(source).Kind() == reflect.Ptr && reflect.ValueOf(source).IsNil()) {
		return nil, errors.New("terma: nil image")
	}
	b := source.Bounds()
	if b.Empty() {
		return nil, errors.New("terma: empty image")
	}
	p := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(p, p.Bounds(), source, b.Min, draw.Src)
	return &ImageResource{pixels: p}, nil
}

// ImageFit controls scaling within an Image's content box.
type ImageFit uint8

const (
	// ImageContain fits the complete image, centered, preserving its aspect ratio.
	ImageContain ImageFit = iota
	// ImageCover fills the content box, cropping the image around its center.
	ImageCover
	// ImageStretch fills the content box without preserving its aspect ratio.
	ImageStretch
)

// Image displays a static decoded image. Its logical representation is always
// coloured half blocks, including in snapshots and screen exports.
type Image struct {
	ID     string
	Source *ImageResource
	Fit    ImageFit
	Style  Style
}

func (i Image) Build(BuildContext) Widget                    { return i }
func (i Image) WidgetID() string                             { return i.ID }
func (i Image) GetStyle() Style                              { return i.Style }
func (i Image) GetContentDimensions() (Dimension, Dimension) { return i.Style.Width, i.Style.Height }
func (i Image) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	cw, ch := 8, 16
	if ctx.renderer != nil {
		cw, ch = ctx.renderer.imageCellSize()
	}
	p, b := toLayoutEdgeInsets(i.Style.Padding), borderToEdgeInsets(i.Style.Border)
	dims := GetWidgetDimensionSet(i)
	minW, maxW, minH, maxH := dimensionSetToMinMax(dims, p, b)
	var node layout.LayoutNode = &layout.BoxNode{
		Padding: p, Border: b, Margin: toLayoutEdgeInsets(i.Style.Margin),
		MinWidth: minW, MaxWidth: maxW, MinHeight: minH, MaxHeight: maxH,
		MeasureFunc: func(c layout.Constraints) (int, int) {
			// Containers resolve percentage dimensions and allocate flex space.
			// Fill bounded allocations, including when this image is the root;
			// retain intrinsic sizing during unbounded scroll measurements.
			// Match layout.isUnbounded: padding can reduce a sentinel slightly.
			if (dims.Width.IsFlex() || dims.Width.IsPercent()) && c.MaxWidth <= 100_000 {
				c.MinWidth = c.MaxWidth
			}
			if (dims.Height.IsFlex() || dims.Height.IsPercent()) && c.MaxHeight <= 100_000 {
				c.MinHeight = c.MaxHeight
			}
			if i.Source == nil || i.Source.pixels == nil {
				return c.Constrain(0, 0)
			}
			size := i.Source.pixels.Bounds().Size()
			w, h := (size.X+cw-1)/cw, (size.Y+ch-1)/ch
			if !c.IsTightWidth() && !c.IsTightHeight() {
				// A bounded Auto image shrinks proportionally; clamping each axis
				// independently would leave a large, empty intrinsic content box.
				naturalW, naturalH := float64(size.X)/float64(cw), float64(size.Y)/float64(ch)
				scale := max(1.0, float64(c.MinWidth)/naturalW, float64(c.MinHeight)/naturalH)
				scale = min(scale, float64(c.MaxWidth)/naturalW, float64(c.MaxHeight)/naturalH)
				w, h = int(math.Ceil(naturalW*scale)), int(math.Ceil(naturalH*scale))
			}
			if c.IsTightWidth() {
				w = c.MinWidth
				h = int(math.Ceil(float64(w*cw*size.Y) / float64(size.X*ch)))
			}
			if c.IsTightHeight() {
				h = c.MinHeight
				if !c.IsTightWidth() {
					w = int(math.Ceil(float64(h*ch*size.X) / float64(size.Y*cw)))
				}
			}
			return c.Constrain(w, h)
		},
	}
	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{child: node, minWidth: dims.MinWidth, maxWidth: dims.MaxWidth, minHeight: dims.MinHeight, maxHeight: dims.MaxHeight, padding: p, border: b}
	}
	return node
}
func (i Image) Render(ctx *RenderContext) {
	ctx.DrawImage(0, 0, ctx.Width, ctx.Height, i.Source, i.Fit)
}

// imageMapping is protocol independent; source crop coordinates are pixels.
type imageMapping struct {
	dest Rect
	crop image.Rectangle
}

func mapImage(source *ImageResource, dest Rect, fit ImageFit, cw, ch int) imageMapping {
	crop := source.pixels.Bounds()
	if dest.IsEmpty() {
		return imageMapping{dest, crop}
	}
	sw, sh := float64(crop.Dx()), float64(crop.Dy())
	dw, dh := float64(dest.Width*cw), float64(dest.Height*ch)
	switch fit {
	case ImageCover:
		if dw/dh > sw/sh {
			h := max(1, int(sw*dh/dw))
			crop.Min.Y = (crop.Dy() - h) / 2
			crop.Max.Y = crop.Min.Y + h
		} else {
			w := max(1, int(sh*dw/dh))
			crop.Min.X = (crop.Dx() - w) / 2
			crop.Max.X = crop.Min.X + w
		}
	case ImageStretch:
	default:
		scale := math.Min(dw/sw, dh/sh)
		w, h := max(1, int(math.Round(sw*scale/float64(cw)))), max(1, int(math.Round(sh*scale/float64(ch))))
		w, h = min(w, dest.Width), min(h, dest.Height)
		dest.X += (dest.Width - w) / 2
		dest.Y += (dest.Height - h) / 2
		dest.Width = w
		dest.Height = h
	}
	return imageMapping{dest, crop}
}

// flattenImageColor blends c over bg. Over no background of the app's own
// (bg.A is 0), the terminal's shows through where the image is fully
// transparent, so those pixels stay transparent; the terminal's colour isn't
// known, so partly transparent ones are blended over black.
func flattenImageColor(c color.NRGBA, bg color.NRGBA) color.NRGBA {
	if bg.A == 0 {
		if c.A == 0 {
			return color.NRGBA{}
		}
		bg = color.NRGBA{A: 255}
	}
	a := uint32(c.A)
	inv := 255 - a
	return color.NRGBA{uint8((uint32(c.R)*a + uint32(bg.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(bg.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(bg.B)*inv + 127) / 255), 255}
}

// imageBackground returns the background an image is drawn over, or no colour
// (transparent) when the app sets none and the terminal's own is behind it.
func imageBackground(c Color) color.NRGBA {
	if !c.IsSet() {
		return color.NRGBA{}
	}
	if !c.IsOpaque() {
		c = c.BlendOver(Black)
	}
	return color.NRGBAModel.Convert(c.toANSI()).(color.NRGBA)
}

// DrawImage draws into a cell rectangle relative to this context. A nil Source
// draws nothing. Later cell painting covers native graphics, even identical writes.
func (ctx *RenderContext) DrawImage(x, y, width, height int, source *ImageResource, fit ImageFit) {
	if source == nil || source.pixels == nil || width <= 0 || height <= 0 {
		return
	}
	cw, ch := 8, 16
	r := ctx.buildContext.renderer
	if r != nil {
		r.imageUsed = true
		cw, ch = r.imageCellSize()
	}
	m := mapImage(source, Rect{ctx.X + x, ctx.Y + y, width, height}, fit, cw, ch)
	visible := m.dest.Intersect(ctx.visible)
	painted := visible.Intersect(ctx.clip)
	slots := ctx.imageSlots()
	slot := *slots
	*slots = slot + 1
	if painted.IsEmpty() {
		return
	}
	var record *imageRecord
	if tracker, ok := ctx.terminal.(*imageBuffer); ok {
		key := imageRecordKey{ctx.imageOwner, slot}
		record = tracker.record(key, source, m, visible)
		if ctx.imageOwner != nil {
			ctx.imageOwner.imageBuffer = tracker
		}
	}
	for ay := painted.Y; ay < painted.Y+painted.Height; ay++ {
		for ax := painted.X; ax < painted.X+painted.Width; ax++ {
			var bg color.NRGBA
			if ctx.inheritedBgAt != nil {
				bg = imageBackground(ctx.inheritedBgAt(ax, ay))
			} else if old := ctx.terminal.CellAt(ax, ay); old != nil && old.Style.Bg != nil {
				bg = color.NRGBAModel.Convert(old.Style.Bg).(color.NRGBA)
			}
			sx := m.crop.Min.X + min(m.crop.Dx()-1, (2*(ax-m.dest.X)+1)*m.crop.Dx()/(2*m.dest.Width))
			top := m.crop.Min.Y + min(m.crop.Dy()-1, (4*(ay-m.dest.Y)+1)*m.crop.Dy()/(4*m.dest.Height))
			bottom := m.crop.Min.Y + min(m.crop.Dy()-1, (4*(ay-m.dest.Y)+3)*m.crop.Dy()/(4*m.dest.Height))
			cell := halfBlockCell(flattenImageColor(source.pixels.NRGBAAt(sx, top), bg), flattenImageColor(source.pixels.NRGBAAt(sx, bottom), bg))
			ctx.terminal.SetCell(ax, ay, &cell)
			if record != nil {
				ctx.terminal.(*imageBuffer).claim(ax, ay, record, cell, bg)
			}
		}
	}
}

// halfBlockCell shows two pixels, one above the other, in a cell. A transparent
// one is left to the terminal's background.
func halfBlockCell(top, bottom color.NRGBA) uv.Cell {
	switch {
	case top.A == 0 && bottom.A == 0:
		return uv.Cell{Content: " ", Width: 1}
	case bottom.A == 0:
		return uv.Cell{Content: "▀", Width: 1, Style: uv.Style{Fg: top}}
	case top.A == 0:
		return uv.Cell{Content: "▄", Width: 1, Style: uv.Style{Fg: bottom}}
	}
	return uv.Cell{Content: "▀", Width: 1, Style: uv.Style{Fg: top, Bg: bottom}}
}

// Subcontexts share a draw sequence so their records cannot alias. Retained
// widgets use their node's counter, reset at the start of each Render call.
func (ctx *RenderContext) imageSlots() *int {
	if ctx.imageSlot == nil {
		ctx.imageSlot = new(int)
	}
	return ctx.imageSlot
}
