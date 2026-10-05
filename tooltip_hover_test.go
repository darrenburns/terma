package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This child exposes its descendants only through Build, like an app widget.
type tooltipHoverComposite struct{}

func (tooltipHoverComposite) Build(BuildContext) Widget {
	return Row{Children: []Widget{Text{Content: "First"}, Text{Content: "Second"}}}
}

type tooltipHoverCountedChild struct{ builds *int }

func (c tooltipHoverCountedChild) WidgetID() string { return "stable-overlay" }
func (c tooltipHoverCountedChild) Build(BuildContext) Widget {
	*c.builds++
	return Text{Content: "Overlay"}
}

func TestTooltipHover_RetainedOverlayScopeChangeRebuilds(t *testing.T) {
	_, renderer := renderForMouse(Text{Content: "Background"}, 30, 6)
	ctx := NewBuildContext(renderer.focusManager, renderer.focusedSignal, renderer.hoveredSignal, renderer.floatCollector)
	ctx.hoverTarget = renderer.hoverTarget
	ctx.hoverScope = "float:1/"
	builds := 0
	child := tooltipHoverCountedChild{builds: &builds}
	node := renderer.buildRetainedNode(nil, child, ctx, renderer.focusCollector, true)
	require.Equal(t, 1, builds)
	ctx.hoverScope = "float:0/"
	next := renderer.buildRetainedNode(node, child, ctx, renderer.focusCollector, false)
	assert.Same(t, node, next, "widget identity survives the move")
	assert.Equal(t, 2, builds, "captured hover selectors must rebuild in their new scope")
	assert.Equal(t, "float:0/", next.buildContext.hoverScope)
}

func TestTooltipHover_CustomChildUpdatesThroughRetainedRenderer(t *testing.T) {
	root := hoverScreen{Tooltip{Content: "Custom help", Position: TooltipBottom, Child: tooltipHoverComposite{}}}
	p := NewPilot(t, root, 30, 6)
	assert.False(t, p.session.renderer.HasFloats())
	p.MouseMove(2, 0)
	require.True(t, p.session.renderer.HasFloats())
	buildFrames := p.session.renderer.Stats().FullRenderCount
	p.MouseMove(7, 0)
	assert.True(t, p.session.renderer.HasFloats())
	assert.Equal(t, buildFrames, p.session.renderer.Stats().FullRenderCount, "moving within the subtree does not rebuild its tooltip")
	p.MouseMove(25, 5)
	assert.False(t, p.session.renderer.HasFloats())
}

func TestTooltipHover_SubtreePathBoundaries(t *testing.T) {
	ctx := BuildContext{path: []int{0, 1}, hoverTarget: NewSignal(hoverTargetInfo{path: "_auto:0.10"})}
	assert.False(t, ctx.isSubtreeHovered(), "sibling index 10 is not a child of index 1")
	ctx.hoverTarget.Set(hoverTargetInfo{path: "_auto:0.1.0"})
	assert.True(t, ctx.isSubtreeHovered())
	ctx.hoverTarget.Set(hoverTargetInfo{path: "_auto:0.1.0", disabled: true})
	assert.False(t, ctx.isSubtreeHovered())
}

func TestTooltipHover_OverlayDoesNotShareMainTreeIdentity(t *testing.T) {
	root := hoverScreen{Stack{Children: []Widget{
		Tooltip{Content: "Main help", Position: TooltipBottom, Child: Text{Content: "Target"}},
		Floating{Visible: true, Config: FloatConfig{Position: FloatPositionTopLeft, Offset: Offset{Y: 3}, Modal: true}, Child: tooltipHoverComposite{}},
	}}}
	p := NewPilot(t, root, 30, 6)
	require.Len(t, p.session.renderer.floatCollector.entries, 1)
	p.MouseMove(2, 0)
	assert.Len(t, p.session.renderer.floatCollector.entries, 1, "modal blocks underlying tooltip")
	p.MouseMove(2, 3)
	assert.Len(t, p.session.renderer.floatCollector.entries, 1, "unrelated overlay descendants cannot trigger main tree help")
}

