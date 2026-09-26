// This demo reproduces the custom-row sizing and modal-backdrop regressions.
// Run with "list" (default) or "backdrop". Ctrl+B changes the backdrop; Ctrl+Q quits.
package main

import (
	"os"

	t "github.com/darrenburns/terma"
)

type app struct {
	mode  string
	items *t.ListState[string]
	color t.Signal[t.Color]
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+b", Name: "Red backdrop", Action: func() { a.color.Set(t.RGBA(255, 0, 0, 0.5)) }},
		{Key: "ctrl+q", Name: "Quit", Action: t.Quit},
	}
}

func (a *app) Build(t.BuildContext) t.Widget {
	if a.mode == "backdrop" {
		return t.Column{
			Style: t.Style{Width: t.Flex(1), Height: t.Flex(1), BackgroundColor: t.RGB(255, 255, 255)},
			Children: []t.Widget{
				t.Text{Content: "MODAL BACKDROP / Ctrl+B changes black to red", Style: t.Style{ForegroundColor: t.Black}},
				t.Text{Content: "Expected: the entire background turns pink.", Style: t.Style{ForegroundColor: t.Black}},
				backdropOwner{color: a.color},
			},
		}
	}
	return t.Column{Children: []t.Widget{
		t.Text{Content: "CUSTOM LIST ROWS / two rows with Height: Flex(1)"},
		t.Text{Content: "Expected: each colored row occupies five lines."},
		t.List[string]{
			ID: "list", State: a.items, Width: t.Flex(1), Height: t.Cells(10),
			RenderItem: func(item string, active, selected bool) t.Widget {
				bg := t.Hex("#164e63")
				if item == "ROW TWO" {
					bg = t.Hex("#5b21b6")
				}
				return t.Text{Content: item, Style: t.Style{Height: t.Flex(1), BackgroundColor: bg, ForegroundColor: t.Hex("#ffffff")}}
			},
		},
	}}
}

type backdropOwner struct{ color t.Signal[t.Color] }

func (b backdropOwner) Build(t.BuildContext) t.Widget {
	return t.Floating{
		Visible: true,
		Config:  t.FloatConfig{Modal: true, Position: t.FloatPositionCenter, BackdropColor: b.color.Get()},
		Child: t.Text{Content: "MODAL STAYS OPEN", Style: t.Style{
			Width: t.Cells(24), Height: t.Cells(3), BackgroundColor: t.Hex("#164e63"), ForegroundColor: t.Hex("#ffffff"),
		}},
	}
}

func main() {
	mode := "list"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if err := t.Run(&app{mode: mode, items: t.NewListState([]string{"ROW ONE", "ROW TWO"}), color: t.NewSignal(t.RGBA(0, 0, 0, 0.5))}); err != nil {
		panic(err)
	}
}
