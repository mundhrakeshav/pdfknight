package direct

import (
	"fmt"
	"math"
	"strconv"
)

// Dark mode color values (normalized 0-1)
var (
	darkBgValue    = 0.102 // #1a1a1a = 26/255
	lightTextValue = 0.878 // #e0e0e0 = 224/255
)

// Transformer handles color value transformations for dark mode
type Transformer struct{}

// NewTransformer creates a new color transformer
func NewTransformer() *Transformer {
	return &Transformer{}
}

// TransformOperator transforms a color operator for dark mode
// Returns the new operator string
func (t *Transformer) TransformOperator(op ColorOperator) string {
	switch op.ColorSpace {
	case "rgb":
		return t.transformRGB(op)
	case "gray":
		return t.transformGray(op)
	case "cmyk":
		return t.transformCMYK(op)
	default:
		return op.FullMatch // Return unchanged if unknown
	}
}

// transformRGB transforms an RGB color operator
func (t *Transformer) transformRGB(op ColorOperator) string {
	r := parseFloat(op.Values[0])
	g := parseFloat(op.Values[1])
	b := parseFloat(op.Values[2])

	// Calculate properties
	saturation := t.getSaturation(r, g, b)
	lightness := t.getLightness(r, g, b)

	var newR, newG, newB float64

	// Check if this is a document color (grayscale or near-grayscale)
	if saturation < 0.15 {
		// Document color - apply smart inversion
		newR, newG, newB = t.invertDocumentColorRGB(lightness)
	} else {
		// Colorful pixel - adjust brightness while preserving hue
		newR, newG, newB = t.adjustColorfulRGB(r, g, b, lightness)
	}

	return fmt.Sprintf("%.3f %.3f %.3f %s", newR, newG, newB, op.Operator)
}

// transformGray transforms a grayscale color operator
func (t *Transformer) transformGray(op ColorOperator) string {
	gray := parseFloat(op.Values[0])

	// Apply smart inversion based on lightness
	var newGray float64
	if gray > 0.9 {
		// White -> dark background
		newGray = darkBgValue
	} else if gray > 0.7 {
		// Light gray -> dark gray
		factor := (gray - 0.7) / 0.2
		newGray = 0.1 + (1-factor)*0.15
	} else if gray < 0.15 {
		// Black -> light text
		newGray = lightTextValue
	} else if gray < 0.4 {
		// Dark gray -> light gray
		factor := gray / 0.4
		newGray = 0.88 - factor*0.3
	} else {
		// Mid gray - simple inversion
		newGray = 1 - gray
	}

	return fmt.Sprintf("%.3f %s", newGray, op.Operator)
}

// transformCMYK transforms a CMYK color operator
func (t *Transformer) transformCMYK(op ColorOperator) string {
	c := parseFloat(op.Values[0])
	m := parseFloat(op.Values[1])
	y := parseFloat(op.Values[2])
	k := parseFloat(op.Values[3])

	// Convert CMYK to RGB for analysis
	r := (1 - c) * (1 - k)
	g := (1 - m) * (1 - k)
	b := (1 - y) * (1 - k)

	saturation := t.getSaturation(r, g, b)
	lightness := t.getLightness(r, g, b)

	var newC, newM, newY, newK float64

	if saturation < 0.15 {
		// Document color - convert to grayscale equivalent
		var newGray float64
		if lightness > 0.9 {
			newGray = darkBgValue
		} else if lightness < 0.15 {
			newGray = lightTextValue
		} else {
			newGray = 1 - lightness
		}
		// Convert gray to CMYK (C=M=Y=0, K=1-gray)
		newK = 1 - newGray
		newC, newM, newY = 0, 0, 0
	} else {
		// Colorful - adjust brightness
		newR, newG, newB := t.adjustColorfulRGB(r, g, b, lightness)
		// Convert back to CMYK
		newC, newM, newY, newK = rgbToCMYK(newR, newG, newB)
	}

	return fmt.Sprintf("%.3f %.3f %.3f %.3f %s", newC, newM, newY, newK, op.Operator)
}

// invertDocumentColorRGB returns RGB values for inverted document color
func (t *Transformer) invertDocumentColorRGB(lightness float64) (r, g, b float64) {
	var newLightness float64

	if lightness > 0.9 {
		newLightness = darkBgValue
	} else if lightness > 0.7 {
		factor := (lightness - 0.7) / 0.2
		newLightness = 0.1 + (1-factor)*0.15
	} else if lightness < 0.15 {
		newLightness = lightTextValue
	} else if lightness < 0.4 {
		factor := lightness / 0.4
		newLightness = 0.88 - factor*0.3
	} else {
		newLightness = 1 - lightness
	}

	return newLightness, newLightness, newLightness
}

// adjustColorfulRGB adjusts colorful pixels for dark mode
// Ensures colored text is bright enough to read on dark background
func (t *Transformer) adjustColorfulRGB(r, g, b, lightness float64) (newR, newG, newB float64) {
	h, s, l := rgbToHSL(r, g, b)

	// For dark mode, ensure minimum lightness of 0.55 for readability
	// Dark colors need to be lightened significantly
	if l < 0.55 {
		// Map 0-0.55 to 0.55-0.75 (lighten dark colors)
		l = 0.55 + (l/0.55)*0.2
	} else if l > 0.85 {
		// Very light colors: reduce slightly but keep visible
		l = 0.7 + (l-0.85)*0.5
	}

	// Boost saturation slightly to maintain color vibrancy
	s = math.Min(1.0, s*1.15)

	return hslToRGB(h, s, l)
}

// getSaturation calculates saturation (0-1)
func (t *Transformer) getSaturation(r, g, b float64) float64 {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))

	if max == min {
		return 0
	}

	l := (max + min) / 2
	if l <= 0.5 {
		return (max - min) / (max + min)
	}
	return (max - min) / (2 - max - min)
}

// getLightness calculates lightness (0-1)
func (t *Transformer) getLightness(r, g, b float64) float64 {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	return (max + min) / 2
}

// parseFloat parses a string to float64
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// rgbToHSL converts RGB (0-1) to HSL
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))

	l = (max + min) / 2

	if max == min {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}

	h /= 6
	return
}

// hslToRGB converts HSL to RGB (0-1)
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)

	return
}

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

// rgbToCMYK converts RGB (0-1) to CMYK (0-1)
func rgbToCMYK(r, g, b float64) (c, m, y, k float64) {
	k = 1 - math.Max(r, math.Max(g, b))
	if k == 1 {
		return 0, 0, 0, 1
	}
	c = (1 - r - k) / (1 - k)
	m = (1 - g - k) / (1 - k)
	y = (1 - b - k) / (1 - k)
	return
}
