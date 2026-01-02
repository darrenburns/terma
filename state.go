package terma

// stateRegistry holds widget state keyed by ID (explicit or auto-generated).
// State persists across rebuilds, allowing widgets to maintain internal state.
var stateRegistry = make(map[string]any)

// GetOrCreateState retrieves existing state for the given ID, or creates new state
// using the factory function if none exists.
func GetOrCreateState[T any](id string, factory func() *T) *T {
	if existing, ok := stateRegistry[id]; ok {
		return existing.(*T)
	}
	state := factory()
	stateRegistry[id] = state
	return state
}

// ClearState removes state for the given ID.
// Use this for cleanup when a widget is permanently removed.
func ClearState(id string) {
	delete(stateRegistry, id)
}

// ClearAllState removes all stored state.
// Primarily useful for testing.
func ClearAllState() {
	stateRegistry = make(map[string]any)
}