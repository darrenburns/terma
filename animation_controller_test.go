package terma

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testAnimator struct{}

func (testAnimator) Advance(time.Duration) bool {
	return true
}

func TestAnimationController_TickChannelLifecycle(t *testing.T) {
	controller := NewAnimationController(60)

	if controller.Tick() != nil {
		t.Fatal("expected Tick to be nil before any animations are registered")
	}

	handle := controller.Register(testAnimator{})
	if handle == nil {
		t.Fatal("expected Register to return a handle")
	}
	if controller.Tick() == nil {
		t.Fatal("expected Tick to return a channel while animations are active")
	}

	controller.Unregister(handle)
	if controller.Tick() != nil {
		t.Fatal("expected Tick to be nil after the last animation unregisters")
	}

	controller.Stop()
	if controller.Tick() != nil {
		t.Fatal("expected Tick to remain nil after Stop")
	}
}

func TestAnimationController_ImmediateUnregisterDoesNotLeakTickerGoroutines(t *testing.T) {
	base := runtime.NumGoroutine()

	for i := 0; i < 50; i++ {
		controller := NewAnimationController(60)
		handle := controller.Register(testAnimator{})
		controller.Unregister(handle)
		controller.Stop()
	}

	require.Eventually(t, func() bool {
		// Allow a little slack for goroutines started elsewhere in the test process.
		return runtime.NumGoroutine() <= base+3
	}, time.Second, 10*time.Millisecond)
}

// durationAnimator records the real elapsed time supplied by the controller.
type durationAnimator struct {
	deltas    []time.Duration
	remaining time.Duration
	callback  func()
}

func (a *durationAnimator) Advance(dt time.Duration) bool {
	a.deltas = append(a.deltas, dt)
	if a.callback != nil {
		a.callback()
	}
	a.remaining -= dt
	return a.remaining > 0
}

func TestAnimationControllerElapsedAndRegistration(t *testing.T) {
	now := time.Unix(1, 0)
	controller := NewAnimationController(60)
	controller.now = func() time.Time { return now }
	defer controller.Stop()
	first := &durationAnimator{remaining: time.Second}
	controller.Register(first)
	now = now.Add(10 * time.Millisecond)
	second := &durationAnimator{remaining: time.Second}
	controller.Register(second)
	now = now.Add(25 * time.Millisecond)
	controller.Update()
	require.Equal(t, []time.Duration{35 * time.Millisecond}, first.deltas)
	require.Equal(t, []time.Duration{25 * time.Millisecond}, second.deltas)
	now = now.Add(time.Second)
	controller.Update()
	require.Equal(t, time.Second, first.deltas[1])
	require.False(t, controller.HasActiveAnimations())
	require.Nil(t, controller.Tick())
	// A fresh ticker must not pass the idle gap to a newly started animation.
	now = now.Add(time.Hour)
	next := &durationAnimator{remaining: time.Second}
	controller.Register(next)
	now = now.Add(5 * time.Millisecond)
	controller.Update()
	require.Equal(t, []time.Duration{5 * time.Millisecond}, next.deltas)
	require.NotNil(t, controller.Tick())
}

func TestAnimationControllerPauseResumeAfterStalledTick(t *testing.T) {
	previous := currentController
	controller := NewAnimationController(60)
	currentController = controller
	defer func() { controller.Stop(); currentController = previous }()
	now := time.Unix(1, 0)
	controller.now = func() time.Time { return now }
	anim := NewAnimation(AnimationConfig[float64]{From: 0, To: 1, Duration: time.Second, Easing: EaseLinear})
	anim.Start()
	now = now.Add(250 * time.Millisecond)
	controller.Update()
	require.InDelta(t, 0.25, anim.Get(), 1e-9)
	anim.Pause()
	now = now.Add(time.Hour)
	// No Update runs while paused, as when external work suspends the event loop.
	anim.Resume()
	now = now.Add(10 * time.Millisecond)
	controller.Update()
	require.InDelta(t, 0.26, anim.Get(), 1e-9)
	anim.Pause()
	now = now.Add(time.Second)
	controller.Update()
	require.InDelta(t, 0.26, anim.Get(), 1e-9)
	anim.Resume()
	now = now.Add(740 * time.Millisecond)
	controller.Update()
	require.True(t, anim.IsComplete())
	require.Nil(t, controller.Tick())
}

func TestAnimationControllerCallbacksAndStop(t *testing.T) {
	now := time.Unix(1, 0)
	controller := NewAnimationController(60)
	controller.now = func() time.Time { return now }
	defer controller.Stop()
	next := &durationAnimator{remaining: time.Second}
	first := &durationAnimator{remaining: time.Millisecond, callback: func() { controller.Register(next) }}
	controller.Register(first)
	now = now.Add(5 * time.Millisecond)
	controller.Update()
	require.Empty(t, next.deltas)
	now = now.Add(10 * time.Millisecond)
	controller.Update()
	require.Equal(t, []time.Duration{10 * time.Millisecond}, next.deltas)
	controller.Stop()
	now = now.Add(time.Hour)
	controller.Update()
	require.Len(t, next.deltas, 1)
	require.Nil(t, controller.Register(testAnimator{}))
	require.Nil(t, controller.Tick())
}

func TestAnimationZeroDelta(t *testing.T) {
	for _, duration := range []time.Duration{0, -time.Second, time.Second} {
		t.Run(duration.String(), func(t *testing.T) {
			completed := 0
			anim := NewAnimation(AnimationConfig[float64]{From: 0, To: 100, Duration: duration, OnComplete: func() { completed++ }})
			anim.Start()
			running := anim.Advance(0)
			if duration > 0 {
				require.True(t, running)
				require.Zero(t, anim.Get())
				require.Zero(t, completed)
			} else {
				require.False(t, running)
				require.Equal(t, float64(100), anim.Get())
				require.Equal(t, 1, completed)
			}
		})
	}
	anim := NewAnimation(AnimationConfig[float64]{From: 0, To: 100, Delay: time.Second})
	anim.Start()
	require.True(t, anim.Advance(0))
	require.Zero(t, anim.Get())
	// At the exact delay boundary, the remaining active duration is zero.
	require.False(t, anim.Advance(time.Second))
	require.Equal(t, float64(100), anim.Get())
}

func TestAnimationControllerZeroElapsedCompletesImmediateAnimation(t *testing.T) {
	previous := currentController
	controller := NewAnimationController(60)
	currentController = controller
	defer func() { controller.Stop(); currentController = previous }()
	now := time.Unix(1, 0)
	controller.now = func() time.Time { return now }
	anim := NewAnimation(AnimationConfig[float64]{From: 0, To: 100})
	anim.Start()
	controller.Update()
	require.True(t, anim.IsComplete())
	require.Equal(t, float64(100), anim.Get())
	require.Nil(t, controller.Tick())
}
