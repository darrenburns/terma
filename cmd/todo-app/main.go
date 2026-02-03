package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	t "terma"
)

// tagPattern matches hashtags: # followed by alphanumeric, underscore, or hyphen
var tagPattern = regexp.MustCompile(`#[a-zA-Z0-9_-]+`)

// darkThemeNames are the dark theme names in display order.
var darkThemeNames = []string{
	t.ThemeNameKanagawa,
	t.ThemeNameRosePine,
	t.ThemeNameCatppuccin,
	t.ThemeNameDracula,
	t.ThemeNameGruvbox,
	t.ThemeNameMonokai,
	t.ThemeNameNord,
	t.ThemeNameSolarized,
	t.ThemeNameTokyoNight,
}

// lightThemeNames are the light theme names in display order.
var lightThemeNames = []string{
	t.ThemeNameKanagawaLotus,
	t.ThemeNameRosePineDawn,
	t.ThemeNameCatppuccinLatte,
	t.ThemeNameDraculaLight,
	t.ThemeNameGruvboxLight,
	t.ThemeNameMonokaiLight,
	t.ThemeNameNordLight,
	t.ThemeNameSolarizedLight,
	t.ThemeNameTokyoNightDay,
}

// isLightTheme returns true if the theme name is a light theme.
func isLightTheme(name string) bool {
	for _, n := range lightThemeNames {
		if n == name {
			return true
		}
	}
	return false
}

// Task represents a single TODO item.
type Task struct {
	ID        string
	Title     string
	Completed bool
	CreatedAt time.Time
}

// TaskList represents a named list of tasks.
type TaskList struct {
	ID          string
	Name        string
	Tasks       *t.ListState[Task]
	ScrollState *t.ScrollState
}

// TodoApp is the main application widget.
type TodoApp struct {
	// Core state
	taskLists     []*TaskList
	activeListIdx t.Signal[int]
	inputState    *t.TextInputState

	// Move menu state
	showMoveMenu     t.Signal[bool]
	moveMenuState    *t.MenuState
	moveMenuAnchorID string

	// Filter state
	filterMode          t.Signal[bool]
	filterInputState    *t.TextInputState
	filterTagAcState    *t.AutocompleteState
	filteredListState   *t.ListState[Task]
	filteredScrollState *t.ScrollState

	// Editing state
	editingIndex   t.Signal[int]
	editInputState *t.TextAreaState

	// Theme picker state
	showThemePicker       t.Signal[bool]
	themeCategory         t.Signal[string] // "dark" or "light"
	darkThemeListState    *t.ListState[string]
	lightThemeListState   *t.ListState[string]
	darkThemeScrollState  *t.ScrollState
	lightThemeScrollState *t.ScrollState
	originalTheme         string

	// Help modal state
	showHelp t.Signal[bool]

	// Celebration animation state
	celebrationAngle *t.Animation[float64]
	wasCelebrating   bool // Track previous celebration state

	// ID counter for generating unique task IDs
	nextID int

	// Tag autocomplete state
	newTaskTagAcState *t.AutocompleteState
	editTagAcState    *t.AutocompleteState
}

// NewTodoApp creates a new todo application.
func NewTodoApp() *TodoApp {
	now := time.Now()
	initialTasks := []Task{
		{ID: "task-1", Title: "Invent a new color #creative #fun", Completed: false, CreatedAt: now},
		{ID: "task-2", Title: "Teach the cat to file taxes #pets #finance", Completed: false, CreatedAt: now},
		{ID: "task-3", Title: "Find out who let the dogs out #pets #mystery", Completed: true, CreatedAt: now},
		{ID: "task-4", Title: "Convince houseplants I'm responsible #home", Completed: false, CreatedAt: now},
		{ID: "task-5", Title: "Reply to email from 2019 #work #overdue", Completed: false, CreatedAt: now},
		{ID: "task-6", Title: "Figure out what the fox says #mystery #fun", Completed: true, CreatedAt: now},
		{ID: "task-7", Title: "Organize sock drawer by emotional value #home #creative", Completed: false, CreatedAt: now},
		{ID: "task-8", Title: "Finally read the terms and conditions #work", Completed: false, CreatedAt: now},
		{ID: "task-9", Title: "Become a morning person (unlikely) #health #fun", Completed: false, CreatedAt: now},
	}

	todayList := &TaskList{
		ID:          "today",
		Name:        "Today's tasks",
		Tasks:       t.NewListState(initialTasks),
		ScrollState: t.NewScrollState(),
	}

	inboxList := &TaskList{
		ID:          "inbox",
		Name:        "Inbox",
		Tasks:       t.NewListState([]Task{}),
		ScrollState: t.NewScrollState(),
	}

	app := &TodoApp{
		taskLists:             []*TaskList{todayList, inboxList},
		activeListIdx:         t.NewSignal(0),
		inputState:            t.NewTextInputState(""),
		showMoveMenu:          t.NewSignal(false),
		moveMenuState:         t.NewMenuState([]t.MenuItem{}),
		filterMode:            t.NewSignal(false),
		filterInputState:      t.NewTextInputState(""),
		filterTagAcState:      t.NewAutocompleteState(),
		filteredListState:     t.NewListState([]Task{}),
		filteredScrollState:   t.NewScrollState(),
		editingIndex:          t.NewSignal(-1),
		editInputState:        t.NewTextAreaState(""),
		showThemePicker:       t.NewSignal(false),
		themeCategory:         t.NewSignal("dark"),
		darkThemeListState:    t.NewListState(darkThemeNames),
		lightThemeListState:   t.NewListState(lightThemeNames),
		darkThemeScrollState:  t.NewScrollState(),
		lightThemeScrollState: t.NewScrollState(),
		showHelp:              t.NewSignal(false),
		nextID:                10,
		newTaskTagAcState:     t.NewAutocompleteState(),
		editTagAcState:        t.NewAutocompleteState(),
	}

	// Initialize celebration animation (loops continuously when started)
	app.celebrationAngle = t.NewAnimation(t.AnimationConfig[float64]{
		From:     0,
		To:       360,
		Duration: 4 * time.Second,
		Easing:   t.EaseLinear,
		OnComplete: func() {
			// Loop while still celebrating
			if app.isAllDone() {
				app.celebrationAngle.Reset()
				app.celebrationAngle.Start()
			}
		},
	})

	app.refreshTagSuggestions()

	return app
}

// generateID creates a unique ID for a new task.
func (a *TodoApp) generateID() string {
	id := fmt.Sprintf("task-%d", a.nextID)
	a.nextID++
	return id
}

func (a *TodoApp) taskRowID(taskID string) string {
	return "task-row-" + taskID
}

// activeList returns the currently active TaskList.
func (a *TodoApp) activeList() *TaskList {
	return a.taskLists[a.activeListIdx.Peek()]
}

// allTasks returns tasks from all lists.
func (a *TodoApp) allTasks() []Task {
	var all []Task
	for _, list := range a.taskLists {
		all = append(all, list.Tasks.GetItems()...)
	}
	return all
}

// isAllDone returns true if all tasks are completed and there's at least one task.
func (a *TodoApp) isAllDone() bool {
	tasks := a.activeList().Tasks.GetItems()
	if len(tasks) == 0 {
		return false
	}
	for _, task := range tasks {
		if !task.Completed {
			return false
		}
	}
	return true
}

