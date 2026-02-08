package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadIndex reads and parses an ArchIndex from a JSON file.
func LoadIndex(path string) (*ArchIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}
	var idx ArchIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing index: %w", err)
	}
	return &idx, nil
}

// SaveIndex writes an ArchIndex to a JSON file with indentation.
func SaveIndex(path string, idx *ArchIndex) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing index: %w", err)
	}
	return nil
}

// ExtractJSON attempts to extract a valid JSON object from potentially messy
// LLM output. It handles:
// - JSON wrapped in markdown code fences
// - JSON preceded/followed by commentary text
// - Clean JSON
func ExtractJSON(input string) ([]byte, error) {
	// First, try the input as-is
	trimmed := strings.TrimSpace(input)
	if json.Valid([]byte(trimmed)) {
		return []byte(trimmed), nil
	}

	// Try stripping markdown code fences
	stripped := stripMarkdownFences(trimmed)
	if stripped != trimmed && json.Valid([]byte(stripped)) {
		return []byte(stripped), nil
	}

	// Find the first { and last } — extract the JSON object
	firstBrace := strings.Index(trimmed, "{")
	lastBrace := strings.LastIndex(trimmed, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		candidate := trimmed[firstBrace : lastBrace+1]
		if json.Valid([]byte(candidate)) {
			return []byte(candidate), nil
		}

		// Try fixing trailing commas (common LLM mistake)
		fixed := fixTrailingCommas(candidate)
		if json.Valid([]byte(fixed)) {
			return []byte(fixed), nil
		}
	}

	return nil, fmt.Errorf("could not extract valid JSON from input")
}

func stripMarkdownFences(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence || (!strings.HasPrefix(trimmed, "```") && len(result) > 0) || strings.HasPrefix(trimmed, "{") {
			result = append(result, line)
		}
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func fixTrailingCommas(s string) string {
	// Remove trailing commas before } or ]
	var result strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == ',' {
			// Look ahead past whitespace for } or ]
			j := i + 1
			for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n' || runes[j] == '\r') {
				j++
			}
			if j < len(runes) && (runes[j] == '}' || runes[j] == ']') {
				continue // Skip the trailing comma
			}
		}
		result.WriteRune(runes[i])
	}
	return result.String()
}
