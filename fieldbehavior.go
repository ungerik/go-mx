package mx

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/domonda/go-errs"
)

// FieldPath names a field inside a (possibly nested) struct as a
// hyphen-separated path (e.g. "AccountingEmail" or
// "Branding-PrimaryColor"). Hyphen is used as the separator because Go
// field names cannot contain it AND shadcn's validateID accepts it,
// letting the same string serve as an HTML input name AND a
// JS-interpolated id.
type FieldPath string

// Append returns a new FieldPath with name joined to p using "-".
// Empty p returns FieldPath(name); empty name returns p unchanged.
func (p FieldPath) Append(name string) FieldPath {
	if p == "" {
		return FieldPath(name)
	}
	if name == "" {
		return p
	}
	return p + "-" + FieldPath(name)
}

// FieldKind is the semantic category [DetectField] assigns to a struct
// field. Renderers use it to pick a widget; multiple distinct widgets
// may share a kind (e.g. FieldKindString covers text/email/url/tel/etc.
// — the [FormTag].Widget override or the type's [FormWidgetHint]
// disambiguates).
type FieldKind int

const (
	// FieldKindUnknown is the zero value and means the field was not
	// recognized; renderers must treat it as a catch-all.
	FieldKindUnknown FieldKind = iota

	// FieldKindSkip means the field is tagged form:"-" or
	// form:"widget=skip" and must not be rendered or parsed.
	FieldKindSkip

	// FieldKindHidden means the field is tagged form:"hidden": render
	// as a hidden input, round-trip the GET value through POST, never
	// display, never accept a user edit beyond what GET seeded.
	FieldKindHidden

	// FieldKindString covers any string-shaped scalar — string,
	// *string, named string types, and the email/url/tel/password
	// widget overrides which are still rendered as <input> of the
	// appropriate type.
	FieldKindString

	// FieldKindTextarea covers form:"widget=textarea" overrides as
	// well as []string (one item per line) and []byte/json.RawMessage
	// when no other override applies.
	FieldKindTextarea

	// FieldKindNumber covers int*/uint*/float* kinds plus their
	// pointer variants.
	FieldKindNumber

	// FieldKindBool covers bool and *bool.
	FieldKindBool

	// FieldKindDateTime covers time.Time (and pointer/Nullable
	// variants of it) as well as types registered via the
	// single-value registry with form:"widget=date|datetime|time".
	FieldKindDateTime

	// FieldKindFile covers form:"widget=file" overrides.
	FieldKindFile

	// FieldKindEnum covers any type implementing the Enums() or
	// EnumStrings() convention. The renderer chooses <select> by
	// default and <radio> when form:"widget=radio" is set.
	FieldKindEnum

	// FieldKindEnumSet covers map[T]struct{} (the set idiom) and []T
	// when T implements the enum convention. The renderer emits a
	// multi-checkbox grid.
	FieldKindEnumSet

	// FieldKindSection means the field is tagged form:"section=..."
	// and recurses into a labeled section.
	FieldKindSection

	// FieldKindNested means the field is tagged form:"nested" and
	// recurses into a section labeled with the field name.
	FieldKindNested

	// FieldKindRepeatable means the field is a slice of struct
	// (`[]T` or `[]*T`) tagged form:"repeatable": it renders as a
	// dynamic list of rows, one per slice element, and binds submitted
	// rows back into the slice. See repeatable.go.
	FieldKindRepeatable

	// FieldKindInline means the field is an anonymous embedded struct
	// with no section/nested tag and is inlined transparently — its
	// visible fields appear at the enclosing level.
	FieldKindInline

	// FieldKindCatchAll means no rule matched. The renderer falls back
	// to a text input and Parse requires the value type to implement
	// encoding.TextUnmarshaler.
	FieldKindCatchAll
)

// String returns the kind's name (handy in error messages and tests).
func (k FieldKind) String() string {
	switch k {
	case FieldKindSkip:
		return "skip"
	case FieldKindHidden:
		return "hidden"
	case FieldKindString:
		return "string"
	case FieldKindTextarea:
		return "textarea"
	case FieldKindNumber:
		return "number"
	case FieldKindBool:
		return "bool"
	case FieldKindDateTime:
		return "datetime"
	case FieldKindFile:
		return "file"
	case FieldKindEnum:
		return "enum"
	case FieldKindEnumSet:
		return "enum-set"
	case FieldKindSection:
		return "section"
	case FieldKindNested:
		return "nested"
	case FieldKindRepeatable:
		return "repeatable"
	case FieldKindInline:
		return "inline"
	case FieldKindCatchAll:
		return "catch-all"
	}
	return "unknown"
}

