package main

import (
	"fmt"
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// SwitcherDemo demonstrates the Switcher widget with three tabs.
// Each tab preserves its state when switching between them.
type SwitcherDemo struct {
	activeTab *t.Signal[string]

	// State for the "Fruits" tab - preserved across switches
	fruitsListState *t.ListState[string]
	fruitsSelected  *t.Signal[string]

	// State for the "Colors" tab - preserved across switches
	colorsListState *t.ListState[string]
	colorsSelected  *t.Signal[string]

	// Counter for the "Counter" tab - preserved across switches
	counter *t.Signal[int]
}

func NewSwitcherDemo() *SwitcherDemo {
	return &SwitcherDemo{
		activeTab: t.NewSignal("fruits"),

		fruitsListState: t.NewListState([]string{
			"Apple", "Banana", "Cherry", "Date", "Elderberry",
			"Fig", "Grape", "Honeydew", "Kiwi", "Lemon",
		}),
		fruitsSelected: t.NewSignal(""),

		colorsListState: t.NewListState([]string{
			"Red", "Orange", "Yellow", "Green", "Blue",
			"Indigo", "Violet", "Pink", "Cyan", "Magenta",
		}),
		colorsSelected: t.NewSignal(""),

		counter: t.NewSignal(0),
	}
}

func (d *SwitcherDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		Top: []t.Widget{
			// Header
			t.Text{
				Content: " Switcher Demo ",
				Style: t.Style{
					ForegroundColor: theme.Background,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			// Tab bar
			d.buildTabBar(ctx),
		},
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Style: t.Style{
				Padding: t.EdgeInsetsAll(1),
			},
			Children: []t.Widget{
				t.Switcher{
					Active: d.activeTab.Get(),
					Height: t.Fr(1),
					Children: map[string]t.Widget{
						"fruits":  d.buildFruitsTab(ctx),
						"colors":  d.buildColorsTab(ctx),
						"counter": d.buildCounterTab(ctx),
					},
				},
			},
		},
	}
}

func (d *SwitcherDemo) buildTabBar(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	active := d.activeTab.Get()

	tabStyle := func(key string) t.Style {
		if key == active {
			return t.Style{
				ForegroundColor: theme.Background,
				BackgroundColor: theme.Accent,
				Padding:         t.EdgeInsetsXY(2, 0),
			}
		}
		return t.Style{
			ForegroundColor: theme.TextMuted,
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(2, 0),
		}
	}

	return t.Row{
		Style: t.Style{
			BackgroundColor: theme.Surface,
		},
		Children: []t.Widget{
			t.Text{Content: "[1] Fruits", Style: tabStyle("fruits")},
			t.Text{Content: "[2] Colors", Style: tabStyle("colors")},
			t.Text{Content: "[3] Counter", Style: tabStyle("counter")},
		},
	}
}

func (d *SwitcherDemo) buildFruitsTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Fruits List",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Spans: t.ParseMarkup("Navigate with [b $Accent]↑/↓[/] or [b $Accent]j/k[/], select with [b $Accent]Enter[/]", theme),
			},
			t.List[string]{
				ID:    "fruits-list",
				State: d.fruitsListState,
				OnSelect: func(item string) {
					d.fruitsSelected.Set(item)
				},
			},
			t.ShowWhen(d.fruitsSelected.Get() != "", t.Text{
				Content: fmt.Sprintf("Selected: %s", d.fruitsSelected.Get()),
				Style:   t.Style{ForegroundColor: theme.Success},
			}),
		},
	}
}

func (d *SwitcherDemo) buildColorsTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Colors List",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Spans: t.ParseMarkup("Navigate with [b $Accent]↑/↓[/] or [b $Accent]j/k[/], select with [b $Accent]Enter[/]", theme),
			},
			t.List[string]{
				ID:    "colors-list",
				State: d.colorsListState,
				OnSelect: func(item string) {
					d.colorsSelected.Set(item)
				},
			},
			t.ShowWhen(d.colorsSelected.Get() != "", t.Text{
				Content: fmt.Sprintf("Selected: %s", d.colorsSelected.Get()),
				Style:   t.Style{ForegroundColor: theme.Success},
			}),
		},
	}
}

func (d *SwitcherDemo) buildCounterTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	count := d.counter.Get()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Counter Demo",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Spans: t.ParseMarkup("Press [b $Accent]+[/] to increment, [b $Accent]-[/] to decrement", theme),
			},
			t.Text{
				Content: fmt.Sprintf("Count: %d", count),
				Style: t.Style{
					ForegroundColor: theme.Accent,
					Padding:         t.EdgeInsetsXY(0, 1),
				},
			},
			t.Text{
				Content: "(State is preserved when you switch tabs)",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

// Keybinds provides tab switching and counter controls.
func (d *SwitcherDemo) Keybinds() []t.Keybind {
	keybinds := []t.Keybind{
		{Key: "1", Name: "Fruits", Action: func() { d.activeTab.Set("fruits") }},
		{Key: "2", Name: "Colors", Action: func() { d.activeTab.Set("colors") }},
		{Key: "3", Name: "Counter", Action: func() { d.activeTab.Set("counter") }},
	}

	// Add counter keybinds only when on the counter tab
	if d.activeTab.Get() == "counter" {
		keybinds = append(keybinds,
			t.Keybind{Key: "+", Name: "Increment", Action: func() {
				d.counter.Set(d.counter.Get() + 1)
			}},
			t.Keybind{Key: "-", Name: "Decrement", Action: func() {
				d.counter.Set(d.counter.Get() - 1)
			}},
		)
	}

	return keybinds
}

func main() {
	app := NewSwitcherDemo()
	t.SetDebugLogging(true)
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
