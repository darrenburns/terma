package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"

	uv "github.com/charmbracelet/ultraviolet"
)

type MouseDemo struct {
	lastDown  t.Signal[string]
	lastClick t.Signal[string]
	lastUp    t.Signal[string]
	lastChain t.Signal[int]

	inputState *t.TextInputState
	areaState  *t.TextAreaState
}

func NewMouseDemo() *MouseDemo {
	return &MouseDemo{
		lastDown:   t.NewSignal(""),
		lastClick:  t.NewSignal(""),
		lastUp:     t.NewSignal(""),
		lastChain:  t.NewSignal(0),
		inputState: t.NewTextInputState(""),
		areaState:  t.NewTextAreaState("Click here and type.\nTry double and triple click."),
	}
}

func (a *MouseDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *MouseDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	focusedID := focusedLabel(ctx)
	hoveredID := ctx.HoveredID()

	return t.Column{
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Mouse Events Demo",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsAll(1),
				},
			},
			t.Text{Content: "Click, double click, and triple click to see chain counts. Click focusable widgets to focus them."},
			t.Text{
				Content: fmt.Sprintf("Focused: %s | Hovered: %s", focusedID, hoveredID),
				Style: t.Style{
					ForegroundColor: theme.Text,
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsAll(1),
				},
			},
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					&t.Button{
						ID:        "btn-primary",
						Label:     "Button A",
						OnPress:   func() {},
						Click:     a.onClick,
						MouseDown: a.onMouseDown,
						MouseUp:   a.onMouseUp,
					},
					&t.Button{
						ID:        "btn-secondary",
						Label:     "Button B",
						OnPress:   func() {},
						Click:     a.onClick,
						MouseDown: a.onMouseDown,
						MouseUp:   a.onMouseUp,
					},
					t.Text{
						ID:        "static-text",
						Content:   "Non-focusable text",
						Click:     a.onClick,
						MouseDown: a.onMouseDown,
						MouseUp:   a.onMouseUp,
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
							Padding:         t.EdgeInsetsAll(1),
						},
					},
				},
			},
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.TextInput{
						ID:          "text-input",
						State:       a.inputState,
						Width:       t.Cells(24),
						Placeholder: "Click to focus",
						Click:       a.onClick,
						MouseDown:   a.onMouseDown,
						MouseUp:     a.onMouseUp,
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
						},
					},
					t.TextArea{
						ID:        "text-area",
						State:     a.areaState,
						Width:     t.Cells(32),
						Height:    t.Cells(4),
						Click:     a.onClick,
						MouseDown: a.onMouseDown,
						MouseUp:   a.onMouseUp,
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
						},
					},
				},
			},
			t.Text{
				Content: fmt.Sprintf("Last MouseDown: %s", a.lastDown.Get()),
				Style:   t.Style{ForegroundColor: theme.Text},
			},
			t.Text{
				Content: fmt.Sprintf("Last Click: %s", a.lastClick.Get()),
				Style:   t.Style{ForegroundColor: theme.Text},
			},
			t.Text{
				Content: fmt.Sprintf("Last MouseUp: %s", a.lastUp.Get()),
				Style:   t.Style{ForegroundColor: theme.Text},
			},
			t.Text{
				Content: fmt.Sprintf("Last Click Count: %d", a.lastChain.Get()),
				Style:   t.Style{ForegroundColor: theme.Text},
			},
			t.Text{Content: "Press q to quit."},
		},
	}
}

func (a *MouseDemo) onMouseDown(ev t.MouseEvent) {
	a.lastDown.Set(formatMouseEvent("down", ev))
	t.Log("MouseDown: %s", formatMouseEvent("down", ev))
}

func (a *MouseDemo) onClick(ev t.MouseEvent) {
	a.lastClick.Set(formatMouseEvent("click", ev))
	a.lastChain.Set(ev.ClickCount)
	t.Log("Click: %s", formatMouseEvent("click", ev))
}

func (a *MouseDemo) onMouseUp(ev t.MouseEvent) {
	a.lastUp.Set(formatMouseEvent("up", ev))
	t.Log("MouseUp: %s", formatMouseEvent("up", ev))
}

func focusedLabel(ctx t.BuildContext) string {
	focused := ctx.Focused()
	if focused == nil {
		return "none"
	}
	if identifiable, ok := focused.(t.Identifiable); ok && identifiable.WidgetID() != "" {
		return identifiable.WidgetID()
	}
	return fmt.Sprintf("%T", focused)
}

func formatMouseEvent(kind string, ev t.MouseEvent) string {
	mod := modString(ev.Mod)
	button := fmt.Sprint(ev.Button)
	return fmt.Sprintf("%s id=%s button=%s mod=%s count=%d x=%d y=%d",
		kind, safeID(ev.WidgetID), button, mod, ev.ClickCount, ev.X, ev.Y)
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
	app := NewMouseDemo()
	if err := t.InitLogger(); err != nil {
		log.Fatal(err)
	}
	defer t.CloseLogger()
	t.Log("Mouse events demo started")
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
