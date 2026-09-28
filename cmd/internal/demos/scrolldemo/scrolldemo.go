// Package scrolldemo demonstrates Scrollable containers driven by ScrollState.
package scrolldemo

import (
	"fmt"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "scroll",
	Title:       "Scroll",
	Description: "Nested scrollables: a list, text, a pinned log and wide content",
}

// ScrollDemo shows Scrollable containers driven by ScrollState: a page that
// scrolls and holds more scrollables, a list, wrapped text, a log pinned to
// its end, wide content that pans sideways, and a panel with scrolling
// disabled.
//
// Keys in the focused scrollable:
//
//	↑/↓ j/k         - scroll one line
//	PgUp/PgDn       - scroll a page (also ctrl+u / ctrl+d)
//	Home/End g/G    - jump to top / end (End re-pins the log)
//	←/→ h/l         - pan sideways (wide panel)
//	tab / shift+tab - move between scrollables
//
// App keys:
//
//	p - pause / resume the log feed
//	b - jump the log to its end and re-pin it
//	r - scroll everything back to the top
//	c - clear the log
//	t - cycle theme
type ScrollDemo struct {
	pageState *t.ScrollState
	listState *t.ScrollState
	textState *t.ScrollState
	logState  *t.ScrollState
	wideState *t.ScrollState
	offState  *t.ScrollState

	logLines   t.AnySignal[[]string]
	logCount   int // numbers log lines; keeps counting after a clear
	feed       *t.Animation[float64]
	feedPaused t.Signal[bool]

	// The feed only runs while the log is on screen. logShown is set each
	// time the log is built; if a whole feed cycle passes without that, the
	// demo is hidden (e.g. the gallery switched away), so the feed stops and
	// feedIdle is set until Build shows the demo again.
	logShown bool
	feedIdle bool
}

const (
	sidebarWidth    = 32
	topRowHeight    = 12
	middleRowHeight = 9
	disabledHeight  = 3
)

const loremText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.\n\nExcepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo."

var logEvents = []string{
	"GET /api/items 200",
	"cache hit items:page=2",
	"POST /api/items 201",
	"worker 3 picked up job",
	"GET /healthz 200",
	"slow query took 412ms",
	"DELETE /api/items/17 204",
	"worker 3 finished job",
}

// New creates the demo.
func New() demokit.Demo {
	d := &ScrollDemo{
		pageState:  t.NewScrollState(),
		listState:  t.NewScrollState(),
		textState:  t.NewScrollState(),
		logState:   t.NewScrollState(),
		wideState:  t.NewScrollState(),
		offState:   t.NewScrollState(),
		logLines:   t.NewAnySignal[[]string](nil),
		feedPaused: t.NewSignal(false),
	}
	// While the log is scrolled to its end, new lines keep it there.
	// Scrolling up unpins it; End (or b) pins it again.
	d.logState.PinToBottom = true
	for i := 0; i < 3; i++ {
		d.appendLog()
	}

	// A looping animation runs code on the UI goroutine at a steady pace:
	// each time it completes, add a log line and start again.
	d.feed = t.NewAnimation(t.AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: 600 * time.Millisecond,
		OnComplete: func() {
			d.appendLog()
			if !d.logShown {
				d.feedIdle = true
				return
			}
			d.logShown = false
			d.startFeed()
		},
	})
	d.logShown = true
	d.feed.Start()
	return d
}

func (d *ScrollDemo) InitialFocus() string { return "scroll-list" }

func (d *ScrollDemo) startFeed() {
	d.feed.Reset()
	d.feed.Start()
}

func (d *ScrollDemo) appendLog() {
	d.logCount++
	line := fmt.Sprintf("%04d  %s", d.logCount, logEvents[d.logCount%len(logEvents)])
	d.logLines.Update(func(lines []string) []string {
		return append(lines[:len(lines):len(lines)], line)
	})
}

// toggleFeed stops or restarts the feed. Stopping (rather than pausing)
// unregisters the animation, so nothing ticks while the feed is paused.
func (d *ScrollDemo) toggleFeed() {
	if d.feedPaused.Peek() {
		d.logShown = true
		d.startFeed()
	} else {
		d.feed.Stop()
	}
	d.feedPaused.Set(!d.feedPaused.Peek())
}

func (d *ScrollDemo) scrollAllToTop() {
	for _, s := range []*t.ScrollState{d.pageState, d.listState, d.textState, d.logState, d.wideState} {
		s.SetOffset(0)
		s.SetOffsetX(0)
	}
}

