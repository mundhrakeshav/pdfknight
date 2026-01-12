package raster

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Renderer handles PDF to image conversion
type Renderer struct {
	dpi int
}

// NewRenderer creates a new Renderer with the specified DPI
func NewRenderer(dpi int) *Renderer {
	return &Renderer{dpi: dpi}
}

// RenderToImages converts a PDF to a slice of images, one per page
// It first tries pdftoppm (poppler-utils), then falls back to a basic approach
func (r *Renderer) RenderToImages(pdfPath string) ([]image.Image, error) {
	// Create temp directory for rendered images
	tempDir, err := os.MkdirTemp("", "pdfdarkmode-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Try pdftoppm first (best quality)
	images, err := r.renderWithPdftoppm(pdfPath, tempDir)
	if err == nil {
		return images, nil
	}

	// Fall back to pdftocairo if pdftoppm fails
	images, err = r.renderWithPdftocairo(pdfPath, tempDir)
	if err == nil {
		return images, nil
	}

	return nil, fmt.Errorf("no PDF renderer available. Please install poppler-utils:\n  macOS: brew install poppler\n  Ubuntu: sudo apt install poppler-utils\n  Windows: download from https://github.com/oschwartz10612/poppler-windows")
}

// renderWithPdftoppm uses poppler's pdftoppm for high-quality rendering
func (r *Renderer) renderWithPdftoppm(pdfPath, tempDir string) ([]image.Image, error) {
	// Check if pdftoppm is available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found: %w", err)
	}

	outputPrefix := filepath.Join(tempDir, "page")

	// Run pdftoppm to convert PDF to PNG images
	cmd := exec.Command("pdftoppm",
		"-png",
		"-r", strconv.Itoa(r.dpi),
		pdfPath,
		outputPrefix,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %w\nOutput: %s", err, string(output))
	}

	return r.loadImagesFromDir(tempDir, "page-*.png")
}

// renderWithPdftocairo uses poppler's pdftocairo as fallback
func (r *Renderer) renderWithPdftocairo(pdfPath, tempDir string) ([]image.Image, error) {
	// Check if pdftocairo is available
	if _, err := exec.LookPath("pdftocairo"); err != nil {
		return nil, fmt.Errorf("pdftocairo not found: %w", err)
	}

	outputPrefix := filepath.Join(tempDir, "page")

	// Run pdftocairo to convert PDF to PNG images
	cmd := exec.Command("pdftocairo",
		"-png",
		"-r", strconv.Itoa(r.dpi),
		pdfPath,
		outputPrefix,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftocairo failed: %w\nOutput: %s", err, string(output))
	}

	return r.loadImagesFromDir(tempDir, "page-*.png")
}

// loadImagesFromDir loads all PNG images matching the pattern from a directory
func (r *Renderer) loadImagesFromDir(dir, pattern string) ([]image.Image, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to glob images: %w", err)
	}

	if len(matches) == 0 {
		// Try without the dash (some versions use different naming)
		matches, err = filepath.Glob(filepath.Join(dir, "page*.png"))
		if err != nil || len(matches) == 0 {
			return nil, fmt.Errorf("no rendered images found")
		}
	}

	// Sort files to ensure correct page order
	sort.Slice(matches, func(i, j int) bool {
		return extractPageNumber(matches[i]) < extractPageNumber(matches[j])
	})

	var images []image.Image
	for _, path := range matches {
		img, err := loadPNG(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load image %s: %w", path, err)
		}
		images = append(images, img)
	}

	return images, nil
}

// extractPageNumber extracts the page number from a filename like "page-01.png"
func extractPageNumber(filename string) int {
	base := filepath.Base(filename)
	base = strings.TrimPrefix(base, "page-")
	base = strings.TrimPrefix(base, "page")
	base = strings.TrimSuffix(base, ".png")
	num, _ := strconv.Atoi(base)
	return num
}

// loadPNG loads a PNG image from a file
func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return png.Decode(f)
}
