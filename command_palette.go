package terma

import (
	"strings"
)

const (
	defaultCommandPaletteWidth       = 60
	defaultCommandPaletteHeight      = 12
	defaultCommandPalettePlaceholder = "Type to search..."
	defaultCommandPaletteEmptyLabel  = "No results"
)

var commandPaletteDividerLine = strings.Repeat("─", 120)

// CommandPaletteItem represents a single entry in the command palette.
type CommandPaletteItem struct {
	Label         string                      // Primary display text
	Hint          string                      // Right-aligned text (e.g., "Ctrl+N")
	HintWidget    func() Widget               // Custom hint widget (color swatches, etc.)
	Description   string                      // Optional secondary line below label
	Action        func()                      // Called on selection (ignored if Children set)
	Children      func() []CommandPaletteItem // Opens nested palette (lazy-loaded)
	ChildrenTitle string                      // Breadcrumb title for nested level
	Disabled      bool                        // Grayed out, not selectable
	Divider       string                      // Empty = plain line, non-empty = titled
	FilterText    string                      // Override Label for filtering
	Data          any                         // User data for custom renderers
}

// IsSelectable returns true if the item can be focused/selected.
func (i CommandPaletteItem) IsSelectable() bool {
	return !i.Disabled && !i.IsDivider()
}

// IsDivider returns true if this item should render as a divider.
func (i CommandPaletteItem) IsDivider() bool {
	if i.Divider != "" {
		return true
	}
	return i.Label == "" && i.Hint == "" && i.Description == "" && i.Action == nil && i.Children == nil && i.HintWidget == nil
}

// GetFilterText returns the text used for filtering.
func (i CommandPaletteItem) GetFilterText() string {
	if i.FilterText != "" {
		return i.FilterText
	}
	return i.Label
}

// CommandPaletteLevel holds the state for a single palette level.
type CommandPaletteLevel struct {
	Title       string
	Items       []CommandPaletteItem
	ListState   *ListState[CommandPaletteItem]
	ScrollState *ScrollState
	FilterState *FilterState
	InputState  *TextInputState
}

// CommandPaletteState holds the stack of palette levels and visibility state.
type CommandPaletteState struct {
	stack   []*CommandPaletteLevel
	Visible Signal[bool]
	depth   Signal[int]

	wasVisible  bool
	lastFocusID string
}

// NewCommandPaletteState creates a new palette state with a root level.
func NewCommandPaletteState(title string, items []CommandPaletteItem) *CommandPaletteState {
	level := newCommandPaletteLevel(title, items)
	state := &CommandPaletteState{
		stack:   []*CommandPaletteLevel{level},
		Visible: NewSignal(false),
		depth:   NewSignal(1),
	}
	state.ensureSelectableCursor(level)
	return state
}

// Open shows the command palette.
func (s *CommandPaletteState) Open() {
	if s == nil {
		return
	}
	s.Visible.Set(true)
}

// Close hides the command palette.
// If keepPosition is false, the palette resets to the root level for the next open.
func (s *CommandPaletteState) Close(keepPosition bool) {
	if s == nil {
		return
	}
	s.Visible.Set(false)
	if !keepPosition {
		s.resetToRoot()
	}
}

// PushLevel adds a new level to the stack.
func (s *CommandPaletteState) PushLevel(title string, items []CommandPaletteItem) {
	if s == nil {
		return
	}
	level := newCommandPaletteLevel(title, items)
	s.stack = append(s.stack, level)
	s.depth.Set(len(s.stack))
	s.ensureSelectableCursor(level)
}

// PopLevel removes the current level, returning true if a level was popped.
func (s *CommandPaletteState) PopLevel() bool {
	if s == nil || len(s.stack) <= 1 {
		return false
	}
	s.stack = s.stack[:len(s.stack)-1]
	s.depth.Set(len(s.stack))
	return true
}

