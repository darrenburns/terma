package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

// App demonstrates Checkbox and CheckboxState: a registration form whose
// submit button is disabled until the terms are accepted, settings with an
// "all" checkbox kept in sync through OnChange and SetChecked, and disabled
// checkboxes for comparison.
//
// Keys:
//
//	tab / shift+tab - move between checkboxes
//	space / enter   - toggle the focused checkbox (or press the button)
//	a / n           - turn every setting on / off
//	r               - reset everything to its defaults
//	t               - cycle theme
type App struct {
	// Registration form
	terms      *t.CheckboxState
	newsletter *t.CheckboxState
	marketing  *t.CheckboxState

	// Settings, plus a checkbox that sets all of them at once
	allSettings   *t.CheckboxState
	darkMode      *t.CheckboxState
	notifications *t.CheckboxState
	autoSave      *t.CheckboxState

	activity   t.AnySignal[[]string] // newest first
	themeIndex t.Signal[int]
}

// option is a checkbox's ID, label, short name and state.
type option struct {
	id    string
	label string
	name  string // used in the State and Activity panels
	state *t.CheckboxState
}

var themeNames = []string{
	t.ThemeNameRosePine,
	t.ThemeNameDracula,
	t.ThemeNameTokyoNight,
	t.ThemeNameCatppuccin,
	t.ThemeNameGruvbox,
	t.ThemeNameNord,
	t.ThemeNameSolarized,
	t.ThemeNameKanagawa,
	t.ThemeNameMonokai,
}

const (
	sidebarWidth    = 32
	maxActivityRows = 8
)

func NewApp() *App {
	a := &App{
		terms:         t.NewCheckboxState(false),
		newsletter:    t.NewCheckboxState(false),
		marketing:     t.NewCheckboxState(false),
		allSettings:   t.NewCheckboxState(false),
		darkMode:      t.NewCheckboxState(false),
		notifications: t.NewCheckboxState(false),
		autoSave:      t.NewCheckboxState(false),
		activity:      t.NewAnySignal[[]string](nil),
		themeIndex:    t.NewSignal(0),
	}
	a.applyDefaults()
	return a
}

func (a *App) formOptions() []option {
	return []option{
		{"terms", "I accept the Terms of Service (required)", "Terms", a.terms},
		{"newsletter", "Subscribe to the newsletter", "Newsletter", a.newsletter},
		{"marketing", "Receive marketing emails", "Marketing", a.marketing},
	}
}

func (a *App) settingOptions() []option {
	return []option{
		{"darkmode", "Dark mode", "Dark mode", a.darkMode},
		{"notifications", "Notifications", "Notifications", a.notifications},
		{"autosave", "Auto-save documents", "Auto-save", a.autoSave},
	}
}

// logActivity records an event at the top of the activity panel.
func (a *App) logActivity(markup string) {
	a.activity.Update(func(events []string) []string {
		events = append([]string{markup}, events...)
		return events[:min(len(events), maxActivityRows)]
	})
}

func (a *App) logToggle(name string, checked bool) {
	if checked {
		a.logActivity(fmt.Sprintf("[$Success]☑[/] [$Text]%s on[/]", name))
	} else {
		a.logActivity(fmt.Sprintf("[$TextMuted]☐ %s off[/]", name))
	}
}

// syncAllSettings checks the "all" box exactly when every setting is on.
func (a *App) syncAllSettings() {
	all := true
	for _, o := range a.settingOptions() {
		all = all && o.state.IsChecked()
	}
	a.allSettings.SetChecked(all)
}

// setAllSettings turns every setting on or off.
func (a *App) setAllSettings(checked bool) {
	for _, o := range a.settingOptions() {
		o.state.SetChecked(checked)
	}
	a.allSettings.SetChecked(checked)
}

// applyDefaults sets every checkbox to its starting value.
func (a *App) applyDefaults() {
	a.terms.SetChecked(false)
	a.newsletter.SetChecked(true)
	a.marketing.SetChecked(false)
	a.darkMode.SetChecked(true)
	a.notifications.SetChecked(true)
	a.autoSave.SetChecked(false)
	a.syncAllSettings()
}

func (a *App) reset() {
	a.applyDefaults()
	a.logActivity("[$Warning]Reset to defaults[/]")
}

