package patterns

import (
	"testing"
)

func TestLoadAll(t *testing.T) {
	patterns, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() returned error: %v", err)
	}

	if len(patterns) < 2 {
		t.Fatalf("expected at least 2 patterns, got %d", len(patterns))
	}

	// Build a lookup by name for easier assertions.
	byName := make(map[string]PatternDef)
	for _, p := range patterns {
		byName[p.Name] = p
	}

	t.Run("hexagonal architecture loads", func(t *testing.T) {
		hex, ok := byName["Hexagonal Architecture"]
		if !ok {
			t.Fatal("Hexagonal Architecture pattern not found")
		}

		// Aliases
		if len(hex.Aliases) < 2 {
			t.Errorf("expected at least 2 aliases, got %d", len(hex.Aliases))
		}

		// Layers
		if len(hex.Layers) != 3 {
			t.Errorf("expected 3 layers, got %d", len(hex.Layers))
		}
		layerIDs := make(map[string]bool)
		for _, l := range hex.Layers {
			layerIDs[l.ID] = true
		}
		for _, expected := range []string{"core", "ports", "adapters"} {
			if !layerIDs[expected] {
				t.Errorf("missing layer %q", expected)
			}
		}

		// Archetypes
		expectedArchetypes := []string{"controller", "repository", "service", "port", "adapter"}
		for _, name := range expectedArchetypes {
			if _, ok := hex.Archetypes[name]; !ok {
				t.Errorf("missing archetype %q", name)
			}
		}

		// Flow patterns
		if len(hex.FlowPatterns) == 0 {
			t.Error("expected at least one flow pattern")
		}

		// Anti patterns
		if len(hex.AntiPatterns) == 0 {
			t.Error("expected at least one anti-pattern")
		}

		// Questions
		if len(hex.Questions) == 0 {
			t.Error("expected at least one question block")
		}

		// Description should not be empty
		if hex.Description == "" {
			t.Error("description should not be empty")
		}

		// Core layer should have characteristics
		for _, l := range hex.Layers {
			if l.ID == "core" {
				if len(l.Characteristics) == 0 {
					t.Error("core layer should have characteristics")
				}
				if len(l.TypicalFiles) == 0 {
					t.Error("core layer should have typical_files")
				}
			}
		}
	})

	t.Run("mvc pattern loads", func(t *testing.T) {
		mvc, ok := byName["Model-View-Controller"]
		if !ok {
			t.Fatal("Model-View-Controller pattern not found")
		}

		// Aliases
		found := false
		for _, a := range mvc.Aliases {
			if a == "MVC" {
				found = true
			}
		}
		if !found {
			t.Error("expected MVC alias")
		}

		// Layers
		if len(mvc.Layers) != 3 {
			t.Errorf("expected 3 layers, got %d", len(mvc.Layers))
		}

		layerIDs := make(map[string]bool)
		for _, l := range mvc.Layers {
			layerIDs[l.ID] = true
		}
		for _, expected := range []string{"models", "views", "controllers"} {
			if !layerIDs[expected] {
				t.Errorf("missing layer %q", expected)
			}
		}

		// Archetypes
		expectedArchetypes := []string{"model", "view", "controller"}
		for _, name := range expectedArchetypes {
			if _, ok := mvc.Archetypes[name]; !ok {
				t.Errorf("missing archetype %q", name)
			}
		}

		// Flow patterns
		if len(mvc.FlowPatterns) == 0 {
			t.Error("expected at least one flow pattern")
		}

		// Anti patterns
		if len(mvc.AntiPatterns) == 0 {
			t.Error("expected at least one anti-pattern")
		}
	})

	t.Run("archetype fields populated", func(t *testing.T) {
		hex := byName["Hexagonal Architecture"]
		ctrl := hex.Archetypes["controller"]
		if ctrl.Description == "" {
			t.Error("controller description should not be empty")
		}
		if len(ctrl.TypicalTechnologies) == 0 {
			t.Error("controller should have typical technologies")
		}
		if len(ctrl.TypicalFiles) == 0 {
			t.Error("controller should have typical files")
		}

		svc := hex.Archetypes["service"]
		if svc.LivesIn != "core" {
			t.Errorf("service lives_in should be 'core', got %q", svc.LivesIn)
		}
	})
}
