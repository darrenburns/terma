package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func cursorPrefixSlotWidth(cursorPrefix, selectedPrefix string) int {
	return max(ansi.StringWidth(cursorPrefix), ansi.StringWidth(selectedPrefix))
}

func padCursorPrefix(prefix string, slotWidth int) string {
	width := ansi.StringWidth(prefix)
	if width >= slotWidth {
		return prefix
	}
	return prefix + strings.Repeat(" ", slotWidth-width)
}
