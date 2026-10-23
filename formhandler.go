package mx

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"sort"
	"strings"
)

// DefaultMaxMemory is the default upper bound on multipart form parsing
// used by [ReflectFormHandler]. 4 MiB matches net/http's
// defaultMaxMemory and is large enough for typical admin forms without
// inviting a memory-pressure attack.
//
// Override globally by reassigning this variable at init time or per
// handler via [ReflectFormConfig.MaxMemory].
var DefaultMaxMemory int64 = 4 << 20

// DefaultMaxUploadSize is the default upper bound on the size of a
// multipart request body that [ReflectFormHandler] reads; past it a POST
// answers 413. Every reflected form accepts a multipart body, also one
// without a file field, and the parser writes file parts past MaxMemory to
// disk, so the bound keeps one request from filling it.
//
// Override globally by reassigning this variable at init time or per
// handler via [ReflectFormConfig.MaxUploadSize].
var DefaultMaxUploadSize int64 = 32 << 20

// FormSubmitLabel is the default text rendered on the form's submit
// button. Reassign to change globally; the renderer layer (html / hx /
// shadcn) may also render its own button instead of relying on this.
var FormSubmitLabel = "Save"

// FormErrorMessage is the form-level message shown when onSubmit returns
// an error that is neither [FieldErrors] nor a [UserMessageError]: such an
// error may carry internal details and a call stack, so its text is shown
// only when [RevealInternalServerErrors] is set (see [FormErrorText]).
// Reassign to change it globally.
var FormErrorMessage = "The form could not be saved."

// MaxFieldErrorLen bounds the error a field shows, and reports through
// [ReportError], for a value a user posted: the parse error of a text,
// number, select or checkbox-set field of html.FieldDecider (and of the
// deciders built on it, such as shadcn's) whose text is longer than this many
// bytes and carries no [UserMessage] reads "the text is not a valid value"
// instead, as such an error typically quotes the posted text whole, a forged
// value of any size included. A short error of a long value shows as it is.
// (Date and time fields take no text longer than 64 bytes at all.) 0 or less
// means no bound. Reassign at init time to change it globally.
var MaxFieldErrorLen = 512

// FormMessage is an error whose text is meant for the user, the
// [UserMessageError] of go-mx itself. When onSubmit returns one, also
// wrapped, the form shows its text as the form-level message.
type FormMessage string

// Error returns the message.
func (m FormMessage) Error() string { return string(m) }

// UserMessage returns the message, which makes a FormMessage a
// [UserMessageError].
func (m FormMessage) UserMessage() string { return string(m) }

// ReflectFormConfig customizes a single [ReflectFormHandlerWith]
// invocation. Zero-valued fields fall back to their package-level
// defaults.
type ReflectFormConfig struct {
	// MaxMemory bounds the multipart form parser and the size of an upload
	// a file field reads into memory ([FormMaxMemory]). 0 means use
	// [DefaultMaxMemory].
	MaxMemory int64

	// MaxUploadSize bounds the size of a multipart request body; past it
	// a POST answers 413. 0 means use [DefaultMaxUploadSize], and a
	// negative value means no bound, for a handler behind a limit of its
	// own: an http.MaxBytesReader with a larger limit is cut to this one.
	// A urlencoded body has net/http's 10 MB limit, or the limit of an
	// http.MaxBytesReader around it.
	MaxUploadSize int64

	// Redirect produces the URL to 303-redirect to on successful
	// submit. nil means "redirect to r.URL.Path" — the same handler
	// re-renders the freshly-saved form.
	Redirect func(*http.Request) string

	// Action produces the URL the rendered <form> submits to (its
	// action attribute). nil means "submit to the request URI that
	// served the form" (path+query, forced same-origin — see
	// [selfSubmitAction]), so a fragment-embedded form (loaded via
	// hx-get into a GET-only page) posts back to its own handler rather
	// than to the embedding page. A non-nil Action is trusted and
	// emitted as-is (escaped, but not scheme/host allow-listed), so do
	// not feed it untrusted input.
	Action func(*http.Request) string

	// SubmitLabel overrides [FormSubmitLabel] for this handler only.
	SubmitLabel string
}