// Build implements the Widget interface.
func (a *TodoApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Check celebration state and manage animation
	celebrating := a.isAllDone()
	if celebrating && !a.wasCelebrating {
		// Just completed all tasks - start celebration
		a.celebrationAngle.Reset()
		a.celebrationAngle.Start()
	} else if !celebrating && a.wasCelebrating {
		// No longer all done - stop celebration
		a.celebrationAngle.Stop()
	}
	a.wasCelebrating = celebrating

	// Background: subtle gradient from top-left when celebrating
	var bgColor t.ColorProvider
	if celebrating {
		bgColor = t.NewGradient(
			theme.Background.Lighten(0.15),
			theme.Background,
		).WithAngle(135)
	} else {
		bgColor = theme.Background
	}

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: bgColor,
			Padding:         t.EdgeInsetsXY(6, 2),
		},
		Children: []t.Widget{
			t.Dock{
				Bottom: []t.Widget{
					t.Column{
						Width:      t.Flex(1),
						CrossAlign: t.CrossAxisCenter,
						Children: []t.Widget{
							t.KeybindBar{},
						},
					},
				},
				Body: a.buildMainContainer(ctx, bgColor),
			},
			a.buildThemePicker(theme),
			a.buildMoveMenu(theme),
			a.buildHelpModal(theme),
		},
	}
}

// buildMainContainer creates the container with gradient border containing input and list.
func (a *TodoApp) buildMainContainer(ctx t.BuildContext, bgColor t.ColorProvider) t.Widget {
	theme := ctx.Theme()

	// Get celebration animation state
	celebrating := a.isAllDone()
	angle := a.celebrationAngle.Value().Get()

	// Calculate task counts for the decoration
	tasks := a.activeList().Tasks.GetItems()
	completed := 0
	for _, task := range tasks {
		if task.Completed {
			completed++
		}
	}

	// Show different count text based on filter mode
	var countText string
	if a.filterMode.Get() {
		filteredCount := len(a.getFilteredTasks())
		countText = fmt.Sprintf("%d of %d", filteredCount, len(tasks))
	} else {
		countText = fmt.Sprintf("%d/%d", completed, len(tasks))
	}

	// Build border based on celebration state
	var border t.Border
	if celebrating {
		// Celebration mode: rotating success gradient with subtle background fade
		border = t.Border{
			Style: t.BorderRounded,
			Decorations: []t.BorderDecoration{
				{Text: "All done!", Position: t.DecorationTopLeft},
				{Text: countText, Position: t.DecorationTopRight},
			},
			Color: t.NewGradient(
				theme.Primary,
				theme.Accent,
			).WithAngle(angle),
		}
	} else {
		// Normal mode: static gradient
		headerText := a.activeList().Name
		if a.filterMode.Get() {
			headerText = "Type to filter"
		} else if selectedCount := len(a.activeList().Tasks.SelectedItems()); selectedCount > 1 {
			headerText = fmt.Sprintf("%d items selected", selectedCount)
		}
		border = t.Border{
			Style: t.BorderRounded,
			Decorations: []t.BorderDecoration{
				{Text: headerText, Position: t.DecorationTopLeft},
				{Text: countText, Position: t.DecorationTopRight},
			},
			Color: t.NewGradient(
				theme.Background.Blend(theme.Primary, 0.5),
				theme.Background,
			).WithAngle(0),
		}
	}

	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style: t.Style{
			BackgroundColor: bgColor,
			Border:          border,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			a.buildInputRow(theme),
			a.buildTaskList(ctx),
		},
	}
}

// buildInputRow creates the new task input row or filter input row.
func (a *TodoApp) buildInputRow(theme t.ThemeData) t.Widget {
	if a.filterMode.Get() {
		return t.Row{
			Width: t.Flex(1),
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Padding:         t.EdgeInsetsXY(1, 0),
			},
			Children: []t.Widget{
				t.Text{
					Content: " / ",
					Style: t.Style{
						ForegroundColor: theme.Accent,
						Bold:            true,
					},
				},
				t.Autocomplete{
					ID:                    "filter-tag-ac",
					State:                 a.filterTagAcState,
					TriggerChars:          []rune{'#'},
					MinChars:              0,
					DisableKeysWhenHidden: true,
					Width:                 t.Flex(1),
					RenderSuggestion:      tagSuggestionRenderer("filter-input"),
					Child: t.TextInput{
						ID:          "filter-input",
						State:       a.filterInputState,
						Placeholder: "Filter tasks...",
						Highlighter: tagHighlighter(theme.Accent),
						Width:       t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Surface,
						},
						OnSubmit: a.handleFilterSubmit,
						OnChange: a.handleFilterChange,
					},
				},
			},
		}
	}

	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.Text{
				Content: " + ",
				Style: t.Style{
					ForegroundColor: theme.Primary,
					Bold:            true,
				},
			},
			t.Autocomplete{
				ID:                    "new-task-tag-ac",
				State:                 a.newTaskTagAcState,
				TriggerChars:          []rune{'#'},
				MinChars:              0,
				DisableKeysWhenHidden: true,
				Width:                 t.Flex(1),
				RenderSuggestion:      tagSuggestionRenderer("new-task-input"),
				Child: t.TextInput{
					ID:          "new-task-input",
					State:       a.inputState,
					Placeholder: "What needs to be done?",
					Highlighter: tagHighlighter(theme.Accent),
					Width:       t.Flex(1),
					Style: t.Style{
						BackgroundColor: theme.Surface,
					},
					OnSubmit: a.addTask,
					ExtraKeybinds: []t.Keybind{
						{Key: "enter", Name: "Create", Action: func() { a.addTask(a.inputState.GetText()) }},
						{Key: "tab", Name: "Tasks", Action: func() {}},
						{Key: "left", Action: a.handleNewTaskInputLeft, Hidden: true},
						{Key: "right", Action: a.handleNewTaskInputRight, Hidden: true},
					},
				},
			},
		},
	}
}

// buildTaskList creates the scrollable task list.
func (a *TodoApp) buildTaskList(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	isFilterMode := a.filterMode.Get()

	// Use filtered list when in filter mode
	listState := a.activeList().Tasks
	scrollState := a.activeList().ScrollState
	if isFilterMode {
		listState = a.filteredListState
		scrollState = a.filteredScrollState
	}

	// Show empty state if no items
	if listState.ItemCount() == 0 {
		message := "There’s nothing to do."
		if isFilterMode {
			message = "Press [b]enter[/] to create this task."
		}
		return t.Column{
			Height:     t.Flex(1),
			Width:      t.Flex(1),
			MainAlign:  t.MainAxisCenter,
			CrossAlign: t.CrossAxisCenter,
			Children: []t.Widget{
				t.Text{
					Spans: t.ParseMarkup(message, theme),
					Style: t.Style{
						ForegroundColor: theme.TextMuted.WithAlpha(0.5),
					},
				},
			},
		}
	}

	// Check if the list is focused
	listFocused := ctx.Focused() != nil && ctx.IsFocused(t.List[Task]{ID: "task-list"})

	return t.Scrollable{
		State:               scrollState,
		ScrollbarThumbColor: theme.Surface,
		ScrollbarTrackColor: theme.Background.Darken(0.05),
		Height:              t.Flex(1),
		Child: t.List[Task]{
			ID:          "task-list",
			State:       listState,
			ScrollState: scrollState,
			RenderItem:  a.renderTaskItem(ctx, listFocused),
			OnSelect:    a.toggleTask,
			MultiSelect: true,
			Blur:        func() { listState.ClearSelection() },
		},
	}
}

