// Package dialogdemo demonstrates modal Dialogs: button variants, rich
// content, forms, and the different ways a dialog can be closed.
package dialogdemo

import (
	"fmt"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "dialog",
	Title:       "Dialog",
	Description: "Modal dialogs with button variants, rich content and a form",
}

// DialogDemo is a gallery of modal Dialogs. Each launcher button opens a
// different style of dialog, and the sidebar records how each one was closed:
// by pressing one of its buttons, or by dismissing it (Esc or a click on the
// backdrop, which calls OnDismiss).
//
//	tab / shift+tab - Move between launcher buttons (or within an open dialog)
//	enter / space   - Press the focused button
//	1-8             - Open a dialog directly
//	escape          - Dismiss the open dialog
//	t               - Cycle theme
type DialogDemo struct {
	active    t.Signal[string] // ID of the open dialog ("" = none)
	opened    t.Signal[int]
	actioned  t.Signal[int]
	cancelled t.Signal[int]
	dismissed t.Signal[int]
	history   t.AnySignal[[]historyEntry]

	// Feedback form state, kept here so it survives the dialog closing.
	category      *t.ListState[string]
	summary       *t.TextInputState
	focusFormNext bool // focus the form's first field on its next build
}

const (
	sidebarWidth = 32
	historyLimit = 20
)

// dialogSpec describes one launcher: the dialog it opens and how it is listed.
type dialogSpec struct {
	id    string
	key   string
	label string
	desc  string
}

// section groups launchers under a panel title.
type section struct {
	title   string
	dialogs []dialogSpec
}

var sections = []section{
	{"Informational", []dialogSpec{
		{"info", "1", "Info", "One button, info variant"},
		{"success", "2", "Success", "Acknowledge a finished task"},
		{"warning", "3", "Warning", "Dismiss or act on a warning"},
	}},
	{"Actions", []dialogSpec{
		{"confirm", "4", "Confirm Action", "Cancel or confirm"},
		{"delete", "5", "Delete Item", "Destructive action, error variant"},
		{"unsaved", "6", "Unsaved Changes", "Three buttons: discard, cancel, save"},
	}},
	{"Content", []dialogSpec{
		{"rich", "7", "Rich Content", "Markup and a checklist in the body"},
		{"form", "8", "Feedback Form", "A list and text input inside a dialog"},
	}},
}

var feedbackCategories = []string{"Bug Report", "Feature Request", "Question", "Other"}

// New creates the demo.
func New() demokit.Demo {
	return &DialogDemo{
		active:    t.NewSignal(""),
		opened:    t.NewSignal(0),
		actioned:  t.NewSignal(0),
		cancelled: t.NewSignal(0),
		dismissed: t.NewSignal(0),
		history:   t.NewAnySignal[[]historyEntry](nil),
		category:  t.NewListState(append([]string(nil), feedbackCategories...)),
		summary:   t.NewTextInputState(""),
	}
}

func (d *DialogDemo) InitialFocus() string { return launcherID(sections[0].dialogs[0].id) }

// outcome is how a dialog was closed.
type outcome int

const (
	outcomeAction  outcome = iota // a button that does something
	outcomeCancel                 // a button that backs out
	outcomeDismiss                // Esc or a click outside (OnDismiss)
)

type historyEntry struct {
	outcome outcome
	message string
}

func (d *DialogDemo) open(id string) {
	if d.active.Peek() != "" {
		return
	}
	if id == "form" {
		d.focusFormNext = true
	}
	d.opened.Update(func(n int) int { return n + 1 })
	d.active.Set(id)
}

// resolve closes the open dialog and records how it was closed.
func (d *DialogDemo) resolve(kind outcome, message string) {
	counter := map[outcome]t.Signal[int]{
		outcomeAction:  d.actioned,
		outcomeCancel:  d.cancelled,
		outcomeDismiss: d.dismissed,
	}[kind]
	counter.Update(func(n int) int { return n + 1 })
	d.history.Update(func(entries []historyEntry) []historyEntry {
		entries = append([]historyEntry{{kind, message}}, entries...)
		if len(entries) > historyLimit {
			entries = entries[:historyLimit]
		}
		return entries
	})
	// Hand focus back to the launcher for this dialog, which may not be the
	// button that had focus before if the dialog was opened with a number key.
	t.RequestFocus(launcherID(d.active.Peek()))
	d.active.Set("")
}