func TestTooltipHover_NonFocusableChildEnterLeave(t *testing.T) {
	for _, id := range []string{"target", ""} {
		t.Run(id, func(t *testing.T) {
			tooltip := Tooltip{ID: "help", Content: "Hover help", Position: TooltipBottom, Offset: 1, Child: Text{ID: id, Content: "Target"}}
			p := NewPilot(t, hoverScreen{tooltip}, 30, 6)
			assert.False(t, p.session.renderer.HasFloats())
			p.MouseMove(2, 0)
			require.True(t, p.session.renderer.HasFloats())
			assert.Equal(t, "", p.FocusedID(), "non-focusable hover does not take keyboard focus")
			anchor := p.session.renderer.WidgetByID("help")
			require.NotNil(t, anchor)
			float := p.session.renderer.TopFloat()
			assert.Equal(t, anchor.Bounds.Y+anchor.Bounds.Height+1, float.Y)
			p.MouseMove(25, 5)
			assert.False(t, p.session.renderer.HasFloats())
		})
	}
}

func TestTooltipHover_ClampedOverlayRemainsVisibleUnderPointer(t *testing.T) {
	root := Tooltip{Content: "Hover help", Child: Text{Content: "Target"}}
	p := NewPilot(t, root, 30, 5)
	p.MouseMove(2, 0)
	assert.True(t, p.session.renderer.HasFloats(), "a top tooltip clamped over its trigger must not flicker")
	p.MouseMove(25, 4)
	assert.False(t, p.session.renderer.HasFloats())
}

func TestTooltipHover_CompositeChildIncludesDescendants(t *testing.T) {
	tooltip := Tooltip{Content: "Composite help", Position: TooltipBottom, Child: Row{
		Children: []Widget{Text{ID: "first", Content: "First"}, Text{Content: "Second"}},
	}}
	p := NewPilot(t, hoverScreen{tooltip}, 30, 6)
	p.MouseMove(2, 0)
	assert.True(t, p.session.renderer.HasFloats())
	p.MouseMove(7, 0)
	assert.True(t, p.session.renderer.HasFloats(), "moving between descendants keeps the tooltip visible")
	p.MouseMove(25, 5)
	assert.False(t, p.session.renderer.HasFloats())
}

func TestTooltipHover_FocusKeepsTooltipVisibleAfterPointerLeaves(t *testing.T) {
	for _, id := range []string{"button", ""} {
		t.Run(id, func(t *testing.T) {
			tooltip := Tooltip{Content: "Focused help", Position: TooltipBottom, Child: Button{ID: id, Label: "Target"}}
			p := NewPilot(t, hoverScreen{tooltip}, 30, 6)
			require.NotEmpty(t, p.FocusedID())
			assert.True(t, p.session.renderer.HasFloats())
			p.MouseMove(2, 0)
			p.MouseMove(25, 5)
			assert.True(t, p.session.renderer.HasFloats())
		})
	}
}

func TestTooltipHover_DisabledChildrenDoNotShowHelp(t *testing.T) {
	for _, outside := range []bool{false, true} {
		t.Run(map[bool]string{false: "child-disabled", true: "tooltip-disabled"}[outside], func(t *testing.T) {
			var tooltip Widget = Tooltip{Content: "Disabled help", Position: TooltipBottom, Child: DisabledWhen(true, Text{ID: "target", Content: "Target"})}
			if outside {
				tooltip = DisabledWhen(true, Tooltip{Content: "Disabled help", Position: TooltipBottom, Child: Text{ID: "target", Content: "Target"}})
			}
			p := NewPilot(t, hoverScreen{tooltip}, 30, 6)
			p.MouseMove(2, 0)
			assert.False(t, p.session.renderer.HasFloats())
		})
	}
}
