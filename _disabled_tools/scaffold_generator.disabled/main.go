// Command scaffold_generator creates clean Go scaffolds from converted C++ code
// It extracts types, generates function stubs with descriptions, and can optionally
// use AI to generate initial implementations
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	inputDir      = flag.String("in", "", "Input directory with converted code")
	outputDir     = flag.String("out", "", "Output directory for clean scaffolds")
	generateImpls = flag.Bool("ai", false, "Use AI to generate initial implementations")
	aiModel       = flag.String("model", "gpt-4o-mini", "AI model to use for generation")
	verbose       = flag.Bool("v", false, "Verbose output")
)

type FunctionSignature struct {
	Name        string
	Receiver    string // empty for package-level functions
	Params      []Parameter
	Returns     []string
	Description string
	CppCode     string // Original C++ for reference
}

type Parameter struct {
	Name string
	Type string
}

type TypeDefinition struct {
	Name   string
	Kind   string // "struct", "enum", "const", "var"
	Fields []Field
	Body   string
}

type Field struct {
	Name string
	Type string
}

type ScaffoldFile struct {
	Package    string
	Imports    []string
	Types      []TypeDefinition
	Functions  []FunctionSignature
	SourceFile string
}

func main() {
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Usage: scaffold_generator -in <converted_dir> -out <scaffold_dir> [-ai] [-model <model>]")
		os.Exit(1)
	}

	generator := NewScaffoldGenerator(*inputDir, *outputDir, *verbose)

	if err := generator.ProcessDirectory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated scaffolds in %s\n", *outputDir)

	if *generateImpls {
		fmt.Println("\n🤖 Generating AI implementations...")
		if err := generator.GenerateAIImplementations(*aiModel); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: AI generation failed: %v\n", err)
		}
	}
}

type ScaffoldGenerator struct {
	inputDir  string
	outputDir string
	verbose   bool
	scaffolds map[string]*ScaffoldFile
}

func NewScaffoldGenerator(inputDir, outputDir string, verbose bool) *ScaffoldGenerator {
	return &ScaffoldGenerator{
		inputDir:  inputDir,
		outputDir: outputDir,
		verbose:   verbose,
		scaffolds: make(map[string]*ScaffoldFile),
	}
}

func (sg *ScaffoldGenerator) ProcessDirectory() error {
	return filepath.Walk(sg.inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		return sg.processFile(path)
	})
}

func (sg *ScaffoldGenerator) processFile(path string) error {
	if sg.verbose {
		fmt.Printf("Processing: %s\n", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	scaffold := sg.extractScaffold(string(content), path)

	// Determine output path
	rel, _ := filepath.Rel(sg.inputDir, path)
	outPath := filepath.Join(sg.outputDir, rel)

	sg.scaffolds[outPath] = scaffold

	return sg.writeScaffold(scaffold, outPath)
}

var (
	packageRe       = regexp.MustCompile(`^package\s+(\w+)`)
	typeRe          = regexp.MustCompile(`^type\s+(\w+)\s+(struct|int|uint8|uint16|uint32)`)
	constRe         = regexp.MustCompile(`^const\s+\(`)
	enumRe          = regexp.MustCompile(`^\s*(\w+)(?:\s+\w+)?\s*=?\s*(?:iota|1\s*<<|\d+)`)
	funcRe          = regexp.MustCompile(`^func\s+(?:\((\w+)\s+\*(\w+)\)\s+)?(\w+)\(([^)]*)\)\s*(.*)`)
	commentMethodRe = regexp.MustCompile(`^\s*//\s*method:\s*(.+)`)
	commentFuncRe   = regexp.MustCompile(`^\s*//\s*func\s+(.+)`)
)

func (sg *ScaffoldGenerator) extractScaffold(content, sourcePath string) *ScaffoldFile {
	scaffold := &ScaffoldFile{
		SourceFile: sourcePath,
		Types:      make([]TypeDefinition, 0),
		Functions:  make([]FunctionSignature, 0),
	}

	lines := strings.Split(content, "\n")
	inStruct := false
	inConst := false
	inCppComment := false
	currentType := &TypeDefinition{}
	currentFunc := &FunctionSignature{}
	cppCodeBuffer := strings.Builder{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract package name
		if m := packageRe.FindStringSubmatch(trimmed); m != nil {
			scaffold.Package = m[1]
			continue
		}

		// Skip AUTO-GENERATED comments
		if strings.Contains(trimmed, "AUTO-GENERATED") || strings.Contains(trimmed, "WILL NOT COMPILE") {
			continue
		}

		// Detect C++ comment blocks
		if strings.HasPrefix(trimmed, "// #include") || strings.HasPrefix(trimmed, "// namespace") {
			inCppComment = true
			continue
		}
		if inCppComment && !strings.HasPrefix(trimmed, "//") {
			inCppComment = false
		}
		if inCppComment {
			cppCodeBuffer.WriteString(trimmed[3:] + "\n") // strip "// "
			continue
		}

		// Extract type definitions
		if m := typeRe.FindStringSubmatch(trimmed); m != nil {
			if inStruct {
				scaffold.Types = append(scaffold.Types, *currentType)
			}
			currentType = &TypeDefinition{
				Name:   m[1],
				Kind:   m[2],
				Fields: make([]Field, 0),
			}
			if m[2] == "struct" {
				inStruct = true
			} else {
				inStruct = false
				scaffold.Types = append(scaffold.Types, *currentType)
			}
			continue
		}

		// Extract struct fields
		if inStruct && !strings.Contains(trimmed, "}") && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				currentType.Fields = append(currentType.Fields, Field{
					Name: parts[0],
					Type: parts[1],
				})
			}
		}

		// End of struct
		if inStruct && trimmed == "}" {
			scaffold.Types = append(scaffold.Types, *currentType)
			inStruct = false
		}

		// Extract commented method signatures
		if m := commentMethodRe.FindStringSubmatch(trimmed); m != nil {
			sig := sg.parseMethodSignature(m[1], cppCodeBuffer.String())
			if sig != nil {
				scaffold.Functions = append(scaffold.Functions, *sig)
				cppCodeBuffer.Reset()
			}
			continue
		}

		// Extract commented function signatures
		if m := commentFuncRe.FindStringSubmatch(trimmed); m != nil {
			sig := sg.parseFunctionSignature(m[1], cppCodeBuffer.String())
			if sig != nil {
				scaffold.Functions = append(scaffold.Functions, *sig)
				cppCodeBuffer.Reset()
			}
			continue
		}

		// Extract actual function definitions (rare in converted code)
		if m := funcRe.FindStringSubmatch(trimmed); m != nil {
			sig := &FunctionSignature{
				Name:     m[3],
				Receiver: m[2],
			}
			// Parse params and returns...
			scaffold.Functions = append(scaffold.Functions, *sig)
		}
	}

	return scaffold
}

