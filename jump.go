package terma

import (
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
)

// DefaultJumpKey is the key that toggles jump mode when Jumper.Key is unset.
const DefaultJumpKey = "ctrl+o"

// DefaultJumpHints are the characters dynamic jump hints are made from when
// Jumper.Hints is unset: the home row first, then the rows above and below.
const DefaultJumpHints = "asdfghjklqwertyuiopzxcvbnm"

// JumpTarget is one entry in a static jump map: typing Key while jump mode is
// active jumps to the widget with the given ID.
type JumpTarget struct {
	// Key is the text to type, usually one character such as "1" or "u".
	// Longer keys ("gt") are typed one character at a time. No key should be
	// a prefix of another, or the longer one can never be reached.
	Key string
	// ID is the widget to label and jump to. If it isn't focusable itself
	// (a panel, say), the jump focuses the first focusable widget inside it.
	ID string
	// Action, if set, runs instead of moving focus: switching a tab, pressing
	// a button, and so on. It is labelled at the widget with ID.
	Action func()
}

// Jumpable is implemented by the items inside a widget that dynamic jump
// hints label one by one: List rows, Tree nodes, Table rows (or cells, in
// cursor selection mode) and tabs. These are what dynamic hints are for,
// content that isn't known until the app runs.
//
// Typing an item's hint calls Jump, then focuses the focusable widget
// holding the item, such as its List. Implement Jumpable on your own widgets
// to make them jump targets; Jump might select a card or open a link.
type Jumpable interface {
	Jump()
}

// jumpOptional lets a Jumpable decline a hint in its current configuration,
// as a table cell outside the first column does in row selection mode.
type jumpOptional interface {
	jumpable() bool
}

// JumpState holds the state of jump mode for a Jumper.
// Create it with NewJumpState and keep it for the life of the app.
type JumpState struct {
	active Signal[bool]
	typed  Signal[string]

	// labels are the targets on screen as of the overlay's last paint, which
	// is where their positions become known. Keys typed in jump mode are
	// matched against them.
	labels []jumpLabel
}

// NewJumpState creates a jump state with jump mode inactive.
func NewJumpState() *JumpState {
	return &JumpState{
		active: NewSignal(false),
		typed:  NewSignal(""),
	}
}

// Activate enters jump mode, labelling every target on screen.
func (s *JumpState) Activate() {
	s.typed.Set("")
	s.active.Set(true)
}

// Deactivate leaves jump mode without jumping.
func (s *JumpState) Deactivate() {
	s.active.Set(false)
	s.typed.Set("")
}

// Toggle enters jump mode if it is inactive, and leaves it otherwise.
func (s *JumpState) Toggle() {
	if s.active.Peek() {
		s.Deactivate()
	} else {
		s.Activate()
	}
}

// IsActive reports whether jump mode is active.
// Reading it during Build subscribes to changes.
func (s *JumpState) IsActive() bool {
	return s.active.Get()
}

// Typed returns what has been typed so far towards a multi-character key.
// Reading it during Build subscribes to changes.
func (s *JumpState) Typed() string {
	return s.typed.Get()
}

// handleKey takes every key press while jump mode is active. Typing a
// target's key jumps to it; typing the start of one or more keys narrows the
// labels to those; anything else leaves jump mode.
func (s *JumpState) handleKey(event KeyEvent, toggleKey string) bool {
	if event.MatchString("escape") || event.MatchString(toggleKey) {
		s.Deactivate()
		return true
	}
	if event.MatchString("backspace") {
		typed := []rune(s.typed.Peek())
		if len(typed) > 0 {
			s.typed.Set(string(typed[:len(typed)-1]))
		}
		return true
	}

	text := event.Text()
	if text == "" || text == " " {
		s.Deactivate()
		return true
	}
	typed := s.typed.Peek() + text
	for _, label := range s.labels {
		if label.key == typed {
			s.jump(label)
			return true
		}
	}
	for _, label := range s.labels {
		if strings.HasPrefix(label.key, typed) {
			s.typed.Set(typed)
			return true
		}
	}
	s.Deactivate()
	return true
}

func (s *JumpState) jump(label jumpLabel) {
	s.Deactivate()
	if label.action != nil {
		label.action()
		return
	}
	RequestFocus(label.focusID)
}

