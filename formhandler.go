package mx

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"reflect"
	"sort"

	"github.com/domonda/go-errs"
)

// DefaultMaxMemory is the default upper bound on multipart form parsing
// used by [ReflectFormHandler]. 4 MiB matches net/http's
// defaultMaxMemory and is large enough for typical admin forms without
// inviting a memory-pressure attack.
//
// Override globally by reassigning this variable at init time or per
// handler via [ReflectFormConfig.MaxMemory].
var DefaultMaxMemory int64 = 4 << 20

// FormSubmitLabel is the default text rendered on the form's submit
// button. Reassign to change globally; the renderer layer (html / hx /
// shadcn) may also render its own button instead of relying on this.
var FormSubmitLabel = "Save"

// ReflectFormConfig customizes a single [ReflectFormHandlerWith]
// invocation. Zero-valued fields fall back to their package-level
// defaults.
type ReflectFormConfig struct {
	// MaxMemory bounds the multipart form parser. 0 means use
	// [DefaultMaxMemory].
	MaxMemory int64

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
// matching field. On any other error the form re-renders with a
// form-level message. On success the handler responds with
// 303 See Other to cfg.Redirect or r.URL.Path when nil.
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
// knobs.
func ReflectFormHandlerWith[T any](
	cfg ReflectFormConfig,
	load func(ctx context.Context) (*T, error),
	onSubmit func(ctx context.Context, t *T) error,
	decider ...FieldDecider,
) http.HandlerFunc {
	if load == nil {
		load = func(context.Context) (*T, error) { return new(T), nil }
	}
	if onSubmit == nil {
		panic("mx.ReflectFormHandler: onSubmit must not be nil")
	}

	maxMem := cfg.MaxMemory
	if maxMem <= 0 {
		maxMem = DefaultMaxMemory
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

	render := func(w http.ResponseWriter, r *http.Request, target *T, fieldErrs map[FieldPath][]error, formMsg string) {
		d := pickDecider(r)
		action := selfSubmitAction(r)
		if cfg.Action != nil {
			action = cfg.Action(r)
		}
		comp := buildFormComponent(target, d, fieldErrs, formMsg, submitLabel, action)
		writeFormResponse(w, r, comp)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			target, err := load(r.Context())
			if err != nil {
				respondLoadError(w, err)
				return
			}
			render(w, r, target, nil, "")

		case http.MethodPost:
			if err := r.ParseMultipartForm(maxMem); err != nil {
				// Fallback for non-multipart submissions.
				if err2 := r.ParseForm(); err2 != nil {
					http.Error(w, "could not parse form", http.StatusBadRequest)
					return
				}
			}
			target, err := load(r.Context())
			if err != nil {
				respondLoadError(w, err)
				return
			}
			d := pickDecider(r)
			// Repeatable add/remove buttons submit a __cmd__ value with
			// formnovalidate: bind whatever the user entered, apply the
			// row mutation, and re-render without running onSubmit. Only
			// commands that resolve to a real repeatable field enter this
			// path; an unknown/injected __cmd__ falls through to a normal
			// submit rather than silently discarding the save.
			if cmd := repeatableCommand(r); cmd != "" && isRepeatableCommand(target, cmd) {
				parseAndValidate(target, d, r)
				applyRepeatableCommand(target, cmd, requestFormValues(r))
				render(w, r, target, nil, "")
				return
			}
			fieldErrs := parseAndValidate(target, d, r)
			if len(fieldErrs) > 0 {
				render(w, r, target, fieldErrs, "")
				return
			}
			if err := onSubmit(r.Context(), target); err != nil {
				if fe, ok := errors.AsType[FieldErrors](err); ok {
					perField := map[FieldPath][]error{}
					for p, e := range fe {
						perField[p] = []error{e}
					}
					render(w, r, target, perField, "")
					return
				}
				render(w, r, target, nil, err.Error())
				return
			}
			dest := r.URL.Path
			if cfg.Redirect != nil {
				dest = cfg.Redirect(r)
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
// ready for onSubmit.
func parseAndValidate[T any](target *T, d FieldDecider, r *http.Request) map[FieldPath][]error {
	out := map[FieldPath][]error{}
	form := requestFormValues(r)
	for visit := range ReflectFormFields(target) {
		if visit.Kind == FieldKindRepeatable {
			parseRepeatable(visit, d, r, out)
			continue
		}
		beh := d(visit.Path, visit.Field, visit.Value)
		if !visit.Value.CanAddr() {
			out[visit.Path] = append(out[visit.Path],
				errs.New("internal: field value not addressable"))
			continue
		}
		// __present gate: skip silently when the field was not
		// rendered. This is the allowlist that defends against
		// mass-assignment via tampered POSTs.
		if !sentinelSet(form, PresentSentinelName(visit.Path)) {
			continue
		}
		// __clear: explicit null intent for nullable fields.
		if sentinelSet(form, ClearSentinelName(visit.Path)) {
			if err := setFieldNull(visit.Value); err != nil {
				out[visit.Path] = append(out[visit.Path], err)
			}
			continue
		}
		// Readonly: render-only; never write a submitted value.
		if visit.Tag.Readonly {
			continue
		}
		if beh.Parse != nil {
			if err := beh.Parse(visit.Path, visit.Field, visit.Value, r); err != nil {
				out[visit.Path] = append(out[visit.Path], err)
				continue
			}
		}
		// Required check (type-aware) — relevant for string and
		// nullable values; the parsers already enforce numeric range
		// via min/max.
		if visit.Tag.Required && isEffectivelyEmpty(visit.Value) {
			// Validation results shown to the user carry no callstack.
			out[visit.Path] = append(out[visit.Path],
				errors.New("required"))
			continue
		}
		// Built-in validation chain.
		if chainErrs := RunValidationChain(visit.Value); len(chainErrs) > 0 {
			out[visit.Path] = append(out[visit.Path], chainErrs...)
		}
		// Decider-supplied Validate runs after the chain (additive).
		if beh.Validate != nil {
			if err := beh.Validate(visit.Path, visit.Field, visit.Value); err != nil {
				out[visit.Path] = append(out[visit.Path], err)
			}
		}
	}
	return out
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
// NullSetter first; falls back to reflect-level zeroing for pointer/
// slice/map/interface kinds.
func setFieldNull(value reflect.Value) error {
	if !value.IsValid() {
		return errs.New("internal: invalid value for clear")
	}
	if value.CanAddr() {
		if ns, ok := value.Addr().Interface().(NullSetter); ok {
			ns.SetNull()
			return nil
		}
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if value.CanSet() {
			value.SetZero()
			return nil
		}
	}
	return errs.New("field is not nullable")
}

// isEffectivelyEmpty reports whether value should be treated as
// "empty" for purposes of a Required check. Numeric and bool fields
// are never empty (a value is always submitted via __present).
func isEffectivelyEmpty(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	if IsNull(safeInterface(value)) {
		return true
	}
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return value.Len() == 0
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
// The handler emits the surrounding <form> element with method=post,
// enctype=multipart/form-data and the given action (by default the URL
// that served the form — see [selfSubmitAction] — or the caller's
// [ReflectFormConfig.Action] when set), then asks the decider to render
// each field (in walk order, grouped by section). The explicit action
// makes the form self-submit to its own handler even when it is loaded
// as an HTMX fragment into a GET-only page — without it the native
// submit would post to the embedding document URL and get a 405. A
// submit button is appended so the form is usable out of the box;
// renderers can layer their own buttons on top by returning extra
// children.
func buildFormComponent[T any](target *T, d FieldDecider, fieldErrs map[FieldPath][]error, formMsg string, submitLabel string, action string) Component {
	type sectionEntry struct {
		name       string
		components []Component
	}
	var (
		root     []Component
		sections []*sectionEntry
		byName   = map[string]*sectionEntry{}
	)
	if formMsg != "" {
		root = append(root, formMessageComponent(formMsg))
	}
	for visit := range ReflectFormFields(target) {
		var wrapped Component
		if visit.Kind == FieldKindRepeatable {
			// The repeatable field owns its own row markup, __present
			// sentinels and add/remove buttons; the core handler only
			// places it into the right section.
			wrapped = renderRepeatable(visit, d, fieldErrs)
		} else {
			beh := d(visit.Path, visit.Field, visit.Value)
			if beh.Render == nil {
				continue
			}
			errs := fieldErrs[visit.Path]
			// Each render emits its own __present sentinel; the handler
			// ALSO ensures one is present (defense-in-depth) when the
			// decider chose to omit it.
			fieldComp := beh.Render(visit.Path, visit.Field, visit.Value, errs)
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

	children := make([]any, 0, len(root)+len(sections)+1)
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

	formAttribs := []any{
		Attribute{Name: "method", Value: "post"},
		Attribute{Name: "enctype", Value: "multipart/form-data"},
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
		RespondNonContextError(w, err)
		return
	}
	w.Header().Set("Content-Type", ContentTypeHTML)
	_, _ = buf.WriteTo(w)
}

// respondLoadError surfaces a load() failure as a 500 with a generic
// message. The detailed error stays server-side unless
// [RevealInternalServerErrors] is true.
func respondLoadError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	msg := "failed to load record"
	if RevealInternalServerErrors {
		msg = err.Error()
	}
	http.Error(w, msg, http.StatusInternalServerError)
}
