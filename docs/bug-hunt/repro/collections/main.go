package main

import (
	"fmt"
	"os"

	t "github.com/darrenburns/terma"
)

type app struct {
	mode   string
	state  *t.ListState[string]
	filter *t.FilterState
	status t.Signal[string]
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "x", Name: "Apply change", Action: func() {
		switch a.mode {
		case "cursor":
			a.filter.Query.Set("b")
			a.status.Set("Expected: banana has the > cursor. Stored cursor remains apple.")
		case "case":
			a.filter.CaseSensitive.Set(true)
			a.status.Set("Expected: only apricot. Uppercase Apple no longer matches lowercase a.")
		case "mode":
			a.filter.Mode.Set(t.FilterFuzzy)
			a.status.Set("Expected: apple now matches ae in fuzzy mode.")
		case "items":
			a.state.Items.Set([]string{"pear", "apple"})
			a.status.Set("Expected: apple remains the only match after replacing the source items.")
		}
	}}}
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	return t.Column{Spacing: 1, Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
		t.Text{Content: "List regression reproduction"},
		t.Text{Content: fmt.Sprintf("Scenario: %s. Press x to apply the change. Use arrows to navigate.", a.mode)},
		t.SignalText(a.status, func(status string) string { return status }),
		t.List[string]{ID: "repro-list", State: a.state, Filter: a.filter, CursorStyle: t.CursorStyle{CursorPrefix: "> "}},
	}}
}

func main() {
	mode := "cursor"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	a := &app{mode: mode, filter: t.NewFilterState(), status: t.NewSignal("Press x to trigger the reproduction.")}
	switch mode {
	case "case":
		a.state = t.NewListState([]string{"Apple", "apricot"})
		a.filter.Query.Set("a")
	case "mode":
		a.state = t.NewListState([]string{"apple", "pear"})
		a.filter.Query.Set("ae")
	case "items":
		a.state = t.NewListState([]string{"apple", "pear"})
		a.filter.Query.Set("app")
	default:
		a.state = t.NewListState([]string{"apple", "banana", "cherry", "blueberry"})
	}
	t.Run(a)
}
