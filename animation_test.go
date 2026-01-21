package terma

import (
	"testing"
	"time"
)

func TestAnimation_InitialValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Before starting, Value should return the From value
	if anim.Value().Get() != 0 {
		t.Errorf("expected initial value 0, got %f", anim.Value().Get())
	}
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
	if sig1.Get() != sig2.Get() {
		t.Errorf("signals should show same value: sig1=%f, sig2=%f", sig1.Get(), sig2.Get())
	}

	// Value should be updated
	if sig1.Get() < 49 || sig1.Get() > 51 {
		t.Errorf("signal should reflect animation progress, got %f", sig1.Get())
	}
}

func TestAnimation_StartBeginAnimation(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	if anim.IsRunning() {
		t.Error("animation should not be running before Start()")
	}

	anim.Start()

	if !anim.IsRunning() {
		t.Error("animation should be running after Start()")
	}
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

	// Value should be approximately 50 (halfway)
	val := anim.Value().Get()
	if val < 49 || val > 51 {
		t.Errorf("expected value near 50 at halfway point, got %f", val)
	}
}

func TestAnimation_CompletesAtEnd(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	if anim.IsComplete() {
		t.Error("animation should not be complete before finishing")
	}

	// Advance past the duration
	anim.Advance(time.Second)

	if !anim.IsComplete() {
		t.Error("animation should be complete after duration")
	}

	// Value should be at the target
	if anim.Value().Get() != 100 {
		t.Errorf("expected final value 100, got %f", anim.Value().Get())
	}
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

	if called {
		t.Error("OnComplete should not be called before animation finishes")
	}

	// Advance to completion
	anim.Advance(time.Second)

	if !called {
		t.Error("OnComplete should be called when animation finishes")
	}
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

	if !anim.IsComplete() {
		t.Error("animation should be marked complete after Stop()")
	}

	if anim.IsRunning() {
		t.Error("animation should not be running after Stop()")
	}

	// OnComplete should NOT be called when stopping
	if called {
		t.Error("OnComplete should not be called when Stop() is used")
	}

	// Value should remain where it was stopped
	val := anim.Value().Get()
	if val < 49 || val > 51 {
		t.Errorf("expected value near 50 after stopping halfway, got %f", val)
	}
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

	if anim.IsRunning() {
		t.Error("animation should not be running after Pause()")
	}

	valBeforePause := anim.Value().Get()

	// Advance while paused - value should not change
	anim.Advance(500 * time.Millisecond)

	if anim.Value().Get() != valBeforePause {
		t.Errorf("value should not change while paused, was %f, now %f", valBeforePause, anim.Value().Get())
	}
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

	if !anim.IsRunning() {
		t.Error("animation should be running after Resume()")
	}

	// Continue advancing
	anim.Advance(750 * time.Millisecond)

	if !anim.IsComplete() {
		t.Error("animation should complete after resuming and advancing to end")
	}

	if anim.Value().Get() != 100 {
		t.Errorf("expected final value 100, got %f", anim.Value().Get())
	}
}

func TestAnimation_ResumeOnlyWorksWhenPaused(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	// Resume on pending animation should have no effect
	anim.Resume()
	if anim.IsRunning() {
		t.Error("Resume() on pending animation should not start it")
	}

	anim.Start()
	anim.Advance(time.Second) // Complete

	// Resume on completed animation should have no effect
	anim.Resume()
	if anim.IsRunning() {
		t.Error("Resume() on completed animation should not restart it")
	}
}

func TestAnimation_Reset(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(time.Second) // Complete

	if anim.Value().Get() != 100 {
		t.Errorf("expected value 100 after completion, got %f", anim.Value().Get())
	}

	anim.Reset()

	// After reset, value should be back to From
	if anim.Value().Get() != 0 {
		t.Errorf("expected value 0 after Reset(), got %f", anim.Value().Get())
	}

	// Animation should not be running (need to call Start)
	if anim.IsRunning() {
		t.Error("animation should not be running after Reset() - must call Start()")
	}

	if anim.IsComplete() {
		t.Error("animation should not be complete after Reset()")
	}
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

	if !anim.IsRunning() {
		t.Error("animation should be running after Reset() and Start()")
	}

	// Advance again
	anim.Advance(500 * time.Millisecond)

	val := anim.Value().Get()
	if val < 49 || val > 51 {
		t.Errorf("expected value near 50 after restart, got %f", val)
	}
}

