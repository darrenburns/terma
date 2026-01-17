package main

import (
	"fmt"
	"log"
	"time"

	t "terma"
)

// TickerDemo demonstrates the problem with signals not triggering re-renders.
// The counter increments every second via time.Ticker, but the UI won't update
// until you interact with it (move mouse, press a key, etc).
//
// Try it: Run this example and watch the counter. It won't visually update
// until you press a key or move your mouse over the terminal.
type TickerDemo struct {
	counter t.Signal[int]
}

func NewTickerDemo() *TickerDemo {
	demo := &TickerDemo{
		counter: t.NewSignal(0),
	}

	// Start a goroutine that increments the counter every second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			demo.counter.Update(func(c int) int { return c + 1 })
			t.Log("Counter updated to %d (but UI won't refresh until next event)", demo.counter.Peek())
		}
	}()

	return demo
}

func (d *TickerDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:      "ticker-root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Signal Re-render Problem Demo",
				Style: t.Style{
					ForegroundColor: theme.Background,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},

			t.Text{
				Content: "This counter increments every second via time.Ticker:",
			},

			t.Text{
				Content: fmt.Sprintf("Counter: %d", d.counter.Get()),
				Style: t.Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},

			t.Spacer{Height: t.Cells(1)},

			t.Text{
				Content: "Notice: The counter IS incrementing in the background",
				Style:   t.Style{ForegroundColor: theme.Warning},
			},
			t.Text{
				Content: "(check terma.log to see updates)",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},

			t.Spacer{Height: t.Cells(1)},

			t.Text{
				Content: "But the UI only refreshes when you interact!",
				Style:   t.Style{ForegroundColor: theme.Error},
			},
			t.Text{
				Content: "Try: Press any key or move your mouse to see it update.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},

			t.Spacer{Height: t.Cells(1)},

			t.Text{
				Spans: t.ParseMarkup("Press [b $Accent]Ctrl+C[/] to quit", theme),
			},
		},
	}
}

func main() {
	t.InitLogger()
	app := NewTickerDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
