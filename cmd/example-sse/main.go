// Command example-sse streams a synthetic chat transcript to dogfood the
// [mx.SSEResponse] flush loop and everything built around it.
//
// It exercises the pieces together, which is the only way some of them fail:
// markup containing newlines (the "data:" line split), a message addressed by a
// [mx.KeyedID] that later events append to out of band (the reason KeyedID
// exists next to mx.UniqueID), resumption after a dropped connection via
// [mx.SSEEvent.ID] and [mx.LastEventID], the [hx.SSEConnect] / [hx.SSESwap] /
// [hx.SSEClose] attributes, [shadcn.StickToBottom] following the growing
// transcript, and an in-band failure delivered as [mx.SSEEventError] once the
// 200 status has been committed.
//
//	go run ./cmd/example-sse
//
// Then browse to http://localhost:8080. Reload to replay the stream. Scroll up
// while it runs to watch the transcript stop following, and scroll back to the
// bottom to watch it resume.
//
// Two flags show what happens when a stream does not simply succeed:
//
//   - -fail reports an error in band instead of finishing, because by then the
//     200 status is committed and there is no 500 left to send.
//   - -drop cuts the first connection midway through the reply. The browser
//     reconnects on its own and sends back the last id it saw, and the handler
//     resumes from the next event — so the transcript completes exactly once.
//     If resumption were broken the whole transcript would appear twice, which
//     is what a stream without event ids does on every dropped connection.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
	"github.com/ungerik/go-mx/shadcn"
)

const (
	// eventMessage is the event name the transcript subscribes to. It is an
	// ordinary caller-chosen name, unlike mx.SSEEventError.
	eventMessage = "message"
	// eventDone is the event name wired to hx.SSEClose, so the client stops
	// reconnecting when the stream ends. SSE has no close frame of its own.
	eventDone = "done"
	// replyKey identifies the assistant message that grows token by token. It
	// is a fixed key here because the example streams one reply; a real
	// application would key it by conversation and message id.
	replyKey = "reply"
)

// tickInterval paces the stream so the browser shows it arriving rather than
// appearing at once. It is a var so tests can drain the stream without waiting.
var tickInterval = 300 * time.Millisecond

// transcriptStyle keeps the example to one file: it is the little that cannot
// come from the shadcn classes, which need a Tailwind build this example
// deliberately does not have. Without an explicit height and overflow there is
// nothing to scroll, so scroll anchoring could not be demonstrated at all.
const transcriptStyle = "height:10rem;overflow:auto;border:1px solid #d4d4d8;border-radius:.5rem;padding:.75rem;font:14px/1.5 ui-monospace,monospace"

// page is the initial document. Everything after it arrives over the stream.
func page() mx.Component {
	return mx.Components{
		mx.Raw("<!DOCTYPE html>"),
		html.HTML(html.Lang("en"),
			html.Head(
				html.Meta(html.CharSet("utf-8")),
				html.TitleElem("go-mx SSE example"),
				hx.ScriptFromCDN,
				hx.ScriptSSEFromCDN,
			),
			html.Body(
				html.H1("go-mx SSE example"),
				// One connection, several subscribers: sse-swap resolves to the
				// nearest sse-connect ancestor, so both the transcript and the
				// error line below read from this single stream. Putting
				// sse-connect on each subscriber instead would open one
				// EventSource per element and stream the whole transcript twice.
				html.Div(
					hx.Ext("sse"),
					hx.SSEConnect("/stream"),
					hx.SSEClose(eventDone),
					shadcn.ScrollArea(
						html.Style(transcriptStyle),
						shadcn.StickToBottom,
						// Appending here rather than replacing is what makes
						// this a transcript instead of one swapping message.
						html.Div(hx.SSESwap(eventMessage), hx.Swap(hx.SwapBeforeEnd)),
					),
					// A failure reported in band cannot be a 500 any more, so it
					// needs somewhere to land. mx.SSEEventError is the reserved
					// name mx.SSEResponse.SendError sends.
					html.Div(
						html.Style("color:#b91c1c;margin-top:.5rem"),
						hx.SSESwap(mx.SSEEventError),
					),
				),
			),
		),
	}
}

// message renders one finished transcript line. The multi-line markup is
// deliberate: it makes the stream exercise the "data:" line split, which is
// what CheckedWriter.WithIndent would produce in a real application.
func message(speaker, text string) mx.Component {
	return html.Div(
		html.Style("margin-bottom:.5rem"),
		html.Strong(speaker+": "),
		mx.RawNewline,
		html.Span(text),
		mx.RawNewline,
	)
}

// transcript is the conversation the stream replays. Enough lines to overflow
// the transcript box, so scrolling up during the stream actually demonstrates
// the anchoring.
var transcript = []struct{ speaker, text string }{
	{"user", "How much did we invoice last quarter?"},
	{"assistant", "Looking at the ledger for Q3…"},
	{"assistant", "Three hundred and twelve invoices, all posted."},
	{"assistant", "Net total is 1.42 million euro."},
	{"user", "How does that compare to Q2?"},
	{"assistant", "Q2 closed at 1.19 million euro."},
	{"assistant", "So Q3 is up a little under 20 percent."},
	{"user", "Any of those still unpaid?"},
	{"assistant", "Forty-one are open, worth 213 thousand euro."},
	{"assistant", "Nine of them are past due."},
	{"user", "Break it down by customer."},
	{"assistant", "Sorting by net total, largest first."},
}

