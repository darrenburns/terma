package terma

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

type signalReadContext struct {
	node  *widgetNode
	phase dependencyMask
}

var currentSignalRead signalReadContext
var currentSignalReadMu sync.Mutex

func withSignalRead[T any](node *widgetNode, phase dependencyMask, fn func() T) T {
	currentSignalReadMu.Lock()
	prev := currentSignalRead
	currentSignalRead = signalReadContext{node: node, phase: phase}
	currentSignalReadMu.Unlock()
	defer func() {
		currentSignalReadMu.Lock()
		currentSignalRead = prev
		currentSignalReadMu.Unlock()
	}()
	return fn()
}

func currentReadSubscription() signalReadContext {
	currentSignalReadMu.Lock()
	defer currentSignalReadMu.Unlock()
	return currentSignalRead
}

var debugRenderCauseEnabled atomic.Bool
var lastRenderCause atomic.Value

// signalCore holds the internal state for Signal.
// All fields are protected by mu for thread-safe access.
type signalCore[T comparable] struct {
	mu        sync.Mutex
	value     T
	listeners map[*widgetNode]dependencyMask
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
			listeners: make(map[*widgetNode]dependencyMask),
		},
	}
}

// Get returns the current value. If called during a tracked render phase,
// the widget is automatically subscribed to future changes for that phase.
// Thread-safe: can be called from any goroutine.
func (s Signal[T]) Get() T {
	read := currentReadSubscription()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()

	if read.node != nil && read.phase != readPhaseNone {
		s.core.listeners[read.node] |= read.phase
		read.node.trackDependency(s.core, read.phase)
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

	// Copy listeners to avoid holding lock during markDirty.
	type listenerEntry struct {
		node *widgetNode
		mask dependencyMask
	}
	listeners := make([]listenerEntry, 0, len(s.core.listeners))
	for listener, mask := range s.core.listeners {
		listeners = append(listeners, listenerEntry{node: listener, mask: mask})
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.node.markDirtyMask(listener.mask)
	}
	recordRenderCause("Signal.Set", value, s.core, 2)
	scheduleRender()
}

// setSilently updates the signal value without notifying listeners or scheduling
// a render. It is only for framework-owned derived state that is synchronized
// during an in-flight render/layout pass.
func (s Signal[T]) setSilently(value T) {
	if s.core == nil {
		return
	}
	s.core.mu.Lock()
	s.core.value = value
	s.core.mu.Unlock()
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

	type listenerEntry struct {
		node *widgetNode
		mask dependencyMask
	}
	listeners := make([]listenerEntry, 0, len(s.core.listeners))
	for listener, mask := range s.core.listeners {
		listeners = append(listeners, listenerEntry{node: listener, mask: mask})
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.node.markDirtyMask(listener.mask)
	}
	recordRenderCause("Signal.Update", newValue, s.core, 2)
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

func (s *signalCore[T]) removeListener(node *widgetNode, mask dependencyMask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.listeners[node]
	next := current &^ mask
	if next == 0 {
		delete(s.listeners, node)
		return
	}
	s.listeners[node] = next
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
	listeners map[*widgetNode]dependencyMask
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
			listeners: make(map[*widgetNode]dependencyMask),
		},
	}
}

// Get returns the current value. If called during a tracked render phase,
// the widget is automatically subscribed to future changes for that phase.
// Thread-safe: can be called from any goroutine.
func (s AnySignal[T]) Get() T {
	read := currentReadSubscription()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()

	if read.node != nil && read.phase != readPhaseNone {
		s.core.listeners[read.node] |= read.phase
		read.node.trackDependency(s.core, read.phase)
	}
	return s.core.value
}

// Set updates the value, notifies all subscribers, and schedules a re-render.
// Thread-safe: can be called from any goroutine.
func (s AnySignal[T]) Set(value T) {
	s.core.mu.Lock()
	s.core.value = value

	type listenerEntry struct {
		node *widgetNode
		mask dependencyMask
	}
	listeners := make([]listenerEntry, 0, len(s.core.listeners))
	for listener, mask := range s.core.listeners {
		listeners = append(listeners, listenerEntry{node: listener, mask: mask})
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.node.markDirtyMask(listener.mask)
	}
	recordRenderCause("AnySignal.Set", value, s.core, 2)
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

	type listenerEntry struct {
		node *widgetNode
		mask dependencyMask
	}
	listeners := make([]listenerEntry, 0, len(s.core.listeners))
	for listener, mask := range s.core.listeners {
		listeners = append(listeners, listenerEntry{node: listener, mask: mask})
	}
	s.core.mu.Unlock()

	for _, listener := range listeners {
		listener.node.markDirtyMask(listener.mask)
	}
	recordRenderCause("AnySignal.Update", s.core.value, s.core, 2)
	scheduleRender()
}

// IsValid returns true if the signal was properly initialized.
// An uninitialized AnySignal (zero value) returns false.
func (s AnySignal[T]) IsValid() bool {
	return s.core != nil
}

func (s *anySignalCore[T]) removeListener(node *widgetNode, mask dependencyMask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.listeners[node]
	next := current &^ mask
	if next == 0 {
		delete(s.listeners, node)
		return
	}
	s.listeners[node] = next
}

// EnableDebugRenderCause turns on tracking of the most recent render cause.
func EnableDebugRenderCause() {
	debugRenderCauseEnabled.Store(true)
}

// LastRenderCause returns a debug string describing the most recent render cause.
func LastRenderCause() string {
	value := lastRenderCause.Load()
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func recordRenderCause(kind string, value any, core any, skip int) {
	if !debugRenderCauseEnabled.Load() {
		return
	}

	pc, file, line, ok := runtime.Caller(skip)
	location := "unknown"
	if ok {
		location = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	funcName := ""
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = fn.Name()
	}

	coreInfo := ""
	if core != nil {
		coreInfo = fmt.Sprintf(" core=%p", core)
	}

	cause := fmt.Sprintf("%s %T%s", kind, value, coreInfo)
	if funcName != "" {
		cause = fmt.Sprintf("%s via %s (%s)", cause, funcName, location)
	} else if location != "unknown" {
		cause = fmt.Sprintf("%s at %s", cause, location)
	}

	lastRenderCause.Store(cause)
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
