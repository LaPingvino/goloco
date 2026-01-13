package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostProcessor fixes common Go syntax errors from mechanical C++ conversion
var (
	inDir  = flag.String("in", "", "Directory with converted Go files to fix")
	outDir = flag.String("out", "", "Output directory for fixed files")
)

func main() {
	flag.Parse()
	if *inDir == "" || *outDir == "" {
		fmt.Println("Usage: fix_conversion -in <dir> -out <dir>")
		os.Exit(1)
	}

	if err := filepath.Walk(*inDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		return processFile(path, *outDir)
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Post-processing complete")
}

func processFile(srcPath, outDir string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	text := string(content)
	text = fixStructFields(text)
	text = fixModuleLevelStatements(text)
	text = fixTypeAnnotations(text)
	text = fixImports(text)
	text = cleanupDuplicates(text)

	rel, _ := filepath.Rel(*inDir, srcPath)
	outPath := filepath.Join(outDir, rel)
	os.MkdirAll(filepath.Dir(outPath), 0755)

	return os.WriteFile(outPath, []byte(text), 0644)
}

// fixStructFields removes 'var' keyword from struct field declarations
// Go struct fields don't need var; converts "var name Type" -> "name Type"
func fixStructFields(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inStruct := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track struct context
		if strings.Contains(trimmed, "type ") && strings.Contains(trimmed, "struct") && strings.Contains(line, "{") {
			inStruct = true
			braceDepth = 1
		}

		if inStruct {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth <= 0 {
				inStruct = false
				result = append(result, line)
				continue
			}

			// Remove "var" keyword from struct fields
			if strings.HasPrefix(trimmed, "var ") && !strings.HasPrefix(trimmed, "var_") {
				indent := getIndent(line)
				rest := strings.TrimPrefix(trimmed, "var ")
				line = indent + rest
			}
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// fixModuleLevelStatements wraps top-level if/for/while/return in init() or helper functions
func fixModuleLevelStatements(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	var pending []string
	inFunc := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track function depth
		if !inFunc && isFunctionDef(trimmed) {
			inFunc = true
			braceDepth = 1
		}
		if inFunc {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth <= 0 {
				inFunc = false
			}
		}

		// If we're at module level and encounter a statement, wrap it
		if !inFunc && !isDeclaration(trimmed) && isStatement(trimmed) {
			pending = append(pending, line)
		} else {
			if len(pending) > 0 {
				// Emit wrapped statements
				result = append(result, "func init() {")
				result = append(result, pending...)
				result = append(result, "}")
				result = append(result, "")
				pending = nil
			}
			result = append(result, line)
		}
	}

	if len(pending) > 0 {
		result = append(result, "func init() {")
		result = append(result, pending...)
		result = append(result, "}")
	}

	return strings.Join(result, "\n")
}

// fixTypeAnnotations attempts to resolve undefined type references
// E.g., "World.Pos2" -> "WorldPos2" or comment it
func fixTypeAnnotations(text string) string {
	// Simple heuristic: replace obvious namespace-qualified types
	replacements := map[string]string{
		"World.Pos2":  "WorldPos2",
		"World.Pos3":  "WorldPos3",
		"uint8_t":     "uint8",
		"uint16_t":    "uint16",
		"uint32_t":    "uint32",
		"int16_t":     "int16",
		"int32_t":     "int32",
		"std::array":  "[]",
		"std::string": "string",
		"std::span":   "[]",
	}

	result := text
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// fixImports comments out remaining include-style comments and ensures package declaration
func fixImports(text string) string {
	// Ensure package declaration exists
	if !strings.Contains(text, "package ") {
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			lines = append([]string{"package unknown"}, lines...)
			text = strings.Join(lines, "\n")
		}
	}

	// Already handled by translator, but ensure no stray imports
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") && !strings.HasPrefix(trimmed, "// import") {
			result = append(result, "// "+line)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// cleanupDuplicates removes obvious duplicated lines and malformed constructs
func cleanupDuplicates(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	seen := make(map[string]int)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip obviously malformed lines
		if strings.HasPrefix(trimmed, "// <skipped>") {
			continue
		}

		// Allow duplicates for common lines (empty, comments)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
			continue
		}

		// De-dup significant lines (but allow methods/functions with same name in different types)
		if !strings.Contains(line, "func") {
			key := trimmed
			seen[key]++
			if seen[key] > 1 {
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// Helper functions

func getIndent(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}

func isFunctionDef(line string) bool {
	return (strings.Contains(line, "func ") && strings.Contains(line, "(")) ||
		strings.HasPrefix(line, "func (")
}

func isDeclaration(line string) bool {
	return strings.HasPrefix(line, "package ") ||
		strings.HasPrefix(line, "import ") ||
		strings.HasPrefix(line, "type ") ||
		strings.HasPrefix(line, "const ") ||
		strings.HasPrefix(line, "var ") ||
		strings.HasPrefix(line, "func ") ||
		strings.HasPrefix(line, "//")
}

func isStatement(line string) bool {
	return strings.HasPrefix(line, "if ") ||
		strings.HasPrefix(line, "for ") ||
		strings.HasPrefix(line, "while ") ||
		strings.HasPrefix(line, "return ") ||
		strings.HasPrefix(line, "switch ") ||
		strings.HasPrefix(line, "case ") ||
		strings.HasPrefix(line, "default") ||
		strings.HasPrefix(line, "break") ||
		strings.HasPrefix(line, "continue") ||
		(strings.Contains(line, "=") && !strings.HasPrefix(line, "//"))
}
