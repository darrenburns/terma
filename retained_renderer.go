package terma

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/darrenburns/terma/layout"
)

type retainedFloat struct {
	entry FloatEntry
	root  *widgetNode
}

type rendererFrameMode string

const (
	rendererFrameNone    rendererFrameMode = ""
	rendererFrameFull    rendererFrameMode = "full"
	rendererFramePartial rendererFrameMode = "partial"
	// rendererFrameReflow rebuilds and lays out, then repaints only damage.
	rendererFrameReflow rendererFrameMode = "reflow"
)

// retainedLayoutNode adapts a retained widget node for layout. It tracks
// layout-phase signal reads, and reuses the node's cached results when nothing
// in its subtree could have changed them. The widget's own layout node (and
// so its children's) is only built when a result has to be computed.
type retainedLayoutNode struct {
	renderer  *Renderer
	node      *widgetNode
	cacheable bool
	raw       layout.LayoutNode
}

// maxLayoutCacheEntries bounds per-node caching; containers may measure a
// child under a few different constraints in one pass.
const maxLayoutCacheEntries = 4

type layoutCacheEntry struct {
	constraints layout.Constraints
	result      layout.ComputedLayout
}

func (r *Renderer) newRetainedLayoutNode(node *widgetNode) *retainedLayoutNode {
	// A layout result depends only on constraints and the subtree's widgets
	// and layout-phase reads; any change there marks the subtree dirty.
	cacheable := r.layoutCacheEnabled && node.subtreeDirtyLevel() < DirtyLayout
	if !cacheable {
		node.layoutCache = node.layoutCache[:0]
		// Reads are recorded afresh as the layout is recomputed. A cacheable
		// node keeps its subscriptions, since a hit skips those reads.
		node.clearDependenciesForPhase(readPhaseLayout)
	}
	return &retainedLayoutNode{renderer: r, node: node, cacheable: cacheable}
}

func (p *retainedLayoutNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	if p.cacheable {
		for _, entry := range p.node.layoutCache {
			if entry.constraints == constraints {
				return entry.result
			}
		}
	}
	result := withSignalRead(p.node, readPhaseLayout, func() layout.ComputedLayout {
		return p.child().ComputeLayout(constraints)
	})
	p.renderer.lastLayoutCount++
	cache := p.node.layoutCache
	if len(cache) == maxLayoutCacheEntries {
		cache = append(cache[:0], cache[1:]...)
	}
	p.node.layoutCache = append(cache, layoutCacheEntry{constraints: constraints, result: result})
	return result
}

func (p *retainedLayoutNode) child() layout.LayoutNode {
	if p.raw == nil {
		p.raw = p.renderer.widgetLayoutNode(p.node)
	}
	return p.raw
}

func (p *retainedLayoutNode) PreservesWidth() bool {
	width, _ := p.sizePreserve()
	return width
}

func (p *retainedLayoutNode) PreservesHeight() bool {
	_, height := p.sizePreserve()
	return height
}

// sizePreserve answers from the node's cached flags when its subtree is clean,
// so asking doesn't force building its layout node.
func (p *retainedLayoutNode) sizePreserve() (width, height bool) {
	node := p.node
	if !p.cacheable || !node.sizePreserveKnown {
		node.preservesWidth, node.preservesHeight = false, false
		if preserver, ok := p.child().(layout.SizePreserver); ok {
			node.preservesWidth, node.preservesHeight = preserver.PreservesWidth(), preserver.PreservesHeight()
		}
		node.sizePreserveKnown = true
	}
	return node.preservesWidth, node.preservesHeight
}

// Update renders the next frame using the retained tree when possible.
// Paint-only signal changes take the partial repaint fast path. Build and
// layout changes reuse clean builds, then lay out and paint the whole tree.
func (r *Renderer) Update(root Widget) []FocusableEntry {
	focusables, _, _ := r.updateInternal(root)
	return focusables
}

func (r *Renderer) updateInternal(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	if r.rootNode == nil || r.fullRenderRequired {
		return r.renderFull(root)
	}
	if r.maxDirtyLevel() >= DirtyLayout {
		return r.renderFrame(root, false)
	}
	if r.hasPaintDirty() {
		return r.renderPartial(root)
	}
	return r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
}

func (r *Renderer) renderFull(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	return r.renderFrame(root, true)
}

