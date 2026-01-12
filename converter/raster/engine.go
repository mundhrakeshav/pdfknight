package raster

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// Engine implements the raster-based PDF dark mode conversion
type Engine struct {
	dpi      int
	renderer *Renderer
	inverter *Inverter
}

// NewEngine creates a new raster conversion engine
func NewEngine(dpi int) *Engine {
	return &Engine{
		dpi:      dpi,
		renderer: NewRenderer(dpi),
		inverter: NewInverter(),
	}
}

// Convert performs the raster-based PDF to dark mode conversion
func (e *Engine) Convert(inputPath, outputPath string) error {
	fmt.Println("  [1/4] Rendering PDF pages to images...")
	images, err := e.renderer.RenderToImages(inputPath)
	if err != nil {
		return fmt.Errorf("failed to render PDF: %w", err)
	}
	fmt.Printf("        Rendered %d page(s)\n", len(images))

	fmt.Println("  [2/4] Applying smart dark mode inversion...")
	invertedImages := make([]image.Image, len(images))
	for i, img := range images {
		invertedImages[i] = e.inverter.InvertImage(img)
		fmt.Printf("        Inverted page %d/%d\n", i+1, len(images))
	}

	fmt.Println("  [3/4] Saving inverted images...")
	tempDir, err := os.MkdirTemp("", "pdfdarkmode-output-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var imagePaths []string
	for i, img := range invertedImages {
		path := filepath.Join(tempDir, fmt.Sprintf("page-%03d.png", i+1))
		if err := savePNG(path, img); err != nil {
			return fmt.Errorf("failed to save image %d: %w", i+1, err)
		}
		imagePaths = append(imagePaths, path)
	}

	fmt.Println("  [4/4] Creating output PDF...")
	if err := e.createPDFFromImages(imagePaths, outputPath); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}

// createPDFFromImages creates a PDF from a list of image files
func (e *Engine) createPDFFromImages(imagePaths []string, outputPath string) error {
	// Use pdfcpu's ImportImages to create PDF from images
	imp := pdfcpu.DefaultImportConfig()
	imp.DPI = e.dpi

	// Import images into a new PDF
	if err := api.ImportImagesFile(imagePaths, outputPath, imp, nil); err != nil {
		return fmt.Errorf("pdfcpu import failed: %w", err)
	}

	return nil
}

// savePNG saves an image as a PNG file
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
