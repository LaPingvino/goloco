// Command batch_ai_impl processes essential functions and generates Go implementations using opencode CLI
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	functionsDB = flag.String("db", "essential_functions.json", "Path to functions database")
	openlocoSrc = flag.String("src", "../OpenLoco/src/OpenLoco/src", "Path to OpenLoco C++ source")
	outputDir   = flag.String("out", "pkg", "Output directory for generated Go code")
	tier        = flag.Int("tier", 1, "Which tier to process (1-5, 0 for all)")
	dryRun      = flag.Bool("dry", false, "Dry run - show what would be done")
	concurrency = flag.Int("concurrent", 3, "Number of concurrent opencode requests")
	model       = flag.String("model", "github-copilot/gpt-5-mini", "Model to use")
	verbose     = flag.Bool("v", false, "Verbose output")
)

type FunctionSpec struct {
	ID            string   `json:"id"`
	Priority      int      `json:"priority"`
	Tier          int      `json:"tier"`
	CppFile       string   `json:"cppFile"`
	Function      string   `json:"function"`
	Signature     string   `json:"signature"`
	GoSignature   string   `json:"goSignature"`
	GoPackage     string   `json:"goPackage"`
	GoFile        string   `json:"goFile"`
	Dependencies  []string `json:"dependencies"`
	Complexity    string   `json:"complexity"`
	EstimatedToks int      `json:"estimatedTokens"`
	Description   string   `json:"description"`
}

type FunctionDatabase struct {
	Functions []FunctionSpec `json:"functions"`
}

