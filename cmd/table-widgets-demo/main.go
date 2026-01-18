package main

import (
	"log"
	"time"

	t "terma"
)

// Task represents a row in our table with various states
type Task struct {
	Name         string
	Status       TaskStatus
	Progress     *t.Animation[float64]      // Animated progress for running tasks
	Metrics      [][]string                 // 2x3 nested table data
	MetricsState *t.TableState[[]string]    // State for nested table
}

type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusComplete
	StatusFailed
)

func (s TaskStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusRunning:
		return "Running"
	case StatusComplete:
		return "Complete"
	case StatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

type App struct {
	tableState   *t.TableState[Task]
	scrollState  *t.ScrollState
	spinnerState *t.SpinnerState
}

// newRunningTask creates a task with an animated progress bar
func newRunningTask(name string, startProgress float64, duration time.Duration, metrics [][]string) Task {
	anim := t.NewAnimation(t.AnimationConfig[float64]{
		From:     startProgress,
		To:       1.0,
		Duration: duration,
		Easing:   t.EaseInOutSine,
	})
	anim.Start()
	return Task{
		Name:         name,
		Status:       StatusRunning,
		Progress:     anim,
		Metrics:      metrics,
		MetricsState: t.NewTableState(metrics),
	}
}

// newStaticTask creates a task with a fixed progress value
func newStaticTask(name string, progress float64, status TaskStatus, metrics [][]string) Task {
	// Use a completed animation for static values
	anim := t.NewAnimation(t.AnimationConfig[float64]{
		From:     progress,
		To:       progress,
		Duration: time.Millisecond,
	})
	return Task{
		Name:         name,
		Status:       status,
		Progress:     anim,
		Metrics:      metrics,
		MetricsState: t.NewTableState(metrics),
	}
}

func NewApp() *App {
	tasks := []Task{
		newStaticTask("Download assets", 1.0, StatusComplete, [][]string{
			{"CPU", "12%"}, {"Mem", "256M"}, {"I/O", "45MB/s"},
		}),
		newRunningTask("Compile shaders", 0.2, 8*time.Second, [][]string{
			{"CPU", "89%"}, {"Mem", "1.2G"}, {"I/O", "12MB/s"},
		}),
		newRunningTask("Process textures", 0.1, 12*time.Second, [][]string{
			{"CPU", "45%"}, {"Mem", "2.1G"}, {"I/O", "120MB/s"},
		}),
		newStaticTask("Build index", 0.0, StatusPending, [][]string{
			{"CPU", "0%"}, {"Mem", "0M"}, {"I/O", "0MB/s"},
		}),
		newStaticTask("Generate thumbnails", 0.0, StatusPending, [][]string{
			{"CPU", "0%"}, {"Mem", "0M"}, {"I/O", "0MB/s"},
		}),
		newRunningTask("Validate checksums", 0.4, 6*time.Second, [][]string{
			{"CPU", "23%"}, {"Mem", "512M"}, {"I/O", "88MB/s"},
		}),
		newStaticTask("Upload to CDN", 0.0, StatusPending, [][]string{
			{"CPU", "0%"}, {"Mem", "0M"}, {"I/O", "0MB/s"},
		}),
		newStaticTask("Network sync failed", 0.33, StatusFailed, [][]string{
			{"CPU", "5%"}, {"Mem", "128M"}, {"I/O", "0MB/s"},
		}),
	}

	spinnerState := t.NewSpinnerState(t.SpinnerDots)
	spinnerState.Start()

	return &App{
		tableState:   t.NewTableState(tasks),
		scrollState:  t.NewScrollState(),
		spinnerState: spinnerState,
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Get spinner frame to subscribe and use in table cells
	spinnerFrame := a.spinnerState.Frame()

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style: t.Style{
			Padding:         t.EdgeInsetsAll(1),
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			t.Text{
				Content: "Table with Arbitrary Widgets",
				Style: t.Style{
					Bold:            true,
					ForegroundColor: theme.Primary,
				},
			},
			t.Text{
				Content: "ProgressBars, Spinners, and nested Tables inside table cells.",
				Style: t.Style{
					ForegroundColor: theme.TextMuted,
					Padding:         t.EdgeInsets{Bottom: 1},
				},
			},
			t.Scrollable{
				State:  a.scrollState,
				Height: t.Flex(1),
				Child: t.Table[Task]{
					State:       a.tableState,
					ScrollState: a.scrollState,
					Columns: []t.TableColumn{
						{Width: t.Cells(22), Header: t.Text{Content: "Task"}},
						{Width: t.Cells(25), Header: t.Text{Content: "Progress"}},
						{Width: t.Cells(12), Header: t.Text{Content: "Status"}},
						{Width: t.Cells(20), Header: t.Text{Content: "Metrics"}},
					},
					ColumnSpacing: 2,
					RenderCell:    a.renderCell(ctx, spinnerFrame),
				},
			},
			t.KeybindBar{
				Style: t.Style{
					Padding:         t.EdgeInsets{Top: 1},
					ForegroundColor: theme.TextMuted,
				},
			},
		},
	}
}

func (a *App) renderCell(ctx t.BuildContext, spinnerFrame string) func(row Task, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
	theme := ctx.Theme()

	return func(row Task, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
		// Determine cell background based on active/selected state
		var bg t.Color
		if active {
			bg = theme.Primary.WithAlpha(0.3)
		} else if selected {
			bg = theme.Surface
		}

		// Get current animated progress value
		progress := row.Progress.Value().Get()

		switch colIndex {
		case 0: // Task name column
			return t.Text{
				Content: row.Name,
				Style: t.Style{
					BackgroundColor: bg,
					ForegroundColor: theme.Text,
				},
			}

		case 1: // Progress column with ProgressBar widget
			var filledColor t.Color
			switch row.Status {
			case StatusComplete:
				filledColor = theme.Success
			case StatusFailed:
				filledColor = theme.Error
			case StatusRunning:
				filledColor = theme.Primary
			default:
				filledColor = theme.TextMuted
			}

			return t.ProgressBar{
				Progress:      progress,
				Width:         t.Flex(1),
				FilledColor:   filledColor,
				UnfilledColor: theme.Surface,
				Style: t.Style{
					BackgroundColor: bg,
				},
			}

		case 2: // Status column with Spinner for running tasks
			switch row.Status {
			case StatusRunning:
				// Show spinner frame next to "Running" text
				return t.Text{
					Content: spinnerFrame + " " + row.Status.String(),
					Style: t.Style{
						ForegroundColor: theme.Primary,
						BackgroundColor: bg,
					},
				}
			case StatusComplete:
				return t.Text{
					Content: "✓ " + row.Status.String(),
					Style: t.Style{
						ForegroundColor: theme.Success,
						BackgroundColor: bg,
					},
				}
			case StatusFailed:
				return t.Text{
					Content: "✗ " + row.Status.String(),
					Style: t.Style{
						ForegroundColor: theme.Error,
						BackgroundColor: bg,
					},
				}
			default:
				return t.Text{
					Content: "○ " + row.Status.String(),
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
						BackgroundColor: bg,
					},
				}
			}

		case 3: // Nested table column
			return t.Table[[]string]{
				State: row.MetricsState,
				Columns: []t.TableColumn{
					{Width: t.Cells(4)},
					{Width: t.Cells(8)},
				},
				ColumnSpacing: 1,
				Style: t.Style{
					BackgroundColor: bg,
				},
			}
		}

		return t.Text{Content: ""}
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	app := NewApp()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
