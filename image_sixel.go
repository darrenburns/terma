package terma

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/darrenburns/terma/internal/imagequant"
)

type sixelKey struct {
	variant    imageVariant
	background color.NRGBA
}
type sixelPrepared struct {
	pixels       *image.Paletted
	used         uint64
	crops        map[image.Rectangle]string
	payloadBytes int
}
type sixelImages struct {
	prepared            map[sixelKey]*sixelPrepared
	clock               uint64
	bytes, preparations int
}

func newSixelImages() *sixelImages { return &sixelImages{prepared: make(map[sixelKey]*sixelPrepared)} }
func (s *sixelImages) prepare(key sixelKey) *sixelPrepared {
	s.clock++
	if p := s.prepared[key]; p != nil {
		p.used = s.clock
		return p
	}
	if !key.variant.valid() {
		return nil
	}
	size := key.variant.width * key.variant.height
	for len(s.prepared) >= 32 || s.bytes+size > maxImageCacheBytes {
		var oldest sixelKey
		var age uint64 = ^uint64(0)
		for k, p := range s.prepared {
			if p.used < age {
				oldest, age = k, p.used
			}
		}
		if age == ^uint64(0) {
			return nil
		}
		p := s.prepared[oldest]
		s.bytes -= len(p.pixels.Pix) + p.payloadBytes
		delete(s.prepared, oldest)
	}
	pixels := key.variant.pixels()
	for y := 0; y < pixels.Rect.Dy(); y++ {
		for x := 0; x < pixels.Rect.Dx(); x++ {
			pixels.SetNRGBA(x, y, flattenImageColor(pixels.NRGBAAt(x, y), key.background))
		}
	}
	p := &sixelPrepared{pixels: imagequant.Prepare(pixels), used: s.clock, crops: make(map[image.Rectangle]string)}
	s.prepared[key] = p
	s.bytes += len(p.pixels.Pix)
	s.preparations++
	return p
}

// sixelCrop always has origin zero: the stock encoder samples from (0,0).
// covered is optional and identifies pixels to replace with transparent index 0.
func sixelCrop(src *image.Paletted, crop image.Rectangle, covered func(x, y int) bool) *image.Paletted {
	crop = crop.Intersect(src.Bounds())
	out := image.NewPaletted(image.Rect(0, 0, crop.Dx(), crop.Dy()), src.Palette)
	for y := 0; y < crop.Dy(); y++ {
		for x := 0; x < crop.Dx(); x++ {
			if covered == nil || !covered(x+crop.Min.X, y+crop.Min.Y) {
				out.SetColorIndex(x, y, src.ColorIndexAt(x+crop.Min.X, y+crop.Min.Y))
			}
		}
	}
	return out
}
func encodeSixel(img *image.Paletted) (string, error) {
	var out bytes.Buffer
	if err := (&sixel.Encoder{}).Encode(&out, img); err != nil {
		return "", err
	}
	// The encoder advances after the final band. At the bottom of the screen
	// that can scroll the whole terminal; the raster height already ends it.
	return "\x1bP0;1q" + strings.TrimSuffix(out.String(), "-") + "\x1b\\", nil
}
func (s *sixelImages) payload(p *sixelPrepared, crop image.Rectangle) string {
	if payload, ok := p.crops[crop]; ok {
		return payload
	}
	payload, err := encodeSixel(sixelCrop(p.pixels, crop, nil))
	if err != nil {
		return ""
	}
	if len(p.crops) >= 64 || p.payloadBytes+len(payload) > 4*1024*1024 {
		s.bytes -= p.payloadBytes
		p.payloadBytes = 0
		clear(p.crops)
	}
	if s.bytes+len(payload) <= maxImageCacheBytes {
		p.crops[crop] = payload
		p.payloadBytes += len(payload)
		s.bytes += len(payload)
	}
	return payload
}

type sixelRegion struct {
	rec        *imageRecord
	background color.NRGBA
	rect       Rect
}

// Disjoint cell runs prevent transparent holes in a later Sixel from erasing
// an earlier image in terminals that store graphics as cell tiles (xterm.js).
func sixelRegions(b *imageBuffer, presentation CellBuffer, debugRows int) []sixelRegion {
	var result []sixelRegion
	previous := make(map[sixelRegion]int)
	for y := debugRows; y < b.buffer.Height(); y++ {
		current := make(map[sixelRegion]int)
		for x := 0; x < b.buffer.Width(); {
			c := b.cells[b.index(x, y)]
			eligible := func(col int) bool {
				idx := b.index(col, y)
				return idx >= 0 && b.cells[idx].owner == c.owner && b.cells[idx].background == c.background && b.owns(col, y, c.owner) && b.cells[idx].preview.Equal(presentation.CellAt(col, y))
			}
			if c.owner == nil || !eligible(x) {
				x++
				continue
			}
			start := x
			for x < b.buffer.Width() && eligible(x) {
				x++
			}
			key := sixelRegion{c.owner, c.background, Rect{X: start, Width: x - start}}
			idx, ok := previous[key]
			if ok {
				result[idx].rect.Height++
			} else {
				idx = len(result)
				region := key
				region.rect.Y = y
				region.rect.Height = 1
				result = append(result, region)
			}
			current[key] = idx
		}
		previous = current
	}
	return result
}
func (s *sixelImages) output(b *imageBuffer, presentation CellBuffer, cw, ch, debugRows int) string {
	var out bytes.Buffer
	for _, region := range sixelRegions(b, presentation, debugRows) {
		rec := region.rec
		p := s.prepare(sixelKey{variantFor(rec, cw, ch), region.background})
		if p == nil {
			continue
		}
		d, a := rec.mapping.dest, region.rect
		crop := image.Rect((a.X-d.X)*cw, (a.Y-d.Y)*ch, (a.X+a.Width-d.X)*cw, (a.Y+a.Height-d.Y)*ch)
		payload := s.payload(p, crop)
		if payload == "" {
			continue
		}
		fmt.Fprintf(&out, "\x1b[%d;%dH", a.Y+1, a.X+1)
		out.WriteString(payload)
	}
	return out.String()
}

// imageTerminal allows output tests to observe the exact Display/Flush boundary.
type imageTerminal interface {
	CellBuffer
	WriteString(string) (int, error)
	Flush() error
	Display() error
	Erase()
}

var _ imageTerminal = (*uv.Terminal)(nil)
