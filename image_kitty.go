package terma

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"sort"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

const maxNativeImagePixels = 4 * 1024 * 1024
const maxImageCacheBytes = 64 * 1024 * 1024
const kittyResponseTimeout = time.Second

type imageVariant struct {
	source        *ImageResource
	crop          image.Rectangle
	width, height int
}

func variantFor(rec *imageRecord, cw, ch int) imageVariant {
	return imageVariant{rec.source, rec.mapping.crop, rec.mapping.dest.Width * cw, rec.mapping.dest.Height * ch}
}
func (v imageVariant) valid() bool {
	return v.width > 0 && v.height > 0 && v.width <= maxNativeImagePixels/v.height
}
func (v imageVariant) pixels() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, v.width, v.height))
	for y := 0; y < v.height; y++ {
		sy := v.crop.Min.Y + min(v.crop.Dy()-1, (2*y+1)*v.crop.Dy()/(2*v.height))
		for x := 0; x < v.width; x++ {
			sx := v.crop.Min.X + min(v.crop.Dx()-1, (2*x+1)*v.crop.Dx()/(2*v.width))
			out.SetNRGBA(x, y, v.source.pixels.NRGBAAt(sx, sy))
		}
	}
	return out
}

type kittyPlacementKey struct{ x, y, cols, rows, cw, ch int }
type kittyPlacement struct {
	id                     int
	key                    kittyPlacementKey
	ready, pending, failed bool
	deadline               time.Time
}
type kittyUpload struct {
	id                     int
	key                    imageVariant
	data                   string
	offset, attempts       int
	ready, pending, failed bool
	deadline               time.Time
	placements             map[kittyPlacementKey]*kittyPlacement
	lastVisible, visible   bool
}
type kittyImages struct {
	uploads     map[imageVariant]*kittyUpload
	ids         [256]*kittyUpload
	active      *kittyUpload
	garbage     []string
	uploadsSent int
	bytes       int
}

func newKittyImages() *kittyImages { return &kittyImages{uploads: make(map[imageVariant]*kittyUpload)} }
func kittyCommand(options, payload string) string {
	if payload != "" {
		return "\x1b_G" + options + ";" + payload + "\x1b\\"
	}
	return "\x1b_G" + options + "\x1b\\"
}
func (k *kittyImages) begin(needed map[imageVariant]bool) {
	for v, u := range k.uploads {
		u.visible = needed[v]
	}
}
func (k *kittyImages) evict(u *kittyUpload) {
	k.garbage = append(k.garbage, kittyCommand(fmt.Sprintf("a=d,d=I,i=%d,q=2", u.id), ""))
	delete(k.uploads, u.key)
	k.ids[u.id] = nil
	k.bytes -= len(u.data)
}
func (k *kittyImages) upload(v imageVariant) *kittyUpload {
	if u := k.uploads[v]; u != nil {
		u.visible = true
		return u
	}
	if !v.valid() {
		return nil
	}
	id := 0
	for i := 1; i < 256; i++ {
		if k.ids[i] == nil {
			id = i
			break
		}
	}
	if id == 0 || k.bytes+v.width*v.height*6 > maxImageCacheBytes {
		for i := 1; i < 256; i++ {
			u := k.ids[i]
			if u != nil && !u.visible && !u.lastVisible && !u.pending && u != k.active {
				k.evict(u)
				if id == 0 {
					id = i
				}
				if k.bytes+v.width*v.height*6 <= maxImageCacheBytes {
					break
				}
			}
		}
	}
	if id == 0 || k.bytes+v.width*v.height*6 > maxImageCacheBytes {
		return nil
	}
	var data bytes.Buffer
	if png.Encode(&data, v.pixels()) != nil {
		return nil
	}
	u := &kittyUpload{id: id, key: v, data: base64.StdEncoding.EncodeToString(data.Bytes()), placements: make(map[kittyPlacementKey]*kittyPlacement), visible: true}
	k.uploads[v] = u
	k.ids[id] = u
	k.bytes += len(u.data)
	return u
}
func (u *kittyUpload) placement(key kittyPlacementKey) *kittyPlacement {
	if p := u.placements[key]; p != nil {
		return p
	}
	// At most 255 placement IDs per upload, never alias one still in use.
	if len(u.placements) >= 255 {
		return nil
	}
	p := &kittyPlacement{id: len(u.placements) + 1, key: key}
	u.placements[key] = p
	return p
}
func (k *kittyImages) expire(now time.Time) {
	for _, u := range k.uploads {
		if u.pending && !now.Before(u.deadline) {
			k.fail(u)
		}
		for _, p := range u.placements {
			if p.pending && !now.Before(p.deadline) {
				p.pending = false
				p.failed = true
			}
		}
	}
}
func (k *kittyImages) fail(u *kittyUpload) {
	u.pending = false
	u.ready = false
	u.offset = 0
	for _, p := range u.placements {
		p.ready = false
		p.pending = false
	}
	if u.attempts >= 2 {
		u.failed = true
	}
}
func (k *kittyImages) handle(e uv.KittyGraphicsEvent) {
	id := e.Options.ID
	if id <= 0 || id >= len(k.ids) {
		return
	}
	u := k.ids[id]
	if u == nil {
		return
	}
	if e.Options.PlacementID != 0 {
		for _, p := range u.placements {
			if p.id == e.Options.PlacementID && p.pending {
				p.pending = false
				p.ready = string(e.Payload) == "OK"
				if !p.ready {
					if bytes.HasPrefix(e.Payload, []byte("ENOENT")) {
						k.fail(u)
					} else {
						p.failed = true
					}
				}
				return
			}
		}
	} else if u.pending {
		if string(e.Payload) == "OK" {
			u.pending = false
			u.ready = true
		} else {
			k.fail(u)
		}
	}
}

