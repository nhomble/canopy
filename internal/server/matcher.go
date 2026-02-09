package server

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/nhomble/canopy/internal/schema"
)

// FindComponent returns the component whose code_refs match the given file path.
// When multiple patterns match, the most specific one wins.
func (idx *ArchiveIndex) FindComponent(filePath string) *schema.Component {
	filePath = filepath.ToSlash(filePath)

	var bestMatch *schema.Component
	bestSpecificity := -1

	for _, entry := range idx.codeRefEntries {
		matched, err := doublestar.Match(entry.Pattern, filePath)
		if err != nil || !matched {
			continue
		}
		specificity := nonGlobPrefixLen(entry.Pattern)
		if specificity > bestSpecificity {
			bestSpecificity = specificity
			bestMatch = entry.Component
		}
	}

	return bestMatch
}

// FindArchetype returns the archetype entry for an exact file path match.
func (idx *ArchiveIndex) FindArchetype(filePath string) *archetypeEntry {
	filePath = filepath.ToSlash(filePath)
	entries := idx.archetypeByFile[filePath]
	if len(entries) > 0 {
		return entries[0]
	}
	return nil
}

// FindFlows returns all flows that include the given ID as a step.
func (idx *ArchiveIndex) FindFlows(id string) []schema.Flow {
	return idx.flowsByStep[id]
}

// FindRelationships returns relationships matching the given symbol and direction.
func (idx *ArchiveIndex) FindRelationships(symbol, direction string) []schema.Relationship {
	switch direction {
	case "upstream":
		return idx.relsByTo[symbol]
	case "downstream":
		return idx.relsByFrom[symbol]
	default: // "both" or empty
		var result []schema.Relationship
		result = append(result, idx.relsByFrom[symbol]...)
		result = append(result, idx.relsByTo[symbol]...)
		return result
	}
}

// nonGlobPrefixLen returns the length of the path prefix before any glob chars.
func nonGlobPrefixLen(pattern string) int {
	for i, c := range pattern {
		if c == '*' || c == '?' || c == '[' || c == '{' {
			return i
		}
	}
	return len(pattern)
}

// NormalizePath converts a file path to forward slashes and strips leading ./ or /
func NormalizePath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	return p
}
