package main

import t "terma"

// ThemePalette stores all styles needed by the diff renderer for one theme.
type ThemePalette struct {
	roleStyles map[TokenRole]t.SpanStyle
	lineStyles map[RenderedLineKind]t.Style
}

func NewThemePalette(theme t.ThemeData) ThemePalette {
	addBg := theme.Background.Blend(theme.Success, 0.14)
	removeBg := theme.Background.Blend(theme.Error, 0.14)
	hunkBg := theme.Background.Blend(theme.Info, 0.13)
	headerBg := theme.Background.Blend(theme.Primary, 0.11)

	return ThemePalette{
		roleStyles: map[TokenRole]t.SpanStyle{
			TokenRoleOldLineNumber:     {Foreground: theme.TextMuted},
			TokenRoleNewLineNumber:     {Foreground: theme.TextMuted},
			TokenRoleLineNumberAdd:     {Foreground: theme.Success},
			TokenRoleLineNumberRemove:  {Foreground: theme.Error},
			TokenRoleDiffPrefixAdd:     {Foreground: theme.Success, Bold: true},
			TokenRoleDiffPrefixRemove:  {Foreground: theme.Error, Bold: true},
			TokenRoleDiffPrefixContext: {Foreground: theme.TextMuted},
			TokenRoleDiffFileHeader:    {Foreground: theme.Primary, Bold: true},
			TokenRoleDiffHunkHeader:    {Foreground: theme.Info, Bold: true},
			TokenRoleDiffMeta:          {Foreground: theme.Warning, Italic: true},
			TokenRoleSyntaxPlain:       {Foreground: theme.Text},
			TokenRoleSyntaxKeyword:     {Foreground: theme.Accent, Bold: true},
			TokenRoleSyntaxType:        {Foreground: theme.Primary},
			TokenRoleSyntaxFunction:    {Foreground: theme.Secondary},
			TokenRoleSyntaxString:      {Foreground: theme.Success},
			TokenRoleSyntaxNumber:      {Foreground: theme.Warning},
			TokenRoleSyntaxComment:     {Foreground: theme.TextMuted, Italic: true},
			TokenRoleSyntaxPunctuation: {Foreground: theme.Text},
		},
		lineStyles: map[RenderedLineKind]t.Style{
			RenderedLineFileHeader: {BackgroundColor: headerBg},
			RenderedLineHunkHeader: {BackgroundColor: hunkBg},
			RenderedLineAdd:        {BackgroundColor: addBg},
			RenderedLineRemove:     {BackgroundColor: removeBg},
		},
	}
}

func (p ThemePalette) StyleForRole(role TokenRole) (t.SpanStyle, bool) {
	style, ok := p.roleStyles[role]
	return style, ok
}

func (p ThemePalette) LineStyleForKind(kind RenderedLineKind) (t.Style, bool) {
	style, ok := p.lineStyles[kind]
	return style, ok
}
