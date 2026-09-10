package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// generate produces the next one or more log lines. A single call can return
// several because that is how a real log arrives: one handled request emits its
// start and its outcome back to back, and rendering that as one burst is what
// makes the stream look like a log rather than a metronome.
//
// The shapes are weighted so that the common ones dominate and the interesting
// ones — a stack trace, a level the theme does not know, a fatal — show up
// often enough to see but rarely enough to stay interesting.
func generate(rnd *rand.Rand) []string {
	switch n := rnd.IntN(100); {
	case n < 40:
		return request(rnd)
	case n < 55:
		return []string{withAttrs(rnd)}
	case n < 65:
		return []string{withTags(rnd)}
	case n < 78:
		return []string{plainLine(rnd)}
	case n < 86:
		return []string{plainLeveledLine(rnd)}
	case n < 92:
		return []string{unknownLevelLine(rnd)}
	case n < 98:
		return []string{stackTraceLine(rnd)}
	default:
		return []string{fatalLine(rnd)}
	}
}

// request is the common case: a request handled, logged at its start and at its
// outcome, sharing a request id the way correlated lines do.
func request(rnd *rand.Rand) []string {
	var (
		id     = requestID(rnd)
		method = pick(rnd, "GET", "GET", "GET", "POST", "PUT", "DELETE")
		path   = pick(rnd, "/api/invoices", "/api/invoices/42", "/api/documents", "/api/partners",
			"/api/bookings?period=2026-09", "/healthz", "/api/documents/upload")
		status = statusCode(rnd)
		level  = "info"
	)
	switch {
	case status >= 500:
		level = "error"
	case status >= 400:
		level = "warn"
	}
	return []string{
		record(
			str("time", stamp()),
			str("level", "debug"),
			str("msg", "request started"),
			str("method", method),
			str("path", path),
			str("request_id", id),
		),
		record(
			str("time", stamp()),
			str("level", level),
			str("msg", "request finished"),
			str("method", method),
			str("path", path),
			str("request_id", id),
			raw("status", fmt.Sprint(status)),
			raw("duration_ms", fmt.Sprintf("%.1f", 0.5+rnd.Float64()*240)),
			raw("cached", fmt.Sprint(rnd.IntN(4) == 0)),
		),
	}
}

// withAttrs carries a nested object, so the renderer's brace formatting shows.
func withAttrs(rnd *rand.Rand) string {
	return record(
		str("time", stamp()),
		str("level", pick(rnd, "info", "info", "debug", "warn")),
		str("msg", pick(rnd,
			"connection pool resized", "flushed write buffer", "reloaded configuration",
			"acquired advisory lock", "vacuum finished")),
		raw("db", record(
			str("host", pick(rnd, "db1.internal", "db2.internal", "replica.internal")),
			raw("port", "5432"),
			raw("in_use", fmt.Sprint(rnd.IntN(40))),
			raw("idle", fmt.Sprint(rnd.IntN(10))),
		)),
		// A null and a big integer id, so both render paths are exercised: the
		// number must survive as its literal rather than through a float64.
		raw("last_error", "null"),
		raw("tenant_id", fmt.Sprint(int64(7000000000000000000)+rnd.Int64N(1000))),
	)
}

// withTags carries an array.
func withTags(rnd *rand.Rand) string {
	tags := pickN(rnd, 2, 4, "billing", "eu-central", "beta", "retry", "batch", "async", "priority")
	quoted := make([]string, len(tags))
	for i, tag := range tags {
		quoted[i] = encode(tag)
	}
	return record(
		str("time", stamp()),
		str("level", pick(rnd, "info", "debug")),
		str("msg", pick(rnd, "job enqueued", "job finished", "batch scheduled")),
		str("job", pick(rnd, "invoice-export", "ocr-pipeline", "sepa-sync", "reindex")),
		raw("tags", "["+strings.Join(quoted, ",")+"]"),
		raw("attempt", fmt.Sprint(1+rnd.IntN(3))),
	)
}

// plainLine is a line that never went through a structured logger, the way
// output from a library or a subprocess shows up in the same stream.
func plainLine(rnd *rand.Rand) string {
	return pick(rnd,
		"http: TLS handshake error from 10.0.3.14:52133: EOF",
		"listening on :8080",
		"    at internal/pool: 12 idle connections reaped",
		"gc 214 @18.402s 0%: 0.019+2.1+0.005 ms clock",
		"redis: connection reused (pool size 8)",
	)
}

