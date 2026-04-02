package terma

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func installTestAppRuntime(t *testing.T) context.CancelFunc {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	oldRenderTrigger := renderTrigger
	renderTrigger = make(chan struct{}, 1)
	setAppRuntimeState(ctx, newDispatchQueue())

	t.Cleanup(func() {
		clearAppRuntimeState()
		renderTrigger = oldRenderTrigger
		cancel()
	})

	return cancel
}

func eventuallyDrain(t *testing.T, fn func() bool) {
	t.Helper()
	require.Eventually(t, func() bool {
		drainPendingDispatches()
		return fn()
	}, time.Second, 10*time.Millisecond)
}

func TestDispatchQueuesUntilDrained(t *testing.T) {
	installTestAppRuntime(t)

	var order []int
	Dispatch(func() { order = append(order, 1) })
	Dispatch(func() { order = append(order, 2) })

	require.Empty(t, order)

	select {
	case <-renderTrigger:
	default:
		t.Fatal("expected Dispatch to schedule a render")
	}

	drainPendingDispatches()
	require.Equal(t, []int{1, 2}, order)
}

func TestDispatchRunsImmediatelyWithoutApp(t *testing.T) {
	var called atomic.Bool
	Dispatch(func() {
		called.Store(true)
	})
	require.True(t, called.Load())
}

func TestTaskSuccessMarksRunningImmediately(t *testing.T) {
	installTestAppRuntime(t)

	task := NewTask[int]()
	release := make(chan struct{})

	task.Start(func(ctx context.Context) (int, error) {
		<-release
		return 42, nil
	})

	require.True(t, task.Running.Peek())
	require.Equal(t, TaskRunning, task.Phase.Peek())

	close(release)

	eventuallyDrain(t, func() bool {
		return task.Phase.Peek() == TaskSuccess
	})

	require.False(t, task.Running.Peek())
	require.Equal(t, 42, task.Value.Peek())
	require.NoError(t, task.Err.Peek())
}

func TestTaskError(t *testing.T) {
	installTestAppRuntime(t)

	task := NewTask[int]()
	expected := errors.New("boom")

	task.Start(func(ctx context.Context) (int, error) {
		return 0, expected
	})

	eventuallyDrain(t, func() bool {
		return task.Phase.Peek() == TaskError
	})

	require.False(t, task.Running.Peek())
	require.ErrorIs(t, task.Err.Peek(), expected)
}

func TestTaskCancelIgnoresLateCompletion(t *testing.T) {
	installTestAppRuntime(t)

	task := NewTask[int]()
	release := make(chan struct{})

	task.Start(func(ctx context.Context) (int, error) {
		<-release
		return 99, nil
	})

	task.Cancel()
	require.Equal(t, TaskCancelled, task.Phase.Peek())
	require.False(t, task.Running.Peek())

	close(release)
	time.Sleep(20 * time.Millisecond)
	drainPendingDispatches()

	require.Equal(t, TaskCancelled, task.Phase.Peek())
	require.Equal(t, 0, task.Value.Peek())
}

func TestTaskStartSuppressesStaleResults(t *testing.T) {
	installTestAppRuntime(t)

	task := NewTask[int]()
	releaseFirst := make(chan struct{})

	task.Start(func(ctx context.Context) (int, error) {
		<-releaseFirst
		return 1, nil
	})

	task.Start(func(ctx context.Context) (int, error) {
		return 2, nil
	})

	eventuallyDrain(t, func() bool {
		return task.Phase.Peek() == TaskSuccess && task.Value.Peek() == 2
	})

	close(releaseFirst)
	time.Sleep(20 * time.Millisecond)
	drainPendingDispatches()

	require.Equal(t, TaskSuccess, task.Phase.Peek())
	require.Equal(t, 2, task.Value.Peek())
	require.NoError(t, task.Err.Peek())
}

func TestTaskUsesAppLifecycleContext(t *testing.T) {
	cancel := installTestAppRuntime(t)

	task := NewTask[int]()
	cancelled := make(chan struct{})
	task.Start(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		close(cancelled)
		return 0, ctx.Err()
	})

	cancel()

	require.Eventually(t, func() bool {
		select {
		case <-cancelled:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}
