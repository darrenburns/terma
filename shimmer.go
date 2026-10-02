package terma

import (
	"math"
	"sync"
	"time"
)

// shimmerIdle is the phase of a stopped shimmer: no highlight is drawn.
const shimmerIdle = -1.0

// ShimmerState drives the highlight of every Shimmer that uses it.
// Create with NewShimmerState and pass to a Shimmer.
type ShimmerState struct {
	mu           sync.Mutex
	period       time.Duration
	elapsed      time.Duration
	running      bool
	pendingStart bool
	phase        Signal[float64]
	handle       *animationHandle
}

// NewShimmerState creates a shimmer whose highlight completes its path once per period.
// A non-positive period defaults to 1.5s.
func NewShimmerState(period time.Duration) *ShimmerState {
	if period <= 0 {
		period = 1500 * time.Millisecond
	}
	return &ShimmerState{period: period, phase: NewSignal(shimmerIdle)}
}

// Start begins the shimmer. If called before the app is running, the shimmer
// starts when it is first painted.
func (s *ShimmerState) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}
	s.running = true
	s.elapsed = 0
	s.phase.Set(0)
	if currentController != nil {
		s.handle = currentController.Register(s)
	} else {
		s.pendingStart = true
	}
}

// Stop halts the shimmer and removes the highlight.
func (s *ShimmerState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil && currentController != nil {
		currentController.Unregister(s.handle)
	}
	s.handle = nil
	s.pendingStart = false
	s.running = false
	s.phase.Set(shimmerIdle)
}

// IsRunning returns true if the shimmer is animating.
func (s *ShimmerState) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Advance implements the Animator interface.
func (s *ShimmerState) Advance(dt time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return false
	}
	s.elapsed = (s.elapsed + dt) % s.period
	s.phase.Set(float64(s.elapsed) / float64(s.period))
	return true
}

// currentPhase subscribes the widget being painted to the phase.
func (s *ShimmerState) currentPhase() float64 {
	s.mu.Lock()
	if s.pendingStart && currentController != nil {
		s.handle = currentController.Register(s)
		s.pendingStart = false
	}
	s.mu.Unlock()
	return s.phase.Get()
}

// ShimmerPath is the route a Shimmer's highlight travels.
type ShimmerPath int

const (
	// ShimmerSweep moves the highlight left to right. It enters fully off the
	// left edge and leaves fully off the right, so each cycle starts and ends
	// on plain base color. Suits text.
	ShimmerSweep ShimmerPath = iota
	// ShimmerPerimeter moves the highlight clockwise around the edge of the
	// region, wrapping seamlessly. Suits borders.
	ShimmerPerimeter
)

// Shimmer is a ColorProvider that blends a highlight band into a base color
// and moves it as its State runs. Use it anywhere a color is accepted:
//
//	state := NewShimmerState(0)
//	state.Start()
//	Text{Content: "Thinking...", Style: Style{ForegroundColor: Shimmer{
//	    State: state, Base: theme.TextMuted, Highlight: theme.Text,
//	}}}
//	Style{Border: RoundedBorder(Shimmer{
//	    State: state, Base: theme.Border, Highlight: theme.Accent, Path: ShimmerPerimeter,
//	})}
//
// The highlight moves on repaint only: no rebuild or relayout.
type Shimmer struct {
	State     *ShimmerState // Required - drives the highlight
	Base      ColorProvider // Required - color away from the band; may be a Gradient
	Highlight Color         // Required - color at the band's center
	BandWidth int           // Width of the band in cells (default 8)
	Path      ShimmerPath   // Route of the band (default ShimmerSweep)
}

// IsSet reports whether the base color is set. This implements ColorProvider.
func (s Shimmer) IsSet() bool {
	return s.Base != nil && s.Base.IsSet()
}

// ColorAt returns the color at (x, y). This implements ColorProvider.
func (s Shimmer) ColorAt(width, height, x, y int) Color {
	base := s.Base.ColorAt(width, height, x, y)
	if s.State == nil {
		return base
	}
	phase := s.State.currentPhase()
	if phase < 0 {
		return base
	}
	halfWidth := 4.0
	if s.BandWidth > 0 {
		halfWidth = float64(s.BandWidth) / 2
	}

	var distance float64
	switch s.Path {
	case ShimmerPerimeter:
		position, length, onEdge := perimeterPosition(width, height, x, y)
		if !onEdge || length <= 0 {
			return base
		}
		distance = math.Abs(position - phase*length)
		distance = min(distance, length-distance)
	default:
		center := -halfWidth + phase*(float64(width)+2*halfWidth)
		distance = math.Abs(float64(x) + 0.5 - center)
	}
	if distance >= halfWidth {
		return base
	}
	return base.Blend(s.Highlight, 0.5*(1+math.Cos(math.Pi*distance/halfWidth)))
}

// perimeterPosition returns how far clockwise from the top-left corner the
// edge cell (x, y) is, and the perimeter's total length. Vertical steps count
// as two cells because terminal cells are about twice as tall as they are
// wide, so the band moves at an even visual speed.
func perimeterPosition(width, height, x, y int) (position, length float64, onEdge bool) {
	w, h := float64(width-1), float64(height-1)
	length = 2*w + 4*h
	fx, fy := float64(x), float64(y)
	switch {
	case y == 0:
		return fx, length, true
	case x == width-1:
		return w + 2*fy, length, true
	case y == height-1:
		return w + 2*h + (w - fx), length, true
	case x == 0:
		return 2*w + 2*h + 2*(h-fy), length, true
	}
	return 0, length, false
}