// CurrentLevel returns the active level.
func (s *CommandPaletteState) CurrentLevel() *CommandPaletteLevel {
	if s == nil || len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

// CurrentItem returns the currently selected item (if any).
func (s *CommandPaletteState) CurrentItem() (CommandPaletteItem, bool) {
	level := s.CurrentLevel()
	if level == nil || level.ListState == nil {
		return CommandPaletteItem{}, false
	}
	cursor := level.ListState.CursorIndex.Peek()
	view := commandPaletteFilteredView(level.Items, level.FilterState)
	for _, idx := range view.Indices {
		if idx == cursor && level.Items[idx].IsSelectable() {
			return level.Items[idx], true
		}
	}
	return CommandPaletteItem{}, false
}

// BreadcrumbPath returns the current breadcrumb trail.
func (s *CommandPaletteState) BreadcrumbPath() []string {
	if s == nil {
		return nil
	}
	path := make([]string, 0, len(s.stack))
	for _, level := range s.stack {
		if level == nil || level.Title == "" {
			continue
		}
		path = append(path, level.Title)
	}
	return path
}

// IsNested returns true if there is more than one level.
func (s *CommandPaletteState) IsNested() bool {
	return s != nil && len(s.stack) > 1
}

// SetItems replaces the items in the current level.
func (s *CommandPaletteState) SetItems(items []CommandPaletteItem) {
	level := s.CurrentLevel()
	if level == nil {
		return
	}
	if items == nil {
		items = []CommandPaletteItem{}
	}
	level.Items = items
	if level.ListState != nil {
		level.ListState.SetItems(items)
	}
	s.ensureSelectableCursor(level)
}

func (s *CommandPaletteState) resetToRoot() {
	root := s.stack[0]
	// Clear root input text
	if root != nil && root.InputState != nil {
		root.InputState.SetText("")
		if root.FilterState != nil {
			root.FilterState.Query.Set("")
		}
	}
	if len(s.stack) <= 1 {
		s.ensureSelectableCursor(root)
		return
	}
	s.stack = []*CommandPaletteLevel{root}
	s.depth.Set(1)
	s.ensureSelectableCursor(root)
}

func (s *CommandPaletteState) ensureSelectableCursor(level *CommandPaletteLevel) {
	if level == nil || level.ListState == nil {
		return
	}
	if len(level.Items) == 0 {
		level.ListState.CursorIndex.Set(0)
		return
	}
	view := commandPaletteFilteredView(level.Items, level.FilterState)
	if len(view.Indices) == 0 {
		level.ListState.CursorIndex.Set(0)
		return
	}
	cursor := level.ListState.CursorIndex.Peek()
	for _, idx := range view.Indices {
		if idx == cursor && level.Items[idx].IsSelectable() {
			return
		}
	}
	if first, ok := firstSelectableIndex(level.Items, view.Indices); ok {
		level.ListState.SelectIndex(first)
		return
	}
	level.ListState.SelectIndex(view.Indices[0])
}

func newCommandPaletteLevel(title string, items []CommandPaletteItem) *CommandPaletteLevel {
	if items == nil {
		items = []CommandPaletteItem{}
	}
	level := &CommandPaletteLevel{
		Title:       title,
		Items:       items,
		ListState:   NewListState(items),
		ScrollState: NewScrollState(),
		FilterState: NewFilterState(),
		InputState:  NewTextInputState(""),
	}
	return level
}

func firstSelectableIndex(items []CommandPaletteItem, indices []int) (int, bool) {
	for _, idx := range indices {
		if idx >= 0 && idx < len(items) && items[idx].IsSelectable() {
			return idx, true
		}
	}
	return 0, false
}

func lastSelectableIndex(items []CommandPaletteItem, indices []int) (int, bool) {
	for i := len(indices) - 1; i >= 0; i-- {
		idx := indices[i]
		if idx >= 0 && idx < len(items) && items[idx].IsSelectable() {
			return idx, true
		}
	}
	return 0, false
}

func commandPaletteFilteredView(items []CommandPaletteItem, filter *FilterState) FilteredView[CommandPaletteItem] {
	query, options := filterStateValuesPeek(filter)
	matchItem := func(item CommandPaletteItem, query string) MatchResult {
		return commandPaletteMatchItem(item, query, options)
	}
	return ApplyFilter(items, query, matchItem)
}

func commandPaletteMatchItem(item CommandPaletteItem, query string, options FilterOptions) MatchResult {
	if item.IsDivider() {
		return MatchResult{Matched: true}
	}

	if query == "" {
		return MatchResult{Matched: true}
	}

	result := MatchString(item.GetFilterText(), query, options)
	if !result.Matched {
		return result
	}

	if item.FilterText != "" && item.FilterText != item.Label {
		labelMatch := MatchString(item.Label, query, options)
		result.Ranges = labelMatch.Ranges
	}

	return result
}

// CommandPalette renders a searchable command palette with nested navigation.
type CommandPalette struct {
	ID             string
	State          *CommandPaletteState
	OnSelect       func(item CommandPaletteItem) // Custom selection handler
	OnCursorChange func(item CommandPaletteItem) // For live previews
	OnDismiss      func()
	RenderItem     func(item CommandPaletteItem, active bool, match MatchResult) Widget
	Width          Dimension // Deprecated: use Style.Width (default: Cells(60))
	Height         Dimension // Deprecated: use Style.Height (default: Cells(12))
	Placeholder    string    // Default: "Type to search..."
	Position       FloatPosition
	Offset         Offset
	Style          Style
}

// Build renders the command palette as a floating modal.
func (p CommandPalette) Build(ctx BuildContext) Widget {
	if p.State == nil {
		return EmptyWidget{}
	}

	visible := p.State.Visible.Get()
	_ = p.State.depth.Get()

	if visible && !p.State.wasVisible {
		if focused := ctx.Focused(); focused != nil {
			if identifiable, ok := focused.(Identifiable); ok {
				p.State.lastFocusID = identifiable.WidgetID()
			}
		}
	}

	if visible {
		RequestFocus(p.inputID())
	} else if p.State.wasVisible {
		if p.State.lastFocusID != "" {
			RequestFocus(p.State.lastFocusID)
		}
	}
	p.State.wasVisible = visible

	if !visible {
		return EmptyWidget{}
	}

	level := p.State.CurrentLevel()
	if level == nil {
		return EmptyWidget{}
	}

	content := p.buildContent(ctx, level)
	float := Floating{
		Visible: true,
		Config: FloatConfig{
			Position:              p.floatPosition(),
			Offset:                p.Offset,
			Modal:                 true,
			DismissOnEsc:          BoolPtr(false),
			DismissOnClickOutside: BoolPtr(true),
			OnDismiss:             p.dismiss,
			BackdropColor:         ctx.Theme().Overlay,
		},
		Child: content,
	}
	return float.Build(ctx)
}

func (p CommandPalette) buildContent(ctx BuildContext, level *CommandPaletteLevel) Widget {
	theme := ctx.Theme()

	containerStyle := p.containerStyle(theme)

	children := make([]Widget, 0, 4)

	headerChildren := make([]Widget, 0, 2)
	if path := p.State.BreadcrumbPath(); len(path) > 0 {
		headerChildren = append(headerChildren, Breadcrumbs{
			ID:        p.ID + "-breadcrumbs",
			Path:      path,
			OnSelect:  p.onBreadcrumbSelect(),
			Separator: ">",
			Style: Style{
				ForegroundColor: theme.TextMuted,
				Width:           Flex(1),
			},
		})
	}
	headerChildren = append(headerChildren, p.buildInput(level, theme))

	children = append(children, Column{
		CrossAlign: CrossAxisStretch,
		Spacing:    0,
		Children:   headerChildren,
	})
	children = append(children, p.buildList(ctx, level, theme))

	if containerStyle.Width.IsUnset() {
		containerStyle.Width = p.paletteWidth()
	}
	if containerStyle.Height.IsUnset() {
		containerStyle.Height = Auto
	}
	if containerStyle.MaxHeight.IsUnset() {
		containerStyle.MaxHeight = p.paletteHeight()
	}
	return Column{
		ID:         p.ID + "-content",
		CrossAlign: CrossAxisStretch,
		Style:      containerStyle,
		Children:   children,
	}
}

func (p CommandPalette) buildInput(level *CommandPaletteLevel, theme ThemeData) Widget {
	if level == nil {
		return EmptyWidget{}
	}

	padding := EdgeInsetsTRBL(1, 1, 1, 1)

	onFilterChange := func(text string) {
		if level.FilterState != nil {
			level.FilterState.Query.Set(text)
		}
		if level.ScrollState != nil {
			level.ScrollState.SetOffset(0)
		}
		if p.State != nil {
			p.State.ensureSelectableCursor(level)
		}
		p.notifyCursorChange()
	}

	return TextInput{
		ID:          p.inputID(),
		State:       level.InputState,
		Placeholder: p.placeholderText(),
		Style: Style{
			BackgroundColor: theme.Surface,
			ForegroundColor: theme.Text,
			Padding:         padding,
			Width:           Flex(1),
		},
		OnChange: onFilterChange,
		ExtraKeybinds: []Keybind{
			{Key: "up", Action: func() { p.moveCursor(-1) }, Hidden: true},
			{Key: "down", Action: func() { p.moveCursor(1) }, Hidden: true},
			{Key: "ctrl+p", Action: func() { p.moveCursor(-1) }, Hidden: true},
			{Key: "ctrl+n", Action: func() { p.moveCursor(1) }, Hidden: true},
			{Key: "home", Action: func() { p.moveCursorToStart() }, Hidden: true},
			{Key: "end", Action: func() { p.moveCursorToEnd() }, Hidden: true},
			{Key: "enter", Action: p.selectCurrent, Hidden: true},
			{Key: "escape", Action: p.handleEscape, Hidden: true},
			{Key: "backspace", Action: func() { p.handleBackspace(level, onFilterChange) }, Hidden: true},
		},
	}
}

func (p CommandPalette) buildList(ctx BuildContext, level *CommandPaletteLevel, theme ThemeData) Widget {
	if level == nil {
		return EmptyWidget{}
	}

	view := commandPaletteFilteredView(level.Items, level.FilterState)
	hasContent := false
	for _, idx := range view.Indices {
		if idx >= 0 && idx < len(level.Items) && !level.Items[idx].IsDivider() {
			hasContent = true
			break
		}
	}

	var listChild Widget
	if !hasContent {
		listChild = Text{
			Content:   defaultCommandPaletteEmptyLabel,
			TextAlign: TextAlignCenter,
			Style: Style{
				ForegroundColor: theme.TextMuted,
				Padding:         EdgeInsetsXY(1, 0),
				Width:           Flex(1),
			},
		}
	} else {
		listChild = List[CommandPaletteItem]{
			ID:                  p.listID(),
			State:               level.ListState,
			ScrollState:         level.ScrollState,
			Filter:              level.FilterState,
			MatchItem:           commandPaletteMatchItem,
			RenderItemWithMatch: p.renderItem(ctx),
			Style: Style{
				BackgroundColor: theme.Surface,
			},
		}
	}

	return Scrollable{
		ID:    p.scrollID(),
		State: level.ScrollState,
		Style: Style{
			BackgroundColor: theme.Surface,
		},
		Child: listChild,
	}
}

func (p CommandPalette) renderItem(ctx BuildContext) func(item CommandPaletteItem, active bool, selected bool, match MatchResult) Widget {
	theme := ctx.Theme()
	return func(item CommandPaletteItem, active bool, selected bool, match MatchResult) Widget {
		if p.RenderItem != nil {
			active = active && item.IsSelectable()
			return p.RenderItem(item, active, match)
		}
		return p.defaultRenderItem(theme, item, active, match)
	}
}

func (p CommandPalette) defaultRenderItem(theme ThemeData, item CommandPaletteItem, active bool, match MatchResult) Widget {
	active = active && item.IsSelectable()

	if item.IsDivider() {
		return p.dividerWidget(theme, item.Divider)
	}

	itemStyle := Style{
		Padding: EdgeInsetsXY(1, 0),
	}
	labelStyle := Style{ForegroundColor: theme.Text}
	descStyle := Style{ForegroundColor: theme.TextMuted}
	hintStyle := Style{ForegroundColor: theme.TextMuted}

	if item.Disabled {
		labelStyle.ForegroundColor = theme.TextDisabled
		descStyle.ForegroundColor = theme.TextDisabled
		hintStyle.ForegroundColor = theme.TextDisabled
	} else if active {
		itemStyle.BackgroundColor = theme.ActiveCursor
		labelStyle.ForegroundColor = theme.SelectionText
		descStyle.ForegroundColor = theme.SelectionText
		hintStyle.ForegroundColor = theme.SelectionText
	}

	labelWidget := p.labelWidget(item.Label, labelStyle, match, theme)
	hintWidget := p.hintWidget(item, hintStyle)
	rowChildren := []Widget{labelWidget}
	if hintWidget != nil {
		rowChildren = append(rowChildren, Spacer{Width: Flex(1)}, hintWidget)
	}
	rowStyle := Style{Width: Flex(1)}
	row := Row{
		Style:      rowStyle,
		CrossAlign: CrossAxisCenter,
		Children:   rowChildren,
	}

	var content Widget = row
	if item.Description != "" {
		descStyle.Width = Flex(1)
		content = Column{
			Children: []Widget{row, Text{Content: item.Description, Style: descStyle}},
			Style:    Style{Width: Flex(1)},
		}
	}

	return Stack{
		Children: []Widget{
			Column{
				Style: func() Style {
					style := itemStyle
					style.Width = Flex(1)
					return style
				}(),
				Children: []Widget{content},
			},
		},
	}
}

func (p CommandPalette) labelWidget(label string, style Style, match MatchResult, theme ThemeData) Widget {
	if style.Width.IsUnset() {
		style.Width = Flex(1)
	}
	if match.Matched && len(match.Ranges) > 0 {
		return Text{
			Spans: HighlightSpans(label, match.Ranges, MatchHighlightStyle(theme)),
			Style: style,
		}
	}
	return Text{
		Content: label,
		Style:   style,
	}
}

func (p CommandPalette) hintWidget(item CommandPaletteItem, style Style) Widget {
	if item.HintWidget != nil {
		return item.HintWidget()
	}
	if item.Hint == "" {
		return nil
	}
	return Text{Content: item.Hint, Style: style}
}

func (p CommandPalette) dividerWidget(theme ThemeData, title string) Widget {
	lineStyle := Style{ForegroundColor: theme.TextMuted}
	dividerPadding := EdgeInsetsTRBL(1, 1, 0, 1)
	if title == "" {
		return Text{
			Content: commandPaletteDividerLine,
			Style:   Style{ForegroundColor: lineStyle.ForegroundColor, Padding: dividerPadding, Width: Flex(1)},
		}
	}

	return Row{
		Style: Style{Padding: dividerPadding, Width: Flex(1)},
		Children: []Widget{
			Text{
				Content: title + " ",
				Style:   Style{ForegroundColor: lineStyle.ForegroundColor, Bold: true},
			},
			Text{
				Content: commandPaletteDividerLine,
				Style: func() Style {
					style := lineStyle
					style.Width = Flex(1)
					return style
				}(),
			},
		},
	}
}

func (p CommandPalette) moveCursor(delta int) {
	level := p.State.CurrentLevel()
	if level == nil || level.ListState == nil || len(level.Items) == 0 {
		return
	}
	view := commandPaletteFilteredView(level.Items, level.FilterState)
	if len(view.Indices) == 0 {
		return
	}

	cursor := level.ListState.CursorIndex.Peek()
	cursorViewIdx := indexOf(view.Indices, cursor)
	if cursorViewIdx < 0 {
		if first, ok := firstSelectableIndex(level.Items, view.Indices); ok {
			level.ListState.SelectIndex(first)
			p.notifyCursorChange()
		}
		return
	}

	step := 1
	if delta < 0 {
		step = -1
	}

	for i := cursorViewIdx + step; i >= 0 && i < len(view.Indices); i += step {
		idx := view.Indices[i]
		if level.Items[idx].IsSelectable() {
			level.ListState.SelectIndex(idx)
			p.notifyCursorChange()
			return
		}
	}
}

func (p CommandPalette) moveCursorToStart() {
	level := p.State.CurrentLevel()
	if level == nil || level.ListState == nil {
		return
	}
	view := commandPaletteFilteredView(level.Items, level.FilterState)
	if first, ok := firstSelectableIndex(level.Items, view.Indices); ok {
		level.ListState.SelectIndex(first)
		p.notifyCursorChange()
	}
}

func (p CommandPalette) moveCursorToEnd() {
	level := p.State.CurrentLevel()
	if level == nil || level.ListState == nil {
		return
	}
	view := commandPaletteFilteredView(level.Items, level.FilterState)
	if last, ok := lastSelectableIndex(level.Items, view.Indices); ok {
		level.ListState.SelectIndex(last)
		p.notifyCursorChange()
	}
}

func (p CommandPalette) selectCurrent() {
	if p.State == nil {
		return
	}
	item, ok := p.State.CurrentItem()
	if !ok {
		return
	}

	if p.OnSelect != nil {
		p.OnSelect(item)
		return
	}

	if item.Children != nil {
		title := item.ChildrenTitle
		if title == "" {
			title = item.Label
		}
		p.State.PushLevel(title, item.Children())
		p.notifyCursorChange()
		RequestFocus(p.inputID())
		return
	}

	if item.Action != nil {
		item.Action()
	}
}

func (p CommandPalette) handleEscape() {
	if p.State == nil {
		return
	}
	if p.State.PopLevel() {
		p.notifyCursorChange()
		return
	}
	p.dismiss()
}

func (p CommandPalette) handleBackspace(level *CommandPaletteLevel, onFilterChange func(string)) {
	if level == nil || level.InputState == nil {
		return
	}
	if level.InputState.GetText() == "" && p.State != nil && p.State.PopLevel() {
		p.notifyCursorChange()
		return
	}
	level.InputState.DeleteBackward()
	onFilterChange(level.InputState.GetText())
}

func (p CommandPalette) notifyCursorChange() {
	if p.OnCursorChange == nil || p.State == nil {
		return
	}
	if item, ok := p.State.CurrentItem(); ok {
		p.OnCursorChange(item)
	}
}

func (p CommandPalette) dismiss() {
	if p.State != nil {
		p.State.Close(false)
	}
	if p.OnDismiss != nil {
		p.OnDismiss()
	}
}

func (p CommandPalette) onBreadcrumbSelect() func(index int) {
	return func(index int) {
		if p.State == nil {
			return
		}
		for len(p.State.stack) > index+1 {
			p.State.PopLevel()
		}
		p.notifyCursorChange()
		RequestFocus(p.inputID())
	}
}

func (p CommandPalette) containerStyle(theme ThemeData) Style {
	style := p.Style
	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}
	return style
}

