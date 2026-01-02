package terma

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
)

// KeyEvent wraps a key press event from ultraviolet.
type KeyEvent struct {
	event uv.KeyPressEvent
}

// MatchString returns true if the key matches one of the given strings.
// Examples: "enter", "tab", "a", "ctrl+a", "shift+enter", "alt+tab"
func (k KeyEvent) MatchString(s ...string) bool {
	return k.event.MatchString(s...)
}

// Key returns the underlying key string representation.
func (k KeyEvent) Key() string {
	return k.event.String()
}

// Keyed is implemented by widgets that provide a stable identity key.
// The key must be unique among siblings and persist across rebuilds.
type Keyed interface {
	Key() string
}

// Focusable is implemented by widgets that can receive keyboard focus.
type Focusable interface {
	// OnKey is called when the widget has focus and a key is pressed.
	// Return true if the key was handled, false to propagate.
	OnKey(event KeyEvent) bool

	// IsFocusable returns whether this widget can currently receive focus.
	// This allows widgets to dynamically enable/disable focus.
	IsFocusable() bool
}

// KeyHandler is implemented by widgets that want to handle key events.
// Unlike Focusable, any widget in the tree can implement this to receive
// bubbling key events from focused descendants.
type KeyHandler interface {
	OnKey(event KeyEvent) bool
}

// Clickable is implemented by widgets that respond to mouse clicks.
type Clickable interface {
	OnClick()
}

// Hoverable is implemented by widgets that respond to mouse hover.
type Hoverable interface {
	OnHover(hovered bool)
}

// FocusableEntry pairs a focusable widget with its identity key and ancestor chain.
type FocusableEntry struct {
	Key       string
	Focusable Focusable
	// Ancestors is the chain of widgets from root to this widget's parent
	// that implement KeyHandler or KeybindProvider.
	// Used for bubbling key events up the tree.
	Ancestors []Widget
}

// FocusManager tracks the currently focused widget and handles navigation.
type FocusManager struct {
	// focusables is the ordered list of focusable widgets (in tab order)
	focusables []FocusableEntry

	// focusedKey is the key of the currently focused widget ("" if none)
	focusedKey string
}

// NewFocusManager creates a new focus manager.
func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

// SetFocusables updates the list of focusable widgets.
// Called after each render to update the tab order.
func (fm *FocusManager) SetFocusables(focusables []FocusableEntry) {
	fm.focusables = focusables

	// Log collected focusables
	if len(focusables) > 0 {
		Log("SetFocusables: %d focusable widgets collected", len(focusables))
		for i, entry := range focusables {
			Log("  [%d] Key=%q Type=%T", i, entry.Key, entry.Focusable)
		}
	}

	// If we have a focused key, verify it still exists
	if fm.focusedKey != "" {
		found := false
		for _, entry := range focusables {
			if entry.Key == fm.focusedKey {
				found = true
				break
			}
		}
		if !found {
			Log("SetFocusables: focused key %q no longer exists, clearing focus", fm.focusedKey)
			fm.focusedKey = ""
		}
	}

	// If no focus and there are focusables, focus the first one
	if fm.focusedKey == "" && len(focusables) > 0 {
		fm.focusedKey = focusables[0].Key
		Log("SetFocusables: auto-focusing first widget %q", fm.focusedKey)
	}
}

// focusedIndex returns the index of the focused widget (-1 if none).
func (fm *FocusManager) focusedIndex() int {
	for i, entry := range fm.focusables {
		if entry.Key == fm.focusedKey {
			return i
		}
	}
	return -1
}

// Focused returns the currently focused widget, or nil if none.
func (fm *FocusManager) Focused() Focusable {
	for _, entry := range fm.focusables {
		if entry.Key == fm.focusedKey {
			return entry.Focusable
		}
	}
	return nil
}

// FocusedKey returns the key of the focused widget ("" if none).
func (fm *FocusManager) FocusedKey() string {
	return fm.focusedKey
}

// ActiveKeybinds returns all declarative keybindings currently active
// based on the focused widget and its ancestors.
// Keybindings are returned in order from focused widget to root,
// matching the order they would be checked when handling key events.
func (fm *FocusManager) ActiveKeybinds() []Keybind {
	// Find the focused entry
	var focusedEntry *FocusableEntry
	for i := range fm.focusables {
		if fm.focusables[i].Key == fm.focusedKey {
			focusedEntry = &fm.focusables[i]
			break
		}
	}

	if focusedEntry == nil {
		return nil
	}

	var keybinds []Keybind

	// Collect from focused widget first
	if provider, ok := focusedEntry.Focusable.(KeybindProvider); ok {
		keybinds = append(keybinds, provider.Keybinds()...)
	}

	// Then collect from ancestors (innermost to outermost/root)
	for i := len(focusedEntry.Ancestors) - 1; i >= 0; i-- {
		if provider, ok := focusedEntry.Ancestors[i].(KeybindProvider); ok {
			keybinds = append(keybinds, provider.Keybinds()...)
		}
	}

	return keybinds
}

