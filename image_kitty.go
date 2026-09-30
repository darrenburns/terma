package terma

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"sort"
	"time"
	"weak"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

const maxNativeImagePixels = 4 * 1024 * 1024
const maxImageCacheBytes = 64 * 1024 * 1024

// kittyResponseTimeout is how long the terminal has to acknowledge an upload
// or placement, from the last byte written. Generous: over a slow link the
// terminal may still be reading a large upload.
const kittyResponseTimeout = 5 * time.Second

// kittyBatchBytes bounds the graphics bytes written in one frame, so the
// event loop gets back to input between them. Chunks are at most 4096 bytes,
// as the protocol recommends.
const kittyBatchBytes = 256 * 1024
const kittyChunkBytes = 4096

// maxHiddenKittyImages bounds the images kept uploaded while none is on
// screen, ready to be shown again without another upload.
const maxHiddenKittyImages = 16

// imageVariant is an image as drawn at one size. It refers to its source
// weakly, so caches keyed by it don't keep a dropped image in memory.
type imageVariant struct {
	source        weak.Pointer[ImageResource]
	crop          image.Rectangle
	width, height int
}

func newImageVariant(source *ImageResource, crop image.Rectangle, width, height int) imageVariant {
	return imageVariant{weak.Make(source), crop, width, height}
}

func variantFor(rec *imageRecord, cw, ch int) imageVariant {
	return newImageVariant(rec.source, rec.mapping.crop, rec.mapping.dest.Width*cw, rec.mapping.dest.Height*ch)
}
func (v imageVariant) valid() bool {
	return v.width > 0 && v.height > 0 && v.width <= maxNativeImagePixels/v.height
}

