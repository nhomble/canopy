package schema

import (
	"os"
	"testing"
)

func TestValidateGoldenIndex(t *testing.T) {
	idx, err := LoadIndex("../../testdata/golden/index.json")
	if err != nil {
		t.Fatalf("loading golden index: %v", err)
	}

	result := ValidateIndex(idx)
	if !result.Valid {
		t.Fatalf("golden index should be valid:\n%s", result.FormatResult())
	}

	if len(idx.Components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(idx.Components))
	}
}

func TestValidateRejectsEmpty(t *testing.T) {
	idx := &ArchIndex{}
	result := ValidateIndex(idx)
	if result.Valid {
		t.Fatal("empty index should fail validation")
	}
}

func TestValidateRejectsDuplicateIDs(t *testing.T) {
	idx := &ArchIndex{
		RepoID: "test",
		Components: []Component{
			{ID: "dup", Name: "A", Layer: "core", CodeRefs: []string{"a/**"}},
			{ID: "dup", Name: "B", Layer: "core", CodeRefs: []string{"b/**"}},
		},
		Archetypes:    map[string][]Archetype{},
		Relationships: []Relationship{},
	}
	result := ValidateIndex(idx)
	if result.Valid {
		t.Fatal("duplicate IDs should fail validation")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "clean JSON",
			input: `{"repo_id": "test"}`,
		},
		{
			name:  "markdown fenced",
			input: "Here is the analysis:\n```json\n{\"repo_id\": \"test\"}\n```\n",
		},
		{
			name:  "with commentary",
			input: "I analyzed the codebase. Here's the result:\n{\"repo_id\": \"test\"}\nHope this helps!",
		},
		{
			name:  "trailing comma",
			input: `{"repo_id": "test", "patterns": ["a",]}`,
		},
		{
			name:    "no JSON",
			input:   "This is just text with no JSON",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ExtractJSON(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("expected non-empty JSON data")
			}
		})
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	idx := &ArchIndex{
		RepoID:   "test",
		Patterns: []string{"hexagonal"},
		Components: []Component{
			{ID: "comp1", Name: "Component 1", Layer: "core", CodeRefs: []string{"src/**"}},
		},
		Archetypes:    map[string][]Archetype{},
		Relationships: []Relationship{},
	}

	tmpFile := t.TempDir() + "/index.json"
	if err := SaveIndex(tmpFile, idx); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadIndex(tmpFile)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.RepoID != idx.RepoID {
		t.Fatalf("repo_id mismatch: %s vs %s", loaded.RepoID, idx.RepoID)
	}
	if len(loaded.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(loaded.Components))
	}

	// Cleanup
	os.Remove(tmpFile)
}
