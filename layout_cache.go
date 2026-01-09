package terma

import "terma/layout"

// ChildLayout holds precomputed layout for a child widget.
type ChildLayout struct {
	// BorderBox is the widget bounds (content + padding + border, excludes margin).
	// Position is relative to parent's content area.
	BorderBox Rect
	// Margin around the border box.
	Margin EdgeInsets
}

// MarginBoxWidth returns the total width including margin (for sibling positioning).
func (cl ChildLayout) MarginBoxWidth() int {
	return cl.BorderBox.Width + cl.Margin.Horizontal()
}

// MarginBoxHeight returns the total height including margin (for sibling positioning).
func (cl ChildLayout) MarginBoxHeight() int {
	return cl.BorderBox.Height + cl.Margin.Vertical()
}

// MarginBoxX returns the X position of the margin box (for RenderChild).
func (cl ChildLayout) MarginBoxX() int {
	return cl.BorderBox.X - cl.Margin.Left
}

// MarginBoxY returns the Y position of the margin box (for RenderChild).
func (cl ChildLayout) MarginBoxY() int {
	return cl.BorderBox.Y - cl.Margin.Top
}

// LayoutCache stores precomputed child layouts for container widgets.
// Created fresh at the start of each render pass to avoid stale data.
type LayoutCache struct {
	cache map[string][]ChildLayout
	// computedCache stores layouts from the new layout system
	computedCache map[string]*layout.ComputedLayout
}

// NewLayoutCache creates a new empty layout cache.
func NewLayoutCache() *LayoutCache {
	return &LayoutCache{
		cache:         make(map[string][]ChildLayout),
		computedCache: make(map[string]*layout.ComputedLayout),
	}
}

// Store saves the child layouts for a container widget identified by its autoID.
func (lc *LayoutCache) Store(autoID string, layouts []ChildLayout) {
	if lc == nil {
		return
	}
	lc.cache[autoID] = layouts
}

// Get retrieves the cached child layouts for a container widget.
// Returns nil, false if no cached layouts exist.
func (lc *LayoutCache) Get(autoID string) ([]ChildLayout, bool) {
	if lc == nil {
		return nil, false
	}
	layouts, ok := lc.cache[autoID]
	return layouts, ok
}

// StoreComputed saves a computed layout from the new layout system.
func (lc *LayoutCache) StoreComputed(path string, computed *layout.ComputedLayout) {
	if lc == nil || lc.computedCache == nil {
		return
	}
	lc.computedCache[path] = computed
}

// LoadComputed retrieves a computed layout from the new layout system.
func (lc *LayoutCache) LoadComputed(path string) *layout.ComputedLayout {
	if lc == nil || lc.computedCache == nil {
		return nil
	}
	return lc.computedCache[path]
}

// Clear removes all cached layouts (call at start of each frame).
func (lc *LayoutCache) Clear() {
	if lc != nil {
		clear(lc.cache)
		clear(lc.computedCache)
	}
}
