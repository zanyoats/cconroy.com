package note_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/zanyoats/cconroy.com/internal/note"
)

func TestParseNoteValid(t *testing.T) {
	filePath := filepath.Join("testdata", "note.md")
	n, err := note.Parse(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if len(n.Headings) != 3 {
		t.Error(err)
	}
}

func TestParseNoteInvalid(t *testing.T) {
	filePath := filepath.Join("testdata", "note_empty.md")
	_, err := note.Parse(filePath)

	if err == nil {
		t.Fatal("expected parsing to return an error")
	}

	expectedErrs := []error{
		note.ErrDateRequired,
		note.ErrTitleRequired,
		note.ErrShortRequired,
		note.ErrTagsRequired,
	}

	for _, expectedErr := range expectedErrs {
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected %q for %q", expectedErr, filePath)
		}
	}
}
