package terma

import (
	"fmt"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeToastClock replaces a ToastState's clock and timers so tests decide
// when time passes.
type fakeToastClock struct {
	now    time.Time
	timers []*fakeToastTimer
}

type fakeToastTimer struct {
	at      time.Time
	fn      func()
	stopped bool
}

func newTestToastState(options ToastOptions) (*ToastState, *fakeToastClock) {
	state := NewToastState(options)
	clock := &fakeToastClock{now: time.Unix(0, 0)}
	state.now = func() time.Time { return clock.now }
	state.after = func(d time.Duration, fn func()) func() bool {
		timer := &fakeToastTimer{at: clock.now.Add(d), fn: fn}
		clock.timers = append(clock.timers, timer)
		return func() bool {
			wasActive := !timer.stopped
			timer.stopped = true
			return wasActive
		}
	}
	return state, clock
}

func (c *fakeToastClock) advance(d time.Duration) {
	c.now = c.now.Add(d)
	for i := 0; i < len(c.timers); i++ {
		timer := c.timers[i]
		if !timer.stopped && !timer.at.After(c.now) {
			timer.stopped = true
			timer.fn()
		}
	}
}

func (c *fakeToastClock) running() int {
	count := 0
	for _, timer := range c.timers {
		if !timer.stopped {
			count++
		}
	}
	return count
}

func toastMessages(state *ToastState) []string {
	var messages []string
	for _, item := range state.view.Peek().visible {
		messages = append(messages, item.toast.Message)
	}
	return messages
}

func TestToastStateQueuesBeyondMaxVisible(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{MaxVisible: 2, Timeout: time.Second})
	state.Info("one")
	state.Info("two")
	state.Info("three")

	assert.Equal(t, []string{"one", "two"}, toastMessages(state))
	assert.Equal(t, 1, state.view.Peek().waiting)
	assert.Equal(t, 2, clock.running(), "a waiting toast's timeout has not started")

	clock.advance(time.Second)
	assert.Equal(t, []string{"three"}, toastMessages(state), "both shown toasts expire and the waiting one is promoted")
	assert.Zero(t, state.view.Peek().waiting)

	clock.advance(999 * time.Millisecond)
	assert.Equal(t, []string{"three"}, toastMessages(state), "a promoted toast gets its full timeout")
	clock.advance(time.Millisecond)
	assert.Empty(t, toastMessages(state))
	assert.Zero(t, state.Len())
}

func TestToastStateTimeouts(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{Timeout: time.Second})
	state.Notify(Toast{Message: "long", Timeout: 3 * time.Second})
	state.Notify(Toast{Message: "sticky", Timeout: -1})
	state.Info("default")

	clock.advance(time.Second)
	assert.Equal(t, []string{"long", "sticky"}, toastMessages(state))
	clock.advance(2 * time.Second)
	assert.Equal(t, []string{"sticky"}, toastMessages(state))
	clock.advance(time.Hour)
	assert.Equal(t, []string{"sticky"}, toastMessages(state), "a negative timeout never expires")
}

func TestToastStatePauseKeepsRemainingTime(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{Timeout: 4 * time.Second})
	id := state.Info("held")

	clock.advance(3 * time.Second)
	state.setPaused(id, true)
	assert.True(t, state.view.Peek().visible[0].paused)
	clock.advance(time.Hour)
	assert.Equal(t, []string{"held"}, toastMessages(state), "a paused toast does not expire")

	state.setPaused(id, false)
	clock.advance(999 * time.Millisecond)
	assert.Equal(t, []string{"held"}, toastMessages(state), "only the unexpired second remains")
	clock.advance(time.Millisecond)
	assert.Empty(t, toastMessages(state))
}

func TestToastStateIgnoresTimerFiringDuringPause(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{Timeout: time.Second})
	id := state.Info("held")
	stale := clock.timers[0]
	state.setPaused(id, true)

	// time.Timer.Stop cannot recall a callback that has already started.
	stale.fn()
	assert.Equal(t, []string{"held"}, toastMessages(state))
}

func TestToastStateDismissAndClear(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{MaxVisible: 1})
	first := state.Info("first")
	state.Info("second")
	waiting := state.Info("third")

	state.Dismiss(waiting)
	assert.Equal(t, []string{"first"}, toastMessages(state))
	assert.Equal(t, 1, state.view.Peek().waiting)

	state.Dismiss(first)
	assert.Equal(t, []string{"second"}, toastMessages(state))
	assert.Equal(t, 1, clock.running(), "the dismissed toast's timer is stopped")

	state.Dismiss(first)
	state.Clear()
	assert.Empty(t, toastMessages(state))
	assert.Zero(t, clock.running())
}

func TestToastStateNotifyFromGoroutines(t *testing.T) {
	state := NewToastState(ToastOptions{MaxVisible: 4, Timeout: time.Millisecond})
	root := Toasts{State: state}
	renderer := NewRenderer(reactivityScreen{uv.NewBuffer(60, 20)}, 60, 20, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(root)

	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 25 {
				id := state.Notify(Toast{Message: fmt.Sprintf("%d/%d", worker, i), Severity: ToastSeverity(i % 4)})
				if i%3 == 0 {
					state.Dismiss(id)
				}
			}
		}()
	}
	for range 50 {
		renderer.Update(root)
	}
	wg.Wait()
	require.Eventually(t, func() bool { return state.Len() == 0 }, 2*time.Second, time.Millisecond)
	renderer.Update(root)
	assert.NotContains(t, renderer.ScreenText(), "/")
}

