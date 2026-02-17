package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type App struct {
	outerScroll *t.ScrollState
	badgeCount  t.Signal[int]
}

func NewApp() *App {
	return &App{
		outerScroll: t.NewScrollState(),
		badgeCount:  t.NewSignal(5),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		Top: []t.Widget{
			t.Text{
				Content: " Scroll + Overflow Test ",
				Style: t.Style{
					ForegroundColor: theme.Background,
					BackgroundColor: theme.Primary,
				},
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Width:  t.Flex(1),
			Height: t.Flex(1),
			Style: t.Style{
				Padding:         t.EdgeInsetsAll(1),
				BackgroundColor: theme.Background,
			},
			Children: []t.Widget{
				// Test 1: Card outside scrollable (should work)
				t.Text{Content: "Test 1: Outside scrollable:", Style: t.Style{ForegroundColor: theme.Primary}},
				a.buildCardWithBadge(ctx, "Outside", 3),

				t.Spacer{Height: t.Cells(1)},

				// Test 2: Card inside scrollable
				t.Text{Content: "Test 2: Inside scrollable:", Style: t.Style{ForegroundColor: theme.Primary}},
				t.Scrollable{
					State:  a.outerScroll,
					Height: t.Cells(10),
					Child: t.Column{
						// Try WITHOUT Flex width to see if that's the issue
						Spacing: 1,
						Style: t.Style{
							Padding:         t.EdgeInsetsAll(1),
							BackgroundColor: theme.Surface,
						},
						Children: []t.Widget{
							t.Text{Content: "Inside scrollable content:", Style: t.Style{ForegroundColor: theme.Text}},
							a.buildCardWithBadge(ctx, "Inside", 7),
							t.Text{Content: "More content below...", Style: t.Style{ForegroundColor: theme.TextMuted}},
							a.buildCardWithBadge(ctx, "Inside2", 99),
						},
					},
				},

				t.Spacer{Height: t.Cells(1)},

				// Test 3: Card inside scrollable with Flex(1) width on inner column
				t.Text{Content: "Test 3: Inside scrollable (Flex width):", Style: t.Style{ForegroundColor: theme.Primary}},
				t.Scrollable{
					State:  t.NewScrollState(),
					Height: t.Cells(10),
					Child: t.Column{
						Width:   t.Flex(1), // This might be causing the issue
						Spacing: 1,
						Style: t.Style{
							Padding:         t.EdgeInsetsAll(1),
							BackgroundColor: t.RGB(60, 40, 40),
						},
						Children: []t.Widget{
							t.Text{Content: "Inside scrollable (flex):", Style: t.Style{ForegroundColor: theme.Text}},
							a.buildCardWithBadge(ctx, "Flex", 42),
						},
					},
				},
			},
		},
	}
}

func (a *App) buildCardWithBadge(ctx t.BuildContext, title string, count int) t.Widget {
	theme := ctx.Theme()

	var badge t.Widget = t.EmptyWidget{}
	if count > 0 {
		badge = t.Positioned{
			Top:   t.IntPtr(-1),
			Right: t.IntPtr(-1),
			Child: t.Text{
				Content: fmt.Sprintf(" %d ", count),
				Style: t.Style{
					ForegroundColor: t.White,
					BackgroundColor: t.RGB(220, 50, 50),
				},
			},
		}
	}

	return t.Stack{
		Children: []t.Widget{
			t.Column{
				Width:  t.Cells(20),
				Height: t.Cells(3),
				Style: t.Style{
					BackgroundColor: t.RGB(50, 50, 70),
					Border:          t.Border{Style: t.BorderRounded, Color: theme.Primary},
					Padding:         t.EdgeInsets{Left: 1, Right: 1},
				},
				Children: []t.Widget{
					t.Text{Content: title, Style: t.Style{ForegroundColor: t.White}},
				},
			},
			badge,
		},
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "j", Name: "Scroll Down", Action: func() {
			a.outerScroll.ScrollDown(1)
		}},
		{Key: "k", Name: "Scroll Up", Action: func() {
			a.outerScroll.ScrollUp(1)
		}},
	}
}

func main() {
	t.InitLogger()
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