func main() {
	flag.Parse()

	// Load functions database
	db, err := loadFunctionDatabase(*functionsDB)
	if err != nil {
		fmt.Printf("Error loading database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📚 Loaded %d functions from database\n", len(db.Functions))

	// Filter by tier
	functions := filterByTier(db.Functions, *tier)
	fmt.Printf("🎯 Processing tier %d: %d functions\n\n", *tier, len(functions))

	if *dryRun {
		fmt.Println("=== DRY RUN - Functions that would be processed ===")
		for i, fn := range functions {
			fmt.Printf("[%d] Priority %d: %s -> %s/%s\n", i+1, fn.Priority, fn.Function, fn.GoPackage, fn.GoFile)
			fmt.Printf("    Complexity: %s, Est. tokens: %d\n", fn.Complexity, fn.EstimatedToks)
			fmt.Printf("    Description: %s\n\n", fn.Description)
		}
		return
	}

	// Process functions with concurrency control
	results := make(chan ProcessResult, len(functions))
	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup

	fmt.Printf("🤖 Using model: %s\n", *model)
	fmt.Printf("⚡ Concurrency: %d parallel requests\n\n", *concurrency)

	for i, fn := range functions {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(idx int, fn FunctionSpec) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			if *verbose {
				fmt.Printf("🔄 [%d/%d] Starting: %s\n", idx+1, len(functions), fn.Function)
			}

			result := processFunction(fn, *openlocoSrc, *outputDir)
			results <- result
		}(i, fn)
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	successful := 0
	failed := 0
	failedFuncs := []string{}

	for result := range results {
		if result.Success {
			successful++
			fmt.Printf("✅ %s\n", result.Function)
		} else {
			failed++
			failedFuncs = append(failedFuncs, result.Function)
			fmt.Printf("❌ %s: %v\n", result.Function, result.Error)
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("📊 Summary\n")
	fmt.Printf(strings.Repeat("=", 50) + "\n")
	fmt.Printf("✅ Successful: %d/%d\n", successful, len(functions))
	fmt.Printf("❌ Failed: %d/%d\n", failed, len(functions))

	if len(failedFuncs) > 0 {
		fmt.Printf("\n⚠️  Failed functions:\n")
		for _, fn := range failedFuncs {
			fmt.Printf("   - %s\n", fn)
		}
	}

	if successful > 0 {
		fmt.Printf("\n✨ Generated code in: %s/\n", *outputDir)
		fmt.Printf("🔨 Next step: go build ./pkg/...\n")
	}
}

type ProcessResult struct {
	Function string
	Success  bool
	Error    error
}

func processFunction(fn FunctionSpec, srcDir, outDir string) ProcessResult {
	// 1. Extract C++ implementation
	cppCode, err := extractCppCode(filepath.Join(srcDir, fn.CppFile), fn.Function)
	if err != nil {
		return ProcessResult{fn.Function, false, fmt.Errorf("extract C++: %w", err)}
	}

	// 2. Build prompt for opencode
	prompt := buildPrompt(fn, cppCode, "")

	// 3. Call opencode CLI with retry logic
	var implementation string
	var callErr error
	maxRetries := 2

	for attempt := 0; attempt <= maxRetries; attempt++ {
		implementation, callErr = callOpencode(prompt)
		if callErr != nil {
			return ProcessResult{fn.Function, false, fmt.Errorf("opencode: %w", callErr)}
		}

		// Validate the implementation
		validationError := validateImplementation(implementation, fn)
		if validationError == "" {
			// Success!
			break
		}

		// If this was the last attempt, fail
		if attempt == maxRetries {
			return ProcessResult{fn.Function, false, fmt.Errorf("validation failed after %d attempts: %s", maxRetries+1, validationError)}
		}

		// Retry with feedback
		if *verbose {
			fmt.Printf("   ⚠️  Attempt %d failed validation, retrying with feedback...\n", attempt+1)
		}
		prompt = buildPrompt(fn, cppCode, validationError)
	}

	// 4. Write Go file
	err = writeGoFile(fn, implementation, outDir)
	if err != nil {
		return ProcessResult{fn.Function, false, fmt.Errorf("write file: %w", err)}
	}

	return ProcessResult{fn.Function, true, nil}
}

// validateImplementation checks if the generated code meets basic requirements
func validateImplementation(code string, fn FunctionSpec) string {
	trimmed := strings.TrimSpace(code)

	if trimmed == "" {
		return "Generated code is empty"
	}

	if strings.Contains(trimmed, "WRITE YOUR IMPLEMENTATION HERE") {
		return "Generated code contains placeholder text"
	}

	// Check if it's just a comment
	lines := strings.Split(trimmed, "\n")
	nonCommentLines := 0
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l != "" && !strings.HasPrefix(l, "//") {
			nonCommentLines++
		}
	}

	if nonCommentLines == 0 {
		return "Generated code contains only comments"
	}

	// Check if it contains function signature (should be body only)
	if strings.Contains(trimmed, "func "+extractFunctionName(fn.GoSignature)) {
		return "Generated code includes function signature - should be body only (code between { and })"
	}

	// Check if it contains package declaration
	if strings.HasPrefix(trimmed, "package ") {
		return "Generated code includes package declaration - should be body only"
	}

	return ""
}

func extractFunctionName(signature string) string {
	// Extract function name from signature like "func GetShade(..."
	if strings.HasPrefix(signature, "func ") {
		rest := strings.TrimPrefix(signature, "func ")
		if idx := strings.Index(rest, "("); idx > 0 {
			return rest[:idx]
		}
	}
	return ""
}

func callOpencode(prompt string) (string, error) {
	// Create an isolated temporary directory for this specific opencode call
	// This prevents opencode from accessing or modifying the main project
	tmpDir, err := os.MkdirTemp("", "opencode-workspace-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the prompt to a file in the isolated directory
	promptFile := filepath.Join(tmpDir, "prompt.md")
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("write prompt: %w", err)
	}

	// Create a stub Go file where opencode can write the output
	stubFile := filepath.Join(tmpDir, "output.go")
	stubContent := `package main

// WRITE YOUR IMPLEMENTATION HERE
`
	if err := os.WriteFile(stubFile, []byte(stubContent), 0644); err != nil {
		return "", fmt.Errorf("write stub: %w", err)
	}

	// Call opencode with isolated workspace
	// Message must come BEFORE -f flag (positional arguments come first)
	cmd := exec.Command("opencode", "run",
		"-m", *model,
		"--format", "json",
		"Read prompt.md and write the Go function implementation to output.go. Include ONLY the function implementation, no package declaration.",
		"-f", promptFile)

	// Set working directory to the isolated temp dir
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execute opencode: %w\nOutput: %s", err, string(output))
	}

	// First try to parse from JSON output (opencode returns code as text)
	implementation, err := parseOpencodeOutput(string(output))
	if err == nil && implementation != "" {
		return implementation, nil
	}

	// Fallback: check if opencode wrote to the file
	generatedCode, readErr := os.ReadFile(stubFile)
	if readErr == nil {
		cleaned := cleanupGeneratedCode(string(generatedCode))
		if cleaned != "" && cleaned != "// WRITE YOUR IMPLEMENTATION HERE" {
			return cleaned, nil
		}
	}

	// Neither method worked
	if err != nil {
		return "", fmt.Errorf("parse JSON: %w", err)
	}
	return "", fmt.Errorf("no code generated")
}

func parseOpencodeOutput(jsonOutput string) (string, error) {
	// opencode outputs JSON events, we need to extract the code from tool_use events
	lines := strings.Split(jsonOutput, "\n")

	var extractedCode string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event struct {
			Type string `json:"type"`
			Part struct {
				Type  string `json:"type"`
				Tool  string `json:"tool"`
				State struct {
					Input struct {
						Content string `json:"content"`
					} `json:"input"`
				} `json:"state"`
			} `json:"part"`
		}

		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip non-JSON lines
		}

		// Look for tool_use events with write tool (opencode generates files)
		if event.Type == "tool_use" && event.Part.Tool == "write" {
			content := event.Part.State.Input.Content
			if content != "" {
				extractedCode = content
				break
			}
		}
	}

	if extractedCode == "" {
		// Fallback: try to extract from text events
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var textEvent struct {
				Type string `json:"type"`
				Part struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"part"`
			}

			if err := json.Unmarshal([]byte(line), &textEvent); err != nil {
				continue
			}

			if textEvent.Type == "text" && textEvent.Part.Text != "" {
				text := textEvent.Part.Text

				// Check if text looks like code (starts with { or has code-like patterns)
				trimmed := strings.TrimSpace(text)
				if strings.HasPrefix(trimmed, "{") ||
				   (strings.Contains(trimmed, "\n\t") && (strings.Contains(trimmed, "return") || strings.Contains(trimmed, "func"))) {
					// This looks like raw code
					extractedCode = text
					break
				}

				// Try to extract code blocks from markdown
				if strings.Contains(text, "```") {
					extracted := extractCodeBlock(text)
					if extracted != "" {
						extractedCode = extracted
						break
					}
				}
			}
		}
	}

	if extractedCode == "" {
		return "", fmt.Errorf("no code generated in output")
	}

	// Clean up the code
	code := cleanupGeneratedCode(extractedCode)
	return code, nil
}