// renderTaskItem returns the render function for task items.
func (a *TodoApp) renderTaskItem(ctx t.BuildContext, listFocused bool) func(Task, bool, bool) t.Widget {
	theme := ctx.Theme()
	editingIdx := a.editingIndex.Get()

	return func(task Task, active bool, selected bool) t.Widget {
		rowID := a.taskRowID(task.ID)

		// Find the index of this task
		tasks := a.activeList().Tasks.GetItems()
		itemIndex := -1
		for i, tsk := range tasks {
			if tsk.ID == task.ID {
				itemIndex = i
				break
			}
		}

		// If this task is being edited, show TextArea
		if editingIdx == itemIndex {
			// Capture index for closures
			idx := itemIndex

			// Align with normal display: "  ○  " = prefix + circle + spacing
			return t.Row{
				ID:    rowID,
				Width: t.Flex(1),
				Children: []t.Widget{
					t.Text{Content: "  ○  "}, // Match the prefix + circle + space
					t.Autocomplete{
						ID:                    "edit-tag-ac",
						State:                 a.editTagAcState,
						TriggerChars:          []rune{'#'},
						MinChars:              0,
						DisableKeysWhenHidden: true,
						Width:                 t.Flex(1),
						RenderSuggestion:      tagSuggestionRenderer("edit-input"),
						Child: t.TextArea{
							ID:          "edit-input",
							State:       a.editInputState,
							Highlighter: tagHighlighter(theme.Accent),
							Width:       t.Flex(1),
							Style: t.Style{
								BackgroundColor: theme.Surface,
							},
							ExtraKeybinds: []t.Keybind{
								{Key: "enter", Name: "Save", Action: func() {
									a.saveEdit(idx, a.editInputState.GetText())
								}},
								{Key: "shift+enter", Name: "Newline", Action: func() {
									a.editInputState.ReplaceSelection("\n")
								}},
							},
						},
					},
				},
			}
		}

		// Normal display mode
		checkbox := "○"
		checkboxStyle := t.Style{ForegroundColor: theme.Border}
		if task.Completed {
			checkbox = "●"
			checkboxStyle.ForegroundColor = theme.Success
		}

		prefix := "  "
		textStyle := t.Style{ForegroundColor: theme.Text}
		rowStyle := t.Style{}

		// Determine background based on state (active+focused takes precedence over selected)
		if active && listFocused {
			// Show cursor and highlight row when list is focused
			prefix = "❯ "
			textStyle.ForegroundColor = theme.Text
			if selected {
				rowStyle.BackgroundColor = t.NewGradient(theme.Primary.WithAlpha(0.35), theme.Primary.WithAlpha(0.03)).WithAngle(90)
			} else {
				rowStyle.BackgroundColor = t.NewGradient(theme.Surface, theme.Background).WithAngle(90)
			}
			if !task.Completed {
				checkboxStyle.ForegroundColor = theme.Primary
			}
		} else if selected {
			// Selected but not active - solid background
			rowStyle.BackgroundColor = t.NewGradient(theme.Primary.WithAlpha(0.15), theme.Primary.WithAlpha(0.03)).WithAngle(90)
		}

		if task.Completed {
			textStyle.ForegroundColor = theme.TextMuted.WithAlpha(0.6)
			textStyle.Strikethrough = true
		}

		// Build the title widget - always highlight tags, also highlight filter matches when filtering
		var titleWidget t.Text
		if a.filterMode.Get() && a.getFilterText() != "" {
			// In filter mode: highlight both filter matches and tags
			titleWidget = t.Text{
				Spans: a.highlightMatchesAndTags(task.Title, textStyle, theme.Accent, theme.Accent.WithAlpha(0.1), theme.Accent),
				Width: t.Flex(1),
			}
		} else {
			// Normal mode: just highlight tags
			titleWidget = t.Text{
				Spans: a.highlightTags(task.Title, textStyle, theme.Accent),
				Width: t.Flex(1),
			}
		}
		titleWidget.Wrap = t.WrapSoft

		return t.Row{
			ID:    rowID,
			Width: t.Flex(1),
			Style: rowStyle,
			Children: []t.Widget{
				t.Text{
					Content: prefix,
					Style: t.Style{
						ForegroundColor: theme.Primary,
					},
				},
				t.Text{
					Content: checkbox + "  ",
					Style:   checkboxStyle,
				},
				titleWidget,
			},
		}
	}
}

// buildThemePicker creates the theme picker modal with dark/light switcher.
func (a *TodoApp) buildThemePicker(theme t.ThemeData) t.Widget {
	category := a.themeCategory.Get()
	isDark := category == "dark"
	listHeight := func(items []string) t.Dimension {
		height := len(items)
		if height > 10 {
			height = 10
		}
		if height == 0 {
			height = 1
		}
		return t.Cells(height)
	}

	// Tab style helpers
	activeTabStyle := t.Style{
		ForegroundColor: theme.Primary,
		Bold:            true,
	}
	inactiveTabStyle := t.Style{
		ForegroundColor: theme.TextMuted,
	}

	darkTabStyle := inactiveTabStyle
	lightTabStyle := inactiveTabStyle
	if isDark {
		darkTabStyle = activeTabStyle
	} else {
		lightTabStyle = activeTabStyle
	}

	return t.Floating{
		Visible: a.showThemePicker.Get(),
		Config: t.FloatConfig{
			Position:      t.FloatPositionCenter,
			Modal:         true,
			OnDismiss:     a.dismissThemePicker,
			BackdropColor: t.Black.WithAlpha(0.04),
		},
		Child: t.Column{
			Spacing: 1,
			Width:   t.Cells(43),
			Style: t.Style{
				BackgroundColor: t.NewGradient(theme.Surface.Lighten(0.3), theme.Surface).WithAngle(45),
				Padding:         t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				t.Text{
					Content: "Select Theme",
					Style: t.Style{
						ForegroundColor: t.NewGradient(theme.Primary.Lighten(0.1), theme.Primary).WithAngle(90),
						Bold:            true,
					},
				},
				// Category tabs
				t.Row{
					Spacing: 2,
					Children: []t.Widget{
						t.Text{Content: "← Dark", Style: darkTabStyle},
						t.Text{Content: "Light →", Style: lightTabStyle},
					},
				},
				// Theme list switcher
				t.Switcher{
					Active: category,
					Children: map[string]t.Widget{
						"dark": t.Scrollable{
							State:  a.darkThemeScrollState,
							Height: listHeight(darkThemeNames),
							Child: t.List[string]{
								ID:             "dark-theme-list",
								State:          a.darkThemeListState,
								ScrollState:    a.darkThemeScrollState,
								OnSelect:       a.selectTheme,
								OnCursorChange: a.previewTheme,
								RenderItem:     a.renderThemeItem(theme),
							},
						},
						"light": t.Scrollable{
							State:  a.lightThemeScrollState,
							Height: listHeight(lightThemeNames),
							Child: t.List[string]{
								ID:             "light-theme-list",
								State:          a.lightThemeListState,
								ScrollState:    a.lightThemeScrollState,
								OnSelect:       a.selectTheme,
								OnCursorChange: a.previewTheme,
								RenderItem:     a.renderThemeItem(theme),
							},
						},
					},
				},
				t.Text{
					Content: "←→ dark/light · ↑↓ navigate · enter select · esc cancel",
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},
			},
		},
	}
}

