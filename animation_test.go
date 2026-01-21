package terma

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimation_InitialValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Before starting, Value should return the From value
	assert.Equal(t, float64(0), anim.Value().Get(), "initial value should be From value")
}

func TestAnimation_ValueReturnsConsistentSignal(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Get signal references before and after advancing
	sig1 := anim.Value()

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	sig2 := anim.Value()

	// Both should reflect the same updated value
	assert.Equal(t, sig1.Get(), sig2.Get(), "signals should show same value")

	// Value should be exactly 50 with linear easing at 50% progress
	assert.Equal(t, float64(50), sig1.Get(), "signal should reflect animation progress")
}

func TestAnimation_StartBeginAnimation(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	assert.False(t, anim.IsRunning(), "animation should not be running before Start()")

	anim.Start()

	assert.True(t, anim.IsRunning(), "animation should be running after Start()")
}

func TestAnimation_AdvanceUpdatesValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	// Advance by 500ms (half the duration)
	anim.Advance(500 * time.Millisecond)

	// Value should be exactly 50 with linear easing at 50% progress
	assert.Equal(t, float64(50), anim.Value().Get(), "value at halfway point")
}

func TestAnimation_CompletesAtEnd(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	assert.False(t, anim.IsComplete(), "animation should not be complete before finishing")

	// Advance past the duration
	anim.Advance(time.Second)

	assert.True(t, anim.IsComplete(), "animation should be complete after duration")
	assert.Equal(t, float64(100), anim.Value().Get(), "final value should be To value")
}

func TestAnimation_OnCompleteCallback(t *testing.T) {
	called := false

	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
		OnComplete: func() {
			called = true
		},
	})

	anim.Start()

	assert.False(t, called, "OnComplete should not be called before animation finishes")

	// Advance to completion
	anim.Advance(time.Second)

	assert.True(t, called, "OnComplete should be called when animation finishes")
}

func TestAnimation_Stop(t *testing.T) {
	called := false

	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
		OnComplete: func() {
			called = true
		},
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond) // Advance halfway

	anim.Stop()

	assert.True(t, anim.IsComplete(), "animation should be marked complete after Stop()")
	assert.False(t, anim.IsRunning(), "animation should not be running after Stop()")
	assert.False(t, called, "OnComplete should not be called when Stop() is used")
	assert.Equal(t, float64(50), anim.Value().Get(), "value should remain where it was stopped")
}

func TestAnimation_Pause(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond) // Advance halfway

	anim.Pause()

	assert.False(t, anim.IsRunning(), "animation should not be running after Pause()")

	valBeforePause := anim.Value().Get()

	// Advance while paused - value should not change
	anim.Advance(500 * time.Millisecond)

	assert.Equal(t, valBeforePause, anim.Value().Get(), "value should not change while paused")
}

func TestAnimation_Resume(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(250 * time.Millisecond) // 25%

	anim.Pause()
	anim.Resume()

	assert.True(t, anim.IsRunning(), "animation should be running after Resume()")

	// Continue advancing
	anim.Advance(750 * time.Millisecond)

	assert.True(t, anim.IsComplete(), "animation should complete after resuming and advancing to end")
	assert.Equal(t, float64(100), anim.Value().Get(), "final value after resume")
}

func TestAnimation_ResumeOnlyWorksWhenPaused(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Resume on pending animation should have no effect
	anim.Resume()
	assert.False(t, anim.IsRunning(), "Resume() on pending animation should not start it")

	anim.Start()
	anim.Advance(time.Second) // Complete

	// Resume on completed animation should have no effect
	anim.Resume()
	assert.False(t, anim.IsRunning(), "Resume() on completed animation should not restart it")
}

func TestAnimation_Reset(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(time.Second) // Complete

	assert.Equal(t, float64(100), anim.Value().Get(), "value after completion")

	anim.Reset()

	assert.Equal(t, float64(0), anim.Value().Get(), "value should be back to From after Reset()")
	assert.False(t, anim.IsRunning(), "animation should not be running after Reset() - must call Start()")
	assert.False(t, anim.IsComplete(), "animation should not be complete after Reset()")
}

