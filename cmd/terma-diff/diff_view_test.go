package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineNumberRolesForLine(tt *testing.T) {
	oldRole, newRole := lineNumberRolesForLine(RenderedLineContext)
	require.Equal(tt, TokenRoleOldLineNumber, oldRole)
	require.Equal(tt, TokenRoleNewLineNumber, newRole)

	oldRole, newRole = lineNumberRolesForLine(RenderedLineAdd)
	require.Equal(tt, TokenRoleLineNumberAdd, oldRole)
	require.Equal(tt, TokenRoleLineNumberAdd, newRole)

	oldRole, newRole = lineNumberRolesForLine(RenderedLineRemove)
	require.Equal(tt, TokenRoleLineNumberRemove, oldRole)
	require.Equal(tt, TokenRoleLineNumberRemove, newRole)
}
