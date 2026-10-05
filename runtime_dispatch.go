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
	appRuntimeMu          sync.RWMutex
	appDispatchQueue      *dispatchQueue
	appLifecycleCtx       context.Context
	headlessDispatchQueue *dispatchQueue
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

func appShuttingDown() bool {
	ctx := currentAppContext()
	return ctx != nil && ctx.Err() != nil
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
// rendered frame. During a headless Renderer render, fn runs after the frame
// finishes, and the renderer draws again before returning. Outside an app or a
// headless render, fn runs immediately. Between Quit and Run returning, fn is
// dropped: no frame will show its effects, and running it on the caller's
// goroutine would race the app's last frame.
func Dispatch(fn func()) {
	if fn == nil {
		return
	}
	if dispatchIfRunning(fn) || appShuttingDown() {
		return
	}
	appRuntimeMu.RLock()
	queue := headlessDispatchQueue
	appRuntimeMu.RUnlock()
	if queue != nil {
		queue.enqueue(fn)
		return
	}
	fn()
}

// beginHeadlessDispatch gives only the outermost headless render ownership of
// frame-boundary work. A running app keeps its queue and event-loop ownership.
func beginHeadlessDispatch() (*dispatchQueue, func()) {
	appRuntimeMu.Lock()
	defer appRuntimeMu.Unlock()
	if (appLifecycleCtx != nil && appLifecycleCtx.Err() == nil && appDispatchQueue != nil) || headlessDispatchQueue != nil {
		return nil, nil
	}
	queue := newDispatchQueue()
	headlessDispatchQueue = queue
	return queue, func() {
		appRuntimeMu.Lock()
		headlessDispatchQueue = nil
		appRuntimeMu.Unlock()
	}
}

// takePending drains one batch only. Work dispatched by a callback is handled
// at the next boundary, so self-dispatch cannot trap the headless caller here.
func (q *dispatchQueue) takePending() []func() {
	q.mu.Lock()
	pending := q.pending
	q.pending = nil
	q.mu.Unlock()
	return pending
}
