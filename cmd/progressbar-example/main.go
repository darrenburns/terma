package main

import (
	"fmt"
	"log"
	"math"
	"time"

	t "github.com/darrenburns/terma"
)

// App demonstrates ProgressBar and Spinner: a bar driven by a looping
// Animation, a bar driven by an AnimatedValue, static bars that show the
// eighth-of-a-cell precision, and every built-in spinner style.
//
// Keys:
//
//	tab        - move between the buttons (enter or space presses)
//	+ / -      - move the interactive bar by 10%
//	← / →      - move it by 1%
//	0 / f      - empty / fill it
//	p          - pause / resume the looping bar
//	s          - start / stop the spinners
//	t          - cycle theme
//	q          - quit
type App struct {
	// progress eases towards whatever it is set to.
	progress *t.AnimatedValue[float64]

	// loop runs from 0 to 1 over and over.
	loop       *t.Animation[float64]
	loopPaused t.Signal[bool]

	// loopSpinner spins next to the looping bar while it runs.
	loopSpinner *t.SpinnerState
	spinners    []namedSpinner
	spinning    t.Signal[bool]

	themeIndex t.Signal[int]
}

type namedSpinner struct {
	name  string
	state *t.SpinnerState
}

var themeNames = []string{
	t.ThemeNameRosePine,
	t.ThemeNameDracula,
	t.ThemeNameTokyoNight,
	t.ThemeNameCatppuccin,
	t.ThemeNameGruvbox,
	t.ThemeNameNord,
	t.ThemeNameSolarized,
	t.ThemeNameKanagawa,
	t.ThemeNameMonokai,
}

const sidebarWidth = 32

func NewApp() *App {
	a := &App{
		progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
			Initial:  0.35,
			Duration: 300 * time.Millisecond,
			Easing:   t.EaseOutCubic,
		}),
		loopPaused:  t.NewSignal(false),
		loopSpinner: t.NewSpinnerState(t.SpinnerDots),
		spinning:    t.NewSignal(true),
		themeIndex:  t.NewSignal(0),
	}

	a.loop = t.NewAnimation(t.AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: 3 * time.Second,
		Easing:   t.EaseInOutSine,
		OnComplete: func() {
			a.loop.Reset()
			a.loop.Start()
		},
	})
	// Safe before the app runs: the animation registers when its value is
	// first read.
	a.loop.Start()
	a.loopSpinner.Start()

	for _, s := range []struct {
		name  string
		style t.SpinnerStyle
	}{
		{"Dots", t.SpinnerDots},
		{"Line", t.SpinnerLine},
		{"Circle", t.SpinnerCircle},
		{"Bounce", t.SpinnerBounce},
		{"Arrow", t.SpinnerArrow},
		{"Braille", t.SpinnerBraille},
		{"Grow", t.SpinnerGrow},
		{"Pulse", t.SpinnerPulse},
		{"DotsBounce", t.SpinnerDotsBounce},
		{"Clock", t.SpinnerClock},
		{"Moon", t.SpinnerMoon},
	} {
		state := t.NewSpinnerState(s.style)
		state.Start()
		a.spinners = append(a.spinners, namedSpinner{name: s.name, state: state})
	}
	return a
}

// nudge moves the interactive bar's target by delta, rounded to whole
// percents so repeated steps don't drift.
func (a *App) nudge(delta float64) {
	target := math.Round((a.progress.Target()+delta)*100) / 100
	a.progress.Set(min(1, max(0, target)))
}

func (a *App) toggleLoop() {
	if a.loopPaused.Peek() {
		a.loop.Resume()
		a.loopSpinner.Start()
	} else {
		a.loop.Pause()
		a.loopSpinner.Stop()
	}
	a.loopPaused.Set(!a.loopPaused.Peek())
}

func (a *App) toggleSpinners() {
	on := !a.spinning.Peek()
	for _, s := range a.spinners {
		if on {
			s.state.Start()
		} else {
			s.state.Stop()
		}
	}
	a.spinning.Set(on)
}

