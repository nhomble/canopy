package server

import (
	_ "embed"
	"net/http"
)

//go:embed web/index.html
var indexHTML []byte

//go:embed web/favicon.png
var faviconPNG []byte

func handleUI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	}
}

func handleFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(faviconPNG)
	}
}
