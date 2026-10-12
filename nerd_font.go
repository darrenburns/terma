package terma

var nerdFontsEnabled = NewSignal(false)

// SetNerdFontsEnabled turns Nerd Font symbols on or off for the whole app.
// They are off by default. When disabled, [NerdFontSymbol] returns its fallback.
// This setting does not replace glyphs written directly into strings.
//
// Terma does not detect Nerd Font availability: there is no portable terminal
// query for it. Applications should enable it only from a user preference.
// Call this during setup or in an event handler, not during Build.
func SetNerdFontsEnabled(enabled bool) {
	nerdFontsEnabled.Set(enabled)
}

// NerdFontsEnabled reports whether Nerd Font symbols are enabled.
// Reads during Build, layout, or paint subscribe to changes in that phase.
func NerdFontsEnabled() bool {
	return nerdFontsEnabled.Get()
}

// NerdFontSymbol returns symbol when [NerdFontsEnabled] is true, or fallback otherwise.
// Both strings are returned unchanged, so they may include padding, multiple
// characters, or be empty. The result can be used anywhere a string is accepted.
//
// The choice is made on each call. Call from Build to update the widget when
// [SetNerdFontsEnabled] changes the preference, including layout when the two
// strings have different widths. A string resolved during setup stays unchanged.
//
// Example:
//
//	Text{Content: NerdFontSymbol("\uf07b", "[dir]") + " Projects"}
func NerdFontSymbol(symbol, fallback string) string {
	if NerdFontsEnabled() {
		return symbol
	}
	return fallback
}