// FocusByKey sets focus to the widget with the given key.
func (fm *FocusManager) FocusByKey(key string) {
	for _, entry := range fm.focusables {
		if entry.Key == key && entry.Focusable.IsFocusable() {
			fm.focusedKey = key
			return
		}
	}
}

// FocusNext moves focus to the next focusable widget (Tab).
func (fm *FocusManager) FocusNext() {
	if len(fm.focusables) == 0 {
		Log("FocusNext: no focusables available")
		return
	}

	oldKey := fm.focusedKey
	startIndex := fm.focusedIndex()
	if startIndex < 0 {
		startIndex = -1
	}

	// Find next focusable widget
	for i := 0; i < len(fm.focusables); i++ {
		nextIndex := (startIndex + 1 + i) % len(fm.focusables)
		if fm.focusables[nextIndex].Focusable.IsFocusable() {
			fm.focusedKey = fm.focusables[nextIndex].Key
			Log("FocusNext: %q -> %q", oldKey, fm.focusedKey)
			return
		}
	}
	Log("FocusNext: no focusable widget found")
}

// FocusPrevious moves focus to the previous focusable widget (Shift+Tab).
func (fm *FocusManager) FocusPrevious() {
	if len(fm.focusables) == 0 {
		Log("FocusPrevious: no focusables available")
		return
	}

	oldKey := fm.focusedKey
	startIndex := fm.focusedIndex()
	if startIndex < 0 {
		startIndex = 0
	}

	// Find previous focusable widget
	for i := 0; i < len(fm.focusables); i++ {
		prevIndex := (startIndex - 1 - i + len(fm.focusables)) % len(fm.focusables)
		if fm.focusables[prevIndex].Focusable.IsFocusable() {
			fm.focusedKey = fm.focusables[prevIndex].Key
			Log("FocusPrevious: %q -> %q", oldKey, fm.focusedKey)
			return
		}
	}
	Log("FocusPrevious: no focusable widget found")
}

// HandleKey routes a key event to the focused widget, bubbling up if not handled.
// For each widget in the chain, declarative keybindings (KeybindProvider) are
// checked first, then the imperative OnKey handler.
// Returns true if the key was handled.
func (fm *FocusManager) HandleKey(event KeyEvent) bool {
	Log("HandleKey: received key %q", event.Key())

	// Handle Tab navigation
	if event.MatchString("tab") {
		Log("HandleKey: tab navigation triggered")
		fm.FocusNext()
		return true
	}
	if event.MatchString("shift+tab") {
		Log("HandleKey: shift+tab navigation triggered")
		fm.FocusPrevious()
		return true
	}

	// Find the focused entry to get the ancestor chain
	var focusedEntry *FocusableEntry
	for i := range fm.focusables {
		if fm.focusables[i].Key == fm.focusedKey {
			focusedEntry = &fm.focusables[i]
			break
		}
	}

	if focusedEntry == nil {
		Log("HandleKey: no focused widget, key unhandled")
		return false
	}

	Log("HandleKey: focused widget is %q (%T), %d ancestors in chain",
		focusedEntry.Key, focusedEntry.Focusable, len(focusedEntry.Ancestors))

	// First, try the focused widget itself
	// Check declarative keybindings first
	if provider, ok := focusedEntry.Focusable.(KeybindProvider); ok {
		Log("HandleKey: checking keybindings on focused widget %q", focusedEntry.Key)
		if matchKeybind(event, provider.Keybinds()) {
			Log("HandleKey: key %q handled by keybinding on focused widget %q", event.Key(), focusedEntry.Key)
			return true
		}
	}
	// Then try imperative OnKey
	Log("HandleKey: calling OnKey on focused widget %q", focusedEntry.Key)
	if focusedEntry.Focusable.OnKey(event) {
		Log("HandleKey: key %q handled by OnKey on focused widget %q", event.Key(), focusedEntry.Key)
		return true
	}
	Log("HandleKey: focused widget %q did not handle key, bubbling up", focusedEntry.Key)

	// Bubble up through ancestors (from innermost to outermost/root)
	for i := len(focusedEntry.Ancestors) - 1; i >= 0; i-- {
		ancestor := focusedEntry.Ancestors[i]
		ancestorKey := ""
		if keyed, ok := ancestor.(Keyed); ok {
			ancestorKey = keyed.Key()
		}

		// Check declarative keybindings first
		if provider, ok := ancestor.(KeybindProvider); ok {
			Log("HandleKey: checking keybindings on ancestor %q (%T)", ancestorKey, ancestor)
			if matchKeybind(event, provider.Keybinds()) {
				Log("HandleKey: key %q handled by keybinding on ancestor %q", event.Key(), ancestorKey)
				return true
			}
		}

		// Then try imperative OnKey
		if handler, ok := ancestor.(KeyHandler); ok {
			Log("HandleKey: calling OnKey on ancestor %q (%T)", ancestorKey, ancestor)
			if handler.OnKey(event) {
				Log("HandleKey: key %q handled by OnKey on ancestor %q", event.Key(), ancestorKey)
				return true
			}
		}
	}

	Log("HandleKey: key %q not handled by any widget", event.Key())
	return false
}

