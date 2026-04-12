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
