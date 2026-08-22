package routes

import "net/url"

const (
	SiteURL      = "https://cconroy.com"
	Home         = "/"
	SiteMap      = "/sitemap.xml"
	Robots       = "/robots.txt"
	Feed         = "/feed.xml"
	Assets       = "/assets"
	AssetsPrefix = "/assets/"
	Notes        = "/notes"
	NotesPrefix  = "/notes/"
	Tags         = "/tags"
	TagsPrefix   = "/tags/"
)

func AssetURL(asset string) string {
	return AssetsPrefix + asset
}

func NoteURL(slug string) string {
	return NotesPrefix + url.PathEscape(slug) + "/"
}

func TagURL(label string) string {
	return TagsPrefix + url.PathEscape(label) + "/"
}

func TagFeedURL(label string) string {
	return TagsPrefix + url.PathEscape(label) + Feed
}
