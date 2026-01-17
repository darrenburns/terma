package terma

// Theme name constants for built-in themes
const (
	ThemeNameRosePine   = "rose-pine"
	ThemeNameDracula    = "dracula"
	ThemeNameTokyoNight = "tokyo-night"
	ThemeNameCatppuccin = "catppuccin"
	ThemeNameGruvbox   = "gruvbox"
	ThemeNameNord      = "nord"
	ThemeNameSolarized = "solarized"
	ThemeNameKanagawa   = "kanagawa"
	ThemeNameMonokai    = "monokai"
)

// ThemeData holds all semantic colors for a theme.
// This is the data structure users provide when registering custom themes.
type ThemeData struct {
	Name string

	// Core branding colors
	Primary   Color
	Secondary Color
	Accent    Color

	// Surface colors (backgrounds)
	Background   Color
	Surface      Color
	SurfaceHover Color

	// Text colors
	Text          Color
	TextMuted     Color
	TextOnPrimary Color

	// Border colors
	Border    Color
	FocusRing Color

	// Feedback colors
	Error   Color
	Warning Color
	Success Color
	Info    Color
}

// Built-in theme definitions

// rosePineThemeData - Soho vibes with muted rose/gold accents
// https://rosepinetheme.com/
var rosePineThemeData = ThemeData{
	Name: ThemeNameRosePine,

	Primary:   Hex("#c4a7e7"), // Iris
	Secondary: Hex("#ebbcba"), // Rose
	Accent:    Hex("#f6c177"), // Gold

	Background:   Hex("#191724"), // Base
	Surface:      Hex("#1f1d2e"), // Surface
	SurfaceHover: Hex("#26233a"), // Overlay

	Text:          Hex("#e0def4"), // Text
	TextMuted:     Hex("#908caa"), // Muted
	TextOnPrimary: Hex("#191724"), // Base (for contrast on primary)

	Border:    Hex("#403d52"), // Highlight Med
	FocusRing: Hex("#c4a7e7"), // Iris

	Error:   Hex("#eb6f92"), // Love
	Warning: Hex("#f6c177"), // Gold
	Success: Hex("#9ccfd8"), // Foam
	Info:    Hex("#31748f"), // Pine
}

// draculaThemeData - Classic dark theme with purple/pink/cyan
// https://draculatheme.com/
var draculaThemeData = ThemeData{
	Name: ThemeNameDracula,

	Primary:   Hex("#bd93f9"), // Purple
	Secondary: Hex("#ff79c6"), // Pink
	Accent:    Hex("#8be9fd"), // Cyan

	Background:   Hex("#282a36"), // Background
	Surface:      Hex("#44475a"), // Current Line
	SurfaceHover: Hex("#6272a4"), // Comment

	Text:          Hex("#f8f8f2"), // Foreground
	TextMuted:     Hex("#8b91a8"), // Lightened comment for better visibility
	TextOnPrimary: Hex("#282a36"), // Background (for contrast)

	Border:    Hex("#44475a"), // Current Line
	FocusRing: Hex("#bd93f9"), // Purple

	Error:   Hex("#ff5555"), // Red
	Warning: Hex("#ffb86c"), // Orange
	Success: Hex("#50fa7b"), // Green
	Info:    Hex("#8be9fd"), // Cyan
}

// tokyoNightThemeData - Cool blues and purples inspired by Tokyo nights
// https://github.com/enkia/tokyo-night-vscode-theme
var tokyoNightThemeData = ThemeData{
	Name: ThemeNameTokyoNight,

	Primary:   Hex("#7aa2f7"), // Blue
	Secondary: Hex("#bb9af7"), // Purple
	Accent:    Hex("#7dcfff"), // Cyan

	Background:   Hex("#1a1b26"), // Background
	Surface:      Hex("#24283b"), // Surface
	SurfaceHover: Hex("#414868"), // Surface Hover

	Text:          Hex("#c0caf5"), // Foreground
	TextMuted:     Hex("#737aa2"), // Lightened comment for better visibility
	TextOnPrimary: Hex("#1a1b26"), // Background (for contrast)

	Border:    Hex("#414868"), // Border
	FocusRing: Hex("#7aa2f7"), // Blue

	Error:   Hex("#f7768e"), // Red
	Warning: Hex("#e0af68"), // Yellow
	Success: Hex("#9ece6a"), // Green
	Info:    Hex("#7dcfff"), // Cyan
}

