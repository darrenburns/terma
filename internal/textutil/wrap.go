// Package textutil contains text operations shared by layout and painting.
package textutil

import "github.com/charmbracelet/x/ansi"

// HardWrap splits a single line at grapheme boundaries. ANSI sequences and
// whitespace are preserved. A grapheme wider than maxWidth occupies a line by
// itself; consuming it guarantees progress even in a one-cell viewport.
func HardWrap(line string, maxWidth int) []string {
	if maxWidth <= 0 || line == "" {
		return []string{line}
	}

	var lines []string
	var state byte
	start, width := 0, 0
	for offset := 0; offset < len(line); {
		_, clusterWidth, size, nextState := ansi.DecodeSequence(line[offset:], state, nil)
		// DecodeSequence fast-paths ASCII bytes individually. An ASCII base
		// followed by Unicode may instead begin a wider grapheme (e.g. 1️⃣).
		if state == ansi.NormalState && line[offset] >= ' ' && line[offset] <= '~' && offset+1 < len(line) && line[offset+1] >= 0x80 {
			cluster, cellWidth := ansi.FirstGraphemeCluster(line[offset:], ansi.GraphemeWidth)
			clusterWidth, size = cellWidth, len(cluster)
		}
		if size == 0 {
			// Malformed escape sequences must not stall wrapping. Keep the byte
			// intact and restart the decoder at the next byte.
			size, nextState = 1, ansi.NormalState
		}
		if clusterWidth > 0 && width > 0 && width+clusterWidth > maxWidth {
			lines = append(lines, line[start:offset])
			start, width = offset, 0
		}
		width += clusterWidth
		offset += size
		state = nextState
	}
	return append(lines, line[start:])
}
