package terma

type cursorTerminal interface {
	MoveTo(x, y int)
	ShowCursor()
	HideCursor()
}

func positionCursor(terminal cursorTerminal, entry *WidgetEntry) {
	terminal.HideCursor()
	if entry == nil {
		return
	}
	provider, ok := entry.Widget.(CursorProvider)
	if !ok {
		return
	}
	x, y, visible := provider.CursorPosition()
	x += entry.Bounds.X
	y += entry.Bounds.Y
	if !entry.Visible.Contains(x, y) {
		return
	}
	terminal.MoveTo(x, y)
	if visible {
		terminal.ShowCursor()
	}
}
