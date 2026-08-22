package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/page"
	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

//go:embed static/*
var static embed.FS
var staticFiles = must(fs.Sub(static, "static"))

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func NewWebRouter(noteOps ops.NoteOps, renderer *page.Renderer) http.Handler {
	mux := http.NewServeMux()

	homeRouter := HomeRouter(noteOps, renderer)
	assetsRouter := http.StripPrefix(
		routes.Assets,
		http.FileServer(http.FS(staticFiles)),
	)
	staticRouter := http.FileServer(http.FS(staticFiles))
	notesRouter := http.StripPrefix(
		routes.Notes,
		NewNotesRouter(noteOps, renderer),
	)
	tagsRouter := http.StripPrefix(
		routes.Tags,
		NewTagsRouter(noteOps, renderer),
	)
	notFoundRouter := NotFoundRouter(renderer)

	mux.Handle(
		"GET "+routes.Home+"{$}",
		homeRouter,
	)
	mux.Handle(
		"GET "+routes.Feed,
		homeRouter,
	)
	mux.Handle(
		"GET "+routes.SiteMap,
		homeRouter,
	)

	mux.Handle(
		"GET "+routes.Robots,
		staticRouter,
	)

	mux.Handle(
		"GET "+routes.AssetsPrefix,
		assetsRouter,
	)

	mux.Handle(
		"GET "+routes.NotesPrefix+"{$}",
		http.RedirectHandler(routes.Home, http.StatusTemporaryRedirect),
	)
	mux.Handle(
		"GET "+routes.NotesPrefix+"{slug}/{$}",
		notesRouter,
	)

	mux.Handle(
		"GET "+routes.TagsPrefix+"{label}/{$}",
		tagsRouter,
	)
	mux.Handle(
		"GET "+routes.TagsPrefix+"{label}"+routes.Feed,
		tagsRouter,
	)

	mux.HandleFunc(
		"GET /",
		notFoundRouter,
	)

	return mux
}