// catppuccinThemeData - Soothing pastel theme (Mocha flavor)
// https://catppuccin.com/
var catppuccinThemeData = ThemeData{
	Name: ThemeNameCatppuccin,

	Primary:   Hex("#cba6f7"), // Mauve
	Secondary: Hex("#f5c2e7"), // Pink
	Accent:    Hex("#94e2d5"), // Teal

	Background:   Hex("#1e1e2e"), // Base
	Surface:      Hex("#313244"), // Surface0
	SurfaceHover: Hex("#45475a"), // Surface1

	Text:          Hex("#cdd6f4"), // Text
	TextMuted:     Hex("#9399b2"), // Overlay2 - more muted
	TextOnPrimary: Hex("#1e1e2e"), // Base

	Border:    Hex("#45475a"), // Surface1
	FocusRing: Hex("#cba6f7"), // Mauve

	Error:   Hex("#f38ba8"), // Red
	Warning: Hex("#fab387"), // Peach
	Success: Hex("#a6e3a1"), // Green
	Info:    Hex("#89b4fa"), // Blue
}

// gruvboxThemeData - Retro groove with warm earthy colors
// https://github.com/morhetz/gruvbox
var gruvboxThemeData = ThemeData{
	Name: ThemeNameGruvbox,

	Primary:   Hex("#d79921"), // Yellow
	Secondary: Hex("#d3869b"), // Purple
	Accent:    Hex("#8ec07c"), // Aqua

	Background:   Hex("#282828"), // bg0
	Surface:      Hex("#3c3836"), // bg1
	SurfaceHover: Hex("#504945"), // bg2

	Text:          Hex("#ebdbb2"), // fg1
	TextMuted:     Hex("#a89984"), // gray
	TextOnPrimary: Hex("#282828"), // bg0

	Border:    Hex("#504945"), // bg2
	FocusRing: Hex("#d79921"), // yellow

	Error:   Hex("#fb4934"), // red
	Warning: Hex("#fe8019"), // orange
	Success: Hex("#b8bb26"), // green
	Info:    Hex("#83a598"), // blue
}

// nordThemeData - Arctic, north-bluish color palette
// https://www.nordtheme.com/
var nordThemeData = ThemeData{
	Name: ThemeNameNord,

	Primary:   Hex("#88c0d0"), // Nord8 - frost
	Secondary: Hex("#81a1c1"), // Nord9 - frost
	Accent:    Hex("#8fbcbb"), // Nord7 - frost

	Background:   Hex("#2e3440"), // Nord0 - polar night
	Surface:      Hex("#3b4252"), // Nord1 - polar night
	SurfaceHover: Hex("#434c5e"), // Nord2 - polar night

	Text:          Hex("#eceff4"), // Nord6 - snow storm
	TextMuted:     Hex("#7b88a1"), // Blend between polar night and snow storm
	TextOnPrimary: Hex("#2e3440"), // Nord0

	Border:    Hex("#4c566a"), // Nord3 - polar night
	FocusRing: Hex("#88c0d0"), // Nord8

	Error:   Hex("#bf616a"), // Nord11 - aurora red
	Warning: Hex("#ebcb8b"), // Nord13 - aurora yellow
	Success: Hex("#a3be8c"), // Nord14 - aurora green
	Info:    Hex("#5e81ac"), // Nord10 - frost
}

// solarizedThemeData - Precision colors for machines and people (Dark)
// https://ethanschoonover.com/solarized/
var solarizedThemeData = ThemeData{
	Name: ThemeNameSolarized,

	Primary:   Hex("#268bd2"), // Blue
	Secondary: Hex("#6c71c4"), // Violet
	Accent:    Hex("#2aa198"), // Cyan

	Background:   Hex("#002b36"), // base03
	Surface:      Hex("#073642"), // base02
	SurfaceHover: Hex("#586e75"), // base01

	Text:          Hex("#839496"), // base0
	TextMuted:     Hex("#657b83"), // base00
	TextOnPrimary: Hex("#fdf6e3"), // base3

	Border:    Hex("#073642"), // base02
	FocusRing: Hex("#268bd2"), // blue

	Error:   Hex("#dc322f"), // red
	Warning: Hex("#b58900"), // yellow
	Success: Hex("#859900"), // green
	Info:    Hex("#2aa198"), // cyan
}

