// Package toastdemo demonstrates toast notifications: each severity, sticky
// toasts, the waiting queue, corner placement, and toasts sent from a
// background goroutine.
package toastdemo

import (
	"fmt"
	"time"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "toast",
	Title:       "Toasts",
	Description: "Stacked, auto-dismissing notifications",
}

var positions = []struct {
	name     string
	position t.FloatPosition
}{
	{"bottom right", t.FloatPositionBottomRight},
	{"bottom center", t.FloatPositionBottomCenter},
	{"bottom left", t.FloatPositionBottomLeft},
	{"top left", t.FloatPositionTopLeft},
	{"top center", t.FloatPositionTopCenter},
	{"top right", t.FloatPositionTopRight},
}

type ToastDemo struct {
	toasts   *t.ToastState
	position t.Signal[int]
	sent     t.Signal[int]
}

// New creates the demo.
func New() demokit.Demo {
	return &ToastDemo{
		toasts:   t.NewToastState(t.ToastOptions{}),
		position: t.NewSignal(0),
		sent:     t.NewSignal(0),
	}
}

func (d *ToastDemo) InitialFocus() string { return "toast-info" }

func (d *ToastDemo) notify(toast t.Toast) {
	d.toasts.Notify(toast)
	d.sent.Update(func(n int) int { return n + 1 })
}

type action struct {
	id, key, label string
	run            func(*ToastDemo)
}

var actions = []action{
	{"toast-info", "1", "Info", func(d *ToastDemo) { d.notify(t.Toast{Message: "Indexing 214 files"}) }},
	{"toast-success", "2", "Success", func(d *ToastDemo) {
		d.notify(t.Toast{Title: "Saved", Message: "Wrote notes.md to disk.", Severity: t.ToastSuccess})
	}},
	{"toast-warning", "3", "Warning", func(d *ToastDemo) {
		d.notify(t.Toast{Message: "Sync is slow. It will keep retrying in the background.", Severity: t.ToastWarning})
	}},
	{"toast-error", "4", "Error", func(d *ToastDemo) {
		d.notify(t.Toast{Title: "Copy failed", Message: "The terminal refused clipboard access.", Severity: t.ToastError})
	}},
	{"toast-sticky", "5", "Sticky", func(d *ToastDemo) {
		d.notify(t.Toast{Title: "Update available", Message: "Stays until clicked.", Timeout: -1})
	}},
	{"toast-background", "6", "From a goroutine", (*ToastDemo).background},
	{"toast-position", "p", "Move corner", func(d *ToastDemo) {
		d.position.Update(func(i int) int { return (i + 1) % len(positions) })
	}},
	{"toast-clear", "c", "Clear", func(d *ToastDemo) { d.toasts.Clear() }},
}

// background sends a toast every 400ms from its own goroutine, as a download
// or build task would.
func (d *ToastDemo) background() {
	go func() {
		for i := 1; i <= 5; i++ {
			d.notify(t.Toast{Message: fmt.Sprintf("Worker finished job %d of 5", i), Severity: t.ToastSuccess, Timeout: 2 * time.Second})
			time.Sleep(400 * time.Millisecond)
		}
	}()
}

func (d *ToastDemo) Keybinds() []t.Keybind {
	binds := make([]t.Keybind, 0, len(actions)+1)
	for _, a := range actions {
		binds = append(binds, t.Keybind{Key: a.key, Name: a.label, Action: func() { a.run(d) }})
	}
	return append(binds, t.Keybind{Key: "t", Name: "Theme", Action: demokit.NextTheme})
}

func (d *ToastDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	buttons := make([]t.Widget, 0, len(actions))
	for _, a := range actions {
		buttons = append(buttons, t.Row{Spacing: 2, Children: []t.Widget{
			t.ParseMarkupToText(fmt.Sprintf("[b $Accent]%s[/]", a.key), theme),
			t.Button{ID: a.id, Label: a.label, Width: t.Cells(20), OnPress: func() { a.run(d) }},
		}})
	}
	position := positions[d.position.Get()]
	return t.Dock{
		ID:     "toast-demo-root",
		Style:  t.Style{BackgroundColor: theme.Background},
		Top:    []t.Widget{demokit.Header{Title: "Toasts", Tagline: "Transient notifications in a corner"}},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Column{
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(2, 1)},
			Children: []t.Widget{
				t.Column{Spacing: 1, Children: buttons},
				demokit.StatRow(theme, "Sent", fmt.Sprintf("[b]%d[/]", d.sent.Get())),
				demokit.StatRow(theme, "Corner", "[b]"+position.name+"[/]"),
				t.ParseMarkupToText("[$TextMuted]Click a toast to dismiss it. Hovering one holds its timer.[/]", theme),
				t.Toasts{ID: "demo-toasts", State: d.toasts, Position: position.position},
			},
		},
	}
}
