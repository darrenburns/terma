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
)

type phaseTrackingLayoutNode struct {
	node  *widgetNode
	child layout.LayoutNode
}

func (p phaseTrackingLayoutNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	return withSignalRead(p.node, readPhaseLayout, func() layout.ComputedLayout {
		return p.child.ComputeLayout(constraints)
	})
}

func (p phaseTrackingLayoutNode) PreservesWidth() bool {
	if preserver, ok := p.child.(layout.SizePreserver); ok {
		return preserver.PreservesWidth()
	}
	return false
}

func (p phaseTrackingLayoutNode) PreservesHeight() bool {
	if preserver, ok := p.child.(layout.SizePreserver); ok {
		return preserver.PreservesHeight()
	}
	return false
}

// Update renders the next frame using the retained tree when possible.
// Paint-only signal changes take the partial repaint fast path; everything
// else falls back to a normal full build/layout/paint pass.
func (r *Renderer) Update(root Widget) []FocusableEntry {
	focusables, _, _ := r.updateInternal(root)
	return focusables
}

func (r *Renderer) updateInternal(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	if r.rootNode == nil || r.fullRenderRequired {
		return r.renderFull(root)
	}
	if r.maxDirtyLevel() >= DirtyLayout {
		return r.renderFull(root)
	}
	if r.hasPaintDirty() {
		return r.renderPartial(root)
	}
	return r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
}

func (r *Renderer) renderFull(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	r.fullRenderRequired = false
	r.lastFrameMode = rendererFrameFull
	r.fullRenderCount++
	r.lastBuildCount = 0
	r.lastLayoutCount = 0
	r.lastPaintCount = 0
	r.lastDamagedRects = nil

	r.focusCollector.Reset()
	r.widgetRegistry.Reset()
	r.floatCollector.Reset()
	r.modalCount = 0

	buildCtx := NewBuildContext(r.focusManager, r.focusedSignal, r.hoveredSignal, r.floatCollector)
	r.rootNode = r.buildRetainedNode(r.rootNode, root, buildCtx, r.focusCollector)

	if r.rootNode == nil {
		r.lastFocusables = nil
		r.lastLayoutWidth = 0
		r.lastLayoutHeight = 0
		return nil, 0, 0
	}

	constraints := layout.Loose(r.width, r.height)
	r.computeRetainedLayout(r.rootNode, constraints)
	layoutWidth = r.rootNode.layout.Box.BorderBoxWidth()
	layoutHeight = r.rootNode.layout.Box.BorderBoxHeight()
	r.lastLayoutWidth = layoutWidth
	r.lastLayoutHeight = layoutHeight

	if scr, ok := r.terminal.(uv.Screen); ok {
		screen.Clear(scr)
	}

	ctx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
	r.paintRetainedNode(ctx, r.rootNode, 0, 0, Rect{}, false, true)
	r.buildAndPaintFloats(ctx, buildCtx)

	focusables = r.focusCollector.Focusables()
	r.lastFocusables = focusables
	r.rootNode.clearDirtyRecursive()
	for _, floatNode := range r.retainedFloats {
		if floatNode.root != nil {
			floatNode.root.clearDirtyRecursive()
		}
	}
	return focusables, layoutWidth, layoutHeight
}

func (r *Renderer) renderPartial(root Widget) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	if r.rootNode == nil {
		return r.renderFull(root)
	}

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

	r.rootNode.clearDirtyRecursive()
	for _, floatNode := range r.retainedFloats {
		if floatNode.root != nil {
			floatNode.root.clearDirtyRecursive()
		}
	}

	return r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
}

func (r *Renderer) buildRetainedNode(old *widgetNode, widget Widget, ctx BuildContext, fc *FocusCollector) *widgetNode {
	if widget == nil {
		widget = EmptyWidget{}
	}

	switch w := widget.(type) {
	case disabledWrapper:
		if w.disabled {
			ctx = ctx.WithDisabled()
		}
		return r.buildRetainedNode(old, w.child, ctx, fc)
	case inertWrapper:
		return r.buildRetainedNode(old, w.child, ctx, nil)
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
			return r.buildRetainedNode(old, EmptyWidget{}, ctx, fc)
		}
		return r.buildRetainedNode(old, w.Child, ctx, fc)
	}

	eventID := widgetIdentity(widget, ctx)
	node := old
	if node == nil || node.identity != eventID {
		if old != nil {
			old.dispose()
		}
		node = newWidgetNode(widget)
	} else if old != nil {
		old.clearDependenciesForPhase(readPhaseBuild)
	}

	node.source = widget
	node.eventWidget = widget
	node.autoID = ctx.AutoID()
	node.eventID = eventID
	node.identity = eventID
	node.buildContext = ctx

	built := withSignalRead(node, readPhaseBuild, func() Widget {
		return widget.Build(ctx)
	})
	if built == nil {
		built = EmptyWidget{}
	}
	node.widget = built
	r.lastBuildCount++

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

	childWidgets := extractChildren(built)
	oldChildren := make(map[string]*widgetNode, len(node.children))
	for _, child := range node.children {
		oldChildren[child.identity] = child
	}

	children := make([]*widgetNode, 0, len(childWidgets))
	for i, childWidget := range childWidgets {
		childCtx := ctx.PushChild(i)
		childID := widgetIdentity(childWidget, childCtx)
		childNode := r.buildRetainedNode(oldChildren[childID], childWidget, childCtx, fc)
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
	node.clearDependenciesForPhase(readPhaseLayout)
	layoutNode := r.layoutNodeFor(node)
	computed := layoutNode.ComputeLayout(constraints)
	r.assignComputedLayout(node, computed)
}