func extractCodeBlock(text string) string {
	// Extract code from markdown code blocks
	parts := strings.Split(text, "```")
	if len(parts) >= 3 {
		code := parts[1]
		// Remove language identifier if present
		if strings.HasPrefix(code, "go\n") {
			code = strings.TrimPrefix(code, "go\n")
		}
		return strings.TrimSpace(code)
	}
	return ""
}

func cleanupGeneratedCode(code string) string {
	// Remove package declaration if present (we'll add it ourselves)
	lines := strings.Split(code, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip package declarations
		if strings.HasPrefix(trimmed, "package ") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}

	code = strings.Join(cleanedLines, "\n")
	code = strings.TrimSpace(code)

	return code
}

func extractCppCode(filePath, functionName string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Try without src prefix
		altPath := strings.Replace(filePath, "src/OpenLoco/src/", "", 1)
		content, err = os.ReadFile(altPath)
		if err != nil {
			return "", err
		}
	}

	lines := strings.Split(string(content), "\n")

	// Extract just the simple function name (e.g., "getShade" from "Colour::getShade")
	simpleName := functionName
	if strings.Contains(functionName, "::") {
		parts := strings.Split(functionName, "::")
		simpleName = parts[len(parts)-1]
	}

	// Find the function - be liberal with matching
	// Look for the simple name followed by an opening parenthesis
	for i, line := range lines {
		// Skip pure comment lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Look for function name + opening paren (function definition or call)
		if strings.Contains(line, simpleName) && strings.Contains(line, "(") {
			// Extract generous context: 50 lines before to help LLM understand context
			contextStart := i - 50
			if contextStart < 0 {
				contextStart = 0
			}

			// Find where the function ends (matching braces)
			endLine := findFunctionEnd(lines, i)
			if endLine == -1 {
				// If we can't find the end, take a generous chunk
				endLine = i + 100
				if endLine >= len(lines) {
					endLine = len(lines) - 1
				}
			}

			// Extract everything from context start to function end
			// This includes namespace declarations, helper functions, constants, etc.
			// The LLM will filter out what's not needed
			contextLines := lines[contextStart : endLine+1]
			extracted := strings.Join(contextLines, "\n")

			// Add a marker to help the LLM identify the target function
			marker := fmt.Sprintf("// TARGET FUNCTION: %s\n// Look for the function '%s' in the code below.\n// Convert ONLY that function to Go.\n\n", functionName, simpleName)

			return marker + extracted, nil
		}
	}

	return "", fmt.Errorf("function %s not found in %s", functionName, filePath)
}

// findFunctionEnd finds the closing brace of a function starting at startLine
func findFunctionEnd(lines []string, startLine int) int {
	braceCount := 0
	inFunction := false

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		for _, ch := range line {
			if ch == '{' {
				braceCount++
				inFunction = true
			} else if ch == '}' {
				braceCount--
				if inFunction && braceCount == 0 {
					return i
				}
			}
		}

		// Safety limit - don't go beyond 200 lines
		if i-startLine > 200 {
			return i
		}
	}

	return -1
}



