package terma

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// bgAt is the background colour of screen cell (x, y).
func bgAt(p *Pilot, x, y int) Color {
	p.t.Helper()
	cell := p.Buffer().CellAt(x, y)
	require.NotNil(p.t, cell)
	return FromANSI(cell.Style.Bg)
}

// drag presses the left button at (x0, y0), moves to (x1, y1) and releases
// there.
func drag(p *Pilot, x0, y0, x1, y1 int) {
	p.t.Helper()
	p.MouseDown(x0, y0, uv.MouseLeft, 0)
	p.MouseMove(x1, y1)
	p.MouseUp(x1, y1, uv.MouseLeft, 0)
}

func numberedItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("Item %02d", i)
	}
	return items
}
