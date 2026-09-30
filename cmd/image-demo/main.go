// Image demo uses generated pixels; no files or network access are required.
package main

import (
	"fmt"
	t "github.com/darrenburns/terma"
	"image"
	"image/color"
	"log"
	"math"
)

type app struct {
	source           *t.ImageResource
	fit              t.Signal[t.ImageFit]
	visible, overlay t.Signal[bool]
	scroll           *t.ScrollState
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
		{Key: "f", Name: "Fit", Action: func() { a.fit.Set((a.fit.Peek() + 1) % 3) }},
		{Key: "i", Name: "Hide/show", Action: func() { a.visible.Set(!a.visible.Peek()) }},
		{Key: "o", Name: "Backdrop", Action: func() { a.overlay.Set(!a.overlay.Peek()) }},
		{Key: "down", Name: "Scroll down", Action: func() { a.scroll.ScrollDown(2) }},
		{Key: "up", Name: "Scroll up", Action: func() { a.scroll.ScrollUp(2) }},
	}
}
func (a *app) Build(ctx t.BuildContext) t.Widget {
	fit := a.fit.Get()
	names := []string{"contain", "cover", "stretch"}
	children := []t.Widget{t.Text{Content: fmt.Sprintf("Static images · %s · f fit / i hide / o backdrop / arrows scroll", names[fit]), Style: t.Style{ForegroundColor: t.White, Padding: t.EdgeInsetsAll(1)}}}
	if a.visible.Get() {
		for i := 0; i < 6; i++ {
			children = append(children, t.Text{Content: fmt.Sprintf("Image %d — shared immutable resource", i+1)}, t.Image{ID: fmt.Sprintf("image-%d", i), Source: a.source, Fit: fit, Style: t.Style{Width: t.Cells(54), Height: t.Cells(14), Padding: t.EdgeInsetsAll(1), Border: t.RoundedBorder(t.Blue), BackgroundColor: t.RGB(20, 30, 50)}})
		}
	} else {
		children = append(children, t.Text{Content: "Images hidden — no stale graphics should remain."})
	}
	children = append(children, t.ShowWhen(a.overlay.Get(), t.Floating{Visible: true, Config: t.FloatConfig{Position: t.FloatPositionCenter, Modal: true, BackdropColor: t.Black.WithAlpha(0.6)}, Child: t.Text{Wrap: t.WrapSoft, Content: "Backdrop: tinted blocks behind this overlay. Press o to close.", Style: t.Style{Width: t.Cells(48), Height: t.Cells(5), Padding: t.EdgeInsetsAll(1), BackgroundColor: t.Blue}}}))
	return t.Dock{Bottom: []t.Widget{t.KeybindBar{}}, Body: t.Scrollable{ID: "images", State: a.scroll, Height: t.Flex(1), Width: t.Flex(1), Child: t.Column{Spacing: 1, Children: children}}}
}
func main() {
	pixels := image.NewNRGBA(image.Rect(0, 0, 320, 160))
	for y := 0; y < 160; y++ {
		for x := 0; x < 320; x++ {
			dx, dy := float64(x-160), float64(y-80)
			alpha := uint8(255)
			if math.Hypot(dx, dy) > 145 {
				alpha = 100
			}
			pixels.SetNRGBA(x, y, color.NRGBA{uint8(x * 255 / 319), uint8(y * 255 / 159), uint8(130 + 80*math.Sin(float64(x+y)/18)), alpha})
		}
	}
	source, err := t.NewImageResource(pixels)
	if err != nil {
		log.Fatal(err)
	}
	if err = t.Run(&app{source: source, fit: t.NewSignal(t.ImageContain), visible: t.NewSignal(true), overlay: t.NewSignal(false), scroll: t.NewScrollState()}); err != nil {
		log.Fatal(err)
	}
}