func TestToastsSnapshot(t *testing.T) {
	background := Column{Width: Flex(1), Height: Flex(1), Children: []Widget{
		Text{Content: "Project files"},
		Text{Content: "  main.go"},
		Text{Content: "  toast.go"},
	}}
	scene := func(toasts Toasts) Widget {
		return Stack{Width: Flex(1), Height: Flex(1), Children: []Widget{background, toasts}}
	}

	state, _ := newTestToastState(ToastOptions{})
	state.Notify(Toast{Title: "Saved", Message: "Wrote 3 files to disk.", Severity: ToastSuccess})
	state.Notify(Toast{Message: "Sync is taking longer than usual, but will keep trying in the background.", Severity: ToastWarning})
	state.Error("Copy failed")
	AssertSnapshotNamed(t, "Toasts_severities", scene(Toasts{State: state}), 60, 16,
		"Three toasts stacked in the bottom-right corner, one cell in from the edges: a green-bordered "+
			"success toast with a bold green 'Saved' title over its message, a yellow warning whose message "+
			"wraps over two lines, and a red 'Copy failed' error. Each has its severity icon on the left.")

	queued, _ := newTestToastState(ToastOptions{MaxVisible: 2})
	queued.Info("Indexing workspace")
	queued.Info("Fetched 12 commits")
	queued.Info("Waiting one")
	queued.Info("Waiting two")
	AssertSnapshotNamed(t, "Toasts_waiting", scene(Toasts{State: queued, Position: FloatPositionTopLeft, Width: 30}), 60, 12,
		"Two blue info toasts 30 cells wide in the top-left corner, then a muted right-aligned '+2 more' "+
			"line beneath them for the toasts still waiting.")

	paused, _ := newTestToastState(ToastOptions{})
	id := paused.Info("Hovered toast")
	paused.Info("Other toast")
	paused.setPaused(id, true)
	AssertSnapshotNamed(t, "Toasts_paused", scene(Toasts{State: paused, Position: FloatPositionTopCenter}), 60, 10,
		"Two info toasts at the top centre. The first, hovered and paused, has the lighter surface-hover background.")
}

func TestToastsClickDismissesAndHoverPauses(t *testing.T) {
	state, clock := newTestToastState(ToastOptions{Timeout: time.Second})
	state.Info("first")
	state.Info("second")
	root := Column{Width: Flex(1), Height: Flex(1), Children: []Widget{
		Text{Content: "body"},
		Toasts{ID: "toasts", State: state, Position: FloatPositionTopLeft, Width: 20},
	}}
	router, renderer := renderForMouse(root, 40, 10)
	// first occupies rows 1-3 and second rows 5-7, starting at column 1.
	require.Equal(t, "toasts-toast-1", hitID(renderer, 2, 2))
	require.Equal(t, "toasts-toast-2", hitID(renderer, 2, 6))

	router.motion(uv.MouseMotionEvent{X: 2, Y: 6}, 0.5, 0.5)
	clock.advance(time.Second)
	assert.Equal(t, []string{"second"}, toastMessages(state), "the hovered toast is held while the other expires")
	renderer.Render(root)

	router.press(uv.MouseClickEvent{X: 2, Y: 2, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
	router.release(uv.MouseReleaseEvent{X: 2, Y: 2, Button: uv.MouseLeft}, 0.5, 0.5)
	assert.Empty(t, toastMessages(state), "clicking the toast now at the top dismisses it")
	renderer.Render(root)
	assert.NotContains(t, hitID(renderer, 2, 2), "toast")
}

type reactivityToastScene struct {
	toasts  *ToastState
	clock   *fakeToastClock
	counter Signal[int]
}

func (s *reactivityToastScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "counter", build: func(BuildContext) Widget {
			return Text{Content: fmt.Sprintf("count %d", s.counter.Get())}
		}},
		Toasts{ID: "toasts", State: s.toasts},
	}}
}

func TestReactivityToasts(t *testing.T) {
	sequence := newReactivitySequence(t, 50, 14, func() *reactivityToastScene {
		state, clock := newTestToastState(ToastOptions{MaxVisible: 2, Timeout: time.Second})
		return &reactivityToastScene{toasts: state, clock: clock, counter: NewSignal(0)}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Notify", func(s *reactivityToastScene) { s.toasts.Success("Saved") })
	sequence.frame("Stack a second", func(s *reactivityToastScene) {
		s.toasts.Notify(Toast{Title: "Copy failed", Message: "Clipboard unavailable", Severity: ToastError})
	})
	sequence.frame("Queue a third", func(s *reactivityToastScene) { s.toasts.Warning("Queued") })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "+1 more")

	work := sequence.frame("Unrelated change keeps toasts", func(s *reactivityToastScene) { s.counter.Set(1) })
	require.Equal(t, 1, work.BuildCount, "only the counter rebuilds; the toast overlay is reused")
	require.Contains(t, sequence.actual.renderer.ScreenText(), "Saved")

	sequence.frame("Pause", func(s *reactivityToastScene) { s.toasts.setPaused(2, true) })
	sequence.frame("Expire unpaused toast", func(s *reactivityToastScene) { s.clock.advance(time.Second) })
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Saved")
	require.Contains(t, sequence.actual.renderer.ScreenText(), "Queued")
	sequence.frame("Dismiss", func(s *reactivityToastScene) { s.toasts.Dismiss(2) })
	sequence.frame("Clear", func(s *reactivityToastScene) { s.toasts.Clear() })
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Queued")
	sequence.frame("Unrelated change after clear", func(s *reactivityToastScene) { s.counter.Set(2) })
}
