// Package palettedemo demonstrates CommandPalette: fuzzy search, nested
// levels, disabled items and live theme previews.
package palettedemo

import (
	"fmt"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "command-palette",
	Title:       "Command Palette",
	Description: "Fuzzy-search commands, nested levels and live theme previews",
}

// CommandPaletteDemo shows a filterable CommandPalette with dividers, hints,
// descriptions, disabled items, nested levels and live theme previews.
//
//	ctrl+p    - Open the command palette
//	t         - Open the palette straight at the Themes level
//	↑↓        - Move through the results
//	enter     - Run a command, or open a nested level
//	backspace - Go back a level when the query is empty
//	escape    - Go back a level, or close the palette at the top level
type CommandPaletteDemo struct {
	palette *t.CommandPaletteState

	// preview is the label of the item under the palette's cursor.
	preview t.Signal[string]
	// path is the palette's breadcrumb trail, kept in a signal so the state
	// panel updates when a level is pushed or popped.
	path t.Signal[string]
	// themeBeforePreview is restored if the Themes level is left without
	// choosing a theme. Empty while no preview is showing.
	themeBeforePreview string

	activity    t.AnySignal[[]activityEntry]
	runs        t.Signal[int]
	showSidebar t.Signal[bool]
	wordWrap    t.Signal[bool]
}

// activityEntry is one line in the activity log: a command run, or a dismissal.
type activityEntry struct {
	label     string
	dismissed bool
}

const (
	themesPaletteTitle = "Themes"
	sidebarWidth       = 32
	maxActivity        = 50
)

// New creates the demo.
func New() demokit.Demo {
	app := &CommandPaletteDemo{
		preview:     t.NewSignal(""),
		path:        t.NewSignal(""),
		activity:    t.NewAnySignal[[]activityEntry](nil),
		runs:        t.NewSignal(0),
		showSidebar: t.NewSignal(true),
		wordWrap:    t.NewSignal(false),
	}

	app.palette = t.NewCommandPaletteState("Commands", []t.CommandPaletteItem{
		{Label: "New File", Hint: "Ctrl+N", Action: app.selectAction("New File")},
		{Label: "Open File", Hint: "Ctrl+O", Action: app.selectAction("Open File")},
		{Label: "Save", Hint: "Ctrl+S", Action: app.selectAction("Save")},
		{Label: "Save As", Hint: "Ctrl+Shift+S", Action: app.selectAction("Save As")},
		{Label: "Revert File", Description: "Nothing to revert", Disabled: true},
		{Divider: "Edit"},
		{Label: "Cut", Hint: "Ctrl+X", Action: app.selectAction("Cut")},
		{Label: "Copy", Hint: "Ctrl+C", Action: app.selectAction("Copy")},
		{Label: "Paste", Hint: "Ctrl+V", Action: app.selectAction("Paste")},
		{Label: "Find in Files", Hint: "Ctrl+Shift+F", Action: app.selectAction("Find in Files")},
		{Divider: "View"},
		{
			Label:       "Toggle Sidebar",
			Hint:        "Ctrl+B",
			Description: "Show or hide the panels on the right",
			Action: app.run("Toggle Sidebar", func() {
				app.showSidebar.Update(func(v bool) bool { return !v })
			}),
		},
		{
			Label:      "Toggle Word Wrap",
			Hint:       "Alt+Z",
			FilterText: "Toggle Word Wrap soft wrap long lines",
			Action: app.run("Toggle Word Wrap", func() {
				app.wordWrap.Update(func(v bool) bool { return !v })
			}),
		},
		{Label: "Profile Settings", Action: app.selectAction("Profile Settings")},
		{
			Label:         themesPaletteTitle,
			ChildrenTitle: themesPaletteTitle,
			Description:   "Preview themes as you move",
			Children:      app.themeItems,
		},
		{
			Label:         "Recent",
			ChildrenTitle: "Recent Files",
			Children: func() []t.CommandPaletteItem {
				return []t.CommandPaletteItem{
					{Label: "app/main.go", Action: app.selectAction("Open app/main.go")},
					{Label: "internal/config.yaml", Action: app.selectAction("Open internal/config.yaml")},
					{Label: "README.md", Action: app.selectAction("Open README.md")},
				}
			},
		},
		{Divider: "Activity"},
		{Label: "Clear Activity", Action: func() {
			app.activity.Set(nil)
			app.runs.Set(0)
			app.palette.Close(false)
		}},
	})

	return app
}

func (a *CommandPaletteDemo) InitialFocus() string { return "open-palette-btn" }