// rebuildAll is used for explicit Render calls and initial/forced frames.
// Reactive updates rebuild only dirty nodes and descendants whose inputs may
// have changed when a parent rebuilt. Focus and float collection still traverse
// the whole tree so their ordering and inherited scopes remain correct.
func (r *Renderer) renderFrame(root Widget, rebuildAll bool) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	r.fullRenderRequired = false
	r.lastFrameMode = rendererFrameFull
	r.fullRenderCount++
	r.lastBuildCount = 0
	r.lastLayoutCount = 0
	r.lastPaintCount = 0
	r.lastAssignCount = 0
	r.lastMeasureCount = 0
	r.lastScanCount = 0
	r.lastDamagedRects = nil

	r.focusCollector.Reset()
	r.widgetRegistry.Reset()
	r.floatCollector.Reset()
	r.modalCount = 0

	buildCtx := NewBuildContext(r.focusManager, r.focusedSignal, r.hoveredSignal, r.floatCollector)
	r.rootNode = r.buildRetainedNode(r.rootNode, root, buildCtx, r.focusCollector, rebuildAll)

	if r.rootNode == nil {
		r.lastFocusables = nil
		r.lastLayoutWidth = 0
		r.lastLayoutHeight = 0
		return nil, 0, 0
	}

	// A forced full render recomputes every layout from scratch.
	r.layoutCacheEnabled = !rebuildAll
	constraints := layout.Loose(r.width, r.height)
	r.computeRetainedLayout(r.rootNode, constraints)
	layoutWidth = r.rootNode.layout.Box.BorderBoxWidth()
	layoutHeight = r.rootNode.layout.Box.BorderBoxHeight()
	r.lastLayoutWidth = layoutWidth
	r.lastLayoutHeight = layoutHeight

	// Opening or closing an overlay repaints everything: a modal's backdrop
	// covers the whole screen. Otherwise repaint only what changed.
	if rebuildAll || !r.floatSetMatches() {
		if scr, ok := r.terminal.(uv.Screen); ok {
			screen.Clear(scr)
		}
		ctx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
		r.paintRetainedNode(ctx, r.rootNode, 0, 0, Rect{}, false, true)
		r.placeFloats(ctx, buildCtx, false)
	} else {
		r.lastFrameMode = rendererFrameReflow
		r.reflowPaint(buildCtx)
	}

	focusables = r.focusCollector.Focusables()
	r.lastFocusables = focusables
	r.clearDirtyFlags()
	return focusables, layoutWidth, layoutHeight
}

func (r *Renderer) clearDirtyFlags() {
	r.lastClearCount = r.rootNode.clearDirtyRecursive()
	for _, floatNode := range r.retainedFloats {
		if floatNode.root != nil {
			r.lastClearCount += floatNode.root.clearDirtyRecursive()
		}
	}
}

func (r *Renderer) renderPartial(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	if r.rootNode == nil {
		return r.renderFull(root)
	}

	r.lastScanCount = 0
	damageRects := r.collectDamageRects()
	if len(damageRects) == 0 {
		return r.renderFull(root)
	}

	r.lastFrameMode = rendererFramePartial
	r.partialRenderCount++
	r.lastBuildCount = 0
	r.lastLayoutCount = 0
	r.lastPaintCount = 0
	r.lastDamagedRects = damageRects

	buildCtx := r.rootNode.buildContext
	for _, rect := range damageRects {
		clipped := rect.Intersect(Rect{X: 0, Y: 0, Width: r.width, Height: r.height})
		if clipped.IsEmpty() {
			continue
		}
		r.clearRect(clipped)
		ctx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
		ctx.clip = ctx.clip.Intersect(clipped)
		r.paintRetainedNode(ctx, r.rootNode, 0, 0, clipped, true, false)
		r.paintRetainedFloats(ctx, clipped)
	}

	r.clearDirtyFlags()

	return r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
}

