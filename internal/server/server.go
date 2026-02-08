package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run loads the index and starts the HTTP server. It blocks until shutdown.
func Run(indexPath string, host string, port int) error {
	idx, err := LoadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("loading index: %w", err)
	}

	cs := NewCursorState(idx)

	mux := http.NewServeMux()
	SetupRoutes(mux, idx, cs)

	handler := corsMiddleware(mux)

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{Addr: addr, Handler: handler}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	archetypeCount := 0
	for _, a := range idx.Raw.Archetypes {
		archetypeCount += len(a)
	}

	log.Printf("arch-index server listening on http://%s (open in browser for graph UI)", addr)
	log.Printf("Loaded: %d components, %d archetypes, %d relationships, %d flows",
		len(idx.Raw.Components), archetypeCount,
		len(idx.Raw.Relationships), len(idx.Raw.Flows))

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
