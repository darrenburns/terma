package terma

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Interpolator defines how to interpolate between two values of type T.
// The parameter t is in range [0, 1] where 0 returns from and 1 returns to.
type Interpolator[T any] func(from, to T, t float64) T

// interpolatorRegistry holds registered interpolators by type name.
var interpolatorRegistry = struct {
	sync.RWMutex
	entries map[string]any
}{
	entries: make(map[string]any),
}

// RegisterInterpolator registers a custom interpolator for type T.
// Built-in types (float64, float32, int, int64, Color) are registered automatically.
func RegisterInterpolator[T any](interpolator Interpolator[T]) {
	interpolatorRegistry.Lock()
	defer interpolatorRegistry.Unlock()

	var zero T
	typeName := reflect.TypeOf(zero).String()
	interpolatorRegistry.entries[typeName] = interpolator
}

// getInterpolator returns the interpolator for type T, or nil if none registered.
func getInterpolator[T any]() Interpolator[T] {
	interpolatorRegistry.RLock()
	defer interpolatorRegistry.RUnlock()

	var zero T
	typeName := reflect.TypeOf(zero).String()

	if interp, ok := interpolatorRegistry.entries[typeName]; ok {
		return interp.(Interpolator[T])
	}
	return nil
}

// Built-in interpolators registered at init.
func init() {
	RegisterInterpolator(interpolateFloat64)
	RegisterInterpolator(interpolateFloat32)
	RegisterInterpolator(interpolateInt)
	RegisterInterpolator(interpolateInt64)
	RegisterInterpolator(interpolateColor)
}

func interpolateFloat64(from, to float64, t float64) float64 {
	return from + (to-from)*t
}

func interpolateFloat32(from, to float32, t float64) float32 {
	return from + float32(float64(to-from)*t)
}

func interpolateInt(from, to int, t float64) int {
	return from + int(float64(to-from)*t)
}

func interpolateInt64(from, to int64, t float64) int64 {
	return from + int64(float64(to-from)*t)
}

func interpolateColor(from, to Color, t float64) Color {
	return from.Blend(to, t)
}

// AnimationState represents the current state of an animation.
type AnimationState int

const (
	AnimationPending AnimationState = iota
	AnimationRunning
	AnimationPaused
	AnimationCompleted
)

// Animation animates a value of type T over time.
// Create with NewAnimation and start with Start().
// Read the current value with Value().Get() during Build() for reactive updates.
type Animation[T any] struct {
	mu sync.Mutex

	// Configuration
	from         T
	to           T
	duration     time.Duration
	easing       EasingFunc
	interpolator Interpolator[T]
	delay        time.Duration

	// State
	state        AnimationState
	elapsed      time.Duration
	delayElapsed time.Duration
	current      T

	// Output signal (updated on each tick)
	signal AnySignal[T]

	// Callbacks
	onComplete func()
	onUpdate   func(T)

	// Controller handle for cleanup
	handle *animationHandle
}

// AnimationConfig holds configuration for creating an Animation.
type AnimationConfig[T any] struct {
	From         T
	To           T
	Duration     time.Duration
	Easing       EasingFunc      // Optional, defaults to EaseLinear
	Interpolator Interpolator[T] // Optional, uses registry if nil
	Delay        time.Duration   // Optional delay before animation starts
	OnComplete   func()          // Optional callback when animation completes
	OnUpdate     func(T)         // Optional callback on each value update
}

// NewAnimation creates a new animation with the given configuration.
// Panics if no interpolator is registered for type T and none is provided in config.
func NewAnimation[T any](config AnimationConfig[T]) *Animation[T] {
	interp := config.Interpolator
	if interp == nil {
		interp = getInterpolator[T]()
	}
	if interp == nil {
		var zero T
		panic(fmt.Sprintf("no interpolator registered for type %T; use RegisterInterpolator or provide one in AnimationConfig", zero))
	}

	easing := config.Easing
	if easing == nil {
		easing = EaseLinear
	}

	return &Animation[T]{
		from:         config.From,
		to:           config.To,
		duration:     config.Duration,
		easing:       easing,
		interpolator: interp,
		delay:        config.Delay,
		state:        AnimationPending,
		current:      config.From,
		signal:       NewAnySignal(config.From),
		onComplete:   config.OnComplete,
		onUpdate:     config.OnUpdate,
	}
}

// Start begins the animation.
// The animation will register with the current AnimationController.
func (a *Animation[T]) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state == AnimationRunning {
		return
	}

	a.state = AnimationRunning
	a.elapsed = 0
	a.delayElapsed = 0

	// Register with global controller
	if currentController != nil {
		a.handle = currentController.Register(a)
	}
}

// Stop halts the animation without completing.
func (a *Animation[T]) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.handle != nil && currentController != nil {
		currentController.Unregister(a.handle)
		a.handle = nil
	}
	a.state = AnimationCompleted
}

// Pause temporarily suspends the animation.
func (a *Animation[T]) Pause() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state == AnimationRunning {
		a.state = AnimationPaused
	}
}

// Resume continues a paused animation.
func (a *Animation[T]) Resume() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state == AnimationPaused {
		a.state = AnimationRunning
	}
}

// Reset restarts the animation from the beginning.
// Does not automatically start the animation; call Start() after Reset().
func (a *Animation[T]) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.elapsed = 0
	a.delayElapsed = 0
	a.current = a.from
	a.signal.Set(a.from)
	a.state = AnimationPending
}

// Value returns a signal containing the current animated value.
// Call Value().Get() during Build() for reactive updates.
func (a *Animation[T]) Value() AnySignal[T] {
	return a.signal
}

// Get returns the current value directly without subscribing.
func (a *Animation[T]) Get() T {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.current
}

// IsRunning returns true if the animation is currently active.
func (a *Animation[T]) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state == AnimationRunning
}

// IsComplete returns true if the animation has finished.
func (a *Animation[T]) IsComplete() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state == AnimationCompleted
}

// Progress returns the current progress as a value from 0.0 to 1.0.
func (a *Animation[T]) Progress() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.duration == 0 {
		return 1.0
	}
	progress := float64(a.elapsed) / float64(a.duration)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// Advance implements the Animator interface.
// Called by AnimationController on each tick.
func (a *Animation[T]) Advance(dt time.Duration) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state != AnimationRunning {
		return a.state != AnimationCompleted
	}

	// Handle delay
	if a.delayElapsed < a.delay {
		a.delayElapsed += dt
		if a.delayElapsed < a.delay {
			return true // Still in delay
		}
		// Delay finished, adjust dt for overflow
		dt = a.delayElapsed - a.delay
	}

	// Advance animation
	a.elapsed += dt

	// Calculate progress (0.0 to 1.0)
	progress := float64(a.elapsed) / float64(a.duration)
	if progress > 1.0 {
		progress = 1.0
	}

	// Apply easing
	easedProgress := a.easing(progress)

	// Interpolate value
	a.current = a.interpolator(a.from, a.to, easedProgress)
	a.signal.Set(a.current)

	// Notify update callback
	if a.onUpdate != nil {
		a.onUpdate(a.current)
	}

	// Check completion
	if progress >= 1.0 {
		a.state = AnimationCompleted
		if a.onComplete != nil {
			a.onComplete()
		}
		return false // Remove from controller
	}

	return true // Keep running
}