// buildMoveMenu creates the move task menu modal.
func (a *TodoApp) buildMoveMenu(theme t.ThemeData) t.Widget {
	return t.ShowWhen(a.showMoveMenu.Get(), t.Column{
		Children: []t.Widget{
			t.Floating{
				Visible: true,
				Config: t.FloatConfig{
					Position:      t.FloatPositionCenter,
					OnDismiss:     a.dismissMoveMenu,
					BackdropColor: t.Black.WithAlpha(0.2),
				},
				Child: t.Spacer{Width: t.Cells(1), Height: t.Cells(1)},
			},
			t.Menu{
				ID:        "move-menu",
				State:     a.moveMenuState,
				AnchorID:  a.moveMenuAnchorID,
				Anchor:    t.AnchorTopLeft,
				OnSelect:  a.handleMoveMenuSelect,
				OnDismiss: a.dismissMoveMenu,
				Style: t.Style{
					BackgroundColor: theme.Surface,
				},
			},
		},
	})
}

// renderThemeItem returns the render function for theme list items.
func (a *TodoApp) renderThemeItem(theme t.ThemeData) func(string, bool, bool) t.Widget {
	currentTheme := t.CurrentThemeName()
	return func(themeName string, active bool, selected bool) t.Widget {
		prefix := "  "
		style := t.Style{ForegroundColor: theme.Text}

		if active {
			prefix = "❯ "
			style.ForegroundColor = theme.Accent
		}

		if themeName == currentTheme {
			if !active {
				style.ForegroundColor = theme.Success
			}
		}

		children := []t.Widget{
			t.Text{
				Content: prefix + themeName,
				Style:   style,
				Width:   t.Cells(20),
			},
		}

		// Only show color swatches for the active item
		if active {
			itemTheme, _ := t.GetTheme(themeName)
			swatch := func(color t.Color) t.Widget {
				return t.Text{
					Content: "██",
					Style:   t.Style{ForegroundColor: color},
				}
			}
			children = append(children,
				swatch(itemTheme.Primary),
				swatch(itemTheme.Secondary),
				swatch(itemTheme.Accent),
				swatch(itemTheme.Success),
				swatch(itemTheme.Error),
			)
		}

		return t.Row{
			Width:    t.Flex(1),
			Spacing:  1,
			Children: children,
		}
	}
}

// Keybinds returns the declarative keybindings for the app.
func (a *TodoApp) Keybinds() []t.Keybind {
	editingIdx := a.editingIndex.Peek()
	isEditing := editingIdx >= 0
	isThemePicker := a.showThemePicker.Peek()
	isMoveMenu := a.showMoveMenu.Peek()
	isFilterMode := a.filterMode.Peek()
	isHelp := a.showHelp.Peek()

	// Help modal - any key closes it
	if isHelp {
		return []t.Keybind{
			{Key: "escape", Name: "Close", Action: a.closeHelp},
			{Key: "?", Name: "Close", Action: a.closeHelp, Hidden: true},
		}
	}

	// Theme picker modal has its own keybinds
	if isThemePicker {
		return []t.Keybind{
			{Key: "escape", Name: "Cancel", Action: a.dismissThemePicker},
			{Key: "left", Name: "Dark", Action: a.showDarkThemes},
			{Key: "right", Name: "Light", Action: a.showLightThemes},
		}
	}

	// Move menu modal has its own keybinds
	if isMoveMenu {
		return []t.Keybind{
			{Key: "escape", Name: "Cancel", Action: a.dismissMoveMenu},
		}
	}

	// Filter mode has its own keybinds
	if isFilterMode {
		return []t.Keybind{
			{Key: "escape", Name: "Clear", Action: a.exitFilterMode},
			{Key: "enter", Name: "Toggle", Action: a.toggleCurrentTask, Hidden: true},
			{Key: " ", Name: "Toggle", Action: a.toggleCurrentTask},
			{Key: "d", Name: "Delete", Action: a.deleteCurrentTask},
			{Key: "up", Action: a.navigateUp, Hidden: true},
			{Key: "down", Action: a.navigateDown, Hidden: true},
		}
	}

	keybinds := []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
		// Navigation between input and list (these bubble up from focused widgets)
		{Key: "up", Action: a.navigateUp, Hidden: true},
		{Key: "down", Action: a.navigateDown, Hidden: true},
		{Key: "left", Name: "Prev List", Action: a.switchToPreviousList, Hidden: true},
		{Key: "right", Name: "Next List", Action: a.switchToNextList, Hidden: true},
	}

	if !isEditing {
		keybinds = append(keybinds,
			t.Keybind{Key: "enter", Name: "Toggle", Action: a.toggleCurrentTask, Hidden: true},
			t.Keybind{Key: " ", Name: "Toggle", Action: a.toggleCurrentTask},
			t.Keybind{Key: "e", Name: "Edit", Action: a.startEdit},
			t.Keybind{Key: "d", Name: "Delete", Action: a.deleteCurrentTask},
			t.Keybind{Key: "m", Name: "Move", Action: a.openMoveMenu},
			t.Keybind{Key: "t", Name: "Theme", Action: a.openThemePicker},
			t.Keybind{Key: "/", Name: "Filter", Action: a.enterFilterMode},
			t.Keybind{Key: "ctrl+j", Name: "Move Down", Action: a.moveTaskDown, Hidden: true},
			t.Keybind{Key: "ctrl+k", Name: "Move Up", Action: a.moveTaskUp, Hidden: true},
			t.Keybind{Key: "?", Name: "Help", Action: a.openHelp},
		)
	} else {
		keybinds = append(keybinds,
			t.Keybind{Key: "escape", Name: "Cancel", Action: a.cancelEdit},
		)
	}

	return keybinds
}

func (a *TodoApp) switchToPreviousList() {
	if a.editingIndex.Peek() >= 0 {
		a.cancelEdit()
	}
	a.activeListIdx.Update(func(idx int) int {
		if idx > 0 {
			return idx - 1
		}
		return len(a.taskLists) - 1
	})
	a.filteredScrollState.SetOffset(0)
	a.refreshFilteredTasks()
}

func (a *TodoApp) switchToNextList() {
	if a.editingIndex.Peek() >= 0 {
		a.cancelEdit()
	}
	a.activeListIdx.Update(func(idx int) int {
		if idx < len(a.taskLists)-1 {
			return idx + 1
		}
		return 0
	})
	a.filteredScrollState.SetOffset(0)
	a.refreshFilteredTasks()
}

func (a *TodoApp) handleNewTaskInputLeft() {
	if a.inputState == nil {
		return
	}
	if a.inputState.GetText() == "" {
		a.switchToPreviousList()
		return
	}
	a.inputState.ClearSelection()
	a.inputState.CursorLeft()
}

func (a *TodoApp) handleNewTaskInputRight() {
	if a.inputState == nil {
		return
	}
	if a.inputState.GetText() == "" {
		a.switchToNextList()
		return
	}
	a.inputState.ClearSelection()
	a.inputState.CursorRight()
}

func (a *TodoApp) openMoveMenu() {
	listState := a.activeList().Tasks
	if a.filterMode.Peek() {
		listState = a.filteredListState
	}
	if listState.ItemCount() == 0 {
		return
	}

	anchorTask, ok := listState.SelectedItem()
	if !ok {
		return
	}
	a.moveMenuAnchorID = a.taskRowID(anchorTask.ID)

	items := a.buildMoveMenuItems()
	if len(items) == 0 {
		return
	}

	a.moveMenuState = t.NewMenuState(items)
	a.showMoveMenu.Set(true)
	t.RequestFocus("move-menu")
}