// batch serializes all graphics commands. Multipart uploads are exclusive, but
// each batch is bounded so the app event loop can process input between them.
func (k *kittyImages) batch(now time.Time) string {
	k.expire(now)
	var out bytes.Buffer
	budget := 4
	for budget > 0 {
		if k.active == nil {
			for len(k.garbage) > 0 && budget > 0 {
				out.WriteString(k.garbage[0])
				k.garbage = k.garbage[1:]
				budget--
			}
			if budget == 0 {
				break
			}
			for id := 1; id < 256; id++ {
				u := k.ids[id]
				if u != nil && u.visible && !u.ready && !u.pending && !u.failed {
					k.active = u
					u.attempts++
					k.uploadsSent++
					break
				}
			}
		}
		if k.active == nil {
			break
		}
		u := k.active
		end := min(len(u.data), u.offset+4096)
		more := 0
		if end < len(u.data) {
			more = 1
		}
		options := fmt.Sprintf("m=%d", more)
		if u.offset == 0 {
			options = fmt.Sprintf("a=t,t=d,f=100,i=%d,q=0,m=%d", u.id, more)
		}
		out.WriteString(kittyCommand(options, u.data[u.offset:end]))
		u.offset = end
		budget--
		if more == 0 {
			u.pending = true
			u.deadline = now.Add(kittyResponseTimeout)
			k.active = nil
		}
	}
	if k.active != nil {
		return out.String()
	}
	for id := 1; id < 256 && budget > 0; id++ {
		u := k.ids[id]
		if u == nil || !u.visible || !u.ready {
			continue
		}
		// Stable ordering makes emitted bytes and retry behavior reproducible.
		list := make([]*kittyPlacement, 0, len(u.placements))
		for _, p := range u.placements {
			list = append(list, p)
		}
		sort.Slice(list, func(i, j int) bool { return list[i].id < list[j].id })
		for _, p := range list {
			if budget == 0 {
				break
			}
			if p.ready || p.pending || p.failed {
				continue
			}
			v := p.key
			out.WriteString(kittyCommand(fmt.Sprintf("a=p,U=1,i=%d,p=%d,x=%d,y=%d,w=%d,h=%d,c=%d,r=%d,q=0", u.id, p.id, v.x*v.cw, v.y*v.ch, v.cols*v.cw, v.rows*v.ch, v.cols, v.rows), ""))
			p.pending = true
			p.deadline = now.Add(kittyResponseTimeout)
			budget--
		}
	}
	return out.String()
}
func (k *kittyImages) pending() bool {
	if k.active != nil || len(k.garbage) > 0 {
		return true
	}
	for _, u := range k.uploads {
		if u.failed {
			continue
		}
		if u.visible && !u.ready {
			return true
		}
		for _, p := range u.placements {
			if u.visible && !p.ready && !p.failed {
				return true
			}
		}
	}
	return false
}
func (k *kittyImages) finish() {
	for _, u := range k.uploads {
		u.lastVisible = u.visible
	}
}
func (k *kittyImages) cleanup() string {
	var out bytes.Buffer
	for id := 1; id < 256; id++ {
		if k.ids[id] != nil {
			out.WriteString(kittyCommand(fmt.Sprintf("a=d,d=I,i=%d,q=2", id), ""))
		}
	}
	return out.String()
}
func (k *kittyImages) paint(dst CellBuffer, b *imageBuffer, cw, ch int) {
	needed := make(map[imageVariant]bool)
	for _, rec := range b.records {
		needed[variantFor(rec, cw, ch)] = true
	}
	k.begin(needed)
	// Allocation order follows screen order, never Go map iteration.
	seen := make(map[*imageRecord]bool)
	for _, cell := range b.cells {
		rec := cell.owner
		if rec == nil || seen[rec] {
			continue
		}
		seen[rec] = true
		u := k.upload(variantFor(rec, cw, ch))
		if u == nil || u.failed {
			continue
		}
		d := rec.mapping.dest
		limit := len(imageDiacritics)
		for ty := 0; ty < d.Height; ty += limit {
			for tx := 0; tx < d.Width; tx += limit {
				tile := Rect{d.X + tx, d.Y + ty, min(limit, d.Width-tx), min(limit, d.Height-ty)}
				if tile.Intersect(rec.visible).IsEmpty() {
					continue
				}
				p := u.placement(kittyPlacementKey{tx, ty, tile.Width, tile.Height, cw, ch})
				if p == nil || !u.ready || !p.ready {
					continue
				}
				area := tile.Intersect(rec.visible).Intersect(Rect{Width: b.buffer.Width(), Height: b.buffer.Height()})
				for y := area.Y; y < area.Y+area.Height; y++ {
					for x := area.X; x < area.X+area.Width; x++ {
						if !b.owns(x, y, rec) {
							continue
						}
						bg := b.cells[b.index(x, y)].background
						dst.SetCell(x, y, &uv.Cell{Content: string([]rune{0x10eeee, imageDiacritics[y-tile.Y], imageDiacritics[x-tile.X]}), Width: 1, Style: uv.Style{Fg: ansi.IndexedColor(u.id), UnderlineColor: ansi.IndexedColor(p.id), Bg: bg}})
					}
				}
			}
		}
	}
}
