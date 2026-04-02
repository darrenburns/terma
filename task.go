package terma

import (
	"context"
	"sync"
)

// TaskPhase describes the current lifecycle phase of a Task.
type TaskPhase int

const (
	TaskIdle TaskPhase = iota
	TaskRunning
	TaskSuccess
	TaskError
	TaskCancelled
)

// Task runs one-shot background work and exposes its state through signals.
// Start cancels any existing run, executes fn in a goroutine, and applies the
// final task state on the UI thread when an app is running.
type Task[T any] struct {
	Phase   Signal[TaskPhase]
	Running Signal[bool]
	Value   AnySignal[T]
	Err     AnySignal[error]

	mu         sync.Mutex
	generation uint64
	cancel     context.CancelFunc
}

// NewTask creates a new idle task.
func NewTask[T any]() *Task[T] {
	var zero T
	return &Task[T]{
		Phase:   NewSignal(TaskIdle),
		Running: NewSignal(false),
		Value:   NewAnySignal(zero),
		Err:     NewAnySignal[error](nil),
	}
}

// Start begins a new run. Any in-flight run is cancelled and ignored if it
// later completes.
func (t *Task[T]) Start(fn func(context.Context) (T, error)) {
	if t == nil || fn == nil {
		return
	}

	baseCtx := currentAppContext()
	appBound := baseCtx != nil
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	ctx, cancel := context.WithCancel(baseCtx)

	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
	}
	t.generation++
	generation := t.generation
	t.cancel = cancel
	t.mu.Unlock()

	t.Err.Set(nil)
	t.Phase.Set(TaskRunning)
	t.Running.Set(true)

	go func() {
		value, err := fn(ctx)
		t.complete(generation, value, err, ctx.Err(), appBound)
	}()
}

// Cancel stops the current run, if any, and marks the task as cancelled.
func (t *Task[T]) Cancel() {
	if t == nil {
		return
	}

	t.mu.Lock()
	if t.cancel == nil {
		t.mu.Unlock()
		return
	}
	t.cancel()
	t.cancel = nil
	t.generation++
	t.mu.Unlock()

	t.Err.Set(nil)
	t.Phase.Set(TaskCancelled)
	t.Running.Set(false)
}

func (t *Task[T]) complete(generation uint64, value T, err error, ctxErr error, appBound bool) {
	apply := func() {
		t.mu.Lock()
		if generation != t.generation {
			t.mu.Unlock()
			return
		}
		t.cancel = nil
		t.mu.Unlock()

		if ctxErr != nil {
			t.Err.Set(nil)
			t.Phase.Set(TaskCancelled)
			t.Running.Set(false)
			return
		}

		if err != nil {
			t.Err.Set(err)
			t.Phase.Set(TaskError)
			t.Running.Set(false)
			return
		}

		t.Value.Set(value)
		t.Err.Set(nil)
		t.Phase.Set(TaskSuccess)
		t.Running.Set(false)
	}

	if dispatchIfRunning(apply) {
		return
	}

	if appBound {
		return
	}

	// Outside a running app, apply synchronously so the helper remains usable in
	// tests and non-UI contexts.
	apply()
}