func (a *TodoApp) dismissMoveMenu() {
	if a.moveMenuState != nil {
		a.moveMenuState.CloseSubmenu()
	}
	a.showMoveMenu.Set(false)
	a.moveMenuAnchorID = ""
	t.RequestFocus("task-list")
}

func (a *TodoApp) handleMoveMenuSelect(item t.MenuItem) {
	if item.Action != nil {
		item.Action()
	}
}

func (a *TodoApp) buildMoveMenuItems() []t.MenuItem {
	currentID := a.activeList().ID
	items := make([]t.MenuItem, 0, len(a.taskLists)-1)
	for _, list := range a.taskLists {
		if list.ID == currentID {
			continue
		}
		listCopy := list
		items = append(items, t.MenuItem{
			Label:  "Move to " + list.Name,
			Action: func() { a.moveTaskToList(listCopy) },
		})
	}
	return items
}

func (a *TodoApp) moveTaskToList(targetList *TaskList) {
	if targetList == nil {
		a.dismissMoveMenu()
		return
	}

	sourceList := a.activeList()
	if sourceList.ID == targetList.ID {
		a.dismissMoveMenu()
		return
	}

	listState := sourceList.Tasks
	selectedTasks := listState.SelectedItems()
	if len(selectedTasks) == 0 {
		if task, ok := listState.SelectedItem(); ok {
			selectedTasks = []Task{task}
		}
	}
	if len(selectedTasks) == 0 {
		a.dismissMoveMenu()
		return
	}

	idsToMove := make(map[string]struct{}, len(selectedTasks))
	for _, task := range selectedTasks {
		idsToMove[task.ID] = struct{}{}
	}

	sourceList.Tasks.RemoveWhere(func(task Task) bool {
		_, shouldMove := idsToMove[task.ID]
		return shouldMove
	})
	listState.ClearSelection()
	listState.ClearAnchor()

	for i := len(selectedTasks) - 1; i >= 0; i-- {
		targetList.Tasks.Prepend(selectedTasks[i])
	}

	a.refreshTagSuggestions()
	a.refreshFilteredTasks()
	a.dismissMoveMenu()
}

// navigateUp handles up arrow for cross-widget navigation.
// Called when: input focused, list at top item, or in edit mode.
func (a *TodoApp) navigateUp() {
	editingIdx := a.editingIndex.Peek()
	listState := a.activeList().Tasks
	if a.filterMode.Peek() {
		listState = a.filteredListState
	}

	if editingIdx >= 0 {
		// In edit mode: cancel edit and move cursor up (or to input if at top)
		a.editingIndex.Set(-1)
		if editingIdx == 0 {
			listState.ClearSelection()
			t.RequestFocus("new-task-input")
		} else {
			listState.SelectPrevious()
			t.RequestFocus("task-list")
		}
	} else {
		// List at top or input focused - move to input
		listState.ClearSelection()
		t.RequestFocus("new-task-input")
	}
}

// navigateDown handles down arrow for cross-widget navigation.
// Called when: input focused, list at bottom item, or in edit mode.
func (a *TodoApp) navigateDown() {
	editingIdx := a.editingIndex.Peek()
	listState := a.activeList().Tasks
	if a.filterMode.Peek() {
		listState = a.filteredListState
	}

	if editingIdx >= 0 {
		// In edit mode: cancel edit and move cursor down
		a.editingIndex.Set(-1)
		itemCount := listState.ItemCount()
		if editingIdx < itemCount-1 {
			listState.SelectNext()
		}
		t.RequestFocus("task-list")
	} else {
		// Input focused or list at bottom - move to list
		t.RequestFocus("task-list")
	}
}

// addTask creates a new task from the input text.
func (a *TodoApp) addTask(title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}

	task := Task{
		ID:        a.generateID(),
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	listState := a.activeList().Tasks
	listState.Prepend(task)
	listState.SelectIndex(0)
	a.inputState.SetText("")
	a.refreshTagSuggestions()
	a.refreshFilteredTasks()
}

// toggleCurrentTask toggles the completion status of selected tasks.
// If multiple items are selected: sets all to completed if any are uncompleted,
// otherwise sets all to uncompleted. If no selection, toggles the cursor item.
func (a *TodoApp) toggleCurrentTask() {
	// Use the appropriate list state based on filter mode
	listState := a.activeList().Tasks
	if a.filterMode.Peek() {
		listState = a.filteredListState
	}

	// Check for multi-select: if items are selected, apply consistent state to all
	selectedTasks := listState.SelectedItems()
	if len(selectedTasks) > 0 {
		// Determine target state: if any are uncompleted, complete all; otherwise uncomplete all
		anyUncompleted := false
		for _, task := range selectedTasks {
			if !task.Completed {
				anyUncompleted = true
				break
			}
		}
		targetState := anyUncompleted // true = mark completed, false = mark uncompleted

		for _, task := range selectedTasks {
			a.setTaskCompleted(task, targetState)
		}
		return
	}

	// No selection - toggle just the cursor item
	if task, ok := listState.SelectedItem(); ok {
		a.toggleTask(task)
	}
}

// toggleTask toggles the completion status of the given task.
func (a *TodoApp) toggleTask(task Task) {
	newState := !task.Completed
	a.updateTaskCompletion(a.activeList().Tasks, task.ID, newState)
	if a.filterMode.Peek() {
		a.updateTaskCompletion(a.filteredListState, task.ID, newState)
	}
}

// setTaskCompleted sets the completion status of the given task to a specific value.
func (a *TodoApp) setTaskCompleted(task Task, completed bool) {
	a.updateTaskCompletion(a.activeList().Tasks, task.ID, completed)
	if a.filterMode.Peek() {
		a.updateTaskCompletion(a.filteredListState, task.ID, completed)
	}
}

func (a *TodoApp) updateTaskCompletion(listState *t.ListState[Task], id string, completed bool) {
	listState.Items.Update(func(items []Task) []Task {
		for i := range items {
			if items[i].ID == id {
				items[i].Completed = completed
				break
			}
		}
		return items
	})
}

func (a *TodoApp) updateTaskTitle(listState *t.ListState[Task], id string, title string) {
	listState.Items.Update(func(items []Task) []Task {
		for i := range items {
			if items[i].ID == id {
				items[i].Title = title
				break
			}
		}
		return items
	})
}

// deleteCurrentTask removes selected tasks.
// If multiple items are selected, deletes all of them. Otherwise deletes the cursor item.
func (a *TodoApp) deleteCurrentTask() {
	// Use the appropriate list state based on filter mode
	isFilterMode := a.filterMode.Peek()
	listState := a.activeList().Tasks
	if isFilterMode {
		listState = a.filteredListState
	}
	sourceList := a.activeList().Tasks

	// Check for multi-select: if items are selected, delete all of them
	selectedTasks := listState.SelectedItems()
	if len(selectedTasks) > 0 {
		// Build a set of IDs to delete
		idsToDelete := make(map[string]struct{}, len(selectedTasks))
		for _, task := range selectedTasks {
			idsToDelete[task.ID] = struct{}{}
		}

		// Remove all matching tasks
		sourceList.RemoveWhere(func(task Task) bool {
			_, shouldDelete := idsToDelete[task.ID]
			return shouldDelete
		})

		a.refreshTagSuggestions()
		a.refreshFilteredTasks()

		listState.ClearSelection()
		listState.ClearAnchor()

		// If in filter mode and no more filtered items, refocus the filter input
		if isFilterMode && len(a.getFilteredTasks()) == 0 {
			t.RequestFocus("filter-input")
		}
		return
	}

	// No selection - delete just the cursor item
	task, ok := listState.SelectedItem()
	if !ok {
		return
	}

	tasks := sourceList.GetItems()
	for i, tsk := range tasks {
		if tsk.ID == task.ID {
			sourceList.RemoveAt(i)
			a.refreshTagSuggestions()
			a.refreshFilteredTasks()

			// If in filter mode and no more filtered items, refocus the filter input
			if isFilterMode && len(a.getFilteredTasks()) == 0 {
				t.RequestFocus("filter-input")
			}
			return
		}
	}
}

