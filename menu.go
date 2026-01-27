package terma

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// MenuItem represents a menu item with optional submenu.
type MenuItem struct {
	Label    string     // Display text
	Shortcut string     // Optional hint (e.g., "Ctrl+S")
	Action   func()     // Callback when selected (ignored if Children is set)
	Children []MenuItem // Submenu items (opens on right arrow)
	Disabled bool       // Shown but not selectable
	Divider  string     // Divider title (empty renders a plain separator)
}

// IsSelectable returns true if the item can be focused/selected.
func (m MenuItem) IsSelectable() bool {
	return !m.Disabled && !m.IsDivider()
}

// IsDivider returns true if this item should render as a divider.
func (m MenuItem) IsDivider() bool {
	if m.Divider != "" {
		return true
	}
	return m.Label == "" && m.Shortcut == "" && m.Action == nil && len(m.Children) == 0
}

// MenuState holds the state for a Menu widget.
type MenuState struct {
	items        []MenuItem
	cursorIndex  Signal[int]
	openSubmenu  Signal[int] // Index of item with open submenu (-1 if none)
	submenuState *MenuState  // State for the open submenu (recursive)
}

// NewMenuState creates a new MenuState with the given items.
func NewMenuState(items []MenuItem) *MenuState {
	if items == nil {
		items = []MenuItem{}
	}
	state := &MenuState{
		items:       items,
		cursorIndex: NewSignal(0),
		openSubmenu: NewSignal(-1),
	}
	state.cursorIndex.Set(state.firstSelectableIndex())
	return state
}

// Items returns the current menu items.
func (s *MenuState) Items() []MenuItem {
	return s.items
}

// CursorIndex returns the current cursor position.
func (s *MenuState) CursorIndex() int {
	return s.cursorIndex.Peek()
}