func TestAnimation_ResetThenStart(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(time.Second) // Complete

	anim.Reset()
	anim.Start()

	assert.True(t, anim.IsRunning(), "animation should be running after Reset() and Start()")

	// Advance again
	anim.Advance(500 * time.Millisecond)

	assert.Equal(t, float64(50), anim.Value().Get(), "value after restart")
}

func TestAnimation_IsRunning(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	// Pending
	assert.False(t, anim.IsRunning(), "IsRunning() should be false when pending")

	anim.Start()
	assert.True(t, anim.IsRunning(), "IsRunning() should be true after Start()")

	anim.Pause()
	assert.False(t, anim.IsRunning(), "IsRunning() should be false when paused")

	anim.Resume()
	assert.True(t, anim.IsRunning(), "IsRunning() should be true after Resume()")

	anim.Advance(time.Second)
	assert.False(t, anim.IsRunning(), "IsRunning() should be false when complete")
}

func TestAnimation_IsComplete(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	assert.False(t, anim.IsComplete(), "IsComplete() should be false when pending")

	anim.Start()
	assert.False(t, anim.IsComplete(), "IsComplete() should be false when running")

	anim.Advance(500 * time.Millisecond)
	assert.False(t, anim.IsComplete(), "IsComplete() should be false before duration elapsed")

	anim.Advance(500 * time.Millisecond)
	assert.True(t, anim.IsComplete(), "IsComplete() should be true after duration elapsed")
}

func TestAnimation_Progress(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	assert.Equal(t, 0.0, anim.Progress(), "Progress() at start")

	anim.Advance(250 * time.Millisecond)
	assert.Equal(t, 0.25, anim.Progress(), "Progress() at 25%")

	anim.Advance(250 * time.Millisecond)
	assert.Equal(t, 0.5, anim.Progress(), "Progress() at 50%")

	anim.Advance(500 * time.Millisecond)
	assert.Equal(t, 1.0, anim.Progress(), "Progress() at end")
}

func TestAnimation_Looping(t *testing.T) {
	loopCount := 0

	var anim *Animation[float64]
	anim = NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: 100 * time.Millisecond,
		OnComplete: func() {
			loopCount++
			if loopCount < 3 {
				anim.Reset()
				anim.Start()
			}
		},
	})

	anim.Start()

	// Run through 3 complete loops
	for i := 0; i < 3; i++ {
		anim.Advance(100 * time.Millisecond)
	}

	assert.Equal(t, 3, loopCount, "loop count")
}

func TestAnimation_EasingFunction(t *testing.T) {
	// Test with a custom easing that always returns 1 (jumps to end)
	jumpToEnd := func(t float64) float64 { return 1.0 }

	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
		Easing:   jumpToEnd,
	})

	anim.Start()
	anim.Advance(1 * time.Millisecond) // Tiny advance

	// Due to easing, value should jump to 100 immediately
	assert.Equal(t, float64(100), anim.Value().Get(), "value with jump-to-end easing")
}

func TestAnimation_DefaultEasingIsLinear(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
		// No Easing specified - should default to linear
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	// Linear easing: 50% time = 50% value
	assert.Equal(t, float64(50), anim.Value().Get(), "value with linear easing at 50%")
}

func TestAnimation_OnUpdateCallback(t *testing.T) {
	var updates []float64

	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
		OnUpdate: func(v float64) {
			updates = append(updates, v)
		},
	})

	anim.Start()
	anim.Advance(250 * time.Millisecond)
	anim.Advance(250 * time.Millisecond)
	anim.Advance(500 * time.Millisecond)

	require.Len(t, updates, 3, "number of updates")

	// Check values are increasing
	for i := 1; i < len(updates); i++ {
		assert.Greater(t, updates[i], updates[i-1], "updates should be increasing")
	}
}

func TestAnimation_Delay(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
		Delay:    500 * time.Millisecond,
	})

	anim.Start()

	// Advance but still within delay
	anim.Advance(250 * time.Millisecond)
	assert.Equal(t, float64(0), anim.Value().Get(), "value during delay")

	// Advance to delay boundary - animation time starts but dt is adjusted to 0
	anim.Advance(250 * time.Millisecond)

	// Now advance 500ms into the actual animation (half of 1s duration)
	anim.Advance(500 * time.Millisecond)
	assert.Equal(t, float64(50), anim.Value().Get(), "value after delay + half duration")
}

