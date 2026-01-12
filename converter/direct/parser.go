package direct

import (
	"regexp"
	"strings"
)

// ColorOperator represents a color-setting operation in a PDF content stream
type ColorOperator struct {
	FullMatch  string   // The complete matched string
	Values     []string // Color values (numbers)
	Operator   string   // The operator (rg, RG, g, G, k, K, sc, SC, scn, SCN)
	ColorSpace string   // Derived color space (rgb, gray, cmyk)
	IsStroke   bool     // True for stroke (uppercase), false for fill
	StartPos   int      // Position in the content stream
	EndPos     int      // End position in the content stream
}

// Parser finds color operators in PDF content streams
type Parser struct {
	// Regex patterns for different color operators
	rgbPattern     *regexp.Regexp // matches "n n n rg" or "n n n RG"
	grayPattern    *regexp.Regexp // matches "n g" or "n G"
	cmykPattern    *regexp.Regexp // matches "n n n n k" or "n n n n K"
	scRgbPattern   *regexp.Regexp // matches "n n n sc" or "n n n SC" or scn/SCN
	scGrayPattern  *regexp.Regexp // matches "n sc" or "n SC" or scn/SCN
	scCmykPattern  *regexp.Regexp // matches "n n n n sc" or "n n n n SC" or scn/SCN
}

// NewParser creates a new content stream parser
func NewParser() *Parser {
	// Number pattern: matches integers and decimals
	num := `[-+]?(?:\d+\.?\d*|\.\d+)`
	ws := `\s+`

	return &Parser{
		// RGB: three numbers followed by rg or RG
		rgbPattern: regexp.MustCompile(`(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(rg|RG)`),
		// Grayscale: one number followed by g or G
		grayPattern: regexp.MustCompile(`(` + num + `)` + ws + `(g|G)\b`),
		// CMYK: four numbers followed by k or K
		cmykPattern: regexp.MustCompile(`(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(k|K)`),
		// sc/SC/scn/SCN operators for different color spaces
		scRgbPattern:  regexp.MustCompile(`(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(scn?|SCN?)\b`),
		scGrayPattern: regexp.MustCompile(`(` + num + `)` + ws + `(scn?|SCN?)\b`),
		scCmykPattern: regexp.MustCompile(`(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(` + num + `)` + ws + `(scn?|SCN?)\b`),
	}
}

// FindColorOperators finds all color operators in a content stream
func (p *Parser) FindColorOperators(content string) []ColorOperator {
	var operators []ColorOperator

	// Find RGB operators (rg/RG)
	for _, match := range p.rgbPattern.FindAllStringSubmatchIndex(content, -1) {
		op := ColorOperator{
			FullMatch:  content[match[0]:match[1]],
			Values:     []string{content[match[2]:match[3]], content[match[4]:match[5]], content[match[6]:match[7]]},
			Operator:   content[match[8]:match[9]],
			ColorSpace: "rgb",
			IsStroke:   content[match[8]:match[9]] == "RG",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	// Find grayscale operators (g/G)
	for _, match := range p.grayPattern.FindAllStringSubmatchIndex(content, -1) {
		fullMatch := content[match[0]:match[1]]
		operator := content[match[4]:match[5]]

		// Skip if this is part of an RGB/CMYK match
		if match[0] > 0 {
			prevChar := content[match[0]-1]
			if prevChar >= '0' && prevChar <= '9' || prevChar == '.' {
				continue
			}
		}

		op := ColorOperator{
			FullMatch:  fullMatch,
			Values:     []string{content[match[2]:match[3]]},
			Operator:   operator,
			ColorSpace: "gray",
			IsStroke:   operator == "G",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	// Find CMYK operators (k/K)
	for _, match := range p.cmykPattern.FindAllStringSubmatchIndex(content, -1) {
		op := ColorOperator{
			FullMatch: content[match[0]:match[1]],
			Values: []string{
				content[match[2]:match[3]],
				content[match[4]:match[5]],
				content[match[6]:match[7]],
				content[match[8]:match[9]],
			},
			Operator:   content[match[10]:match[11]],
			ColorSpace: "cmyk",
			IsStroke:   content[match[10]:match[11]] == "K",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	// Find sc/SC/scn/SCN with 3 values (RGB color space)
	for _, match := range p.scRgbPattern.FindAllStringSubmatchIndex(content, -1) {
		operator := content[match[8]:match[9]]
		op := ColorOperator{
			FullMatch:  content[match[0]:match[1]],
			Values:     []string{content[match[2]:match[3]], content[match[4]:match[5]], content[match[6]:match[7]]},
			Operator:   operator,
			ColorSpace: "rgb",
			IsStroke:   operator == "SC" || operator == "SCN",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	// Find sc/SC/scn/SCN with 1 value (grayscale)
	for _, match := range p.scGrayPattern.FindAllStringSubmatchIndex(content, -1) {
		fullMatch := content[match[0]:match[1]]
		operator := content[match[4]:match[5]]

		// Skip if this is part of a larger pattern
		if match[0] > 0 {
			prevChar := content[match[0]-1]
			if prevChar >= '0' && prevChar <= '9' || prevChar == '.' {
				continue
			}
		}

		op := ColorOperator{
			FullMatch:  fullMatch,
			Values:     []string{content[match[2]:match[3]]},
			Operator:   operator,
			ColorSpace: "gray",
			IsStroke:   operator == "SC" || operator == "SCN",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	// Find sc/SC/scn/SCN with 4 values (CMYK)
	for _, match := range p.scCmykPattern.FindAllStringSubmatchIndex(content, -1) {
		operator := content[match[10]:match[11]]
		op := ColorOperator{
			FullMatch: content[match[0]:match[1]],
			Values: []string{
				content[match[2]:match[3]],
				content[match[4]:match[5]],
				content[match[6]:match[7]],
				content[match[8]:match[9]],
			},
			Operator:   operator,
			ColorSpace: "cmyk",
			IsStroke:   operator == "SC" || operator == "SCN",
			StartPos:   match[0],
			EndPos:     match[1],
		}
		operators = append(operators, op)
	}

	return operators
}

// ReplaceColorOperators replaces color operators in content with new values
// Replacements should be provided as a map from old value to new value
func (p *Parser) ReplaceColorOperators(content string, replacements map[string]string) string {
	result := content

	// Sort replacements by length (longest first) to avoid partial replacements
	// We need to replace from end to start to maintain positions
	type replacement struct {
		old string
		new string
	}
	var repls []replacement
	for old, new := range replacements {
		repls = append(repls, replacement{old: old, new: new})
	}

	// Replace each occurrence
	for _, repl := range repls {
		result = strings.ReplaceAll(result, repl.old, repl.new)
	}

	return result
}