// plainLeveledLine starts with a level word, which the renderer colors the same
// way it colors a JSON level so a mixed stream still reads as one log.
func plainLeveledLine(rnd *rand.Rand) string {
	return pick(rnd,
		"WARN  disk usage on /var/lib/postgresql at 91%",
		"[ERROR] failed to renew certificate: rate limited by ACME server",
		"INFO: cache warm-up finished in 1.4s",
		"debug reconnecting to message broker (attempt 2)",
		"[WARN] clock drift of 340ms detected against ntp peer",
	)
}

// unknownLevelLine uses a level the theme does not define, which renders with
// the undefined level class rather than putting the value into a class name.
func unknownLevelLine(rnd *rand.Rand) string {
	return record(
		str("time", stamp()),
		str("level", pick(rnd, "notice", "verbose", "audit")),
		str("msg", pick(rnd, "permission granted", "configuration snapshot taken", "session opened")),
		str("actor", pick(rnd, "erik@example.com", "svc-importer", "cron")),
	)
}

// stackTraceLine carries a multi-line string value, which is what the renderer
// puts in a <pre> so its indentation survives.
func stackTraceLine(rnd *rand.Rand) string {
	trace := "runtime error: invalid memory address or nil pointer dereference\n" +
		"goroutine 1842 [running]:\n" +
		"\tledger.(*Posting).Apply(0x0, 0xc0004b21e0)\n" +
		"\t\t/srv/app/ledger/posting.go:118 +0x1f\n" +
		"\tledger.(*Batch).Commit(0xc0002a8000)\n" +
		"\t\t/srv/app/ledger/batch.go:64 +0xa5\n" +
		"\thttp.HandlerFunc.ServeHTTP(0xc000123440, {0x8f2e40, 0xc0001a2000})\n" +
		"\t\t/usr/local/go/src/net/http/server.go:2220 +0x29"
	return record(
		str("time", stamp()),
		str("level", "error"),
		str("msg", "recovered from panic while posting batch"),
		str("batch", fmt.Sprintf("B-2026-%04d", rnd.IntN(10000))),
		str("stack", trace),
	)
}

// fatalLine is the rarest shape and the one whose level the example theme
// renders as an icon instead of text.
func fatalLine(rnd *rand.Rand) string {
	return record(
		str("time", stamp()),
		str("level", pick(rnd, "fatal", "panic")),
		str("msg", "lost quorum, shutting down"),
		raw("peers_reachable", fmt.Sprint(rnd.IntN(2))),
		raw("quorum", "2"),
	)
}

// record joins pre-encoded pairs into a JSON object. The pairs are strings
// rather than a map because the field order is part of what is being shown: the
// renderer drops the label of a timestamp only when it leads the record.
func record(pairs ...string) string {
	return "{" + strings.Join(pairs, ",") + "}"
}

// str encodes a key and a string value.
func str(key, val string) string { return encode(key) + ":" + encode(val) }

// raw encodes a key and an already-encoded JSON value.
func raw(key, val string) string { return encode(key) + ":" + val }

// encode JSON-encodes a string, which is what keeps the multi-line stack trace
// and any quote in a message from producing a line that will not parse.
func encode(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

func stamp() string { return time.Now().Format("2006-01-02T15:04:05.000Z07:00") }

func pick(rnd *rand.Rand, options ...string) string {
	return options[rnd.IntN(len(options))]
}

// pickN returns between min and max distinct options.
func pickN(rnd *rand.Rand, min, max int, options ...string) []string {
	shuffled := append([]string(nil), options...)
	rnd.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:min+rnd.IntN(max-min+1)]
}

func requestID(rnd *rand.Rand) string {
	const hex = "0123456789abcdef"
	var b strings.Builder
	for range 16 {
		b.WriteByte(hex[rnd.IntN(len(hex))])
	}
	return b.String()
}

func statusCode(rnd *rand.Rand) int {
	switch n := rnd.IntN(100); {
	case n < 78:
		return pickStatus(rnd, 200, 200, 201, 204)
	case n < 88:
		return pickStatus(rnd, 304, 302)
	case n < 97:
		return pickStatus(rnd, 400, 401, 403, 404, 409, 422)
	default:
		return pickStatus(rnd, 500, 502, 503, 504)
	}
}

func pickStatus(rnd *rand.Rand, codes ...int) int {
	return codes[rnd.IntN(len(codes))]
}
