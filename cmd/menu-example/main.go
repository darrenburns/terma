package main

import (
	"fmt"
	"log"
	"strings"

	t "github.com/darrenburns/terma"
)

// MenuDemo shows dropdown menus anchored to buttons in a menu bar, nested
// submenus, disabled items, dividers and a context menu positioned on screen.
//
//	f      - Open the File menu (anchored to its button)
//	e      - Open the Edit menu (anchored to its button)
//	m      - Open the context menu (centered on screen)
//	↑↓ jk  - Move through a menu
//	→ l    - Open a submenu
//	← h    - Close a submenu (or the menu)
//	enter  - Choose the highlighted item
//	escape - Dismiss the menu
type MenuDemo struct {
	file    *t.MenuState
	edit    *t.MenuState
	context *t.MenuState

	// open names the menu on screen ("file", "edit", "context"), or "" when none is.
	open t.Signal[string]
	// returnFocus is the widget that gets focus back when the menu closes.
	returnFocus string
	// focusedButton is the menu bar button that last had focus.
	focusedButton string

	activity   t.AnySignal[[]activityEntry]
	selections t.Signal[int]
	dismissals t.Signal[int]
	opened     t.Signal[int]
}

// activityEntry is one line in the activity log: a chosen item or a dismissal.
type activityEntry struct {
	menu      string
	path      []string
	dismissed bool
}

const (
	sidebarWidth = 32
	maxActivity  = 50
)

var menuTitles = map[string]string{
	"file":    "File",
	"edit":    "Edit",
	"context": "Context",
}

func NewMenuDemo() *MenuDemo {
	d := &MenuDemo{
		focusedButton: "file-btn",
		open:          t.NewSignal(""),
		activity:      t.NewAnySignal[[]activityEntry](nil),
		selections:    t.NewSignal(0),
		dismissals:    t.NewSignal(0),
		opened:        t.NewSignal(0),
	}
	d.file = t.NewMenuState([]t.MenuItem{
		{Label: "New", Shortcut: "Ctrl+N", Action: d.action("file", "New")},
		{Label: "Open", Shortcut: "Ctrl+O", Action: d.action("file", "Open")},
		{
			Label: "Open Recent",
			Children: []t.MenuItem{
				{Label: "alpha.txt", Action: d.action("file", "Open Recent", "alpha.txt")},
				{Label: "bravo.txt", Action: d.action("file", "Open Recent", "bravo.txt")},
			},
		},
		{Divider: "Settings"},
		{
			Label: "Settings",
			Children: []t.MenuItem{
				{Label: "Editor", Action: d.action("file", "Settings", "Editor")},
				{
					Label: "Theme",
					Children: []t.MenuItem{
						{Label: "Light", Action: d.themeAction("Light", t.ThemeNameRosePineDawn)},
						{Label: "Dark", Action: d.themeAction("Dark", t.ThemeNameRosePine)},
					},
				},
			},
		},
		{Label: "Exit", Action: d.action("file", "Exit")},
	})
	d.edit = t.NewMenuState([]t.MenuItem{
		{Label: "Undo", Shortcut: "Ctrl+Z", Action: d.action("edit", "Undo")},
		{Label: "Redo", Shortcut: "Ctrl+Y", Disabled: true},
		{}, // A plain separator.
		{Label: "Cut", Shortcut: "Ctrl+X", Action: d.action("edit", "Cut")},
		{Label: "Copy", Shortcut: "Ctrl+C", Action: d.action("edit", "Copy")},
		{Label: "Paste", Shortcut: "Ctrl+V", Action: d.action("edit", "Paste")},
		{Divider: "Search"},
		{Label: "Find", Shortcut: "Ctrl+F", Action: d.action("edit", "Find")},
		{Label: "Replace", Shortcut: "Ctrl+H", Action: d.action("edit", "Replace")},
	})
	d.context = t.NewMenuState([]t.MenuItem{
		{Label: "Clear activity", Action: func() {
			d.activity.Set(nil)
			d.action("context", "Clear activity")()
		}},
		{Label: "Reset counters", Action: func() {
			d.action("context", "Reset counters")()
			d.selections.Set(0)
			d.dismissals.Set(0)
			d.opened.Set(0)
		}},
		{},
		{Label: "Export log…", Disabled: true},
	})
	return d
}

// action returns a menu item callback that records the item's path.
func (d *MenuDemo) action(menu string, path ...string) func() {
	return func() { d.record(activityEntry{menu: menu, path: path}) }
}

