package terma

import (
	"sync"
	"sync/atomic"

	"github.com/darrenburns/terma/layout"
)

// Widget is the base interface for all UI elements.
// Leaf widgets (like Text) return themselves from Build().
// Container widgets return a composed widget tree.
type Widget interface {
	Build(ctx BuildContext) Widget
}

// Constraints define the min/max dimensions a widget can occupy.
type Constraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

// Size represents the computed dimensions of a widget.
type Size struct {
	Width, Height int
}

// Layoutable is implemented by widgets that can compute their size
// given constraints and perform layout on their children.
type Layoutable interface {
	Layout(ctx BuildContext, constraints Constraints) Size
}

// Renderable is implemented by widgets that can render themselves.
type Renderable interface {
	Render(ctx *RenderContext)
}

// IntrinsicContentSizer is implemented by widgets that can report their
// content size without rebuilding structural parents. This is used to decide
// whether a paint-only signal change can stay paint-only for auto-sized widgets.
type IntrinsicContentSizer interface {
	ContentWidthHint() int
	ContentHeightHint(width int) int
}

// Dimensioned is implemented by widgets that have explicit dimension preferences.
// This allows parent containers to query child dimensions for fractional layout.
//
// GetContentDimensions returns the content-box dimensions - the space needed for
// the widget's content, NOT including padding or border. The framework automatically
// adds padding and border from the widget's Style to compute the final outer size.
//
// For example, a TextInput might return Cells(1) for height (one line of text content),
// and if it has Style{Padding: EdgeInsetsXY(1,1)}, the final height will be 3 cells.
type Dimensioned interface {
	GetContentDimensions() (width, height Dimension)
}

// Styled is implemented by widgets that have a Style.
// The framework uses this to extract padding and margin for automatic layout.
type Styled interface {
	GetStyle() Style
}

// LayoutNodeBuilder is implemented by widgets that can build a layout node for themselves.
// This enables integration with the new layout system in the layout package.
type LayoutNodeBuilder interface {
	BuildLayoutNode(ctx BuildContext) layout.LayoutNode
}

// ContainerLayoutBuilder is the preferred layout API for container widgets.
// The framework builds child layout nodes first, then passes them to the
// container so it only needs to describe how those children are arranged.
//
// Children are supplied in the widget's logical child order, matching
// ChildProvider/extractChildren ordering.
//
// LayoutNodeBuilder remains supported as a compatibility fallback, especially
// for leaf widgets and older custom containers that still build child layout
// nodes manually inside BuildLayoutNode.
type ContainerLayoutBuilder interface {
	BuildContainerLayoutNode(ctx BuildContext, children []layout.LayoutNode) layout.LayoutNode
}

// GetWidgetDimensionSet returns resolved dimensions for a widget.
// Style dimensions override Dimensioned dimensions when explicitly set.
func GetWidgetDimensionSet(widget Widget) DimensionSet {
	var dims DimensionSet
	if styled, ok := widget.(Styled); ok {
		dims = styled.GetStyle().GetDimensions()
	}
	if dimensioned, ok := widget.(Dimensioned); ok {
		width, height := dimensioned.GetContentDimensions()
		if dims.Width.IsUnset() {
			dims.Width = width
		}
		if dims.Height.IsUnset() {
			dims.Height = height
		}
	}
	return dims
}

// LayoutObserver is implemented by widgets that want access to computed layout data.
// OnLayout is called after layout is computed for the widget, before child render trees are built.
// Use this to read resolved child positions/sizes without re-measuring.
type LayoutObserver interface {
	OnLayout(ctx BuildContext, metrics LayoutMetrics)
}

// ChildProvider exposes a widget's children for render tree construction.
// Implement this for custom containers so computed child layouts are rendered.
type ChildProvider interface {
	ChildWidgets() []Widget
}

type dirtyLevel int32

const (
	DirtyNone dirtyLevel = iota
	DirtyPaint
	DirtyLayout
	DirtyBuild
)

type dependencyMask uint8

const (
	readPhaseNone  dependencyMask = 0
	readPhaseBuild dependencyMask = 1 << iota
	readPhaseLayout
	readPhasePaint
)

type signalDependency interface {
	removeListener(node *widgetNode, mask dependencyMask)
}

type intrinsicSizeCache struct {
	sizer           IntrinsicContentSizer
	autoWidth       bool
	autoHeight      bool
	lastWidthHint   int
	lastHeightHint  int
	lastHeightWidth int
	valid           bool
}

// widgetNode is an internal retained node in the widget tree.
type widgetNode struct {
	parent      *widgetNode
	source      Widget
	widget      Widget
	eventWidget Widget
	children    []*widgetNode

	buildContext BuildContext
	autoID       string
	eventID      string
	identity     string

	layout        layout.ComputedLayout
	bounds        Rect
	subtreeBounds Rect

	dirtySelf    atomic.Int32
	dirtySubtree atomic.Int32

	depsMu sync.Mutex
	deps   map[signalDependency]dependencyMask

	intrinsic intrinsicSizeCache
}

// newWidgetNode creates a new widget node.
func newWidgetNode(widget Widget) *widgetNode {
	node := &widgetNode{
		source:      widget,
		eventWidget: widget,
		widget:      widget,
		deps:        make(map[signalDependency]dependencyMask),
	}
	node.dirtySelf.Store(int32(DirtyBuild))
	node.dirtySubtree.Store(int32(DirtyBuild))
	return node
}

// markDirty marks this node for rebuild.
// Thread-safe: can be called from any goroutine.
func (n *widgetNode) markDirty() {
	n.markDirtyLevel(DirtyBuild)
}

