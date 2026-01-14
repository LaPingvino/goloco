package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIImplementationGenerator uses OpenAI API to generate function implementations
type AIImplementationGenerator struct {
	apiKey  string
	model   string
	client  *http.Client
	verbose bool
}

func NewAIImplementationGenerator(model string, verbose bool) (*AIImplementationGenerator, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	return &AIImplementationGenerator{
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
		verbose: verbose,
	}, nil
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Temperature float64 `json:"temperature"`
	MaxTokens int `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (aig *AIImplementationGenerator) GenerateImplementation(fn *FunctionSignature, contextTypes []TypeDefinition) (string, error) {
	prompt := aig.buildPrompt(fn, contextTypes)

	if aig.verbose {
		fmt.Printf("\n🤖 Generating implementation for %s...\n", fn.Name)
	}

	req := ChatCompletionRequest{
		Model: aig.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are an expert Go programmer helping to port C++ code to Go. Generate clean, idiomatic Go implementations based on C++ logic. Only output the function body code, no explanations.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more deterministic code
		MaxTokens:   1000,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+aig.apiKey)

	resp, err := aig.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w\nBody: %s", err, string(respBody))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	implementation := chatResp.Choices[0].Message.Content

	// Clean up the response - remove markdown code blocks if present
	implementation = strings.TrimPrefix(implementation, "```go\n")
	implementation = strings.TrimPrefix(implementation, "```\n")
	implementation = strings.TrimSuffix(implementation, "```")
	implementation = strings.TrimSpace(implementation)

	return implementation, nil
}

func (aig *AIImplementationGenerator) buildPrompt(fn *FunctionSignature, contextTypes []TypeDefinition) string {
	var prompt strings.Builder

	prompt.WriteString("Convert this C++ function to Go:\n\n")

	// Include function signature
	prompt.WriteString("## Go Function Signature:\n```go\n")
	if fn.Receiver != "" {
		fmt.Fprintf(&prompt, "func (r *%s) %s(", fn.Receiver, fn.Name)
	} else {
		fmt.Fprintf(&prompt, "func %s(", fn.Name)
	}

	for i, p := range fn.Params {
		if i > 0 {
			prompt.WriteString(", ")
		}
		fmt.Fprintf(&prompt, "%s %s", p.Name, p.Type)
	}

	prompt.WriteString(")")

	if len(fn.Returns) > 0 && fn.Returns[0] != "" {
		if len(fn.Returns) == 1 {
			fmt.Fprintf(&prompt, " %s", fn.Returns[0])
		} else {
			fmt.Fprintf(&prompt, " (%s)", strings.Join(fn.Returns, ", "))
		}
	}

	prompt.WriteString(" {\n\t// TODO\n}\n```\n\n")

	// Include C++ code for reference
	if fn.CppCode != "" {
		prompt.WriteString("## Original C++ Code:\n```cpp\n")
		prompt.WriteString(fn.CppCode)
		prompt.WriteString("\n```\n\n")
	}

	// Include relevant type definitions for context
	if len(contextTypes) > 0 {
		prompt.WriteString("## Available Types:\n```go\n")
		for _, td := range contextTypes {
			if td.Kind == "struct" && len(td.Fields) > 0 {
				fmt.Fprintf(&prompt, "type %s struct {\n", td.Name)
				for _, f := range td.Fields {
					fmt.Fprintf(&prompt, "    %s %s\n", f.Name, f.Type)
				}
				prompt.WriteString("}\n\n")
			}
		}
		prompt.WriteString("```\n\n")
	}

	prompt.WriteString("## Instructions:\n")
	prompt.WriteString("1. Translate the C++ logic to idiomatic Go\n")
	prompt.WriteString("2. Use Go naming conventions (camelCase for private, CamelCase for public)\n")
	prompt.WriteString("3. Handle errors appropriately (return error if needed)\n")
	prompt.WriteString("4. Only output the function body (the code between { and }), no signature\n")
	prompt.WriteString("5. If the C++ code is unclear or missing, write a sensible stub that compiles\n\n")

	prompt.WriteString("Output ONLY the function body code:")

	return prompt.String()
}

// BatchGenerateImplementations generates implementations for multiple functions
func (aig *AIImplementationGenerator) BatchGenerateImplementations(
	functions []FunctionSignature,
	contextTypes []TypeDefinition,
	concurrency int,
) (map[string]string, error) {

	results := make(map[string]string)
	errors := make([]error, 0)

	// Simple sequential processing for now (can be parallelized later)
	for i, fn := range functions {
		if aig.verbose {
			fmt.Printf("[%d/%d] Processing %s...\n", i+1, len(functions), fn.Name)
		}

		impl, err := aig.GenerateImplementation(&fn, contextTypes)
		if err != nil {
			if aig.verbose {
				fmt.Printf("  ⚠️  Failed: %v\n", err)
			}
			errors = append(errors, fmt.Errorf("%s: %w", fn.Name, err))
			continue
		}

		results[fn.Name] = impl

		if aig.verbose {
			fmt.Printf("  ✓ Generated %d bytes\n", len(impl))
		}

		// Rate limiting - sleep briefly between requests
		if i < len(functions)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all generations failed: %v", errors[0])
	}

	return results, nil
}

// WriteImplementationsToFile updates a scaffold file with AI-generated implementations
func (aig *AIImplementationGenerator) WriteImplementationsToFile(
	scaffoldPath string,
	implementations map[string]string,
) error {

	content, err := os.ReadFile(scaffoldPath)
	if err != nil {
		return err
	}

	fileContent := string(content)

	// Replace each function stub with implementation
	for funcName, impl := range implementations {
		// Find the function in the file
		funcPattern := fmt.Sprintf(`func (?:\([^)]+\) )?%s\([^)]*\)[^{]*\{[^}]*panic\("not implemented"\)[^}]*\}`, funcName)
		// This is simplified - a proper implementation would use AST parsing

		// For now, just append as comment
		fileContent += fmt.Sprintf("\n// AI-generated implementation for %s:\n/*\n%s\n*/\n", funcName, impl)
	}

	return os.WriteFile(scaffoldPath, []byte(fileContent), 0644)
}

// ExtractFunctionBody extracts just the function body from generated code
func ExtractFunctionBody(code string) string {
	lines := strings.Split(code, "\n")
	var bodyLines []string
	inBody := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip function signature line
		if strings.HasPrefix(trimmed, "func ") {
			inBody = true
			continue
		}

		// Skip opening brace if on its own line
		if trimmed == "{" {
			continue
		}

		// Skip closing brace if on its own line
		if trimmed == "}" {
			break
		}

		if inBody {
			bodyLines = append(bodyLines, line)
		}
	}

	return strings.Join(bodyLines, "\n")
}
