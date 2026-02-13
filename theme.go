package terma

import "sort"

// Theme name constants for built-in themes
const (
	// Dark themes
	ThemeNameRosePine   = "rose-pine"
	ThemeNameDracula    = "dracula"
	ThemeNameTokyoNight = "tokyo-night"
	ThemeNameCatppuccin = "catppuccin"
	ThemeNameGruvbox    = "gruvbox"
	ThemeNameNord       = "nord"
	ThemeNameSolarized  = "solarized"
	ThemeNameKanagawa   = "kanagawa"
	ThemeNameMonokai    = "monokai"

	// Light themes
	ThemeNameRosePineDawn    = "rose-pine-dawn"
	ThemeNameDraculaLight    = "dracula-light"
	ThemeNameTokyoNightDay   = "tokyo-night-day"
	ThemeNameCatppuccinLatte = "catppuccin-latte"
	ThemeNameGruvboxLight    = "gruvbox-light"
	ThemeNameNordLight       = "nord-light"
	ThemeNameSolarizedLight  = "solarized-light"
	ThemeNameKanagawaLotus   = "kanagawa-lotus"
	ThemeNameMonokaiLight    = "monokai-light"
)

const DefaultSelectionAlpha = 0.25

// ThemeData holds all semantic colors for a theme.
// This is the data structure users provide when registering custom themes.
type ThemeData struct {
	Name    string
	IsLight bool // True for light themes, false for dark themes

	// Core branding colors
	Primary   Color
	Secondary Color
	Accent    Color

	// Surface colors (backgrounds)
	Background   Color
	Surface      Color
	SurfaceHover Color
	Surface2     Color // Level 2 (nested elements)
	Surface3     Color // Level 3 (deeply nested)

	// Text colors
	Text            Color
	TextMuted       Color
	TextOnPrimary   Color
	TextOnSecondary Color
	TextOnAccent    Color
	TextDisabled    Color // Disabled state text

	// Border colors
	Border    Color
	FocusRing Color

	// Feedback colors
	Error   Color
	Warning Color
	Success Color
	Info    Color

	// Feedback text variants
	TextOnError   Color
	TextOnWarning Color
	TextOnSuccess Color
	TextOnInfo    Color

	// ActiveCursor colors
	ActiveCursor  Color // Active selection background (cursor/focused item)
	Selection     Color // Dimmer selection background (multi-select without focus)
	SelectionText Color // Text on selection

	// Scrollbar colors
	ScrollbarTrack Color
	ScrollbarThumb Color

	// Overlay/modal
	Overlay Color // Semi-transparent backdrop

	// Input-specific
	Placeholder Color // Placeholder text
	Cursor      Color // Text cursor/caret

	// Link
	Link Color // Clickable text links

	// Label text colors (variant colors blended toward readable text)
	PrimaryText   Color
	SecondaryText Color
	AccentText    Color
	SuccessText   Color
	ErrorText     Color
	WarningText   Color
	InfoText      Color

	// Label background colors (variant colors faded/dimmed)
	PrimaryBg   Color
	SecondaryBg Color
	AccentBg    Color
	SuccessBg   Color
	ErrorBg     Color
	WarningBg   Color
	InfoBg      Color
}

// computeLabelColors fills in derived label colors from base variant colors.
func computeLabelColors(data *ThemeData) {
	autoText := data.Background.AutoText()

	// Text: 50% variant, 50% auto-text for readability with more color
	data.PrimaryText = data.Primary.Blend(autoText, 0.5)
	data.SecondaryText = data.Secondary.Blend(autoText, 0.5)
	data.AccentText = data.Accent.Blend(autoText, 0.5)
	data.SuccessText = data.Success.Blend(autoText, 0.5)
	data.ErrorText = data.Error.Blend(autoText, 0.5)
	data.WarningText = data.Warning.Blend(autoText, 0.5)
	data.InfoText = data.Info.Blend(autoText, 0.5)

	// Background: 35% variant blended into background
	data.PrimaryBg = data.Background.Blend(data.Primary, 0.35)
	data.SecondaryBg = data.Background.Blend(data.Secondary, 0.35)
	data.AccentBg = data.Background.Blend(data.Accent, 0.35)
	data.SuccessBg = data.Background.Blend(data.Success, 0.35)
	data.ErrorBg = data.Background.Blend(data.Error, 0.35)
	data.WarningBg = data.Background.Blend(data.Warning, 0.35)
	data.InfoBg = data.Background.Blend(data.Info, 0.35)
}

