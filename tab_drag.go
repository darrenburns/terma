package terma

import (
	"fmt"
	"slices"
)

// MoveTab moves a tab to a zero-based index without changing the active tab.
// It returns false for an unknown key, invalid index, or unchanged position.
func (s *TabState) MoveTab(key string, index int) bool {
	tabs := s.tabs.Peek()
	from := slices.IndexFunc(tabs, func(tab Tab) bool { return tab.Key == key })
	if from < 0 || index < 0 || index >= len(tabs) || from == index {
		return false
	}
	result := append([]Tab(nil), tabs...)
	item := result[from]
	if from < index {
		copy(result[from:index], result[from+1:index+1])
	} else {
		copy(result[index+1:from+1], result[index:from])
	}
	result[index] = item
	s.tabs.Set(result)
	return true
}

type tabDragItem struct {
	key   string
	width float64
}

type tabDragPreview struct {
	barID    string
	key      string
	revision uint64
	items    []tabDragItem
	index    int
	left     float64
	lastX    float64
	started  bool
}

type tabDragBehavior struct {
	state      *TabState
	barID, key string
}

func (b *tabDragBehavior) arm(drag *dragSession) bool {
	tabs, revision := b.state.tabs.peekWithRevision()
	p := &tabDragPreview{barID: b.barID, key: b.key, revision: revision, lastX: drag.x, index: -1}
	parent := drag.source.parent
	if parent == nil {
		return false
	}
	nodes := make(map[string]*widgetNode, len(parent.children))
	for i, child := range parent.children {
		nodes[child.eventID] = child
		if child == drag.source {
			p.left = float64(drag.slotBounds.X - parent.layout.Children[i].X)
		}
	}
	for i, tab := range tabs {
		node := nodes[tabHeaderID(b.barID, tab.Key)]
		if node == nil {
			return false
		}
		p.items = append(p.items, tabDragItem{key: tab.Key, width: float64(node.layout.Box.MarginBoxWidth())})
		drag.frozen[node] = copyDragLayout(node.layout)
		if tab.Key == b.key {
			p.index = i
		}
	}
	if p.index < 0 {
		return false
	}
	b.state.dragPreview = p
	return true
}

func (b *tabDragBehavior) valid() bool {
	p := b.state.dragPreview
	if p == nil {
		return false
	}
	_, revision := b.state.tabs.peekWithRevision()
	return revision == p.revision && b.state.editingKey.Peek() == ""
}

func (b *tabDragBehavior) move(drag *dragSession) {
	p := b.state.dragPreview
	if p == nil {
		return
	}
	left := drag.x - drag.grabX
	before := p.index
	p.advance(left, drag.x-p.lastX)
	p.lastX = drag.x
	if before != p.index || !p.started {
		p.started = true
		b.state.dragVersion.Update(func(n uint64) uint64 { return n + 1 })
	}
}

func (p *tabDragPreview) advance(left, delta float64) {
	if delta == 0 {
		return
	}
	width := p.items[p.index].width
	for {
		x := p.left
		for _, item := range p.items[:p.index] {
			x += item.width
		}
		switch {
		case delta > 0 && p.index+1 < len(p.items):
			next := p.items[p.index+1]
			if left+width <= x+width+next.width/2 {
				return
			}
			p.items[p.index], p.items[p.index+1] = p.items[p.index+1], p.items[p.index]
			p.index++
		case delta < 0 && p.index > 0:
			previous := p.items[p.index-1]
			if left >= x-previous.width/2 {
				return
			}
			p.items[p.index], p.items[p.index-1] = p.items[p.index-1], p.items[p.index]
			p.index--
		default:
			return
		}
	}
}

func (b *tabDragBehavior) finish(commit bool) {
	p := b.state.dragPreview
	if p == nil {
		return
	}
	valid := b.valid()
	b.state.dragPreview = nil
	b.state.dragVersion.Update(func(n uint64) uint64 { return n + 1 })
	if commit && valid {
		b.state.MoveTab(p.key, p.index)
	}
}

func tabHeaderID(barID, key string) string { return tabPartID(barID, key, "header") }

func tabPartID(barID, key, part string) string {
	return fmt.Sprintf("__terma_tab_%s:%d:%s:%s", part, len(barID), barID, key)
}

func (s *TabState) displayedTabs(barID string, tabs []Tab) []Tab {
	s.dragVersion.Get()
	p := s.dragPreview
	if p == nil || p.barID != barID || !p.started {
		return tabs
	}
	byKey := make(map[string]Tab, len(tabs))
	for _, tab := range tabs {
		byKey[tab.Key] = tab
	}
	result := make([]Tab, 0, len(tabs))
	for _, item := range p.items {
		if tab, ok := byKey[item.key]; ok {
			result = append(result, tab)
		}
	}
	return result
}
