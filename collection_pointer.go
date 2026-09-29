package terma

import uv "github.com/charmbracelet/ultraviolet"

// clickActivates reports whether a press on an item of a list, table or tree
// activates it (calls OnSelect): a double-click, or with activateOnClick a
// plain left click. A shift+click extends the selection instead, and the
// second click of a double-click doesn't activate again.
func clickActivates(event MouseEvent, activateOnClick bool) bool {
	if !activateOnClick {
		return event.ClickCount == 2
	}
	return event.ClickCount == 1 && event.Button == uv.MouseLeft && !event.Mod.Contains(uv.ModShift)
}

// spanAt returns the index of the span containing pos, where span(i) gives
// the start and size of span i of n, in increasing order. It is how
// collections find the item, row or column under the pointer. Empty spans are
// skipped. With clamp, a pos before every span gives the first span, and one
// after or between spans gives the span before it, so a drag past either end
// of a collection lands on its first or last item.
func spanAt(n int, span func(i int) (start, size int), pos int, clamp bool) (int, bool) {
	first, before := -1, -1
	for i := 0; i < n; i++ {
		start, size := span(i)
		if size <= 0 {
			continue
		}
		if pos >= start && pos < start+size {
			return i, true
		}
		if first < 0 {
			first = i
		}
		if start <= pos {
			before = i
		}
	}
	if !clamp || first < 0 {
		return 0, false
	}
	if before < 0 {
		return first, true
	}
	return before, true
}
