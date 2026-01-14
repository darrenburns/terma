package terma

// currentBuildingNode tracks which widget node is currently being built.
// When a Signal.Get() is called during Build(), the node is subscribed.
var currentBuildingNode *widgetNode

// signalCore holds the internal state for Signal.
type signalCore[T comparable] struct {
	value     T
	listeners map[*widgetNode]struct{}
}

// Signal holds reactive state that automatically tracks dependencies.
// When the value changes, all subscribed widget nodes are marked dirty.
// Signal can be stored by value in structs; copies share the same underlying state.
type Signal[T comparable] struct {
	core *signalCore[T]
}

// NewSignal creates a new signal with the given initial value.
func NewSignal[T comparable](initial T) Signal[T] {
	return Signal[T]{
		core: &signalCore[T]{
			value:     initial,
			listeners: make(map[*widgetNode]struct{}),
		},
	}
}

// Get returns the current value. If called during a widget's Build(),
// the widget is automatically subscribed to future changes.
func (s Signal[T]) Get() T {
	if currentBuildingNode != nil {
		s.core.listeners[currentBuildingNode] = struct{}{}
	}
	return s.core.value
}

// Set updates the value. If the value changed, all subscribed widgets
// are marked dirty for rebuild and a re-render is scheduled.
func (s Signal[T]) Set(value T) {
	if s.core.value == value {
		return
	}
	s.core.value = value
	for listener := range s.core.listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// Peek returns the current value without subscribing.
func (s Signal[T]) Peek() T {
	return s.core.value
}

// Update applies a function to the current value and sets the result.
func (s Signal[T]) Update(fn func(T) T) {
	s.Set(fn(s.core.value))
}

// unsubscribe removes a widget node from the listeners.
// Called when a widget is unmounted.
func (s Signal[T]) unsubscribe(node *widgetNode) {
	delete(s.core.listeners, node)
}

// IsValid returns true if the signal was properly initialized.
// An uninitialized Signal (zero value) returns false.
func (s Signal[T]) IsValid() bool {
	return s.core != nil
}

// anySignalCore holds the internal state for AnySignal.
type anySignalCore[T any] struct {
	value     T
	listeners map[*widgetNode]struct{}
}

// AnySignal holds reactive state for non-comparable types (like interfaces).
// Unlike Signal, it always notifies on Set() since equality cannot be checked.
// AnySignal can be stored by value in structs; copies share the same underlying state.
type AnySignal[T any] struct {
	core *anySignalCore[T]
}

// NewAnySignal creates a new signal for non-comparable types.
func NewAnySignal[T any](initial T) AnySignal[T] {
	return AnySignal[T]{
		core: &anySignalCore[T]{
			value:     initial,
			listeners: make(map[*widgetNode]struct{}),
		},
	}
}

// Get returns the current value. If called during a widget's Build(),
// the widget is automatically subscribed to future changes.
func (s AnySignal[T]) Get() T {
	if currentBuildingNode != nil {
		s.core.listeners[currentBuildingNode] = struct{}{}
	}
	return s.core.value
}

// Set updates the value, notifies all subscribers, and schedules a re-render.
func (s AnySignal[T]) Set(value T) {
	s.core.value = value
	for listener := range s.core.listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// Peek returns the current value without subscribing.
func (s AnySignal[T]) Peek() T {
	return s.core.value
}

// Update applies a function to the current value and sets the result.
func (s AnySignal[T]) Update(fn func(T) T) {
	s.Set(fn(s.core.value))
}

// IsValid returns true if the signal was properly initialized.
// An uninitialized AnySignal (zero value) returns false.
func (s AnySignal[T]) IsValid() bool {
	return s.core != nil
}

// scheduleRender signals the app to re-render.
// Non-blocking: drops the signal if one is already pending.
func scheduleRender() {
	if renderTrigger != nil {
		select {
		case renderTrigger <- struct{}{}:
		default:
		}
	}
}
