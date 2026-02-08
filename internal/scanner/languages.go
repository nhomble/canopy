package scanner

import (
	"regexp"
	"strings"
)

// SignaturePattern defines a regex for extracting a specific kind of code signature.
type SignaturePattern struct {
	Kind    string // "function", "class", "method", "interface", "struct", "enum"
	Pattern *regexp.Regexp
}

// LanguageDef describes how to extract imports and signatures from a language.
type LanguageDef struct {
	Name       string
	Extensions []string
	Imports    []*regexp.Regexp
	Signatures []SignaturePattern
	Decorators *regexp.Regexp // optional, nil if not applicable
}

var javaLang = &LanguageDef{
	Name:       "java",
	Extensions: []string{".java"},
	Imports: []*regexp.Regexp{
		regexp.MustCompile(`import\s+([\w.]+(?:\.\*)?);`),
	},
	Signatures: []SignaturePattern{
		{Kind: "class", Pattern: regexp.MustCompile(`(?:public|private|protected)?\s*(?:abstract\s+)?class\s+(\w+)`)},
		{Kind: "interface", Pattern: regexp.MustCompile(`(?:public\s+)?interface\s+(\w+)`)},
		{Kind: "method", Pattern: regexp.MustCompile(`(?:public|private|protected)\s+[\w<>\[\]]+\s+(\w+)\s*\(`)},
		{Kind: "enum", Pattern: regexp.MustCompile(`(?:public\s+)?enum\s+(\w+)`)},
	},
	Decorators: regexp.MustCompile(`@(\w+)`),
}

var typescriptLang = &LanguageDef{
	Name:       "typescript",
	Extensions: []string{".ts", ".tsx"},
	Imports: []*regexp.Regexp{
		regexp.MustCompile(`import\s+.*\s+from\s+['"](.+)['"]`),
		regexp.MustCompile(`import\s+['"](.+)['"]`),
	},
	Signatures: []SignaturePattern{
		{Kind: "function", Pattern: regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)`)},
		{Kind: "class", Pattern: regexp.MustCompile(`(?:export\s+)?class\s+(\w+)`)},
		{Kind: "interface", Pattern: regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`)},
		{Kind: "function", Pattern: regexp.MustCompile(`export\s+const\s+(\w+)\s*=`)},
	},
	Decorators: regexp.MustCompile(`@(\w+)`),
}

var javascriptLang = &LanguageDef{
	Name:       "javascript",
	Extensions: []string{".js", ".jsx", ".mjs", ".cjs"},
	Imports: []*regexp.Regexp{
		regexp.MustCompile(`import\s+.*\s+from\s+['"](.+)['"]`),
		regexp.MustCompile(`import\s+['"](.+)['"]`),
		regexp.MustCompile(`require\s*\(\s*['"](.+)['"]\s*\)`),
	},
	Signatures: []SignaturePattern{
		{Kind: "function", Pattern: regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)`)},
		{Kind: "class", Pattern: regexp.MustCompile(`(?:export\s+)?class\s+(\w+)`)},
		{Kind: "function", Pattern: regexp.MustCompile(`export\s+const\s+(\w+)\s*=`)},
	},
	Decorators: nil,
}

var goLang = &LanguageDef{
	Name:       "go",
	Extensions: []string{".go"},
	Imports: []*regexp.Regexp{
		regexp.MustCompile(`"([\w./\-]+)"`),
	},
	Signatures: []SignaturePattern{
		{Kind: "function", Pattern: regexp.MustCompile(`^func\s+(\w+)\s*\(`)},
		{Kind: "method", Pattern: regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+(\w+)\s*\(`)},
		{Kind: "interface", Pattern: regexp.MustCompile(`^type\s+(\w+)\s+interface\s*\{`)},
		{Kind: "struct", Pattern: regexp.MustCompile(`^type\s+(\w+)\s+struct\s*\{`)},
	},
	Decorators: nil,
}

var pythonLang = &LanguageDef{
	Name:       "python",
	Extensions: []string{".py"},
	Imports: []*regexp.Regexp{
		regexp.MustCompile(`^from\s+([\w.]+)\s+import`),
		regexp.MustCompile(`^import\s+([\w.]+)`),
	},
	Signatures: []SignaturePattern{
		{Kind: "function", Pattern: regexp.MustCompile(`^def\s+(\w+)\s*\(`)},
		{Kind: "class", Pattern: regexp.MustCompile(`^class\s+(\w+)`)},
	},
	Decorators: regexp.MustCompile(`^@(\w+)`),
}

// allLanguages is the registry of supported languages.
var allLanguages = []*LanguageDef{
	javaLang,
	typescriptLang,
	javascriptLang,
	goLang,
	pythonLang,
}

// extensionMap is built at init time for fast lookups.
var extensionMap map[string]*LanguageDef

func init() {
	extensionMap = make(map[string]*LanguageDef)
	for _, lang := range allLanguages {
		for _, ext := range lang.Extensions {
			extensionMap[strings.ToLower(ext)] = lang
		}
	}
}

// GetLanguage returns the language definition for a given file extension.
// Returns nil for unknown extensions.
func GetLanguage(ext string) *LanguageDef {
	return extensionMap[strings.ToLower(ext)]
}