// button returns a dialog button that closes the dialog with the given outcome.
func (d *DialogDemo) button(label string, variant t.ButtonVariant, kind outcome, message string) t.Button {
	return t.Button{Label: label, Variant: variant, OnPress: func() { d.resolve(kind, message) }}
}

// dismissWith returns an OnDismiss handler for the named dialog.
func (d *DialogDemo) dismissWith(title string) func() {
	return func() { d.resolve(outcomeDismiss, title+" dismissed") }
}

func (d *DialogDemo) Keybinds() []t.Keybind {
	var binds []t.Keybind
	for _, s := range sections {
		for _, spec := range s.dialogs {
			binds = append(binds, t.Keybind{
				Key:    spec.key,
				Name:   spec.label,
				Action: func() { d.open(spec.id) },
				Hidden: true,
			})
		}
	}
	return append(binds, t.Keybind{Key: "t", Name: "Theme", Action: demokit.NextTheme})
}

func (d *DialogDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	var panels []t.Widget
	for _, s := range sections {
		panels = append(panels, launcherPanel{demo: d, section: s})
	}
	panels = append(panels, demokit.Fill(aboutPanel{}))

	return t.Dock{
		ID:    "dialog-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top: []t.Widget{
			demokit.Header{Title: "Dialog Gallery", Tagline: "Open, act on and dismiss modals"},
			// Dialogs render as overlays, so the layer can sit anywhere in the
			// tree. Keeping it in its own component means opening and closing
			// a dialog only rebuilds the layer.
			dialogLayer{demo: d},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				t.Column{
					Width:    t.Flex(1),
					Height:   t.Flex(1),
					Spacing:  1,
					Children: panels,
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statsPanel{demo: d},
						demokit.Fill(historyPanel{demo: d}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// focusedID returns the ID of the focused widget, subscribing to focus changes.
func focusedID(ctx t.BuildContext) string {
	if w, ok := ctx.Focused().(t.Identifiable); ok {
		return w.WidgetID()
	}
	return ""
}

func launcherID(id string) string { return "open-" + id }

// launcherPanel is one titled group of launcher buttons. Its border lights up
// while one of its buttons has focus.
type launcherPanel struct {
	demo    *DialogDemo
	section section
}

func (p launcherPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	focused := focusedID(ctx)

	rows := make([]t.Widget, 0, len(p.section.dialogs))
	containsFocus := false
	for _, spec := range p.section.dialogs {
		id := launcherID(spec.id)
		containsFocus = containsFocus || id == focused
		rows = append(rows, t.Row{
			Width:   t.Flex(1),
			Spacing: 2,
			Children: []t.Widget{
				t.ParseMarkupToText(fmt.Sprintf("[b $Accent]%s[/]", spec.key), theme),
				t.Button{
					ID:      id,
					Label:   spec.label,
					OnPress: func() { p.demo.open(spec.id) },
					Style:   t.Style{Width: t.Cells(19)},
				},
				t.Text{
					Content: spec.desc,
					Width:   t.Flex(1),
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
			},
		})
	}

	return t.Column{
		Width:    t.Flex(1),
		Style:    demokit.PanelStyle(theme, p.section.title, containsFocus),
		Children: rows,
	}
}

// aboutPanel explains the behaviour every Dialog gets for free.
type aboutPanel struct{}

func (aboutPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	point := func(markup string) t.Widget {
		return t.Text{
			Spans: t.ParseMarkup("[$Accent]•[/] "+markup, theme),
			Wrap:  t.WrapSoft,
			Style: t.Style{ForegroundColor: theme.TextMuted},
		}
	}
	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, "Every Dialog", false),
		Children: []t.Widget{
			point("Opens centred over a backdrop, wherever it sits in the tree."),
			point("Its [b $Text]first button[/] takes focus when it opens."),
			point("[b $Text]Tab[/] cycles inside the dialog; focus can't leave it."),
			point("[b $Text]Esc[/] or a click on the backdrop calls [b $Text]OnDismiss[/]."),
		},
	}
}

// statsPanel shows which dialog is open and how dialogs have been closed.
type statsPanel struct {
	demo *DialogDemo
}

func (s statsPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	open := "[$TextMuted]none[/]"
	if spec, ok := findDialog(d.active.Get()); ok {
		open = fmt.Sprintf("[b $Primary]%s[/]", spec.label)
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Open", open),
			demokit.StatRow(theme, "Opened", fmt.Sprintf("[b $Text]%d[/]", d.opened.Get())),
			demokit.StatRow(theme, "Actioned", fmt.Sprintf("[b $Success]%d[/]", d.actioned.Get())),
			demokit.StatRow(theme, "Cancelled", fmt.Sprintf("[b $Warning]%d[/]", d.cancelled.Get())),
			demokit.StatRow(theme, "Dismissed", fmt.Sprintf("[b $Info]%d[/]", d.dismissed.Get())),
		},
	}
}

// historyPanel lists how recent dialogs were closed, newest first.
type historyPanel struct {
	demo *DialogDemo
}

func (h historyPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	entries := h.demo.history.Get()

	var lines []string
	for _, e := range entries {
		icon, colour := "✓", "$Success"
		switch e.outcome {
		case outcomeCancel:
			icon, colour = "✗", "$Warning"
		case outcomeDismiss:
			icon, colour = "⎋", "$Info"
		}
		lines = append(lines, fmt.Sprintf("[b %s]%s[/] %s", colour, icon, e.message))
	}

	content := t.Text{
		Spans: t.ParseMarkup("[$TextMuted]Nothing yet. Open a dialog and close it with a button or [b $Info]esc[/].[/]", theme),
		Wrap:  t.WrapSoft,
	}
	if len(lines) > 0 {
		content = t.Text{Spans: t.ParseMarkup(strings.Join(lines, "\n"), theme), Wrap: t.WrapSoft}
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "History", false),
		Children: []t.Widget{content},
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
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("tab ⇧tab", "$Info", "move focus"),
			key("enter", "$Success", "press button"),
			key("1-8", "$Accent", "open a dialog"),
			key("esc", "$Info", "dismiss dialog"),
			key("t", "$Secondary", "cycle theme"),
		},
	}
}

