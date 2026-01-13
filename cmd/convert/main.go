// Command convert translates OpenLoco C++ source files to Go.
// This is a mechanical translator inspired by Go 1.5's C-to-Go bootstrap.
// Output will be broken and require iteration to fix.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	srcDir     = flag.String("src", "", "OpenLoco source directory")
	outDir     = flag.String("out", "", "Output directory for Go files")
	pkg        = flag.String("pkg", "", "Specific package to convert (e.g., Math, Core)")
	verbose    = flag.Bool("v", false, "Verbose output")
	ignoreFile = flag.String("ignore", ".convertignore", "Path to conversion ignore file")
)

func main() {
	flag.Parse()

	if *srcDir == "" || *outDir == "" {
		fmt.Println("Usage: convert -src <openloco/src> -out <output/dir> [-pkg <package>] [-ignore <file>]")
		os.Exit(1)
	}

	converter := NewConverter(*srcDir, *outDir, *verbose, *ignoreFile)

	if *pkg != "" {
		if err := converter.ConvertPackage(*pkg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := converter.ConvertAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

type Converter struct {
	srcDir     string
	outDir     string
	verbose    bool
	ignoreFile string
	ignores    []string
}

func NewConverter(srcDir, outDir string, verbose bool, ignoreFile string) *Converter {
	c := &Converter{srcDir: srcDir, outDir: outDir, verbose: verbose, ignoreFile: ignoreFile}
	c.loadIgnores()
	return c
}

func (c *Converter) loadIgnores() {
	c.ignores = nil
	f, err := os.Open(c.ignoreFile)
	if err != nil {
		// ignore not present
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		c.ignores = append(c.ignores, l)
	}
}

func (c *Converter) isIgnored(path string) bool {
	// path is absolute or relative; normalize to slash-separated relative from srcDir
	rel, err := filepath.Rel(c.srcDir, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	for _, pat := range c.ignores {
		p := strings.TrimSpace(pat)
		if p == "" {
			continue
		}
		// if pattern ends with /** treat as dir prefix
		if strings.HasSuffix(p, "**") || strings.HasSuffix(p, "**/") {
			pp := strings.TrimSuffix(strings.TrimSuffix(p, "**"), "/")
			pp = strings.TrimPrefix(pp, "/")
			pp = filepath.ToSlash(pp)
			if strings.HasPrefix(rel, pp) {
				return true
			}
			continue
		}
		// exact match or dir prefix
		if rel == p || strings.HasPrefix(rel, strings.TrimSuffix(p, "/")+"/") {
			return true
		}
	}
	return false
}

func (c *Converter) ConvertAll() error {
	packages := []string{"Math", "Core", "Diagnostics", "Platform", "Utility", "Gfx"}
	for _, pkg := range packages {
		if err := c.ConvertPackage(pkg); err != nil {
			return err
		}
	}
	openLocoSubs := []string{
		"Audio", "Economy", "Effects", "Entities", "GameCommands",
		"Graphics", "Input", "Localisation", "Map", "Network",
		"Objects", "Paint", "S5", "Scenario", "Ui", "Vehicles", "World",
	}
	for _, sub := range openLocoSubs {
		if err := c.ConvertOpenLocoSub(sub); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
	return nil
}

func (c *Converter) ConvertPackage(pkg string) error {
	pkgDir := filepath.Join(c.srcDir, pkg)
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return fmt.Errorf("package directory not found: %s", pkgDir)
	}

	outPkg := filepath.Join(c.outDir, strings.ToLower(pkg))
	if err := os.MkdirAll(outPkg, 0755); err != nil {
		return err
	}

	return c.convertDir(pkgDir, outPkg, strings.ToLower(pkg))
}

var reservedPackageNames = map[string]string{
	"map":   "worldmap",
	"type":  "types",
	"func":  "funcs",
	"range": "ranges",
}

func sanitizePackageName(name string) string {
	lower := strings.ToLower(name)
	if replacement, ok := reservedPackageNames[lower]; ok {
		return replacement
	}
	return lower
}

func (c *Converter) ConvertOpenLocoSub(sub string) error {
	subDir := filepath.Join(c.srcDir, "OpenLoco", "src", sub)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		return fmt.Errorf("subdir not found: %s", subDir)
	}

	pkgName := sanitizePackageName(sub)
	outPkg := filepath.Join(c.outDir, pkgName)
	if err := os.MkdirAll(outPkg, 0755); err != nil {
		return err
	}

	return c.convertDir(subDir, outPkg, pkgName)
}

func (c *Converter) convertDir(srcDir, outDir, pkgName string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".hpp" && ext != ".cpp" && ext != ".h" {
			return nil
		}

		if strings.Contains(path, "tests") || strings.HasSuffix(info.Name(), "Tests.cpp") {
			return nil
		}

		// Check ignore patterns
		if c.isIgnored(path) {
			if c.verbose {
				fmt.Printf("Skipping (ignored): %s\n", path)
			}
			return nil
		}

		return c.convertFile(path, outDir, pkgName)
	})
}

func (c *Converter) convertFile(srcPath, outDir, pkgName string) error {
	if c.verbose {
		fmt.Printf("Converting: %s\n", srcPath)
	}

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Honor in-file skip markers: // CONVERSION:SKIP-START <id> ... // CONVERSION:SKIP-END <id>
	processed := honorSkipMarkers(string(content))

	goCode := c.translateToGo(processed, pkgName)

	baseName := filepath.Base(srcPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outPath := filepath.Join(outDir, toSnakeCase(baseName)+".go")

	return os.WriteFile(outPath, []byte(goCode), 0644)
}

func honorSkipMarkers(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inSkip := false
	skipID := ""
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, "// CONVERSION:SKIP-START") {
			inSkip = true
			parts := strings.Fields(trim)
			if len(parts) >= 2 {
				skipID = strings.Join(parts[1:], " ")
			}
			out = append(out, fmt.Sprintf("// CONVERSION SKIPPED START %s", skipID))
			continue
		}
		if strings.HasPrefix(trim, "// CONVERSION:SKIP-END") {
			inSkip = false
			out = append(out, fmt.Sprintf("// CONVERSION SKIPPED END %s", skipID))
			skipID = ""
			continue
		}
		if inSkip {
			// replace with comment to preserve line numbers roughly
			out = append(out, fmt.Sprintf("// <skipped> %s", trim))
			continue
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

func (c *Converter) translateToGo(cpp, pkgName string) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	out.WriteString("// AUTO-GENERATED FROM C++ - WILL NOT COMPILE\n")
	out.WriteString("// This is a mechanical translation that needs manual fixing\n\n")

	lines := strings.Split(cpp, "\n")
	t := &translator{
		lines:          lines,
		out:            &out,
		namespaceDepth: 0,
	}
	t.translate()

	return out.String()
}

// (rest of the file unchanged...)

type translator struct {
	lines          []string
	out            *strings.Builder
	namespaceDepth int
	inStruct       bool
	structBraces   int // track brace depth within struct
	inEnum         bool
	inFunction     bool
	funcBraceDepth int
	structName     string
	enumName       string
	enumFirst      bool
	pendingConsts  []string // constants found inside struct, emit after
}

func (t *translator) translate() {
	for i := 0; i < len(t.lines); i++ {
		line := t.lines[i]
		converted := t.translateLine(line)
		if converted != "" {
			t.out.WriteString(converted)
			t.out.WriteString("\n")
		}
	}
}

// Regex patterns
var (
	pragmaRe        = regexp.MustCompile(`^\s*#pragma`)
	includeRe       = regexp.MustCompile(`^\s*#include`)
	ifdefRe         = regexp.MustCompile(`^\s*#if`)
	endifRe         = regexp.MustCompile(`^\s*#endif`)
	defineRe        = regexp.MustCompile(`^\s*#define\s+(\w+)\s*(.*)`)
	namespaceRe     = regexp.MustCompile(`^\s*namespace\s+(\w+(?:::\w+)*)`)
	namespaceEndRe  = regexp.MustCompile(`^\s*}\s*//.*namespace`)
	structFwdRe     = regexp.MustCompile(`^\s*struct\s+(\w+)\s*;`) // forward declaration
	structRe        = regexp.MustCompile(`^\s*(?:template\s*<[^>]*>\s*)?struct\s+(\w+)(?:\s*:\s*(\w+))?`)
	classFwdRe      = regexp.MustCompile(`^\s*class\s+(\w+)\s*;`)
	classRe         = regexp.MustCompile(`^\s*(?:template\s*<[^>]*>\s*)?class\s+(\w+)`)
	enumRe          = regexp.MustCompile(`^\s*enum\s+(?:class\s+)?(\w+)`)
	enumValueRe     = regexp.MustCompile(`^\s*(\w+)\s*(?:=\s*([^,]+))?,?`)
	constexprRe     = regexp.MustCompile(`^\s*(?:static\s+)?constexpr\s+(\w+)\s+(\w+)\s*=\s*(.+);`)
	constexprArrRe  = regexp.MustCompile(`^\s*(?:static\s+)?constexpr\s+std::array<(\w+),\s*(\d+)>\s+(\w+)\s*=`)
	funcDefRe       = regexp.MustCompile(`^\s*(?:static\s+)?(?:inline\s+)?(?:constexpr\s+)?(\w+(?:<[^>]+>)?)\s+(\w+)\s*\(([^)]*)\)\s*(?:const)?\s*(?:noexcept)?\s*\{?`)
	memberVarRe     = regexp.MustCompile(`^\s*(\w+(?:<[^>]+>)?)\s+(\w+)\s*(?:\{[^}]*\})?\s*;`)
	operatorRe      = regexp.MustCompile(`operator([+\-*/=<>!&|]+|\[\]|\(\))`)
	usingRe         = regexp.MustCompile(`^\s*using\s+(\w+)\s*=\s*(.+);`)
	templateRe      = regexp.MustCompile(`^\s*template\s*<`)
	returnRe        = regexp.MustCompile(`^\s*return\s+(.+);`)
	forRe           = regexp.MustCompile(`^\s*for\s*\(([^;]*);([^;]*);([^)]*)\)`)
	whileRe         = regexp.MustCompile(`^\s*while\s*\((.+)\)`)
	ifRe            = regexp.MustCompile(`^\s*if\s*\((.+)\)`)
	elseIfRe        = regexp.MustCompile(`^\s*else\s+if\s*\((.+)\)`)
	elseRe          = regexp.MustCompile(`^\s*else\s*\{?`)
	switchRe        = regexp.MustCompile(`^\s*switch\s*\((.+)\)`)
	caseRe          = regexp.MustCompile(`^\s*case\s+(.+):`)
	defaultRe       = regexp.MustCompile(`^\s*default:`)
	breakRe         = regexp.MustCompile(`^\s*break;`)
	continueRe      = regexp.MustCompile(`^\s*continue;`)
	varDeclRe       = regexp.MustCompile(`^\s*(const\s+)?(\w+(?:<[^>]+>)?)\s+(\w+)\s*=\s*(.+);`)
	varDeclNoInitRe = regexp.MustCompile(`^\s*(const\s+)?(\w+(?:<[^>]+>)?)\s+(\w+)\s*;`)
	assignRe        = regexp.MustCompile(`^\s*(\w+)\s*([+\-*/]?=)\s*(.+);`)
	funcCallRe      = regexp.MustCompile(`^\s*(\w+(?:::\w+)*)\s*\(([^)]*)\)\s*;`)
)

func (t *translator) translateLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines
	if trimmed == "" {
		return ""
	}

	// Handle preprocessor
	if pragmaRe.MatchString(line) {
		return ""
	}
	if includeRe.MatchString(line) {
		return "// " + trimmed
	}
	if ifdefRe.MatchString(line) || endifRe.MatchString(line) {
		return ""
	}
	if m := defineRe.FindStringSubmatch(line); m != nil {
		if m[2] != "" {
			return fmt.Sprintf("const %s = %s", m[1], translateExpr(m[2]))
		}
		return ""
	}

	// Namespace handling - track depth but don't emit braces
	if m := namespaceRe.FindStringSubmatch(line); m != nil {
		t.namespaceDepth++
		if strings.Contains(line, "{") {
			// opening brace on same line
		}
		return fmt.Sprintf("// namespace %s", m[1])
	}
	if namespaceEndRe.MatchString(line) {
		t.namespaceDepth--
		return ""
	}

	// Template lines - skip them, the next line will have the actual definition
	if templateRe.MatchString(line) && !strings.Contains(line, "struct") && !strings.Contains(line, "class") {
		return "// " + trimmed
	}

	// using Type = ...
	if m := usingRe.FindStringSubmatch(line); m != nil {
		goType := translateType(m[2])
		return fmt.Sprintf("type %s = %s", exportName(m[1]), goType)
	}

	// constexpr array
	if m := constexprArrRe.FindStringSubmatch(line); m != nil {
		elemType := translateType(m[1])
		size := m[2]
		name := exportName(m[3])
		return fmt.Sprintf("var %s = [%s]%s{", name, size, elemType)
	}

	// constexpr var (only at top level, not inside struct)
	if !t.inStruct {
		if m := constexprRe.FindStringSubmatch(line); m != nil {
			goType := translateType(m[1])
			name := exportName(m[2])
			value := translateExpr(m[3])
			// Use var for auto/any types since const requires concrete type
			if goType == "any" || strings.Contains(goType, "/*") {
				return fmt.Sprintf("var %s = %s // %s", name, value, m[1])
			}
			return fmt.Sprintf("const %s %s = %s", name, goType, value)
		}
	}

	// Forward declarations - just comment them
	if structFwdRe.MatchString(line) || classFwdRe.MatchString(line) {
		return "// forward: " + trimmed
	}

	// struct/class
	if m := structRe.FindStringSubmatch(line); m != nil {
		baseType := ""
		if len(m) > 2 && m[2] != "" {
			baseType = m[2]
		}
		if t.inStruct {
			// Nested struct - emit as separate type
			t.structBraces++ // track the nested brace
			nestedName := exportName(t.structName) + exportName(m[1])
			if baseType != "" {
				return fmt.Sprintf("}\n\ntype %s struct {\n\t%s // embedded", nestedName, exportName(baseType))
			}
			return fmt.Sprintf("}\n\ntype %s struct {", nestedName)
		}
		t.inStruct = true
		t.structBraces = 1
		t.structName = m[1]
		t.pendingConsts = nil
		if baseType != "" {
			return fmt.Sprintf("type %s struct {\n\t%s // embedded", exportName(m[1]), exportName(baseType))
		}
		return fmt.Sprintf("type %s struct {", exportName(m[1]))
	}
	if m := classRe.FindStringSubmatch(line); m != nil {
		if t.inStruct {
			t.structBraces++
			nestedName := exportName(t.structName) + exportName(m[1])
			return fmt.Sprintf("}\n\ntype %s struct {", nestedName)
		}
		t.inStruct = true
		t.structBraces = 1
		t.structName = m[1]
		t.pendingConsts = nil
		return fmt.Sprintf("type %s struct {", exportName(m[1]))
	}

	// enum
	if m := enumRe.FindStringSubmatch(line); m != nil {
		t.inEnum = true
		t.enumFirst = true
		t.enumName = exportName(m[1])
		return fmt.Sprintf("type %s int\n\nconst (", t.enumName)
	}

	// Inside enum - translate values
	if t.inEnum {
		if trimmed == "};" || trimmed == "}" {
			t.inEnum = false
			t.enumName = ""
			return ")"
		}
		if m := enumValueRe.FindStringSubmatch(trimmed); m != nil {
			name := exportName(m[1])
			if m[2] != "" {
				t.enumFirst = false
				return fmt.Sprintf("\t%s %s = %s", name, t.enumName, translateExpr(m[2]))
			}
			if t.enumFirst {
				t.enumFirst = false
				return fmt.Sprintf("\t%s %s = iota", name, t.enumName)
			}
			return fmt.Sprintf("\t%s", name)
		}
	}

	// End of struct
	if t.inStruct && !t.inFunction {
		if trimmed == "};" || (trimmed == "}" && !strings.Contains(line, "return")) {
			t.structBraces--
			if t.structBraces <= 0 {
				t.inStruct = false
				result := "}"
				// Emit any pending constants after the struct
				for _, c := range t.pendingConsts {
					result += "\n" + c
				}
				t.pendingConsts = nil
				t.structName = ""
				return result
			}
			// Nested struct closed, but parent continues - just close this one
			return "}"
		}
	}

	// Inside struct - member variables and methods
	if t.inStruct && !t.inFunction {
		// Skip access specifiers
		if trimmed == "public:" || trimmed == "private:" || trimmed == "protected:" {
			return ""
		}

		// static constexpr inside struct -> save for later
		if m := constexprRe.FindStringSubmatch(line); m != nil {
			goType := translateType(m[1])
			name := exportName(t.structName) + exportName(m[2])
			value := translateExpr(m[3])
			t.pendingConsts = append(t.pendingConsts, fmt.Sprintf("const %s %s = %s", name, goType, value))
			return ""
		}

		// Operator overloads - comment them
		if operatorRe.MatchString(line) {
			return "\t// " + trimmed
		}

		// Member function with body
		if m := funcDefRe.FindStringSubmatch(line); m != nil {
			if strings.Contains(line, "{") {
				t.inFunction = true
				t.funcBraceDepth = 1
				retType := translateType(m[1])
				name := exportName(m[2])
				params := translateParams(m[3])
				recv := strings.ToLower(t.structName[:1])
				return fmt.Sprintf("func (%s *%s) %s(%s) %s {", recv, exportName(t.structName), name, params, retType)
			}
			// Declaration only
			return "\t// method: " + trimmed
		}

		// Member variable
		if m := memberVarRe.FindStringSubmatch(trimmed); m != nil {
			goType := translateType(m[1])
			name := exportName(m[2])
			return fmt.Sprintf("\t%s %s", name, goType)
		}
	}

	// Standalone function
	if !t.inStruct && !t.inEnum && !t.inFunction {
		if m := funcDefRe.FindStringSubmatch(line); m != nil {
			retType := translateType(m[1])
			name := exportName(m[2])
			params := translateParams(m[3])
			if strings.Contains(line, "{") {
				t.inFunction = true
				t.funcBraceDepth = 1
				return fmt.Sprintf("func %s(%s) %s {", name, params, retType)
			}
			// Forward declaration
			return fmt.Sprintf("// func %s(%s) %s", name, params, retType)
		}
	}

	// Inside function body
	if t.inFunction {
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")
		t.funcBraceDepth += opens - closes

		if t.funcBraceDepth <= 0 {
			t.inFunction = false
			return "}"
		}

		// Translate statements
		indent := "\t"

		// return
		if m := returnRe.FindStringSubmatch(line); m != nil {
			return indent + "return " + translateExpr(m[1])
		}

		// for loop
		if m := forRe.FindStringSubmatch(line); m != nil {
			init := translateStmt(m[1])
			cond := translateExpr(m[2])
			post := translateStmt(m[3])
			result := fmt.Sprintf("%sfor %s; %s; %s {", indent, init, cond, post)
			if !strings.Contains(line, "{") {
				t.funcBraceDepth++ // implicit block
			}
			return result
		}

		// while loop
		if m := whileRe.FindStringSubmatch(line); m != nil {
			cond := translateExpr(m[1])
			result := fmt.Sprintf("%sfor %s {", indent, cond)
			if !strings.Contains(line, "{") {
				t.funcBraceDepth++
			}
			return result
		}

		// if/else if/else
		if m := elseIfRe.FindStringSubmatch(line); m != nil {
			cond := translateExpr(m[1])
			return fmt.Sprintf("%s} else if %s {", indent, cond)
		}
		if elseRe.MatchString(line) {
			return indent + "} else {"
		}
		if m := ifRe.FindStringSubmatch(line); m != nil {
			cond := translateExpr(m[1])
			result := fmt.Sprintf("%sif %s {", indent, cond)
			if !strings.Contains(line, "{") {
				t.funcBraceDepth++
			}
			return result
		}

		// switch
		if m := switchRe.FindStringSubmatch(line); m != nil {
			expr := translateExpr(m[1])
			return fmt.Sprintf("%sswitch %s {", indent, expr)
		}
		if m := caseRe.FindStringSubmatch(line); m != nil {
			return fmt.Sprintf("%scase %s:", indent, translateExpr(m[1]))
		}
		if defaultRe.MatchString(line) {
			return indent + "default:"
		}

		// break/continue
		if breakRe.MatchString(line) {
			return indent + "// break"
		}
		if continueRe.MatchString(line) {
			return indent + "continue"
		}

		// Variable declaration with init
		if m := varDeclRe.FindStringSubmatch(line); m != nil {
			goType := translateType(m[2])
			name := m[3]
			value := translateExpr(m[4])
			if m[1] != "" { // const
				return fmt.Sprintf("%s%s := %s // %s", indent, name, value, goType)
			}
			return fmt.Sprintf("%svar %s %s = %s", indent, name, goType, value)
		}

		// Variable declaration without init
		if m := varDeclNoInitRe.FindStringSubmatch(line); m != nil {
			goType := translateType(m[2])
			name := m[3]
			return fmt.Sprintf("%svar %s %s", indent, name, goType)
		}

		// Assignment
		if m := assignRe.FindStringSubmatch(line); m != nil {
			return fmt.Sprintf("%s%s %s %s", indent, m[1], m[2], translateExpr(m[3]))
		}

		// Function call
		if m := funcCallRe.FindStringSubmatch(line); m != nil {
			funcName := translateExpr(m[1])
			args := translateExpr(m[2])
			return fmt.Sprintf("%s%s(%s)", indent, funcName, args)
		}

		// Closing brace
		if trimmed == "}" {
			return indent + "}"
		}

		// Unknown statement - comment it
		if trimmed != "" && trimmed != "{" {
			return indent + "// " + trimmed
		}
		return ""
	}

	// Stray closing brace from namespace or struct
	if trimmed == "}" {
		if t.namespaceDepth > 0 {
			t.namespaceDepth--
		}
		return ""
	}

	// Catch stray member-like declarations at top level (from nested struct aftermath)
	if !t.inStruct && !t.inFunction && !t.inEnum {
		if m := memberVarRe.FindStringSubmatch(trimmed); m != nil {
			return "// orphan member: " + trimmed
		}
	}

	// Pass through anything else as comment
	if trimmed != "" && trimmed != "{" && trimmed != "};" {
		return "// " + trimmed
	}
	return ""
}

func translateType(cppType string) string {
	cppType = strings.TrimSpace(cppType)

	// Handle array types
	arrRe := regexp.MustCompile(`std::array<(\\w+),\\s*(\\d+)>`)
	if m := arrRe.FindStringSubmatch(cppType); m != nil {
		elemType := translateType(m[1])
		return fmt.Sprintf("[%s]%s", m[2], elemType)
	}

	// Remove const, &, *, noexcept
	cppType = strings.ReplaceAll(cppType, "const ", "")
	cppType = strings.ReplaceAll(cppType, "noexcept", "")
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
		"char":     "byte",
		"auto":     "any",
		"string":   "string",
	}

	if goType, ok := typeMap[cppType]; ok {
		return goType
	}

	// std::string
	if strings.Contains(cppType, "std::string") {
		return "string"
	}

	// std::vector<T> -> []T
	vecRe := regexp.MustCompile(`std::vector<(.+)>`)
	if m := vecRe.FindStringSubmatch(cppType); m != nil {
		return "[]" + translateType(m[1])
	}

	// std::optional<T> -> *T
	optRe := regexp.MustCompile(`std::optional<(.+)>`)
	if m := optRe.FindStringSubmatch(cppType); m != nil {
		return "*" + translateType(m[1])
	}

	// Template types -> placeholder (can't use any for const)
	if strings.Contains(cppType, "<") {
		return "any /* " + cppType + " */ "
	}

	return cppType
}

