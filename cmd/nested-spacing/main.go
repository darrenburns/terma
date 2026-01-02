package main

import (
	"fmt"
	"log"
	"math/rand"

	t "terma"
)

func init() {
	// Initialize logging for debugging
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// HoverText is a Text widget that changes its background color on hover.
type HoverText struct {
	ID              string
	Content         string
	Click           func()
	BaseColor       t.Color
	backgroundColor *t.Signal[t.Color]
}

func NewHoverText(id, content string, baseColor t.Color, click func()) *HoverText {
	return &HoverText{
		ID:              id,
		Content:         content,
		Click:           click,
		BaseColor:       baseColor,
		backgroundColor: t.NewSignal(baseColor),
	}
}

func (h *HoverText) Key() string {
	return h.ID
}

func (h *HoverText) Build(ctx t.BuildContext) t.Widget {
	return t.Text{
		ID:      h.ID,
		Content: h.Content,
		Click:   h.Click,
		Hover:   h.OnHover,
		Style: t.Style{
			Padding:         t.EdgeInsetsAll(1),
			Margin:          t.EdgeInsetsAll(1),
			BackgroundColor: h.backgroundColor.Get(),
		},
	}
}

func (h *HoverText) OnHover(hovered bool) {
	if hovered {
		colors := []t.Color{t.Green, t.Yellow, t.Magenta, t.Cyan, t.Blue}
		h.backgroundColor.Set(colors[rand.Intn(len(colors))])
	} else {
		h.backgroundColor.Set(h.BaseColor)
	}
}

func (h *HoverText) OnClick() {
	if h.Click != nil {
		h.Click()
	}
}

type NestedSpacing struct {
	clickedKey *t.Signal[string]
	text1      *HoverText
	text2      *HoverText
	text3      *HoverText
	text4      *HoverText
}

func NewNestedSpacing() *NestedSpacing {
	n := &NestedSpacing{
		clickedKey: t.NewSignal(""),
	}
	// Create HoverText widgets that manage their own hover state
	n.text1 = NewHoverText("text-1", "text-1\nclick me", t.Red, func() { n.clickedKey.Set("text-1") })
	n.text2 = NewHoverText("text-2", "text-2", t.Red, func() { n.clickedKey.Set("text-2") })
	n.text3 = NewHoverText("text-3", "text-3", t.Red, func() { n.clickedKey.Set("text-3") })
	n.text4 = NewHoverText("text-4", "text-4", t.Red, func() { n.clickedKey.Set("text-4") })
	return n
}

func (n *NestedSpacing) Build(ctx t.BuildContext) t.Widget {
	clicked := n.clickedKey.Get()
	hoveredID := ctx.HoveredID()

	statusText := "Move mouse over widgets to hover, click to select"
	if clicked != "" || hoveredID != "" {
		statusText = fmt.Sprintf("Clicked: %q  |  Hovered: %q", clicked, hoveredID)
	}

	return t.Column{
		ID: "root-column",
		Children: []t.Widget{
			t.Text{
				ID:      "status",
				Content: statusText,
				Style:   t.Style{BackgroundColor: t.BrightBlack, ForegroundColor: t.White},
			},
			n.text1,
			t.Column{
				ID:       "inner-column",
				Children: []t.Widget{n.text2, n.text3},
				Style: t.Style{
					Padding:         t.EdgeInsetsAll(1),
					Margin:          t.EdgeInsetsAll(1),
					BackgroundColor: t.Blue,
				},
			},
			n.text4,
		},
		Style: t.Style{
			Padding:         t.EdgeInsetsAll(1),
			Margin:          t.EdgeInsetsXY(1, 0),
			BackgroundColor: t.Cyan,
		},
	}
}

func main() {
	app := NewNestedSpacing()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