func init() {
	for name, theme := range themeRegistry {
		computeLabelColors(&theme)
		themeRegistry[name] = theme
	}
	// Ensure the initial active theme includes derived label colors.
	if active, ok := themeRegistry[activeThemeName]; ok {
		activeTheme.Set(active)
	}
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
	Surface2:     Hex("#2a273f"), // Slightly lighter overlay
	Surface3:     Hex("#312e4a"), // Even lighter

	Text:            Hex("#e0def4"), // Text
	TextMuted:       Hex("#908caa"), // Muted
	TextOnPrimary:   Hex("#191724"), // Base (for contrast on primary)
	TextOnSecondary: Hex("#191724"), // Base
	TextOnAccent:    Hex("#191724"), // Base
	TextDisabled:    Hex("#6e6a86"), // Subtle

	Border:    Hex("#403d52"), // Highlight Med
	FocusRing: Hex("#c4a7e7"), // Iris

	Error:   Hex("#eb6f92"), // Love
	Warning: Hex("#f6c177"), // Gold
	Success: Hex("#9ccfd8"), // Foam
	Info:    Hex("#31748f"), // Pine

	TextOnError:   Hex("#191724"), // Base
	TextOnWarning: Hex("#191724"), // Base
	TextOnSuccess: Hex("#191724"), // Base
	TextOnInfo:    Hex("#e0def4"), // Text (Pine is darker)

	ActiveCursor:  Hex("#f6c177"),                                  // Accent (Gold)
	Selection:     Hex("#f6c177").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#191724"),                                  // TextOnAccent (Base)

	ScrollbarTrack: Hex("#26233a"), // Overlay
	ScrollbarThumb: Hex("#6e6a86"), // Subtle

	Overlay: Hex("#191724").WithAlpha(0.8), // Base with transparency

	Placeholder: Hex("#6e6a86"), // Subtle
	Cursor:      Hex("#e0def4"), // Text

	Link: Hex("#c4a7e7"), // Iris
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
	Surface2:     Hex("#4d5066"), // Lighter current line
	Surface3:     Hex("#565973"), // Even lighter

	Text:            Hex("#f8f8f2"), // Foreground
	TextMuted:       Hex("#8b91a8"), // Lightened comment for better visibility
	TextOnPrimary:   Hex("#282a36"), // Background (for contrast)
	TextOnSecondary: Hex("#282a36"), // Background
	TextOnAccent:    Hex("#282a36"), // Background
	TextDisabled:    Hex("#6272a4"), // Comment

	Border:    Hex("#44475a"), // Current Line
	FocusRing: Hex("#bd93f9"), // Purple

	Error:   Hex("#ff5555"), // Red
	Warning: Hex("#ffb86c"), // Orange
	Success: Hex("#50fa7b"), // Green
	Info:    Hex("#8be9fd"), // Cyan

	TextOnError:   Hex("#282a36"), // Background
	TextOnWarning: Hex("#282a36"), // Background
	TextOnSuccess: Hex("#282a36"), // Background
	TextOnInfo:    Hex("#282a36"), // Background

	ActiveCursor:  Hex("#8be9fd"),                                  // Accent (Cyan)
	Selection:     Hex("#8be9fd").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#282a36"),                                  // TextOnAccent (Background)

	ScrollbarTrack: Hex("#44475a"), // Current Line
	ScrollbarThumb: Hex("#6272a4"), // Comment

	Overlay: Hex("#282a36").WithAlpha(0.8), // Background with transparency

	Placeholder: Hex("#6272a4"), // Comment
	Cursor:      Hex("#f8f8f2"), // Foreground

	Link: Hex("#8be9fd"), // Cyan
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
	Surface2:     Hex("#2f3549"), // Lighter surface
	Surface3:     Hex("#3b4261"), // Even lighter

	Text:            Hex("#c0caf5"), // Foreground
	TextMuted:       Hex("#737aa2"), // Lightened comment for better visibility
	TextOnPrimary:   Hex("#1a1b26"), // Background (for contrast)
	TextOnSecondary: Hex("#1a1b26"), // Background
	TextOnAccent:    Hex("#1a1b26"), // Background
	TextDisabled:    Hex("#565f89"), // Comment

	Border:    Hex("#414868"), // Border
	FocusRing: Hex("#7aa2f7"), // Blue

	Error:   Hex("#f7768e"), // Red
	Warning: Hex("#e0af68"), // Yellow
	Success: Hex("#9ece6a"), // Green
	Info:    Hex("#7dcfff"), // Cyan

	TextOnError:   Hex("#1a1b26"), // Background
	TextOnWarning: Hex("#1a1b26"), // Background
	TextOnSuccess: Hex("#1a1b26"), // Background
	TextOnInfo:    Hex("#1a1b26"), // Background

	ActiveCursor:  Hex("#7dcfff"),                                  // Accent (Cyan)
	Selection:     Hex("#7dcfff").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#1a1b26"),                                  // TextOnAccent (Background)

	ScrollbarTrack: Hex("#24283b"), // Surface
	ScrollbarThumb: Hex("#565f89"), // Comment

	Overlay: Hex("#1a1b26").WithAlpha(0.8), // Background with transparency

	Placeholder: Hex("#565f89"), // Comment
	Cursor:      Hex("#c0caf5"), // Foreground

	Link: Hex("#7dcfff"), // Cyan
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
	Surface2:     Hex("#585b70"), // Surface2
	Surface3:     Hex("#6c7086"), // Overlay0

	Text:            Hex("#cdd6f4"), // Text
	TextMuted:       Hex("#9399b2"), // Overlay2 - more muted
	TextOnPrimary:   Hex("#1e1e2e"), // Base
	TextOnSecondary: Hex("#1e1e2e"), // Base
	TextOnAccent:    Hex("#1e1e2e"), // Base
	TextDisabled:    Hex("#6c7086"), // Overlay0

	Border:    Hex("#45475a"), // Surface1
	FocusRing: Hex("#cba6f7"), // Mauve

	Error:   Hex("#f38ba8"), // Red
	Warning: Hex("#fab387"), // Peach
	Success: Hex("#a6e3a1"), // Green
	Info:    Hex("#89b4fa"), // Blue

	TextOnError:   Hex("#1e1e2e"), // Base
	TextOnWarning: Hex("#1e1e2e"), // Base
	TextOnSuccess: Hex("#1e1e2e"), // Base
	TextOnInfo:    Hex("#1e1e2e"), // Base

	ActiveCursor:  Hex("#94e2d5"),                                  // Accent (Teal)
	Selection:     Hex("#94e2d5").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#1e1e2e"),                                  // TextOnAccent (Base)

	ScrollbarTrack: Hex("#313244"), // Surface0
	ScrollbarThumb: Hex("#6c7086"), // Overlay0

	Overlay: Hex("#1e1e2e").WithAlpha(0.8), // Base with transparency

	Placeholder: Hex("#6c7086"), // Overlay0
	Cursor:      Hex("#cdd6f4"), // Text

	Link: Hex("#89b4fa"), // Blue
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
	Surface2:     Hex("#665c54"), // bg3
	Surface3:     Hex("#7c6f64"), // bg4

	Text:            Hex("#ebdbb2"), // fg1
	TextMuted:       Hex("#a89984"), // gray
	TextOnPrimary:   Hex("#282828"), // bg0
	TextOnSecondary: Hex("#282828"), // bg0
	TextOnAccent:    Hex("#282828"), // bg0
	TextDisabled:    Hex("#7c6f64"), // bg4

	Border:    Hex("#504945"), // bg2
	FocusRing: Hex("#d79921"), // yellow

	Error:   Hex("#fb4934"), // red
	Warning: Hex("#fe8019"), // orange
	Success: Hex("#b8bb26"), // green
	Info:    Hex("#83a598"), // blue

	TextOnError:   Hex("#282828"), // bg0
	TextOnWarning: Hex("#282828"), // bg0
	TextOnSuccess: Hex("#282828"), // bg0
	TextOnInfo:    Hex("#282828"), // bg0

	ActiveCursor:  Hex("#8ec07c"),                                  // Accent (Aqua)
	Selection:     Hex("#8ec07c").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#282828"),                                  // TextOnAccent (bg0)

	ScrollbarTrack: Hex("#3c3836"), // bg1
	ScrollbarThumb: Hex("#7c6f64"), // bg4

	Overlay: Hex("#282828").WithAlpha(0.8), // bg0 with transparency

	Placeholder: Hex("#7c6f64"), // bg4
	Cursor:      Hex("#ebdbb2"), // fg1

	Link: Hex("#83a598"), // blue
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
	Surface2:     Hex("#4c566a"), // Nord3 - polar night
	Surface3:     Hex("#616e88"), // Lighter polar night

	Text:            Hex("#eceff4"), // Nord6 - snow storm
	TextMuted:       Hex("#7b88a1"), // Blend between polar night and snow storm
	TextOnPrimary:   Hex("#2e3440"), // Nord0
	TextOnSecondary: Hex("#2e3440"), // Nord0
	TextOnAccent:    Hex("#2e3440"), // Nord0
	TextDisabled:    Hex("#4c566a"), // Nord3

	Border:    Hex("#4c566a"), // Nord3 - polar night
	FocusRing: Hex("#88c0d0"), // Nord8

	Error:   Hex("#bf616a"), // Nord11 - aurora red
	Warning: Hex("#ebcb8b"), // Nord13 - aurora yellow
	Success: Hex("#a3be8c"), // Nord14 - aurora green
	Info:    Hex("#5e81ac"), // Nord10 - frost

	TextOnError:   Hex("#2e3440"), // Nord0
	TextOnWarning: Hex("#2e3440"), // Nord0
	TextOnSuccess: Hex("#2e3440"), // Nord0
	TextOnInfo:    Hex("#eceff4"), // Nord6 (frost blue is darker)

	ActiveCursor:  Hex("#8fbcbb"),                                  // Accent (Nord7)
	Selection:     Hex("#8fbcbb").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#2e3440"),                                  // TextOnAccent (Nord0)

	ScrollbarTrack: Hex("#3b4252"), // Nord1
	ScrollbarThumb: Hex("#4c566a"), // Nord3

	Overlay: Hex("#2e3440").WithAlpha(0.8), // Nord0 with transparency

	Placeholder: Hex("#4c566a"), // Nord3
	Cursor:      Hex("#eceff4"), // Nord6

	Link: Hex("#88c0d0"), // Nord8
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
	Surface2:     Hex("#657b83"), // base00
	Surface3:     Hex("#839496"), // base0

	Text:            Hex("#839496"), // base0
	TextMuted:       Hex("#657b83"), // base00
	TextOnPrimary:   Hex("#fdf6e3"), // base3
	TextOnSecondary: Hex("#fdf6e3"), // base3
	TextOnAccent:    Hex("#fdf6e3"), // base3
	TextDisabled:    Hex("#586e75"), // base01

	Border:    Hex("#073642"), // base02
	FocusRing: Hex("#268bd2"), // blue

	Error:   Hex("#dc322f"), // red
	Warning: Hex("#b58900"), // yellow
	Success: Hex("#859900"), // green
	Info:    Hex("#2aa198"), // cyan

	TextOnError:   Hex("#fdf6e3"), // base3
	TextOnWarning: Hex("#fdf6e3"), // base3
	TextOnSuccess: Hex("#fdf6e3"), // base3
	TextOnInfo:    Hex("#fdf6e3"), // base3

	ActiveCursor:  Hex("#2aa198"),                                  // Accent (Cyan)
	Selection:     Hex("#2aa198").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#fdf6e3"),                                  // TextOnAccent (base3)

	ScrollbarTrack: Hex("#073642"), // base02
	ScrollbarThumb: Hex("#586e75"), // base01

	Overlay: Hex("#002b36").WithAlpha(0.8), // base03 with transparency

	Placeholder: Hex("#586e75"), // base01
	Cursor:      Hex("#839496"), // base0

	Link: Hex("#268bd2"), // blue
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
	Surface2:     Hex("#43434f"), // Lighter sumiInk
	Surface3:     Hex("#54546d"), // sumiInk6

	Text:            Hex("#dcd7ba"), // fujiWhite
	TextMuted:       Hex("#9a9a8e"), // Lightened fujiGray for better visibility
	TextOnPrimary:   Hex("#1f1f28"), // sumiInk1
	TextOnSecondary: Hex("#1f1f28"), // sumiInk1
	TextOnAccent:    Hex("#1f1f28"), // sumiInk1
	TextDisabled:    Hex("#727169"), // fujiGray

	Border:    Hex("#54546d"), // sumiInk6
	FocusRing: Hex("#7e9cd8"), // crystalBlue

	Error:   Hex("#e82424"), // samuraiRed
	Warning: Hex("#ff9e3b"), // roninYellow
	Success: Hex("#98bb6c"), // springGreen
	Info:    Hex("#7fb4ca"), // springBlue

	TextOnError:   Hex("#dcd7ba"), // fujiWhite
	TextOnWarning: Hex("#1f1f28"), // sumiInk1
	TextOnSuccess: Hex("#1f1f28"), // sumiInk1
	TextOnInfo:    Hex("#1f1f28"), // sumiInk1

	ActiveCursor:  Hex("#7aa89f"),                                  // Accent (waveAqua2)
	Selection:     Hex("#7aa89f").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#1f1f28"),                                  // TextOnAccent (sumiInk1)

	ScrollbarTrack: Hex("#2a2a37"), // sumiInk3
	ScrollbarThumb: Hex("#54546d"), // sumiInk6

	Overlay: Hex("#1f1f28").WithAlpha(0.8), // sumiInk1 with transparency

	Placeholder: Hex("#727169"), // fujiGray
	Cursor:      Hex("#dcd7ba"), // fujiWhite

	Link: Hex("#7fb4ca"), // springBlue
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
	SurfaceHover: Hex("#49483e"), // ActiveCursor
	Surface2:     Hex("#555549"), // Lighter selection
	Surface3:     Hex("#625f54"), // Even lighter

	Text:            Hex("#f8f8f2"), // Foreground
	TextMuted:       Hex("#a59f85"), // Lightened comment for better visibility
	TextOnPrimary:   Hex("#272822"), // Background
	TextOnSecondary: Hex("#272822"), // Background
	TextOnAccent:    Hex("#272822"), // Background
	TextDisabled:    Hex("#75715e"), // Comment

	Border:    Hex("#49483e"), // ActiveCursor
	FocusRing: Hex("#a6e22e"), // Green

	Error:   Hex("#f92672"), // Pink/Red
	Warning: Hex("#fd971f"), // Orange
	Success: Hex("#a6e22e"), // Green
	Info:    Hex("#66d9ef"), // Cyan

	TextOnError:   Hex("#272822"), // Background
	TextOnWarning: Hex("#272822"), // Background
	TextOnSuccess: Hex("#272822"), // Background
	TextOnInfo:    Hex("#272822"), // Background

	ActiveCursor:  Hex("#66d9ef"),                                  // Accent (Cyan)
	Selection:     Hex("#66d9ef").WithAlpha(DefaultSelectionAlpha), // Accent with alpha for multi-select
	SelectionText: Hex("#272822"),                                  // TextOnAccent (Background)

	ScrollbarTrack: Hex("#3e3d32"), // Line highlight
	ScrollbarThumb: Hex("#75715e"), // Comment

	Overlay: Hex("#272822").WithAlpha(0.8), // Background with transparency

	Placeholder: Hex("#75715e"), // Comment
	Cursor:      Hex("#f8f8f2"), // Foreground

	Link: Hex("#66d9ef"), // Cyan
}