func findDialog(id string) (dialogSpec, bool) {
	for _, s := range sections {
		for _, spec := range s.dialogs {
			if spec.id == id {
				return spec, true
			}
		}
	}
	return dialogSpec{}, false
}

// dialogLayer declares every dialog. Only the active one is visible.
type dialogLayer struct {
	demo *DialogDemo
}

func (l dialogLayer) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := l.demo
	active := d.active.Get()

	wrap := func(text string) t.Widget {
		return t.Text{Content: text, Wrap: t.WrapSoft}
	}

	return t.Column{
		Children: []t.Widget{
			t.Dialog{
				ID:      "dlg-info",
				Visible: active == "info",
				Title:   "Information",
				Content: wrap("This operation completed successfully. No further action is required."),
				Buttons: []t.Button{
					d.button("OK", t.ButtonInfo, outcomeAction, "Info acknowledged"),
				},
				OnDismiss: d.dismissWith("Info"),
			},
			t.Dialog{
				ID:      "dlg-success",
				Visible: active == "success",
				Title:   "Success",
				Content: wrap("Your changes have been saved."),
				Buttons: []t.Button{
					d.button("Great!", t.ButtonSuccess, outcomeAction, "Success acknowledged"),
				},
				OnDismiss: d.dismissWith("Success"),
			},
			t.Dialog{
				ID:      "dlg-warning",
				Visible: active == "warning",
				Title:   "Warning",
				Content: wrap("Your session will expire in 5 minutes. Save your work to avoid losing changes."),
				Buttons: []t.Button{
					d.button("Dismiss", t.ButtonDefault, outcomeCancel, "Warning ignored"),
					d.button("Save Now", t.ButtonWarning, outcomeAction, "Work saved"),
				},
				OnDismiss: d.dismissWith("Warning"),
			},
			t.Dialog{
				ID:      "dlg-confirm",
				Visible: active == "confirm",
				Title:   "Confirm Action",
				Content: wrap("Are you sure you want to proceed? This will apply the pending changes."),
				Buttons: []t.Button{
					d.button("Cancel", t.ButtonDefault, outcomeCancel, "Action cancelled"),
					d.button("Confirm", t.ButtonPrimary, outcomeAction, "Action confirmed"),
				},
				OnDismiss: d.dismissWith("Confirm"),
			},
			t.Dialog{
				ID:      "dlg-delete",
				Visible: active == "delete",
				Title:   "Delete Item",
				Content: t.Column{
					Spacing: 1,
					Children: []t.Widget{
						wrap("Are you sure you want to delete this item?"),
						t.Text{
							Content: "This action cannot be undone.",
							Wrap:    t.WrapSoft,
							Style:   t.Style{ForegroundColor: theme.TextMuted},
						},
					},
				},
				Buttons: []t.Button{
					d.button("Cancel", t.ButtonDefault, outcomeCancel, "Delete cancelled"),
					d.button("Delete", t.ButtonError, outcomeAction, "Item deleted"),
				},
				OnDismiss: d.dismissWith("Delete"),
			},
			t.Dialog{
				ID:      "dlg-unsaved",
				Visible: active == "unsaved",
				Title:   "Unsaved Changes",
				Content: wrap("You have unsaved changes. What would you like to do?"),
				Buttons: []t.Button{
					d.button("Discard", t.ButtonError, outcomeAction, "Changes discarded"),
					d.button("Cancel", t.ButtonDefault, outcomeCancel, "Resumed editing"),
					d.button("Save", t.ButtonSuccess, outcomeAction, "Changes saved"),
				},
				OnDismiss: d.dismissWith("Unsaved changes"),
			},
			t.Dialog{
				ID:      "dlg-rich",
				Visible: active == "rich",
				Title:   "Release Notes",
				Content: t.Column{
					Spacing: 1,
					Children: []t.Widget{
						t.ParseMarkupToText("[b]Version 2.0[/] is now available!", theme),
						t.Text{Content: "What's new:", Style: t.Style{ForegroundColor: theme.TextMuted}},
						t.Column{
							Children: []t.Widget{
								t.ParseMarkupToText("  [b $Success]✓[/] New dialog widget", theme),
								t.ParseMarkupToText("  [b $Success]✓[/] Button variants", theme),
								t.ParseMarkupToText("  [b $Success]✓[/] Improved focus management", theme),
							},
						},
					},
				},
				Buttons: []t.Button{
					d.button("Later", t.ButtonDefault, outcomeCancel, "Update postponed"),
					d.button("Update Now", t.ButtonAccent, outcomeAction, "Update started"),
				},
				OnDismiss: d.dismissWith("Release notes"),
			},
			t.Dialog{
				ID:      "dlg-form",
				Visible: active == "form",
				Title:   "Submit Feedback",
				Content: feedbackForm{demo: d},
				Style:   t.Style{Width: t.Cells(50)},
				Buttons: []t.Button{
					d.button("Cancel", t.ButtonDefault, outcomeCancel, "Feedback cancelled"),
					{Label: "Submit", Variant: t.ButtonPrimary, OnPress: d.submitFeedback},
				},
				OnDismiss: d.dismissWith("Feedback"),
			},
		},
	}
}

