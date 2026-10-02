package terma

import (
	"math"
	"reflect"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/darrenburns/terma/layout"
)

// Draggable lifts Child under the pointer after a left-button movement of one
// cell. ID must be stable and unique. HandleID optionally restricts the gesture
// to one descendant. Without a handle, focusable descendants keep their input.
type Draggable[T any] struct {
	ID       string
	Payload  T
	Child    Widget
	HandleID string
	// Shadow defaults to a soft dark shadow. A transparent color disables it.
	Shadow   *FloatShadow
	behavior dragBehavior
}

func (d Draggable[T]) WidgetID() string          { return d.ID }
func (d Draggable[T]) Build(BuildContext) Widget { return passThrough{child: d.Child} }
func (d Draggable[T]) GetContentDimensions() (Dimension, Dimension) {
	return (passThrough{child: d.Child}).GetContentDimensions()
}

type dragPayload[T any] struct{ value T }

func (d Draggable[T]) dragSource() dragSource {
	shadow := FloatShadow{Color: RGBA(0, 0, 0, .4), Offset: Offset{X: 1, Y: 1}, BlurRadius: 1}
	if d.Shadow != nil {
		shadow = *d.Shadow
	}
	return dragSource{payload: dragPayload[T]{d.Payload}, handleID: d.HandleID, shadow: shadow, behavior: d.behavior}
}

// DropTarget receives payloads from Draggable with the same type parameter.
// Accept must have no side effects. OnDrop runs once on a valid release; it
// owns any application-state changes, including removing or moving the source.
type DropTarget[T any] struct {
	ID     string
	Child  Widget
	Accept func(T) bool
	OnDrop func(T)
}

func (d DropTarget[T]) WidgetID() string          { return d.ID }
func (d DropTarget[T]) Build(BuildContext) Widget { return passThrough{child: d.Child} }
func (d DropTarget[T]) GetContentDimensions() (Dimension, Dimension) {
	return (passThrough{child: d.Child}).GetContentDimensions()
}
func (d DropTarget[T]) acceptsDrop(payload any) bool {
	value, ok := payload.(dragPayload[T])
	return ok && d.OnDrop != nil && (d.Accept == nil || d.Accept(value.value))
}
func (d DropTarget[T]) receiveDrop(payload any) { d.OnDrop(payload.(dragPayload[T]).value) }

type dragSourceProvider interface{ dragSource() dragSource }
type dropReceiver interface {
	acceptsDrop(any) bool
	receiveDrop(any)
}
type dragSource struct {
	payload  any
	handleID string
	shadow   FloatShadow
	behavior dragBehavior
}

type dragBehavior interface {
	arm(*dragSession) bool
	valid() bool
	move(*dragSession)
	finish(bool)
}

type dragPhase uint8

const (
	dragArmed dragPhase = iota
	dragging
)

type dragSession struct {
	phase          dragPhase
	source         *widgetNode
	spec           dragSource
	startX, startY float64
	x, y           float64
	grabX, grabY   float64
	frozen         map[*widgetNode]layout.ComputedLayout
	modalScope     dragModalScope
	paintBounds    Rect
	slotBounds     Rect
	painted        bool
	context        *RenderContext
}

func (d *dragSession) bounds() Rect {
	box := d.frozen[d.source].Box
	return Rect{X: int(math.Floor(d.x - d.grabX)), Y: int(math.Floor(d.y - d.grabY)), Width: box.Width, Height: box.Height}
}

func copyDragLayout(value layout.ComputedLayout) layout.ComputedLayout {
	value.Children = append([]layout.PositionedChild(nil), value.Children...)
	for i := range value.Children {
		value.Children[i].Layout = copyDragLayout(value.Children[i].Layout)
	}
	return value
}

func (m *mouseRouter) armDrag(entry *WidgetEntry, ev uv.MouseClickEvent, subX, subY float64) bool {
	if ev.Button != uv.MouseLeft {
		return false
	}
	var path []*WidgetEntry
	for current := entry; current != nil; current = m.renderer.WidgetByID(current.parentID) {
		path = append(path, current)
		provider, ok := current.EventWidget.(dragSourceProvider)
		if !ok || current.node == nil || current.Disabled {
			continue
		}
		spec := provider.dragSource()
		if spec.handleID != "" {
			matched := false
			for _, child := range path {
				matched = matched || child.ID == spec.handleID
			}
			if !matched {
				return false
			}
		} else {
			for _, child := range path[:len(path)-1] {
				if focusable, ok := child.EventWidget.(Focusable); ok && focusable.IsFocusable() {
					return false
				}
				if _, ok := child.EventWidget.(pointerOwner); ok {
					return false
				}
			}
		}
		x, y := float64(ev.X)+subX, float64(ev.Y)+subY
		d := &dragSession{source: current.node, spec: spec, startX: x, startY: y, x: x, y: y,
			grabX: x - float64(current.Bounds.X), grabY: y - float64(current.Bounds.Y),
			frozen:     map[*widgetNode]layout.ComputedLayout{current.node: copyDragLayout(current.node.layout)},
			slotBounds: current.Bounds, modalScope: m.renderer.dragModalScope()}
		if spec.behavior != nil && !spec.behavior.arm(d) {
			return false
		}
		m.renderer.drag = d
		m.renderer.dragDirty = true
		return true
	}
	return false
}

