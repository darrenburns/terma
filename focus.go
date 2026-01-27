package terma

import uv "github.com/charmbracelet/ultraviolet"

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

// Text returns the actual text input from the key event.
// This returns the literal characters typed, including space as " ".
// Returns empty string for non-text keys like arrows, function keys, etc.
func (k KeyEvent) Text() string {
	return uv.Key(k.event).Text
}

// MouseEvent wraps a mouse interaction with click-chain metadata.
type MouseEvent struct {
	X, Y       int            // Absolute screen coordinates
	LocalX     int            // X offset within the widget (0 = left edge)
	LocalY     int            // Y offset within the widget (0 = top edge)
	Button     uv.MouseButton
	Mod        uv.KeyMod
	ClickCount int // 1=single, 2=double, 3=triple, etc
	WidgetID   string
}

// Identifiable is implemented by widgets that provide an identity.
// If WidgetID() returns a non-empty string, that ID takes precedence
// over the position-based AutoID for focus management and hit testing.
// The ID should be unique among siblings and persist across rebuilds.
type Identifiable interface {
	WidgetID() string
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
	OnClick(event MouseEvent)
}

// MouseDownHandler is implemented by widgets that respond to mouse button presses.
type MouseDownHandler interface {
	OnMouseDown(event MouseEvent)
}

// MouseUpHandler is implemented by widgets that respond to mouse button releases.
type MouseUpHandler interface {
	OnMouseUp(event MouseEvent)
}

// MouseMoveHandler is implemented by widgets that respond to mouse movement during drag.
type MouseMoveHandler interface {
	OnMouseMove(event MouseEvent)
}

// Hoverable is implemented by widgets that respond to mouse hover.
type Hoverable interface {
	OnHover(hovered bool)
}

// KeyCapturer is implemented by widgets that capture certain key events,
// preventing them from bubbling to ancestors. When a KeyCapturer has focus,
// ancestor keybinds are filtered based on CapturesKey() - only keybinds
// for keys that are NOT captured will be shown in the KeybindBar.
//
// For example, a text input captures printable characters (typing "q" inserts
// text rather than triggering a "quit" keybind), but allows "escape" or
// "ctrl+j" to bubble up.
type KeyCapturer interface {
	// CapturesKey returns true if this widget captures the given key,
	// preventing it from bubbling to ancestors.
	CapturesKey(key string) bool
}

// Blurrable is implemented by widgets that want to be notified when they
// lose keyboard focus. The framework calls OnBlur() when focus moves away
// from the widget.
type Blurrable interface {
	OnBlur()
}

// FocusTrapper is implemented by widgets that trap focus within their subtree.
// Tab/Shift+Tab cycling is constrained to focusables within the innermost active trap.
type FocusTrapper interface {
	TrapsFocus() bool
}

// FocusableEntry pairs a focusable widget with its identity and ancestor chain.
type FocusableEntry struct {
	ID        string
	Focusable Focusable
	// Ancestors is the chain of widgets from root to this widget's parent
	// that implement KeyHandler or KeybindProvider.
	// Used for bubbling key events up the tree.
	Ancestors []Widget
	// TrapID is the ID of the innermost FocusTrapper ancestor, or "" if none.
	// Used to constrain Tab/Shift+Tab cycling within a trap scope.
	TrapID string
	// ModalID is the ID of the enclosing modal float, or "" if none.
	// Used to keep focus within the topmost modal when modals are open.
	ModalID string
}

// FocusManager tracks the currently focused widget and handles navigation.
type FocusManager struct {
	// focusables is the ordered list of focusable widgets (in tab order)
	focusables []FocusableEntry

	// focusedID is the ID of the currently focused widget ("" if none)
	focusedID string

	// savedFocusStack stores focused widget IDs saved before entering modals.
	// Push when a modal opens, pop when it closes, to restore prior focus.
	savedFocusStack []string
	// activeModalID tracks the topmost modal ID, if any.
	activeModalID string

	// rootWidget is the root widget of the application, used to include
	// root-level keybinds in ActiveKeybinds() when nothing is focused
	rootWidget Widget
}

// NewFocusManager creates a new focus manager.
func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

