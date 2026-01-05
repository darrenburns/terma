package terma

import (
	"testing"
)

// testTheme provides consistent colors for testing
var testTheme = ThemeData{
	Primary:       Hex("#ff0000"),
	Secondary:     Hex("#00ff00"),
	Accent:        Hex("#0000ff"),
	Background:    Hex("#111111"),
	Surface:       Hex("#222222"),
	SurfaceHover:  Hex("#333333"),
	Text:          Hex("#ffffff"),
	TextMuted:     Hex("#888888"),
	TextOnPrimary: Hex("#000000"),
	Border:        Hex("#444444"),
	FocusRing:     Hex("#555555"),
	Error:         Hex("#ee0000"),
	Warning:       Hex("#eeee00"),
	Success:       Hex("#00ee00"),
	Info:          Hex("#0000ee"),
}

func TestParseMarkup_PlainText(t *testing.T) {
	spans := ParseMarkup("Hello World", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", spans[0].Text)
	}
	if spans[0].Style.Bold || spans[0].Style.Italic || spans[0].Style.Underline != UnderlineNone {
		t.Error("expected no styling")
	}
}

func TestParseMarkup_Bold(t *testing.T) {
	spans := ParseMarkup("[bold]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", spans[0].Text)
	}
	if !spans[0].Style.Bold {
		t.Error("expected bold")
	}
}

func TestParseMarkup_BoldShorthand(t *testing.T) {
	spans := ParseMarkup("[b]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("expected bold")
	}
}

func TestParseMarkup_Italic(t *testing.T) {
	spans := ParseMarkup("[italic]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Italic {
		t.Error("expected italic")
	}
}

func TestParseMarkup_ItalicShorthand(t *testing.T) {
	spans := ParseMarkup("[i]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Italic {
		t.Error("expected italic")
	}
}

func TestParseMarkup_Underline(t *testing.T) {
	spans := ParseMarkup("[underline]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Underline != UnderlineSingle {
		t.Error("expected underline")
	}
}

func TestParseMarkup_UnderlineShorthand(t *testing.T) {
	spans := ParseMarkup("[u]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Underline != UnderlineSingle {
		t.Error("expected underline")
	}
}

func TestParseMarkup_CombinedStyles(t *testing.T) {
	spans := ParseMarkup("[b i u]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("expected bold")
	}
	if !spans[0].Style.Italic {
		t.Error("expected italic")
	}
	if spans[0].Style.Underline != UnderlineSingle {
		t.Error("expected underline")
	}
}

