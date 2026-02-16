package terma

import (
	"testing"
)

func TestBuiltInGalaxyTheme(t *testing.T) {
	galaxy, ok := GetTheme(ThemeNameGalaxy)
	if !ok {
		t.Fatalf("expected built-in theme %q to be registered", ThemeNameGalaxy)
	}

	if galaxy.Name != ThemeNameGalaxy {
		t.Fatalf("theme name: got %q, want %q", galaxy.Name, ThemeNameGalaxy)
	}
	if galaxy.IsLight {
		t.Fatal("galaxy should be a dark theme")
	}

	if galaxy.Primary != Hex("#C45AFF") {
		t.Errorf("Primary: got %v, want %v", galaxy.Primary, Hex("#C45AFF"))
	}
	if galaxy.Secondary != Hex("#A684E8") {
		t.Errorf("Secondary: got %v, want %v", galaxy.Secondary, Hex("#A684E8"))
	}
	if galaxy.Warning != Hex("#FFD700") {
		t.Errorf("Warning: got %v, want %v", galaxy.Warning, Hex("#FFD700"))
	}
	if galaxy.Error != Hex("#FF4500") {
		t.Errorf("Error: got %v, want %v", galaxy.Error, Hex("#FF4500"))
	}
	if galaxy.Success != Hex("#00FA9A") {
		t.Errorf("Success: got %v, want %v", galaxy.Success, Hex("#00FA9A"))
	}
	if galaxy.Accent != Hex("#FF69B4") {
		t.Errorf("Accent: got %v, want %v", galaxy.Accent, Hex("#FF69B4"))
	}
	if galaxy.Background != Hex("#0F0F1F") {
		t.Errorf("Background: got %v, want %v", galaxy.Background, Hex("#0F0F1F"))
	}
	if galaxy.Surface != Hex("#1E1E3F") {
		t.Errorf("Surface: got %v, want %v", galaxy.Surface, Hex("#1E1E3F"))
	}
	if galaxy.SurfaceHover != Hex("#2D2B55") {
		t.Errorf("SurfaceHover: got %v, want %v", galaxy.SurfaceHover, Hex("#2D2B55"))
	}

	expectedContrastText := galaxy.Background.AutoText()
	if galaxy.Text != expectedContrastText {
		t.Errorf("Text: got %v, want %v (background.AutoText())", galaxy.Text, expectedContrastText)
	}
}

func TestExtendTheme_ValidBase(t *testing.T) {
	extended := ExtendTheme("dracula",
		WithPrimary(Hex("#ff5500")),
		WithAccent(Hex("#00ff00")),
	)

	// Should have the new primary color
	if extended.Primary != Hex("#ff5500") {
		t.Errorf("Primary: got %v, want %v", extended.Primary, Hex("#ff5500"))
	}

	// Should have the new accent color
	if extended.Accent != Hex("#00ff00") {
		t.Errorf("Accent: got %v, want %v", extended.Accent, Hex("#00ff00"))
	}

	// Should preserve other colors from base theme
	dracula, _ := GetTheme("dracula")
	if extended.Secondary != dracula.Secondary {
		t.Errorf("Secondary should be preserved: got %v, want %v", extended.Secondary, dracula.Secondary)
	}
	if extended.Background != dracula.Background {
		t.Errorf("Background should be preserved: got %v, want %v", extended.Background, dracula.Background)
	}
}

func TestExtendTheme_InvalidBase(t *testing.T) {
	extended := ExtendTheme("nonexistent-theme",
		WithPrimary(Hex("#ff5500")),
	)

	// Should return zero value
	if extended.Primary != (Color{}) {
		t.Errorf("Expected zero Color for Primary, got %v", extended.Primary)
	}
	if extended.Name != "" {
		t.Errorf("Expected empty Name, got %q", extended.Name)
	}
}

func TestExtendTheme_NoOptions(t *testing.T) {
	extended := ExtendTheme("dracula")
	dracula, _ := GetTheme("dracula")

	// Should be identical to base theme
	if extended.Primary != dracula.Primary {
		t.Errorf("Primary: got %v, want %v", extended.Primary, dracula.Primary)
	}
	if extended.Background != dracula.Background {
		t.Errorf("Background: got %v, want %v", extended.Background, dracula.Background)
	}
}

