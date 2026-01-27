package terma

import "sync"

// currentBuildingNode tracks which widget node is currently being built.
// When a Signal.Get() is called during Build(), the node is subscribed.
// Protected by currentBuildMu for thread-safe access.
var currentBuildingNode *widgetNode
var currentBuildMu sync.Mutex

// signalCore holds the internal state for Signal.
// All fields are protected by mu for thread-safe access.
type signalCore[T comparable] struct {
	mu        sync.Mutex
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
// Thread-safe: can be called from any goroutine.
func (s Signal[T]) Get() T {
	// Read current building node atomically
	currentBuildMu.Lock()
	node := currentBuildingNode
	currentBuildMu.Unlock()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()

	if node != nil {
		s.core.listeners[node] = struct{}{}
	}
	return s.core.value
}

// Set updates the value. If the value changed, all subscribed widgets
// are marked dirty for rebuild and a re-render is scheduled.
// Thread-safe: can be called from any goroutine.
func (s Signal[T]) Set(value T) {
	s.core.mu.Lock()
	if s.core.value == value {
		s.core.mu.Unlock()
		return
	}
	s.core.value = value

	// Copy listeners to avoid holding lock during markDirty
	listeners := make([]*widgetNode, 0, len(s.core.listeners))
	for listener := range s.core.listeners {
		listeners = append(listeners, listener)
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// Peek returns the current value without subscribing.
// Thread-safe: can be called from any goroutine.
func (s Signal[T]) Peek() T {
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	return s.core.value
}

// Update applies a function to the current value and sets the result.
// Thread-safe: can be called from any goroutine. The function is called
// while holding the lock, so it should be fast and not call other Signal methods.
func (s Signal[T]) Update(fn func(T) T) {
	s.core.mu.Lock()
	oldValue := s.core.value
	newValue := fn(oldValue)
	if newValue == oldValue {
		s.core.mu.Unlock()
		return
	}
	s.core.value = newValue

	// Copy listeners to avoid holding lock during markDirty
	listeners := make([]*widgetNode, 0, len(s.core.listeners))
	for listener := range s.core.listeners {
		listeners = append(listeners, listener)
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// unsubscribe removes a widget node from the listeners.
// Called when a widget is unmounted.
// Thread-safe.
func (s Signal[T]) unsubscribe(node *widgetNode) {
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	delete(s.core.listeners, node)
}

// IsValid returns true if the signal was properly initialized.
// An uninitialized Signal (zero value) returns false.
func (s Signal[T]) IsValid() bool {
	return s.core != nil
}

// anySignalCore holds the internal state for AnySignal.
// All fields are protected by mu for thread-safe access.
type anySignalCore[T any] struct {
	mu        sync.Mutex
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
// Thread-safe: can be called from any goroutine.
func (s AnySignal[T]) Get() T {
	// Read current building node atomically
	currentBuildMu.Lock()
	node := currentBuildingNode
	currentBuildMu.Unlock()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()

	if node != nil {
		s.core.listeners[node] = struct{}{}
	}
	return s.core.value
}

// Set updates the value, notifies all subscribers, and schedules a re-render.
// Thread-safe: can be called from any goroutine.
func (s AnySignal[T]) Set(value T) {
	s.core.mu.Lock()
	s.core.value = value

	// Copy listeners to avoid holding lock during markDirty
	listeners := make([]*widgetNode, 0, len(s.core.listeners))
	for listener := range s.core.listeners {
		listeners = append(listeners, listener)
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// Peek returns the current value without subscribing.
// Thread-safe: can be called from any goroutine.
func (s AnySignal[T]) Peek() T {
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	return s.core.value
}

// Update applies a function to the current value and sets the result.
// Thread-safe: can be called from any goroutine. The function is called
// while holding the lock, so it should be fast and not call other Signal methods.
func (s AnySignal[T]) Update(fn func(T) T) {
	s.core.mu.Lock()
	s.core.value = fn(s.core.value)

	// Copy listeners to avoid holding lock during markDirty
	listeners := make([]*widgetNode, 0, len(s.core.listeners))
	for listener := range s.core.listeners {
		listeners = append(listeners, listener)
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.markDirty()
	}
	scheduleRender()
}

// unsubscribe removes a widget node from the listeners.
// Called when a widget is unmounted.
// Thread-safe.
func (s AnySignal[T]) unsubscribe(node *widgetNode) {
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	delete(s.core.listeners, node)
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