// SetCursorIndex sets the cursor position (clamped to valid range).
func (s *MenuState) SetCursorIndex(i int) {
	if len(s.items) == 0 {
		s.cursorIndex.Set(0)
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(s.items) {
		i = len(s.items) - 1
	}
	s.cursorIndex.Set(i)
}

// SelectedItem returns the currently selected item (if any).
func (s *MenuState) SelectedItem() (MenuItem, bool) {
	if len(s.items) == 0 {
		return MenuItem{}, false
	}
	idx := s.cursorIndex.Peek()
	if idx < 0 || idx >= len(s.items) {
		return MenuItem{}, false
	}
	item := s.items[idx]
	if !item.IsSelectable() {
		return MenuItem{}, false
	}
	return item, true
}

// OpenSubmenu opens the submenu for the given item index.
func (s *MenuState) OpenSubmenu(index int) {
	if index < 0 || index >= len(s.items) {
		return
	}
	item := s.items[index]
	if !item.IsSelectable() || len(item.Children) == 0 {
		return
	}
	s.openSubmenu.Set(index)
	s.submenuState = NewMenuState(item.Children)
}

// CloseSubmenu closes any open submenu.
func (s *MenuState) CloseSubmenu() {
	s.openSubmenu.Set(-1)
	s.submenuState = nil
}

// HasOpenSubmenu returns true if there is an open submenu.
func (s *MenuState) HasOpenSubmenu() bool {
	return s.openSubmenu.Peek() >= 0 && s.submenuState != nil
}

func (s *MenuState) firstSelectableIndex() int {
	for i, item := range s.items {
		if item.IsSelectable() {
			return i
		}
	}
	return 0
}

// Menu is a convenience widget for dropdown/context menus.
// It composes Floating + list rendering internally.
type Menu struct {
	ID    string     // Optional unique identifier
	State *MenuState // Required

	// Positioning (choose one approach)
	AnchorID string      // Anchor to widget (dropdown style)
	Anchor   AnchorPoint // Where to anchor (default: AnchorBottomLeft)
	Position FloatPosition
	Offset   Offset

	// Callbacks
	OnSelect  func(item MenuItem) // Called when item selected
	OnDismiss func()              // Called when menu should close

	// Styling
	Width Dimension // Deprecated: use Style.Width
	Style Style // Container style
}

// WidgetID returns the menu's unique identifier.
func (m Menu) WidgetID() string {
	return m.ID
}

// IsFocusable returns true to allow keyboard navigation.
func (m Menu) IsFocusable() bool {
	return m.State != nil
}

// OnKey handles keys not covered by declarative keybindings.
func (m Menu) OnKey(event KeyEvent) bool {
	return false
}

// Keybinds returns the declarative keybindings for this menu.
func (m Menu) Keybinds() []Keybind {
	if m.State == nil {
		return nil
	}
	return []Keybind{
		{Key: "up", Action: m.movePrevious, Hidden: true},
		{Key: "k", Action: m.movePrevious, Hidden: true},
		{Key: "down", Action: m.moveNext, Hidden: true},
		{Key: "j", Action: m.moveNext, Hidden: true},
		{Key: "home", Action: m.moveFirst, Hidden: true},
		{Key: "end", Action: m.moveLast, Hidden: true},
		{Key: "enter", Action: m.selectCurrent, Hidden: true},
		{Key: " ", Action: m.selectCurrent, Hidden: true},
		{Key: "right", Action: m.openSubmenu, Hidden: true},
		{Key: "l", Action: m.openSubmenu, Hidden: true},
		{Key: "left", Action: m.closeSubmenu, Hidden: true},
		{Key: "h", Action: m.closeSubmenu, Hidden: true},
		{Key: "escape", Action: m.dismiss, Hidden: true},
	}
}

// Build renders the menu as a floating overlay.
func (m Menu) Build(ctx BuildContext) Widget {
	if m.State == nil {
		return EmptyWidget{}
	}

	float := Floating{
		Visible: true,
		Config: FloatConfig{
			AnchorID:  m.AnchorID,
			Anchor:    m.anchor(),
			Position:  m.Position,
			Offset:    m.Offset,
			OnDismiss: m.dismissCallback(),
		},
		Child: m.buildMenuContent(ctx),
	}
	return float.Build(ctx)
}

func (m Menu) anchor() AnchorPoint {
	if m.AnchorID == "" {
		return m.Anchor
	}
	if m.Anchor == AnchorTopLeft {
		return AnchorBottomLeft
	}
	return m.Anchor
}

func (m Menu) dismissCallback() func() {
	if m.OnDismiss == nil {
		return nil
	}
	return func() {
		if m.State != nil {
			m.State.CloseSubmenu()
		}
		m.OnDismiss()
	}
}

func (m Menu) buildMenuContent(ctx BuildContext) Widget {
	items := m.State.Items()
	cursorIdx := m.State.cursorIndex.Get()
	openSubmenuIdx := m.State.openSubmenu.Get()
	baseStyle := m.Style
	if baseStyle.Width.IsUnset() {
		baseStyle.Width = m.Width
	}
	width := baseStyle.Width
	layout := computeMenuItemLayout(items, width)
	itemWidth := layout.widthDimension(width)

	children := make([]Widget, 0, len(items))
	for i, item := range items {
		children = append(children, m.renderItem(ctx, item, i, i == cursorIdx, layout, itemWidth))

		if i == openSubmenuIdx && len(item.Children) > 0 && m.State.submenuState != nil {
			children = append(children, Menu{
				ID:        m.submenuID(),
				State:     m.State.submenuState,
				AnchorID:  fmt.Sprintf("%s-item-%d", m.ID, i),
				Anchor:    AnchorRightTop,
				OnSelect:  m.submenuSelectCallback(),
				OnDismiss: m.submenuDismissCallback(),
				Style:     baseStyle,
			})
		}
	}

	menuStyle := m.menuStyle(ctx)
	menuStyle.Width = itemWidth
	return Column{
		ID:         m.ID + "-content",
		CrossAlign: CrossAxisStretch,
		Style:      menuStyle,
		Children:   children,
	}
}

func (m Menu) submenuID() string {
	if m.ID == "" {
		return ""
	}
	return m.ID + "-sub"
}

func (m Menu) submenuSelectCallback() func(item MenuItem) {
	return func(item MenuItem) {
		if m.State != nil {
			m.State.CloseSubmenu()
		}
		RequestFocus(m.ID)
		if m.OnSelect != nil {
			m.OnSelect(item)
			return
		}
		if item.Action != nil {
			item.Action()
		}
	}
}

func (m Menu) submenuDismissCallback() func() {
	return func() {
		if m.State != nil {
			m.State.CloseSubmenu()
		}
		RequestFocus(m.ID)
	}
}

func (m Menu) menuStyle(ctx BuildContext) Style {
	style := m.Style
	theme := ctx.Theme()

	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}

	return style
}