// themeAction switches between a light and a dark theme.
func (d *MenuDemo) themeAction(label, themeName string) func() {
	return func() {
		t.SetTheme(themeName)
		d.record(activityEntry{menu: "file", path: []string{"Settings", "Theme", label}})
	}
}

func (d *MenuDemo) record(entry activityEntry) {
	if entry.dismissed {
		d.dismissals.Update(func(n int) int { return n + 1 })
	} else {
		d.selections.Update(func(n int) int { return n + 1 })
	}
	d.activity.Update(func(entries []activityEntry) []activityEntry {
		entries = append([]activityEntry{entry}, entries...)
		if len(entries) > maxActivity {
			entries = entries[:maxActivity]
		}
		return entries
	})
}

func (d *MenuDemo) state(name string) *t.MenuState {
	switch name {
	case "file":
		return d.file
	case "edit":
		return d.edit
	default:
		return d.context
	}
}

// toggle opens the named menu, or closes it if it is already open. Opening a
// menu while another is showing switches straight to it, like a menu bar.
func (d *MenuDemo) toggle(name string) {
	current := d.open.Peek()
	if current == name {
		d.close()
		return
	}
	if current == "" {
		switch name {
		case "file", "edit":
			d.returnFocus = name + "-btn"
		default:
			d.returnFocus = d.focusedButton
		}
	} else {
		d.state(current).CloseSubmenu()
	}
	d.opened.Update(func(n int) int { return n + 1 })
	d.open.Set(name)
	t.RequestFocus(name + "-menu")
}

func (d *MenuDemo) close() {
	if current := d.open.Peek(); current != "" {
		d.state(current).CloseSubmenu()
	}
	d.open.Set("")
	t.RequestFocus(d.returnFocus)
}

func (d *MenuDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "f", Name: "File menu", Action: func() { d.toggle("file") }},
		{Key: "e", Name: "Edit menu", Action: func() { d.toggle("edit") }},
		{Key: "m", Name: "Context menu", Action: func() { d.toggle("context") }},
	}
}

func (d *MenuDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "menu-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			header{},
			// Menus float above everything, so where they sit in the tree
			// doesn't matter. Here they take no space in the layout.
			openMenu{demo: d},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: []t.Widget{
				t.Column{
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						menuBar{demo: d},
						lastSelectionPanel{demo: d},
						fill(activityPanel{demo: d}),
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{demo: d},
						keysPanel{},
					},
				},
			},
		},
	}
}

// header is the title bar across the top of the screen.
type header struct{}

func (header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ Menu Playground[/]  [$TextMuted]Dropdowns, submenus and context menus[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", theme.Name), theme),
		},
	}
}

// fill gives a component the remaining space in its Column. Rows and Columns
// read Flex from their direct children, so a component's own Flex(1) needs a
// plain wrapper to take effect.
func fill(child t.Widget) t.Widget {
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{child},
	}
}

// panelStyle is the bordered look shared by every panel in the demo.
func panelStyle(theme t.ThemeData, title string, focused bool) t.Style {
	color := theme.Border
	titleColor := "$TextMuted"
	if focused {
		color = theme.FocusRing
		titleColor = "$FocusRing"
	}
	return t.Style{
		BackgroundColor: theme.Background,
		Border:          t.RoundedBorder(color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] %s [/]", titleColor, title))),
		Padding:         t.EdgeInsetsXY(1, 0),
	}
}

// menuBar holds the buttons the dropdowns hang from, and the open menu itself.
type menuBar struct {
	demo *MenuDemo
}

func (m menuBar) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := m.demo
	open := d.open.Get()

	button := func(name, id string) t.Button {
		variant := t.ButtonDefault
		if open == name {
			variant = t.ButtonPrimary
		}
		return t.Button{
			ID:      id,
			Label:   menuTitles[name] + " ▾",
			Variant: variant,
			OnPress: func() { d.toggle(name) },
		}
	}
	fileBtn := button("file", "file-btn")
	editBtn := button("edit", "edit-btn")
	focused := false
	for _, b := range []t.Button{fileBtn, editBtn} {
		if ctx.IsFocused(b) {
			focused = true
			d.focusedButton = b.ID
		}
	}

	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   panelStyle(theme, "Menu bar", focused),
		Children: []t.Widget{
			fileBtn,
			editBtn,
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText("[b $Accent]f[/] [b $Accent]e[/] [$TextMuted]or[/] [b $Accent]m[/] [$TextMuted]opens a menu[/]", theme),
		},
	}
}