func buildPrompt(fn FunctionSpec, cppCode string, feedback string) string {
	var prompt strings.Builder

	prompt.WriteString("# Port OpenLoco C++ Function to Go\n\n")

	// If there's feedback from a previous attempt, include it prominently
	if feedback != "" {
		prompt.WriteString("## ⚠️ IMPORTANT FEEDBACK FROM PREVIOUS ATTEMPT\n")
		prompt.WriteString("Your previous response had this issue:\n")
		prompt.WriteString(fmt.Sprintf("**%s**\n\n", feedback))
		prompt.WriteString("Please fix this in your response.\n\n")
	}

	prompt.WriteString("## Task\n")
	prompt.WriteString(fmt.Sprintf("Implement this function in idiomatic Go for the goloco project.\n\n"))

	prompt.WriteString("## Go Function Signature\n```go\n")
	prompt.WriteString(fn.GoSignature + "\n")
	prompt.WriteString("```\n\n")

	prompt.WriteString("## Original C++ Implementation\n```cpp\n")
	prompt.WriteString(cppCode + "\n")
	prompt.WriteString("```\n\n")

	prompt.WriteString("## Description\n")
	prompt.WriteString(fn.Description + "\n\n")

	if len(fn.Dependencies) > 0 {
		prompt.WriteString("## Available Dependencies (Already Implemented)\n")
		for _, dep := range fn.Dependencies {
			prompt.WriteString(fmt.Sprintf("- `%s`\n", dep))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("## Requirements\n")
	prompt.WriteString("1. Translate the C++ logic to idiomatic Go\n")
	prompt.WriteString("2. Use Go naming conventions (UpperCase exports, camelCase internal)\n")
	prompt.WriteString("3. Handle errors with Go's error type where appropriate\n")
	prompt.WriteString("4. Output ONLY the function body (code between `{` and `}`)\n")
	prompt.WriteString("5. Do NOT include the function signature or package declaration\n")
	prompt.WriteString("6. Add brief comments only for complex/non-obvious logic\n")
	prompt.WriteString("7. If types don't exist, use reasonable stub types (we'll define them later)\n\n")

	prompt.WriteString("## Output Format\n")
	prompt.WriteString("Return ONLY the Go function body. No signature, no package, no explanations.\n")
	prompt.WriteString("Just the code that goes between the opening `{` and closing `}`.\n")

	if feedback != "" {
		prompt.WriteString("\n**REMINDER: Fix the issue mentioned in the feedback above!**\n")
	}

	return prompt.String()
}

func writeGoFile(fn FunctionSpec, implementation, outDir string) error {
	// Create package directory
	pkgDir := filepath.Join(outDir, fn.GoPackage)
	os.MkdirAll(pkgDir, 0755)

	filePath := filepath.Join(pkgDir, fn.GoFile)

	// Check if file exists
	var existingContent string
	if content, err := os.ReadFile(filePath); err == nil {
		existingContent = string(content)
	}

	// Build complete function
	var fileContent strings.Builder

	if existingContent == "" {
		// New file
		fileContent.WriteString(fmt.Sprintf("package %s\n\n", fn.GoPackage))
		fileContent.WriteString("// Auto-generated by batch_ai_impl using opencode + GPT-5-mini\n")
		fileContent.WriteString("// Original: " + fn.CppFile + "\n\n")
	} else {
		// Append to existing
		fileContent.WriteString(existingContent)
		fileContent.WriteString("\n\n")
	}

	fileContent.WriteString("// " + fn.Description + "\n")
	fileContent.WriteString("// Source: " + fn.Function + "\n")
	fileContent.WriteString(fn.GoSignature + " {\n")

	// Indent the implementation
	lines := strings.Split(implementation, "\n")
	for _, line := range lines {
		if line != "" {
			fileContent.WriteString("\t" + line + "\n")
		} else {
			fileContent.WriteString("\n")
		}
	}

	fileContent.WriteString("}\n")

	return os.WriteFile(filePath, []byte(fileContent.String()), 0644)
}

func loadFunctionDatabase(path string) (*FunctionDatabase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var db FunctionDatabase
	err = json.Unmarshal(data, &db)
	return &db, err
}

func filterByTier(functions []FunctionSpec, tier int) []FunctionSpec {
	if tier == 0 {
		return functions
	}

	filtered := make([]FunctionSpec, 0)
	for _, fn := range functions {
		if fn.Tier == tier {
			filtered = append(filtered, fn)
		}
	}
	return filtered
}
