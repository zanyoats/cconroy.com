package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/page"
	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

type tagsRouter struct {
	noteOps  ops.NoteOps
	renderer *page.Renderer
}

func NewTagsRouter(noteOps ops.NoteOps, renderer *page.Renderer) http.Handler {
	router := &tagsRouter{
		noteOps:  noteOps,
		renderer: renderer,
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /{label}/{$}",
		router.showTag,
	)
	mux.HandleFunc(
		"GET /{label}"+routes.Feed,
		FeedHandler(noteOps, feedHandlerOpts{
			tagPathValue: "label",
		}),
	)

	return mux
}

func (t *tagsRouter) showTag(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("label")
	notes, err := t.noteOps.GetNotesByTag(r.Context(), tag)

	switch {
	case err == nil:
		break

	case errors.Is(err, ops.ErrTagNotFound):
		NotFoundRouter(t.renderer)(w, r)
		return

	case errors.Is(err, context.Canceled):
		return

	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "request timed out", http.StatusGatewayTimeout)
		return

	default:
		InternalServerErrorRouter(t.renderer, fmt.Sprintf("error getting tag: %v", err))(w, r)
		return
	}

	var buf bytes.Buffer
	type viewModel struct {
		NoteGroups []notesByYear
		Tag        string
	}
	noteGroups := groupNotesByYear(notes)
	m := viewModel{noteGroups, tag}
	if err := t.renderer.Render(&buf, "tag.gohtml", m); err != nil {
		InternalServerErrorRouter(t.renderer, fmt.Sprintf("error rendering tag: %v", err))(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
