package prompt

import (
	"strings"
	"testing"

	"github.com/nhomble/canopy/internal/patterns"
)

func TestRenderAnalysisPrompt(t *testing.T) {
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
		Patterns: pats,
	}

	result, err := RenderAnalysisPrompt(data)
	if err != nil {
		t.Fatalf("RenderAnalysisPrompt() returned error: %v", err)
	}

	checks := []struct {
		name   string
		substr string
	}{
		{"role", "analyzing the structure of a codebase"},
		{"repo id", "test-repo"},
		{"tree section", "controllers/"},
		{"pattern name hex", "Hexagonal Architecture"},
		{"pattern name mvc", "Model-View-Controller"},
		{"layer core", "core"},
		{"layer adapters", "adapters"},
		{"json schema", "$schema"},
		{"example output", "my-app"},
		{"output rules", "Output ONLY valid JSON"},
		{"stats extension", ".ts"},
		{"archetype controller", "controller"},
		{"archetype repository", "repository"},
		{"flow patterns", "HTTP Request"},
		{"anti patterns", "business logic"},
		{"step 1", "Partition into applications"},
		{"step 4 constraint", "Do NOT create archetypes"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(result, c.substr) {
				t.Errorf("expected output to contain %q", c.substr)
			}
		})
	}

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
