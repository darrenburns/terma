package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type selectHintsProbe struct{ state *SelectState[int] }

func (s *selectHintsProbe) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		SelectBox[int]{ID: "choice", State: s.state, Options: []SelectOption[int]{{Label: "Alpha", Value: 1}}, Style: Style{Width: Cells(20)}},
		Text{Content: "Hints", Height: Cells(7)},
		KeybindBar{},
	}}
}

func TestSelectBoxProbeHintsFollowPopup(t *testing.T) {
	seq := newReactivitySequence(t, 80, 12, func() *selectHintsProbe { return &selectHintsProbe{state: NewSelectState[int]()} })
	seq.frame("closed", nil)
	require.Contains(t, seq.actual.renderer.ScreenText(), "enter Choose")
	for _, side := range []*reactivitySurface[*selectHintsProbe]{seq.actual, seq.expected} {
		dispatchKey(side.renderer, side.focus, side.root, makeKeyEvent(uv.KeyEnter, 0))
	}
	seq.frame("open hints", nil)
	require.Contains(t, seq.actual.renderer.ScreenText(), "enter Commit escape Cancel")
	assertBufferSnapshot(t, "SelectBox_Probe_ReactiveHints", seq.actual.buffer, 80, 12, DefaultSVGOptions(), "Opening SelectBox updates its KeybindBar to Commit and Cancel without changing focus")
	for _, side := range []*reactivitySurface[*selectHintsProbe]{seq.actual, seq.expected} {
		dispatchKey(side.renderer, side.focus, side.root, makeKeyEvent(uv.KeyEscape, 0))
	}
	seq.frame("cancel hints", nil)
	require.Contains(t, seq.actual.renderer.ScreenText(), "enter Choose")
}

type selectDialogProbe struct {
	state         *SelectState[int]
	visible       Signal[bool]
	outsideClicks int
}

func (s *selectDialogProbe) Build(BuildContext) Widget {
	return Dialog{ID: "select-probe-dialog", Visible: s.visible.Get(), OnDismiss: func() { s.visible.Set(false) }, Title: "Select probe", Style: Style{Width: Cells(52)}, Content: Column{Children: []Widget{
		Text{Content: "Outside popup", Height: Cells(2), Click: func(MouseEvent) { s.outsideClicks++ }},
		SelectBox[int]{ID: "choice", State: s.state, Options: []SelectOption[int]{{Label: "Alpha", Value: 1}, {Label: "Beta", Value: 2}}, Style: Style{Width: Cells(20)}},
		Text{Content: "Space below", Height: Cells(6)},
	}}}
}

func TestSelectBoxProbeOutsideClickWithinDialogCancels(t *testing.T) {
	root := &selectDialogProbe{state: NewSelectState[int](), visible: NewSignal(true)}
	p := NewPilot(t, root, 70, 20)
	p.session.focus.FocusByID("choice")
	p.settle()
	p.Press("enter")
	require.Contains(t, p.ScreenText(), "Alpha")
	// Click visible non-focusable dialog content above the popup.
	trigger, ok := p.Bounds("choice")
	require.True(t, ok)
	p.ClickAt(trigger.X+1, trigger.Y-2)
	require.NotContains(t, p.ScreenText(), "Alpha", "outside click cancels even inside a parent modal")
	_, set := root.state.Value()
	require.False(t, set)
	require.Equal(t, "choice", p.FocusedID())
	require.True(t, root.visible.Get(), "parent dialog stays open")
	require.Zero(t, root.outsideClicks, "dismissing press does not activate underlying content")
	p.AssertSnapshot("DialogCancellation", "Outside click closes only the SelectBox popup and leaves its parent Dialog open")
}

type selectNestedProbe struct {
	first, second *SelectState[int]
	options       AnySignal[[]SelectOption[int]]
	scroll        *ScrollState
	changes       []int
}

