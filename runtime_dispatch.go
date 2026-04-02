package terma

import (
	"context"
	"sync"
)

type dispatchQueue struct {
	mu      sync.Mutex
	pending []func()
}

func newDispatchQueue() *dispatchQueue {
	return &dispatchQueue{}
}

func (q *dispatchQueue) enqueue(fn func()) {
	if q == nil || fn == nil {
		return
	}
	q.mu.Lock()
	q.pending = append(q.pending, fn)
	q.mu.Unlock()
}

func (q *dispatchQueue) drain() {
	if q == nil {
		return
	}
	for {
		q.mu.Lock()
		if len(q.pending) == 0 {
			q.mu.Unlock()
			return
		}
		pending := q.pending
		q.pending = nil
		q.mu.Unlock()

		for _, fn := range pending {
			if fn != nil {
				fn()
			}
		}
	}
}

var (
	appRuntimeMu     sync.RWMutex
	appDispatchQueue *dispatchQueue
	appLifecycleCtx  context.Context
)

func setAppRuntimeState(ctx context.Context, queue *dispatchQueue) {
	appRuntimeMu.Lock()
	appLifecycleCtx = ctx
	appDispatchQueue = queue
	appRuntimeMu.Unlock()
}

func clearAppRuntimeState() {
	appRuntimeMu.Lock()
	appLifecycleCtx = nil
	appDispatchQueue = nil
	appRuntimeMu.Unlock()
}

func currentAppContext() context.Context {
	appRuntimeMu.RLock()
	ctx := appLifecycleCtx
	appRuntimeMu.RUnlock()
	return ctx
}

func dispatchIfRunning(fn func()) bool {
	if fn == nil {
		return false
	}

	appRuntimeMu.RLock()
	ctx := appLifecycleCtx
	queue := appDispatchQueue
	appRuntimeMu.RUnlock()

	if ctx == nil || ctx.Err() != nil || queue == nil {
		return false
	}

	queue.enqueue(fn)
	scheduleRender()
	return true
}

func drainPendingDispatches() {
	appRuntimeMu.RLock()
	queue := appDispatchQueue
	appRuntimeMu.RUnlock()
	if queue != nil {
		queue.drain()
	}
}

// Dispatch schedules fn to run on the app/event-loop goroutine before the next
// rendered frame. If no app is running, fn runs immediately.
func Dispatch(fn func()) {
	if fn == nil {
		return
	}
	if dispatchIfRunning(fn) {
		return
	}
	fn()
}
