package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nhomble/canopy/internal/schema"
)

// ArchiveIndex holds the parsed ArchIndex data with pre-built lookup maps
// for O(1) queries on components, archetypes, relationships, and flows.
type ArchiveIndex struct {
	Raw *schema.ArchIndex

	componentByID        map[string]*schema.Component
	archetypeByID        map[string]*archetypeEntry
	archetypeByFile      map[string][]*archetypeEntry
	archetypeToComponent map[string]string // archetype ID → component ID
	codeRefEntries       []codeRefEntry
	relsByFrom           map[string][]schema.Relationship
	relsByTo             map[string][]schema.Relationship
	flowsByStep          map[string][]schema.Flow
}

type codeRefEntry struct {
	Pattern     string
	ComponentID string
	Component   *schema.Component
}

type archetypeEntry struct {
	Category  string
	Archetype *schema.Archetype
}

// LoadIndex reads an index.json file and builds the in-memory index.
func LoadIndex(indexPath string) (*ArchiveIndex, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("reading index file: %w", err)
	}

	var raw schema.ArchIndex
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing index JSON: %w", err)
	}

	return NewIndex(&raw), nil
}

// NewIndex builds an ArchiveIndex from a raw ArchIndex struct.
func NewIndex(raw *schema.ArchIndex) *ArchiveIndex {
	idx := &ArchiveIndex{
		Raw:                  raw,
		componentByID:        make(map[string]*schema.Component, len(raw.Components)),
		archetypeByID:        make(map[string]*archetypeEntry),
		archetypeByFile:      make(map[string][]*archetypeEntry),
		archetypeToComponent: make(map[string]string),
		relsByFrom:           make(map[string][]schema.Relationship),
		relsByTo:             make(map[string][]schema.Relationship),
		flowsByStep:          make(map[string][]schema.Flow),
	}

	for i := range raw.Components {
		comp := &raw.Components[i]
		idx.componentByID[comp.ID] = comp
		for _, pattern := range comp.CodeRefs {
			idx.codeRefEntries = append(idx.codeRefEntries, codeRefEntry{
				Pattern:     pattern,
				ComponentID: comp.ID,
				Component:   comp,
			})
		}
	}

	for category, archetypes := range raw.Archetypes {
		for i := range archetypes {
			arch := &raw.Archetypes[category][i]
			entry := &archetypeEntry{
				Category:  category,
				Archetype: arch,
			}
			idx.archetypeByID[arch.ID] = entry
			idx.archetypeByFile[arch.File] = append(idx.archetypeByFile[arch.File], entry)
		}
	}

	// Map each archetype to its parent component via code_refs glob matching.
	for _, archetypes := range raw.Archetypes {
		for _, arch := range archetypes {
			if comp := idx.FindComponent(arch.File); comp != nil {
				idx.archetypeToComponent[arch.ID] = comp.ID
			}
		}
	}

	for _, rel := range raw.Relationships {
		idx.relsByFrom[rel.From] = append(idx.relsByFrom[rel.From], rel)
		idx.relsByTo[rel.To] = append(idx.relsByTo[rel.To], rel)
	}

	for _, flow := range raw.Flows {
		for _, step := range flow.Steps {
			idx.flowsByStep[step] = append(idx.flowsByStep[step], flow)
		}
	}

	return idx
}

// Graph payload types for the /graph endpoint.

type GraphPayload struct {
	RepoID         string           `json:"repo_id"`
	Patterns       []string         `json:"patterns"`
	Components     []GraphComponent `json:"components"`
	Relationships  []schema.Relationship `json:"relationships"`
	ComponentEdges []ComponentEdge  `json:"component_edges"`
	Flows          []schema.Flow    `json:"flows"`
}

type GraphComponent struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Layer      string           `json:"layer"`
	Archetypes []GraphArchetype `json:"archetypes"`
}

type GraphArchetype struct {
	ID         string `json:"id"`
	Category   string `json:"category"`
	Symbol     string `json:"symbol,omitempty"`
	Technology string `json:"technology,omitempty"`
	File       string `json:"file"`
	Purpose    string `json:"purpose,omitempty"`
}

type ComponentEdge struct {
	From  string   `json:"from"`
	To    string   `json:"to"`
	Count int      `json:"count"`
	Types []string `json:"types"`
}

// BuildGraphPayload returns pre-computed data for the web UI graph view.
func (idx *ArchiveIndex) BuildGraphPayload() *GraphPayload {
	// Build components with nested archetypes.
	compArchetypes := make(map[string][]GraphArchetype)
	for category, archetypes := range idx.Raw.Archetypes {
		for _, arch := range archetypes {
			ga := GraphArchetype{
				ID:         arch.ID,
				Category:   category,
				Symbol:     arch.Symbol,
				Technology: arch.Technology,
				File:       arch.File,
				Purpose:    arch.Purpose,
			}
			compID := idx.archetypeToComponent[arch.ID]
			if compID != "" {
				compArchetypes[compID] = append(compArchetypes[compID], ga)
			}
		}
	}

	components := make([]GraphComponent, 0, len(idx.Raw.Components))
	for _, comp := range idx.Raw.Components {
		components = append(components, GraphComponent{
			ID:         comp.ID,
			Name:       comp.Name,
			Layer:      comp.Layer,
			Archetypes: compArchetypes[comp.ID],
		})
	}

	// Build aggregated component-level edges.
	type edgeKey struct{ from, to string }
	edgeMap := make(map[edgeKey]*ComponentEdge)
	for _, rel := range idx.Raw.Relationships {
		fromComp := idx.archetypeToComponent[rel.From]
		toComp := idx.archetypeToComponent[rel.To]
		if fromComp == "" || toComp == "" || fromComp == toComp {
			continue
		}
		key := edgeKey{fromComp, toComp}
		edge, ok := edgeMap[key]
		if !ok {
			edge = &ComponentEdge{From: fromComp, To: toComp}
			edgeMap[key] = edge
		}
		edge.Count++
		// Add type if not already present.
		found := false
		for _, t := range edge.Types {
			if t == rel.Type {
				found = true
				break
			}
		}
		if !found {
			edge.Types = append(edge.Types, rel.Type)
		}
	}
	componentEdges := make([]ComponentEdge, 0, len(edgeMap))
	for _, edge := range edgeMap {
		componentEdges = append(componentEdges, *edge)
	}

	return &GraphPayload{
		RepoID:         idx.Raw.RepoID,
		Patterns:       idx.Raw.Patterns,
		Components:     components,
		Relationships:  idx.Raw.Relationships,
		ComponentEdges: componentEdges,
		Flows:          idx.Raw.Flows,
	}
}
