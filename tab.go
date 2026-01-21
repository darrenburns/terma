package terma

import "fmt"

// Tab represents a single tab with its key, label, and optional content.
type Tab struct {
	Key     string // Unique identifier (used for switching)
	Label   string // Display text (can differ from Key)
	Content Widget // Optional - used by TabView, ignored by TabBar
}

// TabState holds the state for a TabBar or TabView widget.
// It is the source of truth for tabs and the active tab, and must be provided to TabBar/TabView.
type TabState struct {
	tabs       AnySignal[[]Tab]
	activeKey  Signal[string]
	editingKey Signal[string] // For rename support
}

// NewTabState creates a new TabState with the given tabs.
// The first tab becomes active by default.
func NewTabState(tabs []Tab) *TabState {
	if tabs == nil {
		tabs = []Tab{}
	}
	activeKey := ""
	if len(tabs) > 0 {
		activeKey = tabs[0].Key
	}
	return &TabState{
		tabs:       NewAnySignal(tabs),
		activeKey:  NewSignal(activeKey),
		editingKey: NewSignal(""),
	}
}

// NewTabStateWithActive creates a new TabState with the given tabs and active tab key.
func NewTabStateWithActive(tabs []Tab, activeKey string) *TabState {
	if tabs == nil {
		tabs = []Tab{}
	}
	return &TabState{
		tabs:       NewAnySignal(tabs),
		activeKey:  NewSignal(activeKey),
		editingKey: NewSignal(""),
	}
}

// Tabs returns the current list of tabs (subscribes to changes).
func (s *TabState) Tabs() []Tab {
	return s.tabs.Get()
}

// TabsPeek returns the current list of tabs without subscribing.
func (s *TabState) TabsPeek() []Tab {
	return s.tabs.Peek()
}

// ActiveKey returns the key of the currently active tab (subscribes to changes).
func (s *TabState) ActiveKey() string {
	return s.activeKey.Get()
}

// ActiveKeyPeek returns the key of the currently active tab without subscribing.
func (s *TabState) ActiveKeyPeek() string {
	return s.activeKey.Peek()
}

// SetActiveKey sets the active tab by key.
func (s *TabState) SetActiveKey(key string) {
	s.activeKey.Set(key)
}

// ActiveIndex returns the index of the currently active tab, or -1 if not found.
func (s *TabState) ActiveIndex() int {
	tabs := s.tabs.Peek()
	activeKey := s.activeKey.Peek()
	for i, tab := range tabs {
		if tab.Key == activeKey {
			return i
		}
	}
	return -1
}

// ActiveTab returns the currently active tab, or nil if not found.
func (s *TabState) ActiveTab() *Tab {
	tabs := s.tabs.Peek()
	activeKey := s.activeKey.Peek()
	for i := range tabs {
		if tabs[i].Key == activeKey {
			return &tabs[i]
		}
	}
	return nil
}

// SelectNext moves to the next tab, wrapping to the first if at the end.
func (s *TabState) SelectNext() {
	tabs := s.tabs.Peek()
	if len(tabs) == 0 {
		return
	}
	idx := s.ActiveIndex()
	if idx == -1 {
		idx = 0
	} else {
		idx = (idx + 1) % len(tabs)
	}
	s.activeKey.Set(tabs[idx].Key)
}

// SelectPrevious moves to the previous tab, wrapping to the last if at the beginning.
func (s *TabState) SelectPrevious() {
	tabs := s.tabs.Peek()
	if len(tabs) == 0 {
		return
	}
	idx := s.ActiveIndex()
	if idx == -1 {
		idx = 0
	} else {
		idx = (idx - 1 + len(tabs)) % len(tabs)
	}
	s.activeKey.Set(tabs[idx].Key)
}

// SelectIndex sets the active tab by index.
func (s *TabState) SelectIndex(index int) {
	tabs := s.tabs.Peek()
	if index < 0 || index >= len(tabs) {
		return
	}
	s.activeKey.Set(tabs[index].Key)
}

