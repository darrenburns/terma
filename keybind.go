package terma

// Keybind represents a single declarative keybinding.
// It associates a key pattern with a display name and action callback.
type Keybind struct {
	// Key is the key pattern to match, e.g., "ctrl+s", "enter", "shift+tab"
	Key string
	// Name is a short display name for the binding, e.g., "Save", "Submit"
	Name string
	// Action is the callback to execute when the keybinding is triggered
	Action func()
	// Hidden prevents this keybind from appearing in KeybindBar.
	// Use for internal bindings that shouldn't be displayed to users.
	Hidden bool
}

// KeybindProvider is implemented by widgets that declare keybindings.
// Widgets implementing this interface can define their keybindings declaratively,
// allowing the framework to query and display them (e.g., in a footer).
type KeybindProvider interface {
	Keybinds() []Keybind
}

// matchKeybind checks if the event matches any keybind and executes its action.
// Returns true if a keybind was matched and executed.
func matchKeybind(event KeyEvent, keybinds []Keybind) bool {
	eventKey := event.Key()

	for _, kb := range keybinds {
		matches := event.MatchString(kb.Key)

		// WORKAROUND: ultraviolet's MatchString has a bug with "+" character.
		// It uses "+" as a delimiter for modifiers (e.g., "ctrl+a"), so a literal
		// "+" gets split into empty strings and fails to match.
		// Fall back to direct string comparison when MatchString fails.
		if !matches && eventKey == kb.Key {
			matches = true
		}

		if matches {
			if kb.Action != nil {
				kb.Action()
			}
			return true
		}
	}
	return false
}

