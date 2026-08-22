package server

import (
	"bytes"
	"log"
	"net/http"

	"github.com/zanyoats/cconroy.com/internal/page"
)

func NotFoundRouter(renderer *page.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer

		if err := renderer.Render(&buf, "404.gohtml", nil); err != nil {
			log.Printf("error rendering 404 page: %v", err)
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)

		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("error writing response: %v", err)
		}
	}
}
