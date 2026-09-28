package terma

import "strings"

// ThemePaletteItems returns command palette items for every registered theme,
// grouped under "Dark Themes" and "Light Themes" dividers. Each item shows a
// ThemeSwatch hint, is Current when it is the active theme, and carries the
// theme name in Data so an OnCursorChange handler can preview it.
// onSelect is called with the chosen theme's name.
func ThemePaletteItems(onSelect func(themeName string)) []CommandPaletteItem {
	items := make([]CommandPaletteItem, 0, len(themeRegistry)+2)
	current := CurrentThemeName()
	addGroup := func(title string, names []string) {
		if len(names) == 0 {
			return
		}
		items = append(items, CommandPaletteItem{Divider: title})
		for _, name := range names {
			label := ThemeDisplayName(name)
			items = append(items, CommandPaletteItem{
				Label:      label,
				FilterText: label + " " + name,
				HintWidget: func() Widget { return ThemeSwatch(name) },
				Current:    name == current,
				Data:       name,
				Action: func() {
					if onSelect != nil {
						onSelect(name)
					}
				},
			})
		}
	}
	addGroup("Dark Themes", DarkThemeNames())
	addGroup("Light Themes", LightThemeNames())
	return items
}

// ThemeSwatch returns a small preview chip for a registered theme: its
// background and surface colours fill the chip's left and right halves, with
// primary, secondary and accent dots drawn across them. Unknown themes render
// nothing.
func ThemeSwatch(name string) Widget {
	data, ok := GetTheme(name)
	if !ok {
		return EmptyWidget{}
	}
	bg, surface := data.Background, data.Surface
	cell := func(text string, fg, bg Color) Span {
		return Span{Text: text, Style: SpanStyle{Foreground: fg, Background: bg}}
	}
	return Text{Spans: []Span{
		cell(" ", Color{}, bg),
		cell("●", data.Primary, bg),
		cell(" ", Color{}, bg),
		cell("●", data.Secondary, surface),
		cell(" ", Color{}, surface),
		cell("●", data.Accent, surface),
		cell(" ", Color{}, surface),
	}}
}

// ThemeDisplayName turns a theme name such as "rose-pine" into "Rose Pine".
func ThemeDisplayName(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
