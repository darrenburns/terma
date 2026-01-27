package main

import (
	"fmt"
	"log"

	t "terma"
)

// WidgetShowcase demonstrates all Terma widgets in a single scrollable view
// with easy theme switching.
type WidgetShowcase struct {
	// State for widgets
	listState      *t.ListState[string]
	tableState     *t.TableState[[]string]
	textInputState *t.TextInputState
	textAreaState  *t.TextAreaState
	scrollState    *t.ScrollState

	// Theme management
	themeIndex t.Signal[int]
	themeNames []string
}

func NewWidgetShowcase() *WidgetShowcase {
	// All available themes
	themeNames := []string{
		// Dark themes
		t.ThemeNameRosePine,
		t.ThemeNameDracula,
		t.ThemeNameTokyoNight,
		t.ThemeNameCatppuccin,
		t.ThemeNameGruvbox,
		t.ThemeNameNord,
		t.ThemeNameSolarized,
		t.ThemeNameKanagawa,
		t.ThemeNameMonokai,
		// Light themes
		t.ThemeNameRosePineDawn,
		t.ThemeNameDraculaLight,
		t.ThemeNameTokyoNightDay,
		t.ThemeNameCatppuccinLatte,
		t.ThemeNameGruvboxLight,
		t.ThemeNameNordLight,
		t.ThemeNameSolarizedLight,
		t.ThemeNameKanagawaLotus,
		t.ThemeNameMonokaiLight,
	}
	textAreaState := t.NewTextAreaState("")
	textAreaState.WrapMode.Set(t.WrapSoft)

	return &WidgetShowcase{
		listState: t.NewListState([]string{
			"First item",
			"Second item",
			"Third item",
			"Fourth item",
		}),
		tableState: t.NewTableState([][]string{
			{"Alpha", "Running", "Healthy"},
			{"Bravo", "Stopped", "Warning"},
			{"Charlie", "Running", "Healthy"},
		}),
		textInputState: t.NewTextInputState(""),
		textAreaState:  textAreaState,
		scrollState:    t.NewScrollState(),
		themeIndex:     t.NewSignal(0),
		themeNames:     themeNames,
	}
}

