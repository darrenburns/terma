package main

import (
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
	t.InitDebug()
}

type ColumnLayoutDebug struct{}

func (c *ColumnLayoutDebug) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:     "root",
		Height: t.Flex(1),
		Width:  t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Title
			t.Text{
				Content: "Column Layout Debug - Testing Various Height Configurations",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Main row containing all test columns
			t.Row{
				Height:  t.Flex(1),
				Spacing: 2,
				Children: []t.Widget{
					// Column 1: Fixed heights (4 cells each)
					t.Column{
						Height: t.Flex(1),
						Spacing: 1,
						Style: t.Style{
							Border: t.RoundedBorder(theme.Info,
								t.BorderTitle("Fixed: Cells(4)"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Box 1",
								Height:  t.Cells(4),
								Style: t.Style{
									BackgroundColor: theme.Error,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 2",
								Height:  t.Cells(4),
								Style: t.Style{
									BackgroundColor: theme.Warning,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 3",
								Height:  t.Cells(4),
								Style: t.Style{
									BackgroundColor: theme.Success,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
						},
					},

					// Column 2: All Flex(1)
					t.Column{
						Height: t.Flex(1),
						Spacing: 1,
						Style: t.Style{
							Border: t.RoundedBorder(theme.Secondary,
								t.BorderTitle("All Flex(1)"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Box 1",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Error,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 2",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Warning,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 3",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Success,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
						},
					},

					// Column 3: Flex(2), Flex(1), Flex(1)
					t.Column{
						Height: t.Flex(1),
						Spacing: 1,
						Style: t.Style{
							Border: t.RoundedBorder(theme.Warning,
								t.BorderTitle("Flex(2), Flex(1), Flex(1)"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Box 1 (Flex 2)",
								Height:  t.Flex(2),
								Style: t.Style{
									BackgroundColor: theme.Error,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 2 (Flex 1)",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Warning,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Box 3 (Flex 1)",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Success,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
						},
					},

					// Column 4: Flex(1) + wrapping text
					t.Column{
						Height: t.Flex(1),
						Spacing: 1,
						Style: t.Style{
							Border: t.RoundedBorder(theme.Success,
								t.BorderTitle("Flex(1) + Text"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Flex(1) Box",
								Height:  t.Flex(1),
								Style: t.Style{
									BackgroundColor: theme.Error,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "This is a text widget with wrapping enabled. It should wrap to multiple lines when the content exceeds the available width. The height should be Auto (content-based).",
								Wrap:    t.WrapSoft,
								Style: t.Style{
									BackgroundColor: theme.Surface,
									ForegroundColor: theme.Text,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	app := &ColumnLayoutDebug{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