// SetRootWidget sets the root widget for including root-level keybinds.
func (fm *FocusManager) SetRootWidget(root Widget) {
	fm.rootWidget = root
}

// SetActiveModalID updates the topmost modal ID for focus scoping.
// Pass "" when no modal is active.
func (fm *FocusManager) SetActiveModalID(id string) {
	fm.activeModalID = id
}

// SetFocusables updates the list of focusable widgets.
// Called after each render to update the tab order.
func (fm *FocusManager) SetFocusables(focusables []FocusableEntry) {
	fm.focusables = focusables

	// Log collected focusables
	if len(focusables) > 0 {
		Log("SetFocusables: %d focusable widgets collected", len(focusables))
		for i, entry := range focusables {
			Log("  [%d] ID=%q Type=%T", i, entry.ID, entry.Focusable)
		}
	}

	// If we have a focused ID, verify it still exists
	if fm.focusedID != "" {
		found := false
		for _, entry := range focusables {
			if entry.ID == fm.focusedID {
				found = true
				break
			}
		}
		if !found {
			Log("SetFocusables: focused ID %q no longer exists, clearing focus", fm.focusedID)
			fm.focusedID = ""
		}
	}

	// If no focus and there are focusables, focus the first one
	if fm.focusedID == "" && len(focusables) > 0 {
		fm.focusedID = focusables[0].ID
		Log("SetFocusables: auto-focusing first widget %q", fm.focusedID)
	}
}

// focusedIndex returns the index of the focused widget (-1 if none).
func (fm *FocusManager) focusedIndex() int {
	for i, entry := range fm.focusables {
		if entry.ID == fm.focusedID {
			return i
		}
	}
	return -1
}

// Focused returns the currently focused widget, or nil if none.
func (fm *FocusManager) Focused() Focusable {
	for _, entry := range fm.focusables {
		if entry.ID == fm.focusedID {
			return entry.Focusable
		}
	}
	return nil
}

// FocusedID returns the ID of the focused widget ("" if none).
func (fm *FocusManager) FocusedID() string {
	return fm.focusedID
}

// ActiveKeybinds returns all declarative keybindings currently active
// based on the focused widget and its ancestors, plus root widget keybinds.
// Keybindings are returned in order from focused widget to root,
// matching the order they would be checked when handling key events.
//
// If the focused widget implements KeyCapturer, ancestor keybinds are filtered
// to exclude keys that the focused widget captures (since those keys won't
// bubble up to trigger the ancestor keybinds).
func (fm *FocusManager) ActiveKeybinds() []Keybind {
	// Find the focused entry
	var focusedEntry *FocusableEntry
	for i := range fm.focusables {
		if fm.focusables[i].ID == fm.focusedID {
			focusedEntry = &fm.focusables[i]
			break
		}
	}

	var keybinds []Keybind

	// Check if the focused widget captures certain keys
	var capturer KeyCapturer
	if focusedEntry != nil {
		capturer, _ = focusedEntry.Focusable.(KeyCapturer)
	}

	if focusedEntry != nil {
		// Collect from focused widget first (unfiltered)
		if provider, ok := focusedEntry.Focusable.(KeybindProvider); ok {
			keybinds = append(keybinds, provider.Keybinds()...)
		}

		// Collect from ancestors (innermost to outermost/root), filtering if needed
		for i := len(focusedEntry.Ancestors) - 1; i >= 0; i-- {
			if provider, ok := focusedEntry.Ancestors[i].(KeybindProvider); ok {
				keybinds = appendFilteredKeybinds(keybinds, provider.Keybinds(), capturer)
			}
		}
	}

	// Include root widget keybinds (they're the fallback), filtering if needed
	if fm.rootWidget != nil {
		if provider, ok := fm.rootWidget.(KeybindProvider); ok {
			keybinds = appendFilteredKeybinds(keybinds, provider.Keybinds(), capturer)
		}
	}

	return keybinds
}

// appendFilteredKeybinds appends keybinds to the slice, filtering out any
// that would be captured by the given KeyCapturer (if non-nil).
func appendFilteredKeybinds(dest []Keybind, src []Keybind, capturer KeyCapturer) []Keybind {
	if capturer == nil {
		return append(dest, src...)
	}
	for _, kb := range src {
		if !capturer.CapturesKey(kb.Key) {
			dest = append(dest, kb)
		}
	}
	return dest
}

