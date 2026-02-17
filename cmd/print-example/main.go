// print-example demonstrates Terma's static printing API.
//
// This shows how to render widgets to the terminal without running
// a full interactive TUI application - useful for CLI tools, logs,
// progress output, and one-shot renders.
//
// Usage:
//
//	go run ./cmd/print-example
//	go run ./cmd/print-example | cat   # Test plain text fallback
package main

import (
	"fmt"
	"os"

	"github.com/darrenburns/terma"
)

func main() {
	// Example 1: Simple text
	fmt.Println("=== Example 1: Simple Text ===")
	terma.Print(terma.Row{
		Children: []terma.Widget{
			terma.Text{Content: "Hello from Terma's Print API!"},
		},
		Style: terma.Style{
			Padding: terma.EdgeInsetsAll(3),
		},
	})
	fmt.Println()

	// Example 2: Styled text with colors
	fmt.Println("=== Example 2: Styled Text ===")
	terma.PrintWithSize(
		terma.Text{
			Content: "Colored text with style",
			Style: terma.Style{
				ForegroundColor: terma.Hex("#00FF88"),
				Bold:            true,
			},
		},
		40, 1,
	)
	fmt.Println()

	// Example 3: Column layout
	fmt.Println("=== Example 3: Column Layout ===")
	terma.PrintWithSize(
		terma.Column{
			Children: []terma.Widget{
				terma.Text{
					Content: "Header",
					Style: terma.Style{
						ForegroundColor: terma.Hex("#FFD700"),
						Bold:            true,
					},
				},
				terma.Text{Content: "Line 1: First item"},
				terma.Text{Content: "Line 2: Second item"},
				terma.Text{Content: "Line 3: Third item"},
			},
		},
		40, 4,
	)
	fmt.Println()

	// Example 4: Row layout with spacing
	fmt.Println("=== Example 4: Row Layout ===")
	terma.PrintWithSize(
		terma.Row{
			Spacing: 2,
			Children: []terma.Widget{
				terma.Text{
					Content: "[OK]",
					Style:   terma.Style{ForegroundColor: terma.Hex("#00FF00")},
				},
				terma.Text{Content: "All systems operational"},
			},
		},
		40, 1,
	)
	fmt.Println()

	// Example 5: Widget with border and padding
	fmt.Println("=== Example 5: Bordered Widget ===")
	terma.PrintWithSize(
		terma.Column{
			Style: terma.Style{
				Border:          terma.RoundedBorder(terma.Hex("#5588FF")),
				Padding:         terma.EdgeInsetsXY(1, 0),
				ForegroundColor: terma.Hex("#FFFFFF"),
			},
			Children: []terma.Widget{
				terma.Text{
					Content: "Status Report",
					Style:   terma.Style{Bold: true},
				},
				terma.Text{Content: "CPU: 42%"},
				terma.Text{Content: "Memory: 2.1GB"},
				terma.Text{Content: "Disk: 128GB free"},
			},
		},
		30, 6,
	)
	fmt.Println()

	// Example 6: Using RenderToString for logging
	fmt.Println("=== Example 6: RenderToString ===")
	status := terma.RenderToString(
		terma.Row{
			Spacing: 1,
			Children: []terma.Widget{
				terma.Text{
					Content: "●",
					Style:   terma.Style{ForegroundColor: terma.Hex("#00FF00")},
				},
				terma.Text{Content: "Connected"},
			},
		},
		20, 1,
	)
	fmt.Printf("Status string: %s\n", status)
	fmt.Println()

	// Example 7: Complex nested layout
	fmt.Println("=== Example 7: Complex Layout ===")
	terma.PrintWithSize(
		terma.Column{
			Style: terma.Style{
				Border:          terma.RoundedBorder(terma.Hex("#888888")),
				BackgroundColor: terma.Hex("#1a1a2e"),
			},
			Spacing: 1,
			Children: []terma.Widget{
				terma.Row{
					Children: []terma.Widget{
						terma.Text{
							Content: "Dashboard",
							Style: terma.Style{
								Bold:            true,
								ForegroundColor: terma.Hex("#FFD700"),
							},
						},
						terma.Spacer{},
						terma.Text{
							Content: "v1.0.0",
							Style:   terma.Style{ForegroundColor: terma.Hex("#666666")},
						},
					},
				},
				terma.Row{
					Spacing: 4,
					Children: []terma.Widget{
						terma.Column{
							Children: []terma.Widget{
								terma.Text{
									Content: "Requests",
									Style:   terma.Style{ForegroundColor: terma.Hex("#888888")},
								},
								terma.Text{
									Content: "1,234",
									Style: terma.Style{
										Bold:            true,
										ForegroundColor: terma.Hex("#00FF88"),
									},
								},
							},
						},
						terma.Column{
							Children: []terma.Widget{
								terma.Text{
									Content: "Errors",
									Style:   terma.Style{ForegroundColor: terma.Hex("#888888")},
								},
								terma.Text{
									Content: "3",
									Style: terma.Style{
										Bold:            true,
										ForegroundColor: terma.Hex("#FF4444"),
									},
								},
							},
						},
						terma.Column{
							Children: []terma.Widget{
								terma.Text{
									Content: "Latency",
									Style:   terma.Style{ForegroundColor: terma.Hex("#888888")},
								},
								terma.Text{
									Content: "42ms",
									Style: terma.Style{
										Bold:            true,
										ForegroundColor: terma.Hex("#FFFFFF"),
									},
								},
							},
						},
					},
				},
			},
		},
		50, 6,
	)
	fmt.Println()

	// Example 8: PrintWithOptions for full control
	fmt.Println("=== Example 8: PrintWithOptions ===")
	terma.PrintWithOptions(
		terma.Text{
			Content: "Custom options: no trailing newline ->",
			Style:   terma.Style{ForegroundColor: terma.Hex("#FF88FF")},
		},
		terma.PrintOptions{
			Width:           50,
			Height:          1,
			Writer:          os.Stdout,
			TrailingNewline: false,
		},
	)
	fmt.Println("<- end")
	fmt.Println()

	// Example 9: Plain text (no ANSI)
	fmt.Println("=== Example 9: Plain Text Mode ===")
	terma.PrintWithOptions(
		terma.Text{
			Content: "This has no ANSI codes (NoColor: true)",
			Style:   terma.Style{ForegroundColor: terma.Hex("#FF0000")}, // Color ignored
		},
		terma.PrintOptions{
			Width:   50,
			Height:  1,
			Writer:  os.Stdout,
			NoColor: true,
		},
	)
	fmt.Println()

	fmt.Println("Done! Try piping output: go run ./cmd/print-example | cat")
}