// FormWidgetHint lets a type self-declare its preferred widget. The
// returned name is one of the widget keys recognized by [FormTag]
// (text, email, url, tel, date, datetime, time, password, textarea,
// select, radio, switch, checkbox, file, hidden, skip). The
// [FieldDecider] consults this AFTER an explicit form:"widget=…" tag
// but BEFORE structural type detection.
type FormWidgetHint interface {
	FormWidget() string
}

// FieldBehavior describes the full lifecycle of one form field. The
// three closures are coupled because a single field overrides all
// three together — a custom currency widget needs to render its
// formatted input, parse its locale-aware string, and validate its
// range as one consistent unit.
type FieldBehavior struct {
	// Render emits the widget HTML plus the hidden __present sentinel
	// for this field (and the __clear sentinel when the field is
	// nullable). errs holds any validation errors to display inline;
	// it is nil on a fresh GET and may be non-nil after a failed POST.
	Render func(path FieldPath, field reflect.StructField, value reflect.Value, errs []error) Component

	// Parse reads the submitted value out of r and writes it into
	// value (which is addressable). The handler skips this call when
	// the field's __present sentinel is absent — never touching a
	// loaded record's field that the form did not render. When the
	// __clear sentinel is set, Parse calls NullSetter.SetNull on the
	// addressable value (no further parsing). Otherwise Parse decodes
	// the submitted value.
	Parse func(path FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error

	// Validate runs after Parse succeeds. Returning a non-nil error
	// includes it in the inline error list; the decider may
	// concatenate multiple errors via errors.Join.
	Validate func(path FieldPath, field reflect.StructField, value reflect.Value) error
}

// FieldDecider returns the [FieldBehavior] to use for a single field.
// A custom decider handles the cases it cares about and delegates
// everything else to the decider it composes — typically by reading
// the request-context decider with [DeciderFromContext].
type FieldDecider func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior

// FieldVisit is one yielded item from [ReflectFormFields]: a
// fully-qualified path plus the resolved struct field, its addressable
// value, the [DetectField] kind, and the parsed [FormTag].
type FieldVisit struct {
	Path  FieldPath
	Field reflect.StructField
	Value reflect.Value
	Kind  FieldKind
	Tag   FormTag

	// Section is the nearest enclosing section name (from
	// form:"section=..." or form:"nested" tags). Empty when the field
	// is in the root section.
	Section string
}

// FieldErrors is the structured error type [ReflectFormHandler]'s
// onSubmit callback may return to surface per-field errors inline.
// Keys are field paths (the same value passed to render/parse for
// each field); values are the messages to render under each field.
// FieldErrors implements the error interface via errors.Join.
type FieldErrors map[FieldPath]error

// Error joins every entry's message in deterministic (path-sorted)
// order and returns a single string. It uses errors.Join's formatting
// (newline-separated) when there is more than one entry.
func (e FieldErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	paths := make([]string, 0, len(e))
	for p := range e {
		paths = append(paths, string(p))
	}
	sort.Strings(paths)
	parts := make([]error, 0, len(paths))
	for _, p := range paths {
		parts = append(parts, e[FieldPath(p)])
	}
	return errors.Join(parts...).Error()
}

// Unwrap allows errors.Is/errors.As to inspect each per-field error.
func (e FieldErrors) Unwrap() []error {
	if len(e) == 0 {
		return nil
	}
	parts := make([]error, 0, len(e))
	for _, err := range e {
		parts = append(parts, err)
	}
	return parts
}

// SingleValueTypes is the registry of struct types that the form
// walker MUST treat as a single widget — never as a recursable section
// — regardless of whether they happen to look like Go structs. The
// canonical entry is time.Time: its public fields would otherwise be
// walked into a five-field "section." Callers add their own
// single-value struct types here (currency, address, money amount,
// etc.) so the decider knows to look up a widget instead of recursing.
//
// Types implementing [FormWidgetHint] are treated as single-value
// automatically and do not need to be registered.
var SingleValueTypes = newSingleValueRegistry()

type singleValueRegistry struct {
	mu    sync.RWMutex
	types map[reflect.Type]struct{}
}

func newSingleValueRegistry() *singleValueRegistry {
	r := &singleValueRegistry{types: make(map[reflect.Type]struct{})}
	r.Register(reflect.TypeFor[time.Time]())
	return r
}

// Register marks t as a single-value type. Pointer types are
// normalized to their element type before registration. Calling
// Register with a non-struct type is a no-op.
func (r *singleValueRegistry) Register(t reflect.Type) {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[t] = struct{}{}
}

// Has reports whether t (or its element type, if pointer) is
// registered as a single-value type. Types implementing
// [FormWidgetHint] return true regardless of explicit registration.
func (r *singleValueRegistry) Has(t reflect.Type) bool {
	if t == nil {
		return false
	}
	orig := t
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	hintType := reflect.TypeFor[FormWidgetHint]()
	if orig.Implements(hintType) || reflect.PointerTo(t).Implements(hintType) {
		return true
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.types[t]
	return ok
}

// SetField sets path on the struct rooted at root (a pointer to the
// struct). The path is the hyphen-separated form used by FieldPath.
// Returns an addressable reflect.Value and the resolved StructField.
// Useful when a custom decider needs to round-trip a value into the
// load-then-apply target by path. Returns an error when the path is
// invalid or the field is unexported.
func SetField(root reflect.Value, path FieldPath) (reflect.Value, reflect.StructField, error) {
	for root.Kind() == reflect.Pointer {
		if root.IsNil() {
			return reflect.Value{}, reflect.StructField{}, errs.New("mx.SetField: nil pointer root")
		}
		root = root.Elem()
	}
	if root.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.StructField{}, errs.New("mx.SetField: root is not a struct")
	}
	parts := strings.Split(string(path), "-")
	cur := root
	var field reflect.StructField
	for i, name := range parts {
		if cur.Kind() == reflect.Pointer {
			if cur.IsNil() {
				return reflect.Value{}, reflect.StructField{}, errs.New("mx.SetField: nil pointer in path " + string(path))
			}
			cur = cur.Elem()
		}
		if cur.Kind() != reflect.Struct {
			return reflect.Value{}, reflect.StructField{}, errs.New("mx.SetField: non-struct in path at segment " + parts[i])
		}
		var ok bool
		field, ok = cur.Type().FieldByName(name)
		if !ok {
			return reflect.Value{}, reflect.StructField{}, errs.New("mx.SetField: no field " + name + " in path " + string(path))
		}
		cur = cur.FieldByIndex(field.Index)
	}
	return cur, field, nil
}

// ctxKeyDecider is the unexported request-context key under which
// [Middleware] stores the active [FieldDecider]. It is defined here
// (alongside DeciderFromContext below) so that middleware.go and
// fieldbehavior.go agree on the same key without exporting it.
type ctxKeyDecider struct{}

// DeciderFromContext returns the FieldDecider installed by
// [Middleware], or the unconfigured decider that surfaces a clear
// error when none was installed.
func DeciderFromContext(ctx context.Context) FieldDecider {
	if ctx == nil {
		return unconfiguredDecider
	}
	d, ok := ctx.Value(ctxKeyDecider{}).(FieldDecider)
	if !ok || d == nil {
		return unconfiguredDecider
	}
	return d
}

// unconfiguredDecider is returned by [DeciderFromContext] when no
// [Middleware] wrapped the handler. Every behavior it returns produces
// a clear error pointing the caller at the four wiring options
// (shadcn / hx / html / custom).
var unconfiguredDecider FieldDecider = func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
	const msg = "mx.ReflectFormHandler: no FieldDecider in request context — wrap your handler tree with mx.Middleware(shadcn.FieldDecider) (or hx.FieldDecider, html.FieldDecider, or a custom one)"
	err := errs.New(msg)
	return FieldBehavior{
		Render: func(path FieldPath, field reflect.StructField, value reflect.Value, errs []error) Component {
			return Text(msg)
		},
		Parse: func(path FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error {
			return err
		},
		Validate: func(path FieldPath, field reflect.StructField, value reflect.Value) error {
			return nil
		},
	}
}
