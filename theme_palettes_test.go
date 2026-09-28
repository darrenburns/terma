package terma

import "testing"

var jewelToneThemeNames = []string{
	ThemeNameMoonstone,
	ThemeNameBonsai,
	ThemeNameGarnet,
	ThemeNameKintsugi,
	ThemeNameAmethyst,
	ThemeNameLantern,
}

func TestJewelToneThemes_Readability(t *testing.T) {
	for _, name := range jewelToneThemeNames {
		t.Run(name, func(t *testing.T) {
			theme, ok := GetTheme(name)
			if !ok {
				t.Fatalf("theme %q is not registered", name)
			}
			if theme.IsLight {
				t.Errorf("theme %q should be a dark theme", name)
			}
			for _, background := range []Color{theme.Background, theme.Surface, theme.SurfaceHover} {
				if ratio := theme.Text.ContrastRatio(background); ratio < 4.5 {
					t.Errorf("Text on %s has contrast %.2f, want >= 4.5", background.Hex(), ratio)
				}
				if ratio := theme.TextMuted.ContrastRatio(background); ratio < 4.5 {
					t.Errorf("TextMuted on %s has contrast %.2f, want >= 4.5", background.Hex(), ratio)
				}
			}
			if ratio := theme.SelectionText.ContrastRatio(theme.ActiveCursor); ratio < 4.5 {
				t.Errorf("SelectionText on ActiveCursor has contrast %.2f, want >= 4.5", ratio)
			}
			variants := map[string][2]Color{
				"Primary":   {theme.TextOnPrimary, theme.Primary},
				"Secondary": {theme.TextOnSecondary, theme.Secondary},
				"Accent":    {theme.TextOnAccent, theme.Accent},
				"Error":     {theme.TextOnError, theme.Error},
				"Warning":   {theme.TextOnWarning, theme.Warning},
				"Success":   {theme.TextOnSuccess, theme.Success},
				"Info":      {theme.TextOnInfo, theme.Info},
			}
			for variant, pair := range variants {
				if ratio := pair[0].ContrastRatio(pair[1]); ratio < 4.5 {
					t.Errorf("TextOn%s on %s has contrast %.2f, want >= 4.5", variant, variant, ratio)
				}
			}
		})
	}
}

func TestSnapshot_JewelToneThemes(t *testing.T) {
	originalThemeName := CurrentThemeName()
	defer SetTheme(originalThemeName)

	for _, name := range jewelToneThemeNames {
		SetTheme(name)
		theme, _ := GetTheme(name)
		widget := Column{
			Width:   Cells(44),
			Height:  Cells(12),
			Spacing: 1,
			Style: Style{
				BackgroundColor: theme.Background,
				Padding:         EdgeInsetsAll(1),
			},
			Children: []Widget{
				Text{Content: name, Style: Style{Bold: true, ForegroundColor: theme.Text}},
				Text{Content: "Muted secondary text", Style: Style{ForegroundColor: theme.TextMuted}},
				Row{
					Spacing: 1,
					Children: []Widget{
						Button{ID: name + "-primary", Label: "Primary", Variant: ButtonPrimary},
						Button{ID: name + "-accent", Label: "Accent", Variant: ButtonAccent},
						Button{ID: name + "-info", Label: "Info", Variant: ButtonInfo},
					},
				},
				Row{
					Spacing: 1,
					Children: []Widget{
						Button{ID: name + "-success", Label: "Success", Variant: ButtonSuccess},
						Button{ID: name + "-warning", Label: "Warning", Variant: ButtonWarning},
						Button{ID: name + "-error", Label: "Error", Variant: ButtonError},
					},
				},
			},
		}
		AssertSnapshotNamed(t, t.Name()+"_"+name, widget, 44, 12,
			name+" theme swatch: background, body and muted text, and "+
				"primary/accent/info/success/warning/error buttons using the theme's colors.")
	}
}