func TestExtendTheme_LabelColorsRecomputed(t *testing.T) {
	// Extend with a different background color
	extended := ExtendTheme("dracula",
		WithBackground(Hex("#ffffff")),
		WithPrimary(Hex("#0000ff")),
	)

	// Label colors should be recomputed based on new background
	// PrimaryText is computed as Primary.Blend(Background.AutoText(), 0.5)
	// With white background, auto-text would be dark
	if extended.PrimaryText == (Color{}) {
		t.Error("PrimaryText should be computed, got zero Color")
	}

	// PrimaryBg should be computed
	if extended.PrimaryBg == (Color{}) {
		t.Error("PrimaryBg should be computed, got zero Color")
	}
}

func TestActiveTheme_LabelColorsInitialized(t *testing.T) {
	active := getTheme()
	if active.PrimaryText == (Color{}) {
		t.Fatal("active theme PrimaryText should be initialized")
	}
	if active.PrimaryBg == (Color{}) {
		t.Fatal("active theme PrimaryBg should be initialized")
	}
	if active.SecondaryText == (Color{}) {
		t.Fatal("active theme SecondaryText should be initialized")
	}
	if active.SecondaryBg == (Color{}) {
		t.Fatal("active theme SecondaryBg should be initialized")
	}

	expected, ok := GetTheme(CurrentThemeName())
	if !ok {
		t.Fatalf("current theme %q should exist in registry", CurrentThemeName())
	}
	if active.PrimaryText != expected.PrimaryText {
		t.Fatalf("active PrimaryText mismatch: got %v, want %v", active.PrimaryText, expected.PrimaryText)
	}
	if active.PrimaryBg != expected.PrimaryBg {
		t.Fatalf("active PrimaryBg mismatch: got %v, want %v", active.PrimaryBg, expected.PrimaryBg)
	}
	if active.SecondaryText != expected.SecondaryText {
		t.Fatalf("active SecondaryText mismatch: got %v, want %v", active.SecondaryText, expected.SecondaryText)
	}
	if active.SecondaryBg != expected.SecondaryBg {
		t.Fatalf("active SecondaryBg mismatch: got %v, want %v", active.SecondaryBg, expected.SecondaryBg)
	}
}

func TestExtendAndRegisterTheme_Success(t *testing.T) {
	ok := ExtendAndRegisterTheme("test-custom-theme", "tokyo-night",
		WithPrimary(Hex("#ff0000")),
	)

	if !ok {
		t.Fatal("ExtendAndRegisterTheme returned false for valid base theme")
	}

	// Verify the theme was registered
	registered, found := GetTheme("test-custom-theme")
	if !found {
		t.Fatal("Custom theme was not registered")
	}

	if registered.Primary != Hex("#ff0000") {
		t.Errorf("Primary: got %v, want %v", registered.Primary, Hex("#ff0000"))
	}

	// Name should be set by RegisterTheme
	if registered.Name != "test-custom-theme" {
		t.Errorf("Name: got %q, want %q", registered.Name, "test-custom-theme")
	}

	// Clean up
	delete(themeRegistry, "test-custom-theme")
}

func TestExtendAndRegisterTheme_InvalidBase(t *testing.T) {
	ok := ExtendAndRegisterTheme("test-invalid", "nonexistent-theme",
		WithPrimary(Hex("#ff0000")),
	)

	if ok {
		t.Error("ExtendAndRegisterTheme should return false for invalid base theme")
	}

	// Verify the theme was not registered
	_, found := GetTheme("test-invalid")
	if found {
		t.Error("Invalid theme should not be registered")
		delete(themeRegistry, "test-invalid")
	}
}

