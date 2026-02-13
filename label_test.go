package terma

import "testing"

func TestLabelSecondaryUsesSecondaryLabelColors(t *testing.T) {
	theme := getTheme()
	label := Label("secondary", LabelSecondary, theme)

	if label.Style.ForegroundColor != theme.SecondaryText {
		t.Fatalf("foreground: got %v, want %v", label.Style.ForegroundColor, theme.SecondaryText)
	}
	if label.Style.BackgroundColor != theme.SecondaryBg {
		t.Fatalf("background: got %v, want %v", label.Style.BackgroundColor, theme.SecondaryBg)
	}
}
