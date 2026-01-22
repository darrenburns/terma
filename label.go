package terma

// LabelVariant represents the semantic color variant for a Label.
type LabelVariant int

const (
	LabelDefault LabelVariant = iota
	LabelPrimary
	LabelAccent
	LabelSuccess
	LabelError
	LabelWarning
	LabelInfo
)

// Label returns a styled Text widget with the given variant colors.
// Applies padding x=1, y=0 and variant-appropriate foreground/background.
func Label(content string, variant LabelVariant, theme ThemeData) Text {
	fg, bg := labelVariantColors(variant, theme)
	return Text{
		Content: content,
		Style: Style{
			Padding:         EdgeInsetsXY(1, 0),
			ForegroundColor: fg,
			BackgroundColor: bg,
		},
	}
}

func labelVariantColors(variant LabelVariant, theme ThemeData) (fg, bg Color) {
	switch variant {
	case LabelPrimary:
		return theme.PrimaryText, theme.PrimaryBg
	case LabelAccent:
		return theme.AccentText, theme.AccentBg
	case LabelSuccess:
		return theme.SuccessText, theme.SuccessBg
	case LabelError:
		return theme.ErrorText, theme.ErrorBg
	case LabelWarning:
		return theme.WarningText, theme.WarningBg
	case LabelInfo:
		return theme.InfoText, theme.InfoBg
	default:
		return theme.TextMuted, theme.Surface
	}
}