// openMenu builds the floating Menu that is currently open, if any. File and
// Edit drop down from their buttons; the context menu is centered on screen.
type openMenu struct {
	demo *MenuDemo
}

func (o openMenu) Build(ctx t.BuildContext) t.Widget {
	d := o.demo
	name := d.open.Get()
	if name == "" {
		return t.EmptyWidget{}
	}
	menu := t.Menu{
		ID:    name + "-menu",
		State: d.state(name),
		OnSelect: func(item t.MenuItem) {
			d.close()
			if item.Action != nil {
				item.Action()
			}
		},
		OnDismiss: func() {
			d.close()
			d.record(activityEntry{menu: name, dismissed: true})
		},
	}
	switch name {
	case "file":
		menu.AnchorID = "file-btn"
	case "edit":
		menu.AnchorID = "edit-btn"
	default:
		menu.Position = t.FloatPositionCenter
		menu.Style = t.Style{Width: t.Cells(24)}
	}
	return menu
}

// activityPanel lists what has been chosen from the menus, newest first.
type activityPanel struct {
	demo *MenuDemo
}

func (a activityPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	entries := a.demo.activity.Get()

	children := make([]t.Widget, 0, len(entries)+1)
	if len(entries) == 0 {
		children = append(children, t.Text{
			Spans: t.ParseMarkup("[$TextMuted]Choices and dismissals show up here, newest first. Press [b $Accent]f[/] for the File menu, [b $Accent]e[/] for Edit or [b $Accent]m[/] for the context menu, or tab to a button and press [b]enter[/].[/]", theme),
			Wrap:  t.WrapSoft,
		})
	}
	for _, entry := range entries {
		children = append(children, t.ParseMarkupToText(entryMarkup(entry), theme))
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    panelStyle(theme, "Activity", false),
		Children: children,
	}
}

func entryMarkup(entry activityEntry) string {
	title := menuTitles[entry.menu]
	if entry.dismissed {
		return fmt.Sprintf("[$TextMuted]✕ %s menu dismissed[/]", title)
	}
	return fmt.Sprintf("[$Success]✓[/] [b $Accent]%s[/] [$TextMuted]›[/] %s", title, pathMarkup(entry.path))
}

func pathMarkup(path []string) string {
	return strings.Join(path, " [$TextMuted]›[/] ")
}

// statePanel shows the live state of the menus. It reads the signals itself,
// so opening a menu or choosing an item only rebuilds this panel.
type statePanel struct {
	demo *MenuDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	open := "[$TextMuted]none[/]"
	if name := d.open.Get(); name != "" {
		open = fmt.Sprintf("[b $Primary]%s[/]", menuTitles[name])
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "State", false),
		Children: []t.Widget{
			statRow(theme, "Open menu", open),
			statRow(theme, "Selections", fmt.Sprintf("[b $Success]%d[/]", d.selections.Get())),
			statRow(theme, "Dismissals", fmt.Sprintf("[b $Warning]%d[/]", d.dismissals.Get())),
			statRow(theme, "Times opened", fmt.Sprintf("[b $Info]%d[/]", d.opened.Get())),
		},
	}
}

func statRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}

// lastSelectionPanel shows the most recently chosen item and where it came from.
type lastSelectionPanel struct {
	demo *MenuDemo
}

func (l lastSelectionPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	content := t.ParseMarkupToText("[$TextMuted]Nothing chosen yet. Press [b $Accent]f[/] to open the File menu.[/]", theme)
	for _, entry := range l.demo.activity.Get() {
		if entry.dismissed {
			continue
		}
		content = t.ParseMarkupToText(fmt.Sprintf("[b $Primary]%s[/] [$TextMuted]›[/] [b $Accent]%s[/]", menuTitles[entry.menu], strings.Join(entry.path, " › ")), theme)
		break
	}

	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Last selection", false),
		Children: []t.Widget{content},
	}
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
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("f e m", "$Accent", "file / edit / context"),
			key("↑↓ jk", "$Info", "move in a menu"),
			key("→ l", "$Success", "open submenu"),
			key("← h", "$Warning", "close submenu"),
			key("enter", "$Success", "choose item"),
			key("esc", "$Error", "dismiss menu"),
			key("tab", "$Info", "switch button"),
		},
	}
}

func main() {
	app := NewMenuDemo()
	t.RequestFocus("file-btn")
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