func (w *WidgetShowcase) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := w.themeIndex.Get()
	currentTheme := w.themeNames[themeIdx]

	return t.Dock{
		ID: "widget-showcase-root",
		Top: []t.Widget{
			w.buildHeader(theme, currentTheme),
		},
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Scrollable{
			State:  w.scrollState,
			Height: t.Flex(1),
			Child: t.Column{
				Spacing: 2,
				Style: t.Style{
					Padding:         t.EdgeInsetsAll(2),
					BackgroundColor: theme.Background,
				},
				Children: []t.Widget{
					// Text styles section
					w.buildSection(theme, "Text Styles", w.buildTextSection(theme)),

					// Button section
					w.buildSection(theme, "Button", w.buildButtonSection(theme)),

					// ProgressBar section
					w.buildSection(theme, "ProgressBar", w.buildProgressBarSection(theme)),

					// List section
					w.buildSection(theme, "List", w.buildListSection(ctx, theme)),

					// Table section
					w.buildSection(theme, "Table", w.buildTableSection(ctx, theme)),

					// TextInput section
					w.buildSection(theme, "TextInput", w.buildTextInputSection(theme)),

					// TextArea section
					w.buildSection(theme, "TextArea", w.buildTextAreaSection(theme)),

					// Color palette section
					w.buildSection(theme, "Theme Colors", w.buildColorPalette(theme)),

					// Feedback colors section
					w.buildSection(theme, "Feedback Colors", w.buildFeedbackColors(theme)),

					// Bottom spacer
					t.Spacer{Height: t.Cells(2)},
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildHeader(theme t.ThemeData, currentTheme string) t.Widget {
	return t.Column{
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Row{
				Children: []t.Widget{
					t.Text{
						Content: "Widget Showcase",
						Style: t.Style{
							ForegroundColor: theme.Primary,
							Bold:            true,
						},
					},
					t.Spacer{Width: t.Flex(1)},
					t.Text{
						Content: "Theme: ",
						Style:   t.Style{ForegroundColor: theme.TextMuted},
					},
					t.Text{
						Content: currentTheme,
						Style: t.Style{
							ForegroundColor: theme.Accent,
							Bold:            true,
						},
					},
				},
			},
			t.Text{
				Content: "Press [t] next theme, [T] prev theme, [Tab] cycle focus, [q] quit",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (w *WidgetShowcase) buildSection(theme t.ThemeData, title string, content t.Widget) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: title,
				Style: t.Style{
					ForegroundColor: theme.Primary,
					Bold:            true,
				},
			},
			t.Column{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{content},
			},
		},
	}
}

func (w *WidgetShowcase) buildTextSection(theme t.ThemeData) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{Content: "Normal text", Style: t.Style{ForegroundColor: theme.Text}},
			t.Text{Content: "Muted text", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Text{Content: "Bold text", Style: t.Style{ForegroundColor: theme.Text, Bold: true}},
			t.Text{Content: "Italic text", Style: t.Style{ForegroundColor: theme.Text, Italic: true}},
			t.Text{Content: "Underlined text", Style: t.Style{ForegroundColor: theme.Text, Underline: t.UnderlineSingle}},
			t.Text{Content: "Strikethrough text", Style: t.Style{ForegroundColor: theme.Text, Strikethrough: true}},
			t.Text{Content: "Primary colored", Style: t.Style{ForegroundColor: theme.Primary}},
			t.Text{Content: "Accent colored", Style: t.Style{ForegroundColor: theme.Accent}},
			t.Text{Content: "Link styled", Style: t.Style{ForegroundColor: theme.Link, Underline: t.UnderlineSingle}},
		},
	}
}

func (w *WidgetShowcase) buildButtonSection(theme t.ThemeData) t.Widget {
	return t.Row{
		Spacing: 2,
		Children: []t.Widget{
			&t.Button{
				ID:    "btn-primary",
				Label: " Primary Button ",
			},
			&t.Button{
				ID:    "btn-secondary",
				Label: " Secondary ",
				Style: t.Style{
					BackgroundColor: theme.Secondary,
					ForegroundColor: theme.TextOnSecondary,
				},
			},
			&t.Button{
				ID:    "btn-accent",
				Label: " Accent ",
				Style: t.Style{
					BackgroundColor: theme.Accent,
					ForegroundColor: theme.TextOnAccent,
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildProgressBarSection(theme t.ThemeData) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "Primary:  ", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.ProgressBar{Progress: 0.65, Width: t.Cells(30), FilledColor: theme.Primary},
					t.Text{Content: " 65%"},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "Accent:   ", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.ProgressBar{Progress: 0.45, Width: t.Cells(30), FilledColor: theme.Accent},
					t.Text{Content: " 45%"},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "Success:  ", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.ProgressBar{Progress: 1.0, Width: t.Cells(30), FilledColor: theme.Success},
					t.Text{Content: " 100%"},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "Warning:  ", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.ProgressBar{Progress: 0.75, Width: t.Cells(30), FilledColor: theme.Warning},
					t.Text{Content: " 75%"},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "Error:    ", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.ProgressBar{Progress: 0.25, Width: t.Cells(30), FilledColor: theme.Error},
					t.Text{Content: " 25%"},
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildListSection(ctx t.BuildContext, theme t.ThemeData) t.Widget {
	isFocused := ctx.IsFocused(t.List[string]{ID: "showcase-list"})
	focusLabel := "unfocused"
	if isFocused {
		focusLabel = "focused"
	}

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("List widget (%s) - cursor highlight only shows when focused", focusLabel),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.List[string]{
				ID:          "showcase-list",
				State:       w.listState,
				MultiSelect: true,
				Height:      t.Cells(4),
			},
		},
	}
}

func (w *WidgetShowcase) buildTableSection(ctx t.BuildContext, theme t.ThemeData) t.Widget {
	isFocused := ctx.IsFocused(t.Table[[]string]{ID: "showcase-table"})
	focusLabel := "unfocused"
	if isFocused {
		focusLabel = "focused"
	}

	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface2,
		Bold:            true,
	}

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("Table widget (%s) - cursor highlight only shows when focused", focusLabel),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.Table[[]string]{
				ID:            "showcase-table",
				State:         w.tableState,
				SelectionMode: t.TableSelectionRow,
				Columns: []t.TableColumn{
					{Width: t.Cells(12), Header: t.Text{Content: "Name", Style: headerStyle}},
					{Width: t.Cells(12), Header: t.Text{Content: "Status", Style: headerStyle}},
					{Width: t.Cells(12), Header: t.Text{Content: "Health", Style: headerStyle}},
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildTextInputSection(theme t.ThemeData) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Single-line text input - cursor only visible when focused",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.TextInput{
				ID:          "showcase-textinput",
				State:       w.textInputState,
				Placeholder: "Type something here...",
				Width:       t.Cells(40),
				Style: t.Style{
					BackgroundColor: theme.Background,
					ForegroundColor: theme.Text,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildTextAreaSection(theme t.ThemeData) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Multi-line text area - cursor only visible when focused",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.TextArea{
				ID:          "showcase-textarea",
				State:       w.textAreaState,
				Placeholder: "Type multiple lines here...",
				Width:       t.Auto,
				Height:      t.Cells(4),
				Style: t.Style{
					BackgroundColor: theme.Background,
					ForegroundColor: theme.Text,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
	}
}

func (w *WidgetShowcase) buildColorPalette(theme t.ThemeData) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					w.colorSwatch("Primary", theme.Primary, theme.TextOnPrimary),
					w.colorSwatch("Secondary", theme.Secondary, theme.TextOnSecondary),
					w.colorSwatch("Accent", theme.Accent, theme.TextOnAccent),
				},
			},
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					w.colorSwatch("Surface", theme.Surface, theme.Text),
					w.colorSwatch("Surface2", theme.Surface2, theme.Text),
					w.colorSwatch("Surface3", theme.Surface3, theme.Text),
				},
			},
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					w.colorSwatch("ActiveCursor", theme.ActiveCursor, theme.SelectionText),
					w.colorSwatch("Border", theme.Border, theme.Text),
					w.colorSwatch("FocusRing", theme.FocusRing, theme.TextOnPrimary),
				},
			},
		},
	}
}