func newSelectNestedProbe() *selectNestedProbe {
	return &selectNestedProbe{
		first: NewSelectState[int](), second: NewSelectState[int](), scroll: NewScrollState(),
		options: NewAnySignal([]SelectOption[int]{{Label: "Alpha", Value: 1}, {Label: "Locked", Value: 2, Disabled: true}, {Label: "日本語", Value: 3}, {Label: "Café", Value: 4}}),
	}
}
func (s *selectNestedProbe) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Button{ID: "outside", Label: "Outside trap"},
		FocusTrap{ID: "scope", Active: true, Child: Scrollable{ID: "viewport", State: s.scroll, Height: Cells(7), Child: Column{Style: Style{Padding: EdgeInsets{Left: 2, Top: 1}}, Children: []Widget{
			Row{Spacing: 3, Children: []Widget{
				SelectBox[int]{ID: "first", State: s.first, Options: s.options.Get(), Searchable: true, Style: Style{Width: Cells(20), Border: RoundedBorder(RGB(120, 120, 120)), Padding: EdgeInsets{Left: 1, Right: 1}}, OnChange: func(v int) { s.changes = append(s.changes, v) }},
				SelectBox[int]{ID: "second", State: s.second, Options: []SelectOption[int]{{Label: "Second only", Value: 9}}, Style: Style{Width: Cells(18)}},
			}},
			Text{Content: "Scrollable content below", Height: Cells(12)},
		}}}},
		Text{Content: "Below clipped viewport"},
	}}
}

func TestSelectBoxProbeNestedFocusAndOptionUpdates(t *testing.T) {
	seq := newReactivitySequence(t, 68, 18, newSelectNestedProbe)
	seq.frame("initial", nil)
	seq.focus("first")
	seq.frame("focused first", nil)
	key := func(name string, event KeyEvent) {
		t.Helper()
		for _, side := range []*reactivitySurface[*selectNestedProbe]{seq.actual, seq.expected} {
			dispatchKey(side.renderer, side.focus, side.root, event)
			side.focused.Set(side.focus.Focused())
		}
		seq.frame(name, nil)
	}
	key("open", makeKeyEvent(uv.KeyEnter, 0))
	key("skip disabled", makeKeyEvent(uv.KeyDown, 0))
	require.Contains(t, seq.actual.renderer.ScreenText(), "› 日本語")
	assertBufferSnapshot(t, "SelectBox_Probe_NestedLayout", seq.actual.buffer, 68, 18, DefaultSVGOptions(), "SelectBox popup escapes Scrollable clipping; bordered padded Row and FocusTrap preserve Unicode selection")
	seq.frame("reorder while open", func(s *selectNestedProbe) {
		s.options.Set([]SelectOption[int]{{Label: "日本語", Value: 3}, {Label: "Alpha", Value: 1}, {Label: "Café", Value: 4}})
	})
	require.Contains(t, seq.actual.renderer.ScreenText(), "› 日本語")
	seq.frame("disable active while open", func(s *selectNestedProbe) {
		s.options.Set([]SelectOption[int]{{Label: "日本語", Value: 3, Disabled: true}, {Label: "Alpha", Value: 1}, {Label: "Café", Value: 4}})
	})
	require.Contains(t, seq.actual.renderer.ScreenText(), "› Alpha")
	key("commit fallback", makeKeyEvent(uv.KeyEnter, 0))
	value, set := seq.actual.root.first.Value()
	require.True(t, set)
	require.Equal(t, 1, value)
	require.Equal(t, []int{1}, seq.actual.root.changes)
	key("open again", makeKeyEvent(uv.KeyEnter, 0))
	key("search", makeCharEvent('é'))
	require.Contains(t, seq.actual.renderer.ScreenText(), "› Café")
	key("tab cancels into second", makeKeyEvent(uv.KeyTab, 0))
	require.Equal(t, "second", seq.actual.focus.FocusedID())
	require.NotContains(t, seq.actual.renderer.ScreenText(), "Search:")
	key("open second", makeKeyEvent(uv.KeyEnter, 0))
	require.Contains(t, seq.actual.renderer.ScreenText(), "› Second only")
	key("reverse tab cancels into first", makeKeyEvent(uv.KeyTab, uv.ModShift))
	require.Equal(t, "first", seq.actual.focus.FocusedID())
	key("reverse wraps within trap", makeKeyEvent(uv.KeyTab, uv.ModShift))
	require.Equal(t, "second", seq.actual.focus.FocusedID())
	_, set = seq.actual.root.second.Value()
	require.False(t, set)
	require.Equal(t, []int{1}, seq.actual.root.changes)
}

func TestSelectBoxProbeClippedTriggerDoesNotReceiveClick(t *testing.T) {
	state := NewSelectState[int]()
	scroll := NewScrollState()
	root := Scrollable{State: scroll, Height: Cells(4), Child: Column{Style: Style{Padding: EdgeInsets{Left: 2}}, Children: []Widget{
		SelectBox[int]{ID: "clipped-choice", State: state, Options: []SelectOption[int]{{Label: "Alpha", Value: 1}}, Style: Style{Width: Cells(20)}},
		Text{Content: "Filler", Height: Cells(12)},
	}}}
	p := NewPilot(t, root, 36, 8)
	scroll.SetOffset(5)
	require.NotContains(t, p.ScreenText(), "Choose an option")
	p.ClickAt(3, 0)
	require.NotContains(t, p.ScreenText(), "Alpha")
	_, set := state.Value()
	require.False(t, set)
}

