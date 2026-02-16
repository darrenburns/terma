package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	t "terma"
)

func TestThemePalette_UsesTextVariantsForSemanticRoles(tt *testing.T) {
	theme, ok := t.GetTheme(t.CurrentThemeName())
	if !ok {
		tt.Fatalf("theme %q not found", t.CurrentThemeName())
	}

	palette := NewThemePalette(theme)

	assertRoleColor(tt, palette, TokenRoleLineNumberAdd, theme.SuccessText)
	assertRoleColor(tt, palette, TokenRoleLineNumberRemove, theme.ErrorText)
	assertRoleColor(tt, palette, TokenRoleDiffPrefixAdd, theme.SuccessText)
	assertRoleColor(tt, palette, TokenRoleDiffPrefixRemove, theme.ErrorText)
	assertRoleColor(tt, palette, TokenRoleDiffFileHeader, theme.PrimaryText)
	assertRoleColor(tt, palette, TokenRoleDiffHunkHeader, theme.TextMuted.Blend(theme.InfoText, 0.35))
	assertRoleColor(tt, palette, TokenRoleDiffMeta, theme.WarningText)
	assertRoleColor(tt, palette, TokenRoleOldLineNumber, theme.TextMuted.Blend(theme.TextDisabled, 0.35))
	assertRoleColor(tt, palette, TokenRoleNewLineNumber, theme.TextMuted.Blend(theme.TextDisabled, 0.35))
	assertRoleColor(tt, palette, TokenRoleSyntaxKeyword, theme.AccentText)
	assertRoleColor(tt, palette, TokenRoleSyntaxType, theme.PrimaryText)
	assertRoleColor(tt, palette, TokenRoleSyntaxFunction, theme.SecondaryText)
	assertRoleColor(tt, palette, TokenRoleSyntaxString, theme.SuccessText)
	assertRoleColor(tt, palette, TokenRoleSyntaxNumber, theme.WarningText)

	hunkStyle, ok := palette.StyleForRole(TokenRoleDiffHunkHeader)
	if !ok {
		tt.Fatalf("missing style for role %v", TokenRoleDiffHunkHeader)
	}
	if hunkStyle.Bold {
		tt.Fatalf("hunk header should not be bold")
	}
}

func assertRoleColor(tt *testing.T, palette ThemePalette, role TokenRole, expected t.Color) {
	tt.Helper()
	style, ok := palette.StyleForRole(role)
	if !ok {
		tt.Fatalf("missing style for role %v", role)
	}
	if style.Foreground != expected {
		tt.Fatalf("role %v foreground: got %v, want %v", role, style.Foreground, expected)
	}
}

func TestThemePalette_GutterTintIsDarkerForAddAndRemove(tt *testing.T) {
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	palette := NewThemePalette(theme)

	addLineStyle, ok := palette.LineStyleForKind(RenderedLineAdd)
	require.True(tt, ok)
	addGutterStyle, ok := palette.GutterStyleForKind(RenderedLineAdd)
	require.True(tt, ok)
	require.NotNil(tt, addLineStyle.BackgroundColor)
	require.NotNil(tt, addGutterStyle.BackgroundColor)
	addLineBg := addLineStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
	addGutterBg := addGutterStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
	require.Less(tt, addGutterBg.Luminance(), addLineBg.Luminance())

	removeLineStyle, ok := palette.LineStyleForKind(RenderedLineRemove)
	require.True(tt, ok)
	removeGutterStyle, ok := palette.GutterStyleForKind(RenderedLineRemove)
	require.True(tt, ok)
	require.NotNil(tt, removeLineStyle.BackgroundColor)
	require.NotNil(tt, removeGutterStyle.BackgroundColor)
	removeLineBg := removeLineStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
	removeGutterBg := removeGutterStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
	require.Less(tt, removeGutterBg.Luminance(), removeLineBg.Luminance())

	contextGutterStyle, ok := palette.GutterStyleForKind(RenderedLineContext)
	require.True(tt, ok)
	require.NotNil(tt, contextGutterStyle.BackgroundColor)
	contextGutterBg := contextGutterStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
	require.Less(tt, contextGutterBg.Luminance(), theme.Background.Luminance())
}