// run returns a command that performs effect, logs label and closes the palette.
func (a *CommandPaletteDemo) run(label string, effect func()) func() {
	return func() {
		if effect != nil {
			effect()
		}
		a.record(activityEntry{label: label})
		a.palette.Close(false)
	}
}

func (a *CommandPaletteDemo) selectAction(label string) func() {
	return a.run(label, nil)
}

func (a *CommandPaletteDemo) selectTheme(themeName string) {
	a.run("Theme "+t.ThemeDisplayName(themeName), func() {
		t.SetTheme(themeName)
		a.themeBeforePreview = ""
	})()
}

func (a *CommandPaletteDemo) record(entry activityEntry) {
	if !entry.dismissed {
		a.runs.Update(func(n int) int { return n + 1 })
	}
	a.activity.Update(func(entries []activityEntry) []activityEntry {
		entries = append([]activityEntry{entry}, entries...)
		if len(entries) > maxActivity {
			entries = entries[:maxActivity]
		}
		return entries
	})
}

func (a *CommandPaletteDemo) themeItems() []t.CommandPaletteItem {
	return t.ThemePaletteItems(a.selectTheme)
}

func (a *CommandPaletteDemo) togglePalette() {
	if a.palette.Visible.Peek() {
		a.restoreTheme()
		a.palette.Close(false)
		return
	}
	a.open()
}

// open shows the palette and seeds the state panel with its starting position.
func (a *CommandPaletteDemo) open() {
	a.palette.Open()
	a.syncPosition()
}

// openThemes opens the palette directly at the Themes level.
func (a *CommandPaletteDemo) openThemes() {
	a.palette.PushLevel(themesPaletteTitle, a.themeItems())
	a.open()
}

// syncPosition copies the palette's breadcrumb trail and cursor item into signals.
func (a *CommandPaletteDemo) syncPosition() {
	a.path.Set(strings.Join(a.palette.BreadcrumbPath(), " › "))
	if item, ok := a.palette.CurrentItem(); ok {
		a.preview.Set(item.Label)
	}
}

func (a *CommandPaletteDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+p", Name: "Command palette", Action: a.togglePalette},
		{Key: "t", Name: "Themes", Action: a.openThemes},
	}
}

// handleCursorChange previews the theme under the cursor in the Themes level,
// and restores the original theme once the cursor leaves it.
func (a *CommandPaletteDemo) handleCursorChange(item t.CommandPaletteItem) {
	a.syncPosition()
	level := a.palette.CurrentLevel()
	themeName, isTheme := item.Data.(string)
	if level == nil || level.Title != themesPaletteTitle || !isTheme {
		a.restoreTheme()
		return
	}
	if a.themeBeforePreview == "" {
		a.themeBeforePreview = t.CurrentThemeName()
	}
	t.SetTheme(themeName)
}

func (a *CommandPaletteDemo) handleDismiss() {
	a.restoreTheme()
	a.record(activityEntry{label: "Palette dismissed", dismissed: true})
}

func (a *CommandPaletteDemo) restoreTheme() {
	if a.themeBeforePreview != "" {
		t.SetTheme(a.themeBeforePreview)
		a.themeBeforePreview = ""
	}
}

func (a *CommandPaletteDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	main := t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			launchPanel{demo: a},
			lastCommandPanel{demo: a},
			demokit.Fill(activityPanel{demo: a}),
		},
	}
	body := []t.Widget{main}
	// Append the sidebar only when shown: an empty child would still take a
	// Spacing gap in the Row.
	if a.showSidebar.Get() {
		body = append(body, t.Column{
			Width:   t.Cells(sidebarWidth),
			Height:  t.Flex(1),
			Spacing: 1,
			Children: []t.Widget{
				statePanel{demo: a},
				keysPanel{},
			},
		})
	}

	return t.Dock{
		ID: "command-palette-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			demokit.Header{Title: "Command Palette", Tagline: "Search, nest and preview commands"},
			// The palette floats above everything, so where it sits in the
			// tree doesn't matter. Here it takes no space in the layout.
			t.CommandPalette{
				ID:             "command-palette",
				State:          a.palette,
				Position:       t.FloatPositionTopCenter,
				OnCursorChange: a.handleCursorChange,
				OnDismiss:      a.handleDismiss,
			},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: body,
		},
	}
}

// launchPanel explains how to open the palette and offers buttons that do it.
type launchPanel struct {
	demo *CommandPaletteDemo
}

