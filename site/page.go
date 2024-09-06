package site

import "time"

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

type Page interface {
	Title() string
	PublishedTime() time.Time
}
