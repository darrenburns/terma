package main

import (
	"fmt"
	"os"

	t "github.com/darrenburns/terma"
)

type App struct {
	spinner    *t.SpinnerState
	input      *t.TextInputState
	autoInput  *t.TextInputState
	running    t.Signal[bool]
	showBanner t.Signal[bool]
}

func NewApp() *App {
	app := &App{
		spinner:    t.NewSpinnerState(t.SpinnerDots),
		input:      t.NewTextInputState("edit me without rebuilding"),
		autoInput:  t.NewTextInputState("auto"),
		running:    t.NewSignal(true),
		showBanner: t.NewSignal(false),
	}
	app.spinner.Start()
	return app
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{
			Key:  "ctrl+g",
			Name: "Toggle spinner",
			Action: func() {
				if a.running.Get() {
					a.spinner.Stop()
					a.running.Set(false)
					return
				}
				a.spinner.Start()
				a.running.Set(true)
			},
		},
		{
			Key:  "ctrl+b",
			Name: "Toggle build-phase banner",
			Action: func() {
				a.showBanner.Set(!a.showBanner.Get())
			},
		},
		{
			Key:    "ctrl+q",
			Name:   "Quit",
			Action: t.Quit,
		},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	running := a.running.Get()
	showBanner := a.showBanner.Get()
	overlayEnabled := os.Getenv("TERMA_DEBUG_OVERLAY") != ""

	children := []t.Widget{
		t.Text{
			Content: "Phase-Aware Invalidation Demo",
			Style: t.Style{
				Bold:            true,
				ForegroundColor: theme.Primary,
			},
		},
		t.Text{
			Content: "The spinner is animating continuously. With the debug overlay enabled, its ticks should show mode=partial, build=0, layout=0.",
			Wrap:    t.WrapSoft,
			Style:   t.Style{ForegroundColor: theme.TextMuted},
		},
		t.Text{
			Content: "Type into the inputs below and move the cursor with the arrow keys. The fixed-width input should stay on the partial repaint path, while the auto-width input should force relayout when its content width changes.",
			Wrap:    t.WrapSoft,
			Style:   t.Style{ForegroundColor: theme.TextMuted},
		},
		overlayNotice(theme, overlayEnabled),
		buildPhaseBanner(theme, showBanner),
		panel(theme, "Fixed-width TextInput", t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Text{
					Content: "This input is fixed-width, so typing and cursor movement should not require a relayout of the app.",
					Wrap:    t.WrapSoft,
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.TextInput{
					ID:          "demo-input",
					State:       a.input,
					Placeholder: "Type here",
					Width:       t.Cells(32),
					Style: t.Style{
						Border:          t.RoundedBorder(theme.Border),
						Padding:         t.EdgeInsetsXY(1, 0),
						BackgroundColor: theme.Surface2,
					},
				},
				t.Text{
					Content: "Expected while editing: last frame stays on partial repaint, rebuilt=0, relaid out=0, repaint area stays around the input.",
					Wrap:    t.WrapSoft,
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
			},
		}),
		panel(theme, "Auto-width TextInput", t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Text{
					Content: "This input uses Auto width. Cursor movement should stay partial, but adding or removing characters that change its intrinsic width should force a full render.",
					Wrap:    t.WrapSoft,
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.TextInput{
					ID:          "demo-auto-input",
					State:       a.autoInput,
					Placeholder: "Auto width",
					Width:       t.Auto,
					Style: t.Style{
						Border:          t.RoundedBorder(theme.Border),
						Padding:         t.EdgeInsetsXY(1, 0),
						BackgroundColor: theme.Surface2,
					},
				},
				t.Text{
					Content: "Expected while editing: left/right cursor moves can stay partial, but any text edit that changes the input's width should switch the overlay to a full render.",
					Wrap:    t.WrapSoft,
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
			},
		}),
		t.Row{
			Spacing: 2,
			Width:   t.Flex(1),
			Children: []t.Widget{
				panel(theme, "Static tree", t.Column{
					Spacing: 1,
					Children: []t.Widget{
						statusLine("Root widgets stay in place while the spinner ticks.", theme),
						statusLine("Only the damaged region should repaint.", theme),
						statusLine("Press `ctrl+b` to force a structural change and compare.", theme),
						t.Text{Content: ""},
						staticLog(theme),
					},
				}),
				panel(theme, "Spinner probe", t.Column{
					Spacing: 1,
					Children: []t.Widget{
						t.Row{
							Spacing: 1,
							Children: []t.Widget{
								t.Spinner{
									ID:    "demo-spinner",
									State: a.spinner,
									Style: t.Style{ForegroundColor: theme.Success},
								},
								t.Text{
									Content: spinnerStatus(running),
									Style: t.Style{
										Bold:            true,
										ForegroundColor: theme.Text,
									},
								},
							},
						},
						t.Text{
							Content: "Expected while running: partial frame, zero builds, zero layouts, small damage union around the spinner.",
							Wrap:    t.WrapSoft,
							Style:   t.Style{ForegroundColor: theme.TextMuted},
						},
						t.Text{
							Content: "Press ctrl+g to stop and restart the spinner. Press ctrl+b to toggle a build-phase banner.",
							Wrap:    t.WrapSoft,
							Style:   t.Style{ForegroundColor: theme.TextMuted},
						},
						t.Text{Content: ""},
						meterRow(theme, "Static panel A", 82),
						meterRow(theme, "Static panel B", 47),
						meterRow(theme, "Static panel C", 61),
					},
				}),
			},
		},
	}

	return t.Dock{
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},
		},
		Body: t.Column{
			Spacing: 1,
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Style: t.Style{
				Padding:         t.EdgeInsetsAll(1),
				BackgroundColor: theme.Background,
			},
			Children: children,
		},
	}
}