func (r *Renderer) buildRetainedNode(old *widgetNode, widget Widget, ctx BuildContext, fc *FocusCollector, rebuild bool) *widgetNode {
	if widget == nil {
		widget = EmptyWidget{}
	}

	switch w := widget.(type) {
	case disabledWrapper:
		if w.disabled {
			ctx = ctx.WithDisabled()
		}
		return r.buildRetainedNode(old, w.child, ctx, fc, rebuild)
	case inertWrapper:
		return r.buildRetainedNode(old, w.child, ctx, nil, rebuild)
	case FocusTrap:
		if fc != nil && w.TrapsFocus() {
			trapID := w.WidgetID()
			if trapID == "" {
				trapID = ctx.AutoID()
			}
			fc.PushTrap(trapID)
			defer fc.PopTrap()
		}
		if w.Child == nil {
			return r.buildRetainedNode(old, EmptyWidget{}, ctx, fc, rebuild)
		}
		return r.buildRetainedNode(old, w.Child, ctx, fc, rebuild)
	}

	node := old
	if rebuild || node == nil {
		eventID := widgetIdentity(widget, ctx)
		if node == nil || node.identity != eventID {
			if old != nil {
				old.dispose()
			}
			node = newWidgetNode(widget)
		}
		node.autoID = ctx.AutoID()
		node.eventID = eventID
		node.identity = eventID
	} else if _, ok := widget.(Identifiable); ok {
		// The parent's retained build supplied this widget at the same path, so
		// its auto ID is unchanged, but a pointer widget can change its own ID.
		if eventID := widgetIdentity(widget, ctx); eventID != node.identity {
			replaced := node
			node = newWidgetNode(widget)
			// The parent isn't repainting, so inherit the old painted area for
			// the new node's damage to cover.
			node.bounds, node.subtreeBounds = replaced.bounds, replaced.subtreeBounds
			replaced.dispose()
			node.autoID = ctx.AutoID()
			node.eventID = eventID
			node.identity = eventID
		}
	}
	eventID := node.eventID
	rebuild = rebuild || node.dirtyLevel() == DirtyBuild

	node.source = widget
	node.eventWidget = widget
	node.buildContext = ctx

	if rebuild {
		// Its widget may have new properties even if no signal of its own
		// changed, so treat it as changed for layout caching and damage.
		node.setDirtySelf(DirtyBuild)
		node.setDirtySubtree(DirtyBuild)
		node.clearDependenciesForPhase(readPhaseBuild)
		floatStart := r.floatCollector.Len()
		built := withSignalRead(node, readPhaseBuild, func() Widget {
			return widget.Build(ctx)
		})
		if built == nil {
			built = EmptyWidget{}
		}
		node.widget = built
		// Floating.Build registers overlays rather than returning them as
		// children. Keep this node's registrations for frames that reuse Build.
		node.floats = node.floats[:0]
		for i := floatStart; i < len(r.floatCollector.entries); i++ {
			node.floats = append(node.floats, r.floatCollector.entries[i])
			r.floatCollector.entries[i].fresh = true
		}
		r.lastBuildCount++
	} else {
		for _, entry := range node.floats {
			r.floatCollector.Add(entry)
		}
	}

	var ancestorsPushed bool
	if fc != nil {
		if trapper, ok := widget.(FocusTrapper); ok && trapper.TrapsFocus() {
			fc.PushTrap(eventID)
			defer fc.PopTrap()
		}
		fc.Collect(widget, node.autoID, ctx)
		if fc.ShouldTrackAncestor(widget) {
			fc.PushAncestor(widget)
			ancestorsPushed = true
		}
	}
	if ancestorsPushed {
		defer fc.PopAncestor()
	}

	childWidgets := extractChildren(node.widget)
	if !rebuild {
		// The retained build still supplies the same children and wrappers.
		// Walk them to reach dirty descendants and recollect focus scopes.
		for i, childWidget := range childWidgets {
			previous := node.children[i]
			child := r.buildRetainedNode(previous, childWidget, ctx.PushChild(i), fc, false)
			if child != previous {
				// A child replaced itself beneath this clean node. Connect it so
				// its signal changes reach the renderer, and mark the ancestors
				// changed so none reuses a layout computed for the old child.
				child.parent = node
				for ancestor := node; ancestor != nil; ancestor = ancestor.parent {
					ancestor.setDirtySubtree(DirtyBuild)
				}
			}
			node.children[i] = child
		}
		return node
	}
	oldChildren := make(map[string]*widgetNode, len(node.children))
	for _, child := range node.children {
		oldChildren[child.identity] = child
	}

	children := make([]*widgetNode, 0, len(childWidgets))
	for i, childWidget := range childWidgets {
		childCtx := ctx.PushChild(i)
		childID := widgetIdentity(childWidget, childCtx)
		childNode := r.buildRetainedNode(oldChildren[childID], childWidget, childCtx, fc, true)
		if childNode == nil {
			continue
		}
		childNode.parent = node
		children = append(children, childNode)
		delete(oldChildren, childID)
	}

	for _, child := range oldChildren {
		child.dispose()
	}

	node.children = children
	return node
}

