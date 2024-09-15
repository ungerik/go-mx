package web

import (
	"bytes"
	"context"
	"errors"
	"iter"
	"path/filepath"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-mx"
)

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

type GlobPageSource struct {
	Pattern  string
	PageType string
}

func (s *GlobPageSource) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return func(yield func(*Page, error) bool) {
		files, err := filepath.Glob(s.Pattern)
		if err != nil {
			yield(nil, err)
			return
		}
		for _, file := range files {
			content := fs.File(file)
			if content.IsDir() {
				// TODO
			}
			switch content.ExtLower() {
			case ".md":
				data, err := content.ReadAll()
				if err != nil {
					yield(nil, err)
					return
				}
				var matter struct {
					Title       string    `yaml:"title"       toml:"title"       json:"title"`
					Date        time.Time `yaml:"date"        toml:"date"        json:"date"`
					PublishDate time.Time `yaml:"publishDate" toml:"publishDate" json:"publishDate"`
					Draft       bool      `yaml:"draft"       toml:"draft"       json:"draft"`
					Tags        []string  `yaml:"tags"        toml:"tags"        json:"tags"`
				}
				_, err = frontmatter.Parse(bytes.NewReader(data), &matter)
				if errors.Is(err, frontmatter.ErrNotFound) {
				}
				if err != nil {
					yield(nil, err)
					return
				}
				page := &Page{
					ContentType: mx.ContentTypeMarkdown,
					Type:        s.PageType,
					Title:       matter.Title,
					Created:     matter.Date,
					LastUpdated: matter.Date,
					Published:   matter.PublishDate,
					Tags:        matter.Tags,
				}
				if matter.Draft {
					page.Published = time.Time{}
				}
				if withContent {
					page.Content = data
				}
				if !yield(page, nil) {
					return
				}

			case ".html", ".htm":
				page := &Page{
					ContentType: mx.ContentTypeHTML,
					Type:        s.PageType,
				}
				// TODO parse HTML title
				if withContent {
					page.Content, err = content.ReadAll()
					if err != nil {
						yield(nil, err)
						return
					}
				}
				if !yield(page, nil) {
					return
				}

			case ".txt":
				page := &Page{
					ContentType: mx.ContentTypePlainText,
					Type:        s.PageType,
					Title:       content.Name(), // Use filename as title
				}
				if withContent {
					page.Content, err = content.ReadAll()
					if err != nil {
						yield(nil, err)
						return
					}
				}
				if !yield(page, nil) {
					return
				}
			}
		}
	}
}