// pixels scales the variant's source, or returns nil if it has been dropped.
func (v imageVariant) pixels() *image.NRGBA {
	source := v.source.Value()
	if source == nil {
		return nil
	}
	out := image.NewNRGBA(image.Rect(0, 0, v.width, v.height))
	for y := 0; y < v.height; y++ {
		sy := v.crop.Min.Y + min(v.crop.Dy()-1, (2*y+1)*v.crop.Dy()/(2*v.height))
		for x := 0; x < v.width; x++ {
			sx := v.crop.Min.X + min(v.crop.Dx()-1, (2*x+1)*v.crop.Dx()/(2*v.width))
			out.SetNRGBA(x, y, source.pixels.NRGBAAt(sx, sy))
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
	id  int
	key imageVariant
	// data is the encoded image while it is being sent; it is let go once the
	// terminal has the image, and encoded again if it has to be resent.
	data                   string
	encoding               bool
	offset, attempts       int
	ready, pending, failed bool
	deadline               time.Time
	placements             map[kittyPlacementKey]*kittyPlacement
	lastVisible, visible   bool
	lastSeen               uint64 // The frame it was last on screen.
}
type kittyImages struct {
	uploads     map[imageVariant]*kittyUpload
	ids         [256]*kittyUpload
	active      *kittyUpload
	garbage     []string
	uploadsSent int
	bytes       int
	frame       uint64
	worker      imageWorker
	// lastID is the id most recently given out. IDs are given out in turn, so
	// one is reused only long after it was deleted: a late reply about the
	// image it named can't be taken for one about its successor.
	lastID int
}

func newKittyImages() *kittyImages { return &kittyImages{uploads: make(map[imageVariant]*kittyUpload)} }
func kittyCommand(options, payload string) string {
	if payload != "" {
		return "\x1b_G" + options + ";" + payload + "\x1b\\"
	}
	return "\x1b_G" + options + "\x1b\\"
}

// freeID returns the next unused image id after the last one given out, or 0.
func (k *kittyImages) freeID() int {
	for n := 1; n < 256; n++ {
		id := (k.lastID+n-1)%255 + 1
		if k.ids[id] == nil {
			return id
		}
	}
	return 0
}
func (k *kittyImages) begin(needed map[imageVariant]bool) {
	k.frame++
	for v, u := range k.uploads {
		u.visible = needed[v]
		if u.visible {
			u.lastSeen = k.frame
		} else if v.source.Value() == nil && !u.pending && u != k.active {
			// Its source is gone, so it can never be shown again.
			k.evict(u)
		}
	}
}
func (k *kittyImages) evict(u *kittyUpload) {
	k.garbage = append(k.garbage, kittyCommand(fmt.Sprintf("a=d,d=I,i=%d,q=2", u.id), ""))
	delete(k.uploads, u.key)
	k.ids[u.id] = nil
	k.bytes -= len(u.data)
}

// evictable reports whether an upload can be deleted without disturbing the
// screen or a transfer.
func (k *kittyImages) evictable(u *kittyUpload) bool {
	return !u.visible && !u.lastVisible && !u.pending && u != k.active
}
func (k *kittyImages) upload(v imageVariant) *kittyUpload {
	if u := k.uploads[v]; u != nil {
		u.visible = true
		return u
	}
	if !v.valid() || v.source.Value() == nil {
		return nil
	}
	id := k.freeID()
	if id == 0 || k.bytes+v.width*v.height*6 > maxImageCacheBytes {
		for i := 1; i < 256; i++ {
			u := k.ids[i]
			if u != nil && k.evictable(u) {
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
	u := &kittyUpload{id: id, key: v, placements: make(map[kittyPlacementKey]*kittyPlacement), visible: true, lastSeen: k.frame}
	k.lastID = id
	k.uploads[v] = u
	k.ids[id] = u
	k.encode(u)
	return u
}

// encode encodes the upload's image as PNG, off the event loop in an app.
func (k *kittyImages) encode(u *kittyUpload) {
	u.encoding = true
	v := u.key
	runImageWork(k.worker, func() string {
		pixels := v.pixels()
		if pixels == nil {
			return ""
		}
		var data bytes.Buffer
		if png.Encode(&data, pixels) != nil {
			return ""
		}
		return base64.StdEncoding.EncodeToString(data.Bytes())
	}, func(data string) {
		u.encoding = false
		if k.ids[u.id] != u {
			return // Deleted meanwhile.
		}
		if data == "" {
			u.failed = true
			return
		}
		u.data = data
		k.bytes += len(data)
	})
}

// release lets go of an upload's encoded data once the terminal has it.
func (k *kittyImages) release(u *kittyUpload) {
	k.bytes -= len(u.data)
	u.data = ""
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
		k.release(u)
		return
	}
	if u.data == "" && !u.encoding {
		// Let go of once sent; the terminal has lost it since.
		k.encode(u)
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
			k.release(u)
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
	budget := kittyBatchBytes
	for budget > 0 {
		if k.active == nil {
			for len(k.garbage) > 0 && budget > 0 {
				out.WriteString(k.garbage[0])
				budget -= len(k.garbage[0])
				k.garbage = k.garbage[1:]
			}
			if budget <= 0 {
				break
			}
			for id := 1; id < 256; id++ {
				u := k.ids[id]
				if u != nil && u.visible && !u.ready && !u.pending && !u.failed && u.data != "" {
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
		end := min(len(u.data), u.offset+kittyChunkBytes)
		more := 0
		if end < len(u.data) {
			more = 1
		}
		options := fmt.Sprintf("m=%d", more)
		if u.offset == 0 {
			options = fmt.Sprintf("a=t,t=d,f=100,i=%d,q=0,m=%d", u.id, more)
		}
		out.WriteString(kittyCommand(options, u.data[u.offset:end]))
		budget -= end - u.offset
		u.offset = end
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
			if budget <= 0 {
				break
			}
			if p.ready || p.pending || p.failed {
				continue
			}
			v := p.key
			cmd := kittyCommand(fmt.Sprintf("a=p,U=1,i=%d,p=%d,x=%d,y=%d,w=%d,h=%d,c=%d,r=%d,q=0", u.id, p.id, v.x*v.cw, v.y*v.ch, v.cols*v.cw, v.rows*v.ch, v.cols, v.rows), "")
			out.WriteString(cmd)
			budget -= len(cmd)
			p.pending = true
			p.deadline = now.Add(kittyResponseTimeout)
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

// finish ends a frame, deleting the images longest off screen beyond the
// number kept for showing again.
func (k *kittyImages) finish() {
	var hidden []*kittyUpload
	for _, u := range k.uploads {
		u.lastVisible = u.visible
		if k.evictable(u) {
			hidden = append(hidden, u)
		}
	}
	if len(hidden) <= maxHiddenKittyImages {
		return
	}
	sort.Slice(hidden, func(i, j int) bool {
		if hidden[i].lastSeen != hidden[j].lastSeen {
			return hidden[i].lastSeen < hidden[j].lastSeen
		}
		return hidden[i].id < hidden[j].id
	})
	for _, u := range hidden[:len(hidden)-maxHiddenKittyImages] {
		k.evict(u)
	}
}

// cleanup returns the commands deleting every image from the terminal. An
// upload part-way through its chunks is ended first: the terminal takes no
// other graphics command until it is, and would read the deletes as more of
// its data. Its (now broken) image is deleted with the rest.
func (k *kittyImages) cleanup() string {
	var out bytes.Buffer
	if k.active != nil {
		out.WriteString(kittyCommand("m=0,q=2", ""))
	}
	for _, cmd := range k.garbage {
		out.WriteString(cmd)
	}
	for id := 1; id < 256; id++ {
		if k.ids[id] != nil {
			out.WriteString(kittyCommand(fmt.Sprintf("a=d,d=I,i=%d,q=2", id), ""))
		}
	}
	return out.String()
}

// reset returns cleanup's commands and forgets every image, keeping the turn
// of image ids.
func (k *kittyImages) reset() string {
	out := k.cleanup()
	*k = kittyImages{uploads: make(map[imageVariant]*kittyUpload), lastID: k.lastID, worker: k.worker, frame: k.frame}
	return out
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
						style := uv.Style{Fg: ansi.IndexedColor(u.id), UnderlineColor: ansi.IndexedColor(p.id)}
						if bg := b.cells[b.index(x, y)].background; bg.A != 0 {
							// Otherwise the terminal's background shows
							// through the image's transparent pixels.
							style.Bg = bg
						}
						dst.SetCell(x, y, &uv.Cell{Content: string([]rune{0x10eeee, imageDiacritics[y-tile.Y], imageDiacritics[x-tile.X]}), Width: 1, Style: style})
					}
				}
			}
		}
	}
}
