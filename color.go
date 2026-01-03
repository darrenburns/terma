package terma

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Color represents a terminal color with full RGB and HSL support.
// The zero value (Color{}) is transparent/default - inherits from terminal.
type Color struct {
	r, g, b uint8
	set     bool // distinguishes "not set" from "black" (RGB 0,0,0)
}

// RGB creates a color from red, green, blue components (0-255).
func RGB(r, g, b uint8) Color {
	return Color{r: r, g: g, b: b, set: true}
}

// Hex creates a color from a hex string.
// Accepts formats: "#RRGGBB", "RRGGBB", "#RGB", "RGB".
func Hex(hex string) Color {
	hex = strings.TrimPrefix(hex, "#")

	// Handle short form (#RGB)
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return Color{} // invalid, return default
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return Color{} // invalid, return default
	}

	return RGB(r, g, b)
}

// HSL creates a color from hue (0-360), saturation (0-1), lightness (0-1).
func HSL(h, s, l float64) Color {
	r, g, b := hslToRGB(h, s, l)
	return RGB(r, g, b)
}

// ANSI color constants (backwards compatible names with true color values).
var (
	Black         = RGB(0, 0, 0)
	Red           = RGB(170, 0, 0)
	Green         = RGB(0, 170, 0)
	Yellow        = RGB(170, 170, 0)
	Blue          = RGB(0, 0, 170)
	Magenta       = RGB(170, 0, 170)
	Cyan          = RGB(0, 170, 170)
	White         = RGB(170, 170, 170)
	BrightBlack   = RGB(85, 85, 85)
	BrightRed     = RGB(255, 85, 85)
	BrightGreen   = RGB(85, 255, 85)
	BrightYellow  = RGB(255, 255, 85)
	BrightBlue    = RGB(85, 85, 255)
	BrightMagenta = RGB(255, 85, 255)
	BrightCyan    = RGB(85, 255, 255)
	BrightWhite   = RGB(255, 255, 255)
)

// --- Inspection Methods ---

// RGB returns the red, green, and blue components (0-255).
func (c Color) RGB() (r, g, b uint8) {
	return c.r, c.g, c.b
}

// HSL returns the hue (0-360), saturation (0-1), and lightness (0-1).
func (c Color) HSL() (h, s, l float64) {
	return rgbToHSL(c.r, c.g, c.b)
}

// Hex returns the color as a hex string "#RRGGBB".
func (c Color) Hex() string {
	if !c.set {
		return ""
	}
	return fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
}

// IsSet returns true if the color was explicitly set.
func (c Color) IsSet() bool {
	return c.set
}

// IsDark returns true if the color's lightness is less than 0.5.
func (c Color) IsDark() bool {
	_, _, l := c.HSL()
	return l < 0.5
}

// IsLight returns true if the color's lightness is >= 0.5.
func (c Color) IsLight() bool {
	return !c.IsDark()
}

// Luminance returns the relative luminance of the color (0-1).
// Uses the WCAG formula for calculating relative luminance.
func (c Color) Luminance() float64 {
	// Convert to linear RGB
	rLinear := linearize(float64(c.r) / 255)
	gLinear := linearize(float64(c.g) / 255)
	bLinear := linearize(float64(c.b) / 255)

	// Apply luminance weights
	return 0.2126*rLinear + 0.7152*gLinear + 0.0722*bLinear
}

// ContrastRatio returns the WCAG contrast ratio between two colors.
// The ratio ranges from 1:1 (identical) to 21:1 (black on white).
func (c Color) ContrastRatio(other Color) float64 {
	l1 := c.Luminance()
	l2 := other.Luminance()

	// Ensure l1 is the lighter color
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// --- Fluent Manipulation Methods ---

// Lighten increases the lightness of the color.
// amount should be between 0 and 1.
func (c Color) Lighten(amount float64) Color {
	if !c.set {
		return c
	}
	h, s, l := c.HSL()
	l = clamp01(l + amount)
	return HSL(h, s, l)
}

// Darken decreases the lightness of the color.
// amount should be between 0 and 1.
func (c Color) Darken(amount float64) Color {
	if !c.set {
		return c
	}
	h, s, l := c.HSL()
	l = clamp01(l - amount)
	return HSL(h, s, l)
}

// Saturate increases the saturation of the color.
// amount should be between 0 and 1.
func (c Color) Saturate(amount float64) Color {
	if !c.set {
		return c
	}
	h, s, l := c.HSL()
	s = clamp01(s + amount)
	return HSL(h, s, l)
}

// Desaturate decreases the saturation of the color.
// amount should be between 0 and 1.
func (c Color) Desaturate(amount float64) Color {
	if !c.set {
		return c
	}
	h, s, l := c.HSL()
	s = clamp01(s - amount)
	return HSL(h, s, l)
}

// Rotate rotates the hue by the given number of degrees.
func (c Color) Rotate(degrees float64) Color {
	if !c.set {
		return c
	}
	h, s, l := c.HSL()
	h = math.Mod(h+degrees, 360)
	if h < 0 {
		h += 360
	}
	return HSL(h, s, l)
}

// Complement returns the complementary color (hue rotated 180 degrees).
func (c Color) Complement() Color {
	return c.Rotate(180)
}

// Invert returns the inverted color.
func (c Color) Invert() Color {
	if !c.set {
		return c
	}
	return RGB(255-c.r, 255-c.g, 255-c.b)
}

// Blend mixes this color with another color.
// ratio of 0 returns this color, ratio of 1 returns the other color.
func (c Color) Blend(other Color, ratio float64) Color {
	if !c.set {
		return other
	}
	if !other.set {
		return c
	}

	ratio = clamp01(ratio)
	invRatio := 1 - ratio

	r := uint8(float64(c.r)*invRatio + float64(other.r)*ratio)
	g := uint8(float64(c.g)*invRatio + float64(other.g)*ratio)
	b := uint8(float64(c.b)*invRatio + float64(other.b)*ratio)

	return RGB(r, g, b)
}

// --- Internal Methods ---

// toANSI converts to charmbracelet/x/ansi color.Color for rendering.
func (c Color) toANSI() color.Color {
	if !c.set {
		return nil // default/transparent
	}
	return ansi.RGBColor{R: c.r, G: c.g, B: c.b}
}

// --- Helper Functions ---

// rgbToHSL converts RGB (0-255) to HSL (h: 0-360, s: 0-1, l: 0-1).
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	// Lightness
	l = (max + min) / 2

	if delta == 0 {
		// Achromatic (gray)
		return 0, 0, l
	}

	// Saturation
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2 - max - min)
	}

	// Hue
	switch max {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/delta + 2
	case bf:
		h = (rf-gf)/delta + 4
	}
	h *= 60

	return h, s, l
}

// hslToRGB converts HSL (h: 0-360, s: 0-1, l: 0-1) to RGB (0-255).
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		// Achromatic (gray)
		v := uint8(l * 255)
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	h = h / 360

	r = uint8(hueToRGB(p, q, h+1.0/3.0) * 255)
	g = uint8(hueToRGB(p, q, h) * 255)
	b = uint8(hueToRGB(p, q, h-1.0/3.0) * 255)

	return r, g, b
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}

	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// linearize converts sRGB component to linear RGB for luminance calculation.
func linearize(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// clamp01 clamps a value to the range [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
