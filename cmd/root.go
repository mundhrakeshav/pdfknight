package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"pdfdarkmode/converter"
	"pdfdarkmode/converter/colors"
)

var (
	outputFile     string
	mode           string
	dpi            int
	preserveImages bool
	colorScheme    string
	bgColor        string
	textColor      string

	// Version info
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// SetVersionInfo sets the version information from main
func SetVersionInfo(v, bt, gc string) {
	version = v
	buildTime = bt
	gitCommit = gc
}

var rootCmd = &cobra.Command{
	Use:   "pdfdarkmode <input.pdf>",
	Short: "Convert PDFs to dark mode",
	Long: `A CLI tool to convert PDF documents to dark mode.

Supports two conversion modes:
  - raster: Converts pages to images, inverts colors, reassembles (reliable)
  - direct: Modifies PDF color operators directly (preserves vectors/text)

Available color schemes: dark, sepia, nord, solarized, gruvbox, dracula, monokai
Or use --bg-color and --text-color for custom colors (hex format: #1a1a1a)`,
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

		// Determine color scheme
		scheme, err := resolveColorScheme()
		if err != nil {
			return err
		}

		// Create converter options
		opts := converter.Options{
			InputFile:      inputFile,
			OutputFile:     outputFile,
			Mode:           mode,
			DPI:            dpi,
			PreserveImages: preserveImages,
			ColorScheme:    scheme,
		}

		// Run conversion
		fmt.Printf("Converting %s to dark mode using %s mode...\n", inputFile, mode)
		fmt.Printf("Color scheme: %s (bg: %s, text: %s)\n", scheme.Name, scheme.Background.Hex(), scheme.Text.Hex())
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

// resolveColorScheme determines the color scheme based on flags
func resolveColorScheme() (colors.Scheme, error) {
	// Custom colors take precedence
	if bgColor != "" || textColor != "" {
		// If only one is specified, use defaults for the other
		bg := bgColor
		text := textColor
		if bg == "" {
			bg = colors.DefaultScheme().Background.Hex()
		}
		if text == "" {
			text = colors.DefaultScheme().Text.Hex()
		}
		return colors.NewCustomScheme(bg, text)
	}

	// If no scheme specified, prompt interactively
	if colorScheme == "" {
		return selectColorSchemeInteractively(), nil
	}

	// Try to get the named scheme
	return colors.GetScheme(colorScheme)
}

func selectColorSchemeInteractively() colors.Scheme {
	fmt.Println("\nSelect color scheme:")

	// Get sorted scheme names for consistent display
	schemeNames := colors.ListSchemes()
	sort.Strings(schemeNames)

	for i, name := range schemeNames {
		scheme := colors.AvailableSchemes[name]
		fmt.Printf("  [%d] %-10s (bg: %s, text: %s)\n", i+1, name, scheme.Background.Hex(), scheme.Text.Hex())
	}
	fmt.Println("  [c] custom    - Enter your own hex colors")

	fmt.Print("\nEnter choice: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Check for custom option
	if input == "c" || input == "custom" {
		return promptCustomColors()
	}

	// Try to parse as number
	if num, err := fmt.Sscanf(input, "%d", new(int)); err == nil && num == 1 {
		var idx int
		fmt.Sscanf(input, "%d", &idx)
		if idx >= 1 && idx <= len(schemeNames) {
			name := schemeNames[idx-1]
			return colors.AvailableSchemes[name]
		}
	}

	// Try to parse as scheme name
	if scheme, err := colors.GetScheme(input); err == nil {
		return scheme
	}

	fmt.Println("Invalid choice, using default 'dark' scheme")
	return colors.DefaultScheme()
}

func promptCustomColors() colors.Scheme {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter background color (hex, e.g., #1a1a1a): ")
	bgInput, _ := reader.ReadString('\n')
	bgInput = strings.TrimSpace(bgInput)

	fmt.Print("Enter text color (hex, e.g., #e0e0e0): ")
	textInput, _ := reader.ReadString('\n')
	textInput = strings.TrimSpace(textInput)

	scheme, err := colors.NewCustomScheme(bgInput, textInput)
	if err != nil {
		fmt.Printf("Invalid colors: %v\nUsing default 'dark' scheme\n", err)
		return colors.DefaultScheme()
	}

	return scheme
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pdfdarkmode %s\n", version)
		fmt.Printf("  Build time: %s\n", buildTime)
		fmt.Printf("  Git commit: %s\n", gitCommit)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output PDF file (default: <input>_dark.pdf)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "", "Conversion mode: 'raster' or 'direct'")
	rootCmd.Flags().IntVar(&dpi, "dpi", 150, "DPI for raster mode (default: 150)")
	rootCmd.Flags().BoolVar(&preserveImages, "preserve-images", true, "Preserve images in direct mode (default: true)")

	// Color options
	rootCmd.Flags().StringVarP(&colorScheme, "scheme", "s", "", "Color scheme: dark, sepia, nord, solarized, gruvbox, dracula, monokai")
	rootCmd.Flags().StringVar(&bgColor, "bg-color", "", "Custom background color (hex, e.g., #1a1a1a)")
	rootCmd.Flags().StringVar(&textColor, "text-color", "", "Custom text color (hex, e.g., #e0e0e0)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(schemesCmd)
}

var schemesCmd = &cobra.Command{
	Use:   "schemes",
	Short: "List available color schemes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available color schemes:")
		fmt.Println()

		schemeNames := colors.ListSchemes()
		sort.Strings(schemeNames)

		for _, name := range schemeNames {
			scheme := colors.AvailableSchemes[name]
			fmt.Printf("  %-10s  Background: %s  Text: %s\n", name, scheme.Background.Hex(), scheme.Text.Hex())
		}

		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  pdfdarkmode --scheme nord input.pdf")
		fmt.Println("  pdfdarkmode --bg-color '#282a36' --text-color '#f8f8f2' input.pdf")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
