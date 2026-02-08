package schema

// ArchIndex is the root data structure stored in .arch/index.json.
// It represents the full architectural analysis of a codebase.
type ArchIndex struct {
	RepoID        string                `json:"repo_id"`
	Patterns      []string              `json:"patterns"`
	Components    []Component           `json:"components"`
	Archetypes    map[string][]Archetype `json:"archetypes"`
	Relationships []Relationship        `json:"relationships"`
	Flows         []Flow                `json:"flows,omitempty"`
}

// Component represents a logical grouping of code (e.g., a microservice,
// a bounded context, a subsystem).
type Component struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Layer          string    `json:"layer"`
	CodeRefs       []string  `json:"code_refs"`
	Provides       *Provides `json:"provides,omitempty"`
	NestedAnalysis string    `json:"nested_analysis,omitempty"`
	Analyzed       bool      `json:"analyzed"`
}

// Provides describes what a component exports (interfaces, symbols).
type Provides struct {
	Interface string   `json:"interface,omitempty"`
	Symbols   []string `json:"symbols,omitempty"`
}

// Archetype represents a specific code element classified by its architectural
// role (e.g., controller, repository, middleware, service).
type Archetype struct {
	ID             string   `json:"id"`
	File           string   `json:"file"`
	Symbol         string   `json:"symbol,omitempty"`
	Technology     string   `json:"technology,omitempty"`
	EntryPointType string   `json:"entry_point_type,omitempty"`
	Routes         []string `json:"routes,omitempty"`
	Topics         []string `json:"topics,omitempty"`
	AppliesTo      []string `json:"applies_to,omitempty"`
	Order          int      `json:"order,omitempty"`
	Purpose        string   `json:"purpose,omitempty"`
	Entity         string   `json:"entity,omitempty"`
	TargetService  string   `json:"target_service,omitempty"`
}

// Relationship describes a dependency or interaction between two
// components or archetypes.
type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Flow string `json:"flow,omitempty"`
}

// Flow describes a path that a request or data takes through the system.
type Flow struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Steps   []string `json:"steps"`
	Pattern string   `json:"pattern,omitempty"`
}

// Config represents the user configuration stored in .arch/config.json.
type Config struct {
	Version         string   `json:"version"`
	RepoID          string   `json:"repo_id"`
	IgnorePatterns  []string `json:"ignore_patterns"`
	MaxFileSizeBytes int64   `json:"max_file_size_bytes"`
	SampleFileCount int      `json:"sample_file_count"`
}

// DefaultConfig returns sensible defaults for a new project.
func DefaultConfig(repoID string) Config {
	return Config{
		Version: "0.1.0",
		RepoID:  repoID,
		IgnorePatterns: []string{
			"node_modules", "vendor", ".git", "dist", "build",
			"__pycache__", ".venv", "target", ".idea", ".vscode",
		},
		MaxFileSizeBytes: 1 << 20, // 1MB
		SampleFileCount:  5,
	}
}