// Jumper adds a jump mode to its Child, inspired by Posting's jump mode and
// Vimium's link hints. Pressing Key (ctrl+o by default) overlays a short
// label on each target; typing a label moves focus straight to its target.
// Escape, the toggle key, a click or any key that matches no label leaves
// jump mode.
//
// Targets come from two places, which can be combined:
//
//   - Targets is a static jump map. Each key always leads to the same widget,
//     so it can be learned: "1" for the sidebar, "u" for the URL bar.
//   - Dynamic labels what is on screen that no key could be assigned to in
//     advance, as Vimium does with links: each visible row of a List, Tree
//     or Table, each tab, and anything else implementing Jumpable, plus any
//     other focusable widget without a static key. Jumping to a row moves
//     the cursor there and focuses its list. Hints are made from Hints, are
//     as short as the number of targets allows, and are assigned in reading
//     order. Longer hints are typed one character at a time, with the labels
//     narrowing as you go.
//
// Only widgets that are on screen are labelled. While focus is inside a
// modal (or any focus trap), only targets inside it are labelled. The overlay
// sits above every other overlay, and doesn't take focus: the focused widget
// keeps its focus styling while you choose where to go.
//
// Example:
//
//	Jumper{
//	    State: a.jump,
//	    Targets: []JumpTarget{
//	        {Key: "1", ID: "sidebar"},
//	        {Key: "u", ID: "url"},
//	        {Key: "r", ID: "response-tabs", Action: a.showResponse},
//	    },
//	    Dynamic: true,
//	    Child:   a.body(),
//	}
type Jumper struct {
	State   *JumpState   // Required
	Targets []JumpTarget // Static jump map
	// Dynamic labels the rows, tabs and other Jumpable items in view, and
	// any other focusable widget that isn't in Targets.
	Dynamic bool
	// Hints are the characters dynamic hints are made from.
	// Defaults to DefaultJumpHints.
	Hints string
	// Key toggles jump mode. Defaults to DefaultJumpKey. It is bound on the
	// Jumper, so it works while focus is anywhere inside Child. Call
	// State.Activate from your own keybind to enter jump mode another way.
	Key   string
	Child Widget
	// LabelStyle styles the labels. Unset colors default to the theme's
	// accent; labels are always padded by one cell on each side.
	LabelStyle Style
	// Backdrop dims the screen beneath the labels. Defaults to a light tint
	// of the theme's Overlay; use a fully transparent color to turn it off.
	Backdrop Color
}

// jumpOverlayID identifies the overlay in hit testing.
const jumpOverlayID = "__jump_overlay"

func (j Jumper) toggleKey() string {
	if j.Key == "" {
		return DefaultJumpKey
	}
	return j.Key
}

// Build registers the label overlay while jump mode is active and returns
// Child unchanged, so a Jumper never affects layout.
func (j Jumper) Build(ctx BuildContext) Widget {
	if j.State != nil && j.State.active.Get() && ctx.floatCollector != nil {
		state, toggleKey := j.State, j.toggleKey()
		ctx.floatCollector.Add(FloatEntry{
			Config: FloatConfig{Position: FloatPositionTopLeft},
			Child:  jumpOverlay{jumper: j, typed: j.State.typed.Get()},
			// Above every other overlay, taking keys before the focused widget.
			topmost: true,
			captureKey: func(event KeyEvent) bool {
				return state.handleKey(event, toggleKey)
			},
		})
	}
	if j.Child == nil {
		return EmptyWidget{}
	}
	return j.Child
}

// Keybinds binds the toggle key, showing it in the KeybindBar while focus is
// inside Child.
func (j Jumper) Keybinds() []Keybind {
	if j.State == nil {
		return nil
	}
	return []Keybind{{Key: j.toggleKey(), Name: "Jump", Action: j.State.Toggle}}
}

// globalKeybinds binds the toggle key wherever focus is, such as in a dialog,
// whose overlay isn't inside Child.
func (j Jumper) globalKeybinds() []Keybind {
	return j.Keybinds()
}

// jumpLabel is a target resolved against what is on screen.
type jumpLabel struct {
	key     string
	focusID string // Widget focused by the jump ("" when action is set)
	action  func()
	at      Rect // Visible area of the labelled widget
}

// jumpOverlay covers the screen while jump mode is active, drawing a label
// on each target and swallowing clicks.
type jumpOverlay struct {
	jumper Jumper
	typed  string
}

func (o jumpOverlay) WidgetID() string              { return jumpOverlayID }
func (o jumpOverlay) Build(ctx BuildContext) Widget { return o }

// BuildLayoutNode fills the screen, so clicks anywhere reach the overlay.
func (o jumpOverlay) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	return &layout.BoxNode{ExpandWidth: true, ExpandHeight: true}
}

// OnClick leaves jump mode: the overlay covers everything beneath it.
func (o jumpOverlay) OnClick(MouseEvent) {
	o.jumper.State.Deactivate()
}

