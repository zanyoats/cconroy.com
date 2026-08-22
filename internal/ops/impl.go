package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/zanyoats/cconroy.com/internal/note"
)

var (
	ErrNoteNotFound = errors.New("note not found")
	ErrTagNotFound  = errors.New("tag not found")
)

type NotesImpl struct {
	readDB       *sql.DB
	writeDB      *sql.DB
	notesPattern string
}

func NewNotesImpl(readDB *sql.DB, writeDB *sql.DB, notesPath string) *NotesImpl {
	result := new(NotesImpl)
	result.readDB = readDB
	result.writeDB = writeDB
	result.notesPattern = filepath.Join(notesPath, "*.md")
	return result
}

func (s *NotesImpl) PublishStaticNotes(ctx context.Context) error {
	notes, err := s.collectStaticNotes()
	if err != nil {
		return err
	}

	return upsertNotesIntoDB(ctx, s.writeDB, notes)
}

func (s *NotesImpl) GetNote(ctx context.Context, slug string) (*note.Note, error) {
	query := `
	SELECT
		n.id           ,
		'' AS slug     ,
		n.title        ,
		n.date         ,
		'' AS short    ,
		n.body_html    ,
		n.headings     ,
		COALESCE(
			(
				SELECT json_group_array(tag)
				FROM (
					SELECT t.tag
					FROM notes_tags AS nt
					JOIN tags AS t ON t.id = nt.tag_id
					WHERE nt.note_id = n.id
					ORDER BY t.tag
				)
			),
			'[]'
		) AS tags      ,
		n.updated_at
	FROM notes AS n
	WHERE n.slug = ?;
	`
	row := s.readDB.QueryRowContext(ctx, query, slug)
	n, err := scanNote(row)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("slug %q: %w", slug, ErrNoteNotFound)
	} else if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *NotesImpl) GetNotesByTag(ctx context.Context, label string) ([]note.Note, error) {
	notes, err := queryNotes(ctx, s.readDB, noteQueryOptions{tag: label})
	if err != nil {
		return nil, err
	}

	if len(notes) > 0 {
		return notes, nil
	}

	return nil, fmt.Errorf("slug %q: %w", label, ErrTagNotFound)
}

func (s *NotesImpl) GetTag(ctx context.Context, label string) (Tag, error) {
	const query = `
	SELECT
		t.tag,
		t.updated_at
	FROM tags AS t
	WHERE t.tag = ?
		AND EXISTS (
			SELECT 1
			FROM notes_tags AS nt
			WHERE nt.tag_id = t.id
		);
	`

	var tag Tag
	err := s.readDB.QueryRowContext(ctx, query, label).Scan(
		&tag.Label,
		&tag.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, fmt.Errorf("tag %q: %w", label, ErrTagNotFound)
	}
	if err != nil {
		return Tag{}, fmt.Errorf("query tag %q: %w", label, err)
	}

	return tag, nil
}

func (s *NotesImpl) GetRecentFeedNotes(ctx context.Context, limit int) (FeedResult, error) {
	if limit < 0 {
		return FeedResult{}, fmt.Errorf("notes limit cannot be negative")
	}

	notes, err := queryNotes(ctx, s.readDB, noteQueryOptions{limit: &limit})
	if err != nil {
		return FeedResult{}, err
	}

	result := FeedResult{Notes: notes}
	for _, n := range notes {
		if n.UpdatedAt.After(result.LastBuildDate) {
			result.LastBuildDate = n.UpdatedAt
		}
	}

	return result, nil
}

func (s *NotesImpl) GetAllNotes(ctx context.Context) ([]note.Note, error) {
	return queryNotes(ctx, s.readDB, noteQueryOptions{})
}

func (s *NotesImpl) GetAllTags(ctx context.Context) ([]Tag, error) {
	const query = `
	SELECT
		t.tag,
		t.updated_at
	FROM tags AS t
	WHERE EXISTS (
		SELECT 1
		FROM notes_tags AS nt
		WHERE nt.tag_id = t.id
	)
	ORDER BY t.tag COLLATE NOCASE;
	`

	rows, err := s.readDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Label, &tag.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func (s *NotesImpl) collectStaticNotes() ([]note.Note, error) {
	paths, err := filepath.Glob(s.notesPattern)
	if err != nil {
		return []note.Note{}, err
	}

	var notes []note.Note
	for _, p := range paths {
		parsed, err := note.Parse(p)
		if err != nil {
			return []note.Note{}, err
		}

		base := filepath.Base(p)
		ext := filepath.Ext(p)
		parsed.Slug = strings.TrimSuffix(base, ext)

		notes = append(notes, parsed)
	}

	if len(notes) == 0 {
		return []note.Note{}, errors.New("no notes were found")
	}

	return notes, nil
}

// NOTE: ophaned tags can happend and are acceptable.
func upsertNotesIntoDB(ctx context.Context, writeDB *sql.DB, notes []note.Note) error {
	tx, err := writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("creating transaction for notes: %w", err)
	}
	defer tx.Rollback()

	if err := prepareSchema(ctx, tx); err != nil {
		return err
	}

	if err := upsertNotes(ctx, tx, notes); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit notes: %w", err)
	}

	return nil
}