type dragModalScope struct {
	owner *widgetNode
	count int
}

func (r *Renderer) dragModalScope() dragModalScope {
	var scope dragModalScope
	for _, entry := range r.floatCollector.entries {
		if entry.Config.Modal {
			scope.owner = entry.owner
			scope.count++
		}
	}
	return scope
}

func (m *mouseRouter) moveDrag(x, y int, subX, subY float64) bool {
	r := m.renderer
	d := r.drag
	if d == nil {
		return false
	}
	if d.spec.behavior != nil && !d.spec.behavior.valid() {
		r.cancelDrag()
		return true
	}
	d.x, d.y = float64(x)+subX, float64(y)+subY
	if d.phase == dragArmed {
		if math.Abs(d.x-d.startX) < 1 && math.Abs(d.y-d.startY) < 1 {
			return false
		}
		d.phase = dragging
		r.dragDamage = append(r.dragDamage, d.slotBounds)
	}
	if d.spec.behavior != nil {
		d.spec.behavior.move(d)
	}
	r.dragDirty = true
	return true
}

func (m *mouseRouter) releaseDrag(ev uv.MouseReleaseEvent, subX, subY float64) {
	r := m.renderer
	m.moveDrag(ev.X, ev.Y, subX, subY)
	d := r.drag
	if d == nil {
		return
	}
	if d.phase != dragging {
		r.cancelDrag()
		return
	}
	var receiver dropReceiver
	if d.spec.behavior == nil {
		entry, consumed := m.target(ev.X, ev.Y, false)
		if !consumed {
			for entry != nil {
				if target, ok := entry.EventWidget.(dropReceiver); ok && !entry.Disabled && target.acceptsDrop(d.spec.payload) {
					receiver = target
					break
				}
				entry = r.WidgetByID(entry.parentID)
			}
		}
	}
	// End the drag before callbacks can remove the source or begin another one.
	r.endDrag(true)
	if receiver != nil {
		receiver.receiveDrop(d.spec.payload)
	}
}

func (r *Renderer) cancelDrag() bool {
	if r.drag == nil {
		return false
	}
	r.endDrag(false)
	return true
}

func (r *Renderer) endDrag(commit bool) {
	d := r.drag
	if d == nil {
		return
	}
	r.drag = nil
	r.dragDirty = true
	r.dragDamage = append(r.dragDamage, d.paintBounds, d.slotBounds)
	for node := range d.frozen {
		invalidateDragLayout(node)
	}
	if d.spec.behavior != nil {
		d.spec.behavior.finish(commit && d.phase == dragging)
	}
}

func invalidateDragLayout(node *widgetNode) {
	node.layoutCache = nil
	node.markDirtyLevel(DirtyLayout)
	for _, child := range node.children {
		invalidateDragLayout(child)
	}
}

func (r *Renderer) reconcileDrag() bool {
	d := r.drag
	if d == nil {
		return false
	}
	found := containsDragNode(r.rootNode, d.source)
	for _, floating := range r.retainedFloats {
		found = found || containsDragNode(floating.root, d.source)
	}
	provider, draggable := d.source.eventWidget.(dragSourceProvider)
	compatible := draggable && reflect.TypeOf(provider.dragSource().payload) == reflect.TypeOf(d.spec.payload)
	if !found || !compatible || d.source.buildContext.IsDisabled() || d.modalScope != r.dragModalScope() || (d.spec.behavior != nil && !d.spec.behavior.valid()) {
		return r.cancelDrag()
	}
	return false
}

func containsDragNode(node, source *widgetNode) bool {
	if node == nil {
		return false
	}
	if node == source {
		return true
	}
	for _, child := range node.children {
		if containsDragNode(child, source) {
			return true
		}
	}
	return false
}

func (r *Renderer) resetDragPaint() {
	if r.drag != nil {
		r.drag.painted = false
	}
}

func (r *Renderer) dragOwnsFloat(owner *widgetNode) bool {
	if r.drag == nil || r.drag.phase != dragging {
		return false
	}
	for node := owner; node != nil; node = node.parent {
		if node == r.drag.source {
			return true
		}
		for _, floating := range r.retainedFloats {
			if floating.entry.dragOwned && floating.root == node {
				return true
			}
		}
	}
	return false
}

func (r *Renderer) paintDrag(ctx *RenderContext, damage Rect, partial, recordRegistry bool) {
	d := r.drag
	if d == nil || d.phase != dragging || d.painted || d.context == nil {
		return
	}
	d.painted = true
	box := d.bounds()
	extent := d.spec.shadow.bounds(box)
	if r.geometryOnly {
		r.reflowDamage = append(r.reflowDamage, r.overlayDamage(d.paintBounds), r.overlayDamage(extent))
	}
	d.paintBounds = extent
	floatCtx := *ctx
	floatCtx.inheritedBgAt = d.context.inheritedBgAt
	if !r.geometryOnly {
		paintFloatShadow(&floatCtx, box, d.spec.shadow)
	}
	r.paintingDrag = true
	margin := d.source.layout.Box.Margin
	r.paintRetainedNode(&floatCtx, d.source, box.X-margin.Left, box.Y-margin.Top, damage, partial, recordRegistry)
	r.paintingDrag = false
}
