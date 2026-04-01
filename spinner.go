package terma

import (
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
)

// SpinnerStyle defines the visual appearance of a spinner.
type SpinnerStyle struct {
	Frames    []string
	FrameTime time.Duration
}

// Built-in spinner styles.
var (
	// SpinnerDots is the classic braille dots spinner.
	SpinnerDots = SpinnerStyle{
		Frames:    []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FrameTime: 80 * time.Millisecond,
	}

	// SpinnerLine is a simple rotating line spinner.
	SpinnerLine = SpinnerStyle{
		Frames:    []string{"-", "\\", "|", "/"},
		FrameTime: 100 * time.Millisecond,
	}

	// SpinnerCircle is a quarter-filled circle spinner.
	SpinnerCircle = SpinnerStyle{
		Frames:    []string{"◐", "◓", "◑", "◒"},
		FrameTime: 120 * time.Millisecond,
	}

	// SpinnerBounce is a bouncing dot spinner.
	SpinnerBounce = SpinnerStyle{
		Frames:    []string{"⠁", "⠂", "⠄", "⠂"},
		FrameTime: 120 * time.Millisecond,
	}

	// SpinnerArrow is a rotating arrow spinner.
	SpinnerArrow = SpinnerStyle{
		Frames:    []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		FrameTime: 100 * time.Millisecond,
	}

	// SpinnerBraille is a detailed braille pattern spinner.
	SpinnerBraille = SpinnerStyle{
		Frames:    []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FrameTime: 80 * time.Millisecond,
	}

	// SpinnerGrow is a growing bar spinner.
	SpinnerGrow = SpinnerStyle{
		Frames:    []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"},
		FrameTime: 80 * time.Millisecond,
	}

	// SpinnerPulse is a pulsing block spinner.
	SpinnerPulse = SpinnerStyle{
		Frames:    []string{"█", "▓", "▒", "░", "▒", "▓"},
		FrameTime: 100 * time.Millisecond,
	}

	// SpinnerClock is a clock-like spinner.
	SpinnerClock = SpinnerStyle{
		Frames:    []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"},
		FrameTime: 100 * time.Millisecond,
	}

	// SpinnerMoon is a moon phase spinner.
	SpinnerMoon = SpinnerStyle{
		Frames:    []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
		FrameTime: 120 * time.Millisecond,
	}

	// SpinnerDotsBounce is a horizontal bouncing dots spinner.
	SpinnerDotsBounce = SpinnerStyle{
		Frames:    []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"},
		FrameTime: 100 * time.Millisecond,
	}
)

// SpinnerState holds the animation state for a Spinner.
// Create with NewSpinnerState and pass to the Spinner widget.
type SpinnerState struct {
	animation *FrameAnimation[string]
}

// NewSpinnerState creates a new spinner state with the given style.
func NewSpinnerState(style SpinnerStyle) *SpinnerState {
	anim := NewFrameAnimation(FrameAnimationConfig[string]{
		Frames:    style.Frames,
		FrameTime: style.FrameTime,
		Loop:      true,
	})
	return &SpinnerState{animation: anim}
}

// Start begins the spinner animation.
func (s *SpinnerState) Start() {
	s.animation.Start()
}

// Stop halts the spinner animation.
func (s *SpinnerState) Stop() {
	s.animation.Stop()
}

// IsRunning returns true if the spinner is currently animating.
func (s *SpinnerState) IsRunning() bool {
	return s.animation.IsRunning()
}

// Frame returns the current animation frame.
func (s *SpinnerState) Frame() string {
	return s.animation.Value().Get()
}

// Spinner displays an animated loading indicator.
// Use pure composition to add labels:
//
//	Row{Children: []Widget{
//	    Spinner{State: spinnerState},
//	    Text{Content: " Loading..."},
//	}}
type Spinner struct {
	ID     string        // Optional unique identifier
	State  *SpinnerState // Required - holds animation state
	Width  Dimension     // Deprecated: use Style.Width
	Height Dimension     // Deprecated: use Style.Height
	Style  Style         // Optional styling
}

// WidgetID returns the spinner's unique identifier.
func (s Spinner) WidgetID() string {
	return s.ID
}

// GetContentDimensions returns the dimensions.
func (s Spinner) GetContentDimensions() (width, height Dimension) {
	dims := s.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	return width, height
}

// GetStyle returns the style.
func (s Spinner) GetStyle() Style {
	return s.Style
}

// Build returns self so spinner frames can update during paint without rebuilding
// the widget tree.
func (s Spinner) Build(ctx BuildContext) Widget {
	return s
}

func (s Spinner) resolvedStyle() Style {
	style := s.Style
	if style.Width.IsUnset() {
		style.Width = s.Width
	}
	if style.Height.IsUnset() {
		style.Height = s.Height
	}
	return style
}

func (s Spinner) widestFrame() string {
	if s.State == nil || s.State.animation == nil {
		return " "
	}
	frames := s.State.animation.frames
	if len(frames) == 0 {
		return " "
	}
	widest := frames[0]
	widestWidth := ansi.StringWidth(widest)
	for _, frame := range frames[1:] {
		if width := ansi.StringWidth(frame); width > widestWidth {
			widest = frame
			widestWidth = width
		}
	}
	return widest
}

func (s Spinner) currentFrame() string {
	if s.State == nil || s.State.animation == nil {
		return " "
	}
	return s.State.animation.Value().Get()
}

func (s Spinner) ContentWidthHint() int {
	return max(1, ansi.StringWidth(s.widestFrame()))
}

func (s Spinner) ContentHeightHint(width int) int {
	return 1
}

func (s Spinner) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	text := Text{
		Content: s.widestFrame(),
		Style:   s.resolvedStyle(),
	}
	return text.BuildLayoutNode(ctx)
}

func (s Spinner) Render(ctx *RenderContext) {
	Text{
		Content: s.currentFrame(),
		Style:   s.resolvedStyle(),
	}.Render(ctx)
}