// Render resolves the targets against this frame's layout, then draws them.
// The overlay is painted after everything else, so every widget on screen is
// in the registry by now.
func (o jumpOverlay) Render(ctx *RenderContext) {
	theme := getTheme()
	backdrop := o.jumper.Backdrop
	if !backdrop.IsSet() {
		backdrop = theme.Overlay.WithAlpha(0.4)
	}
	ctx.DrawBackdrop(0, 0, ctx.Width, ctx.Height, backdrop)

	var focusables []FocusableEntry
	if ctx.focusCollector != nil {
		focusables = ctx.focusCollector.Focusables()
	}
	focusedID := ""
	if ctx.focusManager != nil {
		focusedID = ctx.focusManager.FocusedID()
	}
	labels := resolveJumpLabels(o.jumper, ctx.widgetRegistry, focusables, focusedID)
	o.jumper.State.labels = labels

	style := o.jumper.LabelStyle
	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Accent
	}
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = theme.TextOnAccent
	}
	typedStyle := style
	typedStyle.Bold = false
	if fg, ok := style.ForegroundColor.(Color); ok {
		typedStyle.ForegroundColor = fg.WithAlpha(0.45)
	}

	screen := Rect{Width: ctx.Width, Height: ctx.Height}
	var placed []Rect
	for _, label := range labels {
		if !strings.HasPrefix(label.key, o.typed) {
			continue
		}
		width := ansi.StringWidth(label.key) + 2
		rect := placeJumpLabel(Rect{X: label.at.X - ctx.X, Y: label.at.Y - ctx.Y, Width: width, Height: 1}, placed, screen)
		placed = append(placed, rect)

		ctx.DrawStyledText(rect.X, rect.Y, " ", style)
		x := rect.X + 1
		ctx.DrawStyledText(x, rect.Y, o.typed, typedStyle)
		x += ansi.StringWidth(o.typed)
		rest := label.key[len(o.typed):]
		ctx.DrawStyledText(x, rect.Y, rest+" ", style)
	}
}

// placeJumpLabel moves a label right past any label already placed that it
// would cover, keeping it on screen.
func placeJumpLabel(rect Rect, placed []Rect, screen Rect) Rect {
	for moved := true; moved; {
		moved = false
		for _, other := range placed {
			if rect.Intersects(other) {
				rect.X = other.X + other.Width
				moved = true
			}
		}
	}
	rect.X = min(rect.X, screen.Width-rect.Width)
	rect.X = max(rect.X, 0)
	return rect
}

// resolveJumpLabels works out which targets are on screen, what each jumps
// to, and the key for each. Static targets come first, in the order given,
// then dynamic ones in reading order.
func resolveJumpLabels(j Jumper, registry *WidgetRegistry, focusables []FocusableEntry, focusedID string) []jumpLabel {
	if registry == nil {
		return nil
	}
	onScreen := make(map[string]*WidgetEntry, len(registry.entries))
	for i := range registry.entries {
		entry := &registry.entries[i]
		if entry.ID == "" || entry.ID == jumpOverlayID {
			continue
		}
		if _, seen := onScreen[entry.ID]; !seen {
			onScreen[entry.ID] = entry
		}
	}

	// Jumps stay inside the focus trap (a modal, say) that holds focus.
	trapID := ""
	for _, entry := range focusables {
		if entry.ID == focusedID {
			trapID = entry.TrapID
			break
		}
	}
	var candidates []FocusableEntry
	for _, entry := range focusables {
		if entry.TrapID != trapID || !entry.Focusable.IsFocusable() {
			continue
		}
		if widget := onScreen[entry.ID]; widget != nil && !widget.Visible.IsEmpty() {
			candidates = append(candidates, entry)
		}
	}
	isCandidate := func(id string) bool {
		return slices.ContainsFunc(candidates, func(entry FocusableEntry) bool { return entry.ID == id })
	}

	var labels []jumpLabel
	taken := make(map[string]bool)
	reserved := make([]string, 0, len(j.Targets))
	for _, target := range j.Targets {
		if target.Key == "" {
			continue
		}
		reserved = append(reserved, target.Key)
		widget := onScreen[target.ID]
		if widget == nil || widget.Visible.IsEmpty() {
			continue
		}
		label := jumpLabel{key: target.Key, action: target.Action, at: widget.Visible}
		_, isItem := widget.EventWidget.(Jumpable)
		switch {
		case isCandidate(target.ID):
			label.focusID = target.ID
		case target.Action == nil && isItem:
			item, ok := jumpItem(widget, candidates, onScreen, trapID)
			if !ok {
				continue
			}
			label.action = item.action
		case target.Action == nil:
			// A container: jump to the first focusable inside it.
			for _, entry := range candidates {
				if containsRect(widget.Bounds, onScreen[entry.ID].Visible) {
					label.focusID = entry.ID
					break
				}
			}
			if label.focusID == "" {
				continue
			}
		case trapID != "":
			// An action on a widget outside the trap holding focus.
			continue
		}
		taken[target.ID] = true
		if label.focusID != "" {
			taken[label.focusID] = true
		}
		labels = append(labels, label)
	}

	if !j.Dynamic {
		return labels
	}

	// Items in view (list rows, tree nodes, tabs) are labelled one by one,
	// in place of the widget holding them: that is content nobody could have
	// mapped to a key in advance.
	var dynamic []jumpLabel
	holders := make(map[string]bool)
	for i := range registry.entries {
		entry := &registry.entries[i]
		if taken[entry.ID] {
			continue
		}
		if item, ok := jumpItem(entry, candidates, onScreen, trapID); ok {
			dynamic = append(dynamic, item)
			holders[item.focusID] = true
		}
	}
	for _, entry := range candidates {
		if !taken[entry.ID] && !holders[entry.ID] {
			dynamic = append(dynamic, jumpLabel{focusID: entry.ID, at: onScreen[entry.ID].Visible})
		}
	}
	slices.SortStableFunc(dynamic, func(a, b jumpLabel) int {
		if a.at.Y != b.at.Y {
			return a.at.Y - b.at.Y
		}
		return a.at.X - b.at.X
	})
	alphabet := j.Hints
	if alphabet == "" {
		alphabet = DefaultJumpHints
	}
	hints := jumpHints(len(dynamic), alphabet, reserved)
	for i := range hints {
		dynamic[i].key = hints[i]
		labels = append(labels, dynamic[i])
	}
	return labels
}