// moveTaskUp moves selected tasks up in the list.
// If multiple items are selected, moves all of them as a block.
func (a *TodoApp) moveTaskUp() {
	listState := a.activeList().Tasks
	tasks := listState.GetItems()
	selectedIndices := listState.SelectedIndices()

	// If there's a selection, move the entire selection block
	if len(selectedIndices) > 0 {
		firstIdx := selectedIndices[0]
		lastIdx := selectedIndices[len(selectedIndices)-1]

		if firstIdx <= 0 {
			return // Already at top
		}

		// Move the item above the selection to below the selection
		itemAbove := tasks[firstIdx-1]
		copy(tasks[firstIdx-1:lastIdx], tasks[firstIdx:lastIdx+1])
		tasks[lastIdx] = itemAbove
		listState.SetItems(tasks)
		a.refreshFilteredTasks()

		// Update selection indices (all shift up by 1)
		listState.SelectRange(firstIdx-1, lastIdx-1)

		// Move cursor up
		cursorIdx := listState.CursorIndex.Peek()
		listState.SelectIndex(cursorIdx - 1)
		return
	}

	// No selection - move just the cursor item
	idx := listState.CursorIndex.Peek()
	if idx <= 0 || idx >= len(tasks) {
		return
	}

	tasks[idx], tasks[idx-1] = tasks[idx-1], tasks[idx]
	listState.SetItems(tasks)
	a.refreshFilteredTasks()
	listState.SelectIndex(idx - 1)
}

// moveTaskDown moves selected tasks down in the list.
// If multiple items are selected, moves all of them as a block.
func (a *TodoApp) moveTaskDown() {
	listState := a.activeList().Tasks
	tasks := listState.GetItems()
	selectedIndices := listState.SelectedIndices()

	// If there's a selection, move the entire selection block
	if len(selectedIndices) > 0 {
		firstIdx := selectedIndices[0]
		lastIdx := selectedIndices[len(selectedIndices)-1]

		if lastIdx >= len(tasks)-1 {
			return // Already at bottom
		}

		// Move the item below the selection to above the selection
		itemBelow := tasks[lastIdx+1]
		copy(tasks[firstIdx+1:lastIdx+2], tasks[firstIdx:lastIdx+1])
		tasks[firstIdx] = itemBelow
		listState.SetItems(tasks)
		a.refreshFilteredTasks()

		// Update selection indices (all shift down by 1)
		listState.SelectRange(firstIdx+1, lastIdx+1)

		// Move cursor down
		cursorIdx := listState.CursorIndex.Peek()
		listState.SelectIndex(cursorIdx + 1)
		return
	}

	// No selection - move just the cursor item
	idx := listState.CursorIndex.Peek()
	if idx < 0 || idx >= len(tasks)-1 {
		return
	}

	tasks[idx], tasks[idx+1] = tasks[idx+1], tasks[idx]
	listState.SetItems(tasks)
	a.refreshFilteredTasks()
	listState.SelectIndex(idx + 1)
}

// startEdit begins inline editing of the current task.
func (a *TodoApp) startEdit() {
	listState := a.activeList().Tasks
	idx := listState.CursorIndex.Peek()
	tasks := listState.GetItems()
	if idx >= 0 && idx < len(tasks) {
		listState.ClearSelection()
		a.editInputState.SetText(tasks[idx].Title)
		a.editInputState.ClearSelection()
		a.editInputState.CursorEnd() // Position cursor at end of text
		a.editingIndex.Set(idx)
		t.RequestFocus("edit-input")
	}
}

// saveEdit saves the edited task title.
func (a *TodoApp) saveEdit(index int, newTitle string) {
	newTitle = strings.TrimSpace(newTitle)
	if newTitle == "" {
		a.cancelEdit()
		return
	}

	listState := a.activeList().Tasks
	tasks := listState.GetItems()
	if index >= 0 && index < len(tasks) {
		tasks[index].Title = newTitle
		listState.SetItems(tasks)
		if a.filterMode.Peek() {
			a.updateTaskTitle(a.filteredListState, tasks[index].ID, newTitle)
		}
		a.refreshTagSuggestions()
		a.refreshFilteredTasks()
	}
	a.editingIndex.Set(-1)
	t.RequestFocus("task-list")
}

// cancelEdit cancels the current edit.
func (a *TodoApp) cancelEdit() {
	a.editingIndex.Set(-1)
	t.RequestFocus("task-list")
}

// openThemePicker shows the theme picker modal.
func (a *TodoApp) openThemePicker() {
	// Store original theme to restore on cancel
	a.originalTheme = t.CurrentThemeName()

	// Determine which category and select current theme
	if isLightTheme(a.originalTheme) {
		a.themeCategory.Set("light")
		for i, name := range lightThemeNames {
			if name == a.originalTheme {
				a.lightThemeListState.SelectIndex(i)
				break
			}
		}
		t.RequestFocus("light-theme-list")
	} else {
		a.themeCategory.Set("dark")
		for i, name := range darkThemeNames {
			if name == a.originalTheme {
				a.darkThemeListState.SelectIndex(i)
				break
			}
		}
		t.RequestFocus("dark-theme-list")
	}

	a.showThemePicker.Set(true)
}

// dismissThemePicker hides the theme picker and restores original theme.
func (a *TodoApp) dismissThemePicker() {
	// Restore original theme
	t.SetTheme(a.originalTheme)
	a.showThemePicker.Set(false)
	t.RequestFocus("task-list")
}

// previewTheme updates the theme as user navigates the list.
func (a *TodoApp) previewTheme(themeName string) {
	t.SetTheme(themeName)
}

// selectTheme confirms the theme selection and closes the picker.
func (a *TodoApp) selectTheme(themeName string) {
	t.SetTheme(themeName)
	a.showThemePicker.Set(false)
	t.RequestFocus("task-list")
}

// showDarkThemes switches to the dark themes category.
func (a *TodoApp) showDarkThemes() {
	if a.themeCategory.Peek() == "dark" {
		return
	}
	a.themeCategory.Set("dark")
	// Preview the currently selected dark theme
	if idx := a.darkThemeListState.CursorIndex.Peek(); idx >= 0 && idx < len(darkThemeNames) {
		t.SetTheme(darkThemeNames[idx])
	}
	t.RequestFocus("dark-theme-list")
}

// showLightThemes switches to the light themes category.
func (a *TodoApp) showLightThemes() {
	if a.themeCategory.Peek() == "light" {
		return
	}
	a.themeCategory.Set("light")
	// Preview the currently selected light theme
	if idx := a.lightThemeListState.CursorIndex.Peek(); idx >= 0 && idx < len(lightThemeNames) {
		t.SetTheme(lightThemeNames[idx])
	}
	t.RequestFocus("light-theme-list")
}

