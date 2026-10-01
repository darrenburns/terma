package terma

import (
	"fmt"
	"strings"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// SelectOption is a labelled value in a SelectBox. Values should be stable,
// reflexive comparable keys; use scalar IDs for records containing maps or slices.
type SelectOption[T comparable] struct {
	Label    string
	Value    T
	Disabled bool
}

type selectValue[T comparable] struct {
	value T
	set   bool
}

type selectCursor[T comparable] struct {
	value T
	index int
	set   bool
}

// SelectState holds a committed value independently of an in-progress choice.
// Create it once with NewSelectState, outside Build, and retain it across builds.
type SelectState[T comparable] struct {
	selected Signal[selectValue[T]]
	open     Signal[bool]
	query    Signal[string]
	cursor   Signal[selectCursor[T]]
}

// NewSelectState creates a select with no committed value.
func NewSelectState[T comparable]() *SelectState[T] {
	return &SelectState[T]{
		selected: NewSignal(selectValue[T]{}),
		open:     NewSignal(false),
		query:    NewSignal(""),
		cursor:   NewSignal(selectCursor[T]{}),
	}
}

// Value returns the committed value and whether it is set. It subscribes the
// current reactive phase to changes. The zero value is a valid selection.
func (s *SelectState[T]) Value() (T, bool) {
	v := s.selected.Get()
	return v.value, v.set
}

// SetValue sets a committed value and cancels any in-progress choice. It never
// invokes the widget's OnChange callback. Values absent from Options are retained.
func (s *SelectState[T]) SetValue(value T) {
	s.cancel()
	s.selected.Set(selectValue[T]{value: value, set: true})
}

// Clear unsets the committed value and cancels any in-progress choice.
func (s *SelectState[T]) Clear() {
	s.cancel()
	s.selected.Set(selectValue[T]{})
}

func (s *SelectState[T]) cancel() {
	s.open.Set(false)
	s.query.Set("")
	s.cursor.Set(selectCursor[T]{})
}

// SelectBox is a focusable dropdown that chooses a typed value. Highlighting and
// searching do not commit; Enter or an enabled row click commits. Escape, Tab,
// blur and outside clicks cancel. The existing Select function projects signals,
// so this widget uses the distinct name SelectBox.
type SelectBox[T comparable] struct {
	ID          string
	State       *SelectState[T]
	Options     []SelectOption[T]
	Placeholder string
	Searchable  bool
	Disabled    bool
	MaxVisible  int // Maximum visible option rows; defaults to 8.
	Style       Style
	PopupStyle  Style
	OnChange    func(T) // User commits only, once per changed value, after updating state.
}

func (s SelectBox[T]) WidgetID() string { return s.ID }

func (s SelectBox[T]) IsFocusable() bool { return s.State != nil && !s.Disabled }

func (s SelectBox[T]) GetContentDimensions() (Dimension, Dimension) {
	d := s.Style.GetDimensions()
	return d.Width, d.Height
}

// OnBlur cancels a pending choice without changing the committed value.
func (s SelectBox[T]) OnBlur() {
	if s.State != nil {
		s.State.cancel()
	}
}

func (s SelectBox[T]) maxVisible() int {
	if s.MaxVisible > 0 {
		return s.MaxVisible
	}
	return 8
}

// matching derives source indices, keeping navigation and rendering in one view.
func (s SelectBox[T]) matching(query string) []int {
	query = strings.ToLower(query)
	indices := make([]int, 0, len(s.Options))
	for i, option := range s.Options {
		if query == "" || strings.Contains(strings.ToLower(option.Label), query) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (s SelectBox[T]) activeIndex(indices []int, cursor selectCursor[T]) int {
	if cursor.set {
		// Preserve the exact occurrence of duplicate values when possible.
		for _, i := range indices {
			if i == cursor.index && !s.Options[i].Disabled && s.Options[i].Value == cursor.value {
				return i
			}
		}
		for _, i := range indices {
			if !s.Options[i].Disabled && s.Options[i].Value == cursor.value {
				return i
			}
		}
	}
	for _, i := range indices {
		if !s.Options[i].Disabled {
			return i
		}
	}
	return -1
}

func (s SelectBox[T]) highlight(index int) {
	cursor := selectCursor[T]{}
	if index >= 0 && index < len(s.Options) {
		cursor = selectCursor[T]{value: s.Options[index].Value, index: index, set: true}
	}
	s.State.cursor.Set(cursor)
}

func (s SelectBox[T]) show() {
	if !s.IsFocusable() {
		return
	}
	s.State.query.Set("")
	selected := s.State.selected.Peek()
	cursor := selectCursor[T]{value: selected.value, index: -1, set: selected.set}
	s.highlight(s.activeIndex(s.matching(""), cursor))
	s.State.open.Set(true)
}

func (s SelectBox[T]) move(delta int) {
	if !s.State.open.Peek() {
		s.show()
		return
	}
	indices := s.matching(s.State.query.Peek())
	active := s.activeIndex(indices, s.State.cursor.Peek())
	enabled := make([]int, 0, len(indices))
	position := 0
	for _, i := range indices {
		if !s.Options[i].Disabled {
			if i == active {
				position = len(enabled)
			}
			enabled = append(enabled, i)
		}
	}
	if len(enabled) > 0 {
		s.highlight(enabled[max(0, min(len(enabled)-1, position+delta))])
	}
}

func (s SelectBox[T]) commit(index int) {
	if !s.IsFocusable() || !s.State.open.Peek() || index < 0 || index >= len(s.Options) || s.Options[index].Disabled {
		return
	}
	old := s.State.selected.Peek()
	value := s.Options[index].Value
	s.State.SetValue(value)
	if (!old.set || old.value != value) && s.OnChange != nil {
		s.OnChange(value)
	}
}

func (s SelectBox[T]) setQuery(query string) {
	s.State.query.Set(query)
	s.highlight(s.activeIndex(s.matching(query), selectCursor[T]{}))
}

// Keybinds describes the actions available in the current popup state.
func (s SelectBox[T]) Keybinds() []Keybind {
	if !s.IsFocusable() {
		return nil
	}
	keys := []Keybind{
		{Key: "up", Name: "Previous", Action: func() { s.move(-1) }},
		{Key: "down", Name: "Next", Action: func() { s.move(1) }},
	}
	// KeybindBar reads this during Build; subscribe so its hints follow the popup.
	if !s.State.open.Get() {
		return append(keys,
			Keybind{Key: "enter", Name: "Choose", Action: s.show},
			Keybind{Key: " ", Name: "Choose", Action: s.show, Hidden: true})
	}
	keys = append(keys,
		Keybind{Key: "enter", Name: "Commit", Action: func() { s.commit(s.activeIndex(s.matching(s.State.query.Peek()), s.State.cursor.Peek())) }},
		Keybind{Key: "escape", Name: "Cancel", Action: s.State.cancel},
		Keybind{Key: "home", Name: "First", Action: func() { s.move(-len(s.Options)) }, Hidden: true},
		Keybind{Key: "end", Name: "Last", Action: func() { s.move(len(s.Options)) }, Hidden: true},
		Keybind{Key: "pgup", Name: "Page up", Action: func() { s.move(-s.maxVisible()) }, Hidden: true},
		Keybind{Key: "pgdown", Name: "Page down", Action: func() { s.move(s.maxVisible()) }, Hidden: true})
	if s.Searchable {
		keys = append(keys,
			Keybind{Key: "backspace", Name: "Delete", Hidden: true, Action: func() {
				runes := []rune(s.State.query.Peek())
				if len(runes) > 0 {
					s.setQuery(string(runes[:len(runes)-1]))
				}
			}},
			Keybind{Key: "ctrl+u", Name: "Clear search", Action: func() { s.setQuery("") }})
	}
	return keys
}

// CapturesKey keeps ancestor character shortcuts out of search input and hints.
func (s SelectBox[T]) CapturesKey(key string) bool {
	runes := []rune(key)
	return s.IsFocusable() && s.Searchable && len(runes) == 1 && unicode.IsPrint(runes[0])
}

// OnKey handles printable search input and leaves unrelated keys unhandled.
func (s SelectBox[T]) OnKey(event KeyEvent) bool {
	if !s.IsFocusable() {
		return false
	}
	// This also supports callers that dispatch directly, without FocusManager.
	if matchKeybind(event, s.Keybinds()) {
		return true
	}
	if event.MatchString("tab", "shift+tab") {
		s.State.cancel()
		return false
	}
	if s.Searchable && event.Text() != "" {
		for _, r := range event.Text() {
			if !unicode.IsPrint(r) {
				return false
			}
		}
		if !s.State.open.Peek() {
			s.show()
		}
		s.setQuery(s.State.query.Peek() + event.Text())
		return true
	}
	return false
}

// OnClick toggles the popup on a left click. Other buttons are ignored.
func (s SelectBox[T]) OnClick(event MouseEvent) {
	if !s.IsFocusable() || event.Button != uv.MouseLeft {
		return
	}
	if s.State.open.Peek() {
		s.State.cancel()
	} else {
		s.show()
	}
}

func (s SelectBox[T]) label() string {
	if s.State != nil {
		value, set := s.State.Value()
		if set {
			for _, option := range s.Options {
				if option.Value == value {
					return option.Label
				}
			}
			return "Unavailable selection"
		}
	}
	if s.Placeholder != "" {
		return s.Placeholder
	}
	return "Choose an option"
}

// Build derives the closed control and registers an anchored popup without
// mutating state. Keyboard focus remains on the control throughout a choice.
func (s SelectBox[T]) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	style := s.Style
	if style.ForegroundColor == nil {
		style.ForegroundColor = theme.Text
	}
	if style.BackgroundColor == nil {
		style.BackgroundColor = theme.Surface
	}
	disabled := s.Disabled || s.State == nil || ctx.IsDisabled()
	if disabled {
		style.ForegroundColor = theme.TextDisabled
	} else if ctx.IsFocused(s) {
		style.ForegroundColor = theme.SelectionText
		style.BackgroundColor = theme.ActiveCursor
	}
	id := s.ID
	if id == "" {
		id = ctx.AutoID()
	}
	open := s.State != nil && s.State.open.Get() && !disabled
	arrow := " ▾"
	if open {
		arrow = " ▴"
	}
	return Column{Width: style.Width, Children: []Widget{
		selectTrigger{Text: Text{Content: s.label() + arrow, Style: style}, click: func(event MouseEvent) {
			if !disabled && event.Button == uv.MouseLeft {
				RequestFocus(id)
				s.OnClick(event)
			}
		}},
		selectOverlay{Floating: Floating{
			Visible: open,
			Config:  FloatConfig{AnchorID: id, Anchor: AnchorBottomLeft, OnDismiss: s.OnBlur},
			BuildChild: func(ctx BuildContext, geometry FloatGeometry) Widget {
				return s.popup(ctx, geometry)
			},
		}, cancel: s.OnBlur},
	}}
}

func (s SelectBox[T]) popup(ctx BuildContext, geometry FloatGeometry) Widget {
	query := s.State.query.Get()
	indices := s.matching(query)
	active := s.activeIndex(indices, s.State.cursor.Get())
	theme := ctx.Theme()
	style := s.PopupStyle
	if style.ForegroundColor == nil {
		style.ForegroundColor = theme.Text
	}
	if style.BackgroundColor == nil {
		style.BackgroundColor = theme.Surface
	}
	if style.Border.Style == BorderNone {
		style.Border = RoundedBorder(theme.Primary)
	}
	if style.Border.Color == nil {
		style.Border.Color = theme.Primary
	}
	// A tiny terminal still gets a usable option row; decorations yield first.
	if geometry.Screen.Height < 3 || geometry.Screen.Width < 3 {
		style.Border = Border{}
		style.Padding = EdgeInsets{}
	}
	insets := 2*style.Border.Width() + style.Padding.Horizontal()
	verticalInsets := 2*style.Border.Width() + style.Padding.Vertical()
	width := max(geometry.AnchorBounds.Width, 18+insets)
	for _, i := range indices {
		width = max(width, ansi.StringWidth(s.Options[i].Label)+4+insets)
	}
	style.Width = Cells(min(max(1, geometry.Screen.Width), width))
	available := max(1, geometry.Screen.Height-verticalInsets)
	searchRow := s.Searchable && available >= 3
	if searchRow {
		available--
	}
	count := min(s.maxVisible(), available)
	footer := len(indices) > count && available >= 2
	if footer {
		count = min(s.maxVisible(), available-1)
	}
	position := 0
	for j, i := range indices {
		if i == active {
			position = j
		}
	}
	start := max(0, min(position-count+1, len(indices)-count))
	end := min(len(indices), start+count)
	children := make([]Widget, 0, count+2)
	if searchRow {
		label := "Type to search…"
		if query != "" {
			label = "Search: " + query
		}
		children = append(children, Text{Content: label, Style: Style{ForegroundColor: theme.TextMuted}})
	}
	for _, i := range indices[start:end] {
		option := s.Options[i]
		rowStyle := Style{Width: Flex(1), ForegroundColor: theme.Text}
		prefix := "  "
		if option.Disabled {
			prefix = "× "
			rowStyle.ForegroundColor = theme.TextDisabled
		} else if i == active {
			prefix = "› "
			rowStyle.BackgroundColor = theme.ActiveCursor
			rowStyle.ForegroundColor = theme.SelectionText
		}
		children = append(children, selectOptionRow{
			Text:   Text{Content: prefix + option.Label, Style: rowStyle},
			commit: func() { s.commit(i) },
		})
	}
	if len(indices) == 0 {
		label := "No matches"
		if len(s.Options) == 0 {
			label = "No options"
		}
		children = append(children, Text{Content: label, Style: Style{ForegroundColor: theme.TextMuted}})
	}
	if footer {
		children = append(children, Text{Content: fmt.Sprintf("%d–%d of %d", start+1, end, len(indices)), Style: Style{ForegroundColor: theme.TextMuted}})
	}
	return selectPopup{Column: Column{Style: style, Children: children}, move: s.move}
}

type selectOptionRow struct {
	Text
	commit func()
}

func (r selectOptionRow) Build(BuildContext) Widget { return r }
func (r selectOptionRow) OnClick(event MouseEvent) {
	if event.Button == uv.MouseLeft {
		r.commit()
	}
}

type selectPopup struct {
	Column
	move func(int)
}

func (p selectPopup) Build(BuildContext) Widget { return p }
func (p selectPopup) ChildWidgets() []Widget    { return p.Children }
func (p selectPopup) OnMouseWheel(event MouseEvent) bool {
	switch event.Button {
	case uv.MouseWheelUp:
		p.move(-1)
	case uv.MouseWheelDown:
		p.move(1)
	default:
		return false
	}
	return true
}

// Tab is processed by FocusManager before OnKey, and a lone focusable does not
// blur. Observe it in the overlay capture phase, then permit normal traversal.
type selectOverlay struct {
	Floating
	cancel func()
}

func (o selectOverlay) Build(ctx BuildContext) Widget {
	if o.Visible && ctx.floatCollector != nil {
		ctx.floatCollector.Add(FloatEntry{
			Config: o.Config, BuildChild: o.BuildChild,
			captureKey: func(event KeyEvent) bool {
				if event.MatchString("tab", "shift+tab") {
					o.cancel()
				}
				return false
			},
		})
	}
	return EmptyWidget{}
}

// A composed control needs its own leaf hit target; the enclosing Column does
// not automatically route leaf mouse clicks to the focusable SelectBox.
type selectTrigger struct {
	Text
	click func(MouseEvent)
}

func (t selectTrigger) Build(BuildContext) Widget { return t }
func (t selectTrigger) OnClick(event MouseEvent)  { t.click(event) }
