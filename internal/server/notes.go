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
)

type notesRouter struct {
	noteOps  ops.NoteOps
	renderer *page.Renderer
}

func NewNotesRouter(noteOps ops.NoteOps, renderer *page.Renderer) http.Handler {
	router := &notesRouter{
		noteOps:  noteOps,
		renderer: renderer,
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /{slug}/{$}",
		router.showNote,
	)

	return mux
}

func (n *notesRouter) showNote(w http.ResponseWriter, r *http.Request) {
	item, err := n.noteOps.GetNote(r.Context(), r.PathValue("slug"))
	switch {
	case err == nil:
		break

	case errors.Is(err, ops.ErrNoteNotFound):
		NotFoundRouter(n.renderer)(w, r)
		return

	case errors.Is(err, context.Canceled):
		return

	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "request timed out", http.StatusGatewayTimeout)
		return

	default:
		InternalServerErrorRouter(n.renderer, fmt.Sprintf("error getting note: %v", err))(w, r)
		return
	}

	type viewModel struct {
		*note.Note
		TOC []*note.TOCItem
	}
	m := viewModel{
		Note: item,
		TOC:  note.BuildTOC(item.Headings),
	}
	var buf bytes.Buffer
	if err := n.renderer.Render(&buf, "note.gohtml", m); err != nil {
		InternalServerErrorRouter(n.renderer, fmt.Sprintf("error rendering note: %v", err))(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
