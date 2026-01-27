package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"
)

// App demonstrates FocusTrap with 3 columns, each acting as a focus trap.
type App struct {
	trapActive t.Signal[bool]
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
		{Key: "1", Name: "Column 1", Action: func() { t.RequestFocus("col1-btn1") }},
		{Key: "2", Name: "Column 2", Action: func() { t.RequestFocus("col2-btn1") }},
		{Key: "3", Name: "Column 3", Action: func() { t.RequestFocus("col3-btn1") }},
		{Key: "f", Name: "Toggle Trap", Action: func() {
			a.trapActive.Update(func(v bool) bool { return !v })
		}},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	active := a.trapActive.Get()
	theme := ctx.Theme()

	statusText := "OFF"
	statusColor := theme.Error
	if active {
		statusText = "ON"
		statusColor = theme.Success
	}

	// Determine which column prefix the focused widget belongs to
	focusedPrefix := ""
	focused := ctx.Focused()
	if focused != nil {
		if id, ok := focused.(t.Identifiable); ok {
			wid := id.WidgetID()
			for _, prefix := range []string{"col1", "col2", "col3"} {
				if strings.HasPrefix(wid, prefix) {
					focusedPrefix = prefix
					break
				}
			}
		}
	}

	return t.Dock{
		Bottom: []t.Widget{t.KeybindBar{}},
		Top: []t.Widget{
			t.Column{
				Style: t.Style{
					Padding: t.EdgeInsets{Left: 1, Right: 1, Top: 1, Bottom: 1},
				},
				Spacing: 1,
				Children: []t.Widget{
					t.Text{
						Spans: []t.Span{
							{Text: "Focus Trap Demo", Style: t.SpanStyle{Bold: true}},
							{Text: "  "},
							{Text: "Trap: ", Style: t.SpanStyle{Foreground: theme.TextMuted}},
							{Text: statusText, Style: t.SpanStyle{Bold: true, Foreground: statusColor}},
						},
					},
					t.Text{
						Content: "Press 1/2/3 to jump to a column. Tab/Shift+Tab to cycle. f to toggle trap.",
						Style:   t.Style{ForegroundColor: theme.TextMuted},
					},
				},
			},
		},
		Body: t.Row{
			Style: t.Style{
				Padding: t.EdgeInsets{Left: 1, Right: 1},
			},
			Spacing: 1,
			Children: []t.Widget{
				buildColumn(ctx, "col1", "Column 1", active, focusedPrefix == "col1"),
				buildColumn(ctx, "col2", "Column 2", active, focusedPrefix == "col2"),
				buildColumn(ctx, "col3", "Column 3", active, focusedPrefix == "col3"),
			},
		},
	}
}

func buildColumn(ctx t.BuildContext, prefix, title string, trapActive bool, hasFocus bool) t.Widget {
	theme := ctx.Theme()

	borderColor := theme.Border
	titleMarkup := fmt.Sprintf(" %s ", title)
	if hasFocus {
		borderColor = theme.FocusRing
		titleMarkup = fmt.Sprintf("[b $FocusRing] %s [/]", title)
	}

	return t.FocusTrap{
		ID:     prefix + "-trap",
		Active: trapActive,
		Child: t.Column{
			Width:  t.Flex(1),
			Height: t.Flex(1),
			Style: t.Style{
				Border: t.Border{
					Style:       t.BorderRounded,
					Color:       borderColor,
					Decorations: []t.BorderDecoration{{Markup: titleMarkup, Position: t.DecorationTopLeft}},
				},
				Padding: t.EdgeInsets{Left: 1, Right: 1, Top: 1, Bottom: 1},
			},
			Spacing: 1,
			Children: []t.Widget{
				t.Button{ID: prefix + "-btn1", Label: "Button A", OnPress: func() {}},
				t.Button{ID: prefix + "-btn2", Label: "Button B", OnPress: func() {}},
				t.Button{ID: prefix + "-btn3", Label: "Button C", OnPress: func() {}},
			},
		},
	}
}

func main() {
	app := &App{
		trapActive: t.NewSignal(true),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