func TestAnimation_GetReturnsCurrentValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	assert.Equal(t, float64(0), anim.Get(), "Get() initially")

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	assert.Equal(t, float64(50), anim.Get(), "Get() at halfway")
}

func TestAnimation_IntegerInterpolation(t *testing.T) {
	anim := NewAnimation(AnimationConfig[int]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	assert.Equal(t, 50, anim.Value().Get(), "int value at halfway")
}

func TestAnimation_ZeroDuration(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: 0,
	})

	// Progress should be 1.0 for zero duration
	assert.Equal(t, 1.0, anim.Progress(), "Progress() for zero duration")

	// Verify Advance() works correctly with zero duration
	anim.Start()
	continuing := anim.Advance(time.Millisecond)

	assert.False(t, continuing, "Advance() should return false for zero-duration animation")
	assert.True(t, anim.IsComplete(), "zero-duration animation should be complete after Advance()")
	assert.Equal(t, float64(100), anim.Value().Get(), "value after zero-duration animation")
}

func TestAnimation_StartIsIdempotentWhileRunning(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	valBefore := anim.Value().Get()

	// Calling Start() again while running should not restart
	anim.Start()

	assert.Equal(t, valBefore, anim.Value().Get(), "Start() while running should not reset value")
}

func TestAnimation_PauseOnlyWorksWhenRunning(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Pause on pending animation should have no effect
	anim.Pause()
	anim.Start()

	assert.True(t, anim.IsRunning(), "animation should still be able to start after Pause() on pending state")
}

func TestAnimation_AdvanceReturnValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	// Should return true while animation continues
	assert.True(t, anim.Advance(500*time.Millisecond), "Advance() should return true while animation is in progress")

	// Should return false when animation completes
	assert.False(t, anim.Advance(500*time.Millisecond), "Advance() should return false when animation completes")
}

// =============================================================================
// AnimatedValue tests
// =============================================================================
// AnimatedValue wraps Animation and provides a simpler interface for
// animating between values set via Set(). The animation mechanics are
// tested above; these tests verify the AnimatedValue-specific behavior.

func TestAnimatedValue_InitialValue(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  42.0,
		Duration: time.Second,
	})

	assert.Equal(t, 42.0, av.Get(), "initial value")
	assert.Equal(t, 42.0, av.Target(), "initial target")
}

func TestAnimatedValue_Peek(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  100.0,
		Duration: time.Second,
	})

	assert.Equal(t, 100.0, av.Peek(), "Peek()")
}

func TestAnimatedValue_SetImmediate(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(75.0)

	assert.Equal(t, 75.0, av.Get(), "immediate value")
	assert.Equal(t, 75.0, av.Target(), "target after SetImmediate")
	assert.False(t, av.IsAnimating(), "should not be animating after SetImmediate")
}

func TestAnimatedValue_SetUpdatesTarget(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0)

	assert.Equal(t, 100.0, av.Target(), "target after Set")
}

func TestAnimatedValue_SetSameValueNoOp(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  50.0,
		Duration: time.Second,
	})

	// Set to same value should not start animation
	av.Set(50.0)

	assert.False(t, av.IsAnimating(), "setting same value should not start animation")
}

func TestAnimatedValue_IsAnimatingAfterSet(t *testing.T) {
	// Note: Animation.Start() sets state to Running even without a controller,
	// so IsAnimating() works correctly without needing a real controller.
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	assert.False(t, av.IsAnimating(), "should not be animating initially")

	av.Set(100.0)

	assert.True(t, av.IsAnimating(), "should be animating after Set()")
}

func TestAnimatedValue_Signal(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  25.0,
		Duration: time.Second,
	})

	sig := av.Signal()

	assert.Equal(t, 25.0, sig.Get(), "Signal().Get() initial value")

	av.SetImmediate(50.0)

	assert.Equal(t, 50.0, sig.Get(), "Signal().Get() after SetImmediate")
}

func TestAnimatedValue_MultipleSetImmediate(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(10.0)
	av.SetImmediate(20.0)
	av.SetImmediate(30.0)

	assert.Equal(t, 30.0, av.Get(), "final value after multiple SetImmediate")
}

