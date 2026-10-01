package terma

import (
	"fmt"
	"math"
	"reflect"
	"sort"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/darrenburns/terma/layout"
)

type retainedFloat struct {
	entry FloatEntry
	root  *widgetNode
	// The overlay's hit-test entries are registry entries [registryStart, registryEnd).
	registryStart int
	registryEnd   int
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
// child under a few different constraints in one pass. A Scrollable whose
// content overflows measures it under five, and must find all of them cached
// or every scroll step lays its whole content out again.
const maxLayoutCacheEntries = 8

type layoutCacheEntry struct {
	constraints layout.Constraints
	result      layout.ComputedLayout
}

func (r *Renderer) newRetainedLayoutNode(node *widgetNode) *retainedLayoutNode {
	// Each node has one adapter per layout pass, kept on the node, so a pass
	// allocates nothing per child and never discards results it computed.
	if node.layoutEpoch == r.layoutEpoch {
		return &node.layoutAdapter
	}
	node.layoutEpoch = r.layoutEpoch
	// A layout result depends only on constraints and the subtree's widgets
	// and layout-phase reads; any change there marks the subtree dirty.
	cacheable := r.layoutCacheEnabled && node.subtreeDirtyLevel() < DirtyLayout
	node.layoutAdapter = retainedLayoutNode{renderer: r, node: node, cacheable: cacheable}
	if !cacheable && r.layoutCacheEnabled && node.dirtyLevel() < DirtyLayout && r.patchLayoutCache(node) {
		node.layoutAdapter.cacheable = true
		cacheable = true
	}
	if !cacheable {
		node.layoutCache = node.layoutCache[:0]
		// Reads are recorded afresh as the layout is recomputed. A cacheable
		// node keeps its subscriptions, since a hit skips those reads.
		node.clearDependenciesForPhase(readPhaseLayout)
	}
	return &node.layoutAdapter
}

// boxOnlyLayout reports whether a layout node uses its children's results
// only through their boxes (and SizePreserver answers), and wraps them only
// according to their widgets' dimensions. Such a node's result, given the
// same constraints, is unchanged by a child whose box is unchanged.
func boxOnlyLayout(raw layout.LayoutNode) bool {
	switch raw.(type) {
	case *layout.ColumnNode, *layout.RowNode, *layout.ScrollableNode, passThroughLayoutNode:
		return true
	}
	return false
}

// patchLayoutCache revalidates the cached layouts of a clean node whose
// subtree changed, without laying out its clean children again. Each changed
// child is laid out under every constraint it was measured with last time;
// if all of its boxes are unchanged, the node's own results are too, apart
// from the child layouts they embed, which are replaced in place. A row
// rebuilding in a list of thousands then costs one row's layout.
//
// It reports false, leaving the node to be laid out in full, whenever that
// can't be shown: a changed box or size preference, changed dimensions that
// decide how the node wraps the child, or a child layout it can't match to a
// constraint.
func (r *Renderer) patchLayoutCache(node *widgetNode) bool {
	if !node.boxOnlyLayout || len(node.layoutCache) == 0 {
		return false
	}
	widgets := extractChildren(node.widget)
	if len(widgets) != len(node.children) {
		return false
	}
	type childPatch struct {
		index  int
		before []layoutCacheEntry
		after  []layout.ComputedLayout
	}
	var patches []childPatch
	for i, child := range node.children {
		if child.subtreeDirtyLevel() < DirtyLayout {
			continue
		}
		if !child.parentDimsKnown || GetWidgetDimensionSet(widgets[i]) != child.parentDims || len(child.layoutCache) == 0 {
			return false
		}
		patch := childPatch{index: i, before: append([]layoutCacheEntry(nil), child.layoutCache...)}
		preserveKnown, preservesWidth, preservesHeight := child.sizePreserveKnown, child.preservesWidth, child.preservesHeight
		adapter := r.newRetainedLayoutNode(child)
		patch.after = make([]layout.ComputedLayout, len(patch.before))
		for k := range patch.before {
			patch.after[k] = adapter.ComputeLayout(patch.before[k].constraints)
			if patch.after[k].Box != patch.before[k].result.Box {
				return false
			}
		}
		if preserveKnown {
			if width, height := adapter.sizePreserve(); width != preservesWidth || height != preservesHeight {
				return false
			}
		}
		patches = append(patches, patch)
	}

	// Match every embedded layout of a changed child to the constraint it was
	// computed under before changing anything.
	type replacement struct {
		target *layout.ComputedLayout
		layout layout.ComputedLayout
	}
	var replacements []replacement
	for e := range node.layoutCache {
		children := node.layoutCache[e].result.Children
		for _, patch := range patches {
			if patch.index >= len(children) {
				return false
			}
			target := &children[patch.index].Layout
			match := -1
			for k := range patch.before {
				old := patch.before[k].result
				if old.Box != target.Box || !sameChildren(old.Children, target.Children) {
					continue
				}
				// Equal boxes with no children can come from several
				// constraints; any of them serves only if the new results
				// are childless too, and so identical.
				if match >= 0 && (len(patch.after[match].Children) > 0 || len(patch.after[k].Children) > 0) {
					return false
				}
				match = k
			}
			if match < 0 {
				return false
			}
			replacements = append(replacements, replacement{target: target, layout: patch.after[match]})
		}
	}
	for _, rep := range replacements {
		*rep.target = rep.layout
	}
	return true
}

func (p *retainedLayoutNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	// Construction discards stale entries for dirty and forced layouts. Results
	// computed since then are reusable too: flex and stretch can measure the
	// same child under identical constraints several times within one frame.
	cache := p.node.layoutCache
	for i := range cache {
		if cache[i].constraints == constraints {
			return cache[i].result
		}
	}
	result := withSignalRead(p.node, readPhaseLayout, func() layout.ComputedLayout {
		return p.child().ComputeLayout(constraints)
	})
	p.renderer.lastLayoutCount++
	cache = p.node.layoutCache
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
// Headless Dispatch work settles as described by Render.
func (r *Renderer) Update(root Widget) []FocusableEntry {
	focusables, _, _ := r.renderHeadlessFrames(root, true)
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
		// A paint-only update can scroll an anchor. Resolve its geometry after
		// measuring the main tree, rather than reusing the previous snapshot.
		for _, float := range r.retainedFloats {
			if float.entry.BuildChild != nil && float.entry.Config.AnchorID != "" {
				return r.renderFrame(root, false)
			}
		}
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
	buildCtx.hoverTarget = r.hoverTarget
	buildCtx.renderer = r
	r.rootNode = r.buildRetainedNode(r.rootNode, root, buildCtx, r.focusCollector, rebuildAll)
	r.floatCollector.raiseTopmost()

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
		ctx := NewRenderContext(r.terminal, r.width, r.height, r.focusCollector, r.focusManager, buildCtx, r.widgetRegistry)
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
	damageRects, found := r.collectDamageRects()
	if !found {
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
		ctx := NewRenderContext(r.terminal, r.width, r.height, r.focusCollector, r.focusManager, buildCtx, r.widgetRegistry)
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
	// Hover selectors and retained hit entries capture the overlay scope.
	// A moved overlay may retain its widget identity but need a new scope.
	rebuild = rebuild || node.buildContext.hoverScope != ctx.hoverScope

	node.source = widget
	node.eventWidget = widget
	node.buildContext = ctx

	if rebuild {
		// Its widget may have new properties even if no signal of its own
		// changed, so treat it as changed for layout caching and damage.
		// Its recorded hit targets hold the old widget, so they can't be
		// replayed even if it's measured at the same place again.
		node.registered = nil
		// Its children may change, so which of them show must be found again.
		node.shownValid = false
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
		if needsOwnNode(built, widget) {
			// A composite returned directly from Build only works once its own
			// Build runs (a Dialog registers its overlay, a Button builds its
			// label), so it gets a node of its own beneath this one.
			built = passThrough{child: built}
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
		fc.CollectGlobalKeybinds(widget, ctx)
		if fc.ShouldTrackAncestor(widget) {
			fc.PushAncestor(widget)
			ancestorsPushed = true
		}
	}
	if ancestorsPushed {
		defer fc.PopAncestor()
	}

	childWidgets := extractChildren(node.widget)
	quiet := walkQuietWidget(widget, node)
	if !rebuild {
		// The retained build still supplies the same children and wrappers.
		// Walk them to reach dirty descendants and recollect focus scopes.
		// A quiet subtree with nothing to rebuild has neither, so a long list
		// of plain rows costs nothing here.
		for i, childWidget := range childWidgets {
			previous := node.children[i]
			if previous.walkQuiet && previous.subtreeDirtyLevel() < DirtyBuild {
				continue
			}
			child := r.buildRetainedNode(previous, childWidget, ctx.PushChild(i), fc, false)
			if child != previous {
				// A child replaced itself beneath this clean node. Connect it so
				// its signal changes reach the renderer, and mark the ancestors
				// changed so none reuses a layout computed for the old child.
				child.parent = node
				node.shownValid = false
				for ancestor := node; ancestor != nil; ancestor = ancestor.parent {
					ancestor.setDirtySubtree(DirtyBuild)
				}
			}
			node.children[i] = child
			quiet = quiet && child.walkQuiet
		}
		node.walkQuiet = quiet
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
		quiet = quiet && childNode.walkQuiet
	}

	for _, child := range oldChildren {
		child.dispose()
	}

	node.children = children
	node.walkQuiet = quiet
	return node
}

// walkQuietWidget reports whether a node adds nothing of its own to what a
// frame collects while walking the tree, and cannot change its identity
// without being rebuilt. Such subtrees need no walk until something in them
// rebuilds.
func walkQuietWidget(widget Widget, node *widgetNode) bool {
	if len(node.floats) > 0 {
		return false
	}
	switch widget.(type) {
	case Focusable, FocusTrapper, globalKeybindProvider:
		return false
	case Identifiable:
		// A pointer widget can change its own ID (see buildRetainedNode).
		return reflect.ValueOf(widget).Kind() != reflect.Pointer
	}
	return true
}

// needsOwnNode reports whether a widget returned from Build is a composite that
// must be built itself. Widgets that lay out or render themselves are used as
// the node's output directly, unless they take focus or keys: focus and key
// dispatch go to a node's source widget, so a TextInput returned straight from
// a component's Build needs a node of its own to be focusable. A widget
// returning its own type is treated as final, so self-returning widgets aren't
// wrapped again and again.
func needsOwnNode(built, source Widget) bool {
	if reflect.TypeOf(built) == reflect.TypeOf(source) {
		return false
	}
	switch built.(type) {
	case Focusable, KeyHandler, KeybindProvider:
		return true
	case Renderable, LayoutNodeBuilder, ContainerLayoutBuilder, ChildProvider:
		return false
	}
	return true
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
	r.layoutEpoch++
	computed := r.newRetainedLayoutNode(node).ComputeLayout(constraints)
	r.assignComputedLayout(node, &computed)
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
			raw := builder.BuildContainerLayoutNode(node.buildContext, children)
			// Record what patchLayoutCache must check is unchanged.
			node.boxOnlyLayout = boxOnlyLayout(raw)
			if node.boxOnlyLayout {
				widgets := extractChildren(node.widget)
				for i, child := range node.children {
					child.parentDimsKnown = i < len(widgets)
					if child.parentDimsKnown {
						child.parentDims = GetWidgetDimensionSet(widgets[i])
					}
				}
			}
			return raw
		}
		node.boxOnlyLayout = false
		if builder, ok := node.widget.(LayoutNodeBuilder); ok {
			return builder.BuildLayoutNode(node.buildContext)
		}
		return buildFallbackLayoutNode(node.widget, node.buildContext)
	})
}

func (r *Renderer) assignComputedLayout(node *widgetNode, computed *layout.ComputedLayout) {
	if node == nil {
		return
	}
	// A clean subtree handed back the very result it was last assigned (a
	// cache hit returns the same Children slice) has nothing to update.
	// Paint-only changes beneath it can't have changed any layout in it.
	sameSlice := sameChildren(computed.Children, node.layout.Children)
	if node.subtreeDirtyLevel() < DirtyLayout && computed.Box == node.layout.Box && sameSlice {
		node.layoutReused = true
		return
	}
	node.layoutReused = false
	node.layout = *computed
	if !sameSlice {
		node.stackedAxis = stackedAxis(node, computed.Children)
	}
	r.lastAssignCount++
	if observer, ok := node.widget.(LayoutObserver); ok {
		withSignalRead(node, readPhaseLayout, func() struct{} {
			observer.OnLayout(node.buildContext, LayoutMetrics{layout: *computed})
			return struct{}{}
		})
	}
	node.updateIntrinsicCache()
	limit := min(len(node.children), len(computed.Children))
	for i := 0; i < limit; i++ {
		child := node.children[i]
		if sameSlice && node.stackedAxis != stackedNone && child.subtreeDirtyLevel() >= DirtyLayout {
			// The same layout, patched for changed children with unchanged
			// boxes (see patchLayoutCache), keeps its stacking unless a
			// changed child became a Stack, which can draw beyond its box.
			if _, isStack := child.widget.(Stack); isStack {
				node.stackedAxis = stackedNone
			}
		}
		r.assignComputedLayout(child, &computed.Children[i].Layout)
	}
	for i := limit; i < len(node.children); i++ {
		node.children[i].layoutReused = false
		node.children[i].layout = layout.ComputedLayout{}
		node.children[i].updateIntrinsicCache()
	}
	if scrollable, ok := node.widget.(Scrollable); ok {
		box := scrollable.resolvedScrollBox(computed.Box)
		if box != computed.Box {
			computed.Box = box
			node.layout.Box = box
			// Cached measurements captured the offset before child observers ran.
			for ancestor := node; ancestor != nil; ancestor = ancestor.parent {
				ancestor.layoutCache = ancestor.layoutCache[:0]
			}
		}
	}
}

func (r *Renderer) paintRetainedNode(ctx *RenderContext, node *widgetNode, screenX, screenY int, damage Rect, partial bool, recordRegistry bool) Rect {
	if node == nil {
		return Rect{}
	}
	if partial && !node.subtreeBounds.Intersects(damage) {
		return Rect{}
	}

	box := &node.layout.Box
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

	// Outside a Stack, a node and everything beneath it draw only within its
	// border box. If that is entirely out of view (a row scrolled out of a
	// long list, say) there is nothing to draw, measure or hit-test, so its
	// subtree is not visited: scrolling costs what is on screen, not what
	// could be scrolled to. Its empty subtreeBounds tells later frames that
	// nothing beneath it is visible (see collectDamageRects).
	if !partial && !nodeBounds.Intersects(ctx.visible) {
		if _, isStack := node.widget.(Stack); !isStack {
			if r.geometryOnly {
				r.recordReflowDamage(node, nodeBounds, Rect{}, ctx.visible)
			}
			node.registered = nil
			node.prevBox = *box
			node.bounds = nodeBounds
			node.subtreeBounds = Rect{}
			return Rect{}
		}
	}

	selfCtx := *ctx
	selfCtx.currentEventID = node.eventID
	ctx = &selfCtx

	if r.geometryOnly {
		// A clean subtree with the same layout in the same place is exactly as
		// it was: replay its hit-test entries instead of walking it.
		if recordRegistry && node.layoutReused && node.subtreeDirtyLevel() == DirtyNone && node.bounds == nodeBounds && node.hitClip == ctx.clip && node.registered != nil {
			start := len(r.widgetRegistry.entries)
			r.widgetRegistry.appendEntries(node.registered)
			node.registered = r.registeredSince(start)
			return node.subtreeBounds
		}

		// Measure only: record where everything is without drawing.
		r.lastMeasureCount++
		registryStart := len(r.widgetRegistry.entries)
		if recordRegistry {
			r.recordRegistry(node, nodeBounds, ctx.clip)
		}
		subtreeBounds := nodeBounds
		// Measuring needs no style: it only affects how children are drawn.
		r.forEachChildContext(ctx, node, Style{}, partial, absBorderX, absBorderY, absContentX, absContentY, func(childCtx *RenderContext, child *widgetNode, x, y int) {
			subtreeBounds = subtreeBounds.Union(r.paintRetainedNode(childCtx, child, x, y, damage, partial, recordRegistry))
		})
		subtreeBounds = subtreeBounds.Intersect(ctx.visible)
		if scrollable, ok := node.widget.(Scrollable); ok && scrollable.State != nil {
			usableBox := box.UsableContentBox()
			scrollable.State.updateHorizontalLayout(usableBox.Width, box.VirtualWidth)
			scrollable.State.updateLayout(usableBox.Height, box.VirtualHeight)
		}
		r.recordReflowDamage(node, nodeBounds, subtreeBounds, ctx.visible)
		if recordRegistry {
			node.registered = r.registeredSince(registryStart)
			node.hitClip = ctx.clip
		}
		node.prevBox = *box
		node.bounds = nodeBounds
		node.subtreeBounds = subtreeBounds
		return subtreeBounds
	}

	r.lastPaintCount++
	node.clearDependenciesForPhase(readPhasePaint)

	// A node out of view has nothing to show, and paints (and subscribes)
	// again once it comes into view.
	if underlay, ok := node.widget.(underlayPainter); ok && underlay.hasUnderlay() && nodeBounds.Intersects(ctx.visible) {
		underlayCtx := ctx.SubContext(absBorderX, absBorderY, box.Width, box.Height)
		withSignalRead(node, readPhasePaint, func() struct{} {
			underlay.paintUnderlay(underlayCtx)
			return struct{}{}
		})
	}

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

	if renderable, ok := node.widget.(Renderable); ok {
		contentCtx := ctx.SubContext(absContentX, absContentY, box.ContentWidth(), box.ContentHeight())
		contentCtx.imageOwner = node
		node.imageSlot = 0
		contentCtx.imageSlot = &node.imageSlot
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
		r.recordRegistry(node, nodeBounds, ctx.clip)
	}

	subtreeBounds := nodeBounds
	r.forEachChildContext(ctx, node, style, partial, absBorderX, absBorderY, absContentX, absContentY, func(childCtx *RenderContext, child *widgetNode, x, y int) {
		childBounds := r.paintRetainedNode(childCtx, child, x, y, damage, partial, recordRegistry)
		if childBounds.IsEmpty() && partial {
			// Skipped outside the damage; its recorded area is still current.
			childBounds = child.subtreeBounds
		}
		subtreeBounds = subtreeBounds.Union(childBounds)
	})
	subtreeBounds = subtreeBounds.Intersect(ctx.visible)

	if scrollable, ok := node.widget.(Scrollable); ok {
		if scrollable.State != nil {
			usableBox := box.UsableContentBox()
			scrollable.State.updateHorizontalLayout(usableBox.Width, box.VirtualWidth)
			scrollable.State.updateLayout(usableBox.Height, box.VirtualHeight)
		}

		if box.IsScrollableY() && !scrollable.DisableScroll {
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
			// The scrollbar's reads (the exact scroll position, and focus for
			// the thumb colour) are this node's paint dependencies, so changing
			// them repaints just the scrollbar (see scrollbarDamage).
			withSignalRead(node, readPhasePaint, func() struct{} {
				focused := scrollable.IsFocusable() && ctx.focusManager != nil && ctx.IsFocused(node.widget)
				scrollable.renderScrollbar(scrollbarCtx, box.ScrollOffsetY, focused)
				return struct{}{}
			})
		}
	}

	if recordRegistry {
		node.registered = r.registeredSince(registryStart)
		node.hitClip = ctx.clip
	}
	if !partial {
		node.prevBox = *box
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
//
// Of children stacked along an axis (see stackedAxis), it visits only those
// that can be on screen or were last time: the rest are out of view and
// were left so, which is all that visiting them would record.
func (r *Renderer) forEachChildContext(ctx *RenderContext, node *widgetNode, style Style, partial bool, absBorderX, absBorderY, absContentX, absContentY int, visit func(*RenderContext, *widgetNode, int, int)) {
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

	positions := node.layout.Children
	n := min(len(node.children), len(positions))
	lo, hi := 0, n
	if node.stackedAxis != stackedNone && node.shownValid {
		// Children outside [shownLo, shownHi) show nothing, so a partial
		// repaint has nothing else to redraw.
		lo, hi = min(node.shownLo, n), min(node.shownHi, n)
		if !partial {
			visibleLo, visibleHi := stackedVisibleRange(positions[:n], node.stackedAxis, childClipCtx.X, childClipCtx.Y, childClipCtx.visible)
			if visibleLo < visibleHi {
				if lo >= hi {
					lo, hi = visibleLo, visibleHi
				} else {
					lo, hi = min(lo, visibleLo), max(hi, visibleHi)
				}
			}
		}
	}
	shownLo, shownHi := n, 0
	for i := lo; i < hi; i++ {
		child := node.children[i]
		pos := positions[i]
		// Layouts position a child's border box, with its margin already
		// applied. Painting starts from the margin box and applies the margin
		// itself, so step back to the margin box origin.
		margin := pos.Layout.Box.Margin
		visit(childClipCtx, child, pos.X-margin.Left, pos.Y-margin.Top)
		if !child.subtreeBounds.IsEmpty() {
			shownLo, shownHi = min(shownLo, i), i+1
		}
	}
	if !partial {
		// Every child outside what was visited shows nothing.
		if shownLo >= shownHi {
			shownLo, shownHi = 0, 0
		}
		node.shownLo, node.shownHi, node.shownValid = shownLo, shownHi, true
	}
}

// Axes along which a node's children are stacked (see stackedAxis).
const (
	stackedNone int8 = iota
	stackedVertical
	stackedHorizontal
)

// minStackedChildren is how many children make finding the visible ones by
// binary search worthwhile.
const minStackedChildren = 32

// stackedAxis reports an axis along which a node's laid-out children are
// stacked: they start in non-decreasing order, and those starting at the same
// place form a group that everything before it ends ahead of. A Column's or
// Row's children are stacked, as are a table's cells, a row of them per
// group. The ones in a given area can then be found by binary search. A Stack
// child can draw beyond its box, so it rules this out.
func stackedAxis(node *widgetNode, positions []layout.PositionedChild) int8 {
	n := min(len(node.children), len(positions))
	if n < minStackedChildren {
		return stackedNone
	}
	for i := 0; i < n; i++ {
		if _, isStack := node.children[i].widget.(Stack); isStack {
			return stackedNone
		}
	}
	for _, axis := range []int8{stackedVertical, stackedHorizontal} {
		if stackedAlong(positions[:n], axis) {
			return axis
		}
	}
	return stackedNone
}

func stackedAlong(positions []layout.PositionedChild, axis int8) bool {
	groupStart, endBefore, maxEnd := math.MinInt, math.MinInt, math.MinInt
	for i := range positions {
		start, end := stackedSpan(positions, i, axis)
		if end < start {
			return false
		}
		switch {
		case start < groupStart:
			return false
		case start > groupStart:
			// A new group: everything before it must end ahead of it.
			endBefore = maxEnd
			if endBefore > start {
				return false
			}
			groupStart = start
		}
		maxEnd = max(maxEnd, end)
	}
	return true
}

// stackedSpan returns where a child's border box starts and ends on an axis.
func stackedSpan(positions []layout.PositionedChild, i int, axis int8) (start, end int) {
	pos := &positions[i]
	if axis == stackedHorizontal {
		return pos.X, pos.X + pos.Layout.Box.Width
	}
	return pos.Y, pos.Y + pos.Layout.Box.Height
}

// stackedVisibleRange returns the children of a node stacked along axis
// whose border boxes, placed from originX/originY, may intersect visible; all
// others don't.
func stackedVisibleRange(positions []layout.PositionedChild, axis int8, originX, originY int, visible Rect) (lo, hi int) {
	if visible.IsEmpty() {
		return 0, 0
	}
	viewStart, viewEnd := visible.Y-originY, visible.Y+visible.Height-originY
	if axis == stackedHorizontal {
		viewStart, viewEnd = visible.X-originX, visible.X+visible.Width-originX
	}
	start := func(i int) int {
		s, _ := stackedSpan(positions, i, axis)
		return s
	}
	// Groups before the last one starting at or ahead of the view end ahead
	// of it, so they can't be visible.
	if k := sort.Search(len(positions), func(i int) bool { return start(i) > viewStart }); k > 0 {
		groupStart := start(k - 1)
		lo = sort.Search(len(positions), func(i int) bool { return start(i) >= groupStart })
	}
	hi = sort.Search(len(positions), func(i int) bool { return start(i) >= viewEnd })
	return lo, max(lo, hi)
}

// recordRegistry records the node as a hit target. clip is the area it may
// draw in, so only the part of it actually on screen receives pointer events.
func (r *Renderer) recordRegistry(node *widgetNode, bounds, clip Rect) {
	eventWidget := node.eventWidget
	if eventWidget == nil {
		eventWidget = node.widget
	}
	r.widgetRegistry.recordTree(node.widget, eventWidget, node.eventID, node.buildContext.hoverScope+node.autoID, bounds, bounds.Intersect(clip), node.buildContext.IsDisabled())
	if node.parent != nil {
		r.widgetRegistry.entries[len(r.widgetRegistry.entries)-1].parentID = node.parent.eventID
	}
}

// recordReflowDamage marks what a reflow frame must repaint for this node:
// its old and new subtree if it was invalidated, otherwise its own area if it
// moved, resized, or its box changed (for example a scrollbar's extent).
//
// Nodes are only rebuilt, added or removed beneath an invalidated ancestor,
// whose subtree damage already covers them.
//
// Damage is clipped to what the node can show, so scrolling doesn't repaint
// around the viewport where scrolled-out content would lie. If that visible
// area itself moved, the ancestor whose box changed damages the old one.
func (r *Renderer) recordReflowDamage(node *widgetNode, bounds, subtreeBounds, visible Rect) {
	add := func(rect Rect) {
		rect = rect.Intersect(visible)
		if !rect.IsEmpty() {
			r.reflowDamage = append(r.reflowDamage, rect)
		}
	}
	if node.dirtyLevel() == DirtyPaint && node.bounds == bounds && node.prevBox == node.layout.Box {
		if rect, ok := scrollbarDamage(node); ok {
			add(rect)
			return
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
	ctx := NewRenderContext(r.terminal, r.width, r.height, r.focusCollector, r.focusManager, buildCtx, r.widgetRegistry)
	r.reflowDamage = r.reflowDamage[:0]
	r.floatsChanged = false
	r.geometryOnly = true
	r.paintRetainedNode(ctx, r.rootNode, 0, 0, Rect{}, false, true)
	r.placeFloats(ctx, buildCtx, true)
	r.geometryOnly = false

	fullScreen := Rect{Width: r.width, Height: r.height}
	// A topmost overlay (jump mode's labels) is drawn from where everything
	// beneath it is, so any change beneath can move what it shows.
	if r.floatsChanged || (len(r.reflowDamage) > 0 && r.hasTopmostFloat()) {
		r.reflowDamage = append(r.reflowDamage, fullScreen)
	}
	rects := coalesceDamage(r.reflowDamage, fullScreen)
	r.lastDamagedRects = rects
	for _, rect := range rects {
		r.clearRect(rect)
		paintCtx := NewRenderContext(r.terminal, r.width, r.height, r.focusCollector, r.focusManager, buildCtx, r.widgetRegistry)
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
// with the last frame's: the same number, with the same backdrops.
func (r *Renderer) floatSetMatches() bool {
	if len(r.retainedFloats) != r.floatCollector.Len() {
		return false
	}
	for i, retained := range r.retainedFloats {
		if !sameFloatBackdrop(retained.entry.Config, r.floatCollector.entries[i].Config) {
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

	// Only the topmost modal pulls focus into itself. If every open modal did,
	// two of them would take focus from each other on every frame.
	topModal := -1
	for i, entry := range r.floatCollector.entries {
		if entry.Config.Modal {
			topModal = i
		}
	}

	for i := 0; i < len(r.floatCollector.entries); i++ {
		if r.floatCollector.deferTopmost(i) {
			i--
			continue
		}
		entry := r.floatCollector.entries[i]
		focusableCountBefore := r.focusCollector.Len()
		child := entry.Child
		geometryChanged := false
		if entry.BuildChild != nil {
			entry.geometry = r.floatGeometry(entry.Config)
			geometryChanged = i >= len(oldFloats) || oldFloats[i].entry.geometry != entry.geometry
			child = floatChildBuilder{build: entry.BuildChild, geometry: entry.geometry}
		}
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
		floatCtx := buildCtx
		floatCtx.hoverScope = fmt.Sprintf("float:%d/", i)
		if entry.Config.hoverScope != "" {
			floatCtx.hoverScope = entry.Config.hoverScope
		}
		floatRoot := r.buildRetainedNode(oldRoot, child, floatCtx, r.focusCollector, !measure || entry.fresh || geometryChanged)
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

			// A pending request for a widget inside the modal also counts:
			// it was made while opening the modal, or after removing the
			// focused widget, and should win over the first focusable.
			focusedID := r.focusManager.FocusedID()
			alreadyInside := i != topModal
			for _, fe := range r.focusCollector.Focusables()[focusableCountBefore:] {
				if fe.ID == focusedID || (pendingFocusID != "" && fe.ID == pendingFocusID) {
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

		registryStart := len(r.widgetRegistry.entries)
		r.paintRetainedNode(ctx, floatRoot, x, y, Rect{}, false, true)
		r.retainedFloats = append(r.retainedFloats, retainedFloat{
			entry:         entry,
			root:          floatRoot,
			registryStart: registryStart,
			registryEnd:   len(r.widgetRegistry.entries),
		})
	}

	for i := len(r.retainedFloats); i < len(oldFloats); i++ {
		if oldFloats[i].root != nil {
			oldFloats[i].root.dispose()
		}
	}
	if measure && !sameFloatSet(oldFloats, r.retainedFloats) {
		// Nested overlays (registered by overlay content) appeared,
		// disappeared or changed backdrop.
		r.floatsChanged = true
	}
}

func sameFloatSet(a, b []retainedFloat) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameFloatBackdrop(a[i].entry.Config, b[i].entry.Config) {
			return false
		}
	}
	return true
}

// A modal backdrop covers the whole screen, so its changes cannot be repaired
// by repainting just the overlay content's damage. Non-modal colors are unused.
func sameFloatBackdrop(a, b FloatConfig) bool {
	return a.Modal == b.Modal && (!a.Modal || a.BackdropColor == b.BackdropColor)
}

func (r *Renderer) hasTopmostFloat() bool {
	for _, floatNode := range r.retainedFloats {
		if floatNode.entry.topmost {
			return true
		}
	}
	return false
}

// captureKey offers a key press to the top overlay, if it captures keys.
// It reports whether the key was taken.
func (r *Renderer) captureKey(event KeyEvent) bool {
	top := r.TopFloat()
	return top != nil && top.captureKey != nil && top.captureKey(event)
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

// collectDamageRects returns the areas that paint-dirty nodes cover, and
// whether any were found. A dirty node that isn't visible (for example one
// scrolled out of view) has nothing to repaint.
func (r *Renderer) collectDamageRects() (rects []Rect, found bool) {
	appendDirty := func(node *widgetNode) {}
	appendDirty = func(node *widgetNode) {
		if node == nil {
			return
		}
		r.lastScanCount++
		if node.dirtyLevel() == DirtyPaint {
			found = true
			rect, ok := scrollbarDamage(node)
			if !ok {
				rect = node.subtreeBounds
			}
			if !rect.IsEmpty() {
				rects = append(rects, rect)
			}
			if !ok {
				// This repaint includes the whole visible subtree, so additional
				// damage from its descendants would repaint the same cells again.
				// A scrollbar-only repaint still needs to find dirty content.
				return
			}
		}
		for _, child := range node.children {
			// Dirty nodes only lie beneath ancestors whose subtree flag is set.
			if child.subtreeDirtyLevel() == DirtyNone {
				continue
			}
			if child.subtreeBounds.IsEmpty() {
				// Nothing beneath it is visible, and painting may have skipped
				// it, leaving its descendants' recorded areas out of date.
				found = true
				continue
			}
			appendDirty(child)
		}
	}
	appendDirty(r.rootNode)
	for _, floatNode := range r.retainedFloats {
		appendDirty(floatNode.root)
	}
	return rects, found
}

// scrollbarDamage returns the scrollbar column of a Scrollable node, which is
// all that its own paint-phase changes can alter: it only reads signals while
// drawing its scrollbar. Moving the thumb by less than a line therefore
// repaints one column instead of everything the Scrollable contains.
func scrollbarDamage(node *widgetNode) (Rect, bool) {
	if _, ok := node.widget.(Scrollable); !ok {
		return Rect{}, false
	}
	box := node.layout.Box
	if !box.IsScrollableY() || box.ContentWidth() <= 0 {
		return Rect{}, false
	}
	borderX, borderY := box.BorderOrigin()
	contentX, contentY := box.ContentOrigin()
	column := Rect{
		X:      node.bounds.X + contentX - borderX + box.ContentWidth() - 1,
		Y:      node.bounds.Y + contentY - borderY,
		Width:  1,
		Height: box.ContentHeight(),
	}
	// The recorded subtree area is clipped to what is visible.
	return column.Intersect(node.subtreeBounds), true
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