func (a *App) cycleTheme() {
	a.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "+", Name: "+10%", Action: func() { a.nudge(0.1) }},
		{Key: "-", Name: "-10%", Action: func() { a.nudge(-0.1) }},
		{Key: "right", Name: "+1%", Action: func() { a.nudge(0.01) }, Hidden: true},
		{Key: "left", Name: "-1%", Action: func() { a.nudge(-0.01) }, Hidden: true},
		{Key: "0", Name: "Empty", Action: func() { a.progress.Set(0) }, Hidden: true},
		{Key: "f", Name: "Fill", Action: func() { a.progress.Set(1) }, Hidden: true},
		{Key: "p", Name: "Pause loop", Action: a.toggleLoop},
		{Key: "s", Name: "Spinners", Action: a.toggleSpinners},
		{Key: "t", Name: "Theme", Action: a.cycleTheme},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID:    "progress-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top: []t.Widget{
			header{themeName: themeNames[a.themeIndex.Get()]},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
			Children: []t.Widget{
				t.Column{
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						loopPanel{app: a},
						interactivePanel{app: a},
						staticPanel{},
						spinnersPanel{app: a},
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{app: a},
						keysPanel{},
					},
				},
			},
		},
	}
}

// header is the title bar across the top of the screen.
type header struct {
	themeName string
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ Progress & Spinners[/]  [$TextMuted]Smooth bars, eighth-cell precision[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", h.themeName), theme),
		},
	}
}

// panelStyle is the bordered look shared by every panel in the demo.
func panelStyle(theme t.ThemeData, title string, focused bool) t.Style {
	color := theme.Border
	titleColor := "$TextMuted"
	if focused {
		color = theme.FocusRing
		titleColor = "$FocusRing"
	}
	return t.Style{
		BackgroundColor: theme.Background,
		Border:          t.RoundedBorder(color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] %s [/]", titleColor, title))),
		Padding:         t.EdgeInsetsXY(1, 0),
	}
}

// percentBar is a right-aligned percentage followed by a bar that fills the
// rest of the row.
func percentBar(theme t.ThemeData, progress float64, color t.Color) t.Widget {
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content:   fmt.Sprintf("%.0f%%", progress*100),
				Width:     t.Cells(4),
				TextAlign: t.TextAlignRight,
				Style:     t.Style{ForegroundColor: theme.Text, Bold: true},
			},
			t.ProgressBar{Progress: progress, Width: t.Flex(1), FilledColor: color},
		},
	}
}

// loopPanel shows the bar driven by the looping Animation. It reads the
// animation's value itself, so each frame only rebuilds this panel.
type loopPanel struct {
	app *App
}

func (l loopPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	value := l.app.loop.Value().Get()
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   panelStyle(theme, "Looping · Animation", false),
		Children: []t.Widget{
			t.Spinner{State: l.app.loopSpinner, Style: t.Style{ForegroundColor: theme.Accent}},
			percentBar(theme, value, theme.Accent),
		},
	}
}

// interactivePanel shows the bar driven by an AnimatedValue and the buttons
// that move it. Its border lights up while one of the buttons has focus.
type interactivePanel struct {
	app *App
}

func (p interactivePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := p.app
	// A Button's Style applies when it isn't focused and its Variant colours
	// when it is, so this pairing gives a clear focus highlight.
	button := func(id, label string, onPress func()) t.Widget {
		return t.Button{
			ID:      id,
			Label:   label,
			Variant: t.ButtonPrimary,
			Style:   t.Style{ForegroundColor: theme.Text, BackgroundColor: theme.Surface},
			OnPress: onPress,
		}
	}
	buttons := []t.Widget{
		button("decrement", "-10%", func() { a.nudge(-0.1) }),
		button("increment", "+10%", func() { a.nudge(0.1) }),
		button("reset", "Empty", func() { a.progress.Set(0) }),
		button("fill", "Fill", func() { a.progress.Set(1) }),
	}
	focused := false
	for _, b := range buttons {
		focused = focused || ctx.IsFocused(b)
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Interactive · AnimatedValue", focused),
		Children: []t.Widget{
			percentBar(theme, a.progress.Get(), theme.Secondary),
			t.Row{
				Spacing:  1,
				Style:    t.Style{Padding: t.EdgeInsets{Left: 5}},
				Children: buttons,
			},
		},
	}
}