func TestAnimatedValue_SetImmediateStopsAnimation(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0) // Start animation

	assert.True(t, av.IsAnimating(), "should be animating after Set()")

	av.SetImmediate(50.0) // Should stop animation and set value

	assert.False(t, av.IsAnimating(), "SetImmediate should stop running animation")
	assert.Equal(t, 50.0, av.Get(), "value after SetImmediate")
}

func TestAnimatedValue_RetargetingUpdatesTarget(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(50.0)
	assert.Equal(t, 50.0, av.Target(), "first target")

	// Retarget while animating
	av.Set(100.0)
	assert.Equal(t, 100.0, av.Target(), "retargeted target")
}

func TestAnimatedValue_IntegerType(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[int]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(42)

	assert.Equal(t, 42, av.Get(), "int value")
}

// =============================================================================
// AnyAnimatedValue tests
// =============================================================================
// AnyAnimatedValue is for non-comparable types. It always starts a new
// animation on Set() since equality cannot be checked.

func TestAnyAnimatedValue_InitialValue(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  42.0,
		Duration: time.Second,
	})

	assert.Equal(t, 42.0, av.Get(), "initial value")
	assert.Equal(t, 42.0, av.Target(), "initial target")
}

func TestAnyAnimatedValue_Peek(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  100.0,
		Duration: time.Second,
	})

	assert.Equal(t, 100.0, av.Peek(), "Peek()")
}

func TestAnyAnimatedValue_SetImmediate(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(75.0)

	assert.Equal(t, 75.0, av.Get(), "immediate value")
	assert.Equal(t, 75.0, av.Target(), "target after SetImmediate")
	assert.False(t, av.IsAnimating(), "should not be animating after SetImmediate")
}

func TestAnyAnimatedValue_SetUpdatesTarget(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0)

	assert.Equal(t, 100.0, av.Target(), "target after Set")
}

func TestAnyAnimatedValue_SetAlwaysStartsAnimation(t *testing.T) {
	// Unlike AnimatedValue, AnyAnimatedValue always starts animation on Set
	// because it cannot check equality for non-comparable types
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  50.0,
		Duration: time.Second,
	})

	// Set to same value still starts animation (can't check equality)
	av.Set(50.0)

	assert.True(t, av.IsAnimating(), "Set() should always start animation for AnyAnimatedValue")
}

func TestAnyAnimatedValue_IsAnimatingAfterSet(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	assert.False(t, av.IsAnimating(), "should not be animating initially")

	av.Set(100.0)

	assert.True(t, av.IsAnimating(), "should be animating after Set()")
}

func TestAnyAnimatedValue_Signal(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  25.0,
		Duration: time.Second,
	})

	sig := av.Signal()

	assert.Equal(t, 25.0, sig.Get(), "Signal().Get() initial value")

	av.SetImmediate(50.0)

	assert.Equal(t, 50.0, sig.Get(), "Signal().Get() after SetImmediate")
}

func TestAnyAnimatedValue_SetImmediateStopsAnimation(t *testing.T) {
	av := NewAnyAnimatedValue(AnyAnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0) // Start animation

	assert.True(t, av.IsAnimating(), "should be animating after Set()")

	av.SetImmediate(50.0) // Should stop animation and set value

	assert.False(t, av.IsAnimating(), "SetImmediate should stop running animation")
	assert.Equal(t, 50.0, av.Get(), "value after SetImmediate")
}

// =============================================================================
// FrameAnimation tests
// =============================================================================

func TestFrameAnimation_InitialValue(t *testing.T) {
	frames := []string{"a", "b", "c", "d"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
	})

	assert.Equal(t, "a", fa.Value().Get(), "initial value should be first frame")
	assert.Equal(t, "a", fa.Get(), "Get() should return first frame")
	assert.Equal(t, 0, fa.Index(), "initial index should be 0")
}

func TestFrameAnimation_Start(t *testing.T) {
	frames := []string{"a", "b", "c"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
	})

	assert.False(t, fa.IsRunning(), "should not be running before Start()")

	fa.Start()

	assert.True(t, fa.IsRunning(), "should be running after Start()")
}

