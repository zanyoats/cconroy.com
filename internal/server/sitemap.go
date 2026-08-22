package server

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

type URLSetDocument struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func SitemapHandler(
	noteOps ops.NoteOps,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tags []ops.Tag
		notes, err := noteOps.GetAllNotes(r.Context())
		if err == nil {
			tags, err = noteOps.GetAllTags(r.Context())
		}

		switch {
		case err == nil:
			break

		case errors.Is(err, context.Canceled):
			return

		case errors.Is(err, context.DeadlineExceeded):
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return

		default:
			log.Printf("error getting notes or tags for sitemap: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		doc := URLSetDocument{
			XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  make([]SitemapURL, 1, 1+len(notes)+len(tags)),
		}
		doc.URLs[0].Loc = routes.SiteURL + routes.Home

		var homeUpdatedAt time.Time
		for _, n := range notes {
			if n.UpdatedAt.After(homeUpdatedAt) {
				homeUpdatedAt = n.UpdatedAt
			}

			doc.URLs = append(doc.URLs, SitemapURL{
				Loc:     routes.SiteURL + routes.NoteURL(n.Slug),
				LastMod: n.UpdatedAt.UTC().Format(time.RFC3339),
			})
		}
		if !homeUpdatedAt.IsZero() {
			doc.URLs[0].LastMod = homeUpdatedAt.UTC().Format(time.RFC3339)
		}
		for _, tag := range tags {
			doc.URLs = append(doc.URLs, SitemapURL{
				Loc:     routes.SiteURL + routes.TagURL(tag.Label),
				LastMod: tag.UpdatedAt.UTC().Format(time.RFC3339),
			})
		}

		var body bytes.Buffer
		body.WriteString(xml.Header)

		encoder := xml.NewEncoder(&body)
		encoder.Indent("", "  ")
		if err := encoder.Encode(doc); err != nil {
			log.Printf("error encoding sitemap: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		if _, err := body.WriteTo(w); err != nil {
			log.Printf("error writing sitemap response: %v", err)
		}
	}
}
