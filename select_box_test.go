package terma

import (
	"fmt"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selectTestWidget() SelectBox[int] {
	return SelectBox[int]{ID: "choice", State: NewSelectState[int](), Options: []SelectOption[int]{
		{Label: "Zero", Value: 0}, {Label: "Locked", Value: 1, Disabled: true},
		{Label: "日本語", Value: 2}, {Label: "Café Montréal", Value: 3},
	}, Searchable: true, Style: Style{Width: Cells(24)}}
}

func TestSelectStateUnsetAndZero(t *testing.T) {
	s := NewSelectState[int]()
	value, set := s.Value()
	assert.Zero(t, value)
	assert.False(t, set)
	s.SetValue(0)
	value, set = s.Value()
	assert.Zero(t, value)
	assert.True(t, set)
	s.Clear()
	_, set = s.Value()
	assert.False(t, set)
}

func TestSelectBoxCommitAndCallback(t *testing.T) {
	s := selectTestWidget()
	calls := []int{}
	s.OnChange = func(value int) {
		current, set := s.State.Value()
		assert.True(t, set)
		assert.Equal(t, value, current)
		assert.False(t, s.State.open.Peek(), "popup closes before callback")
		calls = append(calls, value)
	}
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	_, set := s.State.Value()
	assert.False(t, set, "opening is not a commit")
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, []int{0}, calls, "unset to zero is a change")
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, []int{0}, calls, "same value is not a change")
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.OnKey(makeKeyEvent(uv.KeyDown, 0))
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, []int{0, 2}, calls, "navigation skips disabled values")
	s.State.SetValue(3)
	s.State.Clear()
	assert.Equal(t, []int{0, 2}, calls, "programmatic updates don't emit UI callbacks")
}

func TestSelectBoxCancellation(t *testing.T) {
	for _, method := range []string{"escape", "tab", "shift-tab", "blur", "outside", "toggle"} {
		t.Run(method, func(t *testing.T) {
			s := selectTestWidget()
			s.State.SetValue(3)
			s.OnChange = func(int) { t.Fatal("cancellation committed") }
			s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
			s.OnKey(makeCharEvent('日'))
			switch method {
			case "escape":
				assert.True(t, s.OnKey(makeKeyEvent(uv.KeyEscape, 0)))
			case "tab":
				assert.False(t, s.OnKey(makeKeyEvent(uv.KeyTab, 0)))
			case "shift-tab":
				assert.False(t, s.OnKey(makeKeyEvent(uv.KeyTab, uv.ModShift)))
			case "blur":
				s.OnBlur()
			case "outside":
				built := s.Build(BuildContext{}).(Column)
				built.Children[1].(selectOverlay).Config.OnDismiss()
			case "toggle":
				s.OnClick(MouseEvent{Button: uv.MouseLeft})
			}
			value, set := s.State.Value()
			assert.True(t, set)
			assert.Equal(t, 3, value)
			assert.False(t, s.State.open.Peek())
			assert.Empty(t, s.State.query.Peek())
		})
	}
}

func TestSelectBoxSearchUnicodeAndNoMatches(t *testing.T) {
	s := selectTestWidget()
	for _, r := range "CAFÉ" {
		s.OnKey(makeCharEvent(r))
	}
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	value, set := s.State.Value()
	assert.True(t, set)
	assert.Equal(t, 3, value)
	s.OnKey(makeCharEvent('日'))
	s.OnKey(makeCharEvent('語'))
	assert.Equal(t, "日語", s.State.query.Peek())
	s.OnKey(makeKeyEvent(uv.KeyBackspace, 0))
	assert.Equal(t, "日", s.State.query.Peek(), "backspace removes a rune, not a byte")
	s.OnKey(makeCharEvent('x'))
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	assert.True(t, s.State.open.Peek(), "no-match Enter is inert")
	value, _ = s.State.Value()
	assert.Equal(t, 3, value)
	s.OnKey(makeKeyEvent('u', uv.ModCtrl))
	assert.Empty(t, s.State.query.Peek())
	s.OnKey(makeCharEvent('日'))
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	value, _ = s.State.Value()
	assert.Equal(t, 2, value)
}