func (m Menu) renderItem(ctx BuildContext, item MenuItem, index int, active bool, layout menuItemLayout, itemWidth Dimension) Widget {
	theme := ctx.Theme()

	if item.IsDivider() {
		prefix, line := layout.dividerParts(item.Divider)
		lineColor := theme.TextMuted.Blend(theme.Surface, 0.7)
		rowStyle := Style{
			Padding: EdgeInsetsXY(layout.paddingX, 0),
			Width:   itemWidth,
		}
		return Row{
			Style: rowStyle,
			Children: []Widget{
				Text{
					Content: prefix,
					Style:   Style{ForegroundColor: theme.TextMuted},
				},
				Text{
					Content: line,
					Style:   Style{ForegroundColor: lineColor},
				},
			},
		}
	}

	itemStyle := Style{
		Padding: EdgeInsetsXY(layout.paddingX, 0),
		Width:   itemWidth,
	}
	labelStyle := Style{ForegroundColor: theme.Text}
	suffixStyle := Style{ForegroundColor: theme.TextMuted}
	if item.Disabled {
		labelStyle.ForegroundColor = theme.TextMuted
		suffixStyle.ForegroundColor = theme.TextMuted
	} else if active && ctx.IsFocused(m) {
		itemStyle.BackgroundColor = theme.ActiveCursor
		labelStyle.ForegroundColor = theme.SelectionText
		suffixStyle.ForegroundColor = theme.ActiveCursor.AutoText()
	}

	return Row{
		ID:         fmt.Sprintf("%s-item-%d", m.ID, index),
		Spacing:    layout.spacing,
		Style:      itemStyle,
		CrossAlign: CrossAxisCenter,
		Children: []Widget{
			Text{
				Content: item.Label,
				Style:   labelStyle,
			},
			Spacer{Width: Flex(1)},
			Text{
				Content: menuItemSuffix(item),
				Style:   suffixStyle,
			},
		},
	}
}

func (m Menu) moveNext() {
	if m.State == nil {
		return
	}
	items := m.State.Items()
	if len(items) == 0 {
		return
	}
	current := m.State.cursorIndex.Peek()
	if current < 0 || current >= len(items) {
		current = 0
	}
	for i := 1; i <= len(items); i++ {
		next := (current + i) % len(items)
		if items[next].IsSelectable() {
			m.State.SetCursorIndex(next)
			m.closeOpenSubmenuIfNeeded(next)
			return
		}
	}
}

func (m Menu) movePrevious() {
	if m.State == nil {
		return
	}
	items := m.State.Items()
	if len(items) == 0 {
		return
	}
	current := m.State.cursorIndex.Peek()
	if current < 0 || current >= len(items) {
		current = 0
	}
	for i := 1; i <= len(items); i++ {
		prev := (current - i + len(items)) % len(items)
		if items[prev].IsSelectable() {
			m.State.SetCursorIndex(prev)
			m.closeOpenSubmenuIfNeeded(prev)
			return
		}
	}
}

func (m Menu) moveFirst() {
	if m.State == nil {
		return
	}
	items := m.State.Items()
	for i, item := range items {
		if item.IsSelectable() {
			m.State.SetCursorIndex(i)
			m.closeOpenSubmenuIfNeeded(i)
			return
		}
	}
}

func (m Menu) moveLast() {
	if m.State == nil {
		return
	}
	items := m.State.Items()
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].IsSelectable() {
			m.State.SetCursorIndex(i)
			m.closeOpenSubmenuIfNeeded(i)
			return
		}
	}
}

