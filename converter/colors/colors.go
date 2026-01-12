package colors

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Scheme represents a color scheme for dark mode conversion
type Scheme struct {
	Name       string
	Background Color // Dark background color
	Text       Color // Light text color
}

// Color represents a color with both normalized (0-1) and 8-bit (0-255) values
type Color struct {
	R8, G8, B8 uint8   // 8-bit values (0-255)
	R, G, B    float64 // Normalized values (0-1)
}

// NewColorFromHex creates a Color from a hex string (e.g., "#1a1a1a" or "1a1a1a")
func NewColorFromHex(hex string) (Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return Color{}, fmt.Errorf("invalid hex color: %s (expected 6 characters)", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid red component in hex: %s", hex)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid green component in hex: %s", hex)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return Color{}, fmt.Errorf("invalid blue component in hex: %s", hex)
	}

	return NewColorFromRGB8(uint8(r), uint8(g), uint8(b)), nil
}

// NewColorFromRGB8 creates a Color from 8-bit RGB values
func NewColorFromRGB8(r, g, b uint8) Color {
	return Color{
		R8: r, G8: g, B8: b,
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
	}
}

// ToRGBA converts to Go's color.RGBA
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{R: c.R8, G: c.G8, B: c.B8, A: 255}
}

// Hex returns the hex string representation
func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R8, c.G8, c.B8)
}

// Predefined color schemes
var (
	// SchemeDark is the default dark mode scheme (#1a1a1a background, #e0e0e0 text)
	SchemeDark = Scheme{
		Name:       "dark",
		Background: NewColorFromRGB8(26, 26, 26),   // #1a1a1a
		Text:       NewColorFromRGB8(224, 224, 224), // #e0e0e0
	}

	// SchemeSepia is a warm sepia-toned scheme
	SchemeSepia = Scheme{
		Name:       "sepia",
		Background: NewColorFromRGB8(30, 25, 20),   // #1e1914
		Text:       NewColorFromRGB8(230, 218, 200), // #e6dac8
	}

	// SchemeNord is inspired by the Nord color palette
	SchemeNord = Scheme{
		Name:       "nord",
		Background: NewColorFromRGB8(46, 52, 64),   // #2e3440
		Text:       NewColorFromRGB8(236, 239, 244), // #eceff4
	}

	// SchemeSolarized is inspired by Solarized Dark
	SchemeSolarized = Scheme{
		Name:       "solarized",
		Background: NewColorFromRGB8(0, 43, 54),    // #002b36
		Text:       NewColorFromRGB8(131, 148, 150), // #839496
	}

	// SchemeGruvbox is inspired by Gruvbox Dark
	SchemeGruvbox = Scheme{
		Name:       "gruvbox",
		Background: NewColorFromRGB8(40, 40, 40),   // #282828
		Text:       NewColorFromRGB8(235, 219, 178), // #ebdbb2
	}

	// SchemeDracula is inspired by Dracula theme
	SchemeDracula = Scheme{
		Name:       "dracula",
		Background: NewColorFromRGB8(40, 42, 54),   // #282a36
		Text:       NewColorFromRGB8(248, 248, 242), // #f8f8f2
	}

	// SchemeMonokai is inspired by Monokai theme
	SchemeMonokai = Scheme{
		Name:       "monokai",
		Background: NewColorFromRGB8(39, 40, 34),   // #272822
		Text:       NewColorFromRGB8(248, 248, 240), // #f8f8f0
	}

	// AvailableSchemes maps scheme names to their definitions
	AvailableSchemes = map[string]Scheme{
		"dark":      SchemeDark,
		"sepia":     SchemeSepia,
		"nord":      SchemeNord,
		"solarized": SchemeSolarized,
		"gruvbox":   SchemeGruvbox,
		"dracula":   SchemeDracula,
		"monokai":   SchemeMonokai,
	}
)

// GetScheme returns a scheme by name, or an error if not found
func GetScheme(name string) (Scheme, error) {
	name = strings.ToLower(name)
	if scheme, ok := AvailableSchemes[name]; ok {
		return scheme, nil
	}
	return Scheme{}, fmt.Errorf("unknown color scheme: %s", name)
}

// ListSchemes returns a list of available scheme names
func ListSchemes() []string {
	names := make([]string, 0, len(AvailableSchemes))
	for name := range AvailableSchemes {
		names = append(names, name)
	}
	return names
}

// NewCustomScheme creates a custom scheme from hex colors
func NewCustomScheme(bgHex, textHex string) (Scheme, error) {
	bg, err := NewColorFromHex(bgHex)
	if err != nil {
		return Scheme{}, fmt.Errorf("invalid background color: %w", err)
	}
	text, err := NewColorFromHex(textHex)
	if err != nil {
		return Scheme{}, fmt.Errorf("invalid text color: %w", err)
	}
	return Scheme{
		Name:       "custom",
		Background: bg,
		Text:       text,
	}, nil
}

// DefaultScheme returns the default dark scheme
func DefaultScheme() Scheme {
	return SchemeDark
}