// kanagawaThemeData - Dark theme inspired by Katsushika Hokusai's famous wave painting
// https://github.com/rebelot/kanagawa.nvim
var kanagawaThemeData = ThemeData{
	Name: ThemeNameKanagawa,

	Primary:   Hex("#7e9cd8"), // crystalBlue
	Secondary: Hex("#957fb8"), // oniViolet
	Accent:    Hex("#7aa89f"), // waveAqua2

	Background:   Hex("#1f1f28"), // sumiInk1
	Surface:      Hex("#2a2a37"), // sumiInk3
	SurfaceHover: Hex("#363646"), // sumiInk4

	Text:          Hex("#dcd7ba"), // fujiWhite
	TextMuted:     Hex("#9a9a8e"), // Lightened fujiGray for better visibility
	TextOnPrimary: Hex("#1f1f28"), // sumiInk1

	Border:    Hex("#54546d"), // sumiInk6
	FocusRing: Hex("#7e9cd8"), // crystalBlue

	Error:   Hex("#e82424"), // samuraiRed
	Warning: Hex("#ff9e3b"), // roninYellow
	Success: Hex("#98bb6c"), // springGreen
	Info:    Hex("#7fb4ca"), // springBlue
}

// monokaiThemeData - Iconic theme from Sublime Text
// https://monokai.pro/
var monokaiThemeData = ThemeData{
	Name: ThemeNameMonokai,

	Primary:   Hex("#a6e22e"), // Green
	Secondary: Hex("#ae81ff"), // Purple
	Accent:    Hex("#66d9ef"), // Cyan

	Background:   Hex("#272822"), // Background
	Surface:      Hex("#3e3d32"), // Line highlight
	SurfaceHover: Hex("#49483e"), // Selection

	Text:          Hex("#f8f8f2"), // Foreground
	TextMuted:     Hex("#a59f85"), // Lightened comment for better visibility
	TextOnPrimary: Hex("#272822"), // Background

	Border:    Hex("#49483e"), // Selection
	FocusRing: Hex("#a6e22e"), // Green

	Error:   Hex("#f92672"), // Pink/Red
	Warning: Hex("#fd971f"), // Orange
	Success: Hex("#a6e22e"), // Green
	Info:    Hex("#66d9ef"), // Cyan
}

// themeRegistry holds all registered themes
var themeRegistry = map[string]ThemeData{
	ThemeNameRosePine:   rosePineThemeData,
	ThemeNameDracula:    draculaThemeData,
	ThemeNameTokyoNight: tokyoNightThemeData,
	ThemeNameCatppuccin: catppuccinThemeData,
	ThemeNameGruvbox:   gruvboxThemeData,
	ThemeNameNord:      nordThemeData,
	ThemeNameSolarized: solarizedThemeData,
	ThemeNameKanagawa:   kanagawaThemeData,
	ThemeNameMonokai:    monokaiThemeData,
}

// activeTheme is the signal holding the current theme
var activeTheme = NewAnySignal(rosePineThemeData)

// activeThemeName tracks the current theme name
var activeThemeName = ThemeNameRosePine

// SetTheme switches to the theme with the given name.
// If the theme is not found, this logs a warning and does nothing.
func SetTheme(name string) {
	data, ok := themeRegistry[name]
	if !ok {
		Log("Theme not found: %s", name)
		return
	}
	activeThemeName = name
	activeTheme.Set(data)
}

// RegisterTheme registers a custom theme with the given name.
// If a theme with this name already exists, it is replaced.
// If this is the currently active theme, the change takes effect immediately.
func RegisterTheme(name string, data ThemeData) {
	data.Name = name
	themeRegistry[name] = data
	// If this is the active theme, update it
	if name == activeThemeName {
		activeTheme.Set(data)
	}
}

// CurrentThemeName returns the name of the currently active theme.
func CurrentThemeName() string {
	return activeThemeName
}

// ThemeNames returns a slice of all registered theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	return names
}

// GetTheme returns the ThemeData for the given theme name.
// Returns the theme data and true if found, or zero value and false if not found.
func GetTheme(name string) (ThemeData, bool) {
	data, ok := themeRegistry[name]
	return data, ok
}

// getTheme returns the ThemeData for the active theme.
// This is called internally by BuildContext.Theme().
func getTheme() ThemeData {
	return activeTheme.Get()
}
