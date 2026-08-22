package page

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"

	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

var (
	//go:embed templates
	templates embed.FS
)

type Renderer struct {
	pages map[string]*template.Template
}

var templateFuncs = template.FuncMap{
	"homeURL":    func() string { return routes.Home },
	"notesURL":   func() string { return routes.Notes },
	"feedURL":    func() string { return routes.Feed },
	"assetURL":   routes.AssetURL,
	"noteURL":    routes.NoteURL,
	"tagURL":     routes.TagURL,
	"tagFeedURL": routes.TagFeedURL,
}

func NewRenderer() (*Renderer, error) {
	base, err := template.
		New("site").
		Funcs(templateFuncs).
		ParseFS(
			templates,
			"templates/layout.gohtml",
			"templates/partials/*.gohtml",
		)
	if err != nil {
		return nil, err
	}

	pagePaths, err := fs.Glob(templates, "templates/pages/*.gohtml")
	if err != nil {
		return nil, err
	}
	if len(pagePaths) == 0 {
		return nil, errors.New("no pages found to render")
	}

	pages := make(map[string]*template.Template, len(pagePaths))

	for _, p := range pagePaths {
		clone, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("clone page error for %q: %w", p, err)
		}

		page, err := clone.ParseFS(templates, p)
		if err != nil {
			return nil, fmt.Errorf("page parse error for %q: %w", p, err)
		}

		pageName := path.Base(p)
		pages[pageName] = page
	}

	return &Renderer{pages}, nil
}

func (r *Renderer) Render(w io.Writer, n string, m any) error {
	t, ok := r.pages[n]
	if !ok {
		return fmt.Errorf("page %q was not found", n)
	}

	if err := t.ExecuteTemplate(w, "layout", m); err != nil {
		return fmt.Errorf("error executing template for page %q: %w", n, err)
	}

	return nil
}