// jumpItem resolves a Jumpable on screen. Jumping calls Jump, then focuses
// the innermost focusable holding the item, reported as the label's
// focusID. An item outside every focusable is only reachable while no focus
// trap is active.
func jumpItem(entry *WidgetEntry, candidates []FocusableEntry, onScreen map[string]*WidgetEntry, trapID string) (jumpLabel, bool) {
	item, ok := entry.EventWidget.(Jumpable)
	if !ok || entry.Visible.IsEmpty() {
		return jumpLabel{}, false
	}
	if optional, ok := item.(jumpOptional); ok && !optional.jumpable() {
		return jumpLabel{}, false
	}
	holder := ""
	holderArea := 0
	for _, candidate := range candidates {
		bounds := onScreen[candidate.ID].Bounds
		area := bounds.Width * bounds.Height
		if containsRect(bounds, entry.Visible) && (holder == "" || area < holderArea) {
			holder, holderArea = candidate.ID, area
		}
	}
	if holder == "" && trapID != "" {
		return jumpLabel{}, false
	}
	return jumpLabel{
		focusID: holder,
		at:      entry.Visible,
		action: func() {
			item.Jump()
			if holder != "" {
				RequestFocus(holder)
			}
		},
	}, true
}

func containsRect(outer, inner Rect) bool {
	return !inner.IsEmpty() &&
		inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.Width <= outer.X+outer.Width &&
		inner.Y+inner.Height <= outer.Y+outer.Height
}

// jumpHints returns up to n hints made from the characters of alphabet. No
// hint is a prefix of another or of a reserved key, and no reserved key is a
// prefix of a hint, so typing a hint always leads to exactly one target.
// Hints are as short as possible, and shorter hints come first.
//
// Like Vimium, it starts from one hint per character and, while there are
// too few, replaces the shortest hint with every one-character extension of
// it.
func jumpHints(n int, alphabet string, reserved []string) []string {
	var chars []string
	for _, r := range alphabet {
		if c := string(r); !slices.Contains(chars, c) {
			chars = append(chars, c)
		}
	}
	if n <= 0 || len(chars) == 0 {
		return nil
	}

	var hints []string
	var extend func(prefix string)
	extend = func(prefix string) {
		for _, c := range chars {
			hint := prefix + c
			usable := true
			for _, key := range reserved {
				if strings.HasPrefix(hint, key) {
					usable = false // The key would be typed first.
					break
				}
				if strings.HasPrefix(key, hint) {
					usable = false // Typing the hint would be ambiguous.
					extend(hint)
					break
				}
			}
			if usable {
				hints = append(hints, hint)
			}
		}
	}
	extend("")

	// Extending a hint gains len(chars)-1 hints, so a single character can't
	// make more than there are to begin with.
	hintLength := func(a, b string) int { return len(a) - len(b) }
	for len(hints) < n && len(hints) > 0 && len(chars) > 1 {
		shortest := slices.IndexFunc(hints, func(h string) bool {
			return len(h) == len(slices.MinFunc(hints, hintLength))
		})
		hint := hints[shortest]
		hints = slices.Delete(hints, shortest, shortest+1)
		extend(hint)
	}
	slices.SortStableFunc(hints, hintLength)
	if len(hints) > n {
		hints = hints[:n]
	}
	return hints
}
