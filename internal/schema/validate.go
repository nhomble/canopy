package schema

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ValidationResult holds the outcome of validating an ArchIndex.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// ValidationError describes a specific validation failure.
type ValidationError struct {
	Path    string // e.g., "components[0].id"
	Message string
}

func (e ValidationError) String() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

// ValidateIndex performs structural and semantic validation on an ArchIndex.
func ValidateIndex(idx *ArchIndex) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Required fields
	if idx.RepoID == "" {
		result.addError("repo_id", "repo_id is required")
	}
	if len(idx.Components) == 0 {
		result.addError("components", "at least one component is required")
	}

	// Collect all known IDs
	ids := make(map[string]bool)

	// Validate components
	for i, comp := range idx.Components {
		prefix := fmt.Sprintf("components[%d]", i)
		if comp.ID == "" {
			result.addError(prefix+".id", "component id is required")
		} else if ids[comp.ID] {
			result.addError(prefix+".id", fmt.Sprintf("duplicate id: %s", comp.ID))
		} else {
			ids[comp.ID] = true
		}
		if comp.Name == "" {
			result.addError(prefix+".name", "component name is required")
		}
		if comp.Layer == "" {
			result.addError(prefix+".layer", "component layer is required")
		}
		if len(comp.CodeRefs) == 0 {
			result.addError(prefix+".code_refs", "at least one code_ref is required")
		}
		for j, ref := range comp.CodeRefs {
			if !isValidGlob(ref) {
				result.addError(fmt.Sprintf("%s.code_refs[%d]", prefix, j),
					fmt.Sprintf("invalid glob pattern: %s", ref))
			}
		}
	}

	// Validate archetypes
	for category, archetypes := range idx.Archetypes {
		for i, arch := range archetypes {
			prefix := fmt.Sprintf("archetypes.%s[%d]", category, i)
			if arch.ID == "" {
				result.addError(prefix+".id", "archetype id is required")
			} else if ids[arch.ID] {
				result.addError(prefix+".id", fmt.Sprintf("duplicate id: %s", arch.ID))
			} else {
				ids[arch.ID] = true
			}
			if arch.File == "" {
				result.addError(prefix+".file", "archetype file is required")
			}
		}
	}

	// Validate relationships reference valid IDs
	for i, rel := range idx.Relationships {
		prefix := fmt.Sprintf("relationships[%d]", i)
		if rel.From == "" {
			result.addError(prefix+".from", "from is required")
		} else if !ids[rel.From] {
			result.addWarning(fmt.Sprintf("%s.from: references unknown id: %s", prefix, rel.From))
		}
		if rel.To == "" {
			result.addError(prefix+".to", "to is required")
		} else if !ids[rel.To] {
			result.addWarning(fmt.Sprintf("%s.to: references unknown id: %s", prefix, rel.To))
		}
		if rel.Type == "" {
			result.addError(prefix+".type", "relationship type is required")
		}
	}

	// Validate flows reference valid IDs
	for i, flow := range idx.Flows {
		prefix := fmt.Sprintf("flows[%d]", i)
		if flow.ID == "" {
			result.addError(prefix+".id", "flow id is required")
		}
		if flow.Name == "" {
			result.addError(prefix+".name", "flow name is required")
		}
		if len(flow.Steps) == 0 {
			result.addError(prefix+".steps", "at least one step is required")
		}
		for j, step := range flow.Steps {
			if !ids[step] {
				result.addWarning(fmt.Sprintf("%s.steps[%d]: references unknown id: %s", prefix, j, step))
			}
		}
	}

	return result
}

func (r *ValidationResult) addError(path, msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{Path: path, Message: msg})
}

func (r *ValidationResult) addWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

// FormatResult returns a human-readable string of the validation result.
func (r *ValidationResult) FormatResult() string {
	var sb strings.Builder
	if r.Valid {
		sb.WriteString("Validation passed.\n")
	} else {
		sb.WriteString("Validation failed.\n")
	}
	for _, e := range r.Errors {
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n", e))
	}
	for _, w := range r.Warnings {
		sb.WriteString(fmt.Sprintf("  WARNING: %s\n", w))
	}
	return sb.String()
}

func isValidGlob(pattern string) bool {
	_, err := doublestar.Match(pattern, "test")
	return err == nil
}
