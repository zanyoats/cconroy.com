package note

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type Note struct {
	Id                 int64
	Slug, Title, Short string
	Date               time.Time
	Tags               []string
	Headings           []Heading
	BodyHTML           template.HTML
	UpdatedAt          time.Time
}

type tomlFrontMatter struct {
	Date  string   `toml:"date"`
	Title string   `toml:"title"`
	Short string   `toml:"short"`
	Tags  []string `toml:"tags"`
}

type Heading struct {
	Id    string `json:"id"`
	Text  string `json:"text"`
	Level int    `json:"level"`
}

type TOCItem struct {
	Heading
	Children []*TOCItem
}

var (
	ErrDateRequired  = errors.New("date is required")
	ErrInvalidDate   = errors.New("date is invalid")
	ErrTitleRequired = errors.New("title is required")
	ErrShortRequired = errors.New("short is required")
	ErrTagsRequired  = errors.New("tags is required")
)

func (m tomlFrontMatter) validate(filePath string) (date time.Time, err error) {
	var errs []error

	if len(m.Date) == 0 {
		errs = append(errs, ErrDateRequired)
	} else if date, err = time.Parse(time.DateOnly, m.Date); err != nil {
		errs = append(errs, fmt.Errorf("%q: %w", m.Date, ErrInvalidDate))
	}

	if len(m.Title) == 0 {
		errs = append(errs, ErrTitleRequired)
	}
	if len(m.Short) == 0 {
		errs = append(errs, ErrShortRequired)
	}
	if len(m.Tags) == 0 {
		errs = append(errs, ErrTagsRequired)
	}

	if len(errs) > 0 {
		err = fmt.Errorf("note %q validation failed:\n%w", filePath, errors.Join(errs...))
	}

	return
}

func Parse(filePath string) (Note, error) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return Note{}, err
	}

	front, body, err := splitFrontMatter(contents)
	if err != nil {
		return Note{}, err
	}

	markdown := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAttribute(),     // Supports: # Title {#custom-id}
			parser.WithAutoHeadingID(), // Generates IDs when omitted
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	document := markdown.Parser().Parse(text.NewReader(body))

	note := Note{}
	if err = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		// Only process each tag once, when "entering"
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		idValue, ok := heading.AttributeString("id")
		if !ok {
			return ast.WalkContinue, nil
		}

		textNode, ok := heading.FirstChild().(*ast.Text)
		if !ok {
			return ast.WalkContinue, nil
		}

		note.Headings = append(note.Headings, Heading{
			Id:    string(idValue.([]byte)),
			Text:  string(textNode.Value(body)),
			Level: heading.Level,
		})

		return ast.WalkContinue, nil
	}); err != nil {
		return Note{}, err
	}

	var htmlBuf bytes.Buffer
	if err = markdown.Renderer().Render(&htmlBuf, body, document); err != nil {
		return Note{}, err
	}
	// this html buffer comes from `data/notes/*` which is a trusted source
	note.BodyHTML = template.HTML(htmlBuf.String())

	var frontMatter tomlFrontMatter
	if _, err := toml.Decode(string(front), &frontMatter); err != nil {
		return Note{}, err
	}
	if note.Date, err = frontMatter.validate(filePath); err != nil {
		return Note{}, err
	}
	note.Title = frontMatter.Title
	note.Short = frontMatter.Short
	note.Tags = frontMatter.Tags

	return note, nil
}

// Only supports LF line delimited files; CRLF is rejected
func splitFrontMatter(data []byte) (metadata []byte, body []byte, err error) {
	openDelim := []byte("+++")
	closeDelim := []byte("\n+++\n")

	after, found := bytes.CutPrefix(data, openDelim)
	if !found {
		err = errors.New("expected to start with front matter delimiter")
		return
	}

	if len(after) == 0 || after[0] != '\n' {
		err = errors.New("expected to start with front matter delimiter")
		return
	}

	endingDelimIndex := bytes.Index(after, closeDelim)
	if endingDelimIndex == -1 {
		err = errors.New("closing front matter delimiter not found")
		return
	}

	metadata = after[:endingDelimIndex]
	body = after[endingDelimIndex+len(closeDelim):]

	if len(bytes.TrimSpace(body)) == 0 {
		err = errors.New("empty body was found")
	}

	return
}

func BuildTOC(headings []Heading) []*TOCItem {
	root := &TOCItem{
		Heading: Heading{Level: 0},
	}

	stack := []*TOCItem{root}

	for _, heading := range headings {
		item := &TOCItem{Heading: heading}

		for len(stack) > 1 &&
			stack[len(stack)-1].Level >= heading.Level {
			stack = stack[:len(stack)-1]
		}

		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, item)
		stack = append(stack, item)
	}

	return root.Children
}
