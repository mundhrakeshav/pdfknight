package converter

import (
	"fmt"

	"pdfdarkmode/converter/colors"
	"pdfdarkmode/converter/direct"
	"pdfdarkmode/converter/raster"
)

// Options holds the configuration for PDF conversion
type Options struct {
	InputFile      string
	OutputFile     string
	Mode           string        // "raster" or "direct"
	DPI            int           // DPI for raster mode
	PreserveImages bool          // Preserve images in direct mode
	ColorScheme    colors.Scheme // Color scheme for dark mode
}

// Converter interface defines the contract for PDF conversion engines
type Converter interface {
	Convert(input, output string) error
}

// Convert performs the PDF to dark mode conversion using the specified mode
func Convert(opts Options) error {
	var conv Converter

	switch opts.Mode {
	case "raster":
		conv = raster.NewEngine(opts.DPI, opts.ColorScheme)
	case "direct":
		conv = direct.NewEngine(opts.PreserveImages, opts.ColorScheme)
	default:
		return fmt.Errorf("unknown mode: %s", opts.Mode)
	}

	return conv.Convert(opts.InputFile, opts.OutputFile)
}