func widgetIdentity(widget Widget, ctx BuildContext) string {
	if widget == nil {
		return ctx.AutoID()
	}
	switch w := widget.(type) {
	case disabledWrapper:
		return widgetIdentity(w.child, ctx)
	case inertWrapper:
		return widgetIdentity(w.child, ctx)
	case FocusTrap:
		if w.Child == nil {
			return widgetIdentity(EmptyWidget{}, ctx)
		}
		return widgetIdentity(w.Child, ctx)
	}
	if identifiable, ok := widget.(Identifiable); ok && identifiable.WidgetID() != "" {
		return identifiable.WidgetID()
	}
	return ctx.AutoID()
}

func (r *Renderer) computeRetainedLayout(node *widgetNode, constraints layout.Constraints) {
	if node == nil {
		return
	}
	computed := r.newRetainedLayoutNode(node).ComputeLayout(constraints)
	r.assignComputedLayout(node, computed)
}

// sameChildren reports whether two child layout slices are the same slice.
func sameChildren(a, b []layout.PositionedChild) bool {
	if len(a) != len(b) {
		return false
	}
	return len(a) == 0 || &a[0] == &b[0]
}

// widgetLayoutNode builds the widget's own layout node, with its children
// adapted as retained layout nodes.
func (r *Renderer) widgetLayoutNode(node *widgetNode) layout.LayoutNode {
	return withSignalRead(node, readPhaseLayout, func() layout.LayoutNode {
		if builder, ok := node.widget.(ContainerLayoutBuilder); ok {
			children := make([]layout.LayoutNode, len(node.children))
			for i, child := range node.children {
				children[i] = r.newRetainedLayoutNode(child)
			}
			return builder.BuildContainerLayoutNode(node.buildContext, children)
		}
		if builder, ok := node.widget.(LayoutNodeBuilder); ok {
			return builder.BuildLayoutNode(node.buildContext)
		}
		return buildFallbackLayoutNode(node.widget, node.buildContext)
	})
}

func (r *Renderer) assignComputedLayout(node *widgetNode, computed layout.ComputedLayout) {
	if node == nil {
		return
	}
	// A clean subtree handed back the very result it was last assigned (a
	// cache hit returns the same Children slice) has nothing to update.
	if node.subtreeDirtyLevel() == DirtyNone && computed.Box == node.layout.Box && sameChildren(computed.Children, node.layout.Children) {
		node.layoutReused = true
		return
	}
	node.layoutReused = false
	node.layout = computed
	r.lastAssignCount++
	if observer, ok := node.widget.(LayoutObserver); ok {
		withSignalRead(node, readPhaseLayout, func() struct{} {
			observer.OnLayout(node.buildContext, LayoutMetrics{layout: computed})
			return struct{}{}
		})
	}
	node.updateIntrinsicCache()
	limit := min(len(node.children), len(computed.Children))
	for i := 0; i < limit; i++ {
		r.assignComputedLayout(node.children[i], computed.Children[i].Layout)
	}
	for i := limit; i < len(node.children); i++ {
		node.children[i].layoutReused = false
		node.children[i].layout = layout.ComputedLayout{}
		node.children[i].updateIntrinsicCache()
	}
}

