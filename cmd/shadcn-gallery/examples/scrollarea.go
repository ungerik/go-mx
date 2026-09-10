package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ScrollAreaDemo renders a fixed-height scroll area listing version tags.
func ScrollAreaDemo() mx.Component {
	return shadcn.ScrollArea(html.Class("h-48 w-56 rounded-md border p-4 text-sm"),
		html.DivClass("mb-3 font-medium leading-none", "Tags"),
		mx.ForEach([]string{
			"v1.2.0-beta.50", "v1.2.0-beta.49", "v1.2.0-beta.48", "v1.2.0-beta.47",
			"v1.2.0-beta.46", "v1.2.0-beta.45", "v1.2.0-beta.44", "v1.2.0-beta.43",
			"v1.2.0-beta.42", "v1.2.0-beta.41", "v1.2.0-beta.40", "v1.2.0-beta.39",
		}, func(tag string) mx.Component {
			return html.DivClass("border-b py-1.5 last:border-b-0", tag)
		}),
	)
}

// ScrollAreaStickToBottom renders a scroll area that follows content appended
// after the initial render, the behavior a chat transcript needs.
//
// Two things are visible on a static page: it starts scrolled to the *bottom*
// rather than the top, and the button grows the list from the client. The
// append is what [shadcn.StickToBottom]'s MutationObserver watches for, so it
// demonstrates the same path content takes when it arrives over an
// [mx.SSEResponse] stream — scroll up and the following stops, scroll back down
// and it resumes.
func ScrollAreaStickToBottom() mx.Component {
	// One binding for the element's id and for the selector the script looks it
	// up by, so editing a part cannot break the append silently.
	logKey := []any{"gallery", "stick-to-bottom", "log"}
	logID := mx.KeyedIDValue(logKey...)
	appendLine := /*js*/ `var l=document.getElementById('` + logID + `');` +
		`var d=document.createElement('div');` +
		`d.className='border-b py-1.5 last:border-b-0';` +
		`d.textContent='Rebuilt at '+new Date().toLocaleTimeString();` +
		`l.appendChild(d)`
	return html.DivClass("flex flex-col items-start gap-3",
		shadcn.ScrollArea(
			html.Class("h-48 w-56 rounded-md border p-4 text-sm"),
			shadcn.StickToBottom,
			html.DivClass("mb-3 font-medium leading-none", "Build log"),
			html.Div(mx.KeyedID(logKey...),
				mx.ForEach([]string{
					"Resolving dependencies", "Compiling mx", "Compiling html",
					"Compiling svg", "Compiling hx", "Compiling shadcn",
					"Linking", "Done in 4.2s",
				}, func(line string) mx.Component {
					return html.DivClass("border-b py-1.5 last:border-b-0", line)
				}),
			),
		),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault,
			html.OnClick(appendLine),
			"Append line"),
	)
}
