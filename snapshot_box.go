package terma

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
)

const (
	boxLeft uint8 = 1 << iota
	boxRight
	boxUp
	boxDown
)

type svgBoxShape struct {
	arms    uint8
	weight  int // 1: light, 2: heavy, 3: double
	rounded bool
	dashes  int
}

// These are the box characters used by borders, dividers and tree branches.
// Drawing them in cell coordinates avoids font-dependent gaps and offsets.
func boxDrawingShape(content string) (svgBoxShape, bool) {
	r, size := utf8.DecodeRuneInString(content)
	if size != len(content) {
		return svgBoxShape{}, false
	}
	groups := [...]struct {
		chars string
		arms  uint8
	}{
		{"─━═", boxLeft | boxRight},
		{"│┃║", boxUp | boxDown},
		{"┌┏╔", boxRight | boxDown},
		{"┐┓╗", boxLeft | boxDown},
		{"└┗╚", boxRight | boxUp},
		{"┘┛╝", boxLeft | boxUp},
		{"├┣╠", boxRight | boxUp | boxDown},
		{"┤┫╣", boxLeft | boxUp | boxDown},
		{"┬┳╦", boxLeft | boxRight | boxDown},
		{"┴┻╩", boxLeft | boxRight | boxUp},
		{"┼╋╬", boxLeft | boxRight | boxUp | boxDown},
		{"╴╸", boxLeft}, {"╶╺", boxRight},
		{"╵╹", boxUp}, {"╷╻", boxDown},
	}
	for _, group := range groups {
		weight := 1
		for _, char := range group.chars {
			if r == char {
				return svgBoxShape{arms: group.arms, weight: weight}, true
			}
			weight++
		}
	}
	switch r {
	case '╭':
		return svgBoxShape{arms: boxRight | boxDown, weight: 1, rounded: true}, true
	case '╮':
		return svgBoxShape{arms: boxLeft | boxDown, weight: 1, rounded: true}, true
	case '╰':
		return svgBoxShape{arms: boxRight | boxUp, weight: 1, rounded: true}, true
	case '╯':
		return svgBoxShape{arms: boxLeft | boxUp, weight: 1, rounded: true}, true
	case '┄', '┅', '┆', '┇', '┈', '┉', '┊', '┋', '╌', '╍', '╎', '╏':
		shape := svgBoxShape{weight: 1, dashes: 3}
		if strings.ContainsRune("┄┅┈┉╌╍", r) {
			shape.arms = boxLeft | boxRight
		} else {
			shape.arms = boxUp | boxDown
		}
		if strings.ContainsRune("┅┇┉┋╍╏", r) {
			shape.weight = 2
		}
		if r >= '┈' && r <= '┋' {
			shape.dashes = 4
		} else if r >= '╌' && r <= '╏' {
			shape.dashes = 2
		}
		return shape, true
	}
	return svgBoxShape{}, false
}

func writeBoxDrawing(sb *strings.Builder, cell *uv.Cell, x, y, w, h float64, pageBg Color) bool {
	shape, ok := boxDrawingShape(cell.Content)
	if !ok || cell.Width != 1 {
		return false
	}
	fg, bg := FromANSI(cell.Style.Fg), FromANSI(cell.Style.Bg)
	if cell.Style.Attrs&uv.AttrReverse != 0 {
		fg, bg = bg, fg
		if !fg.IsSet() {
			fg = pageBg
		}
		if !bg.IsSet() {
			bg = RGB(255, 255, 255)
		}
		fmt.Fprintf(sb, "  <rect x=\"%.4f\" y=\"%.4f\" width=\"%.4f\" height=\"%.4f\" fill=\"%s\"/>\n", x, y, w, h, bg.Hex())
	}
	if !fg.IsSet() {
		fg = RGB(255, 255, 255)
	}
	cx, cy := x+w/2, y+h/2
	stroke := math.Min(w, h) / 8
	if shape.weight == 2 {
		stroke *= 2
	}
	var path strings.Builder
	line := func(x1, y1, x2, y2 float64) {
		fmt.Fprintf(&path, "M %.4f %.4f L %.4f %.4f ", x1, y1, x2, y2)
	}
	if shape.rounded {
		dx, dy := 1.0, 1.0
		if shape.arms&boxLeft != 0 {
			dx = -1
		}
		if shape.arms&boxUp != 0 {
			dy = -1
		}
		radius := math.Min(w, h) / 2
		fmt.Fprintf(&path, "M %.4f %.4f L %.4f %.4f Q %.4f %.4f %.4f %.4f L %.4f %.4f",
			cx+dx*w/2, cy, cx+dx*radius, cy, cx, cy, cx, cy+dy*radius, cx, cy+dy*h/2)
	} else if shape.weight == 3 {
		// Each quadrant joins the corresponding rails. At a missing arm the
		// rails continue through the junction; at a corner they turn together.
		gap := stroke
		for _, dx := range []float64{-1, 1} {
			for _, dy := range []float64{-1, 1} {
				horizontal := boxRight
				if dx < 0 {
					horizontal = boxLeft
				}
				vertical := boxDown
				if dy < 0 {
					vertical = boxUp
				}
				hasH, hasV := shape.arms&horizontal != 0, shape.arms&vertical != 0
				if hasH && hasV {
					fmt.Fprintf(&path, "M %.4f %.4f L %.4f %.4f L %.4f %.4f ", cx+dx*w/2, cy+dy*gap, cx+dx*gap, cy+dy*gap, cx+dx*gap, cy+dy*h/2)
				} else if hasH {
					endX := cx
					if shape.arms&(boxLeft|boxRight) != boxLeft|boxRight {
						endX -= dx * gap
					}
					line(cx+dx*w/2, cy+dy*gap, endX, cy+dy*gap)
				} else if hasV {
					endY := cy
					if shape.arms&(boxUp|boxDown) != boxUp|boxDown {
						endY -= dy * gap
					}
					line(cx+dx*gap, cy+dy*h/2, cx+dx*gap, endY)
				}
			}
		}
	} else if shape.arms == boxLeft|boxRight {
		line(x, cy, x+w, cy)
	} else if shape.arms == boxUp|boxDown {
		line(cx, y, cx, y+h)
	} else {
		if shape.arms&boxLeft != 0 {
			line(x, cy, cx, cy)
		}
		if shape.arms&boxRight != 0 {
			line(cx, cy, x+w, cy)
		}
		if shape.arms&boxUp != 0 {
			line(cx, y, cx, cy)
		}
		if shape.arms&boxDown != 0 {
			line(cx, cy, cx, y+h)
		}
	}
	dashAttr := ""
	shapeAttr := ` shape-rendering="crispEdges"`
	if shape.rounded {
		shapeAttr = "" // Keep curved corners smooth; straight segments snap to pixels.
	}
	if shape.dashes > 0 {
		length := w
		if shape.arms&boxUp != 0 {
			length = h
		}
		dash := length / float64(shape.dashes*2)
		dashAttr = fmt.Sprintf(` stroke-dasharray="%.4f %.4f"`, dash, dash)
	}
	fmt.Fprintf(sb, "  <path d=\"%s\" fill=\"none\" stroke=\"%s\" stroke-width=\"%.4f\"%s%s/>\n", strings.TrimSpace(path.String()), fg.Hex(), stroke, dashAttr, shapeAttr)
	return true
}