func (r *Renderer) paintRetainedNode(ctx *RenderContext, node *widgetNode, screenX, screenY int, damage Rect, partial bool, recordRegistry bool) Rect {
	if node == nil {
		return Rect{}
	}
	if partial && !node.subtreeBounds.Intersects(damage) {
		return Rect{}
	}

	selfCtx := *ctx
	selfCtx.currentEventID = node.eventID
	ctx = &selfCtx

	box := node.layout.Box
	borderX, borderY := box.BorderOrigin()
	contentX, contentY := box.ContentOrigin()

	absBorderX := screenX + borderX
	absBorderY := screenY + borderY
	absContentX := screenX + contentX
	absContentY := screenY + contentY

	trueAbsBorderX := ctx.X + absBorderX
	trueAbsBorderY := ctx.Y + absBorderY

	nodeBounds := Rect{
		X:      trueAbsBorderX,
		Y:      trueAbsBorderY,
		Width:  box.Width,
		Height: box.Height,
	}
	if partial && !nodeBounds.Intersects(damage) && !node.subtreeBounds.Intersects(damage) {
		return Rect{}
	}

	var style Style
	if styled, ok := node.widget.(Styled); ok {
		style = styled.GetStyle()
	}

	if r.geometryOnly {
		// A clean subtree with the same layout in the same place is exactly as
		// it was: replay its hit-test entries instead of walking it.
		if recordRegistry && node.layoutReused && node.subtreeDirtyLevel() == DirtyNone && node.bounds == nodeBounds && node.registered != nil {
			start := len(r.widgetRegistry.entries)
			r.widgetRegistry.entries = append(r.widgetRegistry.entries, node.registered...)
			node.registered = r.registeredSince(start)
			return node.subtreeBounds
		}

		// Measure only: record where everything is without drawing.
		r.lastMeasureCount++
		registryStart := len(r.widgetRegistry.entries)
		if recordRegistry {
			r.recordRegistry(node, nodeBounds)
		}
		subtreeBounds := nodeBounds
		r.forEachChildContext(ctx, node, style, absBorderX, absBorderY, absContentX, absContentY, func(childCtx *RenderContext, child *widgetNode, x, y int) {
			subtreeBounds = subtreeBounds.Union(r.paintRetainedNode(childCtx, child, x, y, damage, partial, recordRegistry))
		})
		if scrollable, ok := node.widget.(Scrollable); ok && scrollable.State != nil {
			usableBox := box.UsableContentBox()
			scrollable.State.updateHorizontalLayout(usableBox.Width, box.VirtualWidth)
			scrollable.State.updateLayout(usableBox.Height, box.VirtualHeight)
		}
		r.recordReflowDamage(node, nodeBounds, subtreeBounds)
		if recordRegistry {
			node.registered = r.registeredSince(registryStart)
		}
		node.prevBox = box
		node.bounds = nodeBounds
		node.subtreeBounds = subtreeBounds
		return subtreeBounds
	}

	r.lastPaintCount++

	if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
		sampleColor := style.BackgroundColor.ColorAt(box.Width, box.Height, 0, 0)
		useBackdrop := !sampleColor.IsOpaque()

		if useBackdrop {
			backdropCtx := ctx.SubContext(absBorderX, absBorderY, box.Width, box.Height)
			backdropCtx.DrawBackdrop(0, 0, box.Width, box.Height, style.BackgroundColor)
		} else {
			for row := 0; row < box.Height; row++ {
				absY := trueAbsBorderY + row
				if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
					continue
				}
				for col := 0; col < box.Width; col++ {
					absX := trueAbsBorderX + col
					if absX < ctx.clip.X || absX >= ctx.clip.X+ctx.clip.Width {
						continue
					}
					cellColor := style.BackgroundColor.ColorAt(box.Width, box.Height, col, row)
					ctx.terminal.SetCell(absX, absY, &uv.Cell{
						Content: " ",
						Width:   1,
						Style:   uv.Style{Bg: cellColor.toANSI()},
					})
				}
			}
		}
	}

	if !style.Border.IsZero() {
		borderCtx := ctx.SubContext(absBorderX, absBorderY, box.Width, box.Height)
		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
			parentCallback := ctx.inheritedBgAt

			borderCtx.inheritedBgAt = func(absX, absY int) Color {
				relX := absX - originX
				relY := absY - originY
				cellColor := bg.ColorAt(w, h, relX, relY)
				if !cellColor.IsOpaque() && parentCallback != nil {
					inherited := parentCallback(absX, absY)
					if !inherited.IsSet() {
						inherited = Black
					}
					cellColor = cellColor.BlendOver(inherited)
				}
				return cellColor
			}
		}
		borderCtx.DrawBorder(0, 0, box.Width, box.Height, style.Border)
	}

	node.clearDependenciesForPhase(readPhasePaint)
	if renderable, ok := node.widget.(Renderable); ok {
		contentCtx := ctx.SubContext(absContentX, absContentY, box.ContentWidth(), box.ContentHeight())
		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
			parentCallback := ctx.inheritedBgAt

			contentCtx.inheritedBgAt = func(absX, absY int) Color {
				relX := absX - originX
				relY := absY - originY
				cellColor := bg.ColorAt(w, h, relX, relY)
				if !cellColor.IsOpaque() && parentCallback != nil {
					inherited := parentCallback(absX, absY)
					if !inherited.IsSet() {
						inherited = Black
					}
					cellColor = cellColor.BlendOver(inherited)
				}
				return cellColor
			}
		}
		withSignalRead(node, readPhasePaint, func() struct{} {
			renderable.Render(contentCtx)
			return struct{}{}
		})
	}

	registryStart := len(r.widgetRegistry.entries)
	if recordRegistry {
		r.recordRegistry(node, nodeBounds)
	}

	subtreeBounds := nodeBounds
	r.forEachChildContext(ctx, node, style, absBorderX, absBorderY, absContentX, absContentY, func(childCtx *RenderContext, child *widgetNode, x, y int) {
		childBounds := r.paintRetainedNode(childCtx, child, x, y, damage, partial, recordRegistry)
		if childBounds.IsEmpty() && partial {
			// Skipped outside the damage; its recorded area is still current.
			childBounds = child.subtreeBounds
		}
		subtreeBounds = subtreeBounds.Union(childBounds)
	})

	if scrollable, ok := node.widget.(Scrollable); ok {
		if scrollable.State != nil {
			usableBox := box.UsableContentBox()
			scrollable.State.updateHorizontalLayout(usableBox.Width, box.VirtualWidth)
			scrollable.State.updateLayout(usableBox.Height, box.VirtualHeight)
		}

		if box.IsScrollableY() && !scrollable.DisableScroll {
			focused := ctx.focusManager != nil && ctx.IsFocused(node.widget)
			scrollbarCtx := ctx.SubContext(absContentX, absContentY, box.ContentWidth(), box.ContentHeight())
			if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
				bg := style.BackgroundColor
				w, h := box.Width, box.Height
				originX, originY := trueAbsBorderX, trueAbsBorderY
				scrollbarCtx.inheritedBgAt = func(absX, absY int) Color {
					relX := absX - originX
					relY := absY - originY
					return bg.ColorAt(w, h, relX, relY)
				}
			}
			scrollable.renderScrollbar(scrollbarCtx, box.ScrollOffsetY, focused)
		}
	}

	if recordRegistry {
		node.registered = r.registeredSince(registryStart)
	}
	if !partial {
		node.prevBox = box
	}
	node.bounds = nodeBounds
	node.subtreeBounds = subtreeBounds
	return subtreeBounds
}

