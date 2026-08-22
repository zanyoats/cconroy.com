package page_test

import (
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/zanyoats/cconroy.com/internal/note"
	"github.com/zanyoats/cconroy.com/internal/page"
)

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	os.Exit(m.Run())
}

func TestNoteRender(t *testing.T) {
	item := &note.Note{
		Date: time.Date(
			2025,
			time.December,
			25,
			0, 0, 0, 0,
			time.UTC,
		),
		UpdatedAt: time.Date(
			2026,
			time.August,
			17,
			12, 0, 0, 0,
			time.UTC,
		),
		Title: "Sunt Lorem culpa ipsum duis.",
		Short: "Adipisicing ut sit commodo laboris magna cillum do in ullamco consequat.",
		Tags:  []string{"Foo", "Bar", "Baz"},
		Headings: []note.Heading{
			{Id: "foo", Text: "Foo", Level: 2},
			{Id: "bar", Text: "Bar", Level: 3},
		},
		BodyHTML: template.HTML("<div>hello, world</div>"),
	}
	model := struct {
		*note.Note
		TOC []*note.TOCItem
	}{
		Note: item,
		TOC:  note.BuildTOC(item.Headings),
	}

	renderer, err := page.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := renderer.Render(&buf, "note.gohtml", model); err != nil {
		t.Fatal(err)
	}

	approvals.VerifyString(t, buf.String())
}