// ============================================================================
// Light Theme Definitions
// ============================================================================

// rosePineDawnThemeData - Rosé Pine Dawn light variant
// https://rosepinetheme.com/
var rosePineDawnThemeData = ThemeData{
	Name:    ThemeNameRosePineDawn,
	IsLight: true,

	Primary:   Hex("#907aa9"), // Iris
	Secondary: Hex("#d7827e"), // Rose
	Accent:    Hex("#ea9d34"), // Gold

	Background:   Hex("#faf4ed"), // Base
	Surface:      Hex("#fffaf3"), // Surface
	SurfaceHover: Hex("#f2e9e1"), // Overlay
	Surface2:     Hex("#e4dcd5"), // Slightly darker
	Surface3:     Hex("#d7cfc8"), // Even darker

	Text:            Hex("#575279"), // Text
	TextMuted:       Hex("#9893a5"), // Muted
	TextOnPrimary:   Hex("#faf4ed"), // Base
	TextOnSecondary: Hex("#faf4ed"), // Base
	TextOnAccent:    Hex("#faf4ed"), // Base
	TextDisabled:    Hex("#b4aeb8"), // Subtle

	Border:    Hex("#dfdad9"), // Highlight Med
	FocusRing: Hex("#907aa9"), // Iris

	Error:   Hex("#b4637a"), // Love
	Warning: Hex("#ea9d34"), // Gold
	Success: Hex("#56949f"), // Foam
	Info:    Hex("#286983"), // Pine

	TextOnError:   Hex("#faf4ed"), // Base
	TextOnWarning: Hex("#faf4ed"), // Base
	TextOnSuccess: Hex("#faf4ed"), // Base
	TextOnInfo:    Hex("#faf4ed"), // Base

	ActiveCursor:  Hex("#ea9d34"),                 // Accent (Gold)
	Selection:     Hex("#ea9d34").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#faf4ed"),                 // TextOnAccent (Base)

	ScrollbarTrack: Hex("#f2e9e1"), // Overlay
	ScrollbarThumb: Hex("#b4aeb8"), // Subtle

	Overlay: Hex("#575279").WithAlpha(0.5), // Text with transparency

	Placeholder: Hex("#b4aeb8"), // Subtle
	Cursor:      Hex("#575279"), // Text

	Link: Hex("#907aa9"), // Iris
}