// registeredSince returns a capacity-limited view of the hit-test entries
// recorded since start, so later appends never write through it.
func (r *Renderer) registeredSince(start int) []WidgetEntry {
	entries := r.widgetRegistry.entries
	return entries[start:len(entries):len(entries)]
}

// forEachChildContext derives each child's clipped/scrolled context exactly as
// painting does, so measuring and painting agree on where children land.
func (r *Renderer) forEachChildContext(ctx *RenderContext, node *widgetNode, style Style, absBorderX, absBorderY, absContentX, absContentY int, visit func(*RenderContext, *widgetNode, int, int)) {
	if len(node.children) == 0 {
		return
	}
	box := node.layout.Box
	var childClipCtx *RenderContext
	if _, isStack := node.widget.(Stack); isStack {
		childClipCtx = ctx.OverflowSubContext(absBorderX, absBorderY, box.Width, box.Height)
	} else if box.IsScrollableX() || box.IsScrollableY() {
		usableBox := box.UsableContentBox()
		childClipCtx = ctx.ScrolledSubContext(
			absContentX,
			absContentY,
			usableBox.Width,
			usableBox.Height,
			box.ScrollOffsetX,
			box.ScrollOffsetY,
		)
	} else {
		usableBox := box.UsableContentBox()
		childClipCtx = ctx.SubContext(absContentX, absContentY, usableBox.Width, usableBox.Height)
	}

	if !r.geometryOnly && style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
		bg := style.BackgroundColor
		w, h := box.Width, box.Height
		originX, originY := ctx.X+absBorderX, ctx.Y+absBorderY
		parentCallback := ctx.inheritedBgAt

		childClipCtx.inheritedBgAt = func(absX, absY int) Color {
			relX := absX - originX
			relY := absY - originY
			cellColor := bg.ColorAt(w, h, relX, relY)
			if !cellColor.IsOpaque() && parentCallback != nil {
				inherited := parentCallback(absX, absY)
				if !inherited.IsSet() {
					inherited = Black
				}
				cellColor = cellColor.BlendOver(inherited)
			}
			return cellColor
		}
	}

	for i, child := range node.children {
		if i >= len(node.layout.Children) {
			break
		}
		pos := node.layout.Children[i]
		visit(childClipCtx, child, pos.X, pos.Y)
	}
}

func (r *Renderer) recordRegistry(node *widgetNode, bounds Rect) {
	eventWidget := node.eventWidget
	if eventWidget == nil {
		eventWidget = node.widget
	}
	r.widgetRegistry.Record(node.widget, eventWidget, node.eventID, bounds)
}

