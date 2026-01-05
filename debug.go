package terma

import (
	"fmt"
	"sync"
	"time"
)

// DebugMetrics holds performance and diagnostic information about the application.
type DebugMetrics struct {
	LastBuildDuration  time.Duration
	WidgetCount        int // Visible widgets (registered in hit-test registry)
	TotalWidgetCount   int // All widgets including those scrolled out of view
	FocusedWidgetID    string
	FocusedWidgetType  string
	FrameRate          float64
}

// Debugger provides real-time performance metrics and diagnostic information.
type Debugger struct {
	metrics      DebugMetrics
	visible      Signal[bool]
	frameHistory []time.Time
	historySize  int
	mu           sync.RWMutex
}

var globalDebugger *Debugger

// InitDebug initializes the global debugger.
// Call this at the start of your application to enable debug overlay.
func InitDebug() {
	globalDebugger = &Debugger{
		visible:      NewSignal(false),
		frameHistory: make([]time.Time, 0, 30),
		historySize:  30,
	}
}

// Toggle toggles the visibility of the debug overlay.
func (d *Debugger) Toggle() {
	d.visible.Update(func(v bool) bool { return !v })
}

// RecordFrame records metrics for a completed frame.
func (d *Debugger) RecordFrame(buildTime time.Duration, widgetCount, totalWidgetCount int, focusedID string, focused Focusable) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.metrics.LastBuildDuration = buildTime
	d.metrics.WidgetCount = widgetCount
	d.metrics.TotalWidgetCount = totalWidgetCount
	d.metrics.FocusedWidgetID = focusedID

	// Update focused widget type
	if focused != nil {
		d.metrics.FocusedWidgetType = fmt.Sprintf("%T", focused)
	} else {
		d.metrics.FocusedWidgetType = "none"
	}

	// Update frame history for FPS calculation
	d.frameHistory = append(d.frameHistory, now)
	if len(d.frameHistory) > d.historySize {
		d.frameHistory = d.frameHistory[1:]
	}

	// Calculate FPS over history window
	if len(d.frameHistory) > 1 {
		elapsed := d.frameHistory[len(d.frameHistory)-1].Sub(d.frameHistory[0])
		d.metrics.FrameRate = float64(len(d.frameHistory)-1) / elapsed.Seconds()
	}
}

// WrapRoot wraps the user's root widget with the debug overlay.
func (d *Debugger) WrapRoot(root Widget) Widget {
	return &debugWrapper{
		root:         root,
		debugOverlay: d.createOverlay(),
	}
}

// createOverlay creates the debug overlay widget.
func (d *Debugger) createOverlay() Widget {
	return &DebugOverlay{
		visible: d.visible,
		metrics: &d.metrics,
		mu:      &d.mu,
	}
}

// debugWrapper wraps the root widget with the debug overlay.
// It delegates KeybindProvider and KeyHandler to the wrapped root
// so that root-level keybindings still work when debug mode is enabled.
type debugWrapper struct {
	root         Widget
	debugOverlay Widget
}

// Keybinds delegates to the wrapped root widget if it implements KeybindProvider.
// This ensures root-level keybindings are still accessible when debug mode wraps the root.
func (w *debugWrapper) Keybinds() []Keybind {
	if provider, ok := w.root.(KeybindProvider); ok {
		return provider.Keybinds()
	}
	return nil
}

// OnKey delegates to the wrapped root widget if it implements KeyHandler.
// This ensures root-level key handling still works when debug mode wraps the root.
func (w *debugWrapper) OnKey(event KeyEvent) bool {
	if handler, ok := w.root.(KeyHandler); ok {
		return handler.OnKey(event)
	}
	return false
}

func (w *debugWrapper) Build(ctx BuildContext) Widget {
	// Build the root widget first
	builtRoot := w.root.Build(ctx)

	// Build the debug overlay
	builtOverlay := w.debugOverlay.Build(ctx)

	return Column{
		Width:  Fr(1),
		Height: Fr(1),
		Children: []Widget{
			builtRoot,
			builtOverlay,
		},
	}
}

// DebugOverlay is the floating overlay widget that displays debug metrics.
type DebugOverlay struct {
	visible Signal[bool]
	metrics *DebugMetrics
	mu      *sync.RWMutex
}

func (d *DebugOverlay) Build(ctx BuildContext) Widget {
	if !d.visible.Get() {
		return EmptyWidget{}
	}

	// Read metrics with lock
	d.mu.RLock()
	buildTime := d.metrics.LastBuildDuration
	visibleWidgetCount := d.metrics.WidgetCount
	totalWidgetCount := d.metrics.TotalWidgetCount
	frameRate := d.metrics.FrameRate
	focusedID := d.metrics.FocusedWidgetID
	focusedType := d.metrics.FocusedWidgetType
	d.mu.RUnlock()

	// Format build time
	buildTimeMs := float64(buildTime.Microseconds()) / 1000.0

	// Truncate focused ID if too long
	if len(focusedID) > 30 {
		focusedID = focusedID[:27] + "..."
	}
	if focusedID == "" {
		focusedID = "none"
	}

	// Truncate focused type if too long
	if len(focusedType) > 40 {
		focusedType = focusedType[:37] + "..."
	}

	theme := ctx.Theme()

	return Floating{
		Visible: true,
		Config: FloatConfig{
			Position: FloatPositionBottomRight,
			Offset:   Offset{X: -1, Y: -1},
			Modal:    false,
		},
		Child: Column{
			Style: Style{
				BackgroundColor: theme.Surface,
				ForegroundColor: theme.Text,
				Border:          Border{Style: BorderRounded, Color: theme.Primary},
				Padding:         EdgeInsetsAll(1),
			},
			Spacing: 0,
			Children: []Widget{
				Text{
					Content: "DEBUG METRICS",
					Style: Style{
						ForegroundColor: theme.Accent,
					},
				},
				Text{Content: fmt.Sprintf("Build: %.2fms", buildTimeMs)},
				Text{Content: fmt.Sprintf("Widgets: %d/%d", visibleWidgetCount, totalWidgetCount)},
				Text{Content: fmt.Sprintf("FPS: %.1f", frameRate)},
				Text{Content: fmt.Sprintf("Focus: %s", focusedID)},
				Text{
					Content: fmt.Sprintf("Type: %s", focusedType),
					Style: Style{
						ForegroundColor: theme.TextMuted,
					},
				},
			},
		},
	}
}