// draculaLightThemeData - Dracula light variant (Alucard)
var draculaLightThemeData = ThemeData{
	Name:    ThemeNameDraculaLight,
	IsLight: true,

	Primary:   Hex("#9580ff"), // Purple
	Secondary: Hex("#ff80bf"), // Pink
	Accent:    Hex("#80ffea"), // Cyan (darkened for light bg)

	Background:   Hex("#f8f8f2"), // Light background
	Surface:      Hex("#f0f0ea"), // Slightly darker
	SurfaceHover: Hex("#e8e8e0"), // Even darker
	Surface2:     Hex("#e0e0d8"), // More contrast
	Surface3:     Hex("#d8d8d0"), // Even more

	Text:            Hex("#282a36"), // Dark text
	TextMuted:       Hex("#6272a4"), // Comment
	TextOnPrimary:   Hex("#f8f8f2"), // Light
	TextOnSecondary: Hex("#f8f8f2"), // Light
	TextOnAccent:    Hex("#282a36"), // Dark (cyan is light)
	TextDisabled:    Hex("#8b91a8"), // Muted

	Border:    Hex("#d8d8d0"), // Light border
	FocusRing: Hex("#9580ff"), // Purple

	Error:   Hex("#ff5555"), // Red
	Warning: Hex("#ffb86c"), // Orange
	Success: Hex("#50fa7b"), // Green
	Info:    Hex("#8be9fd"), // Cyan

	TextOnError:   Hex("#f8f8f2"), // Light
	TextOnWarning: Hex("#282a36"), // Dark
	TextOnSuccess: Hex("#282a36"), // Dark
	TextOnInfo:    Hex("#282a36"), // Dark

	ActiveCursor:  Hex("#80ffea"),                 // Accent (Cyan)
	Selection:     Hex("#80ffea").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#282a36"),                 // TextOnAccent (Dark)

	ScrollbarTrack: Hex("#e8e8e0"), // Surface hover
	ScrollbarThumb: Hex("#6272a4"), // Comment

	Overlay: Hex("#282a36").WithAlpha(0.5), // Dark with transparency

	Placeholder: Hex("#8b91a8"), // Muted
	Cursor:      Hex("#282a36"), // Dark text

	Link: Hex("#9580ff"), // Purple
}

