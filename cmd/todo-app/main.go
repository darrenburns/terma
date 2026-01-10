//go:build ignore
// +build ignore

// DISABLED: Uses t.Dock and t.GradientBox which don't exist yet

package main

import (
	"fmt"
	"strings"
	"time"

	t "terma"
)

// Task represents a single TODO item.
type Task struct {
	ID        string
	Title     string
	Completed bool
	CreatedAt time.Time
}

// TodoApp is the main application widget.
type TodoApp struct {
	// Core state
	tasks       *t.ListState[Task]
	inputState  *t.TextInputState
	scrollState *t.ScrollState

	// Editing state
	editingIndex   t.Signal[int]
	editInputState *t.TextInputState

	// Theme picker state
	showThemePicker  t.Signal[bool]
	themeListState   *t.ListState[string]
	themeScrollState *t.ScrollState
	originalTheme    string

	// ID counter for generating unique task IDs
	nextID int
}

// NewTodoApp creates a new TODO application.
func NewTodoApp() *TodoApp {
	return &TodoApp{
		tasks:            t.NewListState([]Task{}),
		inputState:       t.NewTextInputState(""),
		scrollState:      t.NewScrollState(),
		editingIndex:     t.NewSignal(-1),
		editInputState:   t.NewTextInputState(""),
		showThemePicker:  t.NewSignal(false),
		themeListState:   t.NewListState(t.ThemeNames()),
		themeScrollState: t.NewScrollState(),
		nextID:           1,
	}
}

// generateID creates a unique ID for a new task.
func (a *TodoApp) generateID() string {
	id := fmt.Sprintf("task-%d", a.nextID)
	a.nextID++
	return id
}

// Build implements the Widget interface.
func (a *TodoApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Request focus on edit input when editing starts
	if a.editingIndex.Get() >= 0 {
		t.RequestFocus("edit-input")
	}

	// Request focus on theme list when picker opens
	if a.showThemePicker.Get() {
		t.RequestFocus("theme-list")
	}

	return t.Row{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Children: []t.Widget{
			t.GradientBox{
				Gradient: t.NewGradient(
					theme.Primary.Darken(0.6),
					theme.Background,
				),
				Width:  t.Flex(1),
				Height: t.Flex(1),
				Child: t.Dock{
					Style: t.Style{
						Padding: t.EdgeInsetsXY(2, 1), // Horizontal and vertical padding
					},
					Top: []t.Widget{
						a.buildHeader(theme),
						t.Text{Content: ""}, // Spacer
						a.buildInputRow(theme),
						t.Text{Content: ""}, // Spacer
					},
					Bottom: []t.Widget{
						t.Text{Content: ""}, // Spacer
						a.buildStatusBar(theme),
						t.KeybindBar{},
					},
					Body: a.buildTaskList(ctx),
				},
			},
			a.buildThemePicker(theme),
		},
	}
}

// buildHeader creates the header with app title.
func (a *TodoApp) buildHeader(theme t.ThemeData) t.Widget {
	return t.Text{
		Content: "Todo",
		Style: t.Style{
			ForegroundColor: theme.Primary,
			Bold:            true,
		},
	}
}

// buildInputRow creates the new task input row.
func (a *TodoApp) buildInputRow(theme t.ThemeData) t.Widget {
	return t.Row{
		Children: []t.Widget{
			t.Text{
				Content: "  ＋ ",
				Style: t.Style{
					ForegroundColor: theme.Accent,
				},
			},
			t.TextInput{
				ID:          "new-task-input",
				State:       a.inputState,
				Placeholder: "What needs to be done?",
				Width:       t.Flex(1),
				OnSubmit:    a.addTask,
			},
		},
	}
}

// buildTaskList creates the scrollable task list.
func (a *TodoApp) buildTaskList(ctx t.BuildContext) t.Widget {
	// Check if the list is focused
	listFocused := ctx.Focused() != nil && ctx.IsFocused(t.List[Task]{ID: "task-list"})

	return t.Scrollable{
		State:  a.scrollState,
		Height: t.Flex(1),
		Child: t.List[Task]{
			ID:          "task-list",
			State:       a.tasks,
			ScrollState: a.scrollState,
			RenderItem:  a.renderTaskItem(ctx, listFocused),
			OnSelect:    a.toggleTask,
		},
	}
}

