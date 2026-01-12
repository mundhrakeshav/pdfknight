package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pdfdarkmode/converter"
)

var (
	outputFile     string
	mode           string
	dpi            int
	preserveImages bool
)

var rootCmd = &cobra.Command{
	Use:   "pdfdarkmode <input.pdf>",
	Short: "Convert PDFs to dark mode",
	Long: `A CLI tool to convert PDF documents to dark mode.

Supports two conversion modes:
  - raster: Converts pages to images, inverts colors, reassembles (reliable)
  - direct: Modifies PDF color operators directly (preserves vectors/text)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inputFile)
		}

		// Set default output file if not specified
		if outputFile == "" {
			outputFile = strings.TrimSuffix(inputFile, ".pdf") + "_dark.pdf"
		}

		// If mode not specified, ask user interactively
		if mode == "" {
			mode = selectModeInteractively()
		}

		// Validate mode
		if mode != "raster" && mode != "direct" {
			return fmt.Errorf("invalid mode: %s (must be 'raster' or 'direct')", mode)
		}

		// Create converter options
		opts := converter.Options{
			InputFile:      inputFile,
			OutputFile:     outputFile,
			Mode:           mode,
			DPI:            dpi,
			PreserveImages: preserveImages,
		}

		// Run conversion
		fmt.Printf("Converting %s to dark mode using %s mode...\n", inputFile, mode)
		if err := converter.Convert(opts); err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		fmt.Printf("Successfully created: %s\n", outputFile)
		return nil
	},
}

func selectModeInteractively() string {
	fmt.Println("\nSelect conversion mode:")
	fmt.Println("  [1] raster  - Converts pages to images, then inverts")
	fmt.Println("                + Works with any PDF")
	fmt.Println("                - Larger file size, no text selection")
	fmt.Println("  [2] direct  - Modifies PDF color operators directly")
	fmt.Println("                + Preserves vectors, text, small file size")
	fmt.Println("                - May not work with complex PDFs")
	fmt.Print("\nEnter choice (1 or 2): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1", "raster":
		return "raster"
	case "2", "direct":
		return "direct"
	default:
		fmt.Println("Invalid choice, defaulting to 'raster' mode")
		return "raster"
	}
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output PDF file (default: <input>_dark.pdf)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "", "Conversion mode: 'raster' or 'direct'")
	rootCmd.Flags().IntVar(&dpi, "dpi", 150, "DPI for raster mode (default: 150)")
	rootCmd.Flags().BoolVar(&preserveImages, "preserve-images", true, "Preserve images in direct mode (default: true)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