// tokyoNightDayThemeData - Tokyo Night Day light variant
// https://github.com/enkia/tokyo-night-vscode-theme
var tokyoNightDayThemeData = ThemeData{
	Name:    ThemeNameTokyoNightDay,
	IsLight: true,

	Primary:   Hex("#2e7de9"), // Blue
	Secondary: Hex("#9854f1"), // Purple
	Accent:    Hex("#007197"), // Cyan

	Background:   Hex("#e1e2e7"), // Background
	Surface:      Hex("#d5d6db"), // Surface
	SurfaceHover: Hex("#c9cad0"), // Surface hover
	Surface2:     Hex("#bdbec4"), // Darker
	Surface3:     Hex("#b1b2b8"), // Even darker

	Text:            Hex("#3760bf"), // Foreground
	TextMuted:       Hex("#848cb5"), // Comment
	TextOnPrimary:   Hex("#e1e2e7"), // Background
	TextOnSecondary: Hex("#e1e2e7"), // Background
	TextOnAccent:    Hex("#e1e2e7"), // Background
	TextDisabled:    Hex("#9699a3"), // Muted

	Border:    Hex("#c9cad0"), // Border
	FocusRing: Hex("#2e7de9"), // Blue

	Error:   Hex("#f52a65"), // Red
	Warning: Hex("#8c6c3e"), // Yellow
	Success: Hex("#587539"), // Green
	Info:    Hex("#007197"), // Cyan

	TextOnError:   Hex("#e1e2e7"), // Background
	TextOnWarning: Hex("#e1e2e7"), // Background
	TextOnSuccess: Hex("#e1e2e7"), // Background
	TextOnInfo:    Hex("#e1e2e7"), // Background

	ActiveCursor:  Hex("#007197"),                 // Accent (Cyan)
	Selection:     Hex("#007197").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#e1e2e7"),                 // TextOnAccent (Background)

	ScrollbarTrack: Hex("#d5d6db"), // Surface
	ScrollbarThumb: Hex("#848cb5"), // Comment

	Overlay: Hex("#3760bf").WithAlpha(0.5), // Foreground with transparency

	Placeholder: Hex("#9699a3"), // Muted
	Cursor:      Hex("#3760bf"), // Foreground

	Link: Hex("#007197"), // Cyan
}

// catppuccinLatteThemeData - Catppuccin Latte light variant
// https://catppuccin.com/
var catppuccinLatteThemeData = ThemeData{
	Name:    ThemeNameCatppuccinLatte,
	IsLight: true,

	Primary:   Hex("#8839ef"), // Mauve
	Secondary: Hex("#ea76cb"), // Pink
	Accent:    Hex("#179299"), // Teal

	Background:   Hex("#eff1f5"), // Base
	Surface:      Hex("#e6e9ef"), // Surface0
	SurfaceHover: Hex("#dce0e8"), // Surface1
	Surface2:     Hex("#ccd0da"), // Surface2
	Surface3:     Hex("#bcc0cc"), // Overlay0

	Text:            Hex("#4c4f69"), // Text
	TextMuted:       Hex("#7c7f93"), // Overlay2
	TextOnPrimary:   Hex("#eff1f5"), // Base
	TextOnSecondary: Hex("#eff1f5"), // Base
	TextOnAccent:    Hex("#eff1f5"), // Base
	TextDisabled:    Hex("#9ca0b0"), // Overlay1

	Border:    Hex("#dce0e8"), // Surface1
	FocusRing: Hex("#8839ef"), // Mauve

	Error:   Hex("#d20f39"), // Red
	Warning: Hex("#fe640b"), // Peach
	Success: Hex("#40a02b"), // Green
	Info:    Hex("#1e66f5"), // Blue

	TextOnError:   Hex("#eff1f5"), // Base
	TextOnWarning: Hex("#eff1f5"), // Base
	TextOnSuccess: Hex("#eff1f5"), // Base
	TextOnInfo:    Hex("#eff1f5"), // Base

	ActiveCursor:  Hex("#179299"),                 // Accent (Teal)
	Selection:     Hex("#179299").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#eff1f5"),                 // TextOnAccent (Base)

	ScrollbarTrack: Hex("#e6e9ef"), // Surface0
	ScrollbarThumb: Hex("#9ca0b0"), // Overlay1

	Overlay: Hex("#4c4f69").WithAlpha(0.5), // Text with transparency

	Placeholder: Hex("#9ca0b0"), // Overlay1
	Cursor:      Hex("#4c4f69"), // Text

	Link: Hex("#1e66f5"), // Blue
}

