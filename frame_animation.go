package terma

import (
	"sync"
	"time"
)

// FrameAnimation cycles through discrete frames at a given interval.
// Useful for spinners, loading indicators, and sprite animations.
type FrameAnimation[T any] struct {
	mu sync.Mutex

	// Configuration
	frames    []T
	frameTime time.Duration
	loop      bool

	// State
	state        AnimationState
	currentFrame int
	elapsed      time.Duration
	signal       AnySignal[T]
	pendingStart bool // Start() was called before controller was ready

	// Callbacks
	onComplete func()

	// Controller handle
	handle *animationHandle
}

// FrameAnimationConfig configures a FrameAnimation.
type FrameAnimationConfig[T any] struct {
	Frames     []T           // Required: frames to cycle through
	FrameTime  time.Duration // Time per frame
	Loop       bool          // Whether to loop continuously
	OnComplete func()        // Called when non-looping animation ends
}

// NewFrameAnimation creates a new frame-based animation.
// Panics if Frames is empty.
func NewFrameAnimation[T any](config FrameAnimationConfig[T]) *FrameAnimation[T] {
	if len(config.Frames) == 0 {
		panic("FrameAnimation requires at least one frame")
	}

	return &FrameAnimation[T]{
		frames:       config.Frames,
		frameTime:    config.FrameTime,
		loop:         config.Loop,
		state:        AnimationPending,
		currentFrame: 0,
		signal:       NewAnySignal(config.Frames[0]),
		onComplete:   config.OnComplete,
	}
}

// Start begins the frame animation.
// If called before the app is running, the animation will automatically
// start when the controller becomes available (on first Value() access).
func (fa *FrameAnimation[T]) Start() {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	if fa.state == AnimationRunning {
		return
	}

	fa.state = AnimationRunning
	fa.elapsed = 0
	fa.currentFrame = 0
	fa.signal.Set(fa.frames[0])

	// Register with global controller, or mark as pending if not ready
	if currentController != nil {
		fa.handle = currentController.Register(fa)
		fa.pendingStart = false
	} else {
		fa.pendingStart = true
	}
}

// Stop halts the animation.
func (fa *FrameAnimation[T]) Stop() {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	if fa.handle != nil && currentController != nil {
		currentController.Unregister(fa.handle)
		fa.handle = nil
	}
	fa.state = AnimationCompleted
}

// Reset restarts from the first frame.
// Does not automatically start the animation; call Start() after Reset().
func (fa *FrameAnimation[T]) Reset() {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	fa.currentFrame = 0
	fa.elapsed = 0
	fa.signal.Set(fa.frames[0])
	fa.state = AnimationPending
}

// Value returns a signal containing the current frame.
// Call Value().Get() during Build() for reactive updates.
func (fa *FrameAnimation[T]) Value() AnySignal[T] {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	// If Start() was called before controller was ready, register now
	if fa.pendingStart && currentController != nil {
		fa.handle = currentController.Register(fa)
		fa.pendingStart = false
	}
	return fa.signal
}

// Get returns the current frame directly without subscribing.
func (fa *FrameAnimation[T]) Get() T {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	return fa.frames[fa.currentFrame]
}

// Index returns the current frame index.
func (fa *FrameAnimation[T]) Index() int {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	return fa.currentFrame
}

// IsRunning returns true if the animation is currently active.
func (fa *FrameAnimation[T]) IsRunning() bool {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	return fa.state == AnimationRunning
}

// Advance implements the Animator interface.
func (fa *FrameAnimation[T]) Advance(dt time.Duration) bool {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	if fa.state != AnimationRunning {
		return fa.state != AnimationCompleted
	}

	fa.elapsed += dt

	// Calculate how many frames to advance
	if fa.elapsed >= fa.frameTime {
		framesElapsed := int(fa.elapsed / fa.frameTime)
		fa.elapsed -= time.Duration(framesElapsed) * fa.frameTime

		if fa.loop {
			fa.currentFrame = (fa.currentFrame + framesElapsed) % len(fa.frames)
		} else {
			fa.currentFrame += framesElapsed
			if fa.currentFrame >= len(fa.frames) {
				fa.currentFrame = len(fa.frames) - 1
				fa.state = AnimationCompleted
				fa.signal.Set(fa.frames[fa.currentFrame])
				if fa.onComplete != nil {
					fa.onComplete()
				}
				return false
			}
		}

		fa.signal.Set(fa.frames[fa.currentFrame])
	}

	return true
}
