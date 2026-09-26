package terma

import "github.com/darrenburns/terma/layout"

// passThrough holds one child without any layout of its own: the child gets
// exactly the constraints the wrapper gets, and the wrapper takes the child's
// size. List and table rows use it to give each row its own Build (so only the
// rows whose state changes rebuild) without changing how rows are laid out.
type passThrough struct {
	child Widget
}

func (p passThrough) Build(BuildContext) Widget { return p }

func (p passThrough) ChildWidgets() []Widget { return []Widget{p.child} }

// GetContentDimensions reports the child's declared dimensions, so containers
// that size children by them (flex, percent) treat the wrapper as the child.
func (p passThrough) GetContentDimensions() (width, height Dimension) {
	dims := GetWidgetDimensionSet(p.child)
	return dims.Width, dims.Height
}

func (p passThrough) BuildContainerLayoutNode(_ BuildContext, children []layout.LayoutNode) layout.LayoutNode {
	if len(children) == 0 {
		return &layout.BoxNode{}
	}
	return passThroughLayoutNode{child: children[0]}
}

// BuildLayoutNode supports containers that build layout nodes from widgets
// directly rather than from retained children.
func (p passThrough) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	childCtx := ctx.PushChild(0)
	built := p.child.Build(childCtx)
	if builder, ok := built.(LayoutNodeBuilder); ok {
		return passThroughLayoutNode{child: builder.BuildLayoutNode(childCtx)}
	}
	return passThroughLayoutNode{child: buildFallbackLayoutNode(built, childCtx)}
}

type passThroughLayoutNode struct {
	child layout.LayoutNode
}

func (n passThroughLayoutNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	child := n.child.ComputeLayout(constraints)
	return layout.ComputedLayout{
		Box: layout.BoxModel{Width: child.Box.MarginBoxWidth(), Height: child.Box.MarginBoxHeight()},
		Children: []layout.PositionedChild{{
			X: child.Box.Margin.Left, Y: child.Box.Margin.Top, Layout: child,
		}},
	}
}

func (n passThroughLayoutNode) PreservesWidth() bool {
	preserver, ok := n.child.(layout.SizePreserver)
	return ok && preserver.PreservesWidth()
}

func (n passThroughLayoutNode) PreservesHeight() bool {
	preserver, ok := n.child.(layout.SizePreserver)
	return ok && preserver.PreservesHeight()
}
