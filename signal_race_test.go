package terma

import (
	"context"
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

type signalRaceApp struct {
	count Signal[int]
}

type signalRaceLabel struct {
	count Signal[int]
}

func (l signalRaceLabel) Build(ctx BuildContext) Widget {
	return Text{Content: fmt.Sprintf("count %d", l.count.Get())}
}

func (a *signalRaceApp) Build(ctx BuildContext) Widget {
	even := Select(a.count, func(n int) bool { return n%2 == 0 })
	return Column{Children: []Widget{
		ShowWhen(even, Text{Content: "even"}),
		Row{Children: []Widget{signalRaceLabel{count: a.count}}},
	}}
}

// Signals are documented as safe to update from any goroutine, such as a
// Task's worker, while the event loop renders.
func TestSignalUpdateFromBackgroundGoroutineWhileRendering(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setAppRuntimeState(ctx, newDispatchQueue())
	defer clearAppRuntimeState()

	app := &signalRaceApp{count: NewSignal(0)}
	buf := uv.NewBuffer(20, 4)
	renderer := NewRenderer(buf, 20, 4, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Update(app)

	const updates = 500
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range updates {
			app.count.Update(func(n int) int { return n + 1 })
		}
	}()
	for running := true; running; {
		select {
		case <-done:
			running = false
		default:
		}
		drainPendingDispatches()
		renderer.Update(app)
	}
	drainPendingDispatches()
	renderer.Update(app)

	if got := app.count.Peek(); got != updates {
		t.Fatalf("count = %d, want %d", got, updates)
	}
	if screen := buf.String(); !strings.Contains(screen, fmt.Sprintf("count %d", updates)) || !strings.Contains(screen, "even") {
		t.Fatalf("final frame missed the last update:\n%s", screen)
	}
}

type signalReadPausingWidget struct {
	inBuild chan struct{}
	resume  chan struct{}
}

func (w signalReadPausingWidget) Build(ctx BuildContext) Widget {
	close(w.inBuild)
	<-w.resume
	return Text{Content: "building"}
}

// A goroutine reading a signal while the renderer builds a widget must not
// subscribe that widget to the signal.
func TestSignalReadFromBackgroundGoroutineDuringBuildDoesNotSubscribe(t *testing.T) {
	read := NewSignal(0)
	selected := NewSignal(0)
	widget := signalReadPausingWidget{inBuild: make(chan struct{}), resume: make(chan struct{})}
	renderer := NewRenderer(uv.NewBuffer(20, 4), 20, 4, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))

	rendered := make(chan struct{})
	go func() {
		defer close(rendered)
		renderer.Update(widget)
	}()
	<-widget.inBuild
	_ = read.Get()
	_ = Select(selected, func(n int) bool { return n > 0 })
	close(widget.resume)
	<-rendered

	read.core.mu.Lock()
	listeners := len(read.core.listeners)
	read.core.mu.Unlock()
	selected.core.mu.Lock()
	selectors := len(selected.core.selectors.byNode)
	selected.core.mu.Unlock()
	if listeners != 0 || selectors != 0 {
		t.Fatalf("background reads subscribed the widget being built: %d listener(s), %d selector(s)", listeners, selectors)
	}
}
