# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Versions use the Go module scheme, `vMAJOR.MINOR.PATCH`, and a release tags
every module of the repo in lockstep: the root module as `vX.Y.Z` and each
submodule with its path as prefix (`fpdf/vX.Y.Z`, `wordpress/vX.Y.Z`).

This file starts with the entry below; the changes made before it are in the
commit history. Nothing is tagged yet, so everything here is unreleased and the
API is still free to change.

## [Unreleased]

### Added

- **`web` package: robots.txt, sitemaps and page metadata for a whole site.**
  A `Site` holds what all pages share — the `BaseURL` every absolute URL is
  built from, the title, the language — and turns its `PageSource`s into the
  files search engines expect:
  - `Site.Sitemap` builds a sitemap of the pages that are meant to be found
    (published, not scheduled, not marked `NoIndex`), sorted by URL so a
    regenerated sitemap only differs where the site did.
  - `Site.RobotsTxt` serves a robots.txt pointing at that sitemap, and
    `Site.RegisterRoutes` registers both on an `http.ServeMux` at the paths the
    robots.txt advertises.
  - `Site.AddPageMetadata` gives a rendered page its canonical link,
    description, author and Open Graph tags, and marks drafts, scheduled pages
    and `NoIndex` pages `noindex, nofollow` so they cannot reach a search index.
- **`web.Robots`** builds a robots.txt from rule groups and sitemap URLs, with
  `AllowAllRobots` and `DisallowAllRobots` for the two cases most sites need.
  It writes through `encoding.TextMarshaler`, `io.WriterTo` or an
  `http.HandlerFunc`, and validates first: a rule that would not mean what it
  says once written — a smuggled line break, a path that is not anchored, a `#`
  that truncates the rule at a comment — is an error instead of a file.
- **`web.Sitemap` and `web.SitemapIndex`** render the sitemaps.org 0.9 protocol
  through the `xml` package. Both are `mx.Component`s, so the same value can be
  written to a file for a statically generated site or served over HTTP, and
  both enforce the protocol's limits (50000 URLs, 50 MiB, 2048-byte `<loc>`,
  the `changefreq` keyword set, the priority range) rather than only documenting
  them.
- **`web.Page`** gains `Path` (the URL a canonical link and a sitemap entry are
  built from), `Description` (the meta description), and `IsPublished`/
  `Indexable`, which put the definition of "should search engines see this" in
  one place.
- **`web.GlobPageSource`** derives `Page.Path` from the file path, so a
  directory of markdown files produces a working sitemap without any extra
  wiring, and reads a `description` from the front matter. Files without front
  matter (HTML, plain text) are dated by their modification time so they can be
  indexable at all.

### Changed

- `web.GlobPageSource.Dir` now scopes the glob: `Pattern` is matched below it
  instead of against the working directory, so the directory is named once
  rather than repeated in both fields where the two could disagree. A `Pattern`
  that is absolute while `Dir` is set is an error, and a directory matching the
  pattern is skipped instead of being read as a page.

### Fixed

- `web.DefaultRenderPage` no longer panics when a page is marked `NoIndex`, and
  now marks drafts and scheduled pages `noindex, nofollow` too — previously
  only an explicit `NoIndex` was honored, so a draft served from a preview
  deployment could be indexed.

[Unreleased]: https://github.com/ungerik/go-mx/commits/main
