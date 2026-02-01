// Command batch_ai_impl processes essential functions and generates Go implementations using opencode CLI.
//
// Key design decisions learned from previous runs:
//   - Output goes to a staging file (_generated_<name>.go) so duplicates cannot accumulate.
//   - Functions that already exist in the target package are skipped entirely.
//   - The prompt includes a snapshot of existing types/signatures in the target package so the
//     model can use real names instead of inventing stub APIs.
//   - After generation, `go build` is attempted on the target package.  If it fails the error
//     output is fed back to the model for one retry before giving up.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
	// moduleRoot is the root of the Go module (where go.mod lives).  Defaults to two levels up
	// from the output dir when running from the repo root.
	moduleRoot = flag.String("module-root", ".", "Path to Go module root (contains go.mod)")
)

// ---------------------------------------------------------------------------
// Data types
// ---------------------------------------------------------------------------

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

type ProcessResult struct {
	Function string
	Success  bool
	Error    error
	Skipped  bool
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	flag.Parse()

	db, err := loadFunctionDatabase(*functionsDB)
	if err != nil {
		fmt.Printf("Error loading database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d functions from database\n", len(db.Functions))

	functions := filterByTier(db.Functions, *tier)
	fmt.Printf("Processing tier %d: %d functions\n\n", *tier, len(functions))

	if *dryRun {
		fmt.Println("=== DRY RUN - Functions that would be processed ===")
		for i, fn := range functions {
			exists := functionExistsInPackage(fn)
			status := ""
			if exists {
				status = " [SKIP — already exists]"
			}
			fmt.Printf("[%d] Priority %d: %s -> %s/%s%s\n", i+1, fn.Priority, fn.Function, fn.GoPackage, fn.GoFile, status)
			fmt.Printf("    Complexity: %s, Est. tokens: %d\n", fn.Complexity, fn.EstimatedToks)
			fmt.Printf("    Description: %s\n\n", fn.Description)
		}
		return
	}

	results := make(chan ProcessResult, len(functions))
	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup

	fmt.Printf("Using model: %s\n", *model)
	fmt.Printf("Concurrency: %d parallel requests\n\n", *concurrency)

	for i, fn := range functions {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, fn FunctionSpec) {
			defer wg.Done()
			defer func() { <-sem }()

			if *verbose {
				fmt.Printf("[%d/%d] Starting: %s\n", idx+1, len(functions), fn.Function)
			}

			result := processFunction(fn)
			results <- result
		}(i, fn)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	successful, failed, skipped := 0, 0, 0
	var failedFuncs []string

	for result := range results {
		switch {
		case result.Skipped:
			skipped++
			fmt.Printf("⏭  %s (already exists)\n", result.Function)
		case result.Success:
			successful++
			fmt.Printf("✅ %s\n", result.Function)
		default:
			failed++
			failedFuncs = append(failedFuncs, result.Function)
			fmt.Printf("❌ %s: %v\n", result.Function, result.Error)
		}
	}

	fmt.Printf("\n%s\nSummary: %d ok  %d skipped  %d failed  (of %d)\n%s\n",
		strings.Repeat("=", 50), successful, skipped, failed, len(functions), strings.Repeat("=", 50))

	if len(failedFuncs) > 0 {
		fmt.Printf("\nFailed functions:\n")
		for _, fn := range failedFuncs {
			fmt.Printf("   - %s\n", fn)
		}
	}
}

// ---------------------------------------------------------------------------
// Processing pipeline
// ---------------------------------------------------------------------------

func processFunction(fn FunctionSpec) ProcessResult {
	// 1. Skip if the function already exists in the target package.
	if functionExistsInPackage(fn) {
		return ProcessResult{Function: fn.Function, Skipped: true}
	}

	// 2. Extract C++ source for the target function.
	cppCode, err := extractCppCode(filepath.Join(*openlocoSrc, fn.CppFile), fn.Function)
	if err != nil {
		return ProcessResult{Function: fn.Function, Success: false, Error: fmt.Errorf("extract C++: %w", err)}
	}

	// 3. Collect existing package context (types, signatures) for the prompt.
	pkgContext := collectPackageContext(fn.GoPackage)

	// 4. Build prompt and call the model.
	prompt := buildPrompt(fn, cppCode, pkgContext, "")
	implementation, err := callOpencode(prompt)
	if err != nil {
		return ProcessResult{Function: fn.Function, Success: false, Error: fmt.Errorf("opencode: %w", err)}
	}

	// 5. Write to a staging file and attempt to build.
	stagingFile := writeStagingFile(fn, implementation)
	buildErr := tryBuild(fn.GoPackage)
	if buildErr != "" {
		// One retry: feed build errors back to the model.
		if *verbose {
			fmt.Printf("   Build failed for %s, retrying with error feedback...\n", fn.Function)
		}
		prompt = buildPrompt(fn, cppCode, pkgContext, buildErr)
		implementation, err = callOpencode(prompt)
		if err != nil {
			os.Remove(stagingFile)
			return ProcessResult{Function: fn.Function, Success: false, Error: fmt.Errorf("opencode retry: %w", err)}
		}
		stagingFile = writeStagingFile(fn, implementation)
		buildErr = tryBuild(fn.GoPackage)
		if buildErr != "" {
			os.Remove(stagingFile)
			return ProcessResult{Function: fn.Function, Success: false, Error: fmt.Errorf("build after retry: %s", buildErr)}
		}
	}

	return ProcessResult{Function: fn.Function, Success: true}
}

// ---------------------------------------------------------------------------
// Skip-existing check
// ---------------------------------------------------------------------------

// functionExistsInPackage returns true if any .go file in the target package
// already declares a function whose name matches the one in GoSignature.
func functionExistsInPackage(fn FunctionSpec) bool {
	name := extractFunctionName(fn.GoSignature)
	if name == "" {
		return false
	}
	pkgDir := filepath.Join(*outputDir, fn.GoPackage)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		// Skip staging files from previous incomplete runs
		if strings.HasPrefix(e.Name(), "_generated_") {
			continue
		}
		path := filepath.Join(pkgDir, e.Name())
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.Name == name {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Package context snapshot
// ---------------------------------------------------------------------------

// collectPackageContext reads all .go files in the target package and extracts
// type declarations, function signatures, and constants so the model can
// reference real names instead of inventing APIs.
func collectPackageContext(goPackage string) string {
	pkgDir := filepath.Join(*outputDir, goPackage)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return ""
	}

	var buf strings.Builder
	fset := token.NewFileSet()

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		path := filepath.Join(pkgDir, e.Name())
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range f.Decls {
			start := fset.Position(decl.Pos())
			end := fset.Position(decl.End())
			data, readErr := os.ReadFile(start.Filename)
			if readErr != nil {
				continue
			}
			// Convert byte offsets (1-based line/col from token.Position)
			src := string(data)
			lines := strings.Split(src, "\n")
			startLine := start.Line - 1
			endLine := end.Line - 1
			if startLine < 0 {
				startLine = 0
			}
			if endLine >= len(lines) {
				endLine = len(lines) - 1
			}
			snippet := strings.Join(lines[startLine:endLine+1], "\n")
			buf.WriteString(snippet)
			buf.WriteString("\n\n")
		}
	}
	return buf.String()
}

// ---------------------------------------------------------------------------
// Staging file and build validation
// ---------------------------------------------------------------------------

// stagingFileName returns the path for a generated staging file.
// Format: pkg/<package>/_generated_<sanitised-function-name>.go
func stagingFileName(fn FunctionSpec) string {
	safe := strings.NewReplacer("::", "_", ".", "_", " ", "_").Replace(fn.Function)
	return filepath.Join(*outputDir, fn.GoPackage, "_generated_"+safe+".go")
}

// writeStagingFile writes the complete .go file (package decl + imports stub + function)
// to a staging path and returns that path.
func writeStagingFile(fn FunctionSpec, implementation string) string {
	path := stagingFileName(fn)

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("package %s\n\n", fn.GoPackage))
	buf.WriteString("// Code generated by batch_ai_impl; DO NOT EDIT.\n")
	buf.WriteString("// Source: " + fn.CppFile + " " + fn.Function + "\n\n")
	buf.WriteString(fn.GoSignature + " {\n")

	for _, line := range strings.Split(implementation, "\n") {
		if line != "" {
			buf.WriteString("\t" + line + "\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("}\n")

	os.WriteFile(path, []byte(buf.String()), 0644)
	return path
}

// tryBuild runs `go build` on the target package and returns the compiler
// error output as a string, or "" on success.
func tryBuild(goPackage string) string {
	cmd := exec.Command("go", "build", "./pkg/"+goPackage+"/...")
	cmd.Dir = *moduleRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out)
	}
	return ""
}