func (d *DialogDemo) submitFeedback() {
	category, _ := d.category.SelectedItem()
	summary := strings.TrimSpace(d.summary.GetText())
	message := "Sent " + strings.ToLower(category)
	if summary != "" {
		message += fmt.Sprintf(": %q", summary)
	}
	d.summary.SetText("")
	d.resolve(outcomeAction, message)
}

// feedbackForm is the body of the feedback dialog: a category list and a
// one-line summary. It shows that dialogs can host any focusable widget.
type feedbackForm struct {
	demo *DialogDemo
}

func (f feedbackForm) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := f.demo

	// Dialogs focus their first button when they open. The form is built
	// after the dialog, so a request made here wins and lands on the list.
	if d.focusFormNext {
		d.focusFormNext = false
		ctx.RequestFocus("feedback-category")
	}

	category := t.List[string]{
		ID:    "feedback-category",
		State: d.category,
		Style: t.Style{Width: t.Flex(1)},
	}
	summary := t.TextInput{
		ID:          "feedback-summary",
		State:       d.summary,
		Placeholder: "One line about it…",
		Width:       t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		OnSubmit: func(string) { d.submitFeedback() },
	}

	// The cursor row is the chosen category, so mark it like a radio button
	// that stays visible after focus moves on to the summary.
	categoryFocused := ctx.IsFocused(category)
	category.RenderItem = func(item string, active, _ bool) t.Widget {
		markup := "[$TextMuted]○[/] " + item
		style := t.Style{Width: t.Flex(1)}
		if active {
			markup = "[b $Primary]●[/] [b]" + item + "[/]"
			if categoryFocused {
				style.BackgroundColor = theme.SurfaceHover
			}
		}
		return t.Text{Spans: t.ParseMarkup(markup, theme), Style: style}
	}

	label := func(text string, focused bool) t.Widget {
		colour := "$TextMuted"
		if focused {
			colour = "$FocusRing"
		}
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%s[/]", colour, text), theme)
	}

	return t.Column{
		Width: t.Flex(1),
		Children: []t.Widget{
			label("Category", categoryFocused),
			category,
			t.Text{Content: ""},
			label("Summary", ctx.IsFocused(summary)),
			summary,
		},
	}
}
