package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
)

const banner = `▀█▀ █▀▀ █▀█ █▀▄▀█ ▄▀█
 █  ██▄ █▀▄ █ ▀ █ █▀█`

type ShimmerDemo struct {
	states []*t.ShimmerState

	standard, rainbow, fast, slow, retry, deploy, scan, kitt, inline, banner  *t.ShimmerState
	card, neon, orbitInner, orbitOuter, button, bar, typing, aurora, skeleton *t.ShimmerState
	repos                                                                     [3]*t.ShimmerState

	scroll *t.ScrollState
}

func NewShimmerDemo() *ShimmerDemo {
	d := &ShimmerDemo{scroll: t.NewScrollState()}
	d.standard = d.newState(1500 * time.Millisecond)
	d.rainbow = d.newState(2500 * time.Millisecond)
	d.fast = d.newState(600 * time.Millisecond)
	d.slow = d.newState(4 * time.Second)
	d.retry = d.newState(1200 * time.Millisecond)
	d.deploy = d.newState(2 * time.Second)
	d.scan = d.newState(1800 * time.Millisecond)
	d.kitt = d.newState(900 * time.Millisecond)
	d.inline = d.newState(1300 * time.Millisecond)
	d.banner = d.newState(2200 * time.Millisecond)
	d.card = d.newState(3 * time.Second)
	d.neon = d.newState(1600 * time.Millisecond)
	d.orbitInner = d.newState(2 * time.Second)
	d.orbitOuter = d.newState(5 * time.Second)
	d.button = d.newState(1400 * time.Millisecond)
	d.bar = d.newState(1100 * time.Millisecond)
	d.typing = d.newState(2400 * time.Millisecond)
	d.aurora = d.newState(6 * time.Second)
	d.skeleton = d.newState(1700 * time.Millisecond)
	for i := range d.repos {
		d.repos[i] = d.newState(time.Duration(1500+i*230) * time.Millisecond)
	}
	return d
}

func (d *ShimmerDemo) newState(period time.Duration) *t.ShimmerState {
	state := t.NewShimmerState(period)
	state.Start()
	d.states = append(d.states, state)
	return state
}

func (d *ShimmerDemo) toggle() {
	for _, state := range d.states {
		if state.IsRunning() {
			state.Stop()
		} else {
			state.Start()
		}
	}
}

func (d *ShimmerDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "space", Name: "Start/stop all", Action: d.toggle},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (d *ShimmerDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Dock{
		Bottom: []t.Widget{t.KeybindBar{Style: t.Style{BackgroundColor: theme.Surface}}},
		Body: t.Scrollable{
			State: d.scroll,
			Child: t.Row{
				Spacing: 4,
				Style:   t.Style{Padding: t.EdgeInsetsXY(2, 1)},
				Children: []t.Widget{
					t.Column{Spacing: 1, Style: t.Style{Width: t.Flex(1)}, Children: d.textExamples(theme)},
					t.Column{Spacing: 1, Style: t.Style{Width: t.Flex(1)}, Children: d.boxExamples(theme)},
				},
			},
		},
	}
}

func example(theme t.ThemeData, number int, caption string, demo t.Widget) t.Widget {
	return t.Column{Children: []t.Widget{
		t.Text{
			Spans: []t.Span{
				t.StyledSpan(fmt.Sprintf("%02d ", number), t.SpanStyle{Foreground: theme.Accent}),
				t.StyledSpan(caption, t.SpanStyle{Foreground: theme.TextMuted, Italic: true}),
			},
		},
		demo,
	}}
}

func shimmerText(content string, s t.Shimmer) t.Text {
	return t.Text{Content: content, Style: t.Style{ForegroundColor: s}}
}