// AddTab adds a tab to the end of the list.
func (s *TabState) AddTab(tab Tab) {
	s.tabs.Update(func(tabs []Tab) []Tab {
		return append(tabs, tab)
	})
	// If this is the first tab, make it active
	if len(s.tabs.Peek()) == 1 {
		s.activeKey.Set(tab.Key)
	}
}

// InsertTab inserts a tab at the specified index.
func (s *TabState) InsertTab(index int, tab Tab) {
	s.tabs.Update(func(tabs []Tab) []Tab {
		if index < 0 {
			index = 0
		}
		if index > len(tabs) {
			index = len(tabs)
		}
		result := make([]Tab, 0, len(tabs)+1)
		result = append(result, tabs[:index]...)
		result = append(result, tab)
		result = append(result, tabs[index:]...)
		return result
	})
	// If this is the first tab, make it active
	if len(s.tabs.Peek()) == 1 {
		s.activeKey.Set(tab.Key)
	}
}

// RemoveTab removes a tab by key. Returns true if the tab was found and removed.
func (s *TabState) RemoveTab(key string) bool {
	tabs := s.tabs.Peek()
	idx := -1
	for i, tab := range tabs {
		if tab.Key == key {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}

	// If removing the active tab, switch to adjacent tab
	if s.activeKey.Peek() == key {
		if len(tabs) > 1 {
			if idx == len(tabs)-1 {
				// Last tab: switch to previous
				s.activeKey.Set(tabs[idx-1].Key)
			} else {
				// Switch to next
				s.activeKey.Set(tabs[idx+1].Key)
			}
		} else {
			// No more tabs
			s.activeKey.Set("")
		}
	}

	s.tabs.Update(func(tabs []Tab) []Tab {
		return append(tabs[:idx], tabs[idx+1:]...)
	})
	return true
}

// MoveTabLeft moves the tab with the given key one position to the left.
// Returns true if the move was successful.
func (s *TabState) MoveTabLeft(key string) bool {
	tabs := s.tabs.Peek()
	idx := -1
	for i, tab := range tabs {
		if tab.Key == key {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return false
	}

	s.tabs.Update(func(tabs []Tab) []Tab {
		tabs[idx], tabs[idx-1] = tabs[idx-1], tabs[idx]
		return tabs
	})
	return true
}

// MoveTabRight moves the tab with the given key one position to the right.
// Returns true if the move was successful.
func (s *TabState) MoveTabRight(key string) bool {
	tabs := s.tabs.Peek()
	idx := -1
	for i, tab := range tabs {
		if tab.Key == key {
			idx = i
			break
		}
	}
	if idx == -1 || idx >= len(tabs)-1 {
		return false
	}

	s.tabs.Update(func(tabs []Tab) []Tab {
		tabs[idx], tabs[idx+1] = tabs[idx+1], tabs[idx]
		return tabs
	})
	return true
}

// SetLabel updates the label of a tab by key.
func (s *TabState) SetLabel(key, label string) {
	s.tabs.Update(func(tabs []Tab) []Tab {
		for i := range tabs {
			if tabs[i].Key == key {
				tabs[i].Label = label
				break
			}
		}
		return tabs
	})
}

// StartEditing begins editing mode for a tab's label.
func (s *TabState) StartEditing(key string) {
	s.editingKey.Set(key)
}

// StopEditing exits editing mode.
func (s *TabState) StopEditing() {
	s.editingKey.Set("")
}

// IsEditing returns true if the given tab is being edited.
func (s *TabState) IsEditing(key string) bool {
	return s.editingKey.Peek() == key && key != ""
}

// EditingKey returns the key of the tab being edited (subscribes to changes).
func (s *TabState) EditingKey() string {
	return s.editingKey.Get()
}

// TabCount returns the number of tabs.
func (s *TabState) TabCount() int {
	return len(s.tabs.Peek())
}

// TabKeybindPattern specifies the style of position-based keybindings.
type TabKeybindPattern int

const (
	// TabKeybindNone disables position-based keybindings.
	TabKeybindNone TabKeybindPattern = iota
	// TabKeybindNumbers uses 1, 2, 3... 9 for tab switching.
	TabKeybindNumbers
	// TabKeybindAltNumbers uses alt+1, alt+2... alt+9 for tab switching.
	TabKeybindAltNumbers
	// TabKeybindCtrlNumbers uses ctrl+1, ctrl+2... ctrl+9 for tab switching.
	TabKeybindCtrlNumbers
)

// TabBar is a focusable widget that renders a horizontal row of tabs.
// It supports keyboard navigation and position-based keybindings.
type TabBar struct {
	ID             string            // Required for focus
	State          *TabState         // Required - holds tabs and active key
	KeybindPattern TabKeybindPattern // Position keybind style
	OnTabChange    func(key string)  // Tab selection callback
	OnTabClose     func(key string)  // Close button callback
	Closable       bool              // Show close buttons
	AllowReorder   bool              // Enable ctrl+left/right reordering
	Width          Dimension         // Optional width
	Height         Dimension         // Optional height
	Style          Style             // Container style
	TabStyle       Style             // Inactive tab style
	ActiveTabStyle Style             // Active tab style
	Click     func(MouseEvent) // Optional callback invoked when clicked
	MouseDown func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent) // Optional callback invoked when mouse is released
	Hover     func(bool)       // Optional callback invoked when hover state changes
	MinMaxDimensions
}

// WidgetID returns the widget's unique identifier.
func (t TabBar) WidgetID() string {
	return t.ID
}

// GetDimensions returns the width and height dimension preferences.
func (t TabBar) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the style of the tab bar.
func (t TabBar) GetStyle() Style {
	return t.Style
}

// IsFocusable returns true to allow keyboard navigation.
func (t TabBar) IsFocusable() bool {
	return true
}

// OnClick is called when the widget is clicked.
func (t TabBar) OnClick(event MouseEvent) {
	if t.Click != nil {
		t.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed.
func (t TabBar) OnMouseDown(event MouseEvent) {
	if t.MouseDown != nil {
		t.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released.
func (t TabBar) OnMouseUp(event MouseEvent) {
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
func (t TabBar) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// Keybinds returns the declarative keybindings for the tab bar.
func (t TabBar) Keybinds() []Keybind {
	if t.State == nil {
		return nil
	}

	keybinds := []Keybind{
		{Key: "left", Name: "Prev Tab", Action: t.selectPrevious, Hidden: true},
		{Key: "h", Name: "Prev Tab", Action: t.selectPrevious},
		{Key: "right", Name: "Next Tab", Action: t.selectNext, Hidden: true},
		{Key: "l", Name: "Next Tab", Action: t.selectNext},
	}

	// Reorder keybinds (if enabled)
	if t.AllowReorder {
		keybinds = append(keybinds,
			Keybind{Key: "ctrl+h", Name: "Move Left", Action: t.moveActiveLeft},
			Keybind{Key: "ctrl+l", Name: "Move Right", Action: t.moveActiveRight},
		)
	}

	// Closable keybind
	if t.Closable {
		keybinds = append(keybinds,
			Keybind{Key: "ctrl+w", Name: "Close Tab", Action: t.closeActive, Hidden: true},
		)
	}

	// Position-based keybinds (Hidden to avoid KeybindBar clutter)
	if t.KeybindPattern != TabKeybindNone {
		tabs := t.State.TabsPeek()
		maxTabs := 9
		if len(tabs) < maxTabs {
			maxTabs = len(tabs)
		}

		for i := 0; i < maxTabs; i++ {
			tab := tabs[i]
			key := t.formatPositionKey(i + 1)
			keybinds = append(keybinds, Keybind{
				Key:    key,
				Name:   tab.Label,
				Action: t.makeSelectAction(tab.Key),
				Hidden: true,
			})
		}
	}

	return keybinds
}

// formatPositionKey returns the key string for a position-based keybind.
func (t TabBar) formatPositionKey(num int) string {
	switch t.KeybindPattern {
	case TabKeybindNumbers:
		return fmt.Sprintf("%d", num)
	case TabKeybindAltNumbers:
		return fmt.Sprintf("alt+%d", num)
	case TabKeybindCtrlNumbers:
		return fmt.Sprintf("ctrl+%d", num)
	default:
		return ""
	}
}

// makeSelectAction returns an action function that selects a specific tab.
func (t TabBar) makeSelectAction(key string) func() {
	return func() {
		t.State.SetActiveKey(key)
		if t.OnTabChange != nil {
			t.OnTabChange(key)
		}
	}
}

// selectPrevious moves to the previous tab.
func (t TabBar) selectPrevious() {
	if t.State == nil {
		return
	}
	oldKey := t.State.ActiveKeyPeek()
	t.State.SelectPrevious()
	newKey := t.State.ActiveKeyPeek()
	if oldKey != newKey && t.OnTabChange != nil {
		t.OnTabChange(newKey)
	}
}

// selectNext moves to the next tab.
func (t TabBar) selectNext() {
	if t.State == nil {
		return
	}
	oldKey := t.State.ActiveKeyPeek()
	t.State.SelectNext()
	newKey := t.State.ActiveKeyPeek()
	if oldKey != newKey && t.OnTabChange != nil {
		t.OnTabChange(newKey)
	}
}

// moveActiveLeft moves the active tab left.
func (t TabBar) moveActiveLeft() {
	if t.State == nil {
		return
	}
	t.State.MoveTabLeft(t.State.ActiveKeyPeek())
}

// moveActiveRight moves the active tab right.
func (t TabBar) moveActiveRight() {
	if t.State == nil {
		return
	}
	t.State.MoveTabRight(t.State.ActiveKeyPeek())
}

// closeActive closes the active tab.
func (t TabBar) closeActive() {
	if t.State == nil {
		return
	}
	key := t.State.ActiveKeyPeek()
	if t.OnTabClose != nil {
		t.OnTabClose(key)
	} else {
		t.State.RemoveTab(key)
	}
}

// OnKey handles keys not covered by declarative keybindings.
func (t TabBar) OnKey(event KeyEvent) bool {
	return false
}

// Build renders the tab bar as a Row of styled text widgets.
func (t TabBar) Build(ctx BuildContext) Widget {
	if t.State == nil {
		return Row{}
	}

	tabs := t.State.Tabs()
	activeKey := t.State.ActiveKey()
	theme := ctx.Theme()

	children := make([]Widget, 0, len(tabs))

	for _, tab := range tabs {
		isActive := tab.Key == activeKey
		tabKey := tab.Key

		// Determine style
		var style Style
		if isActive {
			style = t.ActiveTabStyle
			// Apply theme defaults if no explicit colors set
			if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
				style.ForegroundColor = theme.Background
			}
			if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
				style.BackgroundColor = theme.Accent
			}
		} else {
			style = t.TabStyle
			if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
				style.ForegroundColor = theme.TextMuted
			}
			if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
				style.BackgroundColor = theme.Surface
			}
		}

		// Add default padding if not set
		if style.Padding.Top == 0 && style.Padding.Bottom == 0 &&
			style.Padding.Left == 0 && style.Padding.Right == 0 {
			style.Padding = EdgeInsetsXY(2, 0)
		}

		// Build tab content
		if t.Closable {
			// Tab with separate close button
			labelStyle := style
			labelStyle.Padding = EdgeInsets{Left: style.Padding.Left, Right: 1}

			closeStyle := style
			closeStyle.Padding = EdgeInsets{Right: style.Padding.Right}

			children = append(children, Row{
				Style: Style{BackgroundColor: style.BackgroundColor},
				Children: []Widget{
					Text{
						Content: tab.Label,
						Style:   labelStyle,
						Click: func(MouseEvent) {
							t.State.SetActiveKey(tabKey)
							if t.OnTabChange != nil {
								t.OnTabChange(tabKey)
							}
						},
					},
					Text{
						Content: "×",
						Style:   closeStyle,
						Click: func(MouseEvent) {
							if t.OnTabClose != nil {
								t.OnTabClose(tabKey)
							} else {
								t.State.RemoveTab(tabKey)
							}
						},
					},
				},
			})
		} else {
			// Tab without close button
			children = append(children, Text{
				Content: tab.Label,
				Style:   style,
				Click: func(MouseEvent) {
					t.State.SetActiveKey(tabKey)
					if t.OnTabChange != nil {
						t.OnTabChange(tabKey)
					}
				},
			})
		}
	}

	return Row{
		ID:        t.ID,
		Width:     t.Width,
		Height:    t.Height,
		Style:     t.Style,
		Children:  children,
		Click:     t.Click,
		MouseDown: t.MouseDown,
		MouseUp:   t.MouseUp,
		Hover:     t.Hover,
	}
}

// TabView is a convenience composite that combines TabBar with a content area.
// It renders tabs at the top and the active tab's content below.
type TabView struct {
	ID             string            // Optional identifier
	State          *TabState         // Required - holds tabs and active key
	KeybindPattern TabKeybindPattern // Position keybind style
	OnTabChange    func(key string)  // Tab selection callback
	OnTabClose     func(key string)  // Close button callback
	Closable       bool              // Show close buttons
	AllowReorder   bool              // Enable ctrl+left/right reordering
	Width          Dimension         // Optional width
	Height         Dimension         // Optional height
	Style          Style             // Container style
	TabBarStyle    Style             // Style for the tab bar row
	TabStyle       Style             // Inactive tab style
	ActiveTabStyle Style             // Active tab style
	ContentStyle   Style             // Style for the content area
	MinMaxDimensions
}

// WidgetID returns the widget's unique identifier.
func (t TabView) WidgetID() string {
	return t.ID
}

// GetDimensions returns the width and height dimension preferences.
func (t TabView) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the style of the tab view.
func (t TabView) GetStyle() Style {
	return t.Style
}

// Build renders the tab view as a Column with TabBar and content area.
func (t TabView) Build(ctx BuildContext) Widget {
	if t.State == nil {
		return Column{}
	}

	// Build tab bar
	tabBar := TabBar{
		ID:             t.ID + "-tabbar",
		State:          t.State,
		KeybindPattern: t.KeybindPattern,
		OnTabChange:    t.OnTabChange,
		OnTabClose:     t.OnTabClose,
		Closable:       t.Closable,
		AllowReorder:   t.AllowReorder,
		Style:          t.TabBarStyle,
		TabStyle:       t.TabStyle,
		ActiveTabStyle: t.ActiveTabStyle,
	}

	// Build content area using Switcher
	tabs := t.State.TabsPeek()
	contentMap := make(map[string]Widget, len(tabs))
	for _, tab := range tabs {
		if tab.Content != nil {
			contentMap[tab.Key] = tab.Content
		} else {
			contentMap[tab.Key] = EmptyWidget{}
		}
	}

	content := Switcher{
		Active:   t.State.ActiveKey(),
		Children: contentMap,
		Height:   Flex(1),
		Style:    t.ContentStyle,
	}

	return Column{
		ID:     t.ID,
		Width:  t.Width,
		Height: t.Height,
		Style:  t.Style,
		Children: []Widget{
			tabBar,
			content,
		},
	}
}
