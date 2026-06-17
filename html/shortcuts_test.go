package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortcutConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"meta charset",
			MetaCharset("UTF-8").String(),
			`<meta charset='UTF-8'/>`,
		},
		{
			"meta charset UTF-8 constant",
			string(MetaCharsetUTF8),
			`<meta charset="UTF-8">`,
		},
		{
			"meta name",
			MetaName("description", "A go-mx page").String(),
			`<meta name='description' content='A go-mx page'/>`,
		},
		{
			"meta property (open graph)",
			MetaProperty("og:title", "Hello").String(),
			`<meta property='og:title' content='Hello'/>`,
		},
		{
			"meta viewport",
			MetaViewport("width=device-width, initial-scale=1").String(),
			`<meta name='viewport' content='width=device-width, initial-scale=1'/>`,
		},
		{
			"script src",
			ScriptSrc("/app.js").String(),
			`<script src='/app.js'></script>`,
		},
		{
			"script src with extra attrib",
			ScriptSrc("/app.js", Defer).String(),
			`<script src='/app.js' defer='defer'></script>`,
		},
		{
			"script js (inline, raw)",
			ScriptJS(`console.log(1 < 2)`).String(),
			`<script>console.log(1 < 2)</script>`,
		},
		{
			"script js (multiple args joined by semicolon)",
			ScriptJS(`const x = 1`, `console.log(x)`).String(),
			`<script>const x = 1;console.log(x)</script>`,
		},
		{
			"script module with src",
			ScriptModule(Src("/app.mjs")).String(),
			`<script type='module' src='/app.mjs'></script>`,
		},
		{
			"script module js (inline, raw)",
			ScriptModuleJS(`import {x} from "./m.js"; x()`).String(),
			`<script type='module'>import {x} from "./m.js"; x()</script>`,
		},
		{
			"script module js (multiple args joined by semicolon)",
			ScriptModuleJS(`import {x} from "./m.js"`, `x()`).String(),
			`<script type='module'>import {x} from "./m.js";x()</script>`,
		},
		{
			"stylesheet",
			StyleSheet("/style.css").String(),
			`<link rel='stylesheet' href='/style.css'/>`,
		},
		{
			"icon",
			Icon("/favicon.ico").String(),
			`<link rel='icon' href='/favicon.ico'/>`,
		},
		{
			"link preload with As enum",
			LinkPreload("/font.woff2", AsFont, CrossOrigin("anonymous")).String(),
			`<link rel='preload' href='/font.woff2' as='font' crossorigin='anonymous'/>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

// TestScriptJSRawUnescaped pins the trusted-input contract of ScriptJS and
// ScriptModuleJS documented on those functions: their content is emitted
// verbatim via mx.Raw with no HTML escaping, because a script body cannot be
// HTML-escaped without changing what it executes. The cases below lock that
// no-escaping behavior — in particular that a "</script>" sequence passes
// through unaltered (the XSS vector the doc warns against) — so an accidental
// "add escaping" change is caught instead of silently breaking valid scripts.
func TestScriptJSRawUnescaped(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"ScriptJS emits < > & and quotes verbatim",
			ScriptJS(`if (a < b && c > d) alert("x" + 'y')`).String(),
			`<script>if (a < b && c > d) alert("x" + 'y')</script>`,
		},
		{
			"ScriptJS does not escape a </script> breakout sequence",
			ScriptJS(`var s = "</script>"`).String(),
			`<script>var s = "</script>"</script>`,
		},
		{
			"empty ScriptJS renders an empty, non-void element",
			ScriptJS().String(),
			`<script></script>`,
		},
		{
			"ScriptModuleJS emits < > & verbatim",
			ScriptModuleJS(`export const ok = 1 < 2 && 3 > 2`).String(),
			`<script type='module'>export const ok = 1 < 2 && 3 > 2</script>`,
		},
		{
			"ScriptModuleJS does not escape a </script> breakout sequence",
			ScriptModuleJS(`const s = "</script>"`).String(),
			`<script type='module'>const s = "</script>"</script>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}
