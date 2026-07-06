package mx

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/domonda/go-errs"
)

// OptionsProvider is implemented by types that offer a flat list of
// option values. The [ReflectFormHandler] uses it to render dropdowns
// for types that do not (or cannot) implement the
// Enums()/EnumStrings() convention.
type OptionsProvider interface {
	Options() []string
}

// NamedOption is one entry in a [NamedOptionsProvider]'s option list.
// Name is the human-readable label; Value is what the form submits.
type NamedOption struct {
	Name  string
	Value string
}

// NamedOptionsProvider is implemented by types that offer a list of
// labeled option values for dropdowns and radio groups.
type NamedOptionsProvider interface {
	NamedOptions() []NamedOption
}

// NamedOptionsContextProvider is the context-aware sibling of
// [NamedOptionsProvider] for option lists that depend on the request
// (tenant-scoped database queries, per-user permissions, …). It is
// probed before the static provider interfaces and called at render
// time with the context passed to [Component] Render — for forms built
// by [ReflectFormHandler] that is the HTTP request context.
//
// Returning (nil, nil) means "no options from this receiver" and
// collection falls through to the static conventions; a non-nil error
// aborts the render.
type NamedOptionsContextProvider interface {
	NamedOptionsContext(ctx context.Context) ([]NamedOption, error)
}

// NamedOptionsFunc produces an option list from a context. It is the
// function form of [NamedOptionsContextProvider], used with
// [RegisterNamedOptions].
type NamedOptionsFunc func(ctx context.Context) ([]NamedOption, error)

var (
	namedOptionsMtx      sync.RWMutex
	namedOptionsRegistry = map[string]NamedOptionsFunc{}
)

// RegisterNamedOptions registers provider under name for fields tagged
// form:"options=name". The registry exists for field types that cannot
// implement one of the options-provider interfaces themselves (foreign
// ID types, plain strings). Register once at program initialization:
// an empty name, a nil provider, or a duplicate name panics.
func RegisterNamedOptions(name string, provider NamedOptionsFunc) {
	if name == "" {
		panic("mx.RegisterNamedOptions: empty name")
	}
	if provider == nil {
		panic("mx.RegisterNamedOptions: nil provider for name " + name)
	}
	namedOptionsMtx.Lock()
	defer namedOptionsMtx.Unlock()
	if _, exists := namedOptionsRegistry[name]; exists {
		panic("mx.RegisterNamedOptions: name already registered: " + name)
	}
	namedOptionsRegistry[name] = provider
}

func lookupNamedOptions(name string) NamedOptionsFunc {
	namedOptionsMtx.RLock()
	defer namedOptionsMtx.RUnlock()
	return namedOptionsRegistry[name]
}

var (
	namedOptionsContextProviderType = reflect.TypeFor[NamedOptionsContextProvider]()
	namedOptionsProviderType        = reflect.TypeFor[NamedOptionsProvider]()
	optionsProviderType             = reflect.TypeFor[OptionsProvider]()
)

// OptionsNeedContext reports whether the option list for a field with
// the given tag and type can only be resolved with a real request
// context: the tag names a [RegisterNamedOptions] entry, the type
// implements [NamedOptionsContextProvider], or the type is an
// interface (whose dynamic value may implement it). Renderers use it
// to decide between collecting options at component-build time and
// deferring [CollectOptions] into render time.
func OptionsNeedContext(tag FormTag, t reflect.Type) bool {
	if tag.Options != "" {
		return true
	}
	if t == nil {
		return false
	}
	return t.Kind() == reflect.Interface ||
		typeOrPointerImplements(t, namedOptionsContextProviderType)
}

// hasOptionsProviderInterfaces reports whether t (or *t) implements
// one of the options-provider interfaces. It is the interface-based
// sibling of hasEnumMethods in field detection.
func hasOptionsProviderInterfaces(t reflect.Type) bool {
	return typeOrPointerImplements(t, namedOptionsContextProviderType) ||
		typeOrPointerImplements(t, namedOptionsProviderType) ||
		typeOrPointerImplements(t, optionsProviderType)
}

func typeOrPointerImplements(t, iface reflect.Type) bool {
	if t == nil {
		return false
	}
	if t.Implements(iface) {
		return true
	}
	return t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(iface)
}

