package patterns

import (
	"embed"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

//go:embed definitions/*.yaml
var patternFS embed.FS

// PatternDef represents a parsed architectural pattern definition.
type PatternDef struct {
	Name         string                  `yaml:"name"`
	Aliases      []string                `yaml:"aliases"`
	Description  string                  `yaml:"description"`
	Layers       []LayerDef              `yaml:"layers"`
	Archetypes   map[string]ArchetypeDef `yaml:"archetypes"`
	FlowPatterns []string                `yaml:"flow_patterns"`
	AntiPatterns []string                `yaml:"anti_patterns"`
	Questions    []QuestionDef           `yaml:"questions"`
}

// LayerDef describes one architectural layer within a pattern.
type LayerDef struct {
	ID              string   `yaml:"id"`
	Description     string   `yaml:"description"`
	Characteristics []string `yaml:"characteristics"`
	TypicalFiles    []string `yaml:"typical_files"`
}

// ArchetypeDef describes a code archetype (e.g., controller, repository).
type ArchetypeDef struct {
	Description         string   `yaml:"description"`
	TypicalTechnologies []string `yaml:"typical_technologies"`
	TypicalFiles        []string `yaml:"typical_files"`
	LivesIn             string   `yaml:"lives_in"`
}

// QuestionDef holds diagnostic prompts scoped to a layer.
type QuestionDef struct {
	Layer   string   `yaml:"layer"`
	Prompts []string `yaml:"prompts"`
}

// LoadAll reads every .yaml file from the embedded definitions directory,
// parses each into a PatternDef, and returns the full set.
func LoadAll() ([]PatternDef, error) {
	var patterns []PatternDef

	entries, err := fs.ReadDir(patternFS, "definitions")
	if err != nil {
		return nil, fmt.Errorf("reading pattern definitions dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := patternFS.ReadFile("definitions/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var p PatternDef
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		patterns = append(patterns, p)
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no pattern definitions found")
	}

	return patterns, nil
}