func (sg *ScaffoldGenerator) parseMethodSignature(cppSig, cppCode string) *FunctionSignature {
	// Parse C++ method signature: "bool validate() const"
	// Extract description from preceding C++ code if available

	sig := &FunctionSignature{
		CppCode: cppCode,
	}

	// Simple regex to extract method parts
	methodRe := regexp.MustCompile(`(\w+(?:<[^>]+>)?)\s+(\w+)\s*\(([^)]*)\)`)
	if m := methodRe.FindStringSubmatch(cppSig); m != nil {
		returnType := sg.translateType(m[1])
		sig.Name = exportName(m[2])
		sig.Returns = []string{returnType}

		// Parse parameters
		if m[3] != "" {
			sig.Params = sg.parseParams(m[3])
		}

		// Generate description from C++ code
		sig.Description = sg.generateDescription(sig.Name, cppCode)
	}

	return sig
}

func (sg *ScaffoldGenerator) parseFunctionSignature(cppSig, cppCode string) *FunctionSignature {
	sig := &FunctionSignature{
		CppCode: cppCode,
	}

	// Similar to method parsing
	funcRe := regexp.MustCompile(`(\w+(?:<[^>]+>)?)\s+(\w+)\s*\(([^)]*)\)`)
	if m := funcRe.FindStringSubmatch(cppSig); m != nil {
		returnType := sg.translateType(m[1])
		sig.Name = exportName(m[2])
		sig.Returns = []string{returnType}

		if m[3] != "" {
			sig.Params = sg.parseParams(m[3])
		}

		sig.Description = sg.generateDescription(sig.Name, cppCode)
	}

	return sig
}

func (sg *ScaffoldGenerator) parseParams(paramStr string) []Parameter {
	params := make([]Parameter, 0)
	parts := strings.Split(paramStr, ",")

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Simple pattern: "Type name" or just "Type"
		tokens := strings.Fields(p)
		if len(tokens) >= 2 {
			params = append(params, Parameter{
				Name: tokens[len(tokens)-1],
				Type: sg.translateType(strings.Join(tokens[:len(tokens)-1], " ")),
			})
		} else if len(tokens) == 1 {
			params = append(params, Parameter{
				Name: "arg",
				Type: sg.translateType(tokens[0]),
			})
		}
	}

	return params
}