// ---------------------------------------------------------------------------
// C++ extraction
// ---------------------------------------------------------------------------

func extractCppCode(filePath, functionName string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	simpleName := functionName
	if idx := strings.LastIndex(functionName, "::"); idx >= 0 {
		simpleName = functionName[idx+2:]
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if strings.Contains(line, simpleName) && strings.Contains(line, "(") {
			contextStart := i - 50
			if contextStart < 0 {
				contextStart = 0
			}
			endLine := findFunctionEnd(lines, i)
			if endLine == -1 {
				endLine = i + 100
				if endLine >= len(lines) {
					endLine = len(lines) - 1
				}
			}
			extracted := strings.Join(lines[contextStart:endLine+1], "\n")
			marker := fmt.Sprintf("// TARGET FUNCTION: %s\n// Convert ONLY '%s' to Go.\n\n", functionName, simpleName)
			return marker + extracted, nil
		}
	}

	return "", fmt.Errorf("function %s not found in %s", functionName, filePath)
}

func findFunctionEnd(lines []string, startLine int) int {
	braceCount := 0
	inFunction := false
	for i := startLine; i < len(lines) && i-startLine <= 200; i++ {
		for _, ch := range lines[i] {
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
	}
	return -1
}

// ---------------------------------------------------------------------------
// Prompt building
// ---------------------------------------------------------------------------

func buildPrompt(fn FunctionSpec, cppCode, pkgContext, buildErrors string) string {
	var p strings.Builder

	p.WriteString("# Port OpenLoco C++ Function to Go\n\n")

	if buildErrors != "" {
		p.WriteString("## BUILD ERRORS FROM PREVIOUS ATTEMPT\n")
		p.WriteString("Your previous output produced these compiler errors. Fix them.\n\n")
		p.WriteString("```\n" + buildErrors + "```\n\n")
	}

	p.WriteString("## Task\n")
	p.WriteString("Implement the function below in idiomatic Go for the goloco project.\n\n")

	p.WriteString("## Go Function Signature\n```go\n")
	p.WriteString(fn.GoSignature + "\n```\n\n")

	p.WriteString("## Original C++ Implementation\n```cpp\n")
	p.WriteString(cppCode + "\n```\n\n")

	p.WriteString("## Description\n")
	p.WriteString(fn.Description + "\n\n")

	if pkgContext != "" {
		p.WriteString("## Existing Package Context\n")
		p.WriteString("These types and functions already exist in the target package.\n")
		p.WriteString("Use them — do NOT invent replacements.\n\n")
		p.WriteString("```go\n" + pkgContext + "\n```\n\n")
	}

	if len(fn.Dependencies) > 0 {
		p.WriteString("## Available Dependencies\n")
		for _, dep := range fn.Dependencies {
			p.WriteString("- `" + dep + "`\n")
		}
		p.WriteString("\n")
	}

	p.WriteString("## Requirements\n")
	p.WriteString("1. Translate the C++ logic to idiomatic Go.\n")
	p.WriteString("2. Use Go naming conventions (UpperCase exports, camelCase internal).\n")
	p.WriteString("3. Return errors with Go's error type where appropriate.\n")
	p.WriteString("4. Output ONLY the function body — the code between `{` and `}`.\n")
	p.WriteString("5. Do NOT include the function signature, package declaration, or import statements.\n")
	p.WriteString("6. Only add comments for non-obvious logic.\n")
	p.WriteString("7. Use ONLY types and functions from the existing package context above.\n")
	p.WriteString("   If something genuinely does not exist, use a simple stub (e.g. `log.Println(\"stub\")`).\n\n")

	p.WriteString("## Output Format\n")
	p.WriteString("Return ONLY the function body. Nothing else.\n")

	return p.String()
}

// ---------------------------------------------------------------------------
// opencode integration
// ---------------------------------------------------------------------------

func callOpencode(prompt string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "opencode-workspace-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	promptFile := filepath.Join(tmpDir, "prompt.md")
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("write prompt: %w", err)
	}

	stubFile := filepath.Join(tmpDir, "output.go")
	os.WriteFile(stubFile, []byte("package main\n\n// WRITE YOUR IMPLEMENTATION HERE\n"), 0644)

	cmd := exec.Command("opencode", "run",
		"-m", *model,
		"--format", "json",
		"Read prompt.md and write the Go function implementation to output.go. Include ONLY the function body, no package declaration or signature.",
		"-f", promptFile)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execute opencode: %w\nOutput: %s", err, string(output))
	}

	// Try to extract from opencode's JSON event stream first.
	if impl, err := parseOpencodeOutput(string(output)); err == nil && impl != "" {
		return impl, nil
	}

	// Fall back to reading the stub file that opencode may have written.
	if data, err := os.ReadFile(stubFile); err == nil {
		cleaned := cleanupGeneratedCode(string(data))
		if cleaned != "" && cleaned != "// WRITE YOUR IMPLEMENTATION HERE" {
			return cleaned, nil
		}
	}

	return "", fmt.Errorf("no code generated")
}