func TestAnimation_IsRunning(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	// Pending
	if anim.IsRunning() {
		t.Error("IsRunning() should be false when pending")
	}

	anim.Start()
	if !anim.IsRunning() {
		t.Error("IsRunning() should be true after Start()")
	}

	anim.Pause()
	if anim.IsRunning() {
		t.Error("IsRunning() should be false when paused")
	}

	anim.Resume()
	if !anim.IsRunning() {
		t.Error("IsRunning() should be true after Resume()")
	}

	anim.Advance(time.Second)
	if anim.IsRunning() {
		t.Error("IsRunning() should be false when complete")
	}
}

func TestAnimation_IsComplete(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: time.Second,
	})

	if anim.IsComplete() {
		t.Error("IsComplete() should be false when pending")
	}

	anim.Start()
	if anim.IsComplete() {
		t.Error("IsComplete() should be false when running")
	}

	anim.Advance(500 * time.Millisecond)
	if anim.IsComplete() {
		t.Error("IsComplete() should be false before duration elapsed")
	}

	anim.Advance(500 * time.Millisecond)
	if !anim.IsComplete() {
		t.Error("IsComplete() should be true after duration elapsed")
	}
}

func TestAnimation_Progress(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	if anim.Progress() != 0 {
		t.Errorf("expected Progress() = 0 at start, got %f", anim.Progress())
	}

	anim.Advance(250 * time.Millisecond)
	if anim.Progress() != 0.25 {
		t.Errorf("expected Progress() = 0.25 at 25%%, got %f", anim.Progress())
	}

	anim.Advance(250 * time.Millisecond)
	if anim.Progress() != 0.5 {
		t.Errorf("expected Progress() = 0.5 at 50%%, got %f", anim.Progress())
	}

	anim.Advance(500 * time.Millisecond)
	if anim.Progress() != 1.0 {
		t.Errorf("expected Progress() = 1.0 at end, got %f", anim.Progress())
	}
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

	if loopCount != 3 {
		t.Errorf("expected 3 loops, got %d", loopCount)
	}
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
	if anim.Value().Get() != 100 {
		t.Errorf("expected value 100 with jump-to-end easing, got %f", anim.Value().Get())
	}
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
	val := anim.Value().Get()
	if val != 50 {
		t.Errorf("expected value 50 with linear easing at 50%%, got %f", val)
	}
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

	if len(updates) != 3 {
		t.Errorf("expected 3 updates, got %d", len(updates))
	}

	// Check values are increasing
	for i := 1; i < len(updates); i++ {
		if updates[i] <= updates[i-1] {
			t.Errorf("updates should be increasing: %v", updates)
		}
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
	if anim.Value().Get() != 0 {
		t.Errorf("expected value 0 during delay, got %f", anim.Value().Get())
	}

	// Advance past delay
	anim.Advance(250 * time.Millisecond)
	// Now at delay boundary, animation should start

	anim.Advance(500 * time.Millisecond)
	val := anim.Value().Get()
	if val < 49 || val > 51 {
		t.Errorf("expected value near 50 after delay + half duration, got %f", val)
	}
}

func TestAnimation_GetReturnsCurrentValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	if anim.Get() != 0 {
		t.Errorf("expected Get() = 0 initially, got %f", anim.Get())
	}

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	val := anim.Get()
	if val < 49 || val > 51 {
		t.Errorf("expected Get() near 50, got %f", val)
	}
}

func TestAnimation_IntegerInterpolation(t *testing.T) {
	anim := NewAnimation(AnimationConfig[int]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()
	anim.Advance(500 * time.Millisecond)

	val := anim.Value().Get()
	if val != 50 {
		t.Errorf("expected int value 50, got %d", val)
	}
}

func TestAnimation_ZeroDuration(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: 0,
	})

	// Progress should be 1.0 for zero duration
	if anim.Progress() != 1.0 {
		t.Errorf("expected Progress() = 1.0 for zero duration, got %f", anim.Progress())
	}
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

	if anim.Value().Get() != valBefore {
		t.Errorf("Start() while running should not reset value, was %f, now %f", valBefore, anim.Value().Get())
	}
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

	if !anim.IsRunning() {
		t.Error("animation should still be able to start after Pause() on pending state")
	}
}

