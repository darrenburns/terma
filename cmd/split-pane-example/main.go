package main

import (
	"log"
	"strings"

	t "github.com/darrenburns/terma"
)

type SplitPaneDemo struct {
	mainState  *t.SplitPaneState
	rightState *t.SplitPaneState
}

func NewSplitPaneDemo() *SplitPaneDemo {
	return &SplitPaneDemo{
		mainState:  t.NewSplitPaneState(0.35),
		rightState: t.NewSplitPaneState(0.6),
	}
}

func (d *SplitPaneDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (d *SplitPaneDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Dock{
		Top: []t.Widget{
			t.Text{
				Content: " SplitPane Demo ",
				Width:   t.Flex(1),
				Style: t.Style{
					ForegroundColor: ctx.Theme().Background,
					BackgroundColor: ctx.Theme().Primary,
				},
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: ctx.Theme().Surface,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},
		},
		Body: d.buildBody(ctx),
	}
}

func (d *SplitPaneDemo) buildBody(ctx t.BuildContext) t.Widget {
	return t.SplitPane{
		ID:                "main-split",
		State:             d.mainState,
		Orientation:       t.SplitHorizontal,
		DividerSize:       1,
		DividerBackground: ctx.Theme().Surface,
		DividerFocusForeground: t.NewGradient(
			ctx.Theme().Primary,
			ctx.Theme().Accent,
		).WithAngle(0),
		MinPaneSize: 8,
		First:       d.buildLeftPanel(ctx),
		Second:      d.buildRightPanel(ctx),
	}
}

func (d *SplitPaneDemo) buildLeftPanel(ctx t.BuildContext) t.Widget {
	content := strings.Join([]string{
		"Left pane",
		"",
		"- Drag the divider with the mouse",
		"- Use arrow keys to resize",
		"- Keybinds work even when children are focused",
	}, "\n")

	return t.Text{
		Content: content,
		Wrap:    t.WrapSoft,
		Style: t.Style{
			Width:           t.Flex(1),
			BackgroundColor: ctx.Theme().Surface,
			Padding:         t.EdgeInsetsXY(1, 1),
		},
	}
}

func (d *SplitPaneDemo) buildRightPanel(ctx t.BuildContext) t.Widget {
	return t.SplitPane{
		ID:                "right-split",
		State:             d.rightState,
		Orientation:       t.SplitVertical,
		DividerSize:       1,
		DividerBackground: ctx.Theme().Surface,
		MinPaneSize:       4,
		First: t.Text{
			Content: "Top pane (nested)",
			Width:   t.Flex(1),
			Style: t.Style{
				BackgroundColor: ctx.Theme().Surface,
				Padding:         t.EdgeInsetsXY(1, 1),
			},
		},
		Second: t.Text{
			Content: "Bottom pane (nested)",
			Width:   t.Flex(1),
			Style: t.Style{
				BackgroundColor: ctx.Theme().Surface,
				Padding:         t.EdgeInsetsXY(1, 1),
			},
		},
	}
}

func main() {
	app := NewSplitPaneDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
