// Package splitpanedemo demonstrates nested SplitPanes with draggable,
// keyboard-movable dividers.
package splitpanedemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "split-pane",
	Title:       "SplitPane",
	Description: "Nested split panes with draggable, rotatable dividers",
}

// SplitPaneDemo is a small editor-style workspace built from two nested
// SplitPanes: a file list on one side of the outer split, and a preview and a
// key reference on either side of the inner split.
//
//	mouse        - Drag a divider
//	tab          - Cycle focus: files → outer divider → inner divider
//	← → h l      - Move a horizontal divider (also from inside its panes)
//	↑ ↓          - Move a vertical divider (when it has focus)
//	escape       - Leave a divider and return to the file list
//	o / i        - Rotate the outer / inner split
//	=            - Split both evenly
//	d            - Cycle divider thickness
//	r            - Reset the layout
//	t            - Cycle theme
type SplitPaneDemo struct {
	outer            *t.SplitPaneState
	inner            *t.SplitPaneState
	outerOrientation t.Signal[t.SplitPaneOrientation]
	innerOrientation t.Signal[t.SplitPaneOrientation]
	thickness        t.Signal[int]
	files            *t.ListState[string]
}

const (
	outerStart   = 0.35
	innerStart   = 0.55
	maxThickness = 3

	filesID = "files"
	outerID = "outer-split"
	innerID = "inner-split"
)

// New creates the demo.
func New() demokit.Demo {
	return &SplitPaneDemo{
		outer:            t.NewSplitPaneState(outerStart),
		inner:            t.NewSplitPaneState(innerStart),
		outerOrientation: t.NewSignal(t.SplitHorizontal),
		innerOrientation: t.NewSignal(t.SplitVertical),
		thickness:        t.NewSignal(1),
		files:            t.NewListState(append([]string(nil), fileNames...)),
	}
}

func (d *SplitPaneDemo) InitialFocus() string { return filesID }

func rotate(orientation t.Signal[t.SplitPaneOrientation]) {
	orientation.Update(func(o t.SplitPaneOrientation) t.SplitPaneOrientation {
		if o == t.SplitHorizontal {
			return t.SplitVertical
		}
		return t.SplitHorizontal
	})
}

func (d *SplitPaneDemo) reset() {
	d.outerOrientation.Set(t.SplitHorizontal)
	d.innerOrientation.Set(t.SplitVertical)
	d.thickness.Set(1)
	d.outer.SetPosition(outerStart)
	d.inner.SetPosition(innerStart)
}

func (d *SplitPaneDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "o", Name: "Rotate", Action: func() { rotate(d.outerOrientation) }},
		{Key: "i", Name: "Rotate inner", Action: func() { rotate(d.innerOrientation) }, Hidden: true},
		{Key: "=", Name: "Even", Action: func() {
			d.outer.SetPosition(0.5)
			d.inner.SetPosition(0.5)
		}, Hidden: true},
		{Key: "d", Name: "Thickness", Action: func() {
			d.thickness.Update(func(n int) int { return n%maxThickness + 1 })
		}, Hidden: true},
		{Key: "r", Name: "Reset", Action: d.reset},
		{Key: "t", Name: "Theme", Action: demokit.NextTheme, Hidden: true},
	}
}

func (d *SplitPaneDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID:    "split-pane-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top: []t.Widget{
			demokit.Header{Title: "SplitPane Playground", Tagline: "Drag or nudge the dividers"},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Column{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				statePanel{demo: d},
				demokit.Fill(workspace{demo: d}),
			},
		},
	}
}

func orientationName(o t.SplitPaneOrientation) string {
	if o == t.SplitVertical {
		return "vertical"
	}
	return "horizontal"
}

// statePanel is a strip of live readouts for both splits. It reads the
// signals itself, so dragging a divider rebuilds only this strip.
type statePanel struct {
	demo *SplitPaneDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	focus := "[$TextMuted]—[/]"
	switch focusedID(ctx) {
	case filesID:
		focus = "[b $Accent]file list[/]"
	case outerID:
		focus = "[b $Primary]outer divider[/]"
	case innerID:
		focus = "[b $Secondary]inner divider[/]"
	}

	thickness := d.thickness.Get()
	unit := "cells"
	if thickness == 1 {
		unit = "cell"
	}

	return t.Row{
		Width:   t.Flex(1),
		Spacing: 3,
		Style:   demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			splitReadout(theme, "Outer", "$Primary", d.outerOrientation.Get(), d.outer.DividerPosition.Get()),
			splitReadout(theme, "Inner", "$Secondary", d.innerOrientation.Get(), d.inner.DividerPosition.Get()),
			t.Column{
				Width: t.Flex(1),
				Children: []t.Widget{
					demokit.StatRow(theme, "Focus", focus),
					demokit.StatRow(theme, "Divider", fmt.Sprintf("[b $Info]%d %s[/]", thickness, unit)),
				},
			},
		},
	}
}

