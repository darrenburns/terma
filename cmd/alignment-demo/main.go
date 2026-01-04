package main

import (
	"fmt"
	"log"

	t "terma"
)

// Alignment name helpers
var mainAxisNames = []string{"Start", "Center", "End"}
var crossAxisNames = []string{"Start", "Center", "End"}

// App demonstrates MainAlign and CrossAlign on Row and Column containers.
type App struct {
	rowMainAlign    t.Signal[t.MainAxisAlign]
	rowCrossAlign   t.Signal[t.CrossAxisAlign]
	colMainAlign    t.Signal[t.MainAxisAlign]
	colCrossAlign   t.Signal[t.CrossAxisAlign]
	activeContainer t.Signal[int] // 0 = Row, 1 = Column
}

func (a *App) IsFocusable() bool { return true }
func (a *App) OnKey(event t.KeyEvent) bool { return false }

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "m", Name: "Main Axis", Action: a.cycleMainAlign},
		{Key: "c", Name: "Cross Axis", Action: a.cycleCrossAlign},
		{Key: " ", Name: "Switch", Action: a.switchContainer},
	}
}

func (a *App) cycleMainAlign() {
	if a.activeContainer.Get() == 0 {
		a.rowMainAlign.Update(func(v t.MainAxisAlign) t.MainAxisAlign {
			return (v + 1) % 3
		})
	} else {
		a.colMainAlign.Update(func(v t.MainAxisAlign) t.MainAxisAlign {
			return (v + 1) % 3
		})
	}
}

func (a *App) cycleCrossAlign() {
	// Cycle through Start(1), Center(2), End(3) - skipping Stretch(0)
	cycle := func(v t.CrossAxisAlign) t.CrossAxisAlign {
		next := v + 1
		if next > t.CrossAxisEnd {
			next = t.CrossAxisStart
		}
		return next
	}
	if a.activeContainer.Get() == 0 {
		a.rowCrossAlign.Update(cycle)
	} else {
		a.colCrossAlign.Update(cycle)
	}
}

func (a *App) switchContainer() {
	a.activeContainer.Update(func(v int) int {
		return (v + 1) % 2
	})
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	rowMain := a.rowMainAlign.Get()
	rowCross := a.rowCrossAlign.Get()
	colMain := a.colMainAlign.Get()
	colCross := a.colCrossAlign.Get()
	active := a.activeContainer.Get()

	// Container names for display
	containerName := "Row"
	if active == 1 {
		containerName = "Column"
	}

	// Current alignment info
	var mainName, crossName string
	if active == 0 {
		mainName = mainAxisNames[rowMain]
		crossName = crossAxisNames[rowCross-1] // -1 because we skip Stretch(0)
	} else {
		mainName = mainAxisNames[colMain]
		crossName = crossAxisNames[colCross-1] // -1 because we skip Stretch(0)
	}

	return t.Column{
		Height: t.Fr(1),
		Children: []t.Widget{
			// Keybind bar at top
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: t.Hex("#1a1a2e"),
					Padding:         t.EdgeInsets{Left: 1, Right: 1},
				},
			},

			// Title
			t.Text{
				Content: "=== Alignment Demo ===",
				Style:   t.Style{Padding: t.EdgeInsets{Top: 1, Bottom: 1, Left: 1}},
			},

			// Status line
			t.Text{
				Content: fmt.Sprintf("Active: %s | MainAlign: %s | CrossAlign: %s",
					containerName, mainName, crossName),
				Style: t.Style{
					ForegroundColor: t.Cyan,
					Padding:         t.EdgeInsets{Left: 1, Bottom: 1},
				},
			},

			// Demo containers side by side
			t.Row{
				Height:  t.Fr(1),
				Spacing: 2,
				Style:   t.Style{Padding: t.EdgeInsets{Left: 1, Right: 1}},
				Children: []t.Widget{
					// Row demo
					a.buildRowDemo(ctx, rowMain, rowCross, active == 0),
					// Column demo
					a.buildColumnDemo(ctx, colMain, colCross, active == 1),
				},
			},

			// Instructions
			t.Text{
				Content: "Press 'm' to cycle main axis, 'c' to cycle cross axis, 'space' to switch container",
				Style: t.Style{
					ForegroundColor: t.BrightBlack,
					Padding:         t.EdgeInsets{Left: 1, Top: 1, Bottom: 1},
				},
			},
		},
	}
}

func (a *App) buildRowDemo(ctx t.BuildContext, mainAlign t.MainAxisAlign, crossAlign t.CrossAxisAlign, isActive bool) t.Widget {
	borderColor := t.BrightBlack
	if isActive {
		borderColor = t.Cyan
	}

	return t.Column{
		Width: t.Fr(1),
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("Row (main=horiz, cross=vert)"),
				Style:   t.Style{ForegroundColor: borderColor},
			},
			t.Row{
				Width:      t.Fr(1),
				Height:    t.Fr(1),
				MainAlign:  mainAlign,
				CrossAlign: crossAlign,
				Style: t.Style{
					BackgroundColor: t.Hex("#1e1e2e"),
					Border:          t.Border{Style: t.BorderRounded, Color: borderColor},
				},
				Children: []t.Widget{
					t.Column{
						Style: t.Style{
							BackgroundColor: t.Red,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "A"}},
					},
					t.Column{
						Style: t.Style{
							BackgroundColor: t.Green,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "BB", Style: t.Style{ForegroundColor: t.Black}}},
					},
					t.Column{
						Style: t.Style{
							BackgroundColor: t.Blue,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "CCC"}},
					},
				},
			},
		},
	}
}

func (a *App) buildColumnDemo(ctx t.BuildContext, mainAlign t.MainAxisAlign, crossAlign t.CrossAxisAlign, isActive bool) t.Widget {
	borderColor := t.BrightBlack
	if isActive {
		borderColor = t.Cyan
	}

	return t.Column{
		Width: t.Fr(1),
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("Column (main=vert, cross=horiz)"),
				Style:   t.Style{ForegroundColor: borderColor},
			},
			t.Column{
				Width:      t.Fr(1),
				Height:    t.Fr(1),
				MainAlign:  mainAlign,
				CrossAlign: crossAlign,
				Style: t.Style{
					BackgroundColor: t.Hex("#1e1e2e"),
					Border:          t.Border{Style: t.BorderRounded, Color: borderColor},
				},
				Children: []t.Widget{
					t.Row{
						Style: t.Style{
							BackgroundColor: t.Red,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "A"}},
					},
					t.Row{
						Style: t.Style{
							BackgroundColor: t.Green,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "BB", Style: t.Style{ForegroundColor: t.Black}}},
					},
					t.Row{
						Style: t.Style{
							BackgroundColor: t.Blue,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						Children: []t.Widget{t.Text{Content: "CCC"}},
					},
				},
			},
		},
	}
}

func main() {
	app := &App{
		rowMainAlign:    t.NewSignal(t.MainAxisStart),
		rowCrossAlign:   t.NewSignal(t.CrossAxisStart),
		colMainAlign:    t.NewSignal(t.MainAxisStart),
		colCrossAlign:   t.NewSignal(t.CrossAxisStart),
		activeContainer: t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