func TestAnimation_AdvanceReturnValue(t *testing.T) {
	anim := NewAnimation(AnimationConfig[float64]{
		From:     0,
		To:       100,
		Duration: time.Second,
	})

	anim.Start()

	// Should return true while animation continues
	if !anim.Advance(500 * time.Millisecond) {
		t.Error("Advance() should return true while animation is in progress")
	}

	// Should return false when animation completes
	if anim.Advance(500 * time.Millisecond) {
		t.Error("Advance() should return false when animation completes")
	}
}

// AnimatedValue tests
// AnimatedValue wraps Animation and provides a simpler interface for
// animating between values set via Set(). The animation mechanics are
// tested above; these tests verify the AnimatedValue-specific behavior.

func TestAnimatedValue_InitialValue(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  42.0,
		Duration: time.Second,
	})

	if av.Get() != 42.0 {
		t.Errorf("expected initial value 42.0, got %f", av.Get())
	}

	if av.Target() != 42.0 {
		t.Errorf("expected initial target 42.0, got %f", av.Target())
	}
}

func TestAnimatedValue_Peek(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  100.0,
		Duration: time.Second,
	})

	if av.Peek() != 100.0 {
		t.Errorf("expected Peek() = 100.0, got %f", av.Peek())
	}
}

func TestAnimatedValue_SetImmediate(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(75.0)

	// Value should change immediately without animation
	if av.Get() != 75.0 {
		t.Errorf("expected immediate value 75.0, got %f", av.Get())
	}

	if av.Target() != 75.0 {
		t.Errorf("expected target 75.0 after SetImmediate, got %f", av.Target())
	}

	if av.IsAnimating() {
		t.Error("should not be animating after SetImmediate")
	}
}

func TestAnimatedValue_SetUpdatesTarget(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0)

	// Target should be updated immediately
	if av.Target() != 100.0 {
		t.Errorf("expected target 100.0 after Set, got %f", av.Target())
	}
}

func TestAnimatedValue_SetSameValueNoOp(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  50.0,
		Duration: time.Second,
	})

	// Set to same value should not start animation
	av.Set(50.0)

	if av.IsAnimating() {
		t.Error("setting same value should not start animation")
	}
}

func TestAnimatedValue_IsAnimatingAfterSet(t *testing.T) {
	// Note: Animation.Start() sets state to Running even without a controller,
	// so IsAnimating() works correctly without needing a real controller.
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	if av.IsAnimating() {
		t.Error("should not be animating initially")
	}

	av.Set(100.0)

	if !av.IsAnimating() {
		t.Error("should be animating after Set()")
	}
}

func TestAnimatedValue_Signal(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  25.0,
		Duration: time.Second,
	})

	sig := av.Signal()

	if sig.Get() != 25.0 {
		t.Errorf("Signal().Get() should return initial value, got %f", sig.Get())
	}

	av.SetImmediate(50.0)

	if sig.Get() != 50.0 {
		t.Errorf("Signal().Get() should reflect updated value, got %f", sig.Get())
	}
}

func TestAnimatedValue_MultipleSetImmediate(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(10.0)
	av.SetImmediate(20.0)
	av.SetImmediate(30.0)

	if av.Get() != 30.0 {
		t.Errorf("expected final value 30.0, got %f", av.Get())
	}
}

func TestAnimatedValue_SetImmediateStopsAnimation(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(100.0) // Start animation

	if !av.IsAnimating() {
		t.Error("should be animating after Set()")
	}

	av.SetImmediate(50.0) // Should stop animation and set value

	if av.IsAnimating() {
		t.Error("SetImmediate should stop running animation")
	}

	if av.Get() != 50.0 {
		t.Errorf("expected value 50.0 after SetImmediate, got %f", av.Get())
	}
}

func TestAnimatedValue_RetargetingUpdatesTarget(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[float64]{
		Initial:  0,
		Duration: time.Second,
	})

	av.Set(50.0)
	if av.Target() != 50.0 {
		t.Errorf("expected target 50.0, got %f", av.Target())
	}

	// Retarget while animating
	av.Set(100.0)
	if av.Target() != 100.0 {
		t.Errorf("expected retargeted target 100.0, got %f", av.Target())
	}
}

func TestAnimatedValue_IntegerType(t *testing.T) {
	av := NewAnimatedValue(AnimatedValueConfig[int]{
		Initial:  0,
		Duration: time.Second,
	})

	av.SetImmediate(42)

	if av.Get() != 42 {
		t.Errorf("expected int value 42, got %d", av.Get())
	}
}
