package main

import (
	"log"

	t "github.com/darrenburns/terma"
)

// App demonstrates the Spacer widget.
type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Spacing: 1,
		Style:   t.Style{Padding: t.EdgeInsetsAll(1)},
		Children: []t.Widget{
			t.Text{Content: "=== Spacer Widget Demo ===", Style: t.Style{ForegroundColor: theme.Primary}},

			// Section 1: Basic Spacer in Row
			t.Text{Content: "1. Push items apart in a Row (Spacer{}):", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Row{
				Style: t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					t.Text{Content: "Left", Style: t.Style{ForegroundColor: theme.Primary.AutoText(), BackgroundColor: theme.Primary, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{},
					t.Text{Content: "Right", Style: t.Style{ForegroundColor: theme.Accent.AutoText(), BackgroundColor: theme.Accent, Padding: t.EdgeInsetsXY(1, 0)}},
				},
			},

			// Section 2: Multiple Spacers
			t.Text{Content: "2. Three items with equal spacing:", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Row{
				Style: t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					t.Text{Content: "A", Style: t.Style{ForegroundColor: theme.Primary.AutoText(), BackgroundColor: theme.Primary, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{},
					t.Text{Content: "B", Style: t.Style{ForegroundColor: theme.Accent.AutoText(), BackgroundColor: theme.Accent, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{},
					t.Text{Content: "C", Style: t.Style{ForegroundColor: theme.Success.AutoText(), BackgroundColor: theme.Success, Padding: t.EdgeInsetsXY(1, 0)}},
				},
			},

			// Section 3: Proportional Spacers
			t.Text{Content: "3. Proportional spacing (Flex(1) vs Flex(2)):", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Row{
				Style: t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					t.Text{Content: "A", Style: t.Style{ForegroundColor: theme.Primary.AutoText(), BackgroundColor: theme.Primary, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{Width: t.Flex(1)},
					t.Text{Content: "B", Style: t.Style{ForegroundColor: theme.Accent.AutoText(), BackgroundColor: theme.Accent, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{Width: t.Flex(2)},
					t.Text{Content: "C", Style: t.Style{ForegroundColor: theme.Success.AutoText(), BackgroundColor: theme.Success, Padding: t.EdgeInsetsXY(1, 0)}},
				},
			},

			// Section 4: Fixed-size Spacer
			t.Text{Content: "4. Fixed gap with Cells(5):", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Row{
				Style: t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					t.Text{Content: "Left", Style: t.Style{ForegroundColor: theme.Primary.AutoText(), BackgroundColor: theme.Primary, Padding: t.EdgeInsetsXY(1, 0)}},
					t.Spacer{Width: t.Cells(5)},
					t.Text{Content: "Right", Style: t.Style{ForegroundColor: theme.Accent.AutoText(), BackgroundColor: theme.Accent, Padding: t.EdgeInsetsXY(1, 0)}},
				},
			},

			// Section 5: Vertical Spacer in Column
			t.Text{Content: "5. Vertical spacer in Column:", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Column{
				Height: t.Cells(6),
				Style:  t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					t.Text{Content: "Top", Style: t.Style{ForegroundColor: theme.Primary.AutoText(), BackgroundColor: theme.Primary}},
					t.Spacer{},
					t.Text{Content: "Bottom", Style: t.Style{ForegroundColor: theme.Accent.AutoText(), BackgroundColor: theme.Accent}},
				},
			},

			t.Text{Content: "Press Ctrl+C to quit", Style: t.Style{ForegroundColor: theme.TextMuted}},
		},
	}
}

func main() {
	app := &App{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
