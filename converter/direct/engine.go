package direct

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Engine implements direct PDF manipulation for dark mode conversion
type Engine struct {
	preserveImages bool
	parser         *Parser
	transformer    *Transformer
}

// NewEngine creates a new direct manipulation engine
func NewEngine(preserveImages bool) *Engine {
	return &Engine{
		preserveImages: preserveImages,
		parser:         NewParser(),
		transformer:    NewTransformer(),
	}
}

// Convert performs direct PDF manipulation to convert to dark mode
func (e *Engine) Convert(inputPath, outputPath string) error {
	fmt.Println("  [1/4] Reading PDF structure...")

	// Read the PDF file
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer f.Close()

	// Create a configuration
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Parse the PDF using the api package
	ctx, err := api.ReadContext(f, conf)
	if err != nil {
		return fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Ensure page count is calculated
	if err := ctx.EnsurePageCount(); err != nil {
		return fmt.Errorf("failed to determine page count: %w", err)
	}

	fmt.Printf("        PDF version: %s, Pages: %d\n", ctx.HeaderVersion, ctx.PageCount)

	fmt.Println("  [2/4] Processing page content streams...")
	pagesProcessed := 0
	colorsTransformed := 0

	// Process each page
	for pageNum := 1; pageNum <= ctx.PageCount; pageNum++ {
		count, err := e.processPage(ctx, pageNum)
		if err != nil {
			fmt.Printf("        Warning: failed to process page %d: %v\n", pageNum, err)
			continue
		}
		pagesProcessed++
		colorsTransformed += count
	}

	fmt.Printf("        Processed %d pages, transformed %d color operations\n", pagesProcessed, colorsTransformed)

	fmt.Println("  [3/4] Adding dark background to pages...")
	if err := e.addDarkBackgrounds(ctx); err != nil {
		fmt.Printf("        Warning: could not add backgrounds: %v\n", err)
	}

	fmt.Println("  [4/4] Writing output PDF...")

	// Write the modified PDF
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := api.WriteContext(ctx, outFile); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

// processPage processes a single page's content streams
func (e *Engine) processPage(ctx *model.Context, pageNum int) (int, error) {
	// Get the page dictionary
	pageDict, _, _, err := ctx.PageDict(pageNum, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get page dict: %w", err)
	}

	// Get the Contents entry
	contentsEntry, found := pageDict.Find("Contents")
	if !found {
		return 0, nil // Page has no content
	}

	totalTransformed := 0

	// Handle different content types
	switch contents := contentsEntry.(type) {
	case types.IndirectRef:
		// Single content stream
		count, err := e.processContentStream(ctx, contents)
		if err != nil {
			return 0, err
		}
		totalTransformed += count

	case types.Array:
		// Array of content streams
		for _, item := range contents {
			if ref, ok := item.(types.IndirectRef); ok {
				count, err := e.processContentStream(ctx, ref)
				if err != nil {
					continue
				}
				totalTransformed += count
			}
		}
	}

	return totalTransformed, nil
}

// processContentStream processes a single content stream
func (e *Engine) processContentStream(ctx *model.Context, ref types.IndirectRef) (int, error) {
	// Get the stream object
	obj, err := ctx.Dereference(ref)
	if err != nil {
		return 0, err
	}

	sd, ok := obj.(types.StreamDict)
	if !ok {
		return 0, nil
	}

	// Decode the stream content
	if err := sd.Decode(); err != nil {
		return 0, nil // Skip streams we can't decode
	}

	content := sd.Content
	if content == nil {
		return 0, nil
	}

	// Find and transform color operators
	operators := e.parser.FindColorOperators(string(content))
	if len(operators) == 0 {
		return 0, nil
	}

	// Build replacement map
	replacements := make(map[string]string)
	for _, op := range operators {
		newOp := e.transformer.TransformOperator(op)
		if newOp != op.FullMatch {
			replacements[op.FullMatch] = newOp
		}
	}

	if len(replacements) == 0 {
		return 0, nil
	}

	// Apply replacements
	newContent := e.parser.ReplaceColorOperators(string(content), replacements)

	// Re-encode the stream using pdfcpu's Encode method
	sd.Content = []byte(newContent)
	if err := sd.Encode(); err != nil {
		return 0, fmt.Errorf("failed to encode stream: %w", err)
	}

	// Update length in dictionary
	sd.Dict["Length"] = types.Integer(len(sd.Raw))

	// Update the object in the context
	entry, found := ctx.FindTableEntryForIndRef(&ref)
	if !found {
		return 0, fmt.Errorf("could not find xref entry")
	}
	entry.Object = sd

	return len(replacements), nil
}

// addDarkBackgrounds adds a dark background rectangle to each page
func (e *Engine) addDarkBackgrounds(ctx *model.Context) error {
	for pageNum := 1; pageNum <= ctx.PageCount; pageNum++ {
		if err := e.addPageBackground(ctx, pageNum); err != nil {
			fmt.Printf("        Warning: page %d background failed: %v\n", pageNum, err)
			continue
		}
	}
	return nil
}

// addPageBackground adds a dark background to a single page by PREPENDING to content
func (e *Engine) addPageBackground(ctx *model.Context, pageNum int) error {
	pageDict, _, inhPAttrs, err := ctx.PageDict(pageNum, false)
	if err != nil {
		return err
	}

	// Get MediaBox - try page dict first, then inherited attributes
	var mediaBox *types.Rectangle

	// Try to get MediaBox from page dictionary
	if mb, found := pageDict.Find("MediaBox"); found {
		if arr, ok := mb.(types.Array); ok {
			mediaBox = types.RectForArray(arr)
		}
	}

	// Try inherited attributes if not found
	if mediaBox == nil && inhPAttrs != nil && inhPAttrs.MediaBox != nil {
		mediaBox = inhPAttrs.MediaBox
	}

	// Fallback to US Letter size (612x792 points)
	if mediaBox == nil {
		mediaBox = types.NewRectangle(0, 0, 612, 792)
	}

	// Create background content - this will be PREPENDED to draw behind existing content
	// 1. Draw dark background rectangle (#1a1a1a = 26/255 ≈ 0.102)
	// 2. Set default text/fill color to light gray (#e0e0e0 = 224/255 ≈ 0.878)
	// 3. Set default stroke color to light gray
	// This ensures any text without explicit color uses light color on dark background
	bgContent := fmt.Sprintf("q 0.102 0.102 0.102 rg %.2f %.2f %.2f %.2f re f Q 0.878 0.878 0.878 rg 0.878 0.878 0.878 RG\n",
		mediaBox.LL.X, mediaBox.LL.Y, mediaBox.Width(), mediaBox.Height())

	// Get the Contents entry
	contentsEntry, found := pageDict.Find("Contents")
	if !found {
		// No content - just add background
		return ctx.AppendContent(pageDict, []byte(bgContent))
	}

	// Handle single content stream - prepend background to it
	switch contents := contentsEntry.(type) {
	case types.IndirectRef:
		return e.prependToStream(ctx, contents, []byte(bgContent))
	case types.Array:
		if len(contents) > 0 {
			if ref, ok := contents[0].(types.IndirectRef); ok {
				return e.prependToStream(ctx, ref, []byte(bgContent))
			}
		}
	}

	return nil
}

// prependToStream prepends content to a stream
func (e *Engine) prependToStream(ctx *model.Context, ref types.IndirectRef, prefix []byte) error {
	obj, err := ctx.Dereference(ref)
	if err != nil {
		return err
	}

	sd, ok := obj.(types.StreamDict)
	if !ok {
		return fmt.Errorf("not a stream dict")
	}

	// Decode existing content
	if err := sd.Decode(); err != nil {
		return err
	}

	// Prepend the background
	newContent := append(prefix, sd.Content...)

	// Re-encode
	sd.Content = newContent
	if err := sd.Encode(); err != nil {
		return err
	}

	// Update length
	sd.Dict["Length"] = types.Integer(len(sd.Raw))

	// Update in context
	entry, found := ctx.FindTableEntryForIndRef(&ref)
	if !found {
		return fmt.Errorf("could not find xref entry")
	}
	entry.Object = sd

	return nil
}
