package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// DockDemo demonstrates the Dock widget with header, footer, sidebar, and scrollable body.
type DockDemo struct {
	scrollState  *t.ScrollState
	sidebarFirst *t.Signal[bool]
	selectedItem *t.Signal[string]
}

func NewDockDemo() *DockDemo {
	return &DockDemo{
		scrollState:  t.NewScrollState(),
		sidebarFirst: t.NewSignal(false),
		selectedItem: t.NewSignal("Home"),
	}
}

// Keybinds returns app-level keybindings displayed in the KeybindBar.
func (d *DockDemo) Keybinds() []t.Keybind {
	// Dynamic name based on current state
	toggleName := "Sidebar full height"
	if d.sidebarFirst.Get() {
		toggleName = "Header full width"
	}

	return []t.Keybind{
		{Key: "t", Name: toggleName, Action: func() {
			d.sidebarFirst.Update(func(v bool) bool { return !v })
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (d *DockDemo) Build(ctx t.BuildContext) t.Widget {
	// Determine dock order based on toggle
	var dockOrder []t.Edge
	if d.sidebarFirst.Get() {
		dockOrder = []t.Edge{t.Left, t.Top, t.Bottom, t.Right}
	} else {
		dockOrder = []t.Edge{t.Top, t.Bottom, t.Left, t.Right}
	}

	return t.Dock{
		ID:        "main-dock",
		DockOrder: dockOrder,
		Top: []t.Widget{
			d.buildHeader(),
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: t.BrightBlack,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Left: []t.Widget{
			d.buildSidebar(),
		},
		Body: d.buildBody(),
	}
}

func (d *DockDemo) buildHeader() t.Widget {
	return t.Text{
		Content: " Dock Demo ",
		Width:   t.Fr(1),
		Style: t.Style{
			ForegroundColor: t.Black,
			BackgroundColor: t.Cyan,
		},
	}
}

func (d *DockDemo) buildSidebar() t.Widget {
	items := []string{"Home", "Settings", "Profile", "Help"}
	selected := d.selectedItem.Get()

	children := make([]t.Widget, len(items))
	for i, item := range items {
		itemCopy := item
		bg := t.Blue
		if item == selected {
			bg = t.BrightBlue
		}
		children[i] = t.Text{
			Content: fmt.Sprintf(" %s ", item),
			Style: t.Style{
				ForegroundColor: t.White,
				BackgroundColor: bg,
			},
			Click: func() {
				d.selectedItem.Set(itemCopy)
			},
		}
	}

	return t.Column{
		Width: t.Cells(12),
		Style: t.Style{
			BackgroundColor: t.Blue,
		},
		Children: children,
	}
}

func (d *DockDemo) buildBody() t.Widget {
	// Generate scrollable content
	var lines []string
	lines = append(lines, fmt.Sprintf("Selected: %s", d.selectedItem.Get()))
	lines = append(lines, "")
	lines = append(lines, "This is the main body area.")
	lines = append(lines, "It fills the remaining space after docking.")
	lines = append(lines, "")
	for i := 1; i <= 20; i++ {
		lines = append(lines, fmt.Sprintf("Content line %d", i))
	}

	return t.Scrollable{
		ID:     "body-scroll",
		State:  d.scrollState,
		Width:  t.Fr(1),
		Height: t.Fr(1),
		Child: t.Text{
			Content: strings.Join(lines, "\n"),
			Width:   t.Fr(1),
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 0),
			},
		},
	}
}

func main() {
	app := NewDockDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