func TestParseMarkup_ThemeColor(t *testing.T) {
	spans := ParseMarkup("[$Primary]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Foreground != testTheme.Primary {
		t.Errorf("expected Primary color, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_ThemeColorLowercase(t *testing.T) {
	spans := ParseMarkup("[$primary]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Foreground != testTheme.Primary {
		t.Errorf("expected Primary color, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_ThemeColorSnakeCase(t *testing.T) {
	spans := ParseMarkup("[$text_muted]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Foreground != testTheme.TextMuted {
		t.Errorf("expected TextMuted color, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_ThemeColorSnakeCaseSurfaceHover(t *testing.T) {
	spans := ParseMarkup("[$surface_hover]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Foreground != testTheme.SurfaceHover {
		t.Errorf("expected SurfaceHover color, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_BackgroundColor(t *testing.T) {
	spans := ParseMarkup("[on $Surface]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Background != testTheme.Surface {
		t.Errorf("expected Surface background, got %v", spans[0].Style.Background)
	}
}

func TestParseMarkup_ForegroundAndBackground(t *testing.T) {
	spans := ParseMarkup("[$Primary on $Surface]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Style.Foreground != testTheme.Primary {
		t.Errorf("expected Primary foreground, got %v", spans[0].Style.Foreground)
	}
	if spans[0].Style.Background != testTheme.Surface {
		t.Errorf("expected Surface background, got %v", spans[0].Style.Background)
	}
}

func TestParseMarkup_HexColor(t *testing.T) {
	spans := ParseMarkup("[#ff5500]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	expected := Hex("#ff5500")
	if spans[0].Style.Foreground != expected {
		t.Errorf("expected #ff5500, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_HexBackground(t *testing.T) {
	spans := ParseMarkup("[on #333333]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	expected := Hex("#333333")
	if spans[0].Style.Background != expected {
		t.Errorf("expected #333333, got %v", spans[0].Style.Background)
	}
}

func TestParseMarkup_StyleWithColor(t *testing.T) {
	spans := ParseMarkup("[bold $Accent]Hello[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("expected bold")
	}
	if spans[0].Style.Foreground != testTheme.Accent {
		t.Errorf("expected Accent color, got %v", spans[0].Style.Foreground)
	}
}

func TestParseMarkup_Nesting(t *testing.T) {
	spans := ParseMarkup("[bold]Hello [italic]World[/][/]", testTheme)

	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}

	// First span: "Hello " with bold
	if spans[0].Text != "Hello " {
		t.Errorf("expected 'Hello ', got '%s'", spans[0].Text)
	}
	if !spans[0].Style.Bold {
		t.Error("first span should be bold")
	}
	if spans[0].Style.Italic {
		t.Error("first span should not be italic")
	}

	// Second span: "World" with bold+italic
	if spans[1].Text != "World" {
		t.Errorf("expected 'World', got '%s'", spans[1].Text)
	}
	if !spans[1].Style.Bold {
		t.Error("second span should be bold")
	}
	if !spans[1].Style.Italic {
		t.Error("second span should be italic")
	}
}

func TestParseMarkup_Escape(t *testing.T) {
	spans := ParseMarkup("Use [[brackets]] for arrays", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	// [[ -> [ and ]] -> ]
	if spans[0].Text != "Use [brackets] for arrays" {
		t.Errorf("expected 'Use [brackets] for arrays', got '%s'", spans[0].Text)
	}
}

func TestParseMarkup_EscapeMultiple(t *testing.T) {
	spans := ParseMarkup("[[a]] and [[b]]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "[a] and [b]" {
		t.Errorf("expected '[a] and [b]', got '%s'", spans[0].Text)
	}
}

func TestParseMarkup_EscapeOpenOnly(t *testing.T) {
	// Only [[ without ]] - should still work
	spans := ParseMarkup("[[open bracket", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "[open bracket" {
		t.Errorf("expected '[open bracket', got '%s'", spans[0].Text)
	}
}

func TestParseMarkup_EscapeCloseOnly(t *testing.T) {
	// ]] without [[ - should produce ]
	spans := ParseMarkup("close bracket]]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "close bracket]" {
		t.Errorf("expected 'close bracket]', got '%s'", spans[0].Text)
	}
}

func TestParseMarkup_MixedContent(t *testing.T) {
	spans := ParseMarkup("Press [b $Accent]Enter[/] to continue", testTheme)

	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}

	// "Press "
	if spans[0].Text != "Press " {
		t.Errorf("expected 'Press ', got '%s'", spans[0].Text)
	}
	if spans[0].Style.Bold {
		t.Error("first span should not be bold")
	}

	// "Enter"
	if spans[1].Text != "Enter" {
		t.Errorf("expected 'Enter', got '%s'", spans[1].Text)
	}
	if !spans[1].Style.Bold {
		t.Error("second span should be bold")
	}
	if spans[1].Style.Foreground != testTheme.Accent {
		t.Error("second span should have Accent color")
	}

	// " to continue"
	if spans[2].Text != " to continue" {
		t.Errorf("expected ' to continue', got '%s'", spans[2].Text)
	}
	if spans[2].Style.Bold {
		t.Error("third span should not be bold")
	}
}

func TestParseMarkup_InvalidTag(t *testing.T) {
	// Unclosed bracket - graceful fallback
	spans := ParseMarkup("Hello [bold World", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "Hello [bold World" {
		t.Errorf("expected literal text, got '%s'", spans[0].Text)
	}
}

func TestParseMarkup_EmptyInput(t *testing.T) {
	spans := ParseMarkup("", testTheme)

	if len(spans) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(spans))
	}
}

func TestParseMarkup_OnlyTags(t *testing.T) {
	spans := ParseMarkup("[bold][/]", testTheme)

	if len(spans) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(spans))
	}
}

func TestParseMarkup_AllThemeColors(t *testing.T) {
	colors := []struct {
		name     string
		expected Color
	}{
		{"$Primary", testTheme.Primary},
		{"$Secondary", testTheme.Secondary},
		{"$Accent", testTheme.Accent},
		{"$Background", testTheme.Background},
		{"$Surface", testTheme.Surface},
		{"$SurfaceHover", testTheme.SurfaceHover},
		{"$Text", testTheme.Text},
		{"$TextMuted", testTheme.TextMuted},
		{"$TextOnPrimary", testTheme.TextOnPrimary},
		{"$Border", testTheme.Border},
		{"$FocusRing", testTheme.FocusRing},
		{"$Error", testTheme.Error},
		{"$Warning", testTheme.Warning},
		{"$Success", testTheme.Success},
		{"$Info", testTheme.Info},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			spans := ParseMarkup("["+tc.name+"]x[/]", testTheme)
			if len(spans) != 1 {
				t.Fatalf("expected 1 span, got %d", len(spans))
			}
			if spans[0].Style.Foreground != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, spans[0].Style.Foreground)
			}
		})
	}
}

func TestParseMarkup_CaseInsensitive(t *testing.T) {
	cases := []string{"$primary", "$PRIMARY", "$Primary", "$pRiMaRy"}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			spans := ParseMarkup("["+c+"]x[/]", testTheme)
			if len(spans) != 1 {
				t.Fatalf("expected 1 span, got %d", len(spans))
			}
			if spans[0].Style.Foreground != testTheme.Primary {
				t.Errorf("expected Primary color for %s", c)
			}
		})
	}
}

func TestParseMarkup_StyleCaseInsensitive(t *testing.T) {
	cases := []string{"[BOLD]", "[Bold]", "[bold]", "[B]", "[b]"}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			spans := ParseMarkup(c+"x[/]", testTheme)
			if len(spans) != 1 {
				t.Fatalf("expected 1 span, got %d", len(spans))
			}
			if !spans[0].Style.Bold {
				t.Errorf("expected bold for %s", c)
			}
		})
	}
}