func (d *ShimmerDemo) textExamples(theme t.ThemeData) []t.Widget {
	return []t.Widget{
		example(theme, 1, "the classic", shimmerText("Thinking about your request...",
			t.Shimmer{State: d.standard, Base: theme.TextMuted, Highlight: theme.Text})),

		example(theme, 2, "gradient base, white-hot band", shimmerText("Reticulating splines across the cluster",
			t.Shimmer{
				State:     d.rainbow,
				Base:      t.NewGradient(theme.Primary, theme.Secondary, theme.Accent).WithAngle(90),
				Highlight: t.White,
				BandWidth: 6,
			})),

		example(theme, 3, "narrow and fast", shimmerText("Uploading 42 files ▲▲▲",
			t.Shimmer{State: d.fast, Base: theme.Info.WithAlpha(0.5), Highlight: theme.Info, BandWidth: 3})),

		example(theme, 4, "wide and slow, like breathing", shimmerText("Waiting for a reviewer to approve",
			t.Shimmer{State: d.slow, Base: theme.TextMuted, Highlight: theme.Text, BandWidth: 30})),

		example(theme, 5, "trouble, but trying", shimmerText("✗ Connection lost. Retrying (attempt 3 of 5)",
			t.Shimmer{State: d.retry, Base: theme.Error.WithAlpha(0.55), Highlight: theme.Error})),

		example(theme, 6, "good news in progress", shimmerText("✓ Deploying build #1207 to production",
			t.Shimmer{State: d.deploy, Base: theme.Success.WithAlpha(0.5), Highlight: theme.Success, BandWidth: 12})),

		example(theme, 7, "a scanner behind the text", t.Text{
			Content: " Scanning dependencies for vulnerabilities ",
			Style: t.Style{
				ForegroundColor: theme.Text,
				BackgroundColor: t.Shimmer{State: d.scan, Base: theme.Surface, Highlight: theme.Primary.WithAlpha(0.6), BandWidth: 10},
			},
		}),

		example(theme, 8, "KITT's scanner bar", shimmerText("■■■■■■■■■■■■■■■■■■■■■■■■",
			t.Shimmer{State: d.kitt, Base: theme.Error.WithAlpha(0.15), Highlight: theme.Error, BandWidth: 5})),

		example(theme, 9, "inline, mid-sentence", t.Row{Children: []t.Widget{
			t.Text{Content: "Asking ", Style: t.Style{ForegroundColor: theme.Text}},
			shimmerText("the-oracle-v2",
				t.Shimmer{State: d.inline, Base: theme.Secondary, Highlight: theme.Text, BandWidth: 4}),
			t.Text{Content: " about your query", Style: t.Style{ForegroundColor: theme.Text}},
		}}),

		example(theme, 10, "block letters", shimmerText(banner,
			t.Shimmer{
				State:     d.banner,
				Base:      t.NewGradient(theme.Primary, theme.Accent).WithAngle(90),
				Highlight: t.White,
				BandWidth: 5,
			})),
	}
}

func card(border t.Border, children ...t.Widget) t.Column {
	return t.Column{
		Style:    t.Style{Border: border, Padding: t.EdgeInsetsXY(1, 0)},
		Children: children,
	}
}

func skeletonLine(width int, s t.Shimmer) t.Widget {
	return t.Text{Content: strings.Repeat("█", width), Style: t.Style{ForegroundColor: s}}
}