func (d *ScrollDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "p", Name: "Pause feed", Action: d.toggleFeed},
		{Key: "b", Name: "Log end", Action: d.logState.ScrollToBottom},
		{Key: "r", Name: "All to top", Action: d.scrollAllToTop},
		{Key: "c", Name: "Clear log", Action: func() { d.logLines.Set(nil) }, Hidden: true},
		{Key: "t", Name: "Theme", Action: demokit.NextTheme},
	}
}

func (d *ScrollDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// An animation started before the app runs is only registered once its
	// value is read, so touch it here to get the feed going.
	d.feed.Value()
	// Restart the feed if it stopped while the demo was hidden.
	if d.feedIdle {
		d.feedIdle = false
		if !d.feedPaused.Peek() {
			d.logShown = true
			d.startFeed()
		}
	}

	return t.Dock{
		ID:    "scroll-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top: []t.Widget{
			demokit.Header{Title: "Scroll Playground", Tagline: "Scrollables inside a scrollable"},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				demokit.Fill(pagePanel{demo: d}),
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{demo: d},
						keysPanel{},
					},
				},
			},
		},
	}
}

// scrollPanel is a bordered, focusable Scrollable whose border lights up
// while it has focus.
func scrollPanel(ctx t.BuildContext, id, title string, state *t.ScrollState, width, height t.Dimension, child t.Widget) t.Scrollable {
	s := t.Scrollable{
		ID:        id,
		State:     state,
		Focusable: true,
		Child:     child,
	}
	s.Style = demokit.PanelStyle(ctx.Theme(), title, ctx.IsFocused(s))
	s.Style.Width = width
	s.Style.Height = height
	return s
}

// pagePanel is the outer scrollable. Its content is taller than the screen,
// so it scrolls too, while every panel inside it scrolls on its own.
type pagePanel struct {
	demo *ScrollDemo
}

func (p pagePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := p.demo
	page := scrollPanel(ctx, "page", "Page", d.pageState, t.Flex(1), t.Flex(1), t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.Row{
				Width:   t.Flex(1),
				Spacing: 1,
				Children: []t.Widget{
					scrollPanel(ctx, "scroll-list", "List · 50 rows", d.listState, t.Flex(3), t.Cells(topRowHeight), listContent{}),
					scrollPanel(ctx, "scroll-text", "Wrapped text", d.textState, t.Flex(2), t.Cells(topRowHeight), t.Text{
						Content: loremText,
						Wrap:    t.WrapSoft,
						Width:   t.Flex(1),
						Style:   t.Style{ForegroundColor: theme.Text},
					}),
				},
			},
			t.Row{
				Width:   t.Flex(1),
				Spacing: 1,
				Children: []t.Widget{
					scrollPanel(ctx, "scroll-log", "Log · PinToBottom", d.logState, t.Flex(1), t.Cells(middleRowHeight), logContent{demo: d}),
					scrollPanel(ctx, "scroll-wide", "Wide · pans both ways", d.wideState, t.Flex(1), t.Cells(middleRowHeight), wideContent{}),
				},
			},
			disabledPanel(ctx, d.offState),
			t.Text{
				Spans: t.ParseMarkup("[$TextMuted]The page is a Scrollable too, holding the others. Tab to it and scroll down to reach this line.[/]", theme),
				Wrap:  t.WrapSoft,
				Width: t.Flex(1),
			},
		},
	})
	// Keep the Scrollable a direct child of a flex Column rather than the
	// root of this component: that way it fills the space and is registered
	// as focusable.
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{page},
	}
}

// listContent is 50 rows of striped text.
type listContent struct{}

func (listContent) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	rows := make([]t.Widget, 0, 50)
	for i := 1; i <= 50; i++ {
		color := "$Text"
		if i%2 == 0 {
			color = "$TextMuted"
		}
		rows = append(rows, t.ParseMarkupToText(fmt.Sprintf("[$Accent]%2d[/]  [%s]This is a scrollable list item[/]", i, color), theme))
	}
	return t.Column{Width: t.Flex(1), Children: rows}
}

// logContent renders the log lines. It reads the lines itself so a new line
// only rebuilds the log.
type logContent struct {
	demo *ScrollDemo
}

func (l logContent) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	l.demo.logShown = true
	lines := l.demo.logLines.Get()
	if len(lines) == 0 {
		return t.ParseMarkupToText("[$TextMuted]Log is empty.[/]", theme)
	}
	rows := make([]t.Widget, len(lines))
	for i, line := range lines {
		num, msg, _ := strings.Cut(line, "  ")
		rows[i] = t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]%s[/]  [$Text]%s[/]", num, msg), theme)
	}
	return t.Column{Width: t.Flex(1), Children: rows}
}

// wideContent is wider than its panel, so it scrolls horizontally.
type wideContent struct{}