// gruvboxLightThemeData - Gruvbox Light variant
// https://github.com/morhetz/gruvbox
var gruvboxLightThemeData = ThemeData{
	Name:    ThemeNameGruvboxLight,
	IsLight: true,

	Primary:   Hex("#d79921"), // Yellow
	Secondary: Hex("#b16286"), // Purple
	Accent:    Hex("#689d6a"), // Aqua

	Background:   Hex("#fbf1c7"), // bg0
	Surface:      Hex("#ebdbb2"), // bg1
	SurfaceHover: Hex("#d5c4a1"), // bg2
	Surface2:     Hex("#bdae93"), // bg3
	Surface3:     Hex("#a89984"), // bg4

	Text:            Hex("#3c3836"), // fg1
	TextMuted:       Hex("#7c6f64"), // gray
	TextOnPrimary:   Hex("#fbf1c7"), // bg0
	TextOnSecondary: Hex("#fbf1c7"), // bg0
	TextOnAccent:    Hex("#fbf1c7"), // bg0
	TextDisabled:    Hex("#928374"), // gray

	Border:    Hex("#d5c4a1"), // bg2
	FocusRing: Hex("#d79921"), // yellow

	Error:   Hex("#cc241d"), // red
	Warning: Hex("#d65d0e"), // orange
	Success: Hex("#98971a"), // green
	Info:    Hex("#458588"), // blue

	TextOnError:   Hex("#fbf1c7"), // bg0
	TextOnWarning: Hex("#fbf1c7"), // bg0
	TextOnSuccess: Hex("#fbf1c7"), // bg0
	TextOnInfo:    Hex("#fbf1c7"), // bg0

	ActiveCursor:  Hex("#689d6a"),                 // Accent (Aqua)
	Selection:     Hex("#689d6a").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#fbf1c7"),                 // TextOnAccent (bg0)

	ScrollbarTrack: Hex("#ebdbb2"), // bg1
	ScrollbarThumb: Hex("#928374"), // gray

	Overlay: Hex("#3c3836").WithAlpha(0.5), // fg1 with transparency

	Placeholder: Hex("#928374"), // gray
	Cursor:      Hex("#3c3836"), // fg1

	Link: Hex("#458588"), // blue
}

// nordLightThemeData - Nord Light variant (Snow Storm palette)
// https://www.nordtheme.com/
var nordLightThemeData = ThemeData{
	Name:    ThemeNameNordLight,
	IsLight: true,

	Primary:   Hex("#5e81ac"), // Nord10 - frost
	Secondary: Hex("#81a1c1"), // Nord9 - frost
	Accent:    Hex("#88c0d0"), // Nord8 - frost

	Background:   Hex("#eceff4"), // Nord6 - snow storm
	Surface:      Hex("#e5e9f0"), // Nord5 - snow storm
	SurfaceHover: Hex("#d8dee9"), // Nord4 - snow storm
	Surface2:     Hex("#c9d1dc"), // Darker snow storm
	Surface3:     Hex("#bac3cf"), // Even darker

	Text:            Hex("#2e3440"), // Nord0 - polar night
	TextMuted:       Hex("#4c566a"), // Nord3 - polar night
	TextOnPrimary:   Hex("#eceff4"), // Nord6
	TextOnSecondary: Hex("#2e3440"), // Nord0
	TextOnAccent:    Hex("#2e3440"), // Nord0
	TextDisabled:    Hex("#7b88a1"), // Muted

	Border:    Hex("#d8dee9"), // Nord4
	FocusRing: Hex("#5e81ac"), // Nord10

	Error:   Hex("#bf616a"), // Nord11 - aurora red
	Warning: Hex("#d08770"), // Nord12 - aurora orange
	Success: Hex("#a3be8c"), // Nord14 - aurora green
	Info:    Hex("#5e81ac"), // Nord10 - frost

	TextOnError:   Hex("#eceff4"), // Nord6
	TextOnWarning: Hex("#2e3440"), // Nord0
	TextOnSuccess: Hex("#2e3440"), // Nord0
	TextOnInfo:    Hex("#eceff4"), // Nord6

	ActiveCursor:  Hex("#88c0d0"),                 // Accent (Nord8)
	Selection:     Hex("#88c0d0").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#2e3440"),                 // TextOnAccent (Nord0)

	ScrollbarTrack: Hex("#e5e9f0"), // Nord5
	ScrollbarThumb: Hex("#7b88a1"), // Muted

	Overlay: Hex("#2e3440").WithAlpha(0.5), // Nord0 with transparency

	Placeholder: Hex("#7b88a1"), // Muted
	Cursor:      Hex("#2e3440"), // Nord0

	Link: Hex("#5e81ac"), // Nord10
}

// solarizedLightThemeData - Solarized Light variant
// https://ethanschoonover.com/solarized/
var solarizedLightThemeData = ThemeData{
	Name:    ThemeNameSolarizedLight,
	IsLight: true,

	Primary:   Hex("#268bd2"), // Blue
	Secondary: Hex("#6c71c4"), // Violet
	Accent:    Hex("#2aa198"), // Cyan

	Background:   Hex("#fdf6e3"), // base3
	Surface:      Hex("#eee8d5"), // base2
	SurfaceHover: Hex("#93a1a1"), // base1
	Surface2:     Hex("#839496"), // base0
	Surface3:     Hex("#657b83"), // base00

	Text:            Hex("#657b83"), // base00
	TextMuted:       Hex("#93a1a1"), // base1
	TextOnPrimary:   Hex("#fdf6e3"), // base3
	TextOnSecondary: Hex("#fdf6e3"), // base3
	TextOnAccent:    Hex("#fdf6e3"), // base3
	TextDisabled:    Hex("#93a1a1"), // base1

	Border:    Hex("#eee8d5"), // base2
	FocusRing: Hex("#268bd2"), // blue

	Error:   Hex("#dc322f"), // red
	Warning: Hex("#b58900"), // yellow
	Success: Hex("#859900"), // green
	Info:    Hex("#2aa198"), // cyan

	TextOnError:   Hex("#fdf6e3"), // base3
	TextOnWarning: Hex("#fdf6e3"), // base3
	TextOnSuccess: Hex("#fdf6e3"), // base3
	TextOnInfo:    Hex("#fdf6e3"), // base3

	ActiveCursor:  Hex("#2aa198"),                 // Accent (Cyan)
	Selection:     Hex("#2aa198").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#fdf6e3"),                 // TextOnAccent (base3)

	ScrollbarTrack: Hex("#eee8d5"), // base2
	ScrollbarThumb: Hex("#93a1a1"), // base1

	Overlay: Hex("#002b36").WithAlpha(0.5), // base03 with transparency

	Placeholder: Hex("#93a1a1"), // base1
	Cursor:      Hex("#657b83"), // base00

	Link: Hex("#268bd2"), // blue
}