func parseOpencodeOutput(jsonOutput string) (string, error) {
	for _, line := range strings.Split(jsonOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event struct {
			Type string `json:"type"`
			Part struct {
				Tool  string `json:"tool"`
				State struct {
					Input struct {
						Content string `json:"content"`
					} `json:"input"`
				} `json:"state"`
			} `json:"part"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Type == "tool_use" && event.Part.Tool == "write" && event.Part.State.Input.Content != "" {
			return cleanupGeneratedCode(event.Part.State.Input.Content), nil
		}
	}

	// Second pass: look for code in text events.
	for _, line := range strings.Split(jsonOutput, "\n") {
		line = strings.TrimSpace(line)
		var textEvent struct {
			Type string `json:"type"`
			Part struct {
				Text string `json:"text"`
			} `json:"part"`
		}
		if err := json.Unmarshal([]byte(line), &textEvent); err != nil {
			continue
		}
		if textEvent.Type == "text" {
			if code := extractCodeBlock(textEvent.Part.Text); code != "" {
				return code, nil
			}
		}
	}

	return "", fmt.Errorf("no code in output")
}

func extractCodeBlock(text string) string {
	parts := strings.Split(text, "```")
	if len(parts) >= 3 {
		code := parts[1]
		if strings.HasPrefix(code, "go\n") {
			code = strings.TrimPrefix(code, "go\n")
		}
		return strings.TrimSpace(code)
	}
	return ""
}

func cleanupGeneratedCode(code string) string {
	var lines []string
	for _, line := range strings.Split(code, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func extractFunctionName(signature string) string {
	s := strings.TrimPrefix(signature, "func ")
	// Handle method receivers: func (r *Foo) Bar(...)
	if strings.HasPrefix(s, "(") {
		if idx := strings.Index(s, ") "); idx >= 0 {
			s = s[idx+2:]
		}
	}
	if idx := strings.Index(s, "("); idx > 0 {
		return s[:idx]
	}
	return ""
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
	var out []FunctionSpec
	for _, fn := range functions {
		if fn.Tier == tier {
			out = append(out, fn)
		}
	}
	return out
}
