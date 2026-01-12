# PDF Dark Mode

A CLI tool to convert PDF documents to dark mode, written in Go.

## Features

- **Two conversion modes:**
  - **Raster mode** - Converts pages to images, applies smart inversion, reassembles as PDF
    - Works with any PDF
    - Larger output file size
    - No text selection in output
  - **Direct mode** - Modifies PDF color operators directly
    - Preserves vectors and text selectability
    - Smaller file size
    - May not work with complex PDFs

- **Smart color inversion:**
  - White/light backgrounds → Dark (#1a1a1a)
  - Black/dark text → Light (#e0e0e0)
  - Colorful elements (images, charts) → Preserved with adjusted brightness

## Installation

### Prerequisites

**For raster mode**, you need `poppler-utils` installed:

```bash
# macOS
brew install poppler

# Ubuntu/Debian
sudo apt install poppler-utils

# Windows
# Download from https://github.com/oschwartz10612/poppler-windows
```

### Build from source

```bash
git clone https://github.com/yourusername/pdfdarkmode.git
cd pdfdarkmode
go build -o pdfdarkmode .
```

Or install directly:

```bash
go install github.com/yourusername/pdfdarkmode@latest
```

## Usage

### Interactive mode

```bash
pdfdarkmode input.pdf
```

This will prompt you to select a conversion mode and create `input_dark.pdf`.

### Explicit mode selection

```bash
# Raster mode (recommended for complex PDFs)
pdfdarkmode input.pdf -o output.pdf --mode raster

# Direct mode (recommended for simple text documents)
pdfdarkmode input.pdf -o output.pdf --mode direct
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output` | Output PDF file path | `<input>_dark.pdf` |
| `-m, --mode` | Conversion mode: `raster` or `direct` | Interactive prompt |
| `--dpi` | DPI for raster mode rendering | 150 |
| `--preserve-images` | Preserve images in direct mode | true |

### Examples

```bash
# Convert with default settings
pdfdarkmode document.pdf

# High-quality raster conversion
pdfdarkmode document.pdf -o dark.pdf --mode raster --dpi 300

# Direct manipulation
pdfdarkmode document.pdf -o dark.pdf --mode direct
```

## Mode Comparison

| Aspect | Raster Mode | Direct Mode |
|--------|-------------|-------------|
| Reliability | High - works with any PDF | Medium - may fail on complex PDFs |
| Output quality | Good (configurable DPI) | Perfect (vector preserved) |
| File size | Larger | Same as original |
| Text selection | Lost | Preserved |
| Speed | Slower | Faster |

## How It Works

### Raster Mode

1. Renders each PDF page to a PNG image using `pdftoppm` (poppler)
2. Applies smart color inversion to each pixel:
   - Identifies "document colors" (grayscale) vs "colorful" pixels
   - Inverts document colors for dark mode
   - Adjusts colorful pixels to maintain visibility
3. Reassembles inverted images into a new PDF

### Direct Mode

1. Parses the PDF structure using pdfcpu
2. Finds color operators in page content streams (`rg`, `RG`, `g`, `G`, `k`, `K`)
3. Transforms color values for dark mode
4. Adds a dark background to each page
5. Writes the modified PDF

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