// staticPanel shows fixed values in different colours, then one-cell bars
// stepping through each eighth of a cell.
type staticPanel struct{}

func (staticPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	rows := []t.Widget{}
	for _, v := range []struct {
		progress float64
		color    t.Color
	}{
		{0, theme.Primary},
		{0.125, theme.Primary},
		{0.5, theme.Success},
		{0.75, theme.Warning},
		{1, theme.Error},
	} {
		rows = append(rows, percentBar(theme, v.progress, v.color))
	}

	eighths := []t.Widget{
		t.Text{Content: "8ths", Width: t.Cells(4), TextAlign: t.TextAlignRight, Style: t.Style{ForegroundColor: theme.TextMuted}},
	}
	for i := 0; i <= 8; i++ {
		eighths = append(eighths, t.ProgressBar{Progress: float64(i) / 8, Width: t.Cells(1), FilledColor: theme.Info})
	}
	eighths = append(eighths, t.ParseMarkupToText("[$TextMuted]one cell, 0 → 8 eighths[/]", theme))
	rows = append(rows, t.Row{Spacing: 1, Children: eighths})

	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Static values", false),
		Children: rows,
	}
}

// spinnersPanel shows every built-in SpinnerStyle.
type spinnersPanel struct {
	app *App
}

func (s spinnersPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	const perRow = 4
	var rows []t.Widget
	var cells []t.Widget
	for i, sp := range s.app.spinners {
		cells = append(cells, t.Row{
			Width:   t.Flex(1),
			Spacing: 1,
			Children: []t.Widget{
				t.Spinner{State: sp.state, Style: t.Style{ForegroundColor: theme.Primary}},
				t.Text{Content: sp.name, Style: t.Style{ForegroundColor: theme.TextMuted}},
			},
		})
		if len(cells) == perRow || i == len(s.app.spinners)-1 {
			for len(cells) < perRow {
				cells = append(cells, t.Spacer{Width: t.Flex(1), Height: t.Cells(1)})
			}
			rows = append(rows, t.Row{Width: t.Flex(1), Children: cells})
			cells = nil
		}
	}
	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Spinners", false),
		Children: rows,
	}
}

// statePanel shows the live values. It subscribes to the animations itself,
// so animation frames only rebuild this panel.
type statePanel struct {
	app *App
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := s.app

	loopStatus := "[b $Success]running[/]"
	if a.loopPaused.Get() {
		loopStatus = "[b $Warning]paused[/]"
	}
	value := a.progress.Get()
	target := a.progress.Target()
	moving := "[$TextMuted]idle[/]"
	if value != target {
		moving = "[b $Info]easing[/]"
	}
	spinners := "[b $Success]on[/]"
	if !a.spinning.Get() {
		spinners = "[b $Warning]off[/]"
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "State", false),
		Children: []t.Widget{
			statRow(theme, "Loop", fmt.Sprintf("[b $Accent]%.0f%%[/]", a.loop.Value().Get()*100)),
			statRow(theme, "Loop status", loopStatus),
			statRow(theme, "Bar value", fmt.Sprintf("[b $Secondary]%.1f%%[/]", value*100)),
			statRow(theme, "Bar target", fmt.Sprintf("[b $Secondary]%.0f%%[/]", target*100)),
			statRow(theme, "Bar", moving),
			statRow(theme, "Spinners", spinners),
		},
	}
}

func statRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("+ -", "$Secondary", "bar ±10%"),
			key("← →", "$Secondary", "bar ±1%"),
			key("0 f", "$Secondary", "empty / fill"),
			key("tab", "$Info", "next button"),
			key("p", "$Accent", "pause the loop"),
			key("s", "$Primary", "toggle spinners"),
			key("t q", "$Warning", "theme / quit"),
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	t.RequestFocus("increment")
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
