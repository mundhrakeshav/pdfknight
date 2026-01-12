package raster

import (
	"image"
	"image/color"
	"math"

	"pdfdarkmode/converter/colors"
)

// Inverter handles smart color inversion for dark mode
type Inverter struct {
	scheme colors.Scheme
}

// NewInverter creates a new Inverter with the given color scheme
func NewInverter(scheme colors.Scheme) *Inverter {
	return &Inverter{scheme: scheme}
}

// InvertImage applies smart dark mode inversion to an image
// It inverts document colors (black/white/gray) while preserving colorful elements
func (inv *Inverter) InvertImage(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			newColor := inv.smartInvertPixel(originalColor)
			result.Set(x, y, newColor)
		}
	}

	return result
}

// smartInvertPixel applies smart inversion to a single pixel
func (inv *Inverter) smartInvertPixel(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	a8 := uint8(a >> 8)

	// Calculate color properties
	saturation := inv.getSaturation(r8, g8, b8)
	lightness := inv.getLightness(r8, g8, b8)

	// Determine if this is a "document color" (grayscale or near-grayscale)
	isDocumentColor := saturation < 0.15

	if isDocumentColor {
		// For document colors, apply smart inversion
		return inv.invertDocumentColor(r8, g8, b8, a8, lightness)
	}

	// For colorful pixels (likely images/charts), adjust brightness but preserve hue
	return inv.adjustColorfulPixel(r8, g8, b8, a8, lightness)
}

// invertDocumentColor inverts grayscale document colors for dark mode
func (inv *Inverter) invertDocumentColor(r, g, b, a uint8, lightness float64) color.Color {
	bg := inv.scheme.Background
	txt := inv.scheme.Text

	if lightness > 0.9 {
		// Very light (white background) -> dark background (full RGB)
		return color.RGBA{R: bg.R8, G: bg.G8, B: bg.B8, A: a}
	} else if lightness > 0.7 {
		// Light gray -> interpolate towards background
		factor := (lightness - 0.7) / 0.2 // 1 at 0.9, 0 at 0.7
		newR := txt.R + factor*(bg.R-txt.R)
		newG := txt.G + factor*(bg.G-txt.G)
		newB := txt.B + factor*(bg.B-txt.B)
		return color.RGBA{R: uint8(newR * 255), G: uint8(newG * 255), B: uint8(newB * 255), A: a}
	} else if lightness < 0.15 {
		// Very dark (black text) -> light text (full RGB for tinted text)
		return color.RGBA{R: txt.R8, G: txt.G8, B: txt.B8, A: a}
	} else if lightness < 0.4 {
		// Dark gray -> interpolate from text color towards mid-gray
		factor := lightness / 0.4 // 0 at 0, 1 at 0.4
		midGray := 0.5
		newR := txt.R - factor*(txt.R-midGray)
		newG := txt.G - factor*(txt.G-midGray)
		newB := txt.B - factor*(txt.B-midGray)
		return color.RGBA{R: uint8(newR * 255), G: uint8(newG * 255), B: uint8(newB * 255), A: a}
	}

	// Mid-gray: simple inversion
	inverted := uint8((1 - lightness) * 255)
	return color.RGBA{R: inverted, G: inverted, B: inverted, A: a}
}

// adjustColorfulPixel adjusts colorful pixels for dark mode while preserving hue
func (inv *Inverter) adjustColorfulPixel(r, g, b, a uint8, lightness float64) color.Color {
	// Convert to HSL
	h, s, l := rgbToHSL(r, g, b)

	// Adjust lightness for dark mode viewing
	// Very light colors get darkened, very dark colors get lightened
	if l > 0.7 {
		// Light colorful elements: reduce lightness but keep visible
		l = 0.5 + (l-0.7)*0.5
	} else if l < 0.3 {
		// Dark colorful elements: increase lightness
		l = 0.3 + l*0.3
	}

	// Slightly boost saturation for better visibility on dark background
	s = math.Min(1.0, s*1.1)

	// Convert back to RGB
	newR, newG, newB := hslToRGB(h, s, l)
	return color.RGBA{R: newR, G: newG, B: newB, A: a}
}

// getSaturation calculates the saturation of a color (0-1)
func (inv *Inverter) getSaturation(r, g, b uint8) float64 {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	if max == min {
		return 0
	}

	l := (max + min) / 2
	if l <= 0.5 {
		return (max - min) / (max + min)
	}
	return (max - min) / (2 - max - min)
}

// getLightness calculates the lightness of a color (0-1)
func (inv *Inverter) getLightness(r, g, b uint8) float64 {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	return (max + min) / 2
}

// rgbToHSL converts RGB to HSL color space
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	l = (max + min) / 2

	if max == min {
		h = 0
		s = 0
		return
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	case bf:
		h = (rf-gf)/d + 4
	}

	h /= 6
	return
}

// hslToRGB converts HSL to RGB color space
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		gray := uint8(l * 255)
		return gray, gray, gray
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = uint8(hueToRGB(p, q, h+1.0/3.0) * 255)
	g = uint8(hueToRGB(p, q, h) * 255)
	b = uint8(hueToRGB(p, q, h-1.0/3.0) * 255)

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
