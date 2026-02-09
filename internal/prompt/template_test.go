package prompt

import (
	"strings"
	"testing"

	"github.com/nhomble/arch-index/internal/patterns"
)

func TestRenderAnalysisPrompt(t *testing.T) {
	// Load real pattern definitions.
	pats, err := patterns.LoadAll()
	if err != nil {
		t.Fatalf("patterns.LoadAll() failed: %v", err)
	}

	data := PromptData{
		RepoID: "test-repo",
		Tree: `test-repo/
├── src/
│   ├── controllers/
│   │   └── userController.ts
│   ├── services/
│   │   └── userService.ts
│   └── models/
│       └── user.ts
├── package.json
└── tsconfig.json`,
		Stats: ScanStats{
			TotalFiles: 5,
			TotalDirs:  4,
			FilesByExtension: map[string]int{
				".ts":   3,
				".json": 2,
			},
		},
		TechStack: []TechHint{
			{Source: "package.json", Name: "express", Type: "framework"},
			{Source: "tsconfig.json", Name: "typescript", Type: "language"},
		},
		Signatures: []FileSignatures{
			{
				Path:     "/workspace/test-repo/src/controllers/userController.ts",
				RelPath:  "src/controllers/userController.ts",
				Language: "TypeScript",
				Imports: []ImportStatement{
					{Raw: "import { Router } from 'express'", Source: "express"},
					{Raw: "import { UserService } from '../services/userService'", Source: "../services/userService"},
				},
				Signatures: []Signature{
					{Kind: "class", Name: "UserController", Line: 5, Raw: "export class UserController"},
					{Kind: "method", Name: "getUsers", Line: 10, Raw: "async getUsers(req: Request, res: Response)"},
				},
				Decorators: []string{"@Controller('/users')"},
			},
			{
				Path:     "/workspace/test-repo/src/services/userService.ts",
				RelPath:  "src/services/userService.ts",
				Language: "TypeScript",
				Imports: []ImportStatement{
					{Raw: "import { User } from '../models/user'", Source: "../models/user"},
				},
				Signatures: []Signature{
					{Kind: "class", Name: "UserService", Line: 3, Raw: "export class UserService"},
					{Kind: "method", Name: "findAll", Line: 8, Raw: "async findAll(): Promise<User[]>"},
				},
			},
		},
		Patterns: pats,
	}

	result, err := RenderAnalysisPrompt(data)
	if err != nil {
		t.Fatalf("RenderAnalysisPrompt() returned error: %v", err)
	}

	// Verify key sections are present.
	checks := []struct {
		name    string
		substr  string
	}{
		{"role", "senior software architect"},
		{"repo id", "test-repo"},
		{"tree section", "controllers/"},
		{"tech stack express", "express"},
		{"tech stack typescript", "typescript"},
		{"file signature path", "src/controllers/userController.ts"},
		{"import statement", "import { Router } from 'express'"},
		{"signature name", "UserController"},
		{"signature kind", "[class]"},
		{"decorator", "@Controller('/users')"},
		{"pattern name hex", "Hexagonal Architecture"},
		{"pattern name mvc", "Model-View-Controller"},
		{"layer core", "core"},
		{"layer adapters", "adapters"},
		{"json schema", "$schema"},
		{"example output", "my-app"},
		{"output rules", "Output ONLY valid JSON"},
		{"stats total files", "5"},
		{"stats extension", ".ts"},
		{"archetype controller", "controller"},
		{"archetype repository", "repository"},
		{"flow patterns", "HTTP Request"},
		{"anti patterns", "business logic"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(result, c.substr) {
				t.Errorf("expected output to contain %q", c.substr)
			}
		})
	}

	// Verify it is reasonably long (a good prompt should be substantial).
	if len(result) < 1000 {
		t.Errorf("prompt seems too short: %d bytes", len(result))
	}
}

func TestLoadSchema(t *testing.T) {
	schema, err := LoadSchema()
	if err != nil {
		t.Fatalf("LoadSchema() returned error: %v", err)
	}

	if schema == "" {
		t.Fatal("schema should not be empty")
	}

	// Verify it contains key schema elements.
	checks := []string{
		"repo_id",
		"components",
		"archetypes",
		"relationships",
		"flows",
		"$schema",
		"draft/2020-12",
	}

	for _, c := range checks {
		if !strings.Contains(schema, c) {
			t.Errorf("schema should contain %q", c)
		}
	}
}

func TestRenderAnalysisPromptMinimalData(t *testing.T) {
	// Test with minimal data to ensure template does not panic.
	data := PromptData{
		RepoID: "minimal-repo",
		Tree:   "minimal-repo/\n└── main.go",
		Stats: ScanStats{
			TotalFiles:       1,
			TotalDirs:        1,
			FilesByExtension: map[string]int{".go": 1},
		},
		Patterns: []patterns.PatternDef{},
	}

	result, err := RenderAnalysisPrompt(data)
	if err != nil {
		t.Fatalf("RenderAnalysisPrompt() with minimal data returned error: %v", err)
	}

	if !strings.Contains(result, "minimal-repo") {
		t.Error("expected output to contain repo ID")
	}
}