// buildHelpModal creates the keyboard shortcuts help modal.
func (a *TodoApp) buildHelpModal(theme t.ThemeData) t.Widget {
	// Helper to create a key-action pair
	keyCell := func(key, action string) t.Widget {
		return t.Row{
			Width:   t.Cells(18),
			Spacing: 1,
			Children: []t.Widget{
				t.Text{
					Content: key,
					Width:   t.Cells(7),
					Style: t.Style{
						ForegroundColor: theme.Accent,
						Bold:            true,
					},
				},
				t.Text{
					Content: action,
					Style: t.Style{
						ForegroundColor: theme.Text,
					},
				},
			},
		}
	}

	return t.Floating{
		Visible: a.showHelp.Get(),
		Config: t.FloatConfig{
			Position:      t.FloatPositionCenter,
			Modal:         true,
			OnDismiss:     a.closeHelp,
			BackdropColor: t.Black.WithAlpha(0.3),
		},
		Child: t.Column{
			Width: t.Cells(42),
			Style: t.Style{
				BackgroundColor: t.NewGradient(theme.Surface.Lighten(0.3), theme.Surface).WithAngle(45),
				Padding:         t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				t.Text{
					Content: "Keyboard Shortcuts",
					Style: t.Style{
						ForegroundColor: t.NewGradient(theme.Primary.Lighten(0.1), theme.Primary).WithAngle(90),
						Bold:            true,
					},
				},
				t.Row{
					Children: []t.Widget{
						keyCell("space", "Toggle"),
						keyCell("ctrl+k", "Move ↑"),
					},
				},
				t.Row{
					Children: []t.Widget{
						keyCell("e", "Edit"),
						keyCell("ctrl+j", "Move ↓"),
					},
				},
				t.Row{
					Children: []t.Widget{
						keyCell("d", "Delete"),
						keyCell("/", "Filter"),
					},
				},
				t.Row{
					Children: []t.Widget{
						keyCell("left/right", "Switch list"),
						keyCell("m", "Move"),
					},
				},
				t.Row{
					Children: []t.Widget{
						keyCell("t", "Theme"),
						keyCell("q", "Quit"),
					},
				},
			},
		},
	}
}

// openHelp shows the help modal.
func (a *TodoApp) openHelp() {
	a.showHelp.Set(true)
}

// closeHelp hides the help modal.
func (a *TodoApp) closeHelp() {
	a.showHelp.Set(false)
}

// enterFilterMode activates filter mode and focuses the filter input.
func (a *TodoApp) enterFilterMode() {
	a.filterMode.Set(true)
	a.filterInputState.SetText("")
	a.refreshFilteredTasks()
	t.RequestFocus("filter-input")
}

// handleFilterSubmit handles enter in the filter input.
// If no matches, creates a new task with the filter text.
func (a *TodoApp) handleFilterSubmit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	// If there are matches, do nothing (user can navigate to them)
	if len(a.getFilteredTasks()) > 0 {
		return
	}

	// No matches - create a new task with the filter text
	task := Task{
		ID:        a.generateID(),
		Title:     text,
		Completed: false,
		CreatedAt: time.Now(),
	}
	listState := a.activeList().Tasks
	listState.Prepend(task)
	a.refreshTagSuggestions()
	a.refreshFilteredTasks()

	// Select the newly created task (first item)
	listState.SelectIndex(0)

	// Exit filter mode
	a.filterMode.Set(false)
	a.filterInputState.SetText("")
	t.RequestFocus("task-list")
}

// exitFilterMode deactivates filter mode and clears the filter.
func (a *TodoApp) exitFilterMode() {
	// Remember the selected task from the filtered list
	selectedTask, hasSelection := a.filteredListState.SelectedItem()

	a.filterMode.Set(false)
	a.filterInputState.SetText("")

	// Restore cursor to the same task in the main list
	if hasSelection {
		tasks := a.activeList().Tasks.GetItems()
		for i, task := range tasks {
			if task.ID == selectedTask.ID {
				a.activeList().Tasks.SelectIndex(i)
				break
			}
		}
	}

	t.RequestFocus("task-list")
}

// handleFilterChange updates the filtered list as the user types.
func (a *TodoApp) handleFilterChange(_ string) {
	a.refreshFilteredTasks()
}

// getFilterText returns the current filter text (lowercase for case-insensitive matching).
func (a *TodoApp) getFilterText() string {
	return strings.ToLower(a.filterInputState.GetText())
}

// taskMatchesFilter returns true if the task title contains the filter text.
func (a *TodoApp) taskMatchesFilter(task Task) bool {
	filterText := a.getFilterText()
	if filterText == "" {
		return true
	}
	return strings.Contains(strings.ToLower(task.Title), filterText)
}

