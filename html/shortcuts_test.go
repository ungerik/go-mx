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