func translateParams(params string) string {
	if strings.TrimSpace(params) == "" {
		return ""
	}

	parts := strings.Split(params, ",")
	var goParams []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		p = strings.ReplaceAll(p, "const ", "")
		p = strings.ReplaceAll(p, "&", "")
		p = strings.ReplaceAll(p, "*", "")

		tokens := strings.Fields(p)
		if len(tokens) >= 2 {
			name := tokens[len(tokens)-1]
			typeParts := tokens[:len(tokens)-1]
			goType := translateType(strings.Join(typeParts, " "))
			goParams = append(goParams, fmt.Sprintf("%s %s", name, goType))
		} else if len(tokens) == 1 {
			goParams = append(goParams, translateType(tokens[0]))
		}
	}

	return strings.Join(goParams, ", ")
}

func translateStmt(stmt string) string {
	stmt = strings.TrimSpace(stmt)
	stmt = strings.TrimSuffix(stmt, ";")

	// Type declarations in for init: "int i = 0" -> "i := 0"
	varRe := regexp.MustCompile(`^(\\w+)\\s+(\\w+)\\s*=\\s*(.+)$`)
	if m := varRe.FindStringSubmatch(stmt); m != nil {
		return fmt.Sprintf("%s := %s", m[2], translateExpr(m[3]))
	}

	return translateExpr(stmt)
}

func translateExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	expr = strings.TrimSuffix(expr, ";")

	// nullptr -> nil
	expr = strings.ReplaceAll(expr, "nullptr", "nil")
	expr = strings.ReplaceAll(expr, "NULL", "nil")

	// this-> -> remove (Go uses implicit receiver)
	expr = strings.ReplaceAll(expr, "this->", "")

	// static_cast<T>(x) -> T(x)
	castRe := regexp.MustCompile(`static_cast<(\\w+)>\\(([^)]+)\\)`)
	expr = castRe.ReplaceAllString(expr, "$1($2)")

	// reinterpret_cast -> unsafe conversion comment
	expr = regexp.MustCompile(`reinterpret_cast<[^>]+>\\([^)]+\\)`).ReplaceAllString(expr, "/* unsafe cast */")

	// std:: functions
	expr = strings.ReplaceAll(expr, "std::abs", "abs")
	expr = strings.ReplaceAll(expr, "std::max", "max")
	expr = strings.ReplaceAll(expr, "std::min", "min")
	expr = strings.ReplaceAll(expr, "std::move", "")

	// :: -> . for namespaced calls
	expr = strings.ReplaceAll(expr, "::", ".")

	// Remove trailing f from float literals
	floatRe := regexp.MustCompile(`(\\d+\\.\\d+)f\\b`)
	expr = floatRe.ReplaceAllString(expr, "$1")

	// Remove U/u suffix from integer literals (unsigned)
	uintRe := regexp.MustCompile(`(\\d+)[Uu]\\b`)
	expr = uintRe.ReplaceAllString(expr, "$1")

	// Remove L/l suffix from integer literals (long)
	longRe := regexp.MustCompile(`(\\d+)[Ll]\\b`)
	expr = longRe.ReplaceAllString(expr, "$1")

	// Remove UL/ul/LU/lu suffix
	ulRe := regexp.MustCompile(`(\\d+)(?:[Uu][Ll]|[Ll][Uu])\\b`)
	expr = ulRe.ReplaceAllString(expr, "$1")

	// Remove ULL/ull suffix
	ullRe := regexp.MustCompile(`(\\d+)(?:[Uu][Ll][Ll]|[Ll][Ll][Uu])\\b`)
	expr = ullRe.ReplaceAllString(expr, "$1")

	return expr
}

func exportName(name string) string {
	if name == "" {
		return name
	}
	// k-prefix constants
	if strings.HasPrefix(name, "k") && len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
		return name[1:]
	}
	// m_ member prefix
	if strings.HasPrefix(name, "m_") {
		name = name[2:]
	}
	// _ prefix (private)
	if strings.HasPrefix(name, "_") {
		name = name[1:]
	}

	if len(name) == 0 {
		return "X"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
