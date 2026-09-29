package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hoverRegionProbe struct {
	events []HoverEvent
	clicks int
}

func (p *hoverRegionProbe) WidgetID() string { return "region-owner" }

func (p *hoverRegionProbe) Build(BuildContext) Widget {
	return Text{Content: "owner", Style: Style{Width: Cells(10), Height: Cells(3)}}
}

func (p *hoverRegionProbe) OnClick(MouseEvent) { p.clicks++ }

func (p *hoverRegionProbe) hoverRegionAt(entry *WidgetEntry, x, y int) *WidgetEntry {
	bounds := Rect{X: entry.Bounds.X + 6, Y: entry.Bounds.Y + 1, Width: 2, Height: 1}
	if !bounds.Contains(x, y) {
		return nil
	}
	region := *entry
	region.ID = entry.ID + "-region"
	region.Bounds = bounds
	region.Visible = bounds.Intersect(entry.Visible)
	region.EventWidget = Text{ID: region.ID, Hover: func(event HoverEvent) {
		p.events = append(p.events, event)
	}}
	return &region
}

func TestHoverRegion_TransitionsWithinOwnerAndPreservesClicks(t *testing.T) {
	probe := &hoverRegionProbe{}
	scene := newClickScene(t, probe, 12, 4)
	scene.hover(2, 1)
	assert.Empty(t, probe.events)

	scene.hover(7, 1)
	require.Len(t, probe.events, 1)
	assert.Equal(t, HoverEnter, probe.events[0].Type)
	assert.Equal(t, "region-owner-region", probe.events[0].WidgetID)
	assert.Equal(t, 1, probe.events[0].LocalX)
	assert.Zero(t, probe.events[0].LocalY)

	scene.hover(6, 1)
	assert.Len(t, probe.events, 1, "motion within a region does not enter it again")
	scene.click(7, 1, 0)
	assert.Equal(t, 1, probe.clicks, "the original widget still receives clicks")

	scene.hover(8, 1)
	require.Len(t, probe.events, 2)
	assert.Equal(t, HoverLeave, probe.events[1].Type)
}

func TestHoverRegion_DisabledOwnerDoesNotEnterRegion(t *testing.T) {
	probe := &hoverRegionProbe{}
	scene := newClickScene(t, DisabledWhen(true, probe), 12, 4)
	scene.hover(7, 1)
	scene.click(7, 1, 0)
	assert.Empty(t, probe.events)
	assert.Zero(t, probe.clicks)
}