// notifyBlur calls OnBlur on the currently focused widget if it implements Blurrable.
func (fm *FocusManager) notifyBlur() {
	focused := fm.Focused()
	if focused == nil {
		return
	}
	if blurrable, ok := focused.(Blurrable); ok {
		blurrable.OnBlur()
	}
}

// FocusByID sets focus to the widget with the given ID.
func (fm *FocusManager) FocusByID(id string) {
	if id == fm.focusedID {
		return // No change
	}
	for _, entry := range fm.focusables {
		if entry.ID == id && entry.Focusable.IsFocusable() {
			fm.notifyBlur()
			fm.focusedID = id
			return
		}
	}
}

// IsInModalTrap returns true if the currently focused widget is inside a modal.
// Used to avoid re-saving focus on every render when a modal is already open.
func (fm *FocusManager) IsInModalTrap() bool {
	return fm.FocusedModalID() != ""
}

// SaveFocus pushes the current focused ID onto the saved focus stack.
// Call this before moving focus into a modal so the prior focus can be
// restored when the modal closes via RestoreFocus.
func (fm *FocusManager) SaveFocus() {
	fm.savedFocusStack = append(fm.savedFocusStack, fm.focusedID)
}

// RestoreFocus pops the most recently saved focus ID from the stack
// and moves focus back to it. Returns the restored ID, or "" if the
// stack was empty.
func (fm *FocusManager) RestoreFocus() string {
	if len(fm.savedFocusStack) == 0 {
		return ""
	}
	id := fm.savedFocusStack[len(fm.savedFocusStack)-1]
	fm.savedFocusStack = fm.savedFocusStack[:len(fm.savedFocusStack)-1]
	if id != "" {
		fm.FocusByID(id)
	}
	return id
}

// activeTrapID returns the TrapID of the currently focused widget.
// Returns "" if no widget is focused or the focused widget is not in a trap.
func (fm *FocusManager) activeTrapID() string {
	for _, entry := range fm.focusables {
		if entry.ID == fm.focusedID {
			return entry.TrapID
		}
	}
	return ""
}

// FocusedModalID returns the modal ID for the currently focused widget.
// Returns "" if no widget is focused or the focused widget is not in a modal.
func (fm *FocusManager) FocusedModalID() string {
	for _, entry := range fm.focusables {
		if entry.ID == fm.focusedID {
			return entry.ModalID
		}
	}
	return ""
}

// focusablesInScope returns the focusable entries that are in the given trap scope.
// If trapID is "", returns all focusables (no trap active).
func (fm *FocusManager) focusablesInScope(trapID string) []FocusableEntry {
	if trapID == "" {
		if fm.activeModalID == "" {
			return fm.focusables
		}
		var modalScoped []FocusableEntry
		for _, entry := range fm.focusables {
			if entry.ModalID == fm.activeModalID {
				modalScoped = append(modalScoped, entry)
			}
		}
		return modalScoped
	}
	var scoped []FocusableEntry
	for _, entry := range fm.focusables {
		if entry.TrapID == trapID {
			scoped = append(scoped, entry)
		}
	}
	return scoped
}

// FocusNext moves focus to the next focusable widget (Tab).
// If the focused widget is within a focus trap, cycling is constrained to that trap.
func (fm *FocusManager) FocusNext() {
	if len(fm.focusables) == 0 {
		Log("FocusNext: no focusables available")
		return
	}

	oldID := fm.focusedID
	trapID := fm.activeTrapID()
	candidates := fm.focusablesInScope(trapID)
	if len(candidates) == 0 {
		Log("FocusNext: no focusables in scope (trapID=%q)", trapID)
		return
	}

	// Find the current position within the scoped candidates
	startIndex := -1
	for i, entry := range candidates {
		if entry.ID == fm.focusedID {
			startIndex = i
			break
		}
	}

	// Find next focusable widget within candidates
	for i := 0; i < len(candidates); i++ {
		nextIndex := (startIndex + 1 + i) % len(candidates)
		if candidates[nextIndex].Focusable.IsFocusable() {
			newID := candidates[nextIndex].ID
			if newID != oldID {
				fm.notifyBlur()
			}
			fm.focusedID = newID
			Log("FocusNext: %q -> %q (trapID=%q)", oldID, fm.focusedID, trapID)
			return
		}
	}
	Log("FocusNext: no focusable widget found in scope (trapID=%q)", trapID)
}