func TestSelectBoxSearchSpaceAndNonSearchable(t *testing.T) {
	s := selectTestWidget()
	s.OnKey(makeCharEvent(' '))
	assert.True(t, s.State.open.Peek())
	assert.Empty(t, s.State.query.Peek(), "Space opens without filtering")
	for _, r := range "Café M" {
		s.OnKey(makeCharEvent(r))
	}
	assert.Equal(t, "Café M", s.State.query.Peek())
	s.OnBlur()
	s.Searchable = false
	assert.False(t, s.OnKey(makeCharEvent('a')))
	assert.False(t, s.OnKey(makeKeyEvent(uv.KeyEscape, 0)))
	assert.False(t, s.State.open.Peek())
}

func TestSelectBoxEmptyAndAllDisabled(t *testing.T) {
	for _, empty := range []bool{true, false} {
		s := selectTestWidget()
		if empty {
			s.Options = nil
		} else {
			for i := range s.Options {
				s.Options[i].Disabled = true
			}
		}
		s.OnChange = func(int) { t.Fatal("unavailable option committed") }
		for _, key := range []rune{uv.KeyEnter, uv.KeyDown, uv.KeyUp, uv.KeyHome, uv.KeyEnd, uv.KeyPgDown, uv.KeyPgUp, uv.KeyEnter} {
			require.True(t, s.OnKey(makeKeyEvent(key, 0)))
		}
		_, set := s.State.Value()
		assert.False(t, set)
		assert.True(t, s.State.open.Peek())
		s.OnKey(makeKeyEvent(uv.KeyEscape, 0))
		assert.False(t, s.State.open.Peek())
	}
}

func TestSelectBoxDynamicOptionsAndDuplicates(t *testing.T) {
	s := selectTestWidget()
	s.State.SetValue(2)
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.Options = []SelectOption[int]{{Label: "Moved", Value: 2}, {Label: "Other", Value: 8}}
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	value, _ := s.State.Value()
	assert.Equal(t, 2, value, "reordering preserves the highlighted value")
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.Options = []SelectOption[int]{{Label: "Other", Value: 8}}
	assert.Equal(t, "Unavailable selection", s.label())
	value, _ = s.State.Value()
	assert.Equal(t, 2, value, "removing options does not alter committed state")
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	value, _ = s.State.Value()
	assert.Equal(t, 8, value, "commit resolves a removed highlight against current options")
	s.Options = []SelectOption[int]{{Label: "First", Value: 8}, {Label: "Duplicate", Value: 8}, {Label: "First", Value: 9}}
	s.OnChange = func(int) { t.Fatal("duplicate value should not notify") }
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.OnKey(makeKeyEvent(uv.KeyDown, 0))
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, "First", s.label())
	s.OnChange = nil
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.Options[0].Disabled = true
	s.Options[1].Disabled = true
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	value, _ = s.State.Value()
	assert.Equal(t, 9, value, "disabled highlight falls back to enabled option")
}

func TestSelectBoxMouseAndDisabled(t *testing.T) {
	s := selectTestWidget()
	s.OnClick(MouseEvent{Button: uv.MouseRight})
	assert.False(t, s.State.open.Peek())
	s.OnClick(MouseEvent{Button: uv.MouseLeft})
	popup := s.popup(BuildContext{}, FloatGeometry{Screen: Rect{Width: 60, Height: 20}}).(selectPopup)
	row := popup.Children[2].(selectOptionRow) // search, zero, disabled
	row.OnClick(MouseEvent{Button: uv.MouseLeft})
	_, set := s.State.Value()
	assert.False(t, set)
	row = popup.Children[3].(selectOptionRow)
	row.OnClick(MouseEvent{Button: uv.MouseRight})
	_, set = s.State.Value()
	assert.False(t, set)
	row.OnClick(MouseEvent{Button: uv.MouseLeft})
	value, set := s.State.Value()
	assert.True(t, set)
	assert.Equal(t, 2, value)
	s.Disabled = true
	assert.False(t, s.IsFocusable())
	assert.False(t, s.OnKey(makeKeyEvent(uv.KeyEnter, 0)))
	s.OnClick(MouseEvent{Button: uv.MouseLeft})
	assert.False(t, s.State.open.Peek())
	s.State = nil
	assert.False(t, s.IsFocusable())
	assert.NotPanics(t, func() { RenderToBuffer(s, 20, 5) })
}