// ReflectFormHandler builds an http.Handler that renders, parses, and
// validates a form for *T by reflecting target struct fields.
//
// On GET it calls load(r.Context()) to obtain the seed *T, walks via
// [ReflectFormFields], asks the request-context [FieldDecider] to
// render each visited field, and writes the resulting HTML.
//
// On POST it calls load again to obtain a fresh canonical record,
// parses ONLY the fields the form actually rendered into that record
// (allowlist by construction — never mass-assignment), runs the
// per-field validation chain via the decider, and either re-renders
// with inline errors or calls onSubmit. If onSubmit returns a
// [FieldErrors] the form re-renders with each entry routed to the
// matching field; an entry for a path no input shows goes above the form
// as below, and a FieldErrors of only nil entries is a successful submit.
// On any other error the form re-renders with a
// form-level message, [FormErrorText]: the [UserMessage] of a
// [UserMessageError] such as a [FormMessage], or [FormErrorMessage], with
// the error passed to [ReportError]. On
// success the handler responds with 303 See Other to cfg.Redirect or
// r.URL.Path when nil. A body that cannot be parsed is answered with 400,
// or 413 past the multipart limits, net/http's 10 MB for a urlencoded
// body or a limit set with [http.MaxBytesReader], never treated as a
// submission of nothing, also when a middleware parsed it first and
// ignored the error; a body that is no form (urlencoded or multipart) is
// answered with 415. A middleware that reads a multipart form must parse
// all of it (r.ParseMultipartForm or r.FormValue): after r.ParseForm alone
// the handler cannot tell an unread multipart body from a half-read one
// and answers 400.
//
// load must return a record the handler may change: it binds the
// submission into it, including text a field rejected, which stays there for
// the re-render. Return a deep copy, never a pointer into a cache or store:
// a field of a pointer type binds into a new pointee, but the fields of a
// pointer section, and of a repeatable row's pointer sections, are written
// in place.
//
// When onSubmit returns an error, the form re-renders the record as
// onSubmit left it. It must not reorder or remove the rows of a repeatable
// field then: rows are matched to the loaded rows by position, which their
// readonly and rejected cells are taken from on the next save. Rows it
// appends have no loaded row.
//
// load may be nil for submit-only forms (registration, contact, new
// record) where there is no existing state to load: the handler then
// seeds every GET and every POST with a fresh new(T) instead.
//
// The optional variadic decider, if supplied, overrides
// [DeciderFromContext] for this specific handler.
func ReflectFormHandler[T any](
	load func(ctx context.Context) (*T, error),
	onSubmit func(ctx context.Context, t *T) error,
	decider ...FieldDecider,
) http.HandlerFunc {
	return ReflectFormHandlerWith(ReflectFormConfig{}, load, onSubmit, decider...)
}

// ReflectFormHandlerWith is the configurable variant of
// [ReflectFormHandler]. See [ReflectFormConfig] for the available
// knobs. It is a thin wrapper over [ReflectFormHandlerRedirecting] whose
// onSubmit has no per-submission destination, so a successful submit
// always redirects to cfg.Redirect (or r.URL.Path when nil).
func ReflectFormHandlerWith[T any](
	cfg ReflectFormConfig,
	load func(ctx context.Context) (*T, error),
	onSubmit func(ctx context.Context, t *T) error,
	decider ...FieldDecider,
) http.HandlerFunc {
	if onSubmit == nil {
		panic("mx.ReflectFormHandler: onSubmit must not be nil")
	}
	return ReflectFormHandlerRedirecting(cfg, load,
		func(ctx context.Context, t *T) (string, error) {
			// No per-submission URL: the empty string makes the
			// redirecting handler fall back to cfg.Redirect / r.URL.Path,
			// exactly as this handler did before.
			return "", onSubmit(ctx, t)
		},
		decider...,
	)
}