// FocusCollector collects focusable widgets during render traversal.
type FocusCollector struct {
	focusables []FocusableEntry
	// path tracks the current position in the widget tree for auto-keys
	path []int
	// ancestorStack tracks widgets that implement KeyHandler or KeybindProvider
	// from root to current position
	ancestorStack []Widget
}

// NewFocusCollector creates a new focus collector.
func NewFocusCollector() *FocusCollector {
	return &FocusCollector{
		path: []int{0},
	}
}

// PushChild enters a child widget context for key generation.
func (fc *FocusCollector) PushChild(index int) {
	fc.path = append(fc.path, index)
}

// PopChild exits the current child widget context.
func (fc *FocusCollector) PopChild() {
	if len(fc.path) > 1 {
		fc.path = fc.path[:len(fc.path)-1]
	}
}

// PushAncestor adds a widget to the ancestor chain.
// Called when entering a widget that implements KeyHandler or KeybindProvider.
func (fc *FocusCollector) PushAncestor(widget Widget) {
	fc.ancestorStack = append(fc.ancestorStack, widget)
}

// PopAncestor removes the last widget from the ancestor chain.
// Called when exiting a widget that implements KeyHandler or KeybindProvider.
func (fc *FocusCollector) PopAncestor() {
	if len(fc.ancestorStack) > 0 {
		fc.ancestorStack = fc.ancestorStack[:len(fc.ancestorStack)-1]
	}
}

// ShouldTrackAncestor returns true if the widget should be added to the ancestor chain.
// A widget is tracked if it implements KeyHandler or KeybindProvider.
func (fc *FocusCollector) ShouldTrackAncestor(widget Widget) bool {
	_, isHandler := widget.(KeyHandler)
	_, isProvider := widget.(KeybindProvider)
	return isHandler || isProvider
}

// currentPath returns the current path as a string for auto-key generation.
func (fc *FocusCollector) currentPath() string {
	result := ""
	for i, idx := range fc.path {
		if i > 0 {
			result += "."
		}
		result += fmt.Sprintf("%d", idx)
	}
	return result
}

// Collect adds a focusable widget to the collection.
// If the widget implements Keyed, its key is used; otherwise an auto-key is generated.
func (fc *FocusCollector) Collect(widget Widget) {
	focusable, ok := widget.(Focusable)
	if !ok || !focusable.IsFocusable() {
		return
	}

	// Get the widget's key, falling back to auto-key if not provided or empty
	var key string
	if keyed, ok := widget.(Keyed); ok && keyed.Key() != "" {
		key = keyed.Key()
	} else {
		// Generate auto-key from tree position
		key = "_auto:" + fc.currentPath()
	}

	// Copy the current ancestor chain for this focusable
	var ancestors []Widget
	if len(fc.ancestorStack) > 0 {
		ancestors = make([]Widget, len(fc.ancestorStack))
		copy(ancestors, fc.ancestorStack)
	}

	fc.focusables = append(fc.focusables, FocusableEntry{
		Key:       key,
		Focusable: focusable,
		Ancestors: ancestors,
	})
}

// Focusables returns all collected focusable entries.
func (fc *FocusCollector) Focusables() []FocusableEntry {
	return fc.focusables
}

// Reset clears the collected focusables for a new render pass.
func (fc *FocusCollector) Reset() {
	fc.focusables = fc.focusables[:0]
	fc.path = fc.path[:1]
	fc.path[0] = 0
	fc.ancestorStack = fc.ancestorStack[:0]
}
