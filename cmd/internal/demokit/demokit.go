// Package demokit holds what the widget demos share: the Demo interface they
// implement, the panel look they're drawn with, theme cycling, and Run for
// starting a demo on its own.
//
// Each demo lives in its own package under cmd/internal/demos, is started on
// its own by a small main package in cmd/<name>, and appears alongside the
// others in the cmd/terma-demos gallery.
package demokit

import (
	"fmt"
	"log"
	"slices"

	t "github.com/darrenburns/terma"
)

// Demo is a widget demo that can run on its own or inside the gallery.
type Demo interface {
	t.Widget
	// InitialFocus returns the ID of the widget to focus when the demo is
	// shown, or "" to leave focus where it is.
	InitialFocus() string
}

// Info describes a demo for the gallery.
type Info struct {
	Key         string // Stable identifier, e.g. "table"
	Title       string // Short name, e.g. "Table"
	Description string // One line on what the demo shows
}

// Themes are the themes the demos cycle through.
var Themes = []string{
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

// NextTheme switches to the theme after the current one in Themes.
func NextTheme() {
	i := slices.Index(Themes, t.CurrentThemeName())
	t.SetTheme(Themes[(i+1)%len(Themes)])
}

// SwitchHint is shown on the right of every demo's header when it is set. The
// gallery sets it to the key that switches demos.
var SwitchHint string

// Run starts a demo as a standalone app.
func Run(demo Demo) {
	t.SetTheme(Themes[0])
	if id := demo.InitialFocus(); id != "" {
		t.RequestFocus(id)
	}
	if err := t.Run(demo); err != nil {
		log.Fatal(err)
	}
}

// Header is the title bar across the top of a demo.
type Header struct {
	Title   string // Shown bold after "≡"
	Tagline string // Muted text after the title
	Right   string // Optional markup on the right; defaults to the theme name
}

func (h Header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	right := h.Right
	if right == "" {
		right = fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", theme.Name)
	}
	if SwitchHint != "" {
		right = SwitchHint + "  " + right
	}
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText(fmt.Sprintf("[b $Primary]≡ %s[/]  [$TextMuted]%s[/]", h.Title, h.Tagline), theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(right, theme),
		},
	}
}

// Footer is the KeybindBar along the bottom of a demo.
func Footer(theme t.ThemeData) t.Widget {
	return t.KeybindBar{
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
	}
}

// Fill gives a component the remaining space in its Column. Rows and Columns
// read Flex from their direct children, so a component's own Flex(1) needs a
// plain wrapper to take effect.
func Fill(child t.Widget) t.Widget {
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{child},
	}
}

// PanelStyle is the bordered look shared by every panel. The border takes the
// focus ring colour while the panel's widget has focus.
func PanelStyle(theme t.ThemeData, title string, focused bool) t.Style {
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

// StatRow is one line of a state panel: a muted label on the left and a value
// (markup) on the right.
func StatRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}
