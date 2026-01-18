package main

import (
	"fmt"
	"log"

	t "terma"
)

// TabDemo demonstrates the TabBar and TabView widgets.
// It shows:
// - Basic tab navigation with arrow keys
// - Position-based keybinds (1-9)
// - Tab reordering with ctrl+left/right
// - Closable tabs
// - State preservation across tab switches
type TabDemo struct {
	tabState *t.TabState

	// State for the "Counter" tab - preserved across switches
	counter t.Signal[int]

	// State for the "List" tab
	listState *t.ListState[string]

	// Track closed tabs for demo
	closedTabs t.Signal[int]
}

func NewTabDemo() *TabDemo {
	tabs := []t.Tab{
		{Key: "home", Label: "Home"},
		{Key: "counter", Label: "Counter"},
		{Key: "list", Label: "List"},
		{Key: "info", Label: "Info"},
	}

	return &TabDemo{
		tabState: t.NewTabState(tabs),
		counter:  t.NewSignal(0),
		listState: t.NewListState([]string{
			"Apple", "Banana", "Cherry", "Date", "Elderberry",
		}),
		closedTabs: t.NewSignal(0),
	}
}

func (d *TabDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		Top: []t.Widget{
			// Header
			t.Text{
				Content: " TabBar Demo ",
				Style: t.Style{
					ForegroundColor: theme.Background,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			// Tab bar
			t.TabBar{
				ID:             "tabs",
				State:          d.tabState,
				KeybindPattern: t.TabKeybindNumbers,
				AllowReorder:   true,
				Closable:       true,
				OnTabClose: func(key string) {
					d.tabState.RemoveTab(key)
					d.closedTabs.Update(func(n int) int { return n + 1 })
				},
				OnTabChange: func(key string) {
					t.Log("Tab changed to: %s", key)
				},
				Style: t.Style{
					BackgroundColor: theme.Surface,
				},
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Style: t.Style{
				Padding: t.EdgeInsetsAll(1),
			},
			Height: t.Flex(1),
			Children: []t.Widget{
				d.buildContent(ctx),
			},
		},
	}
}

func (d *TabDemo) buildContent(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	activeKey := d.tabState.ActiveKey()

	switch activeKey {
	case "home":
		return d.buildHomeTab(ctx)
	case "counter":
		return d.buildCounterTab(ctx)
	case "list":
		return d.buildListTab(ctx)
	case "info":
		return d.buildInfoTab(ctx)
	default:
		return t.Text{
			Content: "Select a tab",
			Style:   t.Style{ForegroundColor: theme.TextMuted},
		}
	}
}

func (d *TabDemo) buildHomeTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Welcome to the TabBar Demo!",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Spans: t.ParseMarkup("Use [b $Accent]h/l[/] or [b $Accent]←/→[/] to switch tabs", theme),
			},
			t.Text{
				Spans: t.ParseMarkup("Press [b $Accent]1-4[/] to jump to specific tabs", theme),
			},
			t.Text{
				Spans: t.ParseMarkup("Use [b $Accent]ctrl+h/l[/] to reorder tabs", theme),
			},
			t.Text{
				Spans: t.ParseMarkup("Press [b $Accent]ctrl+w[/] to close the active tab", theme),
			},
			t.Text{
				Content: fmt.Sprintf("Tabs closed so far: %d", d.closedTabs.Get()),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (d *TabDemo) buildCounterTab(ctx t.BuildContext) t.Widget {
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

func (d *TabDemo) buildListTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "List Demo",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Spans: t.ParseMarkup("Navigate with [b $Accent]j/k[/] or [b $Accent]↑/↓[/]", theme),
			},
			t.List[string]{
				ID:    "fruit-list",
				State: d.listState,
			},
		},
	}
}

func (d *TabDemo) buildInfoTab(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "About TabBar",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{Content: "TabBar is a focusable widget that renders tabs."},
			t.Text{Content: "It supports:"},
			t.Text{Content: "  - Keyboard navigation (←/→)"},
			t.Text{Content: "  - Position keybinds (1-9, alt+1-9, ctrl+1-9)"},
			t.Text{Content: "  - Tab reordering (ctrl+←/→)"},
			t.Text{Content: "  - Closable tabs with × button"},
			t.Text{Content: "  - Click to select tab"},
		},
	}
}

// Keybinds provides tab-specific controls.
func (d *TabDemo) Keybinds() []t.Keybind {
	var keybinds []t.Keybind

	// Add counter keybinds only when on the counter tab
	if d.tabState.ActiveKeyPeek() == "counter" {
		keybinds = append(keybinds,
			t.Keybind{Key: "+", Name: "Increment", Action: func() {
				d.counter.Update(func(c int) int { return c + 1 })
			}},
			t.Keybind{Key: "-", Name: "Decrement", Action: func() {
				d.counter.Update(func(c int) int { return c - 1 })
			}},
		)
	}

	return keybinds
}

func main() {
	app := NewTabDemo()
	t.SetDebugLogging(true)
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