// recordReflowDamage marks what a reflow frame must repaint for this node:
// its old and new subtree if it was invalidated, otherwise its own area if it
// moved, resized, or its box changed (for example a scrollbar's extent).
//
// Nodes are only rebuilt, added or removed beneath an invalidated ancestor,
// whose subtree damage already covers them.
func (r *Renderer) recordReflowDamage(node *widgetNode, bounds, subtreeBounds Rect) {
	add := func(rect Rect) {
		if !rect.IsEmpty() {
			r.reflowDamage = append(r.reflowDamage, rect)
		}
	}
	switch {
	case node.dirtyLevel() != DirtyNone:
		add(node.subtreeBounds)
		add(subtreeBounds)
	case node.bounds != bounds || node.prevBox != node.layout.Box:
		add(node.bounds)
		add(bounds)
	}
}

// reflowPaint repaints only what changed after a rebuild and relayout. A
// measuring pass records every node's new position, the hit-test registry and
// the damage; the partial painter then redraws just the damaged areas.
func (r *Renderer) reflowPaint(buildCtx BuildContext) {
	ctx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
	r.reflowDamage = r.reflowDamage[:0]
	r.floatsChanged = false
	r.geometryOnly = true
	r.paintRetainedNode(ctx, r.rootNode, 0, 0, Rect{}, false, true)
	r.placeFloats(ctx, buildCtx, true)
	r.geometryOnly = false

	fullScreen := Rect{Width: r.width, Height: r.height}
	if r.floatsChanged {
		r.reflowDamage = append(r.reflowDamage, fullScreen)
	}
	rects := coalesceDamage(r.reflowDamage, fullScreen)
	r.lastDamagedRects = rects
	for _, rect := range rects {
		r.clearRect(rect)
		paintCtx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
		paintCtx.clip = paintCtx.clip.Intersect(rect)
		r.paintRetainedNode(paintCtx, r.rootNode, 0, 0, rect, true, false)
		r.paintRetainedFloats(paintCtx, rect)
	}
}

// maxDamageRects bounds how many separate regions are repainted; beyond it
// one bounding rectangle is cheaper than many tree walks.
const maxDamageRects = 8

// coalesceDamage clips rects to the screen and merges overlapping ones, so
// no cell is repainted twice.
func coalesceDamage(rects []Rect, screen Rect) []Rect {
	var merged []Rect
	for _, rect := range rects {
		rect = rect.Intersect(screen)
		if rect.IsEmpty() {
			continue
		}
		// Absorb any existing rects this one overlaps, repeating as it grows.
		for i := 0; i < len(merged); {
			if merged[i].Intersects(rect) {
				rect = rect.Union(merged[i])
				merged = append(merged[:i], merged[i+1:]...)
				i = 0
				continue
			}
			i++
		}
		merged = append(merged, rect)
	}
	if len(merged) > maxDamageRects {
		union := merged[0]
		for _, rect := range merged[1:] {
			union = union.Union(rect)
		}
		return []Rect{union}
	}
	return merged
}

// floatSetMatches reports whether this frame's overlays correspond one to one
// with the last frame's: the same number, with the same modal flags.
func (r *Renderer) floatSetMatches() bool {
	if len(r.retainedFloats) != r.floatCollector.Len() {
		return false
	}
	for i, retained := range r.retainedFloats {
		if retained.entry.Config.Modal != r.floatCollector.entries[i].Config.Modal {
			return false
		}
	}
	return true
}

