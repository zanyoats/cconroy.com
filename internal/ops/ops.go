package ops

import (
	"context"
	"time"

	"github.com/zanyoats/cconroy.com/internal/note"
)

type Tag struct {
	Label     string
	UpdatedAt time.Time
}

type FeedResult struct {
	Notes         []note.Note
	LastBuildDate time.Time
}

type NoteOps interface {
	PublishStaticNotes(ctx context.Context) error
	GetNote(ctx context.Context, slug string) (*note.Note, error)
	GetTag(ctx context.Context, label string) (Tag, error)
	GetNotesByTag(ctx context.Context, label string) ([]note.Note, error)
	GetAllNotes(ctx context.Context) ([]note.Note, error)
	GetAllTags(ctx context.Context) ([]Tag, error)
	GetRecentFeedNotes(ctx context.Context, limit int) (FeedResult, error)
}
