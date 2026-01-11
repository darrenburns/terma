package terma

import (
	"fmt"
	"sync"
	"time"
)

// AnimatedValue wraps a value and animates transitions when Set() is called.
// It provides a signal-like interface but smoothly transitions between values.
// When Set() is called multiple times rapidly, the animation retargets from
// the current interpolated position to the new target for smooth motion.
type AnimatedValue[T comparable] struct {
	mu sync.Mutex

	// Configuration
	duration     time.Duration
	easing       EasingFunc
	interpolator Interpolator[T]

	// State
	current   T
	target    T
	signal    Signal[T]
	animation *Animation[T]
}

// AnimatedValueConfig configures an AnimatedValue.
type AnimatedValueConfig[T comparable] struct {
	Initial      T
	Duration     time.Duration
	Easing       EasingFunc      // Optional, defaults to EaseOutQuad
	Interpolator Interpolator[T] // Optional, uses registry if nil
}

// NewAnimatedValue creates a new animated value.
// Panics if no interpolator is registered for type T and none is provided in config.
func NewAnimatedValue[T comparable](config AnimatedValueConfig[T]) *AnimatedValue[T] {
	interp := config.Interpolator
	if interp == nil {
		interp = getInterpolator[T]()
	}
	if interp == nil {
		var zero T
		panic(fmt.Sprintf("no interpolator registered for type %T; use RegisterInterpolator or provide one in AnimatedValueConfig", zero))
	}

	easing := config.Easing
	if easing == nil {
		easing = EaseOutQuad
	}

	return &AnimatedValue[T]{
		duration:     config.Duration,
		easing:       easing,
		interpolator: interp,
		current:      config.Initial,
		target:       config.Initial,
		signal:       NewSignal(config.Initial),
	}
}

// Get returns the current animated value.
// During Build(), this subscribes the widget to updates.
func (av *AnimatedValue[T]) Get() T {
	return av.signal.Get()
}

// Peek returns the current value without subscribing.
func (av *AnimatedValue[T]) Peek() T {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.current
}

// Target returns the target value (where the animation is heading).
func (av *AnimatedValue[T]) Target() T {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.target
}

// Set animates to a new target value.
// If called while an animation is in progress, the animation is
// retargeted: starts from current interpolated value and animates to new target.
func (av *AnimatedValue[T]) Set(value T) {
	av.mu.Lock()
	defer av.mu.Unlock()

	// Skip if already at target
	if av.target == value {
		return
	}

	av.target = value

	// Stop any existing animation
	if av.animation != nil {
		av.animation.Stop()
	}

	// Create new animation from current position to new target
	av.animation = NewAnimation(AnimationConfig[T]{
		From:         av.current,
		To:           value,
		Duration:     av.duration,
		Easing:       av.easing,
		Interpolator: av.interpolator,
		OnUpdate: func(v T) {
			av.mu.Lock()
			av.current = v
			av.mu.Unlock()
			av.signal.Set(v)
		},
		OnComplete: func() {
			av.mu.Lock()
			av.animation = nil
			av.mu.Unlock()
		},
	})

	av.animation.Start()
}

// SetImmediate sets the value without animation.
func (av *AnimatedValue[T]) SetImmediate(value T) {
	av.mu.Lock()
	defer av.mu.Unlock()

	// Stop any running animation
	if av.animation != nil {
		av.animation.Stop()
		av.animation = nil
	}

	av.current = value
	av.target = value
	av.signal.Set(value)
}

// IsAnimating returns true if an animation is in progress.
func (av *AnimatedValue[T]) IsAnimating() bool {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.animation != nil && av.animation.IsRunning()
}

// Signal returns the underlying signal for advanced use cases.
func (av *AnimatedValue[T]) Signal() Signal[T] {
	return av.signal
}

// AnyAnimatedValue wraps a non-comparable value and animates transitions when Set() is called.
// Unlike AnimatedValue, this always animates on Set() since equality cannot be checked.
type AnyAnimatedValue[T any] struct {
	mu sync.Mutex

	// Configuration
	duration     time.Duration
	easing       EasingFunc
	interpolator Interpolator[T]

	// State
	current   T
	target    T
	signal    AnySignal[T]
	animation *Animation[T]
}

// AnyAnimatedValueConfig configures an AnyAnimatedValue.
type AnyAnimatedValueConfig[T any] struct {
	Initial      T
	Duration     time.Duration
	Easing       EasingFunc      // Optional, defaults to EaseOutQuad
	Interpolator Interpolator[T] // Optional, uses registry if nil
}

// NewAnyAnimatedValue creates a new animated value for non-comparable types.
// Panics if no interpolator is registered for type T and none is provided in config.
func NewAnyAnimatedValue[T any](config AnyAnimatedValueConfig[T]) *AnyAnimatedValue[T] {
	interp := config.Interpolator
	if interp == nil {
		interp = getInterpolator[T]()
	}
	if interp == nil {
		var zero T
		panic(fmt.Sprintf("no interpolator registered for type %T; use RegisterInterpolator or provide one in AnyAnimatedValueConfig", zero))
	}

	easing := config.Easing
	if easing == nil {
		easing = EaseOutQuad
	}

	return &AnyAnimatedValue[T]{
		duration:     config.Duration,
		easing:       easing,
		interpolator: interp,
		current:      config.Initial,
		target:       config.Initial,
		signal:       NewAnySignal(config.Initial),
	}
}

// Get returns the current animated value.
// During Build(), this subscribes the widget to updates.
func (av *AnyAnimatedValue[T]) Get() T {
	return av.signal.Get()
}

// Peek returns the current value without subscribing.
func (av *AnyAnimatedValue[T]) Peek() T {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.current
}

// Target returns the target value (where the animation is heading).
func (av *AnyAnimatedValue[T]) Target() T {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.target
}

// Set animates to a new target value.
// Unlike AnimatedValue, this always starts a new animation since equality cannot be checked.
func (av *AnyAnimatedValue[T]) Set(value T) {
	av.mu.Lock()
	defer av.mu.Unlock()

	av.target = value

	// Stop any existing animation
	if av.animation != nil {
		av.animation.Stop()
	}

	// Create new animation from current position to new target
	av.animation = NewAnimation(AnimationConfig[T]{
		From:         av.current,
		To:           value,
		Duration:     av.duration,
		Easing:       av.easing,
		Interpolator: av.interpolator,
		OnUpdate: func(v T) {
			av.mu.Lock()
			av.current = v
			av.mu.Unlock()
			av.signal.Set(v)
		},
		OnComplete: func() {
			av.mu.Lock()
			av.animation = nil
			av.mu.Unlock()
		},
	})

	av.animation.Start()
}

// SetImmediate sets the value without animation.
func (av *AnyAnimatedValue[T]) SetImmediate(value T) {
	av.mu.Lock()
	defer av.mu.Unlock()

	// Stop any running animation
	if av.animation != nil {
		av.animation.Stop()
		av.animation = nil
	}

	av.current = value
	av.target = value
	av.signal.Set(value)
}

// IsAnimating returns true if an animation is in progress.
func (av *AnyAnimatedValue[T]) IsAnimating() bool {
	av.mu.Lock()
	defer av.mu.Unlock()
	return av.animation != nil && av.animation.IsRunning()
}

// Signal returns the underlying signal for advanced use cases.
func (av *AnyAnimatedValue[T]) Signal() AnySignal[T] {
	return av.signal
}