func (p CommandPalette) floatPosition() FloatPosition {
	if p.Position == FloatPositionAbsolute {
		if p.Offset != (Offset{}) {
			return FloatPositionAbsolute
		}
		return FloatPositionTopCenter
	}
	return p.Position
}

func (p CommandPalette) paletteWidth() Dimension {
	width := p.Style.GetDimensions().Width
	if width.IsUnset() {
		width = p.Width
	}
	if width.IsUnset() {
		width = Cells(defaultCommandPaletteWidth)
	}
	return width
}

func (p CommandPalette) paletteHeight() Dimension {
	height := p.Style.GetDimensions().Height
	if height.IsUnset() {
		height = p.Height
	}
	if height.IsUnset() {
		height = Cells(defaultCommandPaletteHeight)
	}
	return height
}

func (p CommandPalette) placeholderText() string {
	if p.Placeholder == "" {
		return defaultCommandPalettePlaceholder
	}
	return p.Placeholder
}

func (p CommandPalette) inputID() string {
	if p.ID == "" {
		return ""
	}
	return p.ID + "-input"
}

func (p CommandPalette) listID() string {
	if p.ID == "" {
		return ""
	}
	return p.ID + "-list"
}

func (p CommandPalette) scrollID() string {
	if p.ID == "" {
		return ""
	}
	return p.ID + "-scroll"
}

func indexOf(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}
