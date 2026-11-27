package main

import (
	"errors"
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// linkAttr matches an href or src attribute, its submatch the value. Every
// writer of the site's HTML (go-mx's CheckedWriter, goldmark) quotes
// attribute values with double quotes, and code blocks escape the quotes of
// their text, so markup shown as code is not matched.
var linkAttr = regexp.MustCompile(`\s(?:href|src)="([^"]*)"`)

// idAttr matches an id or name attribute, the targets of a fragment.
var idAttr = regexp.MustCompile(`\s(?:id|name)="([^"]*)"`)

// checkLinks checks every link of every HTML file of the site in dir that
// stays inside the site: its file has to exist below dir, a link to a
// directory has to name its index.html (opened from disk, a directory link
// shows a file listing), a fragment has to name an id or name in the target
// file, and a root-absolute path ("/x") is refused because the site is
// published below a base path of a host it does not own. Links with a scheme or host, data: and javascript: URLs and
// query-only links are not checked. It returns every broken link at once.
func checkLinks(dir string) error {
	ids := make(map[string]map[string]bool)
	idsOf := func(file string) (map[string]bool, error) {
		if set, ok := ids[file]; ok {
			return set, nil
		}
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		set := make(map[string]bool)
		for _, m := range idAttr.FindAllStringSubmatch(string(data), -1) {
			set[html.UnescapeString(m[1])] = true
		}
		ids[file] = set
		return set, nil
	}

	var broken []error
	err := filepath.WalkDir(dir, func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(file) != ".html" {
			return err
		}
		rel, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}
		from := filepath.ToSlash(rel)
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		for _, m := range linkAttr.FindAllStringSubmatch(string(data), -1) {
			ref := html.UnescapeString(m[1])
			if err := checkLink(dir, from, ref, idsOf); err != nil {
				broken = append(broken, fmt.Errorf("%s: link %q: %w", from, ref, err))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return errors.Join(broken...)
}

// checkLink checks one link ref of the site file from, see checkLinks.
func checkLink(dir, from, ref string, idsOf func(string) (map[string]bool, error)) error {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}
	if u.Scheme != "" || u.Host != "" || (u.Path == "" && u.Fragment == "") {
		return nil
	}
	file := from
	if u.Path != "" {
		// The browser resolves a root-absolute path against the host, which
		// the site shares below its base path.
		if strings.HasPrefix(u.Path, "/") {
			return errors.New("root-absolute path leaves the site")
		}
		file = path.Join(path.Dir(from), u.Path)
		if file == ".." || strings.HasPrefix(file, "../") {
			return errors.New("leaves the site")
		}
		if strings.HasSuffix(u.Path, "/") {
			return errors.New("links a directory instead of its index.html")
		}
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(file)))
		if err != nil {
			return errors.New("no such file")
		}
		if info.IsDir() {
			return errors.New("links a directory instead of its index.html")
		}
	}
	if u.Fragment == "" || path.Ext(file) != ".html" {
		return nil
	}
	set, err := idsOf(file)
	if err != nil {
		return err
	}
	if !set[u.Fragment] {
		return fmt.Errorf("no id %q in %s", u.Fragment, file)
	}
	return nil
}
