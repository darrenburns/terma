# Image

`Image` displays an already decoded, static Go `image.Image`. Create its resource
once and reuse it across widgets and renders:

```go
resource, err := terma.NewImageResource(decoded)
if err != nil {
    return err
}
widget := terma.Image{
    ID: "thumbnail",
    Source: resource,
    Fit: terma.ImageContain,
    Style: terma.Style{Width: terma.Cells(40), Height: terma.Cells(12)},
}
```

Resources copy pixels into immutable, zero-origin NRGBA storage. Nil and empty
inputs are errors. Loading files or URLs remains the application's responsibility.
A nil widget Source draws nothing. Custom widgets can call
`ctx.DrawImage(x, y, width, height, resource, fit)` with cell coordinates.

`ImageContain` (the default) centers the entire image, `ImageCover` fills the box
with a centered crop, and `ImageStretch` fills it without preserving aspect ratio.
Ordinary dimensions, min/max constraints, padding, borders, margins and scrolling
apply. Layout uses pixel dimensions and cell metrics, independently of protocol.
Before metrics arrive and in headless rendering, a cell is 8×16 logical pixels.

## Terminal selection

`TERMA_IMAGE_PROTOCOL=auto|kitty|sixel|blocks` selects the backend. Auto lazily probes
on first use, showing blocks while replies are pending. Detection ends at DA1 or
a 350 ms timeout. Late capability and geometry replies are accepted.

| Terminal capability | Auto output |
| --- | --- |
| Kitty ≥0.28 or Ghostty, successful graphics query, 256-colour or truecolour profile | Unicode placeholders |
| Sixel advertised in DA1 | Sixel |
| Unknown implementation, missing pixel geometry, unsupported output | Coloured `▀` blocks |
| Recognized tmux/screen environment or identity | Blocks |

Explicit `kitty` asserts placeholder support but still requires at least 256
colours and valid pixel geometry. Ghostty uses Kitty placeholders automatically
when its identity and graphics replies confirm support; no override is needed.
Explicit `sixel` asserts Sixel support. Environment checks cannot prove
there is no hidden multiplexer in an SSH connection; passthrough is not provided.
Direct terminals and SSH use in-band image data, with no shared-filesystem need.

`TERMA_IMAGE_CELL_SIZE=WIDTHxHEIGHT` overrides cell metrics for the terminal session.
Otherwise CSI 16 t takes precedence over valid window pixels divided by the grid.
Nonpositive, excessive (>4096), malformed and fractional fallback dimensions are
ignored. Exact cell replies supersede window estimates, which can include padding.

## Composition and costs

The logical screen always contains a half-block preview. Snapshots, accessible
terminal text and screen exports retain it. Every later cell write covers image
cells, including identical writes; translucent backdrops produce tinted blocks.
Kitty retains source alpha, while Sixel and blocks flatten alpha over the effective
background. With no background set, fully transparent pixels show the terminal's
own background and partly transparent ones are blended over black. Composition is
cell based, not pixel based.

Kitty uploads chunked PNGs in batches of up to 256 KiB per frame, retaining blocks
until upload and virtual placement acknowledgements succeed; encoding runs off the
event loop, and the encoded data is released once the terminal has the image.
Movement, scrolling and window resizes reuse uploads.
Missing uploads retry once; repeated errors or a missing acknowledgement fall back
to blocks. Cropping, stretching or pixel-size changes can create a transformed
variant. Tiles use explicit row/column diacritics. Image IDs are limited to 1–255,
and each cached upload supports 255 virtual placements. IDs referenced by the
current or previous presentation cannot be recycled; excess images stay as blocks.
IDs are handed out in turn, so a late reply about a deleted image is never taken
for its successor. At most 16 images are kept uploaded while off screen, and images
whose source has been garbage collected are deleted. Only session-owned image IDs
are deleted, including during shutdown and resume; an upload interrupted by either
is ended before the deletes.

Sixel caches flattened, scaled palettes (255 opaque colours plus transparency),
prepared off the event loop, and reuses prepared pixels for zero-origin crops.
Visible runs are emitted separately so covered cells and neighbouring images are
preserved. Images already on screen are left alone; only when one is removed or
changed does a frame conservatively erase and redraw the terminal to remove stale
pixels. Synchronized output wraps this when supported; other terminals may flicker
then. The terminal's last row always shows blocks, since many terminals scroll
after a Sixel image that reaches it.
Hardware scrolling optimization remains disabled. Cursor state is restored and
synchronization is ended even when output fails.

Native preparation is limited to 4 megapixels per variant. Kitty encoded data and
Sixel caches each have a 64 MiB budget; Sixel retains at most 16 prepared variants
and 64 crop payloads per variant. Larger/excess variants use blocks. Large or
frequently resized images and gradients with many cell backgrounds cost more to
prepare. There is no animation, direct Kitty placement, multiplexer passthrough,
or general pixel compositing.

## Demo and verification

```sh
go run ./cmd/terma-browser -- go run ./cmd/image-demo
TERMA_IMAGE_PROTOCOL=blocks go run ./cmd/terma-browser -- go run ./cmd/image-demo
go test ./...
go test -run '^$' -bench 'BenchmarkImagePresentation|BenchmarkCollectionScroll' -benchmem .
```

In the demo, `f` cycles fits, `i` hides/shows images, `o` toggles a backdrop, and
arrows or the mouse wheel scroll. Use `?cols=100&rows=30` before the launcher's token
fragment for repeatable sizing. Native Kitty acceptance must also cover overlays,
scrolling, removal, resize and suspension; browser Sixel checks cannot certify it.

## Ultraviolet replacement

The module pins `github.com/darrenburns/ultraviolet` at commit
`8cafcd59bdf3983d46b9b3ccfe07e054c90ffa7f` through `replace`, preserving original
module and import paths. It is based on upstream `6b0c0e26fad9`. The isolated patch
makes colour equality preserve indexed encodings, including palette entries with
the same RGB value and indexed/RGB transitions. The fork includes cell and SGR
regressions; Terma tests emitted bytes in ANSI256 and truecolour profiles.
Remove the replacement once the pinned upstream version has the equivalent fix.

**Applications using Terma must carry the same replacement in their own `go.mod`:**
Go does not propagate a dependency's replacements to its consumers. Terma checks
the linked library's encoding behavior and disables Kitty placeholders if the
fix is absent, including with an explicit `kitty` override.

```go
replace github.com/charmbracelet/ultraviolet => github.com/darrenburns/ultraviolet v0.0.0-20260930153435-8cafcd59bdf3
```
