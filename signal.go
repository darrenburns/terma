package terma

// currentBuildingNode tracks which widget node is currently being built.
// When a Signal.Get() is called during Build(), the node is subscribed.
var currentBuildingNode *widgetNode

// Signal holds reactive state that automatically tracks dependencies.
// When the value changes, all subscribed widget nodes are marked dirty.
type Signal[T comparable] struct {
	value     T
	listeners map[*widgetNode]struct{}
}

// NewSignal creates a new signal with the given initial value.
func NewSignal[T comparable](initial T) *Signal[T] {
	return &Signal[T]{
		value:     initial,
		listeners: make(map[*widgetNode]struct{}),
	}
}

// Get returns the current value. If called during a widget's Build(),
// the widget is automatically subscribed to future changes.
func (s *Signal[T]) Get() T {
	if currentBuildingNode != nil {
		s.listeners[currentBuildingNode] = struct{}{}
	}
	return s.value
}

// Set updates the value. If the value changed, all subscribed widgets
// are marked dirty for rebuild.
func (s *Signal[T]) Set(value T) {
	if s.value == value {
		return
	}
	s.value = value
	for listener := range s.listeners {
		listener.markDirty()
	}
}

// Peek returns the current value without subscribing.
func (s *Signal[T]) Peek() T {
	return s.value
}

// Update applies a function to the current value and sets the result.
func (s *Signal[T]) Update(fn func(T) T) {
	s.Set(fn(s.value))
}

// unsubscribe removes a widget node from the listeners.
// Called when a widget is unmounted.
func (s *Signal[T]) unsubscribe(node *widgetNode) {
	delete(s.listeners, node)
}

// AnySignal holds reactive state for non-comparable types (like interfaces).
// Unlike Signal, it always notifies on Set() since equality cannot be checked.
type AnySignal[T any] struct {
	value     T
	listeners map[*widgetNode]struct{}
}

// NewAnySignal creates a new signal for non-comparable types.
func NewAnySignal[T any](initial T) *AnySignal[T] {
	return &AnySignal[T]{
		value:     initial,
		listeners: make(map[*widgetNode]struct{}),
	}
}

// Get returns the current value. If called during a widget's Build(),
// the widget is automatically subscribed to future changes.
func (s *AnySignal[T]) Get() T {
	if currentBuildingNode != nil {
		s.listeners[currentBuildingNode] = struct{}{}
	}
	return s.value
}

// Set updates the value and notifies all subscribers.
func (s *AnySignal[T]) Set(value T) {
	s.value = value
	for listener := range s.listeners {
		listener.markDirty()
	}
}

// Peek returns the current value without subscribing.
func (s *AnySignal[T]) Peek() T {
	return s.value
}