// renderTaskItem returns the render function for task items.
func (a *TodoApp) renderTaskItem(ctx t.BuildContext, listFocused bool) func(Task, bool, bool) t.Widget {
	theme := ctx.Theme()
	editingIdx := a.editingIndex.Get()

	return func(task Task, active bool, selected bool) t.Widget {
		// Find the index of this task
		tasks := a.tasks.GetItems()
		itemIndex := -1
		for i, tsk := range tasks {
			if tsk.ID == task.ID {
				itemIndex = i
				break
			}
		}

		// If this task is being edited, show TextInput
		if editingIdx == itemIndex {
			// Capture index for closures
			idx := itemIndex
			itemCount := a.tasks.ItemCount()

			// Align with normal display: "  ○  " = prefix + circle + spacing
			return t.Row{
				Width: t.Flex(1),
				Children: []t.Widget{
					t.Text{Content: "  ○  "}, // Match the prefix + circle + space
					t.TextInput{
						ID:    "edit-input",
						State: a.editInputState,
						Width: t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Surface,
						},
						OnSubmit: func(text string) {
							a.saveEdit(idx, text)
						},
						ExtraKeybinds: []t.Keybind{
							{Key: "up", Action: func() {
								a.editingIndex.Set(-1)
								if idx == 0 {
									t.RequestFocus("new-task-input")
								} else {
									a.tasks.SelectPrevious()
									t.RequestFocus("task-list")
								}
							}, Hidden: true},
							{Key: "down", Action: func() {
								a.editingIndex.Set(-1)
								if idx < itemCount-1 {
									a.tasks.SelectNext()
								}
								t.RequestFocus("task-list")
							}, Hidden: true},
						},
					},
				},
			}
		}

		// Normal display mode
		checkbox := "○"
		if task.Completed {
			checkbox = "●"
		}

		prefix := "  "
		textStyle := t.Style{ForegroundColor: theme.Text}
		checkboxStyle := t.Style{ForegroundColor: theme.TextMuted}

		if active && listFocused {
			// Show cursor and accent color only when list is focused
			prefix = "▸ "
			textStyle.ForegroundColor = theme.Accent
			checkboxStyle.ForegroundColor = theme.Accent
		}

		if task.Completed {
			textStyle.ForegroundColor = theme.TextMuted
			checkboxStyle.ForegroundColor = theme.Success
		}

		return t.Row{
			Width: t.Flex(1),
			Children: []t.Widget{
				t.Text{
					Content: prefix,
					Style:   textStyle,
				},
				t.Text{
					Content: checkbox + "  ",
					Style:   checkboxStyle,
				},
				t.Text{
					Content: task.Title,
					Style:   textStyle,
					Width:   t.Flex(1),
				},
			},
		}
	}
}

// buildThemePicker creates the theme picker modal.
func (a *TodoApp) buildThemePicker(theme t.ThemeData) t.Widget {
	return t.Floating{
		Visible: a.showThemePicker.Get(),
		Config: t.FloatConfig{
			Position:  t.FloatPositionCenter,
			Modal:     true,
			OnDismiss: a.dismissThemePicker,
		},
		Child: t.Column{
			Width: t.Cells(40),
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Border:          t.RoundedBorder(theme.Primary),
				Padding:         t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				t.Text{
					Content: "Select Theme",
					Style: t.Style{
						ForegroundColor: theme.Primary,
						Bold:            true,
					},
				},
				t.Text{Content: ""}, // Spacer
				t.Scrollable{
					State:  a.themeScrollState,
					Height: t.Cells(12),
					Child: t.List[string]{
						ID:             "theme-list",
						State:          a.themeListState,
						ScrollState:    a.themeScrollState,
						OnSelect:       a.selectTheme,
						OnCursorChange: a.previewTheme,
						RenderItem:     a.renderThemeItem(theme),
					},
				},
				t.Text{Content: ""}, // Spacer
				t.Text{
					Content: "↑↓ navigate · enter select · esc cancel",
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},
			},
		},
	}
}

// renderThemeItem returns the render function for theme list items.
func (a *TodoApp) renderThemeItem(theme t.ThemeData) func(string, bool, bool) t.Widget {
	currentTheme := t.CurrentThemeName()
	return func(themeName string, active bool, selected bool) t.Widget {
		prefix := "  "
		suffix := ""
		style := t.Style{ForegroundColor: theme.Text}

		if active {
			prefix = "▸ "
			style.ForegroundColor = theme.Accent
		}

		if themeName == currentTheme {
			suffix = " ✓"
			if !active {
				style.ForegroundColor = theme.Success
			}
		}

		return t.Text{
			Content: prefix + themeName + suffix,
			Style:   style,
			Width:   t.Flex(1),
		}
	}
}

// buildStatusBar shows the count of remaining items.
func (a *TodoApp) buildStatusBar(theme t.ThemeData) t.Widget {
	tasks := a.tasks.GetItems()
	active := 0
	completed := 0
	for _, task := range tasks {
		if task.Completed {
			completed++
		} else {
			active++
		}
	}

	var status string
	if len(tasks) == 0 {
		status = "No tasks yet"
	} else if active == 0 {
		status = "All done! 🎉"
	} else {
		itemWord := "tasks"
		if active == 1 {
			itemWord = "task"
		}
		status = fmt.Sprintf("%d %s remaining", active, itemWord)
		if completed > 0 {
			status += fmt.Sprintf(" · %d completed", completed)
		}
	}

	return t.Text{
		Content: status,
		Style: t.Style{
			ForegroundColor: theme.TextMuted,
		},
	}
}

