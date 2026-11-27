package snapshot

import (
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterFlags(t *testing.T) {
	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	o := RegisterFlags(fs, "cmd/example-x")
	if o.Enabled() {
		t.Fatal("enabled before -snapshot was given")
	}
	if err := fs.Parse([]string{"-snapshot", "/tmp/x", "-snapshot-index", "../index.html"}); err != nil {
		t.Fatal(err)
	}
	if !o.Enabled() || o.Dir != "/tmp/x" || o.IndexHref != "../index.html" || o.Command != "cmd/example-x" {
		t.Fatalf("options = %+v", o)
	}
}

func TestNote(t *testing.T) {
	o := &Options{Command: "cmd/example-x", IndexHref: "../index.html"}
	note, err := o.Note()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Static snapshot.",
		"go run ./cmd/example-x",
		`href="` + SourceURL + `cmd/example-x"`,
		`href="../index.html"`,
		`role="note"`,
	} {
		if !strings.Contains(note, want) {
			t.Errorf("note misses %q:\n%s", want, note)
		}
	}

	// Without IndexHref the note has no link to the examples page.
	o.IndexHref = ""
	note, err = o.Note()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(note, "All examples") {
		t.Errorf("note links all examples without IndexHref:\n%s", note)
	}

	if _, err := (&Options{}).Note(); err == nil {
		t.Error("a note without Command did not fail")
	}
}

func TestWrite(t *testing.T) {
	o := &Options{Dir: t.TempDir(), Command: "cmd/example-x"}
	if err := o.Write("sub/page.html",
		/*html*/ `<html><body class="x"><p>hi</p></body></html>`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(o.Dir, "sub", "page.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	wantPrefix := /*html*/ `<html><body class="x"><div role="note"`
	if !strings.HasPrefix(page, wantPrefix) || !strings.HasSuffix(page, "<p>hi</p></body></html>") {
		t.Errorf("note not right after <body>:\n%s", page)
	}

	if err := o.Write("fragment.html", "<p>no body</p>"); err == nil {
		t.Error("a page without <body> did not fail")
	}
}

func TestEdit(t *testing.T) {
	got, err := Edit( /*html*/ `<a href="/x">1</a><a href="/x">2</a>`, `href="/x"`, `href="x.html"`, 2)
	if err != nil {
		t.Fatal(err)
	}
	if want := /*html*/ `<a href="x.html">1</a><a href="x.html">2</a>`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if _, err := Edit("abc", "x", "y", 1); err == nil {
		t.Error("a missing target did not fail")
	}
	if _, err := Edit("xx", "x", "y", 1); err == nil {
		t.Error("a target found too often did not fail")
	}
}

func TestInsertAfterStartTags(t *testing.T) {
	page := /*html*/ `<div data-a="">x</div><div data-a="">y</div>`
	got, err := InsertAfterStartTags(page, `data-a=""`, "<b>!</b>", 2)
	if err != nil {
		t.Fatal(err)
	}
	if want := /*html*/ `<div data-a=""><b>!</b>x</div><div data-a=""><b>!</b>y</div>`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if _, err := InsertAfterStartTags(page, `data-a=""`, "", 1); err == nil {
		t.Error("a wrong count did not fail")
	}
	if _, err := InsertAfterStartTags(`<div data-a=""`, `data-a=""`, "", 1); err == nil {
		t.Error("a tag without end did not fail")
	}
}

func TestGetAndPost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/missing":
			http.NotFound(w, r)
		case r.Method == http.MethodPost:
			_ = r.ParseForm()
			_, _ = io.WriteString(w, "posted "+r.PostForm.Get("name"))
		default:
			_, _ = io.WriteString(w, "got "+r.URL.Query().Get("q"))
		}
	})
	got, err := Get(handler, "/?q=1")
	if err != nil || got != "got 1" {
		t.Errorf("Get = %q, %v", got, err)
	}
	got, err = Post(handler, "/", url.Values{"name": {"ada"}})
	if err != nil || got != "posted ada" {
		t.Errorf("Post = %q, %v", got, err)
	}
	if _, err := Get(handler, "/missing"); err == nil {
		t.Error("a 404 did not fail")
	}
}