func (m Menu) closeOpenSubmenuIfNeeded(cursorIdx int) {
	if m.State == nil || !m.State.HasOpenSubmenu() {
		return
	}
	openIdx := m.State.openSubmenu.Peek()
	if openIdx != cursorIdx {
		m.State.CloseSubmenu()
	}
}

func (m Menu) openSubmenu() {
	if m.State == nil {
		return
	}
	item, ok := m.State.SelectedItem()
	if !ok || len(item.Children) == 0 {
		return
	}
	m.State.OpenSubmenu(m.State.cursorIndex.Peek())
	if m.ID != "" {
		RequestFocus(m.submenuID())
	}
}

func (m Menu) closeSubmenu() {
	if m.State == nil {
		return
	}
	if m.State.HasOpenSubmenu() {
		m.State.CloseSubmenu()
		RequestFocus(m.ID)
		return
	}
	m.dismiss()
}

func (m Menu) selectCurrent() {
	if m.State == nil {
		return
	}
	item, ok := m.State.SelectedItem()
	if !ok {
		return
	}

	if len(item.Children) > 0 {
		m.openSubmenu()
		return
	}

	m.State.CloseSubmenu()
	if m.OnSelect != nil {
		m.OnSelect(item)
		return
	}
	if item.Action != nil {
		item.Action()
	}
}

func (m Menu) dismiss() {
	if m.State != nil {
		m.State.CloseSubmenu()
	}
	if m.OnDismiss != nil {
		m.OnDismiss()
	}
}

type menuItemLayout struct {
	labelWidth   int
	suffixWidth  int
	contentWidth int
	rowWidth     int
	paddingX     int
	spacing      int
}

const menuItemPaddingX = 1

func computeMenuItemLayout(items []MenuItem, width Dimension) menuItemLayout {
	layout := menuItemLayout{paddingX: menuItemPaddingX}
	maxDividerPrefixWidth := 0
	for _, item := range items {
		if item.IsDivider() {
			titleWidth := ansi.StringWidth(item.Divider)
			prefixWidth := titleWidth
			if titleWidth > 0 {
				prefixWidth++
			}
			if prefixWidth > maxDividerPrefixWidth {
				maxDividerPrefixWidth = prefixWidth
			}
			continue
		}
		labelWidth := ansi.StringWidth(item.Label)
		if labelWidth > layout.labelWidth {
			layout.labelWidth = labelWidth
		}
		suffixWidth := ansi.StringWidth(menuItemSuffix(item))
		if suffixWidth > layout.suffixWidth {
			layout.suffixWidth = suffixWidth
		}
	}

	if layout.suffixWidth > 0 {
		layout.spacing = 1
	}

	layout.contentWidth = layout.labelWidth
	if layout.suffixWidth > 0 {
		layout.contentWidth += layout.suffixWidth + layout.spacing*2
	}
	if maxDividerPrefixWidth > layout.contentWidth {
		layout.contentWidth = maxDividerPrefixWidth
	}
	if layout.contentWidth < 1 {
		layout.contentWidth = 1
	}

	layout.rowWidth = layout.contentWidth + layout.paddingX*2

	if width.IsCells() {
		layout.rowWidth = max(width.CellsValue(), layout.paddingX*2)
		layout.contentWidth = max(0, layout.rowWidth-layout.paddingX*2)
	}

	return layout
}

func (l menuItemLayout) widthDimension(width Dimension) Dimension {
	if width.IsUnset() || width.IsAuto() {
		return Cells(l.rowWidth)
	}
	return width
}

func (l menuItemLayout) dividerParts(title string) (prefix, line string) {
	lineChar := "─"
	if title == "" {
		return "", strings.Repeat(lineChar, l.contentWidth)
	}
	prefix = title + " "
	prefixWidth := ansi.StringWidth(prefix)
	if prefixWidth >= l.contentWidth {
		return prefix, ""
	}
	return prefix, strings.Repeat(lineChar, max(0, l.contentWidth-prefixWidth))
}

func menuItemSuffix(item MenuItem) string {
	if len(item.Children) > 0 {
		return "▸"
	}
	if item.Shortcut != "" {
		return item.Shortcut
	}
	return ""
}
