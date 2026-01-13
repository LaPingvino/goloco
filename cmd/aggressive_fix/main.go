package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AggressivePostProcessor fixes structural Go syntax errors from C++ conversion
// This is a second-pass that handles more complex issues
var (
	inDir  = flag.String("in", "", "Directory with converted Go files to fix")
	outDir = flag.String("out", "", "Output directory for fixed files")
)

func main() {
	flag.Parse()
	if *inDir == "" || *outDir == "" {
		fmt.Println("Usage: aggressive_fix -in <dir> -out <dir>")
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
	fmt.Println("Aggressive post-processing complete")
}

func processFile(srcPath, outDir string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	text := string(content)
	text = removeNestedStatements(text)
	text = fixStructSyntax(text)
	text = fixFunctionSyntax(text)
	text = cleanupGarbage(text)
	text = ensureValidGoStructure(text)

	rel, _ := filepath.Rel(*inDir, srcPath)
	outPath := filepath.Join(outDir, rel)
	os.MkdirAll(filepath.Dir(outPath), 0755)

	return os.WriteFile(outPath, []byte(text), 0644)
}

// removeNestedStatements removes orphaned if/for/return statements that aren't in functions
func removeNestedStatements(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inFunc := false
	inStruct := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track struct/type context
		if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, "struct") {
			inStruct = true
			if strings.Contains(line, "{") {
				braceDepth = 1
			}
		}

		// Track function context
		if (strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "func (")) && strings.Contains(line, "{") {
			inFunc = true
			braceDepth = 1
		}

		if inFunc || inStruct {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth <= 0 {
				inFunc = false
				inStruct = false
			}
		}

		// Skip orphaned statements at module level
		if !inFunc && !inStruct {
			if isOrphanedStatement(trimmed) {
				result = append(result, "// REMOVED ORPHANED: "+trimmed)
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// fixStructSyntax ensures struct fields are properly formatted
func fixStructSyntax(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inStruct := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, "struct") && strings.Contains(line, "{") {
			inStruct = true
			braceDepth = 1
		}

		if inStruct {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			// Inside struct: validate field syntax
			if braceDepth > 0 && braceDepth <= 1 && !strings.Contains(trimmed, "type") && !strings.Contains(trimmed, "//") && trimmed != "{" && trimmed != "}" {
				// Should be: fieldname Type
				if isMalformedStructField(trimmed) {
					// Skip it or comment it out
					result = append(result, "// MALFORMED FIELD: "+line)
					continue
				}
			}

			if braceDepth <= 0 {
				inStruct = false
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// fixFunctionSyntax ensures function declarations are valid
func fixFunctionSyntax(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Fix constructor-like declarations that aren't Go methods
		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") && !strings.HasPrefix(trimmed, "func") && !strings.HasPrefix(trimmed, "//") {
			if looksLikeBadConstructor(trimmed) {
				result = append(result, "// SKIPPED CONSTRUCTOR: "+line)
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// cleanupGarbage removes obviously malformed content
func cleanupGarbage(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip lines that are clearly incomplete conversions
		if strings.HasPrefix(trimmed, "// <skipped>") {
			continue
		}

		// Skip lines with C++ syntax that didn't convert
		if strings.Contains(trimmed, "::") && !strings.HasPrefix(trimmed, "//") {
			result = append(result, "// SKIPPED C++ SYNTAX: "+line)
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// ensureValidGoStructure ensures the file has valid Go structure
func ensureValidGoStructure(text string) string {
	lines := strings.Split(text, "\n")

	// Ensure package declaration
	hasPkg := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			hasPkg = true
			break
		}
	}

	if !hasPkg && len(lines) > 0 {
		lines = append([]string{"package unknown"}, lines...)
	}

	// Remove trailing orphaned braces
	for len(lines) > 0 {
		lastTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastTrimmed == "}" || lastTrimmed == "" {
			lines = lines[:len(lines)-1]
		} else {
			break
		}
	}

	return strings.Join(lines, "\n")
}

// Helper functions

func isOrphanedStatement(line string) bool {
	return (strings.HasPrefix(line, "if ") ||
		strings.HasPrefix(line, "for ") ||
		strings.HasPrefix(line, "while ") ||
		strings.HasPrefix(line, "return ") ||
		strings.HasPrefix(line, "switch ") ||
		strings.HasPrefix(line, "case ") ||
		strings.HasPrefix(line, "default") ||
		strings.HasPrefix(line, "break") ||
		strings.HasPrefix(line, "continue")) &&
		!strings.HasPrefix(line, "//")
}

func isMalformedStructField(line string) bool {
	// A field should be: Name Type [optional comment]
	// Skip if it looks like C++ method or has C++ syntax
	return strings.Contains(line, "(") ||
		strings.Contains(line, "const ") ||
		strings.Contains(line, "::") ||
		strings.HasPrefix(line, "var ") ||
		strings.HasPrefix(line, "func ") ||
		strings.HasPrefix(line, "if ") ||
		strings.HasPrefix(line, "return ")
}

func looksLikeBadConstructor(line string) bool {
	// Constructor names match struct names: StructName(params)
	// In Go, we use NewStructName() functions
	return !strings.Contains(line, "func ") &&
		!strings.HasPrefix(line, "//") &&
		strings.Contains(line, "(") &&
		strings.HasSuffix(strings.TrimSpace(line), ")")
}
