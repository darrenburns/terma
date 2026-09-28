package terma

import "testing"

func TestThemePaletteItems_GroupsAndMarksCurrent(t *testing.T) {
	original := CurrentThemeName()
	defer SetTheme(original)
	SetTheme(ThemeNameDracula)

	var selected string
	items := ThemePaletteItems(func(name string) { selected = name })

	if items[0].Divider != "Dark Themes" {
		t.Fatalf("expected first item to be the dark themes divider, got %+v", items[0])
	}
	themeCount := 0
	hasLightDivider := false
	for _, item := range items {
		if item.IsDivider() {
			hasLightDivider = hasLightDivider || item.Divider == "Light Themes"
			continue
		}
		themeCount++
		name := item.Data.(string)
		if item.Current != (name == ThemeNameDracula) {
			t.Fatalf("item %q has Current=%v", name, item.Current)
		}
		if item.HintWidget == nil {
			t.Fatalf("item %q has no swatch", name)
		}
	}
	if !hasLightDivider {
		t.Fatal("expected a light themes divider")
	}
	if themeCount != len(ThemeNames()) {
		t.Fatalf("expected %d theme items, got %d", len(ThemeNames()), themeCount)
	}

	items[1].Action()
	if selected != items[1].Data.(string) {
		t.Fatalf("expected onSelect(%q), got %q", items[1].Data, selected)
	}
}

func TestThemeDisplayName(t *testing.T) {
	if got := ThemeDisplayName("rose-pine-dawn"); got != "Rose Pine Dawn" {
		t.Fatalf("got %q", got)
	}
}

func TestSnapshot_CommandPalette_ThemeSwatches(t *testing.T) {
	original := CurrentThemeName()
	defer SetTheme(original)
	SetTheme(ThemeNameCatppuccin)

	state := NewCommandPaletteState("Commands", []CommandPaletteItem{{Label: "Themes"}})
	state.PushLevel("Themes", ThemePaletteItems(nil))
	state.Visible.Set(true)

	widget := CommandPalette{
		ID:       "palette-themes",
		State:    state,
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 2, Y: 1},
	}

	AssertSnapshot(t, widget, 80, 24, "Themes level of the command palette: each theme shows four coloured dots (primary, secondary, accent, background) on the right, and the cursor rests on the current theme, Catppuccin")
}
