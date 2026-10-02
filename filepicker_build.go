package terma

import "fmt"

// Build composes existing controls. It only reads reactive state; directory IO
// and selection changes belong to setup and handlers.
func (p FilePicker) Build(ctx BuildContext) Widget {
	if p.State == nil {
		return Text{Content: "FilePicker requires State"}
	}
	s := p.State
	theme := ctx.Theme()
	hidden := s.ShowHidden.Get()
	_ = s.FilterIndex.Get()
	// List caches matches by query. This matcher also depends on picker
	// configuration, so invalidate only the private derived cache here.
	viewKey := fmt.Sprintf("%d/%t/%d/%v", p.Mode, hidden, p.filterIndex(), p.Filters)
	if s.viewKey != viewKey {
		s.list.resetFilterCache()
		s.viewKey = viewKey
	}
	rows := s.list.Items.Get()
	visible := 0
	for _, e := range rows {
		if p.visible(e) {
			visible++
		}
	}
	filterLabel := "All files"
	if len(p.Filters) > 0 {
		filterLabel = p.Filters[p.filterIndex()].Label
		if filterLabel == "" {
			filterLabel = fmt.Sprintf("Filter %d", p.filterIndex()+1)
		}
	}
	hiddenLabel := "Hidden: off"
	if hidden {
		hiddenLabel = "Hidden: on"
	}
	modeLabel := "Open file"
	acceptLabel := "Open"
	if p.MultiSelect && p.Mode == FilePickerOpen {
		modeLabel = "Open files"
	}
	if p.Mode == FilePickerSave {
		modeLabel = "Save file"
		acceptLabel = "Save"
	}
	if p.Mode == FilePickerDirectory {
		modeLabel = "Choose directory"
		acceptLabel = "Choose"
	}
	errorText := s.Error.Get()
	if err := p.filterError(); err != nil {
		errorText = err.Error()
	}
	completed := s.completed.Get()
	status := fmt.Sprintf("%s · %d visible", modeLabel, visible)
	if completed {
		status = "Selection finished · navigate or Reset to start again"
	}
	content := Widget(Scrollable{ID: p.id("scroll"), State: s.scroll, Style: Style{Width: Flex(1), Height: Flex(1)}, Child: filePickerList{picker: p, List: List[FilePickerEntry]{ID: p.id("list"), State: s.list, ScrollState: s.scroll, Filter: s.filter, MultiSelect: p.MultiSelect && p.Mode == FilePickerOpen,
		MatchItem: func(e FilePickerEntry, _ string, _ FilterOptions) MatchResult {
			return MatchResult{Matched: p.visible(e)}
		},
		RenderItem: func(e FilePickerEntry, active, selected bool) Widget {
			prefix := "     "
			if p.MultiSelect && p.Mode == FilePickerOpen {
				prefix = "[ ]  "
				if selected {
					prefix = "[x]  "
				}
			}
			kind := "     "
			if e.IsDir {
				kind = "[D]  "
			}
			if e.IsSymlink {
				kind = "[L]  "
			}
			if e.Err != nil {
				kind = "[!]  "
			}
			style := Style{Width: Flex(1)}
			if active {
				style.BackgroundColor = theme.Primary
				style.ForegroundColor = theme.Background
			}
			if e.Err != nil {
				style.ForegroundColor = theme.Error
			}
			return Text{Content: prefix + kind + e.Name, Style: style}
		},
		OnSelect: p.activate,
		OnCursorChange: func(e FilePickerEntry) {
			if p.Mode == FilePickerSave && !e.IsDir && e.Err == nil {
				s.FilenameInput.SetText(e.Name)
			}
		},
	}}})
	if visible == 0 {
		content = Column{Style: Style{Width: Flex(1), Height: Flex(1)}, Children: []Widget{Text{Content: "No matching entries", Style: Style{ForegroundColor: theme.TextMuted}}}}
	}
	children := []Widget{
		Text{Content: status, Style: Style{Bold: true}},
		Text{Content: "Directory path (Enter to navigate)"},
		TextInput{ID: p.id("path"), State: s.PathInput, Style: Style{Width: Flex(1)}, OnSubmit: func(path string) {
			if s.Navigate(path) == nil {
				RequestFocus(p.id("list"))
			}
		}},
		Row{Spacing: 1, Children: []Widget{Button{ID: p.id("parent"), Label: "Up", OnPress: p.parent}, Button{ID: p.id("refresh"), Label: "Refresh", OnPress: func() { _ = s.Refresh() }}, Button{ID: p.id("hidden"), Label: hiddenLabel, OnPress: p.toggleHidden}, Button{ID: p.id("filter"), Label: "Filter: " + filterLabel, OnPress: p.nextFilter}}},
		content,
	}
	if p.Mode == FilePickerSave {
		children = append(children, Text{Content: "Filename"}, TextInput{ID: p.id("filename"), State: s.FilenameInput, Style: Style{Width: Flex(1)}, OnSubmit: func(string) { p.Submit() }})
	}
	if errorText != "" {
		children = append(children, Text{Content: errorText, Wrap: WrapSoft, Style: Style{ForegroundColor: theme.Error}})
	}
	children = append(children, Row{Spacing: 2, Children: []Widget{Button{ID: p.id("accept"), Label: acceptLabel, Variant: ButtonPrimary, OnPress: func() { p.Submit() }}, Button{ID: p.id("cancel"), Label: "Cancel", OnPress: p.Cancel}}}, Text{Content: "↑↓ navigate · Enter activate · Tab controls · Esc cancel", Style: Style{ForegroundColor: theme.TextMuted}}, Dialog{ID: p.id("overwrite"), Visible: s.pending.Get() != "", Title: "Confirm overwrite", Content: Text{Content: "A file already exists at:\n" + s.pending.Peek() + "\nReturn this path for replacement?", Wrap: WrapSoft}, Buttons: []Button{{Label: "Keep existing", OnPress: p.cancelOverwrite}, {Label: "Overwrite", Variant: ButtonWarning, OnPress: p.confirmOverwrite}}, OnDismiss: p.cancelOverwrite})
	style := p.Style
	if style.Width.IsUnset() {
		style.Width = Flex(1)
	}
	if style.Height.IsUnset() {
		style.Height = Flex(1)
	}
	return Column{Spacing: 1, Style: style, Children: children}
}

// filePickerList adds selection toggling to the list's navigation bindings.
// Keeping this binding on the list leaves spaces in path/filename inputs alone.
type filePickerList struct {
	List[FilePickerEntry]
	picker FilePicker
}

func (l filePickerList) Keybinds() []Keybind {
	binds := l.List.Keybinds()
	if l.MultiSelect {
		binds = append(binds, Keybind{Key: "space", Name: "Toggle file", Action: func() {
			if _, _, ok := l.normalizeCursorForInteraction(); !ok {
				return
			}
			item, ok := l.State.SelectedItem()
			if ok && !item.IsDir && item.Err == nil && l.picker.visible(item) {
				l.State.ToggleSelection(l.State.CursorIndex.Peek())
			}
		}})
	}
	return binds
}