func (sg *ScaffoldGenerator) translateType(cppType string) string {
	cppType = strings.TrimSpace(cppType)
	cppType = strings.ReplaceAll(cppType, "const ", "")
	cppType = strings.TrimSuffix(cppType, "&")
	cppType = strings.TrimSuffix(cppType, "*")
	cppType = strings.TrimSpace(cppType)

	typeMap := map[string]string{
		"void":     "",
		"bool":     "bool",
		"int":      "int",
		"int8_t":   "int8",
		"int16_t":  "int16",
		"int32_t":  "int32",
		"int64_t":  "int64",
		"uint8_t":  "uint8",
		"uint16_t": "uint16",
		"uint32_t": "uint32",
		"uint64_t": "uint64",
		"size_t":   "int",
		"float":    "float32",
		"double":   "float64",
	}

	if goType, ok := typeMap[cppType]; ok {
		return goType
	}

	return cppType
}

func (sg *ScaffoldGenerator) generateDescription(funcName, cppCode string) string {
	if cppCode == "" {
		return fmt.Sprintf("TODO: Implement %s", funcName)
	}

	// Extract meaningful lines from C++ code
	lines := strings.Split(cppCode, "\n")
	meaningful := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Skip namespace/include lines
		if strings.Contains(line, "namespace") || strings.Contains(line, "include") {
			continue
		}
		meaningful = append(meaningful, line)
		if len(meaningful) >= 5 {
			break
		}
	}

	if len(meaningful) == 0 {
		return fmt.Sprintf("TODO: Implement %s", funcName)
	}

	return fmt.Sprintf("TODO: Implement %s\nC++ Logic:\n// %s", funcName, strings.Join(meaningful, "\n// "))
}

func (sg *ScaffoldGenerator) writeScaffold(scaffold *ScaffoldFile, outPath string) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write package declaration
	fmt.Fprintf(w, "package %s\n\n", scaffold.Package)
	fmt.Fprintf(w, "// Auto-generated scaffold from: %s\n", filepath.Base(scaffold.SourceFile))
	fmt.Fprintf(w, "// This file contains clean type definitions and function stubs\n\n")

	// Write imports (TODO: detect needed imports)
	if len(scaffold.Imports) > 0 {
		fmt.Fprintf(w, "import (\n")
		for _, imp := range scaffold.Imports {
			fmt.Fprintf(w, "\t\"%s\"\n", imp)
		}
		fmt.Fprintf(w, ")\n\n")
	}

	// Write type definitions
	for _, td := range scaffold.Types {
		switch td.Kind {
		case "struct":
			fmt.Fprintf(w, "type %s struct {\n", td.Name)
			for _, field := range td.Fields {
				fmt.Fprintf(w, "\t%s %s\n", field.Name, field.Type)
			}
			fmt.Fprintf(w, "}\n\n")
		case "int", "uint8", "uint16", "uint32":
			fmt.Fprintf(w, "type %s %s\n\n", td.Name, td.Kind)
		}
	}

	// Write function stubs
	for _, fn := range scaffold.Functions {
		sg.writeFunctionStub(w, &fn)
	}

	return nil
}

func (sg *ScaffoldGenerator) writeFunctionStub(w *bufio.Writer, fn *FunctionSignature) {
	// Write description as comment
	if fn.Description != "" {
		for _, line := range strings.Split(fn.Description, "\n") {
			fmt.Fprintf(w, "// %s\n", line)
		}
	}

	// Write function signature
	if fn.Receiver != "" {
		fmt.Fprintf(w, "func (r *%s) %s(", fn.Receiver, fn.Name)
	} else {
		fmt.Fprintf(w, "func %s(", fn.Name)
	}

	// Parameters
	for i, p := range fn.Params {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "%s %s", p.Name, p.Type)
	}

	fmt.Fprintf(w, ")")

	// Return type
	if len(fn.Returns) > 0 && fn.Returns[0] != "" {
		if len(fn.Returns) == 1 {
			fmt.Fprintf(w, " %s", fn.Returns[0])
		} else {
			fmt.Fprintf(w, " (%s)", strings.Join(fn.Returns, ", "))
		}
	}

	fmt.Fprintf(w, " {\n")

	// Stub body
	fmt.Fprintf(w, "\t// TODO: Implement this function\n")
	if len(fn.Returns) > 0 && fn.Returns[0] != "" {
		fmt.Fprintf(w, "\tpanic(\"not implemented\")\n")
	}

	fmt.Fprintf(w, "}\n\n")
}

func (sg *ScaffoldGenerator) GenerateAIImplementations(model string) error {
	// This will be implemented to call GPT-4o-mini for each function stub
	fmt.Println("AI implementation generation will be added next...")
	return nil
}

func exportName(name string) string {
	if name == "" {
		return name
	}
	// Handle k-prefix constants
	if strings.HasPrefix(name, "k") && len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
		return name[1:]
	}
	// Handle m_ member prefix
	if strings.HasPrefix(name, "m_") {
		name = name[2:]
	}
	// Handle _ prefix
	if strings.HasPrefix(name, "_") {
		name = name[1:]
	}
	if len(name) == 0 {
		return "X"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