func (l launchPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	openBtn := t.Button{
		ID:      "open-palette-btn",
		Label:   "Open palette",
		Variant: t.ButtonPrimary,
		OnPress: l.demo.open,
	}
	themesBtn := t.Button{
		ID:      "open-themes-btn",
		Label:   "Themes",
		OnPress: l.demo.openThemes,
	}
	focused := ctx.IsFocused(openBtn) || ctx.IsFocused(themesBtn)

	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "Launch", focused),
		Children: []t.Widget{
			openBtn,
			themesBtn,
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText("[$TextMuted]or press[/] [b $Accent]ctrl+p[/]", theme),
		},
	}
}

// lastCommandPanel shows the most recently run command.
type lastCommandPanel struct {
	demo *CommandPaletteDemo
}

func (l lastCommandPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	content := t.ParseMarkupToText("[$TextMuted]Nothing run yet. Press [b $Accent]ctrl+p[/] and pick a command.[/]", theme)
	for _, entry := range l.demo.activity.Get() {
		if entry.dismissed {
			continue
		}
		content = t.ParseMarkupToText(fmt.Sprintf("[$Success]▶[/] [b $Accent]%s[/]", entry.label), theme)
		break
	}

	return t.Column{
		Width:    t.Flex(1),
		Style:    demokit.PanelStyle(theme, "Last command", false),
		Children: []t.Widget{content},
	}
}

// activityPanel lists commands that have run, newest first.
type activityPanel struct {
	demo *CommandPaletteDemo
}

func (a activityPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	entries := a.demo.activity.Get()

	children := make([]t.Widget, 0, len(entries)+1)
	if len(entries) == 0 {
		children = append(children, t.Text{
			Spans: t.ParseMarkup("[$TextMuted]Commands you run show up here, newest first. Type to fuzzy-search, open [b]Themes[/] to preview themes as you move, or [b]Recent[/] for a nested list. Try [b]Toggle Sidebar[/].[/]", theme),
			Wrap:  t.WrapSoft,
		})
	}
	for _, entry := range entries {
		markup := fmt.Sprintf("[$Success]▶[/] %s", entry.label)
		if entry.dismissed {
			markup = fmt.Sprintf("[$TextMuted]✕ %s[/]", entry.label)
		}
		children = append(children, t.ParseMarkupToText(markup, theme))
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "Activity", false),
		Children: children,
	}
}

// statePanel shows the palette's live state. It reads the signals itself, so
// moving the cursor or typing only rebuilds this panel.
type statePanel struct {
	demo *CommandPaletteDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := s.demo
	visible := a.palette.Visible.Get()
	path := a.path.Get()
	preview := a.preview.Get()

	status := "[$TextMuted]closed[/]"
	level, query, highlighted := "[$TextMuted]—[/]", "[$TextMuted]—[/]", "[$TextMuted]—[/]"
	if visible {
		status = "[b $Success]open[/]"
		level = fmt.Sprintf("[b $Primary]%s[/]", lastSegment(path))
		if current := a.palette.CurrentLevel(); current != nil {
			if q := current.FilterState.Query.Get(); q != "" {
				query = fmt.Sprintf("[b $Warning]%q[/]", q)
			}
		}
		if _, ok := a.palette.CurrentItem(); ok {
			highlighted = fmt.Sprintf("[b $Info]%s[/]", truncate(preview, 16))
		} else {
			highlighted = "[$Warning]no match[/]"
		}
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Palette", status),
			demokit.StatRow(theme, "Level", level),
			demokit.StatRow(theme, "Query", query),
			demokit.StatRow(theme, "Highlighted", highlighted),
			demokit.StatRow(theme, "Commands run", fmt.Sprintf("[b $Secondary]%d[/]", a.runs.Get())),
			demokit.StatRow(theme, "Word wrap", onOff(a.wordWrap.Get())),
		},
	}
}

func onOff(on bool) string {
	if on {
		return "[b $Success]on[/]"
	}
	return "[$TextMuted]off[/]"
}

func lastSegment(path string) string {
	if i := strings.LastIndex(path, " › "); i >= 0 {
		return path[i+len(" › "):]
	}
	return path
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-6s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("ctrl+p", "$Accent", "open palette"),
			key("t", "$Accent", "open at Themes"),
			key("type", "$Warning", "fuzzy filter"),
			key("↑↓", "$Info", "move cursor"),
			key("enter", "$Success", "run / open level"),
			key("bksp", "$Warning", "back (empty query)"),
			key("esc", "$Error", "back / close"),
		},
	}
}