// ReflectFormHandlerRedirecting is like [ReflectFormHandlerWith], but its
// onSubmit returns the URL to 303-redirect to on success. An empty string
// falls back to cfg.Redirect, then to r.URL.Path — identical to the
// non-redirecting handlers.
//
// This is the variant for a form whose submission creates an external
// resource whose URL the caller cannot know up front: a Stripe Checkout
// Session, a signed upload URL, an OAuth handoff, a generated share link.
// onSubmit is the only place that learns the destination, so returning it
// keeps that flow inside the handler instead of forcing a request-scoped
// mutable holder or abandoning ReflectFormHandler for that one form.
//
// It is a separate function rather than a [ReflectFormConfig] field on
// purpose: the config is not generic, and a func(*http.Request, *T) field
// would make the whole struct generic and break every existing
// ReflectFormConfig{} literal. Returning the URL from onSubmit is explicit,
// typed, and does not overload the error channel — a returned error still
// re-renders the form, and a [FieldErrors] still routes per field.
func ReflectFormHandlerRedirecting[T any](
	cfg ReflectFormConfig,
	load func(ctx context.Context) (*T, error),
	onSubmit func(ctx context.Context, t *T) (redirect string, err error),
	decider ...FieldDecider,
) http.HandlerFunc {
	if load == nil {
		load = func(context.Context) (*T, error) { return new(T), nil }
	}
	if onSubmit == nil {
		panic("mx.ReflectFormHandlerRedirecting: onSubmit must not be nil")
	}

	maxMem := cfg.MaxMemory
	if maxMem <= 0 {
		maxMem = DefaultMaxMemory
	}
	maxUpload := cfg.MaxUploadSize
	if maxUpload == 0 {
		maxUpload = DefaultMaxUploadSize
	}
	submitLabel := cfg.SubmitLabel
	if submitLabel == "" {
		submitLabel = FormSubmitLabel
	}

	pickDecider := func(r *http.Request) FieldDecider {
		if len(decider) > 0 && decider[0] != nil {
			return decider[0]
		}
		return DeciderFromContext(r.Context())
	}

	render := func(w http.ResponseWriter, r *http.Request, target *T, fieldErrs map[FieldPath][]error, formMsg string, origins rowOrigins) {
		d := pickDecider(r)
		action := selfSubmitAction(r)
		if cfg.Action != nil {
			action = cfg.Action(r)
		}
		comp := buildFormComponent(r.Context(), target, d, fieldErrs, formMsg, submitLabel, action, origins)
		writeFormResponse(w, r, comp)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			target, err := load(r.Context())
			if err != nil {
				respondLoadError(w, r, err)
				return
			}
			render(w, r, target, nil, "", nil)

		case http.MethodPost:
			// A body that fails to parse, such as one past the multipart
			// part limit, must not read as a submission of nothing: onSubmit
			// would save the unchanged record and drop the user's edits.
			loaded := r.MultipartForm
			if parseErr := parseFormBody(w, r, maxMem, maxUpload); parseErr != nil {
				status := http.StatusBadRequest
				var tooLarge *http.MaxBytesError
				switch {
				case errors.Is(parseErr, errNotAForm):
					status = http.StatusUnsupportedMediaType
				case errors.Is(parseErr, multipart.ErrMessageTooLarge) || errors.As(parseErr, &tooLarge) || parseErr.Error() == errPostTooLarge:
					status = http.StatusRequestEntityTooLarge
				}
				http.Error(w, "could not parse form", status)
				return
			}
			// The server removes the temporary files only of a form set on
			// the request it passed in, which a middleware's copy is not.
			if form := r.MultipartForm; form != nil && form != loaded {
				defer form.RemoveAll()
			}
			r = r.WithContext(context.WithValue(r.Context(), maxMemoryKey{}, maxMem))
			target, err := load(r.Context())
			if err != nil {
				respondLoadError(w, r, err)
				return
			}
			d := pickDecider(r)
			origins := rowOrigins{}
			// Repeatable add/remove buttons submit a __cmd__ value with
			// formnovalidate: bind whatever the user entered, apply the
			// row mutation, and re-render without running onSubmit. Only
			// commands that resolve to a real repeatable field enter this
			// path; an unknown/injected __cmd__ falls through to a normal
			// submit rather than silently discarding the save.
			if cmd := repeatableCommand(r); cmd != "" && isRepeatableCommand(target, d, cmd) {
				// The command saves nothing and shows no error, so the errors
				// of this parse, which the next save reports, are not reported.
				parseAndValidate(target, d, r.WithContext(withoutReports(r.Context())), origins)
				applyRepeatableCommand(target, cmd, requestFormValues(r), origins)
				render(w, r, target, nil, "", origins)
				return
			}
			fieldErrs := parseAndValidate(target, d, r, origins)
			if len(fieldErrs) > 0 {
				render(w, r, target, fieldErrs, "", origins)
				return
			}
			redirect, err := onSubmit(r.Context(), target)
			if fe, ok := errors.AsType[FieldErrors](err); ok {
				// A nil entry, as from FieldErrors{"Name": validate(x)} with a
				// valid x, is no error.
				perField := map[FieldPath][]error{}
				for p, e := range fe {
					if e != nil {
						perField[p] = []error{e}
					}
				}
				if len(perField) > 0 {
					render(w, r, target, perField, "", origins)
					return
				}
				// A FieldErrors whose entries are all nil is no error at all,
				// unless it came wrapped or joined with another error.
				if _, bare := err.(FieldErrors); bare {
					err = nil
				}
			}
			if err != nil {
				render(w, r, target, nil, FormErrorText(r.Context(), err), origins)
				return
			}
			// onSubmit's returned URL wins; an empty string falls back to
			// cfg.Redirect, then r.URL.Path — byte-identical to the
			// non-redirecting handler's success path.
			dest := redirect
			if dest == "" {
				dest = r.URL.Path
				if cfg.Redirect != nil {
					dest = cfg.Redirect(r)
				}
			}
			http.Redirect(w, r, dest, http.StatusSeeOther)

		default:
			w.Header().Set("Allow", "GET, HEAD, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// parseAndValidate walks target, applies the form's submitted values
// onto it (only for fields whose __present sentinel is set), and runs
// each per-field [FieldBehavior.Validate]. Returns a map of paths to
// the errors that prevented save; an empty map means the form is
// ready for onSubmit. The origins of the rebuilt repeatable rows are
// recorded in origins.
func parseAndValidate[T any](target *T, d FieldDecider, r *http.Request, origins rowOrigins) map[FieldPath][]error {
	out := map[FieldPath][]error{}
	form := requestFormValues(r)
	for visit := range ReflectFormFields(target) {
		if visit.Kind == FieldKindRepeatable {
			if repeatableInForm(visit, d) {
				parseRepeatable(visit, d, r, out, origins)
			}
			continue
		}
		if !visit.Value.CanAddr() {
			out[visit.Path] = append(out[visit.Path],
				errors.New("internal: field value not addressable"))
			continue
		}
		parseField(visit, d, r, form, visit.Path, visit.Path, out)
	}
	return out
}

// validateField runs the validation chain (see [validateBound]) and then
// the decider-supplied Validate on a bound field value, at the top level
// and in a repeatable row alike. The errors of a form:"sensitive" field
// read "not a valid value", and are reported ([ReportHiddenError]): a
// type's own message may quote the value, which after a blank submission is
// the stored one. A [UserMessageError], such as a password rule, is meant
// for the user and shows as its UserMessage.
func validateField(ctx context.Context, visit FieldVisit, beh FieldBehavior, path FieldPath, blank bool) []error {
	errs := validateBound(visit.Kind, visit.Value, blank)
	if beh.Validate != nil {
		if err := beh.Validate(path, visit.Field, visit.Value); err != nil {
			errs = append(errs, err)
		}
	}
	if !visit.Tag.Sensitive || len(errs) == 0 {
		return errs
	}
	var shown []error
	masked := false
	for _, err := range errs {
		// Only the message itself: an error wrapping or joined with it may
		// quote the value.
		if msg, ok := UserMessage(err); ok {
			shown = append(shown, FormMessage(msg))
			continue
		}
		ReportHiddenError(ctx, err)
		if !masked {
			// Meant for the user: where no input shows it, it goes above the
			// form as it is, neither hidden nor reported again.
			shown = append(shown, FormMessage("not a valid value"))
			masked = true
		}
	}
	return shown
}

// validateBound runs the validation chain on a bound field value, at the
// top level and in a repeatable row alike. A blank optional enum, one
// submitted empty ([sentBlank]), is "no choice", not an invalid constant:
// go-enum Validate rejects the zero string, which would make the leading
// "(none)" option impossible to save (a decider Validate still runs: that
// hook is the caller's, not the type's enum check). An enum zero bound
// from text that was not empty, such as a forged "0", is validated. The
// elements of an enum set carry the enum's Validate, not the map or slice
// holding them, so a set is validated element by element unless it
// validates itself.
func validateBound(kind FieldKind, value reflect.Value, blank bool) []error {
	if kind == FieldKindEnum && blank {
		return nil
	}
	chainErrs := RunValidationChain(value)
	if len(chainErrs) == 0 && kind == FieldKindEnumSet {
		chainErrs = validateSetElements(value)
	}
	return chainErrs
}

// parseField applies one submitted value onto cv.Value, for a top-level
// field (srcCell and dstCell are its path) and a repeatable row's cell
// (srcCell is the submitted cell path, dstCell its path in the rebuilt
// rows): readonly skip (the value stays as loaded, or as the row started),
// __present gate (the mass-assignment allowlist), __clear handling,
// keepsStored, decider Parse, required check, then the validation chain
// and decider Validate. Errors are keyed by dstCell.
func parseField(cv FieldVisit, d FieldDecider, r *http.Request, form map[string][]string, srcCell, dstCell FieldPath, out map[FieldPath][]error) {
	if cv.Tag.Readonly {
		// Never a submitted value, not even the null of __clear.
		return
	}
	// A value not sent is left alone, but an empty required one, as in a
	// submit-only form's new(T) or an added row, is still an error. The
	// sentinel is checked first: a field that was not sent costs no decider
	// call, so a POST of bare row markers stays cheap.
	present := sentinelSet(form, PresentSentinelName(srcCell))
	if !present && !(cv.Tag.Required && isBlank(cv.Kind, cv.Value)) {
		return
	}
	beh := d(srcCell, cv.Field, cv.Value)
	// A field the decider does not render, or renders as nothing, is not
	// in the form: it is never written, also not by a forged __present
	// sentinel, and shows no error.
	if !fieldInForm(beh, srcCell, cv) {
		return
	}
	if !present {
		out[dstCell] = append(out[dstCell], errors.New("required"))
		return
	}
	keep := true // a blank sensitive input may keep the stored value
	// A field the decider does not parse is not written, not even cleared.
	if beh.Parse != nil && sentinelSet(form, ClearSentinelName(srcCell)) {
		// The forms offer no clear control on a required field, so a clear
		// there is a tampered POST; it is rejected, not saved.
		if cv.Tag.Required {
			out[dstCell] = append(out[dstCell], errors.New("required"))
			return
		}
		cleared, err := clearField(cv.Tag, cv.Kind, cv.Value, form, srcCell)
		if err != nil {
			out[dstCell] = append(out[dstCell], err)
		}
		if cleared {
			return
		}
		keep = false
	}
	if beh.Parse != nil && !(keep && keepsStored(cv.Tag, cv.Kind, cv.Value, form, srcCell)) {
		if err := beh.Parse(srcCell, cv.Field, cv.Value, r); err != nil {
			out[dstCell] = append(out[dstCell], err)
			return
		}
	}
	blank := sentBlank(cv, form, srcCell, beh.Parse != nil && !(keep && keepsStored(cv.Tag, cv.Kind, cv.Value, form, srcCell)))
	if cv.Tag.Required && blank {
		// Validation results shown to the user carry no callstack.
		out[dstCell] = append(out[dstCell], errors.New("required"))
		return
	}
	if errs := validateField(r.Context(), cv, beh, dstCell, blank); len(errs) > 0 {
		out[dstCell] = append(out[dstCell], errs...)
	}
}

// fieldInForm reports whether the decider's behavior beh puts the field of
// cv at path into the form: a nil Render, or one that returns nil, leaves
// it out. Render is asked with a probe ([IsRenderProbe]), which the
// built-in deciders answer without building markup.
func fieldInForm(beh FieldBehavior, path FieldPath, cv FieldVisit) bool {
	return beh.Render != nil && beh.Render(path, cv.Field, cv.Value, renderProbe()) != nil
}

// keepsStored reports whether a submission leaves a field's stored value
// alone: a form:"sensitive" field never shows its value, so leaving it
// blank means "unchanged", as with a password field. Since the input stays
// empty, a value typed into it before an add-row or remove-row click is
// not shown again and must be typed again before saving. An optional one
// is emptied with its clear control. With nothing stored
// there is nothing to keep, so a blank is parsed, and a required field or
// a number's range still rejects it. A checkbox, checkbox set, hidden or
// file field has no such blank, and a select shows its stored option.
// A parser reads the first submitted value, as form.Get returns it.
func keepsStored(tag FormTag, kind FieldKind, value reflect.Value, form url.Values, path FieldPath) bool {
	return KeepsStoredValue(tag, kind, value) && form.Get(string(path)) == ""
}

// clearField applies a __clear sentinel: it sets value to null, and cleared
// reports that nothing is left to parse. A form:"sensitive" input that keeps
// a blank's stored value gets the clear control too, rendered ticked while
// the value is empty. Text typed into it wins over the tick, and a value
// that has no null is cleared by parsing the blank, so its range, required
// check and validation apply; cleared is false for both.
func clearField(tag FormTag, kind FieldKind, value reflect.Value, form url.Values, path FieldPath) (cleared bool, err error) {
	sensitive := tag.Sensitive && kind.KeepsStoredWhenBlank()
	if sensitive && form.Get(string(path)) != "" {
		return false, nil
	}
	if err := setFieldNull(value); err != nil {
		if sensitive {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

// errPostTooLarge is the text of the error net/http's ParseForm returns for
// a urlencoded body past its 10 MB limit, which it applies unless the body
// is an http.MaxBytesReader with a limit of its own. Either answers 413, as
// a body past the multipart limits does.
const errPostTooLarge = "http: POST too large"

// errNotAForm answers a POST whose body is no form, such as JSON or one
// without a content type: parsing it reads nothing, and onSubmit would run
// on a submission of nothing.
var errNotAForm = errors.New("the body is not a form")

// errPreParseFailed answers a body a middleware parsed before the handler
// and failed to read (see [preParseFailed]).
var errPreParseFailed = errors.New("the form body was parsed before the handler and could not be read")

// preParseFailed reports whether the body of r, whose form a middleware
// parsed before the handler, evidently failed to parse: a multipart body
// with no multipart form, which a stopped read leaves half consumed, or a
// urlencoded body with content but no values, as one past a limit leaves
// it. A middleware such as a method override reading r.FormValue ignores
// that error, and the handler must not save what is left. A middleware
// parsing a multipart form must parse all of it (r.ParseMultipartForm or
// r.FormValue), not only r.ParseForm.
func preParseFailed(r *http.Request, mediaType string) bool {
	switch mediaType {
	case ContentTypeMultipartFormData:
		return r.MultipartForm == nil
	case ContentTypeWWWFormURLEncoded:
		return len(r.PostForm) == 0 && r.ContentLength != 0
	}
	return false
}

// parseFormBody parses the request body into r.PostForm, r.Form and, for a
// multipart body, r.MultipartForm. Only the body is the submission, so it
// is parsed on its own: a URL query that cannot be parsed, such as one with
// a semicolon, fails nothing, and its readable values join r.Form. A form
// a middleware parsed already holds the query, unless its parse evidently
// failed ([preParseFailed]). A multipart body past maxUpload, when it is
// positive, is too large.
func parseFormBody(w http.ResponseWriter, r *http.Request, maxMem, maxUpload int64) error {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != ContentTypeWWWFormURLEncoded && mediaType != ContentTypeMultipartFormData {
		return errNotAForm
	}
	if mediaType == ContentTypeMultipartFormData && maxUpload > 0 && r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	}
	parsed := r.Form != nil
	if parsed && preParseFailed(r, mediaType) {
		return errPreParseFailed
	}
	body := *r
	bodyURL := *r.URL
	bodyURL.RawQuery = ""
	body.URL = &bodyURL
	err := body.ParseForm()
	if mediaType == ContentTypeMultipartFormData {
		err = body.ParseMultipartForm(maxMem)
	}
	r.Form, r.PostForm, r.MultipartForm = body.Form, body.PostForm, body.MultipartForm
	if r.Form == nil {
		r.Form = url.Values{}
	}
	if !parsed {
		query, _ := url.ParseQuery(r.URL.RawQuery)
		for name, values := range query {
			r.Form[name] = append(r.Form[name], values...)
		}
	}
	return err
}

type maxMemoryKey struct{}

// FormMaxMemory returns the [ReflectFormConfig.MaxMemory] of the reflected
// form handler parsing the request of ctx, or [DefaultMaxMemory]. A
// decider's file binder reads no larger upload into memory.
func FormMaxMemory(ctx context.Context) int64 {
	if n, ok := ctx.Value(maxMemoryKey{}).(int64); ok {
		return n
	}
	return DefaultMaxMemory
}

// FormErrorText returns the text a form shows the user for err, an error
// it cannot show at a field, such as one onSubmit returned that is no
// [FieldErrors]: its [UserMessage], else [FormErrorMessage], or the error's
// own text when [RevealInternalServerErrors] is set, and "" for a nil err.
// An error that is no user message is passed to [ReportError] with ctx
// ([ReportHiddenError]). A decider of the application can use it for an
// error it does not show.
//
// Unlike a 500, the form is re-rendered also for a context.Canceled or
// context.DeadlineExceeded, which may come from a call onSubmit made while
// the client is still there, so it gets a message like any other error; it
// is not reported.
func FormErrorText(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}
	if msg, ok := UserMessage(err); ok {
		return msg
	}
	ReportHiddenError(ctx, err)
	if RevealInternalServerErrors {
		return err.Error()
	}
	return FormErrorMessage
}

// isBlank reports whether a value not sent counts as no value for a
// required check: [isEffectivelyEmpty], or the zero of an enum of any kind,
// the value an empty select stores, which isEffectivelyEmpty never sees in
// a number.
func isBlank(kind FieldKind, value reflect.Value) bool {
	return isEffectivelyEmpty(value) || (kind == FieldKindEnum && value.IsZero())
}

// sentBlank reports whether a sent field holds no value, for the required
// check and the validation chain. An enum the decider parsed is blank when
// its submitted text was empty and it bound no value, as a parser that
// stores the zero for an empty select leaves it; text that bound a zero,
// such as a Low = iota or an option "0", is a value, which the validation
// chain then checks. Any other field, also an enum the decider did not
// parse from the submission, is blank when [isEffectivelyEmpty].
func sentBlank(cv FieldVisit, form url.Values, path FieldPath, parsed bool) bool {
	if cv.Kind == FieldKindEnum && parsed {
		return form.Get(string(path)) == "" && (cv.Value.IsZero() || isEffectivelyEmpty(cv.Value))
	}
	return isEffectivelyEmpty(cv.Value)
}

// validateSetElements runs [RunValidationChain] on every element of a
// set-shaped value, the keys of a map[T]struct{} or the elements of a
// []T, so a submitted value that is none of the values of an enum T is
// rejected like in a single enum field. The errors of a map are sorted
// by message, because map iteration order is random.
func validateSetElements(value reflect.Value) []error {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	var elemErrs []error
	switch value.Kind() {
	case reflect.Map:
		for key := range value.Seq() {
			elemErrs = append(elemErrs, RunValidationChain(key)...)
		}
		slices.SortFunc(elemErrs, func(a, b error) int { return strings.Compare(a.Error(), b.Error()) })
	case reflect.Slice:
		for _, elem := range value.Seq2() {
			elemErrs = append(elemErrs, RunValidationChain(elem)...)
		}
	}
	return elemErrs
}

func sentinelSet(form map[string][]string, name string) bool {
	vals, ok := form[name]
	if !ok {
		return false
	}
	for _, v := range vals {
		if v != "" {
			return true
		}
	}
	return false
}

// setFieldNull asks the addressable value to clear itself. Tries
// NullSetter first, then a database/sql type's Scan(nil) on a copy (such
// as sql.NullString); falls back to reflect-level zeroing for pointer/
// slice/map/interface kinds.
func setFieldNull(value reflect.Value) error {
	if !value.IsValid() {
		return errors.New("internal: invalid value for clear")
	}
	if value.CanAddr() {
		if ns, ok := value.Addr().Interface().(NullSetter); ok {
			ns.SetNull()
			return nil
		}
		// A database/sql type, such as sql.NullString, scans nil as its
		// null, into a copy: a Scan that fails may change its value.
		if _, ok := value.Addr().Interface().(sqlScanner); ok && value.Kind() == reflect.Struct {
			null := reflect.New(value.Type())
			if null.Interface().(sqlScanner).Scan(nil) == nil {
				value.Set(null.Elem())
				return nil
			}
		}
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if value.CanSet() {
			value.SetZero()
			return nil
		}
	}
	return errors.New("field is not nullable")
}

// isEffectivelyEmpty reports whether value should be treated as
// "empty" for purposes of a Required check. Numeric and bool fields
// are never empty (a value is always submitted via __present).
//
// An array or struct is empty when it is its zero value: an empty
// submission stores the zero value of a fixed-size ID such as uuid.UUID
// or of a time.Time, so a length check alone would accept a blank ID as
// uuid.Nil(). A [Zeroable] type decides that itself, as time.Time does.
func isEffectivelyEmpty(value reflect.Value) (empty bool) {
	// A method that panics, as one promoted through a nil embedded pointer
	// does (IsNull, IsZero), holds no value.
	defer func() {
		if recover() != nil {
			empty = true
		}
	}()
	if !value.IsValid() {
		return true
	}
	if IsNull(safeInterface(value)) {
		return true
	}
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Slice, reflect.Map:
		return value.Len() == 0
	case reflect.Array, reflect.Struct:
		if z, ok := safeInterface(value).(Zeroable); ok {
			return z.IsZero()
		}
		return value.IsZero()
	case reflect.Pointer, reflect.Interface:
		return value.IsNil()
	}
	return false
}

// selfSubmitAction returns the default <form> action when
// [ReflectFormConfig.Action] is not set: the request URI that served the
// form (path+query), so a native submit posts back to this handler.
//
// The request URI is untrusted request input. It reaches the rendered
// action attribute where the CheckedWriter escapes it, so it cannot break
// out of the attribute — but escaping does not change URL resolution. A
// path beginning with "//" is a protocol-relative reference that a browser
// resolves off-origin, which would post the form's fields off site on a
// native submit. Collapse a leading run of slashes to a single "/" so the
// default action is always a same-origin absolute path — which is also
// what net/http's ServeMux redirects such requests to. (A backslash-based
// "/\" reference cannot occur here: [net/url.URL.RequestURI] percent-
// encodes backslashes to %5C, which browsers keep same-origin.)
func selfSubmitAction(r *http.Request) string {
	uri := r.URL.RequestURI()
	if len(uri) > 1 && uri[0] == '/' && uri[1] == '/' {
		i := 0
		for i < len(uri) && uri[i] == '/' {
			i++
		}
		uri = "/" + uri[i:]
	}
	return uri
}

// buildFormComponent constructs the top-level form Component.
//
// The handler emits the surrounding <form> element with method=post, an
// enctype of multipart/form-data when a file field, or one whose
// [FieldBehavior].Multipart is set, renders and
// application/x-www-form-urlencoded otherwise, and the given action (by default the URL
// that served the form — see [selfSubmitAction] — or the caller's
// [ReflectFormConfig.Action] when set), then asks the decider to render
// each field (in walk order, grouped by section). The explicit action
// makes the form self-submit to its own handler even when it is loaded
// as an HTMX fragment into a GET-only page — without it the native
// submit would post to the embedding document URL and get a 405. A
// submit button is appended so the form is usable out of the box;
// renderers can layer their own buttons on top by returning extra
// children.
func buildFormComponent[T any](ctx context.Context, target *T, d FieldDecider, fieldErrs map[FieldPath][]error, formMsg string, submitLabel string, action string, origins rowOrigins) Component {
	type sectionEntry struct {
		name       string
		components []Component
	}
	var (
		messages    []Component
		rowCommands bool // a repeatable field renders its row buttons
		hasFile     bool // a file input needs a multipart body
		root        []Component
		sections    []*sectionEntry
		byName      = map[string]*sectionEntry{}
		// consumed holds the paths whose errors a rendered input shows.
		consumed = map[FieldPath]bool{}
	)
	if formMsg != "" {
		messages = append(messages, formMessageComponent(formMsg))
	}
	for visit := range ReflectFormFields(target) {
		var wrapped Component
		if visit.Kind == FieldKindRepeatable {
			if !repeatableInForm(visit, d) {
				continue
			}
			rowCommands = rowCommands || !visit.Tag.Readonly
			// The repeatable field owns its own row markup, __present
			// sentinels and add/remove buttons; the core handler only
			// places it into the right section.
			var rowFiles bool
			wrapped, rowFiles = renderRepeatable(visit, d, fieldErrs, origins, consumed)
			hasFile = hasFile || rowFiles
		} else {
			beh := d(visit.Path, visit.Field, visit.Value)
			if beh.Render == nil {
				continue
			}
			errs := fieldErrs[visit.Path]
			fieldComp := beh.Render(visit.Path, visit.Field, visit.Value, errs)
			if fieldComp == nil {
				// A field the decider renders as nothing is not in the form,
				// so it gets no __present sentinel either: one would make
				// every save blank it.
				continue
			}
			// A hidden input shows no error of its own.
			if _, hasErrs := fieldErrs[visit.Path]; hasErrs && visit.Kind != FieldKindHidden {
				consumed[visit.Path] = true
			}
			hasFile = hasFile || visit.Kind == FieldKindFile || beh.Multipart
			// Each render emits its own __present sentinel; the handler
			// ALSO ensures one is present (defense-in-depth) when the
			// decider chose to omit it.
			wrapped = wrapWithPresentSentinel(visit.Path, fieldComp)
		}
		if wrapped == nil {
			continue
		}
		if visit.Section == "" {
			root = append(root, wrapped)
			continue
		}
		s, ok := byName[visit.Section]
		if !ok {
			s = &sectionEntry{name: visit.Section}
			byName[visit.Section] = s
			sections = append(sections, s)
		}
		s.components = append(s.components, wrapped)
	}
	// An error no input shows, of a hidden field or row cell, a field the
	// decider does not render or a path that names no field, must not fail
	// the save silently: it becomes a form-level message, written as an
	// onSubmit error is ([FormErrorText]), once per distinct text. Each
	// error is reported, also one whose text is shown already.
	unshown := make([]FieldPath, 0, len(fieldErrs))
	for path := range fieldErrs {
		if !consumed[path] {
			unshown = append(unshown, path)
		}
	}
	slices.Sort(unshown)
	shownMsgs := map[string]bool{}
	for _, path := range unshown {
		for _, err := range fieldErrs[path] {
			if msg := FormErrorText(ctx, err); !shownMsgs[msg] {
				shownMsgs[msg] = true
				messages = append(messages, formMessageComponent(msg))
			}
		}
	}

	children := make([]any, 0, len(messages)+len(root)+len(sections)+2)
	if rowCommands {
		children = append(children, defaultSubmitButton())
	}
	for _, c := range messages {
		children = append(children, c)
	}
	for _, c := range root {
		children = append(children, c)
	}
	// Section order: stable, by first-encountered field — but tests
	// commonly expect alphabetical, so we sort to make output
	// deterministic across compilers.
	sort.SliceStable(sections, func(i, j int) bool {
		return sections[i].name < sections[j].name
	})
	for _, s := range sections {
		children = append(children, renderSection(s.name, s.components))
	}
	children = append(children, submitButton(submitLabel))

	// A form posts urlencoded, which allows 10,000 values, unless a file
	// input needs multipart, whose limit is 1,000 parts: every field takes
	// two (its value and __present sentinel), so 100 rows of 5 cells would
	// already be rejected.
	enctype := ContentTypeWWWFormURLEncoded
	if hasFile {
		enctype = ContentTypeMultipartFormData
	}
	formAttribs := []any{
		Attribute{Name: "method", Value: "post"},
		Attribute{Name: "enctype", Value: enctype},
		Attribute{Name: "action", Value: action},
	}
	all := append(formAttribs, children...)
	return NewElement("form", all...)
}

// renderSection builds a simple <fieldset><legend>name</legend>...
// section. Renderer layers (shadcn) will swap this for shadcn.Card via
// their own decider Render closures; this default exists so a bare
// mx-only application still produces a usable layout.
func renderSection(name string, comps []Component) Component {
	children := make([]any, 0, len(comps)+1)
	children = append(children, NewElement("legend", name))
	for _, c := range comps {
		children = append(children, c)
	}
	return NewElement("fieldset", children...)
}

// submitButton emits a <button type="submit">label</button>.
func submitButton(label string) Component {
	return NewElement("button",
		Attribute{Name: "type", Value: "submit"},
		Text(label),
	)
}

// defaultSubmitButton is the form's first submit button when rows render
// their "Remove" buttons before the Save button: pressing Enter in an
// input clicks the first submit button, which would remove a row instead
// of saving. It is out of sight and out of the tab order rather than
// display:none, which some browsers skip for that implicit submission.
func defaultSubmitButton() Component {
	return NewElement("button",
		Attribute{Name: "type", Value: "submit"},
		Attribute{Name: "tabindex", Value: "-1"},
		Attribute{Name: "aria-hidden", Value: "true"},
		Attribute{Name: "style", Value: "position:absolute;width:1px;height:1px;margin:-1px;padding:0;border:0;overflow:hidden;clip-path:inset(50%)"},
	)
}

// formMessageComponent emits a top-level <p role="alert">msg</p>.
func formMessageComponent(msg string) Component {
	return NewElement("p",
		Attribute{Name: "role", Value: "alert"},
		Text(msg),
	)
}

// wrapWithPresentSentinel ensures every rendered field is accompanied
// by its __present sentinel hidden input. The decider's Render is
// expected to emit it already, but a defensive wrap guarantees the
// allowlist contract even when a custom decider forgets.
func wrapWithPresentSentinel(path FieldPath, field Component) Component {
	hidden := NewVoidElement("input",
		Attribute{Name: "type", Value: "hidden"},
		Attribute{Name: "name", Value: PresentSentinelName(path)},
		Attribute{Name: "value", Value: "1"},
	)
	return Components{hidden, field}
}

// writeFormResponse writes comp to w using the package's default
// writer factory and the standard text/html content type. The
// component is rendered into a buffer first: deferred computations
// (context-dependent option providers, ErrAttrib, NewErrElement) can
// fail mid-render, and streaming would already have sent a 200 status
// and a truncated page. Form pages are small, so buffering is cheap
// and turns every render error into a clean 500.
func writeFormResponse(w http.ResponseWriter, r *http.Request, comp Component) {
	var buf bytes.Buffer
	writer := DefaultWriterFactory.NewWriter(&buf)
	if err := comp.Render(r.Context(), writer); err != nil {
		RespondNonContextError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", ContentTypeHTML)
	_, _ = buf.WriteTo(w)
}

// respondLoadError surfaces a load() failure as a 500 with a generic
// message and reports it to [ReportError]. The detailed error stays
// server-side unless [RevealInternalServerErrors] is true.
func respondLoadError(w http.ResponseWriter, r *http.Request, err error) {
	if msg, ok := hiddenErrorText(r.Context(), err, "failed to load record"); ok {
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
