package main

import (
	"errors"
	"fmt"
	"io"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

const (
	homeListID = "gallery-home-list"
	paletteID  = "gallery-palette"
	switchKey  = "ctrl+g"
)

// Entry is a demo the gallery can show.
type Entry struct {
	demokit.Info
	Command string // The standalone command, e.g. "./cmd/table-demo"
	New     func() demokit.Demo
}

// Gallery shows one demo at a time, with a home page listing them all and a
// command palette for switching. Each demo is created the first time it is
// opened and kept afterwards, so its state survives switching away and back.
type Gallery struct {
	entries   []Entry
	instances map[string]demokit.Demo
	active    t.Signal[string] // Key of the demo on screen; "" for the home page
	palette   *t.CommandPaletteState
	home      *t.ListState[Entry]
	homeWheel *t.ScrollState
}

func NewGallery(entries []Entry) *Gallery {
	return &Gallery{
		entries:   entries,
		instances: make(map[string]demokit.Demo),
		active:    t.NewSignal(""),
		palette:   t.NewCommandPaletteState("Demos", nil),
		home:      t.NewListState(entries),
		homeWheel: t.NewScrollState(),
	}
}

func (g *Gallery) InitialFocus() string { return homeListID }

func (g *Gallery) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: switchKey, Name: "Demos", Action: g.togglePalette},
	}
}

// show switches to the demo with the given key, or to the home page for "".
// Returning home puts the list's cursor on the demo just left.
func (g *Gallery) show(key string) {
	focusID := homeListID
	if key != "" {
		focusID = g.instance(key).InitialFocus()
	} else if previous := g.active.Peek(); previous != "" {
		for i, entry := range g.entries {
			if entry.Key == previous {
				g.home.SelectIndex(i)
			}
		}
	}
	g.active.Set(key)
	if g.palette.Visible.Peek() {
		// The palette hands focus back as it closes.
		g.palette.SetNextFocusIDOnClose(focusID)
		g.palette.Close()
	} else if focusID != "" {
		t.RequestFocus(focusID)
	}
}

func (g *Gallery) instance(key string) demokit.Demo {
	if demo, ok := g.instances[key]; ok {
		return demo
	}
	for _, entry := range g.entries {
		if entry.Key == key {
			demo := entry.New()
			g.instances[key] = demo
			return demo
		}
	}
	return nil
}

func (g *Gallery) togglePalette() {
	if g.palette.Visible.Peek() {
		g.palette.Close()
		return
	}
	g.palette.SetItems(g.paletteItems())
	g.palette.Open()
}

func (g *Gallery) paletteItems() []t.CommandPaletteItem {
	active := g.active.Peek()
	items := []t.CommandPaletteItem{
		{
			Label:       "Home",
			Description: "Every demo at a glance",
			Current:     active == "",
			Action:      func() { g.show("") },
		},
		{Divider: "Demos"},
	}
	for _, entry := range g.entries {
		hint := ""
		if _, open := g.instances[entry.Key]; open && entry.Key != active {
			hint = "open"
		}
		items = append(items, t.CommandPaletteItem{
			Label:       entry.Title,
			Description: entry.Description,
			Hint:        hint,
			Current:     entry.Key == active,
			FilterText:  entry.Title + " " + entry.Description,
			Action:      func() { g.show(entry.Key) },
		})
	}
	return items
}

func (g *Gallery) Build(ctx t.BuildContext) t.Widget {
	var body t.Widget = homePage{gallery: g}
	if key := g.active.Get(); key != "" {
		body = g.instance(key)
	}
	// The palette comes after the demo so it opens above, and takes focus
	// from, any overlay the demo has open.
	return t.Column{
		ID:     "gallery",
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Children: []t.Widget{
			body,
			t.CommandPalette{
				ID:          paletteID,
				State:       g.palette,
				Position:    t.FloatPositionTopCenter,
				Placeholder: "Jump to a demo…",
				Style:       t.Style{Width: t.Cells(64), Height: t.Cells(20)},
			},
		},
	}
}

// homePage lists every demo, with the one under the cursor described alongside.
type homePage struct {
	gallery *Gallery
}

func (h homePage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	g := h.gallery
	list := t.List[Entry]{
		ID:          homeListID,
		State:       g.home,
		ScrollState: g.homeWheel,
		OnSelect:    func(entry Entry) { g.show(entry.Key) },
		RenderItem: func(entry Entry, active bool, _ bool) t.Widget {
			style := t.Style{Padding: t.EdgeInsetsXY(1, 0), Width: t.Flex(1)}
			title, description := theme.Text, theme.TextMuted
			if active {
				style.BackgroundColor = theme.ActiveCursor
				title, description = theme.SelectionText, theme.SelectionText
			}
			return t.Row{
				Style: style,
				Children: []t.Widget{
					t.Text{Content: entry.Title, Width: t.Cells(22), Style: t.Style{ForegroundColor: title, Bold: true}},
					t.Text{Content: entry.Description, Style: t.Style{ForegroundColor: description}},
				},
			}
		},
		Style: t.Style{Width: t.Flex(1)},
	}

	return t.Dock{
		Style:  t.Style{BackgroundColor: theme.Background},
		Top:    []t.Widget{demokit.Header{Title: "Terma Demos", Tagline: "Every widget demo in one app"}},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				t.Column{
					Width:  t.Flex(1),
					Height: t.Flex(1),
					Style:  demokit.PanelStyle(theme, fmt.Sprintf("Demos · %d", len(g.entries)), ctx.IsFocused(list)),
					Children: []t.Widget{
						t.Scrollable{
							State: g.homeWheel,
							Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)},
							Child: list,
						},
					},
				},
				t.Column{
					Width:   t.Cells(34),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						demokit.Fill(aboutPanel{gallery: g}),
						homeKeysPanel{},
					},
				},
			},
		},
	}
}

// aboutPanel describes the demo under the home list's cursor.
type aboutPanel struct {
	gallery *Gallery
}

func (a aboutPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	entry, ok := a.gallery.home.SelectedItem()
	a.gallery.home.CursorIndex.Get() // Rebuild as the cursor moves
	children := []t.Widget{}
	if ok {
		status := "[$TextMuted]not opened yet[/]"
		if _, open := a.gallery.instances[entry.Key]; open {
			status = "[$Success]open · state kept[/]"
		}
		children = append(children,
			t.ParseMarkupToText(fmt.Sprintf("[b $Primary]%s[/]", entry.Title), theme),
			t.Text{Content: entry.Description, Wrap: t.WrapSoft, Style: t.Style{ForegroundColor: theme.Text}},
			t.Text{},
			demokit.StatRow(theme, "Status", status),
			t.Text{},
			t.Text{Content: "Run it on its own:", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Text{Content: "go run " + entry.Command, Wrap: t.WrapSoft, Style: t.Style{ForegroundColor: theme.Accent}},
		)
	}
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "About", false),
		Children: children,
	}
}

type homeKeysPanel struct{}

func (homeKeysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("↑↓ jk", "$Info", "choose a demo"),
			key("enter", "$Success", "open it"),
			key(switchKey, "$Accent", "switch demo"),
			key("t", "$Warning", "cycle theme (in demos)"),
			key("ctrl+c", "$Error", "quit"),
		},
	}
}

// Close releases resources owned by demos after the gallery exits.
func (g *Gallery) Close() error {
	var result error
	for key, demo := range g.instances {
		if closer, ok := demo.(io.Closer); ok {
			result = errors.Join(result, closer.Close())
		}
		delete(g.instances, key)
	}
	return result
}
