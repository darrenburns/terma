package terma

import "testing"

func TestCommandPaletteSetItemsWhileClosedRestsOnCurrentItem(t *testing.T) {
	state := NewCommandPaletteState("Demos", []CommandPaletteItem{
		{Label: "Home"}, {Label: "List"}, {Label: "Table"},
	})

	state.SetItems([]CommandPaletteItem{
		{Label: "Home"}, {Label: "List"}, {Label: "Table", Current: true},
	})

	if item, _ := state.CurrentItem(); item.Label != "Table" {
		t.Fatalf("cursor on %q, want the Current item %q", item.Label, "Table")
	}
}

func TestCommandPaletteSetItemsWhileOpenKeepsCursor(t *testing.T) {
	state := NewCommandPaletteState("Demos", []CommandPaletteItem{
		{Label: "Home"}, {Label: "List"}, {Label: "Table"},
	})
	state.Open()
	state.CurrentLevel().ListState.SelectIndex(1)

	state.SetItems([]CommandPaletteItem{
		{Label: "Home"}, {Label: "List"}, {Label: "Table", Current: true},
	})

	if item, _ := state.CurrentItem(); item.Label != "List" {
		t.Fatalf("cursor on %q, want it left on %q", item.Label, "List")
	}
}