func TestFrameAnimation_AdvanceChangesFrame(t *testing.T) {
	frames := []string{"a", "b", "c", "d"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
	})

	fa.Start()

	// Advance less than one frame time
	fa.Advance(50 * time.Millisecond)
	assert.Equal(t, "a", fa.Get(), "should still be on first frame")
	assert.Equal(t, 0, fa.Index())

	// Advance to complete first frame
	fa.Advance(50 * time.Millisecond)
	assert.Equal(t, "b", fa.Get(), "should be on second frame")
	assert.Equal(t, 1, fa.Index())

	// Advance through multiple frames at once
	fa.Advance(200 * time.Millisecond)
	assert.Equal(t, "d", fa.Get(), "should be on fourth frame")
	assert.Equal(t, 3, fa.Index())
}

func TestFrameAnimation_NonLoopingCompletes(t *testing.T) {
	frames := []string{"a", "b", "c"}
	completed := false

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
		Loop:      false,
		OnComplete: func() {
			completed = true
		},
	})

	fa.Start()

	// Advance through all frames
	continuing := fa.Advance(300 * time.Millisecond)

	assert.False(t, continuing, "Advance() should return false when complete")
	assert.True(t, completed, "OnComplete should be called")
	assert.Equal(t, "c", fa.Get(), "should be on last frame")
	assert.Equal(t, 2, fa.Index())
}

func TestFrameAnimation_LoopingWrapsAround(t *testing.T) {
	frames := []string{"a", "b", "c"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
		Loop:      true,
	})

	fa.Start()

	// Advance through all frames and wrap around
	fa.Advance(300 * time.Millisecond) // Back to frame 0
	assert.Equal(t, "a", fa.Get(), "should wrap to first frame")
	assert.Equal(t, 0, fa.Index())

	fa.Advance(100 * time.Millisecond) // Frame 1
	assert.Equal(t, "b", fa.Get())
	assert.Equal(t, 1, fa.Index())
}

func TestFrameAnimation_LoopingContinues(t *testing.T) {
	frames := []string{"a", "b"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
		Loop:      true,
	})

	fa.Start()

	// Advance through multiple loops
	for i := 0; i < 10; i++ {
		continuing := fa.Advance(100 * time.Millisecond)
		assert.True(t, continuing, "looping animation should always continue")
	}

	assert.True(t, fa.IsRunning(), "should still be running")
}

func TestFrameAnimation_Stop(t *testing.T) {
	frames := []string{"a", "b", "c"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
		Loop:      true,
	})

	fa.Start()
	fa.Advance(150 * time.Millisecond) // Partway through frame 1

	fa.Stop()

	assert.False(t, fa.IsRunning(), "should not be running after Stop()")
}

func TestFrameAnimation_Reset(t *testing.T) {
	frames := []string{"a", "b", "c"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
	})

	fa.Start()
	fa.Advance(200 * time.Millisecond) // On frame 2

	assert.Equal(t, "c", fa.Get())

	fa.Reset()

	assert.Equal(t, "a", fa.Get(), "should be back to first frame after Reset()")
	assert.Equal(t, 0, fa.Index())
	assert.False(t, fa.IsRunning(), "should not be running after Reset()")
}

func TestFrameAnimation_PanicsOnEmptyFrames(t *testing.T) {
	assert.Panics(t, func() {
		NewFrameAnimation(FrameAnimationConfig[string]{
			Frames:    []string{},
			FrameTime: 100 * time.Millisecond,
		})
	}, "should panic with empty frames")
}

func TestFrameAnimation_IntFrames(t *testing.T) {
	frames := []int{0, 1, 2, 3}

	fa := NewFrameAnimation(FrameAnimationConfig[int]{
		Frames:    frames,
		FrameTime: 50 * time.Millisecond,
	})

	fa.Start()

	assert.Equal(t, 0, fa.Get())

	fa.Advance(50 * time.Millisecond)
	assert.Equal(t, 1, fa.Get())

	fa.Advance(100 * time.Millisecond)
	assert.Equal(t, 3, fa.Get())
}

func TestFrameAnimation_SignalUpdates(t *testing.T) {
	frames := []string{"a", "b", "c"}

	fa := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    frames,
		FrameTime: 100 * time.Millisecond,
	})

	sig := fa.Value()
	assert.Equal(t, "a", sig.Get(), "signal initial value")

	fa.Start()
	fa.Advance(100 * time.Millisecond)

	assert.Equal(t, "b", sig.Get(), "signal should update after Advance()")
}
