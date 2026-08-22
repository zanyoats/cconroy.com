package server

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/zanyoats/cconroy.com/internal/note"
	"github.com/zanyoats/cconroy.com/internal/ops"
	"github.com/zanyoats/cconroy.com/internal/server/routes"
)

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	AtomNS  string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Generator     string      `xml:"generator"`
	Language      string      `xml:"language"`
	LastBuildDate string      `xml:"lastBuildDate,omitempty"`
	AtomLink      rssAtomLink `xml:"atom:link"`
	Items         []rssItem   `xml:"item"`
}

type rssAtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
}

const maxNumNotes = 20

type feedHandlerOpts struct {
	// Empty means the existing site-wide notes feed.
	// Otherwise, this names the mux path value containing the tag.
	tagPathValue string
}

func FeedHandler(
	noteOps ops.NoteOps,
	opts feedHandlerOpts,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			notes       []note.Note
			title       string
			channelURL  string
			feedURL     string
			description string
			lastUpdated time.Time
			err         error
		)

		if opts.tagPathValue == "" {
			var result ops.FeedResult
			result, err = noteOps.GetRecentFeedNotes(r.Context(), maxNumNotes)
			notes = result.Notes
			lastUpdated = result.LastBuildDate

			title = "Notes on cconroy.com"
			channelURL = routes.SiteURL + routes.Notes
			feedURL = routes.SiteURL + routes.Feed
			description = "Recent content in notes on cconroy.com"
		} else {
			label := r.PathValue(opts.tagPathValue)

			notes, err = noteOps.GetNotesByTag(r.Context(), label)
			if err == nil {
				var tag ops.Tag
				tag, err = noteOps.GetTag(r.Context(), label)
				if err == nil {
					lastUpdated = tag.UpdatedAt
				}
			}

			if len(notes) > maxNumNotes {
				notes = notes[:maxNumNotes]
			}

			title = label + " on cconroy.com"
			channelURL = routes.SiteURL + routes.TagURL(label)
			feedURL = routes.SiteURL + routes.TagFeedURL(label)
			description = "Recent content in " + label + " on cconroy.com"
		}

		switch {
		case err == nil:
			break

		case errors.Is(err, ops.ErrTagNotFound):
			http.NotFound(w, r)
			return

		case errors.Is(err, context.Canceled):
			return

		case errors.Is(err, context.DeadlineExceeded):
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return

		default:
			log.Printf("error getting notes for RSS: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feed := rssDocument{
			Version: "2.0",
			AtomNS:  "http://www.w3.org/2005/Atom",
			Channel: rssChannel{
				Title:       title,
				Link:        channelURL,
				Description: description,
				Generator:   "cconroy.com",
				Language:    "en-us",
				AtomLink: rssAtomLink{
					Href: feedURL,
					Rel:  "self",
					Type: "application/rss+xml",
				},
			},
		}

		for _, n := range notes {
			noteURL := routes.SiteURL + routes.NoteURL(n.Slug)

			feed.Channel.Items = append(feed.Channel.Items, rssItem{
				Title:       n.Title,
				Link:        noteURL,
				PubDate:     n.Date.Format(time.RFC1123Z),
				GUID:        noteURL,
				Description: n.Short,
			})
		}

		if !lastUpdated.IsZero() {
			feed.Channel.LastBuildDate = lastUpdated.Format(time.RFC1123Z)
		}

		var body bytes.Buffer
		body.WriteString(xml.Header)

		encoder := xml.NewEncoder(&body)
		encoder.Indent("", "  ")
		if err := encoder.Encode(feed); err != nil {
			log.Printf("error encoding RSS: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set(
			"Content-Type",
			"application/rss+xml; charset=utf-8",
		)
		_, _ = body.WriteTo(w)
	}
}