// CollectOptions resolves the option list for an option-driven form
// field (FieldKindEnum / FieldKindEnumSet). Resolution order:
//
//  1. tag.Options names a provider registered with
//     [RegisterNamedOptions]; an unregistered name is an error.
//  2. value implements [NamedOptionsContextProvider],
//     [NamedOptionsProvider], [OptionsProvider], or has an
//     Enums()/EnumStrings() method — probed in that order, with the
//     addressable interface tried first so pointer-receiver methods
//     are reachable.
//  3. The same probes on a zero value of t, so value-independent
//     providers work when no live value exists (enum-set element
//     types).
//
// value may be invalid; t must be the field (or set-element) type.
// Renderers defer this call into render time whenever
// [OptionsNeedContext] reports true, so ctx is the context passed to
// [Component] Render — the HTTP request context in
// [ReflectFormHandler] forms. Static option lists may be collected at
// component-build time with a background context.
func CollectOptions(ctx context.Context, value reflect.Value, tag FormTag, t reflect.Type) ([]NamedOption, error) {
	if tag.Options != "" {
		provider := lookupNamedOptions(tag.Options)
		if provider == nil {
			return nil, errs.Errorf("no options provider registered for name %q — call mx.RegisterNamedOptions", tag.Options)
		}
		return provider(ctx)
	}
	if value.IsValid() {
		opts, err := probeOptionsProviders(ctx, value)
		if opts != nil || err != nil {
			return opts, err
		}
	}
	// Zero-value fallback: reflect.New(t).Elem() is addressable, so
	// pointer-receiver methods are reachable through the probe too.
	return probeOptionsProviders(ctx, reflect.New(t).Elem())
}

// probeOptionsProviders probes value for any of the option-list
// conventions and returns the unified [NamedOption] list. Returns
// (nil, nil) when none match.
func probeOptionsProviders(ctx context.Context, value reflect.Value) ([]NamedOption, error) {
	if !value.IsValid() {
		return nil, nil
	}
	// Never invoke provider methods through a nil pointer: value-receiver
	// methods panic on dispatch and pointer-receiver methods would see
	// receiver state that doesn't exist. Probe an addressable zero
	// element instead.
	for value.Kind() == reflect.Pointer && value.IsNil() {
		value = reflect.New(value.Type().Elem()).Elem()
	}
	probe := func(iface any) ([]NamedOption, error) {
		if iface == nil {
			return nil, nil
		}
		if cp, ok := iface.(NamedOptionsContextProvider); ok {
			opts, err := cp.NamedOptionsContext(ctx)
			if opts != nil || err != nil {
				return opts, err
			}
		}
		if np, ok := iface.(NamedOptionsProvider); ok {
			return np.NamedOptions(), nil
		}
		if op, ok := iface.(OptionsProvider); ok {
			return optionsToNamed(op.Options()), nil
		}
		// Reflective check for Enums()/EnumStrings(). Only niladic
		// methods are callable here — a signature with parameters
		// would panic in Call.
		v := reflect.ValueOf(iface)
		if v.IsValid() {
			if m := v.MethodByName("EnumStrings"); m.IsValid() && m.Type().NumIn() == 0 {
				ret := m.Call(nil)
				if len(ret) == 1 && ret[0].Kind() == reflect.Slice && ret[0].Type().Elem().Kind() == reflect.String {
					return optionsToNamed(stringsFromValue(ret[0])), nil
				}
			}
			if m := v.MethodByName("Enums"); m.IsValid() && m.Type().NumIn() == 0 {
				ret := m.Call(nil)
				if len(ret) == 1 && ret[0].Kind() == reflect.Slice {
					return enumsToNamed(ret[0]), nil
				}
			}
		}
		return nil, nil
	}
	// Try the addressable interface first so pointer-receiver methods
	// are reachable too.
	if value.CanAddr() {
		if opts, err := probe(value.Addr().Interface()); opts != nil || err != nil {
			return opts, err
		}
	}
	if value.CanInterface() {
		if opts, err := probe(value.Interface()); opts != nil || err != nil {
			return opts, err
		}
	}
	// No fresh-zero-value fallback here: [CollectOptions] follows a
	// nil result with a probe of reflect.New(t).Elem(), which covers
	// receiver-state-independent methods.
	return nil, nil
}

func optionsToNamed(opts []string) []NamedOption {
	out := make([]NamedOption, len(opts))
	for i, o := range opts {
		out[i] = NamedOption{Name: o, Value: o}
	}
	return out
}

func stringsFromValue(v reflect.Value) []string {
	out := make([]string, v.Len())
	for i := range v.Len() {
		out[i] = v.Index(i).String()
	}
	return out
}

func enumsToNamed(v reflect.Value) []NamedOption {
	out := make([]NamedOption, v.Len())
	for i := range v.Len() {
		s := fmt.Sprint(v.Index(i).Interface())
		out[i] = NamedOption{Name: s, Value: s}
	}
	return out
}