func TestAllWithOptions(t *testing.T) {
	// Test that all With* functions correctly set their respective fields
	tests := []struct {
		name     string
		opt      ThemeOption
		validate func(td ThemeData) bool
	}{
		{"WithPrimary", WithPrimary(Hex("#111111")), func(td ThemeData) bool { return td.Primary == Hex("#111111") }},
		{"WithSecondary", WithSecondary(Hex("#222222")), func(td ThemeData) bool { return td.Secondary == Hex("#222222") }},
		{"WithAccent", WithAccent(Hex("#333333")), func(td ThemeData) bool { return td.Accent == Hex("#333333") }},
		{"WithBackground", WithBackground(Hex("#444444")), func(td ThemeData) bool { return td.Background == Hex("#444444") }},
		{"WithSurface", WithSurface(Hex("#555555")), func(td ThemeData) bool { return td.Surface == Hex("#555555") }},
		{"WithSurfaceHover", WithSurfaceHover(Hex("#666666")), func(td ThemeData) bool { return td.SurfaceHover == Hex("#666666") }},
		{"WithSurface2", WithSurface2(Hex("#777777")), func(td ThemeData) bool { return td.Surface2 == Hex("#777777") }},
		{"WithSurface3", WithSurface3(Hex("#888888")), func(td ThemeData) bool { return td.Surface3 == Hex("#888888") }},
		{"WithText", WithText(Hex("#999999")), func(td ThemeData) bool { return td.Text == Hex("#999999") }},
		{"WithTextMuted", WithTextMuted(Hex("#aaaaaa")), func(td ThemeData) bool { return td.TextMuted == Hex("#aaaaaa") }},
		{"WithTextOnPrimary", WithTextOnPrimary(Hex("#bbbbbb")), func(td ThemeData) bool { return td.TextOnPrimary == Hex("#bbbbbb") }},
		{"WithTextOnSecondary", WithTextOnSecondary(Hex("#cccccc")), func(td ThemeData) bool { return td.TextOnSecondary == Hex("#cccccc") }},
		{"WithTextOnAccent", WithTextOnAccent(Hex("#dddddd")), func(td ThemeData) bool { return td.TextOnAccent == Hex("#dddddd") }},
		{"WithTextDisabled", WithTextDisabled(Hex("#eeeeee")), func(td ThemeData) bool { return td.TextDisabled == Hex("#eeeeee") }},
		{"WithBorder", WithBorder(Hex("#112233")), func(td ThemeData) bool { return td.Border == Hex("#112233") }},
		{"WithFocusRing", WithFocusRing(Hex("#223344")), func(td ThemeData) bool { return td.FocusRing == Hex("#223344") }},
		{"WithError", WithError(Hex("#334455")), func(td ThemeData) bool { return td.Error == Hex("#334455") }},
		{"WithWarning", WithWarning(Hex("#445566")), func(td ThemeData) bool { return td.Warning == Hex("#445566") }},
		{"WithSuccess", WithSuccess(Hex("#556677")), func(td ThemeData) bool { return td.Success == Hex("#556677") }},
		{"WithInfo", WithInfo(Hex("#667788")), func(td ThemeData) bool { return td.Info == Hex("#667788") }},
		{"WithTextOnError", WithTextOnError(Hex("#778899")), func(td ThemeData) bool { return td.TextOnError == Hex("#778899") }},
		{"WithTextOnWarning", WithTextOnWarning(Hex("#8899aa")), func(td ThemeData) bool { return td.TextOnWarning == Hex("#8899aa") }},
		{"WithTextOnSuccess", WithTextOnSuccess(Hex("#99aabb")), func(td ThemeData) bool { return td.TextOnSuccess == Hex("#99aabb") }},
		{"WithTextOnInfo", WithTextOnInfo(Hex("#aabbcc")), func(td ThemeData) bool { return td.TextOnInfo == Hex("#aabbcc") }},
		{"WithActiveCursor", WithActiveCursor(Hex("#bbccdd")), func(td ThemeData) bool { return td.ActiveCursor == Hex("#bbccdd") }},
		{"WithSelection", WithSelection(Hex("#ccddee")), func(td ThemeData) bool { return td.Selection == Hex("#ccddee") }},
		{"WithSelectionText", WithSelectionText(Hex("#ddeeff")), func(td ThemeData) bool { return td.SelectionText == Hex("#ddeeff") }},
		{"WithScrollbarTrack", WithScrollbarTrack(Hex("#eeff00")), func(td ThemeData) bool { return td.ScrollbarTrack == Hex("#eeff00") }},
		{"WithScrollbarThumb", WithScrollbarThumb(Hex("#ff0011")), func(td ThemeData) bool { return td.ScrollbarThumb == Hex("#ff0011") }},
		{"WithOverlay", WithOverlay(Hex("#001122")), func(td ThemeData) bool { return td.Overlay == Hex("#001122") }},
		{"WithPlaceholder", WithPlaceholder(Hex("#112200")), func(td ThemeData) bool { return td.Placeholder == Hex("#112200") }},
		{"WithCursor", WithCursor(Hex("#220011")), func(td ThemeData) bool { return td.Cursor == Hex("#220011") }},
		{"WithLink", WithLink(Hex("#003344")), func(td ThemeData) bool { return td.Link == Hex("#003344") }},
		{"WithIsLight", WithIsLight(true), func(td ThemeData) bool { return td.IsLight == true }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extended := ExtendTheme("dracula", tt.opt)
			if !tt.validate(extended) {
				t.Errorf("%s did not set the expected value", tt.name)
			}
		})
	}
}

