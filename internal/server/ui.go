package server

import (
	_ "embed"
	"net/http"
)

//go:embed web/index.html
var indexHTML []byte

func handleUI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	}
}
