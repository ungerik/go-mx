package shadcn

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/internal/headless"
)

// render renders a component with a double-quote CheckedWriter and fails the
// test on any render error (which includes the duplicate-attribute check).
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

// runInChrome serves a page holding body at "/" of a test server, together
// with routes, loads it in headless Chrome and returns what script recorded.
// script is the body of an async function run after the page loaded: it
// records its findings as properties of out, may await wait(ms), and
// anything it throws is recorded as out.error. Without Chrome the test is
// skipped (see [headless.Chrome]).
func runInChrome(t *testing.T, body mx.Component, script string, routes map[string]http.HandlerFunc) map[string]any {
	t.Helper()
	chrome := headless.Chrome(t)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page := mx.Components{
			mx.Raw(`<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>`),
			body,
			mx.Raw(`<pre id="result"></pre><script>var wait=function(ms){return new Promise(function(r){setTimeout(r,ms);});};` +
				`window.addEventListener('load',function(){(async function(){var out={};try{` + script +
				`}catch(e){out.error=String(e&&e.stack||e);}document.getElementById('result').textContent=JSON.stringify(out);})();});</script></body></html>`),
		}
		if err := page.Render(r.Context(), mx.NewCheckedWriter(w)); err != nil {
			t.Errorf("render page: %v", err)
		}
	})
	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	dom, err := headless.DumpDOM(ctx, chrome, srv.URL+"/", 10*time.Second)
	if err != nil {
		t.Fatalf("headless Chrome: %v", err)
	}
	result, ok := headless.Result(dom)
	if !ok || result == "" {
		t.Fatalf("the page recorded no result:\n%s", dom)
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("result %q: %v", result, err)
	}
	if e, ok := out["error"]; ok {
		t.Fatalf("page script failed: %v", e)
	}
	return out
}

// checkRecorded compares what [runInChrome] recorded with want, as JSON
// values: strings, numbers (float64) and booleans.
func checkRecorded(t *testing.T, got, want map[string]any) {
	t.Helper()
	for key, w := range want {
		if g, ok := got[key]; !ok {
			t.Errorf("%s: not recorded", key)
		} else if g != w {
			t.Errorf("%s = %#v, want %#v", key, g, w)
		}
	}
}

func TestFinish(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		out := render(t, finish(html.Div(), "thing", "p-2 text-sm"))
		if !strings.Contains(out, `data-slot="thing"`) {
			t.Errorf("missing data-slot: %s", out)
		}
		if !strings.Contains(out, `class="p-2 text-sm"`) {
			t.Errorf("missing merged class: %s", out)
		}
	})

	t.Run("caller class overrides base", func(t *testing.T) {
		out := render(t, finish(html.Div(html.Class("p-8")), "thing", "p-2 text-sm"))
		if strings.Contains(out, "p-2") {
			t.Errorf("conflicting base class p-2 should be dropped: %s", out)
		}
		if !strings.Contains(out, "p-8") || !strings.Contains(out, "text-sm") {
			t.Errorf("expected p-8 and text-sm: %s", out)
		}
	})

	t.Run("two caller class attribs merge into one", func(t *testing.T) {
		out := render(t, finish(html.Div(html.Class("p-2"), html.Class("m-4")), "thing", "text-sm"))
		if n := strings.Count(out, "class="); n != 1 {
			t.Errorf("expected exactly one class attribute, got %d: %s", n, out)
		}
		for _, c := range []string{"text-sm", "p-2", "m-4"} {
			if !strings.Contains(out, c) {
				t.Errorf("missing %s: %s", c, out)
			}
		}
	})

	t.Run("caller data-slot is dropped", func(t *testing.T) {
		out := render(t, finish(html.Div(html.DataAttr("slot", "evil")), "thing", ""))
		if !strings.Contains(out, `data-slot="thing"`) || strings.Contains(out, "evil") {
			t.Errorf("caller data-slot must not override component slot: %s", out)
		}
	})

	t.Run("duplicate attribute last wins, no render error", func(t *testing.T) {
		out := render(t, finish(html.Div(html.ID("a"), html.ID("b")), "thing", ""))
		if !strings.Contains(out, `id="b"`) || strings.Contains(out, `id="a"`) {
			t.Errorf("last id should win: %s", out)
		}
	})

	t.Run("attribute order: data-slot first, class last", func(t *testing.T) {
		out := render(t, finish(html.Div(html.ID("x")), "thing", "p-2"))
		if !strings.HasPrefix(out, `<div data-slot="thing" id="x" class="`) {
			t.Errorf("unexpected attribute order: %s", out)
		}
	})
}
