// float-geometry-demo builds a summary from current-frame anchor geometry.
package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type anchor struct {
	left   t.Signal[int]
	narrow t.Signal[bool]
}

func (*anchor) WidgetID() string { return "geometry-anchor" }
func (a *anchor) Build(ctx t.BuildContext) t.Widget {
	width := t.Percent(75)
	if a.narrow.Get() {
		width = t.Percent(50)
	}
	theme := ctx.Theme()
	return t.Text{Content: "Anchor: summary follows size and position", Style: t.Style{
		Width: width, Height: t.Cells(3), Margin: t.EdgeInsets{Left: a.left.Get()},
		Border: t.RoundedBorder(theme.Primary), BackgroundColor: theme.Surface,
	}}
}

type app struct {
	anchor *anchor
	label  t.Signal[string]
	scroll *t.ScrollState
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "m", Name: "Move", Action: func() { a.anchor.left.Update(func(v int) int { return (v + 2) % 10 }) }},
		{Key: "w", Name: "Width", Action: func() { a.anchor.narrow.Update(func(v bool) bool { return !v }) }},
		{Key: "l", Name: "Label", Action: func() {
			a.label.Update(func(v string) string {
				if v == "Summary" {
					return "Changed"
				}
				return "Summary"
			})
		}},
		{Key: "s", Name: "Scroll", Action: func() { a.scroll.SetOffset(a.scroll.GetOffset() + 1) }},
		{Key: "r", Name: "Reset scroll", Action: func() { a.scroll.SetOffset(0) }},
	}
}
func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{Width: t.Flex(1), Height: t.Flex(1), Children: []t.Widget{
		t.Text{Content: "Deferred geometry: summary is present on frame one"},
		t.Text{Content: "m move | w width | l label | s scroll | r reset | ctrl+c quit"},
		t.Spacer{Height: t.Cells(2)},
		t.Scrollable{ID: "geometry-viewport", Height: t.Cells(6), Width: t.Flex(1), State: a.scroll,
			Child: t.Column{Width: t.Flex(1), Children: []t.Widget{
				a.anchor, t.Spacer{Height: t.Cells(12)}, t.Text{Content: "End of scroll content"},
			}},
		},
		t.Floating{Visible: true, Config: t.FloatConfig{AnchorID: "geometry-anchor", Anchor: t.AnchorBottomLeft},
			BuildChild: func(_ t.BuildContext, g t.FloatGeometry) t.Widget {
				if !g.AnchorFound || g.AnchorVisibleBounds.Height == 0 {
					return nil
				}
				return t.Text{ID: "geometry-summary", Content: fmt.Sprintf("%s: x=%d y=%d width=%d screen=%dx%d", a.label.Get(), g.AnchorBounds.X, g.AnchorBounds.Y, g.AnchorBounds.Width, g.Screen.Width, g.Screen.Height), Style: t.Style{
					Width: t.Cells(g.AnchorBounds.Width), BackgroundColor: theme.Primary, ForegroundColor: theme.Primary.AutoText(),
				}}
			},
		},
	}}
}
func main() {
	a := &app{anchor: &anchor{left: t.NewSignal(2), narrow: t.NewSignal(false)}, label: t.NewSignal("Summary"), scroll: t.NewScrollState()}
	if err := t.Run(a); err != nil {
		log.Fatal(err)
	}
}