func (r *Renderer) layoutNodeFor(node *widgetNode) layout.LayoutNode {
	raw := withSignalRead(node, readPhaseLayout, func() layout.LayoutNode {
		if builder, ok := node.widget.(ContainerLayoutBuilder); ok {
			children := make([]layout.LayoutNode, len(node.children))
			for i, child := range node.children {
				children[i] = buildRetainedChildLayoutNode(child)
			}
			return builder.BuildContainerLayoutNode(node.buildContext, children)
		}
		if builder, ok := node.widget.(LayoutNodeBuilder); ok {
			return builder.BuildLayoutNode(node.buildContext)
		}
		return buildFallbackLayoutNode(node.widget, node.buildContext)
	})
	return phaseTrackingLayoutNode{node: node, child: raw}
}

func (r *Renderer) assignComputedLayout(node *widgetNode, computed layout.ComputedLayout) {
	if node == nil {
		return
	}
	node.layout = computed
	r.lastLayoutCount++
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
		node.children[i].layout = layout.ComputedLayout{}
		node.children[i].updateIntrinsicCache()
	}
}

func buildRetainedChildLayoutNode(child *widgetNode) layout.LayoutNode {
	if child == nil {
		return &layout.BoxNode{}
	}
	child.clearDependenciesForPhase(readPhaseLayout)
	return withSignalRead(child, readPhaseLayout, func() layout.LayoutNode {
		var raw layout.LayoutNode
		if builder, ok := child.widget.(ContainerLayoutBuilder); ok {
			children := make([]layout.LayoutNode, len(child.children))
			for i, grandChild := range child.children {
				children[i] = buildRetainedChildLayoutNode(grandChild)
			}
			raw = builder.BuildContainerLayoutNode(child.buildContext, children)
		} else if builder, ok := child.widget.(LayoutNodeBuilder); ok {
			raw = builder.BuildLayoutNode(child.buildContext)
		} else {
			raw = buildFallbackLayoutNode(child.widget, child.buildContext)
		}
		return phaseTrackingLayoutNode{node: child, child: raw}
	})
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

	r.lastPaintCount++

	var style Style
	if styled, ok := node.widget.(Styled); ok {
		style = styled.GetStyle()
	}

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

	if recordRegistry {
		eventWidget := node.eventWidget
		if eventWidget == nil {
			eventWidget = node.widget
		}
		r.widgetRegistry.Record(node.widget, eventWidget, node.eventID, nodeBounds)
	}

	subtreeBounds := nodeBounds
	if len(node.children) > 0 {
		var childClipCtx *RenderContext
		_, isStack := node.widget.(Stack)
		if isStack {
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

		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
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
			if childBounds := r.paintRetainedNode(childClipCtx, child, pos.X, pos.Y, damage, partial, recordRegistry); !childBounds.IsEmpty() {
				subtreeBounds = subtreeBounds.Union(childBounds)
			}
		}
	}

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

	node.bounds = nodeBounds
	node.subtreeBounds = subtreeBounds
	return subtreeBounds
}

func (r *Renderer) buildAndPaintFloats(ctx *RenderContext, buildCtx BuildContext) {
	oldFloats := r.retainedFloats
	r.retainedFloats = nil
	if r.floatCollector.Len() == 0 {
		for _, old := range oldFloats {
			if old.root != nil {
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
		floatRoot := r.buildRetainedNode(oldRoot, child, buildCtx, r.focusCollector)
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
			r.renderModalBackdrop(ctx, entry.Config.BackdropColor)

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
		if node.dirtyLevel() == DirtyPaint && !node.subtreeBounds.IsEmpty() {
			rects = append(rects, node.subtreeBounds)
		}
		for _, child := range node.children {
			appendDirty(child)
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