// Fuzz through public draft/commit APIs, checking independent safety invariants:
// only Enter may commit, commits must belong to an enabled visible choice,
// drafts/cancellation preserve the last committed value, and programmatic
// changes never emit UI callbacks. Rendering every step exercises tiny layouts.
func FuzzSelectBoxDraftIsolation(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0, 0, 2, 1, 2, 3}, {3, 9, 6, 8, 10, 2, 11, 2}, {255, 8, 8, 8, 0, 0, 9, 2},
		{8, 2, 0, 4, 0, 5, 0, 6, 6, 10, 2, 12, 2},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 80 {
			data = data[:80]
		}
		state := NewSelectState[int]()
		widget := SelectBox[int]{ID: "fuzz-choice", State: state, Searchable: true, MaxVisible: 3, Style: Style{Width: Cells(14)}}
		labels := []string{"Alpha", "Beta", "日本語", "Café", "", "Alpha"}
		for i, b := range data {
			if i == 6 {
				break
			}
			widget.Options = append(widget.Options, SelectOption[int]{Label: labels[int(b)%len(labels)], Value: int(b) % 4, Disabled: b&8 != 0})
		}
		value, set := 0, false
		query := ""
		open := false
		permitCommit := false
		widget.OnChange = func(v int) {
			require.True(t, permitCommit, "only Enter may emit a commit")
			actual, actualSet := state.Value()
			require.True(t, actualSet)
			require.Equal(t, v, actual, "state is updated before callback")
			require.True(t, !set || value != v, "same value must not notify")
			valid := false
			for _, option := range widget.Options {
				if !option.Disabled && option.Value == v && (query == "" || strings.Contains(strings.ToLower(option.Label), query)) {
					valid = true
				}
			}
			require.True(t, valid, "committed option must be enabled and match search")
			value, set = v, true
		}
		width, height := 2, 2
		if len(data) > 0 {
			width += int(data[0]) % 40
			height += int(data[0]) % 14
		}
		for _, b := range data {
			permitCommit = false
			switch b % 14 {
			case 0, 1:
				code := uv.KeyDown
				if b%14 == 1 {
					code = uv.KeyUp
				}
				widget.OnKey(makeKeyEvent(code, 0))
				if !open {
					open = true
					query = ""
				}
			case 2:
				permitCommit = open
				canCommit := false
				for _, option := range widget.Options {
					if !option.Disabled && (query == "" || strings.Contains(strings.ToLower(option.Label), query)) {
						canCommit = true
					}
				}
				widget.OnKey(makeKeyEvent(uv.KeyEnter, 0))
				open = !open || !canCommit
				if !open {
					query = ""
				}
			case 3:
				widget.OnKey(makeKeyEvent(uv.KeyEscape, 0))
				open = false
				query = ""
			case 4:
				value, set = int(b)%5, true
				state.SetValue(value)
				open = false
				query = ""
			case 5:
				state.Clear()
				value, set = 0, false
				open = false
				query = ""
			case 6:
				widget.OnKey(makeCharEvent('a'))
				query += "a"
				open = true
			case 7:
				widget.OnBlur()
				open = false
				query = ""
			case 8:
				for i := range widget.Options {
					widget.Options[i].Disabled = !widget.Options[i].Disabled
				}
			case 9:
				for i, j := 0, len(widget.Options)-1; i < j; i, j = i+1, j-1 {
					widget.Options[i], widget.Options[j] = widget.Options[j], widget.Options[i]
				}
			case 10:
				widget.OnKey(makeKeyEvent(uv.KeyBackspace, 0))
				if open && query != "" {
					query = query[:len(query)-1]
				}
			case 11:
				widget.OnKey(makeKeyEvent('u', uv.ModCtrl))
				if open {
					query = ""
				}
			case 12:
				widget.Options = nil
			case 13:
				widget.OnKey(makeKeyEvent(uv.KeyTab, 0))
				open = false
				query = ""
			}
			actual, actualSet := state.Value()
			require.Equal(t, set, actualSet)
			if set {
				require.Equal(t, value, actual)
			}
			RenderToBuffer(Column{Style: Style{Padding: EdgeInsets{Left: 1}}, Children: []Widget{widget}}, width, height)
		}
	})
}
