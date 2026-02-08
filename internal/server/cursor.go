package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// CursorState tracks the editor's current file and broadcasts changes via SSE.
type CursorState struct {
	mu      sync.Mutex
	file    string
	clients map[chan string]struct{}
	idx     *ArchiveIndex
}

// CursorEvent is the JSON payload sent over SSE.
type CursorEvent struct {
	File        string `json:"file"`
	ComponentID string `json:"component_id,omitempty"`
	ArchetypeID string `json:"archetype_id,omitempty"`
}

// NewCursorState creates a CursorState backed by the given index for resolving IDs.
func NewCursorState(idx *ArchiveIndex) *CursorState {
	return &CursorState{
		clients: make(map[chan string]struct{}),
		idx:     idx,
	}
}

// Set updates the current file and broadcasts the resolved event to all SSE clients.
func (cs *CursorState) Set(file string) {
	file = NormalizePath(file)

	ev := CursorEvent{File: file}

	if comp := cs.idx.FindComponent(file); comp != nil {
		ev.ComponentID = comp.ID
	}
	if arch := cs.idx.FindArchetype(file); arch != nil {
		ev.ArchetypeID = arch.Archetype.ID
	}

	data, _ := json.Marshal(ev)
	msg := string(data)

	cs.mu.Lock()
	cs.file = file
	for ch := range cs.clients {
		select {
		case ch <- msg:
		default:
			// Drop if client is slow
		}
	}
	cs.mu.Unlock()
}

// Subscribe returns a channel that receives SSE messages.
func (cs *CursorState) Subscribe() chan string {
	ch := make(chan string, 8)
	cs.mu.Lock()
	cs.clients[ch] = struct{}{}
	cs.mu.Unlock()
	return ch
}

// Unsubscribe removes a client channel and closes it.
func (cs *CursorState) Unsubscribe(ch chan string) {
	cs.mu.Lock()
	delete(cs.clients, ch)
	cs.mu.Unlock()
	close(ch)
}

// handleCursorPut handles PUT /cursor?file=<path>
func handleCursorPut(cs *CursorState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file parameter is required"})
			return
		}
		cs.Set(file)
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleCursorStream handles GET /cursor/stream (SSE)
func handleCursorStream(cs *CursorState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := cs.Subscribe()
		defer cs.Unsubscribe(ch)

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	}
}