func TestSelectBoxNavigationAndWheel(t *testing.T) {
	s := selectTestWidget()
	s.MaxVisible = 2
	s.OnKey(makeKeyEvent(uv.KeyDown, 0))
	s.OnKey(makeKeyEvent(uv.KeyEnd, 0))
	assert.Equal(t, 3, s.State.cursor.Peek().value)
	s.OnKey(makeKeyEvent(uv.KeyDown, 0))
	assert.Equal(t, 3, s.State.cursor.Peek().value)
	s.OnKey(makeKeyEvent(uv.KeyPgUp, 0))
	assert.Equal(t, 0, s.State.cursor.Peek().value)
	s.OnKey(makeKeyEvent(uv.KeyPgDown, 0))
	assert.Equal(t, 3, s.State.cursor.Peek().value)
	s.OnKey(makeKeyEvent(uv.KeyHome, 0))
	popup := s.popup(BuildContext{}, FloatGeometry{Screen: Rect{Width: 40, Height: 10}}).(selectPopup)
	assert.True(t, popup.OnMouseWheel(MouseEvent{Button: uv.MouseWheelDown}))
	assert.Equal(t, 2, s.State.cursor.Peek().value)
	assert.True(t, popup.OnMouseWheel(MouseEvent{Button: uv.MouseWheelUp}))
	assert.Equal(t, 0, s.State.cursor.Peek().value)
	assert.False(t, popup.OnMouseWheel(MouseEvent{Button: uv.MouseWheelLeft}))
	_, set := s.State.Value()
	assert.False(t, set)
}

func TestSelectBoxBuildDoesNotMutateState(t *testing.T) {
	s := selectTestWidget()
	s.OnKey(makeCharEvent('日'))
	beforeOpen, beforeQuery, beforeCursor, beforeSelected := s.State.open.Peek(), s.State.query.Peek(), s.State.cursor.Peek(), s.State.selected.Peek()
	s.Options = nil
	RenderToBuffer(s, 20, 7)
	assert.Equal(t, beforeOpen, s.State.open.Peek())
	assert.Equal(t, beforeQuery, s.State.query.Peek())
	assert.Equal(t, beforeCursor, s.State.cursor.Peek())
	assert.Equal(t, beforeSelected, s.State.selected.Peek())
}

func TestSelectBoxRetainedRendering(t *testing.T) {
	seq := newReactivitySequence(t, 32, 10, func() *SelectBox[int] { s := selectTestWidget(); return &s })
	seq.frame("initial", nil)
	seq.frame("open", func(s *SelectBox[int]) { s.OnKey(makeKeyEvent(uv.KeyEnter, 0)) })
	seq.frame("navigate", func(s *SelectBox[int]) { s.OnKey(makeKeyEvent(uv.KeyDown, 0)) })
	seq.frame("filter", func(s *SelectBox[int]) { s.OnKey(makeCharEvent('C')) })
	seq.frame("no matches", func(s *SelectBox[int]) { s.OnKey(makeCharEvent('x')) })
	seq.frame("cancel", func(s *SelectBox[int]) { s.OnKey(makeKeyEvent(uv.KeyEscape, 0)) })
	seq.frame("programmatic", func(s *SelectBox[int]) { s.State.SetValue(3) })
	seq.frame("clear", func(s *SelectBox[int]) { s.State.Clear() })
}

func TestSnapshot_SelectBox_States(t *testing.T) {
	for _, name := range []string{"unset", "selected-zero", "disabled", "open", "empty", "all-disabled", "unicode-search", "no-matches", "unavailable"} {
		t.Run(name, func(t *testing.T) {
			s := selectTestWidget()
			switch name {
			case "selected-zero":
				s.State.SetValue(0)
			case "disabled":
				s.Disabled = true
			case "open":
				s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
			case "empty":
				s.Options = nil
				s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
			case "all-disabled":
				for i := range s.Options {
					s.Options[i].Disabled = true
				}
				s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
			case "unicode-search":
				s.OnKey(makeCharEvent('日'))
			case "no-matches":
				s.OnKey(makeCharEvent('x'))
			case "unavailable":
				s.State.SetValue(100)
			}
			AssertSnapshot(t, s, 36, 10, "SelectBox: "+name)
		})
	}
}

