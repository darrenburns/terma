package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"
)

var themeNames = t.DarkThemeNames()

// SimpleTreeDemo demonstrates the most basic usage of Tree[string]
// with filtering via TextInput and multi-select enabled.
type SimpleTreeDemo struct {
	treeState   *t.TreeState[string]
	filterState *t.FilterState
	filterInput *t.TextInputState
	selectedMsg t.Signal[string]
	themeIndex  t.Signal[int]
}

func NewSimpleTreeDemo() *SimpleTreeDemo {
	return &SimpleTreeDemo{
		treeState: t.NewTreeState([]t.TreeNode[string]{
			{
				Data: "Fruits",
				Children: []t.TreeNode[string]{
					{Data: "Apple", Children: []t.TreeNode[string]{}},
					{Data: "Banana", Children: []t.TreeNode[string]{}},
					{Data: "Cherry", Children: []t.TreeNode[string]{}},
				},
			},
			{
				Data: "Vegetables",
				Children: []t.TreeNode[string]{
					{Data: "Carrot", Children: []t.TreeNode[string]{}},
					{Data: "Broccoli", Children: []t.TreeNode[string]{}},
					{Data: "Spinach", Children: []t.TreeNode[string]{}},
				},
			},
			{
				Data: "Grains",
				Children: []t.TreeNode[string]{
					{Data: "Rice", Children: []t.TreeNode[string]{}},
					{Data: "Wheat", Children: []t.TreeNode[string]{}},
					{Data: "Oats", Children: []t.TreeNode[string]{}},
				},
			},
		}),
		filterState: t.NewFilterState(),
		filterInput: t.NewTextInputState(""),
		selectedMsg: t.NewSignal("No selection yet"),
		themeIndex:  t.NewSignal(0),
	}
}

func (d *SimpleTreeDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Theme", Action: d.cycleTheme},
		{Key: "escape", Name: "Clear filter", Action: d.clearFilter},
	}
}

func (d *SimpleTreeDemo) clearFilter() {
	d.filterInput.SetText("")
	d.filterState.Query.Set("")
}

func (d *SimpleTreeDemo) cycleTheme() {
	d.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (d *SimpleTreeDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:      "simple-tree-root",
		Spacing: 1,
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Simple Tree Example",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.ParseMarkupToText("[b $Accent]↑↓[/] navigate [b $Accent]←→[/] expand [b $Accent]Shift+↑↓[/] select", theme),
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "Filter:", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.TextInput{
						ID:          "tree-filter",
						State:       d.filterInput,
						Placeholder: "Type to filter...",
						Width:       t.Cells(30),
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
						},
						OnChange: func(text string) {
							d.filterState.Query.Set(text)
						},
					},
				},
			},
			t.Tree[string]{
				ID:          "food-tree",
				Width:       t.Cells(24),
				State:       d.treeState,
				Filter:      d.filterState,
				MultiSelect: true,
				OnSelect: func(cursor string, selected []string) {
					if len(selected) > 0 {
						d.selectedMsg.Set(fmt.Sprintf("Selected %d items: %s", len(selected), strings.Join(selected, ", ")))
					} else {
						d.selectedMsg.Set(fmt.Sprintf("Selected: %s", cursor))
					}
				},
			},
			t.Text{
				Content: d.selectedMsg.Get(),
				Style:   t.Style{ForegroundColor: theme.Success},
			},
			t.ParseMarkupToText("Press [b $Accent]t[/] to change theme | [b $Error]Ctrl+C[/] to quit", theme),
		},
	}
}

func main() {
	app := NewSimpleTreeDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
