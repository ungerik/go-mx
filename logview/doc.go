// Package logview renders streamed log lines as a readable, searchable surface.
//
// Logs are mostly structured JSON today, and a raw JSON line is unreadable at
// stream speed. [Config.Line] parses one and renders it as its fields: the
// timestamp and level promoted, every other field as key=value with the value
// colored by its JSON type, so keys, strings, numbers, booleans and nulls are
// told apart at a glance. A line that is not a JSON object is rendered as plain
// text, with a leading level word colored the same way, so a mixed stream still
// reads as one log.
//
// [Config.View] is the surface those lines stream into: a filter over the lines
// already received (or, with [Config.SearchParam], a search the handler
// answers), a pause button that holds the view without dropping
// anything, a scroll area that follows the stream, and a sink for an in-band
// stream error. Scrolling up pauses as well — reading something on screen and
// holding the stream still are the same intent — and scrolling back to the
// bottom resumes it, while a pause the button caused stays until the button,
// whose caption swaps to [Labels.Resume], takes it back. Either way the state
// is announced through a status region carrying [Labels.Paused], since the
// caption only reaches a screen reader that has the button focused. The filter
// highlights what it matched, as a CSS custom highlight styled through
// [ClassMatch].
//
// # Streaming a log
//
// The server side is [mx.SSEResponse]. The page renders a View wired to the
// stream, and the handler sends one [Config.Line] per event:
//
//	http.Handle("GET /logs", mx.ComponentHTTPHandler(page(), nil, htmlHeader))
//	http.HandleFunc("GET /logs/stream", func(w http.ResponseWriter, r *http.Request) {
//		sse, err := mx.NewSSEResponse(w, nil)
//		if err != nil {
//			mx.RespondNonContextError(w, r, err)
//			return
//		}
//		defer sse.Close()
//		go sse.KeepaliveLoop(r.Context(), 20*time.Second)
//
//		for line := range tail(r.Context(), mx.LastEventID(r)) {
//			event := mx.SSEEvent{Name: logview.DefaultEvent, ID: line.Seq, Comp: logview.Line(line.Text)}
//			if err := sse.SendEvent(r.Context(), event); err != nil {
//				return
//			}
//		}
//	})
//
// and the page:
//
//	html.Head(
//		hx.ScriptFromCDN,
//		hx.ScriptSSEFromCDN,
//		logview.StyleElement(),
//	),
//	html.Body(
//		logview.View(
//			hx.Ext("sse"),
//			hx.SSEConnect("/logs/stream"),
//			hx.SSEClose("done"),
//		),
//	)
//
// A log stream is idle most of the time, so [mx.SSEResponse.KeepaliveLoop] is
// not optional in practice: the shortest idle timeout on the path decides how
// long a quiet connection survives without it. Setting [mx.SSEEvent.ID] to a
// sequence number and resuming from [mx.LastEventID] is what keeps a reconnect
// from appending the backlog a second time.
//
// # Configuration
//
// Everything is a field of [Config], whose zero value is usable and equals
// [Default]: the field names a record is read with ([Config.TimeKey],
// [Config.LevelKey], [Config.MessageKey]), the limits that keep one line or one
// stream from growing without bound, and the [Theme] that colors it. A Config
// is read-only once it is serving.
//
// # Styling
//
// The markup carries only classes, so the same lines can be styled by any
// theme: [Config.StyleElement] emits the stylesheet, and the [Theme] behind it
// gives every value type and every log level a foreground and background color,
// bold, italic, underline and strikethrough — plus, per level, an image to
// render in place of the level text.
//
// A level value comes from whoever wrote the log line, so it is never used as a
// class directly: a value the theme does not define renders as
// [LevelUndefined]. See [Theme.Levels].
//
// The one thing the package does not style is the scroll area, which is
// [shadcn.ScrollArea] and needs Tailwind like the rest of that package.
//
// # Searching
//
// By default the filter hides the lines in the view that do not match, which
// leaves out the lines [Config.MaxLines] already dropped and the backlog behind
// the stream. Setting [Config.SearchParam] makes it a search of the whole log:
// a change of filter reconnects the stream with the filter as that query
// parameter, and the handler answers with the matching lines of its backlog,
// then the matching live ones, testing each with [Config.Matches]:
//
//	query := r.URL.Query().Get(cfg.SearchParam)
//	for line := range tail(r.Context(), mx.LastEventID(r)) {
//		if !cfg.Matches(line.Text, query) {
//			continue
//		}
//		// send it as above
//	}
//
// # Terminal output
//
// ANSI escape sequences — SGR colors and the other CSI, OSC and short ESC
// sequences — are removed before a line is parsed, so a colorizing logger or
// raw terminal output reads as text rather than as "[31m" garbage. The colors
// themselves are not mapped to the theme: severity, the one thing they encode
// in a log, is what the view colors already. [Config.KeepANSI] keeps them as
// text.
//
// A carriage return is a line break by default, so a progress bar that
// redraws itself with "\r" renders as one line per redraw.
// [Config.OverwriteCR] reads it the way a terminal does instead, keeping only
// the text after the last carriage return of each line segment.
//
// # Expandable records
//
// [Config.Expandable] wraps each JSON record in a <details> whose summary is
// the line as it renders without it, and which opens to the whole record
// pretty-printed: one field per line, nested objects and arrays indented,
// colored with the same classes.
package logview