// FocusPrevious moves focus to the previous focusable widget (Shift+Tab).
// If the focused widget is within a focus trap, cycling is constrained to that trap.
func (fm *FocusManager) FocusPrevious() {
	if len(fm.focusables) == 0 {
		Log("FocusPrevious: no focusables available")
		return
	}

	oldID := fm.focusedID
	trapID := fm.activeTrapID()
	candidates := fm.focusablesInScope(trapID)
	if len(candidates) == 0 {
		Log("FocusPrevious: no focusables in scope (trapID=%q)", trapID)
		return
	}

	// Find the current position within the scoped candidates
	startIndex := 0
	for i, entry := range candidates {
		if entry.ID == fm.focusedID {
			startIndex = i
			break
		}
	}

	// Find previous focusable widget within candidates
	for i := 0; i < len(candidates); i++ {
		prevIndex := (startIndex - 1 - i + len(candidates)) % len(candidates)
		if candidates[prevIndex].Focusable.IsFocusable() {
			newID := candidates[prevIndex].ID
			if newID != oldID {
				fm.notifyBlur()
			}
			fm.focusedID = newID
			Log("FocusPrevious: %q -> %q (trapID=%q)", oldID, fm.focusedID, trapID)
			return
		}
	}
	Log("FocusPrevious: no focusable widget found in scope (trapID=%q)", trapID)
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
		if fm.focusables[i].ID == fm.focusedID {
			focusedEntry = &fm.focusables[i]
			break
		}
	}

	if focusedEntry == nil {
		Log("HandleKey: no focused widget, key unhandled")
		return false
	}

	Log("HandleKey: focused widget is %q (%T), %d ancestors in chain",
		focusedEntry.ID, focusedEntry.Focusable, len(focusedEntry.Ancestors))

	// First, try the focused widget itself
	// Check declarative keybindings first
	if provider, ok := focusedEntry.Focusable.(KeybindProvider); ok {
		Log("HandleKey: checking keybindings on focused widget %q", focusedEntry.ID)
		if matchKeybind(event, provider.Keybinds()) {
			Log("HandleKey: key %q handled by keybinding on focused widget %q", event.Key(), focusedEntry.ID)
			return true
		}
	}
	// Then try imperative OnKey
	Log("HandleKey: calling OnKey on focused widget %q", focusedEntry.ID)
	if focusedEntry.Focusable.OnKey(event) {
		Log("HandleKey: key %q handled by OnKey on focused widget %q", event.Key(), focusedEntry.ID)
		return true
	}
	Log("HandleKey: focused widget %q did not handle key, bubbling up", focusedEntry.ID)

	// Bubble up through ancestors (from innermost to outermost/root)
	for i := len(focusedEntry.Ancestors) - 1; i >= 0; i-- {
		ancestor := focusedEntry.Ancestors[i]
		ancestorID := ""
		if identifiable, ok := ancestor.(Identifiable); ok {
			ancestorID = identifiable.WidgetID()
		}

		// Check declarative keybindings first
		if provider, ok := ancestor.(KeybindProvider); ok {
			Log("HandleKey: checking keybindings on ancestor %q (%T)", ancestorID, ancestor)
			if matchKeybind(event, provider.Keybinds()) {
				Log("HandleKey: key %q handled by keybinding on ancestor %q", event.Key(), ancestorID)
				return true
			}
		}

		// Then try imperative OnKey
		if handler, ok := ancestor.(KeyHandler); ok {
			Log("HandleKey: calling OnKey on ancestor %q (%T)", ancestorID, ancestor)
			if handler.OnKey(event) {
				Log("HandleKey: key %q handled by OnKey on ancestor %q", event.Key(), ancestorID)
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
	// ancestorStack tracks widgets that implement KeyHandler or KeybindProvider
	// from root to current position
	ancestorStack []Widget
	// trapStack tracks the IDs of enclosing FocusTrapper widgets.
	// The last element is the innermost active trap.
	trapStack []string
	// modalStack tracks the IDs of enclosing modal floats.
	// The last element is the topmost modal scope.
	modalStack []string
}

// NewFocusCollector creates a new focus collector.
func NewFocusCollector() *FocusCollector {
	return &FocusCollector{}
}

// PushTrap pushes a focus trap scope onto the stack.
// Called when entering a widget that implements FocusTrapper with TrapsFocus() == true.
func (fc *FocusCollector) PushTrap(id string) {
	fc.trapStack = append(fc.trapStack, id)
}

// PopTrap removes the innermost focus trap scope from the stack.
func (fc *FocusCollector) PopTrap() {
	if len(fc.trapStack) > 0 {
		fc.trapStack = fc.trapStack[:len(fc.trapStack)-1]
	}
}

// PushModal pushes a modal scope onto the stack.
func (fc *FocusCollector) PushModal(id string) {
	fc.modalStack = append(fc.modalStack, id)
}

// PopModal removes the topmost modal scope from the stack.
func (fc *FocusCollector) PopModal() {
	if len(fc.modalStack) > 0 {
		fc.modalStack = fc.modalStack[:len(fc.modalStack)-1]
	}
}

// CurrentModalID returns the ID of the topmost modal scope, or "" if none.
func (fc *FocusCollector) CurrentModalID() string {
	if len(fc.modalStack) > 0 {
		return fc.modalStack[len(fc.modalStack)-1]
	}
	return ""
}

// CurrentTrapID returns the ID of the innermost active focus trap, or "" if none.
func (fc *FocusCollector) CurrentTrapID() string {
	if len(fc.trapStack) > 0 {
		return fc.trapStack[len(fc.trapStack)-1]
	}
	return ""
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

// Collect adds a focusable widget to the collection.
// The autoID is the position-based auto-generated ID from BuildContext.
// If the widget implements Identifiable with a non-empty ID, that is used instead.
// Disabled widgets (ctx.IsDisabled() == true) are skipped and cannot receive focus.
func (fc *FocusCollector) Collect(widget Widget, autoID string, ctx BuildContext) {
	// Skip disabled widgets - they cannot receive focus
	if ctx.IsDisabled() {
		return
	}

	focusable, ok := widget.(Focusable)
	if !ok || !focusable.IsFocusable() {
		return
	}

	// Get the widget's ID, falling back to auto-ID if not provided or empty
	id := autoID
	if identifiable, ok := widget.(Identifiable); ok && identifiable.WidgetID() != "" {
		id = identifiable.WidgetID()
	}

	// Copy the current ancestor chain for this focusable
	var ancestors []Widget
	if len(fc.ancestorStack) > 0 {
		ancestors = make([]Widget, len(fc.ancestorStack))
		copy(ancestors, fc.ancestorStack)
	}

	fc.focusables = append(fc.focusables, FocusableEntry{
		ID:        id,
		Focusable: focusable,
		Ancestors: ancestors,
		TrapID:    fc.CurrentTrapID(),
		ModalID:   fc.CurrentModalID(),
	})
}

// Focusables returns all collected focusable entries.
func (fc *FocusCollector) Focusables() []FocusableEntry {
	return fc.focusables
}

// Reset clears the collected focusables for a new render pass.
func (fc *FocusCollector) Reset() {
	fc.focusables = fc.focusables[:0]
	fc.ancestorStack = fc.ancestorStack[:0]
	fc.trapStack = fc.trapStack[:0]
	fc.modalStack = fc.modalStack[:0]
}

// Len returns the number of focusables collected so far.
func (fc *FocusCollector) Len() int {
	return len(fc.focusables)
}

// FirstIDAfter returns the ID of the first focusable collected after the given index.
// Returns empty string if no focusables were collected after that index.
// This is useful for finding the first focusable in a subtree that was just built.
func (fc *FocusCollector) FirstIDAfter(index int) string {
	if index < len(fc.focusables) {
		return fc.focusables[index].ID
	}
	return ""
}
