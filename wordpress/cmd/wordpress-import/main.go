// Command wordpress-import converts a WordPress WXR export (Tools → Export) into
// a static HTML site rendered with go-mx + shadcn, plus an import-diagnostics
// report.
//
//	wordpress-import -in export.xml -out ./dist          # write the static site
//	wordpress-import -in export.xml -out ./dist -serve :8080   # …and preview it
//	wordpress-import -out ./dist part-1.xml part-2.xml   # merge a split export
//	wordpress-import -in export.xml                       # just print the report
//
// The generated site loads Tailwind v4 from a CDN, so viewing it needs an
// internet connection; the article-body typography is plain CSS and works
// offline. Links are root-absolute, so serve the output from a web root rather
// than opening the files directly.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ungerik/go-mx/wordpress"
)

func main() {
	log.SetFlags(0)
	in := flag.String("in", "", "WXR export file (or pass files as positional arguments)")
	out := flag.String("out", "", "write the static site into this directory")
	serve := flag.String("serve", "", "after writing, serve -out on this address, e.g. :8080")
	base := flag.String("base", "", "URL sub-path to host the site under, e.g. /blog")
	title := flag.String("title", "", "override the site title")
	permalinks := flag.String("permalinks", "slug", "permalink style: slug | dated | id")
	status := flag.String("status", "publish", "comma-separated post statuses to include")
	flag.Parse()

	files := flag.Args()
	if *in != "" {
		files = append([]string{*in}, files...)
	}
	if len(files) == 0 {
		log.Fatal("no input: pass a WXR file with -in or as an argument (export one from WordPress → Tools → Export)")
	}

	site, parseRep, err := wordpress.ParseFiles(files...)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	opt := wordpress.Options{
		SiteTitle:  *title,
		Permalinks: wordpress.PermalinkStyle(*permalinks),
		BasePath:   *base,
		Statuses:   parseStatuses(*status),
	}

	if *out == "" {
		fmt.Print(parseRep.Summary())
		fmt.Println("\n(no -out given, so nothing was written)")
		return
	}

	rep, err := wordpress.WriteStatic(site, *out, opt)
	if err != nil {
		log.Fatalf("write failed: %v", err)
	}
	rep.InheritParse(parseRep)
	fmt.Print(rep.Summary())
	fmt.Printf("\nWrote %d posts and %d pages to %s\n", rep.Counts.Posts, rep.Counts.Pages, *out)
	fmt.Printf("Report: %s/import-report/ (HTML) and %s/import-report.json\n", *out, *out)

	if *serve == "" {
		fmt.Printf("View it: cd %s && python3 -m http.server 8000  →  http://localhost:8000\n", *out)
		return
	}
	fmt.Printf("Serving %s on http://localhost%s (Ctrl-C to stop)\n", *out, *serve)
	log.Fatal(http.ListenAndServe(*serve, http.FileServer(http.Dir(*out))))
}

func parseStatuses(csv string) []wordpress.Status {
	var out []wordpress.Status
	for _, s := range strings.Split(csv, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, wordpress.Status(s))
		}
	}
	return out
}