func (a *App) cycleTheme() {
	a.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "a", Name: "All on", Action: func() {
			a.setAllSettings(true)
			a.logActivity("[$Success]All settings on[/]")
		}},
		{Key: "n", Name: "All off", Action: func() {
			a.setAllSettings(false)
			a.logActivity("[$TextMuted]All settings off[/]")
		}},
		{Key: "r", Name: "Reset", Action: a.reset},
		{Key: "t", Name: "Theme", Action: a.cycleTheme},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID:    "checkbox-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top: []t.Widget{
			header{themeName: themeNames[a.themeIndex.Get()]},
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
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				t.Column{
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						formPanel{app: a},
						settingsPanel{app: a},
						disabledPanel{},
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{app: a},
						fill(activityPanel{app: a}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// header is the title bar across the top of the screen.
type header struct {
	themeName string
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ Checkbox Playground[/]  [$TextMuted]Toggles backed by CheckboxState[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", h.themeName), theme),
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

// checkbox builds a checkbox that logs each toggle, then calls onChange.
func (a *App) checkbox(theme t.ThemeData, o option, onChange func(bool)) *t.Checkbox {
	return &t.Checkbox{
		ID:    o.id,
		State: o.state,
		Label: o.label,
		Style: t.Style{BackgroundColor: theme.Background},
		OnChange: func(checked bool) {
			a.logToggle(o.name, checked)
			if onChange != nil {
				onChange(checked)
			}
		},
	}
}

// anyFocused reports whether one of the widgets has focus. Panels use it to
// light up their border.
func anyFocused(ctx t.BuildContext, widgets []t.Widget) bool {
	for _, w := range widgets {
		if ctx.IsFocused(w) {
			return true
		}
	}
	return false
}

// formPanel is the registration form. Submit stays disabled until the terms
// are accepted.
type formPanel struct {
	app *App
}

func (f formPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := f.app
	accepted := a.terms.Checked.Get()

	var children []t.Widget
	for _, o := range a.formOptions() {
		children = append(children, a.checkbox(theme, o, nil))
	}
	// A Button's Style applies when it isn't focused and its Variant colours
	// when it is, so this pairing gives a clear focus highlight.
	submit := t.Button{
		ID:      "submit",
		Label:   "Submit registration",
		Variant: t.ButtonSuccess,
		Style:   t.Style{ForegroundColor: theme.Text, BackgroundColor: theme.Surface},
		OnPress: func() { a.logActivity("[b $Success]Registration submitted[/]") },
	}
	focused := ctx.IsFocused(submit) || anyFocused(ctx, children)

	hint := "[$Success]Ready to submit[/]"
	if !accepted {
		hint = "[$TextMuted]Accept the terms to enable[/]"
	}
	// Padding rather than Margin: a Margin on a Row or Column child is
	// currently applied twice and pushes the row out of the panel.
	children = append(children, t.Row{
		Spacing: 2,
		Style:   t.Style{Padding: t.EdgeInsets{Top: 1}},
		Children: []t.Widget{
			t.DisabledWhen(!accepted, submit),
			t.ParseMarkupToText(hint, theme),
		},
	})

	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Registration", focused),
		Children: children,
	}
}

// settingsPanel holds the settings and an "all" checkbox kept in sync with
// them: toggling it sets every setting, and toggling a setting updates it.
type settingsPanel struct {
	app *App
}

func (s settingsPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := s.app

	all := a.checkbox(theme, option{"all-settings", "Enable all settings", "All settings", a.allSettings}, a.setAllSettings)
	checkboxes := []t.Widget{all}
	children := []t.Widget{all}
	for _, o := range a.settingOptions() {
		cb := a.checkbox(theme, o, func(bool) { a.syncAllSettings() })
		checkboxes = append(checkboxes, cb)
		// Indent with a padded Row so the focus highlight covers only the box.
		children = append(children, t.Row{
			Style:    t.Style{Padding: t.EdgeInsets{Left: 2}},
			Children: []t.Widget{cb},
		})
	}

	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Settings", anyFocused(ctx, checkboxes)),
		Children: children,
	}
}

// disabledPanel shows how disabled checkboxes look. They can't take focus.
type disabledPanel struct{}

func (disabledPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	style := t.Style{BackgroundColor: theme.Background}
	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Disabled", false),
		Children: []t.Widget{
			t.DisabledWhen(true, &t.Checkbox{
				ID:    "disabled-unchecked",
				State: t.NewCheckboxState(false),
				Label: "Disabled, unchecked",
				Style: style,
			}),
			t.DisabledWhen(true, &t.Checkbox{
				ID:    "disabled-checked",
				State: t.NewCheckboxState(true),
				Label: "Disabled, checked",
				Style: style,
			}),
		},
	}
}

// statePanel shows every checkbox's value. It subscribes to the states
// itself, so a toggle only rebuilds this panel and the one holding the box.
type statePanel struct {
	app *App
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := s.app

	onOff := func(state *t.CheckboxState) string {
		if state.Checked.Get() {
			return "[b $Success]on[/]"
		}
		return "[$TextMuted]off[/]"
	}

	options := append(a.formOptions(), a.settingOptions()...)
	var rows []t.Widget
	checked := 0
	for _, o := range options {
		rows = append(rows, statRow(theme, o.name, onOff(o.state)))
		if o.state.Checked.Get() {
			checked++
		}
	}
	form := "[b $Warning]needs terms[/]"
	if a.terms.Checked.Get() {
		form = "[b $Success]ready[/]"
	}
	rows = append(rows,
		statRow(theme, "Checked", fmt.Sprintf("[b $Primary]%d of %d[/]", checked, len(options))),
		statRow(theme, "Form", form),
	)

	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "State", false),
		Children: rows,
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

// activityPanel lists recent changes, newest first, as reported by the
// checkboxes' OnChange callbacks.
type activityPanel struct {
	app *App
}

func (p activityPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	events := p.app.activity.Get()

	var rows []t.Widget
	if len(events) == 0 {
		rows = append(rows, t.ParseMarkupToText("[$TextMuted]Toggle something…[/]", theme))
	}
	for _, e := range events {
		rows = append(rows, t.ParseMarkupToText(e, theme))
	}
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    panelStyle(theme, "Activity", false),
		Children: rows,
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-9s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("tab ⇧tab", "$Info", "move focus"),
			key("space ↵", "$Success", "toggle / press"),
			key("a n", "$Accent", "settings on / off"),
			key("r", "$Warning", "reset to defaults"),
			key("t", "$Primary", "cycle theme"),
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	t.RequestFocus("terms")
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
