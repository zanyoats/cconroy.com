package server

import (
	"bytes"
	"log"
	"net/http"

	"github.com/zanyoats/cconroy.com/internal/page"
)

func InternalServerErrorRouter(renderer *page.Renderer, logmsg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print(logmsg)

		var buf bytes.Buffer

		if err := renderer.Render(&buf, "500.gohtml", nil); err != nil {
			log.Printf("error rendering 500 page: %v", err)
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)

		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("error writing response: %v", err)
		}
	}
}
