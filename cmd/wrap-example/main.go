//go:build ignore
// +build ignore

// DISABLED: Uses t.Dock and t.InitDebug which don't exist yet

package main

import (
	"fmt"
	"log"

	t "terma"
)

const sampleText = `The quick brown fox jumps over the lazy dog. This is a longer sentence to demonstrate how text wrapping works in different modes. Here's another sentence with some longer words like "extraordinary" and "phenomenal" to test word breaking.`

type App struct {
	width    t.Signal[int]
	height   t.Signal[int]
	wrapMode t.Signal[t.WrapMode]
}

func NewApp() *App {
	return &App{
		width:    t.NewSignal(40),
		height:   t.NewSignal(6),
		wrapMode: t.NewSignal(t.WrapSoft),
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "w", Name: "Cycle wrap", Action: a.cycleWrap},
		{Key: "left", Name: "Width -", Action: func() { a.adjustWidth(-1) }},
		{Key: "right", Name: "Width +", Action: func() { a.adjustWidth(1) }},
		{Key: "down", Name: "Height +", Action: func() { a.adjustHeight(1) }},
		{Key: "up", Name: "Height -", Action: func() { a.adjustHeight(-1) }},
	}
}

func (a *App) cycleWrap() {
	current := a.wrapMode.Get()
	switch current {
	case t.WrapSoft:
		a.wrapMode.Set(t.WrapHard)
	case t.WrapHard:
		a.wrapMode.Set(t.WrapNone)
	case t.WrapNone:
		a.wrapMode.Set(t.WrapSoft)
	}
}

func (a *App) adjustWidth(delta int) {
	newWidth := a.width.Get() + delta
	if newWidth < 10 {
		newWidth = 10
	}
	if newWidth > 100 {
		newWidth = 100
	}
	a.width.Set(newWidth)
}

func (a *App) adjustHeight(delta int) {
	newHeight := a.height.Get() + delta
	if newHeight < 1 {
		newHeight = 1
	}
	if newHeight > 20 {
		newHeight = 20
	}
	a.height.Set(newHeight)
}

func wrapModeName(mode t.WrapMode) string {
	switch mode {
	case t.WrapSoft:
		return "Soft (word boundaries)"
	case t.WrapHard:
		return "Hard (character boundary)"
	case t.WrapNone:
		return "None (truncate)"
	default:
		return "Unknown"
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	width := a.width.Get()
	height := a.height.Get()
	wrapMode := a.wrapMode.Get()

	return t.Dock{
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				t.Text{
					Content: "Text Wrapping Demo",
					Wrap:    t.WrapNone,
					Style: t.Style{
						ForegroundColor: t.Black,
						BackgroundColor: t.Cyan,
						Padding:         t.EdgeInsetsXY(1, 0),
					},
				},

				t.Text{
					Content: fmt.Sprintf("Width: %d cells  |  Height: %d lines  |  Wrap: %s", width, height, wrapModeName(wrapMode)),
					Wrap:    t.WrapNone,
				},

				// The text box with configurable dimensions and wrap mode
				t.Text{
					Content: sampleText,
					Width:   t.Cells(width),
					Height:  t.Cells(height),
					Wrap:    wrapMode,
					Style: t.Style{
						BackgroundColor: t.Hex("#2a2a3a"),
						Padding:         t.EdgeInsetsXY(1, 0),
					},
				},
			},
		},
	}
}

func main() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
	t.InitDebug()

	app := NewApp()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