// splitReadout shows one split's orientation and divider position, with a
// bar that fills to where the divider sits.
func splitReadout(theme t.ThemeData, name, colour string, orientation t.SplitPaneOrientation, position float64) t.Widget {
	filled := theme.Primary
	if colour == "$Secondary" {
		filled = theme.Secondary
	}
	return t.Column{
		Width: t.Flex(1),
		Children: []t.Widget{
			demokit.StatRow(theme, name, fmt.Sprintf("[b %s]%s[/]", colour, orientationName(orientation))),
			t.Row{
				Width:   t.Flex(1),
				Spacing: 1,
				Children: []t.Widget{
					t.ProgressBar{
						Progress:    position,
						FilledColor: filled,
						Style:       t.Style{Width: t.Flex(1)},
					},
					t.ParseMarkupToText(fmt.Sprintf("[b %s]%3.0f%%[/]", colour, position*100), theme),
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

// workspace is the nested pair of SplitPanes that fills the screen.
type workspace struct {
	demo *SplitPaneDemo
}

func (w workspace) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := w.demo
	thickness := d.thickness.Get()
	backToFiles := func() { t.RequestFocus(filesID) }

	inner := d.innerOrientation.Get()
	outer := d.outerOrientation.Get()

	split := t.SplitPane{
		ID:                     outerID,
		State:                  d.outer,
		Orientation:            outer,
		DividerSize:            thickness,
		MinPaneSize:            minPaneSize(outer),
		DividerChar:            dividerChar(outer),
		DividerForeground:      theme.Border,
		DividerFocusForeground: dividerGradient(outer, theme.Primary, theme.Accent),
		OnExitFocus:            backToFiles,
		First:                  filesPane{demo: d},
		Second: t.SplitPane{
			ID:                     innerID,
			State:                  d.inner,
			Orientation:            inner,
			DividerSize:            thickness,
			MinPaneSize:            minPaneSize(inner),
			DividerChar:            dividerChar(inner),
			DividerForeground:      theme.Border,
			DividerFocusForeground: dividerGradient(inner, theme.Secondary, theme.Accent),
			OnExitFocus:            backToFiles,
			First:                  previewPane{demo: d},
			Second:                 keysPane{},
		},
	}

	return split
}

// minPaneSize keeps side-by-side panes wide enough to read, and stacked panes
// tall enough to show a line between their borders.
func minPaneSize(o t.SplitPaneOrientation) int {
	if o == t.SplitVertical {
		return 3
	}
	return 12
}

// dividerChar draws dividers dashed, so they read as handles rather than as
// another panel border.
func dividerChar(o t.SplitPaneOrientation) string {
	if o == t.SplitVertical {
		return "╌"
	}
	return "╎"
}

// dividerGradient runs along the divider, whichever way it points.
func dividerGradient(o t.SplitPaneOrientation, from, to t.Color) t.Gradient {
	angle := 0.0
	if o == t.SplitVertical {
		angle = 90
	}
	return t.NewGradient(from, to).WithAngle(angle)
}

// filesPane is the file list in the outer split's first pane.
type filesPane struct {
	demo *SplitPaneDemo
}

func (f filesPane) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	list := t.List[string]{
		ID:    filesID,
		State: f.demo.files,
		Style: t.Style{Width: t.Flex(1)},
	}
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "Files", ctx.IsFocused(list)),
		Children: []t.Widget{list},
	}
}

// previewPane shows the file under the list cursor, wrapped to the pane.
type previewPane struct {
	demo *SplitPaneDemo
}

func (p previewPane) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	files := p.demo.files
	name := ""
	if idx := files.CursorIndex.Get(); idx >= 0 && idx < len(fileNames) {
		name = fileNames[idx]
	}
	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, name, false),
		Children: []t.Widget{
			t.Text{
				Content: fileContents[name],
				Wrap:    t.WrapSoft,
				Style:   t.Style{ForegroundColor: theme.Text},
			},
		},
	}
}

// keysPane is a quick reference for the keys the demo responds to.
type keysPane struct{}

func (keysPane) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("mouse", "$Accent", "drag a divider"),
			key("tab", "$Info", "cycle files and dividers"),
			key("←→ h l", "$Primary", "move a side-by-side divider"),
			key("↑↓", "$Secondary", "move a stacked divider"),
			key("esc", "$Info", "back to the file list"),
			key("o i", "$Warning", "rotate outer / inner split"),
			key("= d", "$Success", "split evenly / divider thickness"),
			key("r t", "$Error", "reset layout / cycle theme"),
		},
	}
}

var fileNames = []string{"README.md", "main.go", "go.mod", "CHANGELOG.md", "LICENSE"}

var fileContents = map[string]string{
	"README.md": "# Terma\n\n" +
		"A declarative terminal UI framework for Go. Widgets are plain structs, " +
		"state lives in signals, and only the parts of the screen that read a " +
		"changed signal are rebuilt.\n\n" +
		"Drag a divider, or focus one with Tab and nudge it with the arrow keys, " +
		"and watch this text reflow to fit its pane.",
	"main.go": "t.SplitPane{\n" +
		"    ID:          \"outer-split\",\n" +
		"    State:       t.NewSplitPaneState(0.35),\n" +
		"    Orientation: t.SplitHorizontal,\n" +
		"    MinPaneSize: 12,\n" +
		"    First:       files,\n" +
		"    Second:      preview,\n" +
		"}",
	"go.mod": "module example.com/workspace\n\ngo 1.25\n\nrequire github.com/darrenburns/terma v0.0.0",
	"CHANGELOG.md": "## Unreleased\n\n" +
		"- SplitPane repaints its divider when a drag starts and ends.\n" +
		"- Dividers can be moved from the keyboard while a child pane has focus.\n" +
		"- OnExitFocus lets Escape hand focus back to your content.",
	"LICENSE": "MIT License\n\n" +
		"Permission is hereby granted, free of charge, to any person obtaining a " +
		"copy of this software, to deal in the software without restriction.",
}