func (wideContent) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	const cols = 20
	cells := func(format func(c int) string) string {
		var b strings.Builder
		for c := 0; c < cols; c++ {
			fmt.Fprintf(&b, "%-10s", format(c))
		}
		return b.String()
	}
	rows := []t.Widget{
		t.Text{Content: cells(func(c int) string { return fmt.Sprintf("╷%d", c*10) }), Style: t.Style{ForegroundColor: theme.TextMuted}},
		t.Text{Content: cells(func(c int) string { return fmt.Sprintf("col %02d", c+1) }), Style: t.Style{ForegroundColor: theme.Info, Bold: true}},
	}
	for r := 1; r <= 12; r++ {
		rows = append(rows, t.Text{
			Content: cells(func(c int) string { return fmt.Sprintf("%d.%02d", r, c+1) }),
			Style:   t.Style{ForegroundColor: theme.Text},
		})
	}
	return t.Column{Children: rows}
}

// disabledPanel has more text than fits, but DisableScroll clips it instead
// of scrolling. It can't take focus and shows no scrollbar.
func disabledPanel(ctx t.BuildContext, state *t.ScrollState) t.Widget {
	theme := ctx.Theme()
	style := demokit.PanelStyle(theme, "", false)
	style.Border = t.RoundedBorder(theme.Warning, t.BorderTitleMarkup("[b $Warning] DisableScroll [/]"))
	style.Width = t.Flex(1)
	style.Height = t.Cells(disabledHeight)
	return t.Scrollable{
		ID:            "no-scroll",
		State:         state,
		DisableScroll: true,
		Style:         style,
		Child: t.Text{
			Content: "Scrolling is disabled here, so content that overflows is clipped and there is no scrollbar. This panel can't take focus and ignores the mouse wheel. " +
				"This sentence keeps going so that it runs past the bottom border, where it is cut off instead of becoming scrollable. You shouldn't be able to read this far.",
			Wrap:  t.WrapSoft,
			Width: t.Flex(1),
			Style: t.Style{ForegroundColor: theme.TextMuted},
		},
	}
}

// statePanel shows live scroll offsets. It subscribes to the ScrollState
// signals itself, so scrolling only rebuilds this panel.
type statePanel struct {
	demo *ScrollDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	focus := "—"
	if f, ok := ctx.Focused().(t.Identifiable); ok && f.WidgetID() != "" {
		focus = strings.TrimPrefix(f.WidgetID(), "scroll-")
	}

	logOffset := d.logState.Offset.Get()
	logLines := len(d.logLines.Get())
	logNote := ""
	if d.logState.IsPinned() && d.logState.IsAtBottom() {
		logNote = "pinned"
	}

	feed := "[b $Success]live[/]"
	if d.feedPaused.Get() {
		feed = "[b $Warning]paused[/]"
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Focus", fmt.Sprintf("[b $Accent]%s[/]", focus)),
			demokit.StatRow(theme, "Page", offsetMarkup(d.pageState.Offset.Get(), d.pageState, "")),
			demokit.StatRow(theme, "List", offsetMarkup(d.listState.Offset.Get(), d.listState, "")),
			demokit.StatRow(theme, "Text", offsetMarkup(d.textState.Offset.Get(), d.textState, "")),
			demokit.StatRow(theme, "Log", offsetMarkup(logOffset, d.logState, logNote)),
			demokit.StatRow(theme, "Wide (x)", fmt.Sprintf("[b $Info]%d[/]", d.wideState.OffsetX.Get())),
			demokit.StatRow(theme, "Log lines", fmt.Sprintf("[b $Primary]%d[/]", logLines)),
			demokit.StatRow(theme, "Feed", feed),
		},
	}
}

// offsetMarkup formats a vertical offset, noting when it is at the top or the
// end. Callers read the offset with Get, so this is recomputed whenever the
// offset changes.
func offsetMarkup(offset int, state *t.ScrollState, note string) string {
	if note == "" {
		switch {
		case offset == 0:
			note = "top"
		case state.IsAtBottom():
			note = "end"
		}
	}
	if note != "" {
		return fmt.Sprintf("[$TextMuted]%s[/] [b $Info]%d[/]", note, offset)
	}
	return fmt.Sprintf("[b $Info]%d[/]", offset)
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-9s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("↑↓ jk", "$Info", "scroll a line"),
			key("PgUp PgDn", "$Info", "a page (^u ^d)"),
			key("Home End", "$Info", "top / end (g G)"),
			key("←→ hl", "$Info", "pan sideways"),
			key("tab", "$Accent", "next scrollable"),
			key("p b", "$Success", "pause / log end"),
			key("r c", "$Warning", "all to top / clear"),
		},
	}
}
