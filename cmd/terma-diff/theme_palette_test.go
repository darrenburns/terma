package main

import (
	"testing"

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