// replyTokens are appended one event at a time into a single message, the way a
// model's answer arrives.
var replyTokens = []string{"Acme ", "GmbH ", "leads ", "at ", "41%, ", "then ", "Globex ", "at ", "28%."}

// streamSteps is the whole stream as an indexed list: the transcript, then the
// empty container for the reply, then one event per token.
//
// The list exists so a step can be addressed by its position, which is what
// makes the stream resumable — the index becomes the [mx.SSEEvent.ID], and a
// client that reconnects sends the last one it saw back as
// [mx.LastEventID]. A real application would use durable message ids from its
// store rather than positions in a slice, but the shape is the same: an id the
// handler can answer "give me everything after this" for.
func streamSteps() []mx.Component {
	steps := make([]mx.Component, 0, len(transcript)+1+len(replyTokens))
	for _, line := range transcript {
		steps = append(steps, message(line.speaker, line.text))
	}
	// The container the tokens land in. It is addressed by mx.KeyedID, so the
	// selector below reproduces it exactly — including from a different
	// process, which is what makes resuming into it work at all. mx.UniqueID
	// could not do this: its counter yields a different id every render, and
	// restarts from scratch when the process does.
	steps = append(steps, html.Div(
		html.Style("margin-bottom:.5rem"),
		html.Strong("assistant: "),
		html.Span(mx.KeyedID(replyKey)),
	))
	target := "#" + mx.KeyedIDValue(replyKey)
	for _, token := range replyTokens {
		steps = append(steps, html.Span(hx.SwapOOB(hx.SwapBeforeEnd, target), token))
	}
	return steps
}

// dropIndex is the step -drop cuts the connection after: midway through the
// reply, so the reconnect has to resume into a container the *previous*
// connection rendered.
func dropIndex() int { return len(transcript) + 1 + len(replyTokens)/2 }

// resumeIndex returns the step to continue from, given the id the client
// acknowledged.
//
// An id it does not recognize means "start from the beginning" rather than an
// error: the value is echoed back by the client, so it is client-controlled
// input and a handler must not trust it to address anything. Returning a
// position past the end is fine — the loop then sends nothing and closes.
func resumeIndex(lastEventID string, total int) int {
	i, err := strconv.Atoi(lastEventID)
	if err != nil || i < 0 || i >= total {
		return 0
	}
	return i + 1
}

// streamHandler serves the transcript. Note the shape: everything that could
// fail with a 500 has to happen before NewSSEResponse, because it commits the
// 200 status.
func streamHandler(fail, drop bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read this before creating the response: it is the only thing that
		// distinguishes a reconnect from a fresh reader.
		resumeFrom := mx.LastEventID(r)

		sse, err := mx.NewSSEResponse(w, nil)
		if err != nil {
			// Still a clean 500: NewSSEResponse checks that it can flush before
			// writing anything.
			mx.RespondNonContextError(w, err)
			return
		}
		defer sse.Close()

		// A browser waits about three seconds before reconnecting by default,
		// which makes -drop look like a hang. This is the one connection-level
		// setting, so it goes out once, before any event.
		if err := sse.SetRetry(500 * time.Millisecond); err != nil {
			return
		}

		ctx := r.Context()
		steps := streamSteps()
		for i := resumeIndex(resumeFrom, len(steps)); i < len(steps); i++ {
			if drop && resumeFrom == "" && i > dropIndex() {
				// Cut the connection without closing the stream, the way a
				// dropped network or an impatient proxy would. Returning here
				// sends no close event, so the browser reconnects on its own
				// and arrives back with Last-Event-ID set — which is the whole
				// point of the flag.
				return
			}
			if fail && i == len(transcript) {
				// The -fail path: a 200 and a dozen messages are already on the
				// wire, so this can only travel in band.
				_ = sse.SendError(ctx, fmt.Errorf("ledger query timed out after 30s"))
				break
			}
			// A canceled context means the client navigated away or reloaded,
			// which is the normal way this loop ends. SendEvent reports it
			// without writing.
			event := mx.SSEEvent{Name: eventMessage, ID: strconv.Itoa(i), Comp: steps[i]}
			if err := sse.SendEvent(ctx, event); err != nil {
				return
			}
			time.Sleep(tickInterval)
		}

		// Close on the failure path too. A browser's EventSource reconnects on
		// its own when the server just goes away, so a stream that ends without
		// the event hx.SSEClose binds replays itself — and, on the failure path,
		// re-reports the same error forever.
		_ = sse.Send(ctx, eventDone, nil)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	fail := flag.Bool("fail", false, "report an in-band error mid-stream instead of finishing")
	drop := flag.Bool("drop", false, "cut the first connection mid-reply, so the browser reconnects and resumes")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", streamHandler(*fail, *drop))
	mux.HandleFunc("/", mx.ComponentHTTPHandler(page(), mx.DefaultWriterFactory,
		http.Header{"Content-Type": {mx.ContentTypeHTML}}))

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