// Keybinds returns the declarative keybindings for the app.
func (a *TodoApp) Keybinds() []t.Keybind {
	editingIdx := a.editingIndex.Peek()
	isEditing := editingIdx >= 0
	isThemePicker := a.showThemePicker.Peek()

	// Theme picker modal has its own keybinds (handled via Float dismiss)
	if isThemePicker {
		return []t.Keybind{
			{Key: "escape", Name: "Cancel", Action: a.dismissThemePicker},
		}
	}

	keybinds := []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
		// Navigation between input and list (these bubble up from focused widgets)
		{Key: "up", Action: a.navigateUp, Hidden: true},
		{Key: "down", Action: a.navigateDown, Hidden: true},
	}

	if !isEditing {
		keybinds = append(keybinds,
			t.Keybind{Key: "enter", Name: "Toggle", Action: a.toggleCurrentTask, Hidden: true},
			t.Keybind{Key: " ", Name: "Toggle", Action: a.toggleCurrentTask},
			t.Keybind{Key: "e", Name: "Edit", Action: a.startEdit},
			t.Keybind{Key: "d", Name: "Delete", Action: a.deleteCurrentTask},
			t.Keybind{Key: "t", Name: "Theme", Action: a.openThemePicker},
		)
	} else {
		keybinds = append(keybinds,
			t.Keybind{Key: "escape", Name: "Cancel", Action: a.cancelEdit},
		)
	}

	return keybinds
}

// navigateUp handles up arrow for cross-widget navigation.
// Called when: input focused, list at top item, or in edit mode.
func (a *TodoApp) navigateUp() {
	editingIdx := a.editingIndex.Peek()

	if editingIdx >= 0 {
		// In edit mode: cancel edit and move cursor up (or to input if at top)
		a.editingIndex.Set(-1)
		if editingIdx == 0 {
			t.RequestFocus("new-task-input")
		} else {
			a.tasks.SelectPrevious()
			t.RequestFocus("task-list")
		}
	} else {
		// List at top or input focused - move to input
		t.RequestFocus("new-task-input")
	}
}

// navigateDown handles down arrow for cross-widget navigation.
// Called when: input focused, list at bottom item, or in edit mode.
func (a *TodoApp) navigateDown() {
	editingIdx := a.editingIndex.Peek()

	if editingIdx >= 0 {
		// In edit mode: cancel edit and move cursor down
		a.editingIndex.Set(-1)
		itemCount := a.tasks.ItemCount()
		if editingIdx < itemCount-1 {
			a.tasks.SelectNext()
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
	a.tasks.Append(task)
	a.inputState.SetText("")
}

// toggleCurrentTask toggles the completion status of the selected task.
func (a *TodoApp) toggleCurrentTask() {
	if task, ok := a.tasks.SelectedItem(); ok {
		a.toggleTask(task)
	}
}

// toggleTask toggles the completion status of the given task.
func (a *TodoApp) toggleTask(task Task) {
	tasks := a.tasks.GetItems()
	for i, tsk := range tasks {
		if tsk.ID == task.ID {
			tasks[i].Completed = !tasks[i].Completed
			a.tasks.SetItems(tasks)
			return
		}
	}
}

// deleteCurrentTask removes the currently selected task.
func (a *TodoApp) deleteCurrentTask() {
	idx := a.tasks.CursorIndex.Peek()
	a.tasks.RemoveAt(idx)
}

// startEdit begins inline editing of the current task.
func (a *TodoApp) startEdit() {
	idx := a.tasks.CursorIndex.Peek()
	tasks := a.tasks.GetItems()
	if idx >= 0 && idx < len(tasks) {
		a.editInputState.SetText(tasks[idx].Title)
		a.editInputState.CursorEnd() // Position cursor at end of text
		a.editingIndex.Set(idx)
	}
}

// saveEdit saves the edited task title.
func (a *TodoApp) saveEdit(index int, newTitle string) {
	newTitle = strings.TrimSpace(newTitle)
	if newTitle == "" {
		a.cancelEdit()
		return
	}

	tasks := a.tasks.GetItems()
	if index >= 0 && index < len(tasks) {
		tasks[index].Title = newTitle
		a.tasks.SetItems(tasks)
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

	// Find and select current theme in list
	themes := a.themeListState.GetItems()
	for i, name := range themes {
		if name == a.originalTheme {
			a.themeListState.SelectIndex(i)
			break
		}
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

func main() {
	t.SetTheme("rose-pine")
	t.Run(NewTodoApp())
}