// placeFloats builds, lays out, positions and paints the frame's overlays.
// When measuring (a reflow frame), overlays whose owner reused its build keep
// theirs too, and nothing is drawn: positions, hit targets and damage are
// recorded for the partial painter instead.
func (r *Renderer) placeFloats(ctx *RenderContext, buildCtx BuildContext, measure bool) {
	oldFloats := r.retainedFloats
	r.retainedFloats = nil
	if r.floatCollector.Len() == 0 {
		for _, old := range oldFloats {
			if old.root != nil {
				if measure {
					r.floatsChanged = true
				}
				old.root.dispose()
			}
		}
		return
	}

	for i := 0; i < len(r.floatCollector.entries); i++ {
		entry := r.floatCollector.entries[i]
		focusableCountBefore := r.focusCollector.Len()
		child := entry.Child
		if entry.Config.Modal {
			child = FocusTrap{
				ID:     fmt.Sprintf("__modal_float_%d", i),
				Active: true,
				Child:  child,
			}
		}

		var oldRoot *widgetNode
		if i < len(oldFloats) {
			oldRoot = oldFloats[i].root
		}
		floatRoot := r.buildRetainedNode(oldRoot, child, buildCtx, r.focusCollector, !measure || entry.fresh)
		if measure && oldRoot != nil && floatRoot != oldRoot {
			// Nothing else records where the replaced overlay was drawn.
			r.reflowDamage = append(r.reflowDamage, oldRoot.subtreeBounds)
		}
		r.computeRetainedLayout(floatRoot, layout.Loose(r.width, r.height))

		floatWidth := floatRoot.layout.Box.MarginBoxWidth()
		floatHeight := floatRoot.layout.Box.MarginBoxHeight()

		var x, y int
		if entry.Config.AnchorID != "" {
			anchor := r.widgetRegistry.WidgetByID(entry.Config.AnchorID)
			x, y = calculateAnchorPosition(anchor, entry.Config.Anchor, floatWidth, floatHeight, entry.Config.Offset)
		} else {
			x, y = calculateAbsolutePosition(entry.Config.Position, r.width, r.height, floatWidth, floatHeight, entry.Config.Offset)
		}
		x, y = clampToScreen(x, y, floatWidth, floatHeight, r.width, r.height)

		entry.X = x
		entry.Y = y
		entry.Width = floatWidth
		entry.Height = floatHeight
		r.floatCollector.entries[i] = entry

		if entry.Config.Modal {
			r.modalCount++
			if !measure {
				r.renderModalBackdrop(ctx, entry.Config.BackdropColor)
			}

			focusedID := r.focusManager.FocusedID()
			alreadyInside := false
			for _, fe := range r.focusCollector.Focusables()[focusableCountBefore:] {
				if fe.ID == focusedID {
					alreadyInside = true
					break
				}
			}
			if !alreadyInside {
				if firstID := r.focusCollector.FirstFocusableIDAfter(focusableCountBefore); firstID != "" {
					pendingFocusID = firstID
				}
			}
		}

		r.paintRetainedNode(ctx, floatRoot, x, y, Rect{}, false, true)
		r.retainedFloats = append(r.retainedFloats, retainedFloat{
			entry: entry,
			root:  floatRoot,
		})
	}

	for i := len(r.retainedFloats); i < len(oldFloats); i++ {
		if oldFloats[i].root != nil {
			oldFloats[i].root.dispose()
		}
	}
	if measure && !sameFloatSet(oldFloats, r.retainedFloats) {
		// Nested overlays (registered by overlay content) appeared,
		// disappeared or changed modality.
		r.floatsChanged = true
	}
}

func sameFloatSet(a, b []retainedFloat) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].entry.Config.Modal != b[i].entry.Config.Modal {
			return false
		}
	}
	return true
}

func (r *Renderer) paintRetainedFloats(ctx *RenderContext, damage Rect) {
	for _, floatNode := range r.retainedFloats {
		if floatNode.entry.Config.Modal {
			r.renderModalBackdrop(ctx, floatNode.entry.Config.BackdropColor)
		}
		if floatNode.root != nil {
			r.paintRetainedNode(ctx, floatNode.root, floatNode.entry.X, floatNode.entry.Y, damage, true, false)
		}
	}
}

func (r *Renderer) clearRect(rect Rect) {
	if rect.IsEmpty() {
		return
	}
	if scr, ok := r.terminal.(uv.Screen); ok {
		screen.ClearArea(scr, uv.Rect(rect.X, rect.Y, rect.Width, rect.Height))
		return
	}
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			r.terminal.SetCell(x, y, nil)
		}
	}
}

func (r *Renderer) collectDamageRects() []Rect {
	var rects []Rect
	appendDirty := func(node *widgetNode) {}
	appendDirty = func(node *widgetNode) {
		if node == nil {
			return
		}
		r.lastScanCount++
		if node.dirtyLevel() == DirtyPaint && !node.subtreeBounds.IsEmpty() {
			rects = append(rects, node.subtreeBounds)
		}
		for _, child := range node.children {
			// Dirty nodes only lie beneath ancestors whose subtree flag is set.
			if child.subtreeDirtyLevel() != DirtyNone {
				appendDirty(child)
			}
		}
	}
	appendDirty(r.rootNode)
	for _, floatNode := range r.retainedFloats {
		appendDirty(floatNode.root)
	}
	return rects
}

func (r *Renderer) maxDirtyLevel() dirtyLevel {
	level := DirtyNone
	if r.rootNode != nil && r.rootNode.subtreeDirtyLevel() > level {
		level = r.rootNode.subtreeDirtyLevel()
	}
	for _, floatNode := range r.retainedFloats {
		if floatNode.root != nil && floatNode.root.subtreeDirtyLevel() > level {
			level = floatNode.root.subtreeDirtyLevel()
		}
	}
	return level
}

func (r *Renderer) hasPaintDirty() bool {
	return r.maxDirtyLevel() == DirtyPaint
}