// kanagawaLotusThemeData - Kanagawa Lotus light variant
// https://github.com/rebelot/kanagawa.nvim
var kanagawaLotusThemeData = ThemeData{
	Name:    ThemeNameKanagawaLotus,
	IsLight: true,

	Primary:   Hex("#4d699b"), // lotusBlue
	Secondary: Hex("#624c83"), // lotusViolet
	Accent:    Hex("#597b75"), // lotusAqua

	Background:   Hex("#f2ecbc"), // lotusWhite0
	Surface:      Hex("#e7dba0"), // lotusWhite1
	SurfaceHover: Hex("#d5cea3"), // lotusWhite2
	Surface2:     Hex("#c9c3a0"), // lotusWhite3
	Surface3:     Hex("#bdb89d"), // lotusWhite4

	Text:            Hex("#545464"), // lotusInk1
	TextMuted:       Hex("#8a8980"), // lotusFuji
	TextOnPrimary:   Hex("#f2ecbc"), // lotusWhite0
	TextOnSecondary: Hex("#f2ecbc"), // lotusWhite0
	TextOnAccent:    Hex("#f2ecbc"), // lotusWhite0
	TextDisabled:    Hex("#a09f95"), // Muted

	Border:    Hex("#d5cea3"), // lotusWhite2
	FocusRing: Hex("#4d699b"), // lotusBlue

	Error:   Hex("#c84053"), // lotusRed
	Warning: Hex("#77713f"), // lotusYellow
	Success: Hex("#6f894e"), // lotusGreen
	Info:    Hex("#4d699b"), // lotusBlue

	TextOnError:   Hex("#f2ecbc"), // lotusWhite0
	TextOnWarning: Hex("#f2ecbc"), // lotusWhite0
	TextOnSuccess: Hex("#f2ecbc"), // lotusWhite0
	TextOnInfo:    Hex("#f2ecbc"), // lotusWhite0

	ActiveCursor:  Hex("#597b75"),                 // Accent (lotusAqua)
	Selection:     Hex("#597b75").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#f2ecbc"),                 // TextOnAccent (lotusWhite0)

	ScrollbarTrack: Hex("#e7dba0"), // lotusWhite1
	ScrollbarThumb: Hex("#a09f95"), // Muted

	Overlay: Hex("#545464").WithAlpha(0.5), // lotusInk1 with transparency

	Placeholder: Hex("#a09f95"), // Muted
	Cursor:      Hex("#545464"), // lotusInk1

	Link: Hex("#4d699b"), // lotusBlue
}

// monokaiLightThemeData - Monokai Light variant
var monokaiLightThemeData = ThemeData{
	Name:    ThemeNameMonokaiLight,
	IsLight: true,

	Primary:   Hex("#7a8c21"), // Green (darkened)
	Secondary: Hex("#8c6bc8"), // Purple
	Accent:    Hex("#0f9fbf"), // Cyan (darkened)

	Background:   Hex("#fafafa"), // Light background
	Surface:      Hex("#f0f0f0"), // Surface
	SurfaceHover: Hex("#e5e5e5"), // Surface hover
	Surface2:     Hex("#dadada"), // Darker
	Surface3:     Hex("#cfcfcf"), // Even darker

	Text:            Hex("#272822"), // Dark text (original bg)
	TextMuted:       Hex("#75715e"), // Comment
	TextOnPrimary:   Hex("#fafafa"), // Light
	TextOnSecondary: Hex("#fafafa"), // Light
	TextOnAccent:    Hex("#fafafa"), // Light
	TextDisabled:    Hex("#a59f85"), // Muted comment

	Border:    Hex("#dadada"), // Border
	FocusRing: Hex("#7a8c21"), // Green

	Error:   Hex("#f92672"), // Pink/Red
	Warning: Hex("#fd971f"), // Orange
	Success: Hex("#7a8c21"), // Green
	Info:    Hex("#0f9fbf"), // Cyan

	TextOnError:   Hex("#fafafa"), // Light
	TextOnWarning: Hex("#272822"), // Dark
	TextOnSuccess: Hex("#fafafa"), // Light
	TextOnInfo:    Hex("#fafafa"), // Light

	ActiveCursor:  Hex("#0f9fbf"),                 // Accent (Cyan)
	Selection:     Hex("#0f9fbf").WithAlpha(0.12), // Accent with alpha for multi-select
	SelectionText: Hex("#fafafa"),                 // TextOnAccent (Light)

	ScrollbarTrack: Hex("#e5e5e5"), // Surface hover
	ScrollbarThumb: Hex("#a59f85"), // Muted comment

	Overlay: Hex("#272822").WithAlpha(0.5), // Dark with transparency

	Placeholder: Hex("#a59f85"), // Muted comment
	Cursor:      Hex("#272822"), // Dark text

	Link: Hex("#0f9fbf"), // Cyan
}

