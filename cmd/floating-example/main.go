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

// FloatingDemo demonstrates the Floating widget for popups and modals.
type FloatingDemo struct {
	showDropdown t.Signal[bool]
	showModal    t.Signal[bool]
	statusMsg    t.Signal[string]
}

func NewFloatingDemo() *FloatingDemo {
	return &FloatingDemo{
		showDropdown: t.NewSignal(false),
		showModal:    t.NewSignal(false),
		statusMsg:    t.NewSignal("Click a button to see floating widgets"),
	}
}

func (d *FloatingDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:      "floating-demo-root",
		Spacing: 1,
		Width:   t.Fr(1),
		Height:  t.Fr(1),
		Style: t.Style{
			Padding:         t.EdgeInsetsXY(2, 1),
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Floating Widgets Demo",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Cyan,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},

			// Instructions
			t.Text{
				Spans: t.ParseMarkup("Press [b #00ffff]Escape[/] to dismiss floats • Click outside dropdown to close", t.ThemeData{}),
			},

			// Row with buttons
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					// Dropdown trigger button
					&t.Button{
						ID:    "dropdown-btn",
						Label: "Open Dropdown",
						OnPress: func() {
							d.showDropdown.Set(true)
						},
					},

					// Modal trigger button
					&t.Button{
						ID:    "modal-btn",
						Label: "Open Modal",
						OnPress: func() {
							d.showModal.Set(true)
						},
					},
				},
			},

			// Dropdown menu (anchored to button)
			t.Floating{
				Visible: d.showDropdown.Get(),
				Config: t.FloatConfig{
					AnchorID:  "dropdown-btn",
					Anchor:    t.AnchorBottomLeft,
					OnDismiss: func() { d.showDropdown.Set(false) },
				},
				Child: d.buildDropdownMenu(),
			},

			// Modal dialog (centered)
			t.Floating{
				Visible: d.showModal.Get(),
				Config: t.FloatConfig{
					Position:      t.FloatPositionCenter,
					Modal:         true,
					BackdropColor: theme.Background.WithAlpha(0.2),
					OnDismiss:     func() { d.showModal.Set(false) },
				},
				Child: d.buildModalDialog(),
			},

			// Status message
			t.Text{
				Content: d.statusMsg.Get(),
				Style: t.Style{
					ForegroundColor: t.BrightYellow,
				},
			},

			// Quit instructions
			t.Text{
				Spans: t.ParseMarkup("Press [b #ff5555]Ctrl+C[/] to quit", t.ThemeData{}),
			},
		},
	}
}

// buildDropdownMenu creates the dropdown menu content.
func (d *FloatingDemo) buildDropdownMenu() t.Widget {
	return t.Column{
		Style: t.Style{
			BackgroundColor: t.RGB(40, 40, 40),
			Border:          t.Border{Style: t.BorderSquare, Color: t.BrightBlue},
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			d.menuItem("New File", func() {
				d.statusMsg.Set("Selected: New File")
				d.showDropdown.Set(false)
			}),
			d.menuItem("Open File", func() {
				d.statusMsg.Set("Selected: Open File")
				d.showDropdown.Set(false)
			}),
			d.menuItem("Save", func() {
				d.statusMsg.Set("Selected: Save")
				d.showDropdown.Set(false)
			}),
			d.menuItem("Exit", func() {
				d.statusMsg.Set("Selected: Exit")
				d.showDropdown.Set(false)
			}),
		},
	}
}

// menuItem creates a clickable menu item.
func (d *FloatingDemo) menuItem(label string, onClick func()) t.Widget {
	return t.Text{
		Content: " " + label + " ",
		Click:   onClick,
		Style: t.Style{
			ForegroundColor: t.White,
		},
	}
}

// buildModalDialog creates the modal dialog content.
func (d *FloatingDemo) buildModalDialog() t.Widget {
	return t.Column{
		Style: t.Style{
			BackgroundColor: t.RGB(50, 50, 60),
			Border: t.Border{
				Style:       t.BorderRounded,
				Color:       t.BrightMagenta,
				Decorations: []t.BorderDecoration{t.BorderTitleCenter(" Confirm Action ")},
			},
			Padding: t.EdgeInsetsAll(2),
		},
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Are you sure you want to proceed?",
				Style: t.Style{
					ForegroundColor: t.White,
				},
			},
			t.Text{
				Content: "This action cannot be undone.",
				Style: t.Style{
					ForegroundColor: t.BrightBlack,
				},
			},
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					&t.Button{
						ID:    "cancel-btn",
						Label: "Cancel",
						OnPress: func() {
							d.statusMsg.Set("Action cancelled")
							d.showModal.Set(false)
						},
					},
					&t.Button{
						ID:    "confirm-btn",
						Label: "Confirm",
						OnPress: func() {
							d.statusMsg.Set("Action confirmed!")
							d.showModal.Set(false)
						},
					},
				},
			},
		},
	}
}

func main() {
	app := NewFloatingDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