func TestExtendTheme_DoesNotModifyOriginal(t *testing.T) {
	// Get the original dracula theme
	originalDracula, _ := GetTheme("dracula")
	originalPrimary := originalDracula.Primary

	// Extend with different primary
	_ = ExtendTheme("dracula", WithPrimary(Hex("#ff0000")))

	// Verify original is unchanged
	currentDracula, _ := GetTheme("dracula")
	if currentDracula.Primary != originalPrimary {
		t.Errorf("Original theme was modified: got %v, want %v", currentDracula.Primary, originalPrimary)
	}
}

// =============================================================================
// Snapshot Tests - Visual verification of theme inheritance
// =============================================================================

func TestSnapshot_ThemeInheritance_ExtendedTheme(t *testing.T) {
	// Save original theme to restore after test
	originalThemeName := CurrentThemeName()
	defer SetTheme(originalThemeName)

	// Create a dramatically different theme by extending dracula
	// Use bright, obvious colors that are clearly different from the base
	ExtendAndRegisterTheme("test-neon", "dracula",
		WithPrimary(Hex("#ff0000")),    // Bright red (dracula has purple #bd93f9)
		WithAccent(Hex("#00ff00")),     // Bright green (dracula has cyan #8be9fd)
		WithSuccess(Hex("#ffff00")),    // Yellow (dracula has green #50fa7b)
		WithError(Hex("#ff00ff")),      // Magenta (dracula has red #ff5555)
		WithBackground(Hex("#000033")), // Dark blue background
		WithSurface(Hex("#000066")),    // Lighter blue surface
	)
	defer delete(themeRegistry, "test-neon") // Clean up

	// Set the extended theme as active
	SetTheme("test-neon")

	// Create a widget that displays multiple theme colors prominently
	// This uses buttons with different variants to show Primary, Accent, Success, Error
	widget := Column{
		Width:  Cells(40),
		Height: Cells(11),
		Style: Style{
			Padding: EdgeInsetsAll(1),
		},
		Children: []Widget{
			Text{Content: "Extended Theme Test", Style: Style{Bold: true}},
			Spacer{Height: Cells(1)},
			Row{
				Spacing: 1,
				Children: []Widget{
					Button{ID: "primary", Label: "Primary", Variant: ButtonPrimary},
					Button{ID: "accent", Label: "Accent", Variant: ButtonAccent},
				},
			},
			Spacer{Height: Cells(1)},
			Row{
				Spacing: 1,
				Children: []Widget{
					Button{ID: "success", Label: "Success", Variant: ButtonSuccess},
					Button{ID: "error", Label: "Error", Variant: ButtonError},
				},
			},
		},
	}

	AssertSnapshot(t, widget, 40, 11,
		"Theme inheritance demo: dark blue background (#000033), "+
			"Primary button in bright RED (#ff0000), "+
			"Accent button in bright GREEN (#00ff00), "+
			"Success button in YELLOW (#ffff00), "+
			"Error button in MAGENTA (#ff00ff). "+
			"These colors prove ExtendTheme modified the base dracula theme.")
}
