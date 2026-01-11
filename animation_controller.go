package terma

import (
	"sync"
	"time"
)

// currentController is the animation controller for the currently running app.
// Set by Run() and used by animations to register themselves.
var currentController *AnimationController

// Animator is the interface for anything that can be animated.
type Animator interface {
	// Advance moves the animation forward by the given duration.
	// Returns true if the animation is still running, false if complete.
	Advance(dt time.Duration) bool
}

// animationHandle is an opaque reference to a registered animation.
type animationHandle struct {
	animation Animator
}

// AnimationController manages all active animations.
// It provides a tick channel that only sends when animations are active,
// implementing the "no animations = no ticker" optimization.
type AnimationController struct {
	mu         sync.Mutex
	animations map[*animationHandle]struct{}
	ticker     *time.Ticker
	tickChan   chan time.Time
	fps        int
	stopped    bool
}

// NewAnimationController creates a new controller with the given target FPS.
func NewAnimationController(fps int) *AnimationController {
	if fps <= 0 {
		fps = 60
	}
	return &AnimationController{
		animations: make(map[*animationHandle]struct{}),
		tickChan:   make(chan time.Time, 1),
		fps:        fps,
	}
}

// Tick returns a channel that receives when animation updates are needed.
// Returns nil when no animations are active (nil channel blocks forever in select).
func (ac *AnimationController) Tick() <-chan time.Time {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if len(ac.animations) == 0 || ac.stopped {
		return nil
	}
	return ac.tickChan
}

// Register adds an animation to be managed by the controller.
// Returns a handle that can be used to unregister the animation.
func (ac *AnimationController) Register(anim Animator) *animationHandle {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.stopped {
		return nil
	}

	handle := &animationHandle{animation: anim}
	ac.animations[handle] = struct{}{}

	// Start ticker if this is the first animation
	if len(ac.animations) == 1 {
		ac.startTicker()
	}

	return handle
}

// Unregister removes an animation from the controller.
func (ac *AnimationController) Unregister(handle *animationHandle) {
	if handle == nil {
		return
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()

	delete(ac.animations, handle)

	// Stop ticker if no more animations
	if len(ac.animations) == 0 {
		ac.stopTicker()
	}
}

// Update advances all animations and removes completed ones.
// Called from the event loop when Tick() receives.
func (ac *AnimationController) Update() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.stopped {
		return
	}

	dt := time.Duration(float64(time.Second) / float64(ac.fps))

	// Collect handles to remove
	var toRemove []*animationHandle

	for handle := range ac.animations {
		if !handle.animation.Advance(dt) {
			toRemove = append(toRemove, handle)
		}
	}

	// Remove completed animations
	for _, handle := range toRemove {
		delete(ac.animations, handle)
	}

	// Stop ticker if all animations complete
	if len(ac.animations) == 0 {
		ac.stopTicker()
	}
}

// Stop halts the controller and cleans up resources.
// Called when Run() exits.
func (ac *AnimationController) Stop() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.stopped = true
	ac.stopTicker()
	ac.animations = make(map[*animationHandle]struct{})
}

// HasActiveAnimations returns true if any animations are running.
func (ac *AnimationController) HasActiveAnimations() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return len(ac.animations) > 0
}

// startTicker begins the animation tick loop.
// Must be called with mutex held.
func (ac *AnimationController) startTicker() {
	if ac.ticker != nil {
		return
	}

	interval := time.Duration(float64(time.Second) / float64(ac.fps))
	ac.ticker = time.NewTicker(interval)

	// Pump ticker events to tickChan
	go func() {
		for t := range ac.ticker.C {
			select {
			case ac.tickChan <- t:
			default:
				// Drop tick if channel is full (avoid blocking)
			}
		}
	}()
}

// stopTicker halts the animation tick loop.
// Must be called with mutex held.
func (ac *AnimationController) stopTicker() {
	if ac.ticker == nil {
		return
	}
	ac.ticker.Stop()
	ac.ticker = nil
}
