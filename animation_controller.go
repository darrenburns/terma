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
	animation   Animator
	lastAdvance time.Time
}

// AnimationController manages all active animations.
// It provides a tick channel that only sends when animations are active,
// implementing the "no animations = no ticker" optimization.
type AnimationController struct {
	mu         sync.Mutex
	animations map[*animationHandle]struct{}
	ticker     *time.Ticker
	fps        int
	stopped    bool
	now        func() time.Time
}

// NewAnimationController creates a new controller with the given target FPS.
func NewAnimationController(fps int) *AnimationController {
	if fps <= 0 {
		fps = 60
	}
	return &AnimationController{
		animations: make(map[*animationHandle]struct{}),
		fps:        fps,
		now:        time.Now,
	}
}

// Tick returns a channel that receives when animation updates are needed.
// Returns nil when no animations are active (nil channel blocks forever in select).
func (ac *AnimationController) Tick() <-chan time.Time {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if len(ac.animations) == 0 || ac.stopped || ac.ticker == nil {
		return nil
	}
	return ac.ticker.C
}

// Register adds an animation to be managed by the controller.
// Returns a handle that can be used to unregister the animation.
func (ac *AnimationController) Register(anim Animator) *animationHandle {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.stopped {
		return nil
	}

	handle := &animationHandle{animation: anim, lastAdvance: ac.now()}
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
	if ac.stopped {
		ac.mu.Unlock()
		return
	}

	// Sample processing time rather than ticker timestamps: ticks may be dropped
	// while rendering is slow. Each handle starts at its own registration time.
	now := ac.now()
	handles := make([]*animationHandle, 0, len(ac.animations))
	for handle := range ac.animations {
		handles = append(handles, handle)
	}
	ac.mu.Unlock()

	// Advance outside the lock (callbacks may register, remove, or resume an
	// animation). Read its timestamp immediately before advancing so a resume
	// from another callback cannot inherit elapsed time from before the pause.
	var toRemove []*animationHandle
	for _, handle := range handles {
		ac.mu.Lock()
		_, active := ac.animations[handle]
		dt := now.Sub(handle.lastAdvance)
		if dt < 0 {
			dt = 0
		} else {
			handle.lastAdvance = now
		}
		ac.mu.Unlock()
		if active && !handle.animation.Advance(dt) {
			toRemove = append(toRemove, handle)
		}
	}

	// Re-acquire lock to clean up
	ac.mu.Lock()
	for _, handle := range toRemove {
		delete(ac.animations, handle)
	}

	// Stop ticker if all animations complete
	if len(ac.animations) == 0 {
		ac.stopTicker()
	}
	ac.mu.Unlock()
}

// resetElapsed excludes paused time even if no update ran during the pause.
func (ac *AnimationController) resetElapsed(handle *animationHandle) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if _, active := ac.animations[handle]; active {
		handle.lastAdvance = ac.now()
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
