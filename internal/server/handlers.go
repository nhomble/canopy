package server

import (
	"encoding/json"
	"net/http"
)

// Response types

type ContextResponse struct {
	Component     *ComponentSummary  `json:"component,omitempty"`
	Layer         string             `json:"layer,omitempty"`
	Archetype     *ArchetypeSummary  `json:"archetype,omitempty"`
	Flows         []FlowSummary      `json:"flows,omitempty"`
	ZoomAvailable bool               `json:"zoom_available"`
	ZoomAnalyzed  bool               `json:"zoom_analyzed"`
}

type ComponentSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ArchetypeSummary struct {
	Category   string `json:"category"`
	ID         string `json:"id"`
	Technology string `json:"technology,omitempty"`
}

type FlowSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SetupRoutes registers all HTTP handlers on the given mux.
func SetupRoutes(mux *http.ServeMux, idx *ArchiveIndex, cs *CursorState) {
	mux.HandleFunc("GET /{$}", handleUI())
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /graph", handleGraph(idx))
	mux.HandleFunc("GET /context", handleContext(idx))
	mux.HandleFunc("GET /components", handleComponents(idx))
	mux.HandleFunc("GET /archetypes/{category}", handleArchetypes(idx))
	mux.HandleFunc("GET /relationships", handleRelationships(idx))
	mux.HandleFunc("GET /flows", handleFlows(idx))
	mux.HandleFunc("PUT /cursor", handleCursorPut(cs))
	mux.HandleFunc("GET /cursor/stream", handleCursorStream(cs))
}

func handleGraph(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, idx.BuildGraphPayload())
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleContext(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file parameter is required"})
			return
		}

		file = NormalizePath(file)

		resp := ContextResponse{}

		// Find component
		comp := idx.FindComponent(file)
		if comp != nil {
			resp.Component = &ComponentSummary{ID: comp.ID, Name: comp.Name}
			resp.Layer = comp.Layer
			resp.ZoomAvailable = comp.NestedAnalysis != ""
			resp.ZoomAnalyzed = comp.Analyzed

			// Find flows through this component
			flows := idx.FindFlows(comp.ID)
			for _, f := range flows {
				resp.Flows = append(resp.Flows, FlowSummary{ID: f.ID, Name: f.Name})
			}
		}

		// Find archetype
		arch := idx.FindArchetype(file)
		if arch != nil {
			resp.Archetype = &ArchetypeSummary{
				Category:   arch.Category,
				ID:         arch.Archetype.ID,
				Technology: arch.Archetype.Technology,
			}

			// Also find flows through this archetype
			flows := idx.FindFlows(arch.Archetype.ID)
			for _, f := range flows {
				// Avoid duplicates
				found := false
				for _, existing := range resp.Flows {
					if existing.ID == f.ID {
						found = true
						break
					}
				}
				if !found {
					resp.Flows = append(resp.Flows, FlowSummary{ID: f.ID, Name: f.Name})
				}
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func handleComponents(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"components": idx.Raw.Components,
		})
	}
}

func handleArchetypes(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		archetypes, ok := idx.Raw.Archetypes[category]
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "archetype category not found: " + category,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			category: archetypes,
		})
	}
}

func handleRelationships(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		direction := r.URL.Query().Get("direction")
		if direction == "" {
			direction = "both"
		}

		if symbol == "" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"relationships": idx.Raw.Relationships,
			})
			return
		}

		rels := idx.FindRelationships(symbol, direction)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"relationships": rels,
		})
	}
}

func handleFlows(idx *ArchiveIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		through := r.URL.Query().Get("through")

		if through == "" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"flows": idx.Raw.Flows,
			})
			return
		}

		flows := idx.FindFlows(through)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"flows": flows,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