func buildPhaseBanner(theme t.ThemeData, visible bool) t.Widget {
	if !visible {
		return t.Text{
			Content: "Build-phase banner is hidden. Press `ctrl+b` to insert a widget from Build() and force a full render.",
			Wrap:    t.WrapSoft,
			Style: t.Style{
				ForegroundColor: theme.Surface2.AutoText(),
				BackgroundColor: theme.Surface2,
				Padding:         t.EdgeInsetsXY(1, 0),
			},
		}
	}

	return t.Text{
		Content: "Build-phase banner is visible. This widget was conditionally inserted from Build(), so toggling it should produce a full render.",
		Wrap:    t.WrapSoft,
		Style: t.Style{
			ForegroundColor: theme.Warning.AutoText(),
			BackgroundColor: theme.Warning,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
	}
}

func overlayNotice(theme t.ThemeData, enabled bool) t.Widget {
	if enabled {
		return t.Text{
			Content: "Debug overlay enabled. Watch the top rows for frame mode and build/layout counts.",
			Style: t.Style{
				ForegroundColor: theme.Success.AutoText(),
				BackgroundColor: theme.Success,
				Padding:         t.EdgeInsetsXY(1, 0),
			},
		}
	}

	return t.Text{
		Content: "Run this demo with TERMA_DEBUG_OVERLAY=1 to see mode, build/layout/paint counts, and damaged region details.",
		Wrap:    t.WrapSoft,
		Style: t.Style{
			ForegroundColor: theme.Accent.AutoText(),
			BackgroundColor: theme.Accent,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
	}
}

func panel(theme t.ThemeData, title string, child t.Widget) t.Widget {
	return t.Column{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Border:          t.RoundedBorder(theme.Border, t.BorderTitle(title)),
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{child},
	}
}

func statusLine(content string, theme t.ThemeData) t.Widget {
	return t.Text{
		Content: content,
		Wrap:    t.WrapSoft,
		Style:   t.Style{ForegroundColor: theme.TextMuted},
	}
}

func staticLog(theme t.ThemeData) t.Widget {
	lines := make([]t.Widget, 0, 10)
	for i := 1; i <= 10; i++ {
		lines = append(lines, t.Text{
			Content: fmt.Sprintf("Static row %02d  |  cache warm  |  unchanged subtree", i),
			Style:   t.Style{ForegroundColor: theme.Text},
		})
	}
	return t.Column{
		Spacing:  0,
		Children: lines,
	}
}

func meterRow(theme t.ThemeData, label string, percent int) t.Widget {
	return t.Column{
		Spacing: 0,
		Children: []t.Widget{
			t.Text{
				Content: label,
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.ProgressBar{
				Progress: float64(percent) / 100.0,
				Style: t.Style{
					Width:           t.Cells(24),
					ForegroundColor: theme.Primary,
				},
			},
		},
	}
}

func spinnerStatus(running bool) string {
	if running {
		return "Spinner running"
	}
	return "Spinner stopped"
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		panic(err)
	}
}