// themeRegistry holds all registered themes
var themeRegistry = map[string]ThemeData{
	// Dark themes
	ThemeNameRosePine:   rosePineThemeData,
	ThemeNameDracula:    draculaThemeData,
	ThemeNameTokyoNight: tokyoNightThemeData,
	ThemeNameCatppuccin: catppuccinThemeData,
	ThemeNameGruvbox:    gruvboxThemeData,
	ThemeNameNord:       nordThemeData,
	ThemeNameSolarized:  solarizedThemeData,
	ThemeNameKanagawa:   kanagawaThemeData,
	ThemeNameMonokai:    monokaiThemeData,
	// Light themes
	ThemeNameRosePineDawn:    rosePineDawnThemeData,
	ThemeNameDraculaLight:    draculaLightThemeData,
	ThemeNameTokyoNightDay:   tokyoNightDayThemeData,
	ThemeNameCatppuccinLatte: catppuccinLatteThemeData,
	ThemeNameGruvboxLight:    gruvboxLightThemeData,
	ThemeNameNordLight:       nordLightThemeData,
	ThemeNameSolarizedLight:  solarizedLightThemeData,
	ThemeNameKanagawaLotus:   kanagawaLotusThemeData,
	ThemeNameMonokaiLight:    monokaiLightThemeData,
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
	computeLabelColors(&data)
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

// ThemeNames returns a slice of all registered theme names in alphabetical order.
func ThemeNames() []string {
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LightThemeNames returns a slice of all registered light theme names in alphabetical order.
func LightThemeNames() []string {
	names := make([]string, 0)
	for name, data := range themeRegistry {
		if data.IsLight {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// DarkThemeNames returns a slice of all registered dark theme names in alphabetical order.
func DarkThemeNames() []string {
	names := make([]string, 0)
	for name, data := range themeRegistry {
		if !data.IsLight {
			names = append(names, name)
		}
	}
	sort.Strings(names)
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

// ============================================================================
// Theme Inheritance API
// ============================================================================

// ThemeOption is a functional option for modifying theme data.
type ThemeOption func(*ThemeData)

// ExtendTheme creates a new theme based on an existing one with modifications.
// Returns zero ThemeData if base theme not found.
func ExtendTheme(baseName string, opts ...ThemeOption) ThemeData {
	base, ok := GetTheme(baseName)
	if !ok {
		Log("ExtendTheme: base theme not found: %s", baseName)
		return ThemeData{}
	}

	// Apply all options to the copy
	for _, opt := range opts {
		opt(&base)
	}

	// Recompute derived colors
	computeLabelColors(&base)

	return base
}

// ExtendAndRegisterTheme extends a theme and registers it in one call.
// Returns false if base theme not found.
func ExtendAndRegisterTheme(newName, baseName string, opts ...ThemeOption) bool {
	extended := ExtendTheme(baseName, opts...)
	if extended.Name == "" && extended.Primary == (Color{}) {
		return false
	}
	RegisterTheme(newName, extended)
	return true
}

// Core branding options

// WithPrimary sets the Primary color.
func WithPrimary(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Primary = c
	}
}

// WithSecondary sets the Secondary color.
func WithSecondary(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Secondary = c
	}
}

// WithAccent sets the Accent color.
func WithAccent(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Accent = c
	}
}

// Surface options

// WithBackground sets the Background color.
func WithBackground(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Background = c
	}
}

// WithSurface sets the Surface color.
func WithSurface(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Surface = c
	}
}

// WithSurfaceHover sets the SurfaceHover color.
func WithSurfaceHover(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.SurfaceHover = c
	}
}

// WithSurface2 sets the Surface2 color.
func WithSurface2(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Surface2 = c
	}
}

// WithSurface3 sets the Surface3 color.
func WithSurface3(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Surface3 = c
	}
}

// Text options

// WithText sets the Text color.
func WithText(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Text = c
	}
}

// WithTextMuted sets the TextMuted color.
func WithTextMuted(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextMuted = c
	}
}

// WithTextOnPrimary sets the TextOnPrimary color.
func WithTextOnPrimary(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnPrimary = c
	}
}

// WithTextOnSecondary sets the TextOnSecondary color.
func WithTextOnSecondary(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnSecondary = c
	}
}

// WithTextOnAccent sets the TextOnAccent color.
func WithTextOnAccent(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnAccent = c
	}
}

// WithTextDisabled sets the TextDisabled color.
func WithTextDisabled(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextDisabled = c
	}
}

// Border options

// WithBorder sets the Border color.
func WithBorder(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Border = c
	}
}

// WithFocusRing sets the FocusRing color.
func WithFocusRing(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.FocusRing = c
	}
}

// Feedback options

// WithError sets the Error color.
func WithError(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Error = c
	}
}

// WithWarning sets the Warning color.
func WithWarning(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Warning = c
	}
}

// WithSuccess sets the Success color.
func WithSuccess(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Success = c
	}
}

// WithInfo sets the Info color.
func WithInfo(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Info = c
	}
}

// WithTextOnError sets the TextOnError color.
func WithTextOnError(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnError = c
	}
}

// WithTextOnWarning sets the TextOnWarning color.
func WithTextOnWarning(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnWarning = c
	}
}

// WithTextOnSuccess sets the TextOnSuccess color.
func WithTextOnSuccess(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnSuccess = c
	}
}

// WithTextOnInfo sets the TextOnInfo color.
func WithTextOnInfo(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.TextOnInfo = c
	}
}

// Selection options

// WithActiveCursor sets the ActiveCursor color.
func WithActiveCursor(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.ActiveCursor = c
	}
}

// WithSelection sets the Selection color.
func WithSelection(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Selection = c
	}
}

// WithSelectionText sets the SelectionText color.
func WithSelectionText(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.SelectionText = c
	}
}

// Scrollbar options

// WithScrollbarTrack sets the ScrollbarTrack color.
func WithScrollbarTrack(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.ScrollbarTrack = c
	}
}

// WithScrollbarThumb sets the ScrollbarThumb color.
func WithScrollbarThumb(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.ScrollbarThumb = c
	}
}

// Other options

// WithOverlay sets the Overlay color.
func WithOverlay(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Overlay = c
	}
}

// WithPlaceholder sets the Placeholder color.
func WithPlaceholder(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Placeholder = c
	}
}

// WithCursor sets the Cursor color.
func WithCursor(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Cursor = c
	}
}

// WithLink sets the Link color.
func WithLink(c Color) ThemeOption {
	return func(t *ThemeData) {
		t.Link = c
	}
}

// Metadata options

// WithIsLight sets the IsLight flag.
func WithIsLight(isLight bool) ThemeOption {
	return func(t *ThemeData) {
		t.IsLight = isLight
	}
}
