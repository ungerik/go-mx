package shadcntheme

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/internal/headless"
	"github.com/ungerik/go-mx/shadcn"
)

// TestModeToggleInChrome runs InitScript and the ModeToggle menu next to the
// shadcn Sidebar's script in headless Chrome: inline scripts share the global
// scope, and a top-level variable of one (the Sidebar's cookie match) once
// overwrote the other's media query, so every choice threw. It checks each
// choice of the menu and that neither script leaves a variable behind.
// Without Chrome the test is skipped (see headless.Chrome).
func TestModeToggleInChrome(t *testing.T) {
	chrome := headless.Chrome(t)
	page := mx.Components{
		mx.Raw(`<!DOCTYPE html><html><head><meta charset="utf-8">`),
		html.ScriptJS(InitScript),
		mx.Raw(`</head><body>`),
		shadcn.SidebarProvider(shadcn.Sidebar(ModeToggle("theme-menu")), shadcn.SidebarInset("content")),
		mx.Raw(`<pre id="result"></pre><script>window.addEventListener('load',function(){var out={};try{` +
			`var root=document.documentElement,items=document.querySelectorAll('#theme-menu [role=menuitem]');` +
			`out.leaked=['m','w','p'].filter(function(n){return n in window;}).join(',');` +
			`items[1].click();out.dark=root.classList.contains('dark');out.savedDark=localStorage.theme;` +
			`items[0].click();out.light=!root.classList.contains('dark');out.savedLight=localStorage.theme;` +
			`items[2].click();out.system=root.classList.contains('dark')===matchMedia('(prefers-color-scheme: dark)').matches;out.savedSystem=localStorage.theme===undefined;` +
			`}catch(e){out.error=String(e&&e.stack||e);}document.getElementById('result').textContent=JSON.stringify(out);});</script></body></html>`),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Render(r.Context(), mx.NewCheckedWriter(w)); err != nil {
			t.Errorf("render page: %v", err)
		}
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	dom, err := headless.DumpDOM(ctx, chrome, srv.URL+"/", 5*time.Second)
	if err != nil {
		t.Fatalf("headless Chrome: %v", err)
	}
	result, ok := headless.Result(dom)
	if !ok || result == "" {
		t.Fatalf("the page recorded no result:\n%s", dom)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("result %q: %v", result, err)
	}
	want := map[string]any{
		"leaked": "", "dark": true, "savedDark": "dark", "light": true,
		"savedLight": "light", "system": true, "savedSystem": true,
	}
	if e, ok := out["error"]; ok {
		t.Fatalf("page script failed: %v", e)
	}
	for key, w := range want {
		if out[key] != w {
			t.Errorf("%s = %#v, want %#v", key, out[key], w)
		}
	}
}