type noteScanner interface {
	Scan(dest ...any) error
}

func scanNote(scan noteScanner) (note.Note, error) {
	var (
		n            note.Note
		tagsJSON     string
		headingsJSON string
	)

	if err := scan.Scan(
		&n.Id,
		&n.Slug,
		&n.Title,
		&n.Date,
		&n.Short,
		&n.BodyHTML,
		&headingsJSON,
		&tagsJSON,
		&n.UpdatedAt,
	); err != nil {
		return n, fmt.Errorf("scan note: %w", err)
	}

	if err := json.Unmarshal([]byte(headingsJSON), &n.Headings); err != nil {
		return n, fmt.Errorf("decode headings for %q: %w", n.Slug, err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
		return n, fmt.Errorf("decode tags for %q: %w", n.Slug, err)
	}

	return n, nil
}

type noteQueryOptions struct {
	tag   string
	limit *int
}

func queryNotes(ctx context.Context, readDB *sql.DB, opts noteQueryOptions) ([]note.Note, error) {
	query := `
	SELECT
		n.id           ,
		n.slug         ,
		n.title        ,
		n.date         ,
		n.short        ,
		'' AS body_html,
		n.headings     ,
		COALESCE(
			(
				SELECT json_group_array(tag)
				FROM (
					SELECT t.tag
					FROM notes_tags AS nt
					JOIN tags AS t ON t.id = nt.tag_id
					WHERE nt.note_id = n.id
					ORDER BY t.tag
				)
			),
			'[]'
		) AS tags      ,
		n.updated_at
	FROM notes AS n
	`
	args := []any{}

	if opts.tag != "" {
		query += `
			WHERE EXISTS (
				SELECT 1
				FROM notes_tags AS nt
				JOIN tags AS t ON t.id = nt.tag_id
				WHERE nt.note_id = n.id
					AND t.tag = ?
			)
		`
		args = append(args, opts.tag)
	}

	query += ` ORDER BY n.date DESC, n.id DESC`

	if opts.limit != nil {
		query += " LIMIT ?"
		args = append(args, *opts.limit)
	}

	rows, err := readDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()

	notes := make([]note.Note, 0)

	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
}

func prepareSchema(ctx context.Context, tx *sql.Tx) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS notes (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		slug          TEXT    NOT NULL                 ,
		title         TEXT    NOT NULL                 ,
		date          DATE    NOT NULL                 ,
		short         TEXT    NOT NULL                 ,
		headings      TEXT    NOT NULL DEFAULT '[]'
			CHECK (json_valid(headings))               ,
		body_html     TEXT    NOT NULL                 ,
		updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

		UNIQUE (slug)
	);

	CREATE TABLE IF NOT EXISTS tags (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		tag        TEXT    COLLATE NOCASE NOT NULL  ,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (tag)
	);

	CREATE TABLE IF NOT EXISTS notes_tags (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		note_id    INTEGER NOT NULL                 ,
		tag_id     INTEGER NOT NULL                 ,

		UNIQUE (note_id, tag_id)                    ,
		FOREIGN KEY (note_id)
			REFERENCES notes (id)
			ON DELETE CASCADE                       ,
		FOREIGN KEY (tag_id)
			REFERENCES tags (id)
			ON DELETE CASCADE
	);
	`

	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("creating schema for notes: %w", err)
	}

	return nil
}

func upsertNotes(ctx context.Context, tx *sql.Tx, notes []note.Note) error {
	const noteQuery = `
	INSERT INTO notes (
		slug    ,
		title   ,
		date    ,
		short   ,
		headings,
		body_html
	)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(slug) DO UPDATE SET
		title         = excluded.title        ,
		date          = excluded.date         ,
		short         = excluded.short        ,
		headings      = excluded.headings     ,
		body_html     = excluded.body_html    ,
		updated_at    = CURRENT_TIMESTAMP
	WHERE
		   notes.title     IS NOT excluded.title
		OR notes.date      IS NOT excluded.date
		OR notes.short     IS NOT excluded.short
		OR notes.headings  IS NOT excluded.headings
		OR notes.body_html IS NOT excluded.body_html
	RETURNING id;
	`

	for _, n := range notes {
		if strings.TrimSpace(n.Slug) == "" {
			return fmt.Errorf("cannot publish note %q: slug is empty", n.Title)
		}

		headingsJSON, err := json.Marshal(n.Headings)
		if err != nil {
			return fmt.Errorf("encode headings for %q: %w", n.Slug, err)
		}

		/*
		** Insert or update this note
		**/
		var noteID int64
		var noteUpdated bool
		err = tx.QueryRowContext(
			ctx,
			noteQuery,
			n.Slug,
			n.Title,
			n.Date.Format(time.DateOnly),
			n.Short,
			string(headingsJSON),
			string(n.BodyHTML),
		).Scan(&noteID)
		switch {
		case err == nil:
			noteUpdated = true
		case errors.Is(err, sql.ErrNoRows):
			noteUpdated = false

			if err := tx.QueryRowContext(
				ctx,
				`SELECT id FROM notes WHERE slug = ?`,
				n.Slug,
			).Scan(&noteID); err != nil {
				return fmt.Errorf("query unchanged note %q: %w", n.Slug, err)
			}
		default:
			return fmt.Errorf("upsert note %q: %w", n.Slug, err)
		}

		/*
		** Collect note's "old" tags
		**/
		oldTagIDs := make(map[int64]struct{})
		rows, err := tx.QueryContext(
			ctx,
			`SELECT tag_id FROM notes_tags WHERE note_id = ?`,
			noteID,
		)
		if err != nil {
			return fmt.Errorf("query existing tags for note %q: %w", n.Slug, err)
		}
		for rows.Next() {
			var tagID int64
			if err := rows.Scan(&tagID); err != nil {
				rows.Close()
				return fmt.Errorf("scan existing tag for note %q: %w", n.Slug, err)
			}
			oldTagIDs[tagID] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate existing tags for note %q: %w", n.Slug, err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close existing tags for note %q: %w", n.Slug, err)
		}

		/*
		** Find or Insert note's "new" tags
		**/
		newTagIDs := make(map[int64]struct{}, len(n.Tags))
		for _, tag := range n.Tags {
			if strings.TrimSpace(tag) == "" {
				return fmt.Errorf("note %q contains an empty tag", n.Slug)
			}

			const tagQuery = `
			INSERT INTO tags (tag)
			VALUES (?)
			ON CONFLICT(tag) DO NOTHING
			RETURNING id;
			`

			var tagID int64
			err := tx.QueryRowContext(ctx, tagQuery, tag).Scan(&tagID)
			if errors.Is(err, sql.ErrNoRows) {
				err = tx.QueryRowContext(
					ctx,
					`SELECT id FROM tags WHERE tag = ?`,
					tag,
				).Scan(&tagID)
			}
			if err != nil {
				return fmt.Errorf(
					"query existing tag %q for note %q: %w",
					tag, n.Slug, err,
				)
			}
			newTagIDs[tagID] = struct{}{}
		}

		/*
		** Collect affected tag set, detect tag membershift changes
		**/
		affectedTagIDs := make(map[int64]struct{})
		membershipChanged := false

		// compare oldTagIDS -> newTagIDS
		// operation: remove note assoc. for old tags not in the new tags set
		for tagID := range oldTagIDs {
			if _, remains := newTagIDs[tagID]; remains {
				continue
			}

			if _, err := tx.ExecContext(
				ctx,
				`DELETE FROM notes_tags WHERE note_id = ? AND tag_id = ?`,
				noteID,
				tagID,
			); err != nil {
				return fmt.Errorf("remove tag association from note %q: %w", n.Slug, err)
			}
			affectedTagIDs[tagID] = struct{}{}
			membershipChanged = true
		}

		// compare newTagIDS -> oldTagIDS
		// operation: add note assoc. for new tags not in the old tags set
		for tagID := range newTagIDs {
			if _, exists := oldTagIDs[tagID]; exists {
				continue
			}

			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO notes_tags (note_id, tag_id) VALUES (?, ?)`,
				noteID,
				tagID,
			); err != nil {
				return fmt.Errorf("associate tag with note %q: %w", n.Slug, err)
			}
			affectedTagIDs[tagID] = struct{}{}
			membershipChanged = true
		}

		// if the note changed or tag membership detected a change then,
		// `touch` each of its "new" tags
		if noteUpdated || membershipChanged {
			for tagID := range newTagIDs {
				affectedTagIDs[tagID] = struct{}{}
			}
		}

		// if the note was not changed but tag membership was, we need to
		// explicitly `touch` the updated time for the note.
		if !noteUpdated && membershipChanged {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE notes
					 SET updated_at = CURRENT_TIMESTAMP
				 WHERE id = ?`,
				noteID,
			); err != nil {
				return fmt.Errorf("update timestamp for retagged note %q: %w", n.Slug, err)
			}
		}

		// lastly, `touch` all the affected tags set
		for tagID := range affectedTagIDs {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE tags
					 SET updated_at = CURRENT_TIMESTAMP
				 WHERE id = ?`,
				tagID,
			); err != nil {
				return fmt.Errorf("update timestamp for tag %d: %w", tagID, err)
			}
		}
	} // end for each note loop

	return nil
}