func TestParseMarkup_WhitespaceInTag(t *testing.T) {
	spans := ParseMarkup("[  bold   $Primary  ]x[/]", testTheme)

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("expected bold")
	}
	if spans[0].Style.Foreground != testTheme.Primary {
		t.Error("expected Primary color")
	}
}

func TestParseMarkup_ComplexNesting(t *testing.T) {
	spans := ParseMarkup("A[bold]B[italic]C[underline]D[/]E[/]F[/]G", testTheme)

	// A - plain
	// B - bold
	// C - bold+italic
	// D - bold+italic+underline
	// E - bold+italic
	// F - bold
	// G - plain

	if len(spans) != 7 {
		t.Fatalf("expected 7 spans, got %d", len(spans))
	}

	// Verify each span
	expectations := []struct {
		text      string
		bold      bool
		italic    bool
		underline UnderlineStyle
	}{
		{"A", false, false, UnderlineNone},
		{"B", true, false, UnderlineNone},
		{"C", true, true, UnderlineNone},
		{"D", true, true, UnderlineSingle},
		{"E", true, true, UnderlineNone},
		{"F", true, false, UnderlineNone},
		{"G", false, false, UnderlineNone},
	}

	for i, exp := range expectations {
		if spans[i].Text != exp.text {
			t.Errorf("span %d: expected text '%s', got '%s'", i, exp.text, spans[i].Text)
		}
		if spans[i].Style.Bold != exp.bold {
			t.Errorf("span %d: expected bold=%v, got %v", i, exp.bold, spans[i].Style.Bold)
		}
		if spans[i].Style.Italic != exp.italic {
			t.Errorf("span %d: expected italic=%v, got %v", i, exp.italic, spans[i].Style.Italic)
		}
		if spans[i].Style.Underline != exp.underline {
			t.Errorf("span %d: expected underline=%v, got %v", i, exp.underline, spans[i].Style.Underline)
		}
	}
}
