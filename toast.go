package terma

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// ToastSeverity selects a toast's icon and theme color.
type ToastSeverity int

const (
	ToastInfo ToastSeverity = iota
	ToastSuccess
	ToastWarning
	ToastError
)

func (s ToastSeverity) icon() string {
	switch s {
	case ToastSuccess:
		return "✓"
	case ToastWarning:
		return "!"
	case ToastError:
		return "✗"
	default:
		return "i"
	}
}

func (s ToastSeverity) color(theme ThemeData) Color {
	switch s {
	case ToastSuccess:
		return theme.Success
	case ToastWarning:
		return theme.Warning
	case ToastError:
		return theme.Error
	default:
		return theme.Info
	}
}

// Toast is one notification.
type Toast struct {
	Title    string        // Optional bold heading
	Message  string        // Body text; wraps to the toast's width
	Severity ToastSeverity // Icon and color (default ToastInfo)
	// Timeout is how long the toast stays once it is on screen. Zero uses the
	// state's default; a negative value keeps it until it is dismissed.
	Timeout time.Duration
}

// ToastID identifies a toast returned by ToastState.Notify.
type ToastID uint64

// ToastOptions configures a ToastState.
type ToastOptions struct {
	MaxVisible int           // Toasts on screen at once (default 3); the rest wait
	Timeout    time.Duration // Default timeout (default 4s)
}

// ToastState holds an app's notifications. Every method is safe to call from
// any goroutine. Create one with NewToastState, show it with a Toasts widget,
// and call Notify (or Info, Success, Warning, Error) to add a toast.
type ToastState struct {
	mu         sync.Mutex
	maxVisible int
	timeout    time.Duration
	nextID     ToastID
	queue      []*toastEntry // arrival order; the first maxVisible are on screen
	view       AnySignal[toastView]

	now   func() time.Time
	after func(time.Duration, func()) (stop func() bool)
}

type toastEntry struct {
	id    ToastID
	toast Toast
	// remaining is the unexpired timeout; negative means sticky. It is only
	// current while the timer is not running.
	remaining time.Duration
	deadline  time.Time
	stop      func() bool
	// generation invalidates a timer that fired while being stopped.
	generation int
	paused     bool
}

// toastView is the immutable snapshot the Toasts widget builds from.
type toastView struct {
	visible []toastItem
	waiting int
}

type toastItem struct {
	id     ToastID
	toast  Toast
	paused bool
}

// NewToastState creates an empty toast queue.
func NewToastState(options ToastOptions) *ToastState {
	if options.MaxVisible <= 0 {
		options.MaxVisible = 3
	}
	if options.Timeout <= 0 {
		options.Timeout = 4 * time.Second
	}
	return &ToastState{
		maxVisible: options.MaxVisible,
		timeout:    options.Timeout,
		view:       NewAnySignal(toastView{}),
		now:        time.Now,
		after: func(d time.Duration, fn func()) func() bool {
			return time.AfterFunc(d, fn).Stop
		},
	}
}

// Notify adds a toast and returns its ID. Its timeout starts once it is on
// screen, so toasts queued behind MaxVisible are not missed.
func (s *ToastState) Notify(toast Toast) ToastID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	remaining := toast.Timeout
	if remaining == 0 {
		remaining = s.timeout
	}
	s.queue = append(s.queue, &toastEntry{id: s.nextID, toast: toast, remaining: remaining})
	s.syncLocked()
	return s.nextID
}

// Info adds an info toast with a message.
func (s *ToastState) Info(message string) ToastID {
	return s.Notify(Toast{Message: message, Severity: ToastInfo})
}

// Success adds a success toast with a message.
func (s *ToastState) Success(message string) ToastID {
	return s.Notify(Toast{Message: message, Severity: ToastSuccess})
}

// Warning adds a warning toast with a message.
func (s *ToastState) Warning(message string) ToastID {
	return s.Notify(Toast{Message: message, Severity: ToastWarning})
}

// Error adds an error toast with a message.
func (s *ToastState) Error(message string) ToastID {
	return s.Notify(Toast{Message: message, Severity: ToastError})
}

// Dismiss removes a toast, shown or waiting. Unknown IDs are ignored.
func (s *ToastState) Dismiss(id ToastID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.indexLocked(id)
	if i < 0 {
		return
	}
	s.stopLocked(s.queue[i])
	s.queue = slices.Delete(s.queue, i, i+1)
	s.syncLocked()
}

// Clear removes every toast.
func (s *ToastState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.queue {
		s.stopLocked(entry)
	}
	s.queue = nil
	s.syncLocked()
}

// Len returns the number of toasts, shown or waiting.
func (s *ToastState) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}

// setPaused holds a shown toast's timeout while the pointer is over it.
func (s *ToastState) setPaused(id ToastID, paused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.indexLocked(id)
	if i < 0 || s.queue[i].paused == paused {
		return
	}
	entry := s.queue[i]
	entry.paused = paused
	if paused {
		s.stopLocked(entry)
	}
	s.syncLocked()
}

func (s *ToastState) expire(id ToastID, generation int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.indexLocked(id)
	if i < 0 || s.queue[i].generation != generation {
		return
	}
	s.queue = slices.Delete(s.queue, i, i+1)
	s.syncLocked()
}