func (w *WidgetShowcase) colorSwatch(name string, bg, fg t.Color) t.Widget {
	return t.Text{
		Content: fmt.Sprintf(" %-10s ", name),
		Style: t.Style{
			BackgroundColor: bg,
			ForegroundColor: fg,
		},
	}
}

func (w *WidgetShowcase) buildFeedbackColors(theme t.ThemeData) t.Widget {
	return t.Row{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: " Error ",
				Style: t.Style{
					BackgroundColor: theme.Error,
					ForegroundColor: theme.TextOnError,
				},
			},
			t.Text{
				Content: " Warning ",
				Style: t.Style{
					BackgroundColor: theme.Warning,
					ForegroundColor: theme.TextOnWarning,
				},
			},
			t.Text{
				Content: " Success ",
				Style: t.Style{
					BackgroundColor: theme.Success,
					ForegroundColor: theme.TextOnSuccess,
				},
			},
			t.Text{
				Content: " Info ",
				Style: t.Style{
					BackgroundColor: theme.Info,
					ForegroundColor: theme.TextOnInfo,
				},
			},
		},
	}
}

func (w *WidgetShowcase) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Next theme", Action: w.nextTheme},
		{Key: "T", Name: "Prev theme", Action: w.prevTheme},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (w *WidgetShowcase) nextTheme() {
	w.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(w.themeNames)
		t.SetTheme(w.themeNames[next])
		return next
	})
}

func (w *WidgetShowcase) prevTheme() {
	w.themeIndex.Update(func(i int) int {
		prev := i - 1
		if prev < 0 {
			prev = len(w.themeNames) - 1
		}
		t.SetTheme(w.themeNames[prev])
		return prev
	})
}

func main() {
	app := NewWidgetShowcase()

	// Start with first theme
	t.SetTheme(app.themeNames[0])

	fmt.Println("Starting Widget Showcase...")
	fmt.Println("Press 't' to cycle through themes, Tab to move focus between widgets")

	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
