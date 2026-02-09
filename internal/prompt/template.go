package prompt

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/nhomble/canopy/internal/patterns"
)

//go:embed templates/*.md.tmpl
var templateFS embed.FS

//go:embed schemas/*.json
var schemaFS embed.FS

// ScanStats holds high-level statistics from a codebase scan.
type ScanStats struct {
	TotalFiles       int
	TotalDirs        int
	FilesByExtension map[string]int
}

// PromptData is the complete data bag passed to the analysis prompt template.
type PromptData struct {
	RepoID        string
	Tree          string
	Stats         ScanStats
	Patterns      []patterns.PatternDef
	OutputSchema  string
	ExampleOutput string
}

// exampleOutput is a small, valid JSON example illustrating the expected format.
var exampleOutput = `{
  "repo_id": "my-app",
  "patterns": ["Hexagonal Architecture"],
  "components": [
    {
      "id": "user-domain",
      "name": "User Domain",
      "layer": "core",
      "code_refs": ["src/domain/user/"],
      "provides": {
        "interface": "UserService",
        "symbols": ["User", "UserService", "CreateUserUseCase"]
      },
      "analyzed": true
    },
    {
      "id": "user-api",
      "name": "User API",
      "layer": "adapters",
      "code_refs": ["src/adapters/http/userController.ts"],
      "analyzed": true
    },
    {
      "id": "user-persistence",
      "name": "User Persistence",
      "layer": "adapters",
      "code_refs": ["src/adapters/db/userRepository.ts"],
      "analyzed": true
    },
    {
      "id": "admin-dashboard",
      "name": "Admin Dashboard",
      "layer": "app",
      "code_refs": ["dashboard/**"],
      "analyzed": true
    }
  ],
  "archetypes": {
    "controller": [
      {
        "id": "user-controller",
        "file": "src/adapters/http/userController.ts",
        "symbol": "UserController",
        "technology": "express",
        "entry_point_type": "http",
        "routes": ["/api/users", "/api/users/:id"]
      }
    ],
    "repository": [
      {
        "id": "user-repo",
        "file": "src/adapters/db/userRepository.ts",
        "symbol": "UserRepository",
        "technology": "typeorm",
        "entity": "User"
      }
    ],
    "service": [
      {
        "id": "user-service",
        "file": "src/domain/user/userService.ts",
        "symbol": "UserService",
        "purpose": "User CRUD and business rules"
      }
    ]
  },
  "relationships": [
    {"from": "user-controller", "to": "user-service", "type": "calls", "flow": "create-user"},
    {"from": "user-service", "to": "user-repo", "type": "calls", "flow": "create-user"},
    {"from": "admin-dashboard", "to": "user-api", "type": "calls"}
  ],
  "flows": [
    {
      "id": "create-user",
      "name": "Create User",
      "steps": ["user-controller", "user-service", "user-repo"],
      "pattern": "Hexagonal Architecture"
    }
  ]
}`

// RenderAnalysisPrompt renders the root analysis prompt with the given data.
func RenderAnalysisPrompt(data PromptData) (string, error) {
	if data.OutputSchema == "" {
		s, err := LoadSchema()
		if err != nil {
			return "", fmt.Errorf("loading output schema: %w", err)
		}
		data.OutputSchema = s
	}

	if data.ExampleOutput == "" {
		data.ExampleOutput = exampleOutput
	}

	tmplContent, err := templateFS.ReadFile("templates/analyze-root.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("analyze-root").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// LoadSchema reads the index-schema.json from the embedded filesystem.
func LoadSchema() (string, error) {
	data, err := schemaFS.ReadFile("schemas/index-schema.json")
	if err != nil {
		return "", fmt.Errorf("reading schema: %w", err)
	}
	return string(data), nil
}
