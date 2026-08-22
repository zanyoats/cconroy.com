package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/zanyoats/cconroy.com/internal/note"
	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/page"
	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

func HomeRouter(noteOps ops.NoteOps, renderer *page.Renderer) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /{$}",
		showHome(noteOps, renderer),
	)
	mux.HandleFunc(
		"GET "+routes.Feed,
		FeedHandler(noteOps, feedHandlerOpts{}),
	)
	mux.HandleFunc(
		"GET "+routes.SiteMap,
		SitemapHandler(noteOps),
	)

	return mux
}

func showHome(noteOps ops.NoteOps, renderer *page.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notes, err := noteOps.GetAllNotes(r.Context())
		switch {
		case err == nil:
			break

		case errors.Is(err, context.Canceled):
			return

		case errors.Is(err, context.DeadlineExceeded):
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return

		default:
			InternalServerErrorRouter(renderer, fmt.Sprintf("error getting notes: %v", err))(w, r)
			return
		}

		var buf bytes.Buffer
		type viewModel struct {
			NoteGroups []notesByYear
			HasMore    bool
		}
		noteGroups := groupNotesByYear(notes)
		m := viewModel{noteGroups, false}
		if err := renderer.Render(&buf, "home.gohtml", m); err != nil {
			InternalServerErrorRouter(renderer, fmt.Sprintf("error rendering home: %v", err))(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("error writing response: %v", err)
		}
	}
}

type notesByYear struct {
	Year  int
	Notes []note.Note
}

func groupNotesByYear(notes []note.Note) []notesByYear {
	groups := make([]notesByYear, 0)

	for _, n := range notes {
		year := n.Date.Year()

		if len(groups) == 0 || groups[len(groups)-1].Year != year {
			groups = append(groups, notesByYear{Year: year})
		}

		i := len(groups) - 1
		groups[i].Notes = append(groups[i].Notes, n)
	}

	return groups
}