func (s *ToastState) indexLocked(id ToastID) int {
	return slices.IndexFunc(s.queue, func(entry *toastEntry) bool { return entry.id == id })
}

// stopLocked cancels the entry's timer and banks its unexpired time.
func (s *ToastState) stopLocked(entry *toastEntry) {
	if entry.stop == nil {
		return
	}
	entry.stop()
	entry.stop = nil
	entry.generation++
	entry.remaining = max(entry.deadline.Sub(s.now()), 0)
}

// syncLocked runs timers for the shown, unpaused toasts and publishes a new
// snapshot. Publishing under the lock keeps snapshots from concurrent callers
// in order; Set only marks subscribers dirty.
func (s *ToastState) syncLocked() {
	shown := min(len(s.queue), s.maxVisible)
	view := toastView{visible: make([]toastItem, 0, shown), waiting: len(s.queue) - shown}
	for _, entry := range s.queue[:shown] {
		if entry.stop == nil && !entry.paused && entry.remaining >= 0 {
			id, generation := entry.id, entry.generation
			entry.deadline = s.now().Add(entry.remaining)
			entry.stop = s.after(entry.remaining, func() { s.expire(id, generation) })
		}
		view.visible = append(view.visible, toastItem{id: entry.id, toast: entry.toast, paused: entry.paused})
	}
	s.view.Set(view)
}

// Toasts shows a ToastState's notifications stacked in a corner of the screen,
// above the rest of the app. Place it anywhere in the tree; it takes no space.
// Clicking a toast dismisses it, and hovering one holds its timeout.
//
//	toasts := NewToastState(ToastOptions{})
//	Stack{Children: []Widget{body, Toasts{State: toasts}}}
//	toasts.Success("Saved")
type Toasts struct {
	ID    string
	State *ToastState // Required
	// Position is the screen corner or edge the stack sits against (default
	// FloatPositionBottomRight). FloatPositionCenter and FloatPositionAbsolute
	// are treated as the default.
	Position FloatPosition
	Width    int // Toast width in cells, including the border (default 40)
}

// WidgetID returns the widget's ID.
func (t Toasts) WidgetID() string {
	return t.ID
}

// Build registers the overlay while there are toasts to show.
func (t Toasts) Build(ctx BuildContext) Widget {
	if t.State == nil {
		return EmptyWidget{}
	}
	view := t.State.view.Get()
	if len(view.visible) == 0 {
		return EmptyWidget{}
	}
	theme := ctx.Theme()
	width := t.Width
	if width <= 0 {
		width = 40
	}
	prefix := t.ID
	if prefix == "" {
		prefix = ctx.AutoID()
	}

	cards := make([]Widget, 0, len(view.visible)+1)
	for _, item := range view.visible {
		cards = append(cards, t.card(theme, prefix, width, item))
	}
	if view.waiting > 0 {
		cards = append(cards, Text{
			Content:   fmt.Sprintf("+%d more", view.waiting),
			TextAlign: TextAlignRight,
			Style:     Style{Width: Cells(width), ForegroundColor: theme.TextMuted},
		})
	}

	position, offset := t.placement()
	Floating{
		Visible: true,
		Config:  FloatConfig{Position: position, Offset: offset},
		Child:   Column{Spacing: 1, Children: cards},
	}.Build(ctx)
	return EmptyWidget{}
}

func (t Toasts) card(theme ThemeData, prefix string, width int, item toastItem) Widget {
	accent := item.toast.Severity.color(theme)
	background := theme.Surface
	if item.paused {
		background = theme.SurfaceHover
	}
	var lines []Widget
	if item.toast.Title != "" {
		lines = append(lines, Text{
			Content: item.toast.Title,
			Wrap:    WrapSoft,
			Style:   Style{ForegroundColor: accent, Bold: true},
		})
	}
	if item.toast.Message != "" {
		lines = append(lines, Text{Content: item.toast.Message, Wrap: WrapSoft})
	}

	state, id := t.State, item.id
	return Row{
		ID:      fmt.Sprintf("%s-toast-%d", prefix, id),
		Spacing: 1,
		Style: Style{
			Width:           Cells(max(width-4, 1)),
			BackgroundColor: background,
			ForegroundColor: theme.Text,
			Border:          RoundedBorder(accent),
			Padding:         EdgeInsetsXY(1, 0),
		},
		Click: func(MouseEvent) { state.Dismiss(id) },
		Hover: func(event HoverEvent) { state.setPaused(id, event.Type == HoverEnter) },
		Children: []Widget{
			Text{Content: item.toast.Severity.icon(), Style: Style{ForegroundColor: accent, Bold: true}},
			Column{Style: Style{Width: Flex(1)}, Children: lines},
		},
	}
}

// placement keeps the stack one cell in from the screen edges it sits on.
func (t Toasts) placement() (FloatPosition, Offset) {
	switch t.Position {
	case FloatPositionTopLeft:
		return t.Position, Offset{X: 1, Y: 1}
	case FloatPositionTopCenter:
		return t.Position, Offset{Y: 1}
	case FloatPositionTopRight:
		return t.Position, Offset{X: -1, Y: 1}
	case FloatPositionBottomLeft:
		return t.Position, Offset{X: 1, Y: -1}
	case FloatPositionBottomCenter:
		return t.Position, Offset{Y: -1}
	default:
		return FloatPositionBottomRight, Offset{X: -1, Y: -1}
	}
}
