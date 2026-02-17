package main

import (
	"log"
	"math"
	"math/rand"
	"time"

	t "github.com/darrenburns/terma"
)

// App demonstrates the Sparkline widget with various configurations.
type App struct {
	// Live data that updates over time
	liveData t.AnySignal[[]float64]
}

func NewApp() *App {
	// Initialize with some data
	initialData := make([]float64, 30)
	for i := range initialData {
		initialData[i] = rand.Float64() * 100
	}

	app := &App{
		liveData: t.NewAnySignal(initialData),
	}

	// Start a goroutine to update live data periodically
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			data := app.liveData.Get()
			// Shift data left and add new value
			newData := make([]float64, len(data))
			copy(newData, data[1:])
			newData[len(newData)-1] = rand.Float64() * 100
			app.liveData.Set(newData)
		}
	}()

	return app
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Sample data sets
	sineWave := generateSineWave(40, 2)
	randomData := []float64{23, 45, 12, 78, 34, 56, 89, 12, 45, 67, 90, 23, 45, 67, 89, 12, 34, 56, 78, 90}
	trendUp := []float64{10, 15, 12, 20, 25, 22, 30, 35, 40, 45, 42, 50, 55, 60, 65}
	trendDown := []float64{90, 85, 80, 75, 78, 70, 65, 60, 55, 58, 50, 45, 40, 35, 30}
	spiky := []float64{10, 90, 10, 90, 10, 90, 10, 90, 10, 90, 10, 90, 10, 90, 10}

	return t.Column{
		Spacing: 1,
		Width:   t.Flex(1),
		Style:   t.Style{Padding: t.EdgeInsetsAll(1), BackgroundColor: theme.Background},
		Children: []t.Widget{
			t.Text{Content: "Sparkline Widget Demo", Style: t.Style{ForegroundColor: theme.Primary, Bold: true}},
			t.Text{Content: "Compact inline charts using Unicode bar characters", Style: t.Style{ForegroundColor: theme.TextMuted}},

			// Live updating sparkline
			t.Text{Content: "Live data (updates every 200ms):", Style: t.Style{ForegroundColor: theme.Accent}},
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Sparkline{
						Values:       a.liveData.Get(),
						Width:        t.Cells(30),
						ColorByValue: true,
						ValueColorScale: t.NewGradient(
							theme.Success,
							theme.Warning,
							theme.Error,
						),
					},
					t.Text{Content: "(color by value)", Style: t.Style{ForegroundColor: theme.TextMuted}},
				},
			},

			// Basic sparkline
			t.Text{Content: "Basic sparkline:", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Sparkline{
				Values: randomData,
				Style:  t.Style{ForegroundColor: theme.Primary},
			},

			// Sine wave
			t.Text{Content: "Sine wave pattern:", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Sparkline{
				Values: sineWave,
				Style:  t.Style{ForegroundColor: theme.Accent},
			},

			// Trend examples
			t.Text{Content: "Trend patterns:", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Column{
						Children: []t.Widget{
							t.Text{Content: "Upward", Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{
								Values: trendUp,
								Style:  t.Style{ForegroundColor: theme.Success},
							},
						},
					},
					t.Column{
						Children: []t.Widget{
							t.Text{Content: "Downward", Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{
								Values: trendDown,
								Style:  t.Style{ForegroundColor: theme.Error},
							},
						},
					},
					t.Column{
						Children: []t.Widget{
							t.Text{Content: "Volatile", Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{
								Values: spiky,
								Style:  t.Style{ForegroundColor: theme.Warning},
							},
						},
					},
				},
			},

			// ColorByValue demo
			t.Text{Content: "Color by value (gradient from low to high):", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Sparkline{
				Values:       randomData,
				ColorByValue: true,
				ValueColorScale: t.NewGradient(
					theme.TextMuted,
					theme.Primary,
				),
			},

			// Custom gradient
			t.Text{Content: "Multi-stop gradient (green -> yellow -> red):", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Sparkline{
				Values:       randomData,
				ColorByValue: true,
				ValueColorScale: t.NewGradient(
					theme.Success,
					theme.Warning,
					theme.Error,
				),
			},

			// Fixed width examples
			t.Text{Content: "Resampling (same data at different widths):", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Column{
				Spacing: 0,
				Children: []t.Widget{
					t.Row{
						Spacing: 1,
						Children: []t.Widget{
							t.Text{Content: "10 cells:", Width: t.Cells(10), Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{Values: sineWave, Width: t.Cells(10), Style: t.Style{ForegroundColor: theme.Primary}},
						},
					},
					t.Row{
						Spacing: 1,
						Children: []t.Widget{
							t.Text{Content: "20 cells:", Width: t.Cells(10), Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{Values: sineWave, Width: t.Cells(20), Style: t.Style{ForegroundColor: theme.Primary}},
						},
					},
					t.Row{
						Spacing: 1,
						Children: []t.Widget{
							t.Text{Content: "40 cells:", Width: t.Cells(10), Style: t.Style{ForegroundColor: theme.TextMuted}},
							t.Sparkline{Values: sineWave, Width: t.Cells(40), Style: t.Style{ForegroundColor: theme.Primary}},
						},
					},
				},
			},
		},
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{
			Key:    "q",
			Name:   "Quit",
			Action: t.Quit,
		},
	}
}

func generateSineWave(points int, cycles float64) []float64 {
	data := make([]float64, points)
	for i := 0; i < points; i++ {
		t := float64(i) / float64(points-1) * cycles * 2 * math.Pi
		data[i] = (math.Sin(t) + 1) / 2 * 100 // Normalize to 0-100
	}
	return data
}

func main() {
	t.SetTheme(t.ThemeNameRosePine)
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
