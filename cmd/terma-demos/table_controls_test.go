package main

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/stretchr/testify/require"
)

func invokeBinding(test *testing.T, widget t.KeybindProvider, key string) {
	test.Helper()
	for _, bind := range widget.Keybinds() {
		if bind.Key == key {
			bind.Action()
			return
		}
	}
	test.Fatalf("missing key binding %s", key)
}

func tableNameCell(test *testing.T, buffer *uv.Buffer, name string) *uv.Cell {
	test.Helper()
	for y := 0; y < 30; y++ {
		var line strings.Builder
		for x := 0; x < 100; x++ {
			cell := buffer.CellAt(x, y)
			if cell == nil || cell.Content == "" {
				line.WriteByte(' ')
			} else {
				line.WriteString(cell.Content)
			}
		}
		text := line.String()
		if strings.HasPrefix(strings.TrimLeft(text, " "), name+" ") {
			return buffer.CellAt(strings.Index(text, name), y)
		}
	}
	test.Fatalf("table row %q not visible", name)
	return nil
}

func TestGalleryTableMultiselectColours(test *testing.T) {
	galleryTheme(test)
	g := NewGallery(demos)
	test.Cleanup(func() { require.NoError(test, g.Close()) })
	openDemo(test, g, "Table controls")
	table := renderGallery(g).WidgetByID("people").EventWidget.(t.KeybindProvider)
	invokeBinding(test, table, "shift+down")

	for _, name := range demokit.Themes {
		test.Run(name, func(test *testing.T) {
			t.SetTheme(name)
			theme, ok := t.GetTheme(name)
			require.True(test, ok)
			for _, focus := range []string{"people", "filter"} {
				test.Run(focus, func(test *testing.T) {
					t.RequestFocus(focus)
					buffer := t.RenderToBuffer(g, 100, 30)
					cursor, selected := tableNameCell(test, buffer, "Ada"), tableNameCell(test, buffer, "Zoe")
					cursorBG, cursorFG := theme.Selection.BlendOver(theme.Background), theme.Text
					if focus == "people" {
						cursorBG, cursorFG = theme.ActiveCursor, theme.SelectionText
					}
					require.Equal(test, cursorBG.Hex(), t.FromANSI(cursor.Style.Bg).Hex(), "selected cursor must retain the normal Table focus colour")
					require.Equal(test, cursorFG.Hex(), t.FromANSI(cursor.Style.Fg).Hex(), "cursor text must contrast with its background")
					require.Equal(test, theme.Selection.BlendOver(theme.Background).Hex(), t.FromANSI(selected.Style.Bg).Hex(), "other selections must remain visibly distinct")
					require.Equal(test, theme.Text.Hex(), t.FromANSI(selected.Style.Fg).Hex())
					if name == demokit.Themes[0] {
						t.RequestFocus(focus)
						t.AssertSnapshot(test, g, 100, 30)
					}
				})
			}
		})
	}
}
