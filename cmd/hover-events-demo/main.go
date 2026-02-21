package main

import (
	"fmt"
	"log"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	t "github.com/darrenburns/terma"
)

type HoverEventsDemo struct {
	currentHovered t.Signal[string]
	lastEvent      t.Signal[string]
	history        t.Signal[string]
	eventCount     t.Signal[int]
	enterCount     t.Signal[int]
	leaveCount     t.Signal[int]
	checkboxState  *t.CheckboxState
}

func NewHoverEventsDemo() *HoverEventsDemo {
	return &HoverEventsDemo{
		currentHovered: t.NewSignal("(none)"),
		lastEvent:      t.NewSignal("(none yet)"),
		history:        t.NewSignal("(no transitions yet)"),
		eventCount:     t.NewSignal(0),
		enterCount:     t.NewSignal(0),
		leaveCount:     t.NewSignal(0),
		checkboxState:  t.NewCheckboxState(false),
	}
}

func (a *HoverEventsDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "c", Name: "Clear Stats", Action: a.reset},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *HoverEventsDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	contextHovered := ctx.HoveredID()
	if contextHovered == "" {
		contextHovered = "(none)"
	}

	return t.Dock{
		Bottom: []t.Widget{t.KeybindBar{}},
		Body: t.Column{
			Spacing: 1,
			Style: t.Style{
				Padding:         t.EdgeInsetsAll(1),
				BackgroundColor: theme.Background,
			},
			Children: []t.Widget{
				t.Text{
					Content: "Hover Events Demo",
					Style: t.Style{
						ForegroundColor: theme.TextOnPrimary,
						BackgroundColor: theme.Primary,
						Padding:         t.EdgeInsetsAll(1),
						Bold:            true,
					},
				},
				t.Text{Content: "Move your mouse between targets to trigger HoverEnter/HoverLeave transitions."},
				t.Text{Content: "Press c to reset counters, q to quit."},
				t.Row{
					Spacing: 2,
					Children: []t.Widget{
						t.Text{Content: fmt.Sprintf("Events: %d", a.eventCount.Get()), Style: t.Style{ForegroundColor: theme.Text}},
						t.Text{Content: fmt.Sprintf("Enter: %d", a.enterCount.Get()), Style: t.Style{ForegroundColor: theme.Success}},
						t.Text{Content: fmt.Sprintf("Leave: %d", a.leaveCount.Get()), Style: t.Style{ForegroundColor: theme.Warning}},
					},
				},
				t.Text{Content: fmt.Sprintf("Event-tracked hovered ID: %s", a.currentHovered.Get())},
				t.Text{Content: fmt.Sprintf("BuildContext.HoveredID(): %s", contextHovered)},
				t.Text{
					Content: fmt.Sprintf("Last event: %s", a.lastEvent.Get()),
					Style: t.Style{
						BackgroundColor: theme.Surface,
						Padding:         t.EdgeInsetsAll(1),
						ForegroundColor: theme.Text,
					},
				},
				t.Text{Content: "Hover targets:"},
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						t.Button{
							ID:    "hover-btn-a",
							Label: a.targetLabel("hover-btn-a", "Button A"),
							Hover: a.onHover,
						},
						t.Button{
							ID:    "hover-btn-b",
							Label: a.targetLabel("hover-btn-b", "Button B"),
							Hover: a.onHover,
						},
						t.Button{
							ID:    "hover-btn-c",
							Label: a.targetLabel("hover-btn-c", "Button C"),
							Hover: a.onHover,
						},
					},
				},
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						t.Text{
							ID:      "hover-text",
							Content: a.targetLabel("hover-text", " Hover this text target "),
							Hover:   a.onHover,
							Style: t.Style{
								Padding:         t.EdgeInsetsAll(1),
								BackgroundColor: theme.Surface,
								ForegroundColor: theme.Text,
							},
						},
						&t.Checkbox{
							ID:    "hover-checkbox",
							State: a.checkboxState,
							Label: a.targetLabel("hover-checkbox", " Checkbox target"),
							Hover: a.onHover,
						},
					},
				},
				t.Text{Content: "Recent transitions (newest first):"},
				t.Text{
					Content: a.history.Get(),
					Style: t.Style{
						BackgroundColor: theme.Surface,
						Padding:         t.EdgeInsetsAll(1),
						ForegroundColor: theme.Text,
					},
				},
			},
		},
	}
}

func (a *HoverEventsDemo) targetLabel(id, label string) string {
	if a.currentHovered.Get() == id {
		return "[hovered] " + label
	}
	return label
}

func (a *HoverEventsDemo) onHover(event t.HoverEvent) {
	a.eventCount.Update(func(v int) int { return v + 1 })

	switch event.Type {
	case t.HoverEnter:
		a.enterCount.Update(func(v int) int { return v + 1 })
		a.currentHovered.Set(safeID(event.WidgetID))
	case t.HoverLeave:
		a.leaveCount.Update(func(v int) int { return v + 1 })
		a.currentHovered.Set(safeID(event.NextWidgetID))
	}

	line := formatHoverEvent(event)
	a.lastEvent.Set(line)
	a.history.Set(prependHistory(a.history.Get(), line, 8))
}

func (a *HoverEventsDemo) reset() {
	a.currentHovered.Set("(none)")
	a.lastEvent.Set("(none yet)")
	a.history.Set("(no transitions yet)")
	a.eventCount.Set(0)
	a.enterCount.Set(0)
	a.leaveCount.Set(0)
}

func prependHistory(existing, next string, maxLines int) string {
	if maxLines < 1 {
		maxLines = 1
	}

	if existing == "" || existing == "(no transitions yet)" {
		return next
	}

	lines := []string{next}
	for _, line := range strings.Split(existing, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "(no transitions yet)" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= maxLines {
			break
		}
	}

	return strings.Join(lines, "\n")
}

func formatHoverEvent(event t.HoverEvent) string {
	return fmt.Sprintf(
		"%s target=%s prev=%s next=%s pos=(%d,%d) local=(%d,%d) mod=%s button=%s",
		hoverTypeLabel(event.Type),
		safeID(event.WidgetID),
		safeID(event.PreviousWidgetID),
		safeID(event.NextWidgetID),
		event.X,
		event.Y,
		event.LocalX,
		event.LocalY,
		modString(event.Mod),
		fmt.Sprint(event.Button),
	)
}

func hoverTypeLabel(eventType t.HoverEventType) string {
	switch eventType {
	case t.HoverEnter:
		return "enter"
	case t.HoverLeave:
		return "leave"
	default:
		return "unknown"
	}
}

func modString(mod uv.KeyMod) string {
	if mod == 0 {
		return "none"
	}
	parts := []string{}
	if mod.Contains(uv.ModCtrl) {
		parts = append(parts, "ctrl")
	}
	if mod.Contains(uv.ModAlt) {
		parts = append(parts, "alt")
	}
	if mod.Contains(uv.ModShift) {
		parts = append(parts, "shift")
	}
	if mod.Contains(uv.ModMeta) {
		parts = append(parts, "meta")
	}
	if mod.Contains(uv.ModSuper) {
		parts = append(parts, "super")
	}
	if mod.Contains(uv.ModHyper) {
		parts = append(parts, "hyper")
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "+")
}

func safeID(id string) string {
	if id == "" {
		return "(none)"
	}
	return id
}

func main() {
	app := NewHoverEventsDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