func TestSnapshot_SelectBox_Scrolling(t *testing.T) {
	s := selectTestWidget()
	s.Options = nil
	s.MaxVisible = 4
	for i := 0; i < 30; i++ {
		s.Options = append(s.Options, SelectOption[int]{Label: fmt.Sprintf("Option %02d", i), Value: i})
	}
	s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
	s.OnKey(makeKeyEvent(uv.KeyEnd, 0))
	AssertSnapshot(t, s, 32, 10, "Last option visible in a four-row window with 27–30 of 30 footer")
}

func TestSnapshot_SelectBox_NarrowAndBottom(t *testing.T) {
	for _, size := range [][2]int{{12, 6}, {2, 2}, {32, 9}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			s := selectTestWidget()
			s.Style.Width = Cells(size[0])
			s.OnKey(makeKeyEvent(uv.KeyEnter, 0))
			s.OnKey(makeKeyEvent(uv.KeyEnd, 0))
			widget := Column{Children: []Widget{Text{Content: "Above"}, s}}
			if size[1] == 9 {
				widget.Children[0] = Text{Content: "Above", Style: Style{Height: Cells(7)}}
			}
			AssertSnapshot(t, widget, size[0], size[1], "Popup clamps to viewport; active option stays in window")
		})
	}
}

func TestSelectBoxTabCancelsWithOnlyOneFocusable(t *testing.T) {
	for _, mod := range []uv.KeyMod{0, uv.ModShift} {
		seq := newReactivitySequence(t, 32, 10, func() *SelectBox[int] { s := selectTestWidget(); return &s })
		seq.frame("initial", nil)
		seq.frame("open", func(s *SelectBox[int]) { s.OnKey(makeKeyEvent(uv.KeyEnter, 0)) })
		for _, side := range []*reactivitySurface[*SelectBox[int]]{seq.actual, seq.expected} {
			dispatchKey(side.renderer, side.focus, side.root, makeKeyEvent(uv.KeyTab, mod))
			assert.False(t, side.root.State.open.Peek())
			assert.Equal(t, "choice", side.focus.FocusedID())
		}
		seq.frame("tab cancelled", nil)
	}
}

func TestSelectBoxKeybindDiscovery(t *testing.T) {
	s := selectTestWidget()
	assert.True(t, s.CapturesKey("a"))
	assert.True(t, s.CapturesKey("+"))
	assert.False(t, s.CapturesKey("ctrl+a"))
	assert.False(t, matchKeybind(makeKeyEvent(uv.KeyEscape, 0), s.Keybinds()))
	assert.True(t, matchKeybind(makeKeyEvent(uv.KeyEnter, 0), s.Keybinds()))
	assert.True(t, s.State.open.Peek())
	assert.True(t, matchKeybind(makeKeyEvent(uv.KeyEscape, 0), s.Keybinds()))
	assert.False(t, s.State.open.Peek())
}

func TestSelectBoxMouseRouting(t *testing.T) {
	s := selectTestWidget()
	previous := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previous })
	router, renderer := renderForMouse(s, 36, 12)
	router.press(uv.MouseClickEvent{X: 3, Y: 0, Button: uv.MouseLeft}, .5, .5, time.Now())
	router.release(uv.MouseReleaseEvent{X: 3, Y: 0, Button: uv.MouseLeft}, .5, .5)
	require.True(t, s.State.open.Peek(), "click on the visible trigger opens it")
	assert.Equal(t, s.ID, pendingFocusID)
	renderer.Render(s)
	router.press(uv.MouseClickEvent{X: 30, Y: 10, Button: uv.MouseLeft}, .5, .5, time.Now())
	router.release(uv.MouseReleaseEvent{X: 30, Y: 10, Button: uv.MouseLeft}, .5, .5)
	assert.False(t, s.State.open.Peek(), "outside click cancels")
	_, set := s.State.Value()
	assert.False(t, set)
}