func (d *ShimmerDemo) boxExamples(theme t.ThemeData) []t.Widget {
	skeleton := t.Shimmer{State: d.skeleton, Base: theme.Surface, Highlight: theme.TextMuted.WithAlpha(0.45), BandWidth: 14}
	orbit := t.Shimmer{
		State:     d.orbitInner,
		Base:      t.Shimmer{State: d.orbitOuter, Base: theme.Border, Highlight: theme.Secondary, BandWidth: 14, Path: t.ShimmerPerimeter},
		Highlight: theme.Accent,
		BandWidth: 6,
		Path:      t.ShimmerPerimeter,
	}

	var repoRows []t.Widget
	for i, name := range []string{"terma", "ultraviolet", "lipgloss"} {
		repoRows = append(repoRows, t.Text{
			Content: fmt.Sprintf("◇ darrenburns/%-12s ░░░░░░", name),
			Style:   t.Style{ForegroundColor: t.Shimmer{State: d.repos[i], Base: theme.TextMuted, Highlight: theme.Text, BandWidth: 10}},
		})
	}

	return []t.Widget{
		example(theme, 11, "a comet around the card", card(
			t.RoundedBorder(t.Shimmer{State: d.card, Base: theme.Border, Highlight: theme.Accent, BandWidth: 12, Path: t.ShimmerPerimeter}, t.BorderTitle("Indexing")),
			t.Text{Content: "Scanning 1,204 files in the workspace"},
		)),

		example(theme, 12, "neon sign", card(
			t.HeavyBorder(t.Shimmer{State: d.neon, Base: theme.Error.WithAlpha(0.4), Highlight: theme.Warning, BandWidth: 8, Path: t.ShimmerPerimeter}),
			shimmerText("OPEN 24/7 · COFFEE · DEBUGGING",
				t.Shimmer{State: d.neon, Base: theme.Error.WithAlpha(0.6), Highlight: theme.Warning, BandWidth: 8}),
		)),

		example(theme, 13, "two orbits: a shimmer inside a shimmer", card(
			t.DoubleBorder(orbit, t.BorderTitleCenter("orbit")),
			t.Text{Content: "Fast accent comet over a slow secondary tide", Style: t.Style{ForegroundColor: theme.TextMuted}},
		)),

		example(theme, 14, "a pill that wants to be pressed", t.Row{Children: []t.Widget{t.Text{
			Content: "  ✦ Generate  ",
			Style: t.Style{
				Bold:            true,
				ForegroundColor: theme.TextOnPrimary,
				BackgroundColor: t.Shimmer{State: d.button, Base: theme.Primary, Highlight: theme.Accent, BandWidth: 6},
			},
		}}}),

		example(theme, 15, "indeterminate progress bar", shimmerText(strings.Repeat("━", 40),
			t.Shimmer{State: d.bar, Base: theme.Border, Highlight: theme.Primary, BandWidth: 10})),

		example(theme, 16, "someone is typing", card(
			t.RoundedBorder(t.Shimmer{State: d.typing, Base: theme.Border, Highlight: theme.Primary, BandWidth: 10, Path: t.ShimmerPerimeter}, t.BorderTitle("assistant")),
			shimmerText("● ● ●  composing a reply",
				t.Shimmer{State: d.typing, Base: theme.TextMuted, Highlight: theme.Primary, BandWidth: 6}),
		)),

		example(theme, 17, "aurora: gradient border, white comet", card(
			t.RoundedBorder(t.Shimmer{
				State:     d.aurora,
				Base:      t.NewGradient(theme.Success, theme.Info, theme.Secondary).WithAngle(90),
				Highlight: t.White,
				BandWidth: 8,
				Path:      t.ShimmerPerimeter,
			}),
			t.Text{Content: "Northern lights over a slow loop", Style: t.Style{ForegroundColor: theme.TextMuted}},
		)),

		example(theme, 18, "skeleton screen", t.Column{Children: []t.Widget{
			skeletonLine(14, skeleton),
			skeletonLine(36, skeleton),
			skeletonLine(28, skeleton),
		}}),

		example(theme, 19, "staggered rows drift out of phase", t.Column{Children: repoRows}),

		example(theme, 20, "dashed border, hazard band", card(
			t.DashedBorder(t.Shimmer{State: d.neon, Base: theme.Warning.WithAlpha(0.35), Highlight: theme.Warning, BandWidth: 16, Path: t.ShimmerPerimeter}, t.BorderTitle("migration")),
			t.Text{Content: "Rewriting 3,812 rows. Don't close this window.", Style: t.Style{ForegroundColor: theme.Warning}},
		)),
	}
}

func main() {
	if err := t.Run(NewShimmerDemo()); err != nil {
		log.Fatal(err)
	}
}