// getFilteredTasks returns only the tasks that match the current filter.
func (a *TodoApp) getFilteredTasks() []Task {
	filterText := a.getFilterText()
	if filterText == "" {
		return a.activeList().Tasks.GetItems()
	}

	var filtered []Task
	for _, task := range a.activeList().Tasks.GetItems() {
		if a.taskMatchesFilter(task) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// highlightMatches creates spans with highlighted matching substrings.
func (a *TodoApp) highlightMatches(title string, baseStyle t.Style, highlightColor t.Color, highlightBackgroundColor t.Color) []t.Span {
	filterText := a.getFilterText()
	if filterText == "" {
		return []t.Span{{Text: title, Style: styleToSpanStyle(baseStyle)}}
	}

	var spans []t.Span
	titleLower := strings.ToLower(title)
	pos := 0

	for pos < len(title) {
		// Find next match
		matchIdx := strings.Index(titleLower[pos:], filterText)
		if matchIdx == -1 {
			// No more matches, add remaining text
			if pos < len(title) {
				spans = append(spans, t.Span{Text: title[pos:], Style: styleToSpanStyle(baseStyle)})
			}
			break
		}

		// Add text before match
		matchStart := pos + matchIdx
		if matchStart > pos {
			spans = append(spans, t.Span{Text: title[pos:matchStart], Style: styleToSpanStyle(baseStyle)})
		}

		// Add highlighted match
		matchEnd := matchStart + len(filterText)
		highlightStyle := styleToSpanStyle(baseStyle)
		highlightStyle.Underline = t.UnderlineSingle
		highlightStyle.UnderlineColor = highlightColor
		highlightStyle.Background = highlightBackgroundColor
		spans = append(spans, t.Span{Text: title[matchStart:matchEnd], Style: highlightStyle})

		pos = matchEnd
	}

	return spans
}

// styleToSpanStyle converts a Style to a SpanStyle.
func styleToSpanStyle(s t.Style) t.SpanStyle {
	return t.SpanStyle{
		Foreground:    colorProviderToColor(s.ForegroundColor),
		Bold:          s.Bold,
		Italic:        s.Italic,
		Strikethrough: s.Strikethrough,
	}
}

// extractTags returns all unique tags from a task title.
func extractTags(title string) []string {
	matches := tagPattern.FindAllString(title, -1)
	seen := make(map[string]bool)
	var tags []string
	for _, tag := range matches {
		lower := strings.ToLower(tag)
		if !seen[lower] {
			seen[lower] = true
			tags = append(tags, tag)
		}
	}
	return tags
}

// buildTagSuggestions builds a sorted list of unique tag suggestions from tasks.
func buildTagSuggestions(tasks []Task) []t.Suggestion {
	seen := make(map[string]string)
	counts := make(map[string]int)
	for _, task := range tasks {
		for _, tag := range extractTags(task.Title) {
			trimmed := strings.TrimPrefix(tag, "#")
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, ok := seen[key]; !ok {
				seen[key] = trimmed
			}
			counts[key]++
		}
	}

	labels := make([]string, 0, len(seen))
	for _, label := range seen {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		return strings.ToLower(labels[i]) < strings.ToLower(labels[j])
	})

	suggestions := make([]t.Suggestion, 0, len(labels))
	for _, label := range labels {
		key := strings.ToLower(label)
		count := counts[key]
		suggestions = append(suggestions, t.Suggestion{
			Label:       label,
			Value:       "#" + label,
			Description: fmt.Sprintf("(%d)", count),
		})
	}
	return suggestions
}

func (a *TodoApp) refreshTagSuggestions() {
	suggestions := buildTagSuggestions(a.allTasks())
	a.newTaskTagAcState.SetSuggestions(suggestions)
	a.filterTagAcState.SetSuggestions(suggestions)
	a.editTagAcState.SetSuggestions(suggestions)
}

func (a *TodoApp) refreshFilteredTasks() {
	if !a.filterMode.Peek() {
		return
	}
	a.filteredListState.SetItems(a.getFilteredTasks())
}

func tagSuggestionRenderer(childID string) func(t.Suggestion, bool, t.MatchResult, t.BuildContext) t.Widget {
	return func(item t.Suggestion, active bool, match t.MatchResult, ctx t.BuildContext) t.Widget {
		theme := ctx.Theme()
		showCursor := false
		if active {
			focused := ctx.Focused()
			if focused != nil {
				if identifiable, ok := focused.(t.Identifiable); ok && identifiable.WidgetID() == childID {
					showCursor = true
				}
			}
		}

		textColor := theme.Text
		if showCursor {
			textColor = theme.SelectionText
		}

		style := t.Style{
			Padding: t.EdgeInsets{Left: 1, Right: 1},
		}
		if showCursor {
			style.BackgroundColor = theme.ActiveCursor
		}

		spans := make([]t.Span, 0, 4)
		if item.Icon != "" {
			spans = append(spans, t.Span{Text: item.Icon + " "})
		}
		if match.Matched && len(match.Ranges) > 0 {
			spans = append(spans, t.HighlightSpans(item.Label, match.Ranges, t.MatchHighlightStyle(theme))...)
		} else {
			spans = append(spans, t.Span{Text: item.Label})
		}
		if item.Description != "" {
			countColor := theme.TextMuted.WithAlpha(0.7)
			if showCursor {
				countColor = theme.SelectionText.WithAlpha(0.7)
			}
			spans = append(spans,
				t.Span{Text: " "},
				t.Span{Text: item.Description, Style: t.SpanStyle{Foreground: countColor}},
			)
		}

		return t.Row{
			Style: style,
			Children: []t.Widget{
				t.Text{
					Spans: spans,
					Style: t.Style{ForegroundColor: textColor},
				},
			},
		}
	}
}

// tagHighlighter returns a Highlighter that highlights #tags in the accent color.
func tagHighlighter(accentColor t.Color) t.HighlighterFunc {
	return func(text string, graphemes []string) []t.TextHighlight {
		matches := tagPattern.FindAllStringIndex(text, -1)
		if len(matches) == 0 {
			return nil
		}

		// Build a map from byte offset to grapheme index
		byteToGrapheme := make(map[int]int)
		bytePos := 0
		for i, g := range graphemes {
			byteToGrapheme[bytePos] = i
			bytePos += len(g)
		}
		byteToGrapheme[bytePos] = len(graphemes) // end position

		var highlights []t.TextHighlight
		for _, match := range matches {
			startGrapheme, ok1 := byteToGrapheme[match[0]]
			endGrapheme, ok2 := byteToGrapheme[match[1]]
			if ok1 && ok2 {
				highlights = append(highlights, t.TextHighlight{
					Start: startGrapheme,
					End:   endGrapheme,
					Style: t.SpanStyle{
						Foreground: accentColor,
						Italic:     true,
					},
				})
			}
		}
		return highlights
	}
}

// highlightTags creates spans with highlighted tags.
func (a *TodoApp) highlightTags(title string, baseStyle t.Style, tagColor t.Color) []t.Span {
	matches := tagPattern.FindAllStringIndex(title, -1)
	if len(matches) == 0 {
		return []t.Span{{Text: title, Style: styleToSpanStyle(baseStyle)}}
	}

	var spans []t.Span
	pos := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add text before the tag
		if start > pos {
			spans = append(spans, t.Span{Text: title[pos:start], Style: styleToSpanStyle(baseStyle)})
		}

		// Add the highlighted tag
		tagStyle := styleToSpanStyle(baseStyle)
		tagStyle.Foreground = tagColor
		tagStyle.Italic = true
		spans = append(spans, t.Span{Text: title[start:end], Style: tagStyle})

		pos = end
	}

	// Add remaining text after last tag
	if pos < len(title) {
		spans = append(spans, t.Span{Text: title[pos:], Style: styleToSpanStyle(baseStyle)})
	}

	return spans
}

// highlightMatchesAndTags creates spans with both filter matches and tags highlighted.
func (a *TodoApp) highlightMatchesAndTags(title string, baseStyle t.Style, matchColor t.Color, matchBgColor t.Color, tagColor t.Color) []t.Span {
	filterText := a.getFilterText()

	// Build a character-level style map
	type charStyle struct {
		isTag         bool
		isFilterMatch bool
	}
	styles := make([]charStyle, len(title))

	// Mark tag positions
	tagMatches := tagPattern.FindAllStringIndex(title, -1)
	for _, match := range tagMatches {
		for i := match[0]; i < match[1]; i++ {
			styles[i].isTag = true
		}
	}

	// Mark filter match positions
	if filterText != "" {
		titleLower := strings.ToLower(title)
		pos := 0
		for {
			idx := strings.Index(titleLower[pos:], filterText)
			if idx == -1 {
				break
			}
			start := pos + idx
			end := start + len(filterText)
			for i := start; i < end; i++ {
				styles[i].isFilterMatch = true
			}
			pos = end
		}
	}

	// Build spans by grouping consecutive characters with the same style
	var spans []t.Span
	pos := 0

	for pos < len(title) {
		currentStyle := styles[pos]
		end := pos + 1
		for end < len(title) && styles[end] == currentStyle {
			end++
		}

		spanStyle := styleToSpanStyle(baseStyle)
		if currentStyle.isTag {
			spanStyle.Foreground = tagColor
			spanStyle.Italic = true
		}
		if currentStyle.isFilterMatch {
			spanStyle.Underline = t.UnderlineSingle
			spanStyle.UnderlineColor = matchColor
			spanStyle.Background = matchBgColor
		}

		spans = append(spans, t.Span{Text: title[pos:end], Style: spanStyle})
		pos = end
	}

	return spans
}

// colorProviderToColor converts a ColorProvider to a Color.
func colorProviderToColor(cp t.ColorProvider) t.Color {
	if cp == nil {
		return t.Color{}
	}
	if c, ok := cp.(t.Color); ok {
		return c
	}
	return t.Color{}
}

func main() {
	t.SetTheme("kanagawa")
	_ = t.Run(NewTodoApp())
}