// isDirty returns whether this node needs rebuild.
// Thread-safe.
func (n *widgetNode) isDirty() bool {
	return n.dirtyLevel() != DirtyNone
}

// clearDirty marks this node as clean.
// Thread-safe.
func (n *widgetNode) clearDirty() {
	n.dirtySelf.Store(int32(DirtyNone))
	n.dirtySubtree.Store(int32(DirtyNone))
}

func (n *widgetNode) dirtyLevel() dirtyLevel {
	return dirtyLevel(n.dirtySelf.Load())
}

func (n *widgetNode) subtreeDirtyLevel() dirtyLevel {
	return dirtyLevel(n.dirtySubtree.Load())
}

func (n *widgetNode) setDirtySelf(level dirtyLevel) {
	for {
		current := dirtyLevel(n.dirtySelf.Load())
		if current >= level {
			return
		}
		if n.dirtySelf.CompareAndSwap(int32(current), int32(level)) {
			return
		}
	}
}

func (n *widgetNode) setDirtySubtree(level dirtyLevel) {
	for {
		current := dirtyLevel(n.dirtySubtree.Load())
		if current >= level {
			return
		}
		if n.dirtySubtree.CompareAndSwap(int32(current), int32(level)) {
			return
		}
	}
}

func (n *widgetNode) markDirtyMask(mask dependencyMask) {
	n.markDirtyLevel(dirtyLevelForMask(mask))
}

func (n *widgetNode) markDirtyLevel(level dirtyLevel) {
	if n == nil || level == DirtyNone {
		return
	}
	if level == DirtyPaint && n.intrinsicSizeChanged() {
		level = DirtyLayout
	}
	n.setDirtySelf(level)
	for current := n; current != nil; current = current.parent {
		current.setDirtySubtree(level)
	}
}

func (n *widgetNode) intrinsicSizeChanged() bool {
	cache := n.intrinsic
	if !cache.valid || cache.sizer == nil {
		return false
	}
	if cache.autoWidth && cache.sizer.ContentWidthHint() != cache.lastWidthHint {
		return true
	}
	if cache.autoHeight && cache.sizer.ContentHeightHint(cache.lastHeightWidth) != cache.lastHeightHint {
		return true
	}
	return false
}

func (n *widgetNode) updateIntrinsicCache() {
	cache := intrinsicSizeCache{}
	sizer, ok := n.widget.(IntrinsicContentSizer)
	if !ok {
		n.intrinsic = cache
		return
	}
	dims := GetWidgetDimensionSet(n.widget)
	cache.autoWidth = dims.Width.IsAuto() && !dims.Width.IsUnset()
	cache.autoHeight = dims.Height.IsAuto() && !dims.Height.IsUnset()
	if !cache.autoWidth && !cache.autoHeight {
		n.intrinsic = cache
		return
	}
	cache.sizer = sizer
	cache.valid = true
	if cache.autoWidth {
		cache.lastWidthHint = max(1, sizer.ContentWidthHint())
	}
	if cache.autoHeight {
		cache.lastHeightWidth = max(1, n.layout.Box.ContentWidth())
		cache.lastHeightHint = max(1, sizer.ContentHeightHint(cache.lastHeightWidth))
	}
	n.intrinsic = cache
}

func (n *widgetNode) clearDependenciesForPhase(mask dependencyMask) {
	if n == nil || mask == readPhaseNone {
		return
	}
	n.depsMu.Lock()
	defer n.depsMu.Unlock()
	for dep, currentMask := range n.deps {
		if currentMask&mask == 0 {
			continue
		}
		nextMask := currentMask &^ mask
		dep.removeListener(n, mask)
		if nextMask == 0 {
			delete(n.deps, dep)
			continue
		}
		n.deps[dep] = nextMask
	}
}

func (n *widgetNode) clearAllDependencies() {
	if n == nil {
		return
	}
	n.depsMu.Lock()
	deps := n.deps
	n.deps = make(map[signalDependency]dependencyMask)
	n.depsMu.Unlock()
	for dep, mask := range deps {
		dep.removeListener(n, mask)
	}
}

func (n *widgetNode) dispose() {
	if n == nil {
		return
	}
	for _, child := range n.children {
		child.dispose()
	}
	n.children = nil
	n.clearAllDependencies()
	n.clearDirty()
}

func (n *widgetNode) trackDependency(dep signalDependency, mask dependencyMask) {
	if n == nil || dep == nil || mask == readPhaseNone {
		return
	}
	n.depsMu.Lock()
	defer n.depsMu.Unlock()
	n.deps[dep] |= mask
}

func (n *widgetNode) recomputeDirtySubtree() dirtyLevel {
	if n == nil {
		return DirtyNone
	}
	level := n.dirtyLevel()
	for _, child := range n.children {
		if childLevel := child.recomputeDirtySubtree(); childLevel > level {
			level = childLevel
		}
	}
	n.dirtySubtree.Store(int32(level))
	return level
}

func (n *widgetNode) clearDirtyRecursive() {
	if n == nil {
		return
	}
	n.dirtySelf.Store(int32(DirtyNone))
	for _, child := range n.children {
		child.clearDirtyRecursive()
	}
	n.dirtySubtree.Store(int32(DirtyNone))
}

func dirtyLevelForMask(mask dependencyMask) dirtyLevel {
	switch {
	case mask&readPhaseBuild != 0:
		return DirtyBuild
	case mask&readPhaseLayout != 0:
		return DirtyLayout
	case mask&readPhasePaint != 0:
		return DirtyPaint
	default:
		return DirtyNone
	}
}
