package mx

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// Registry entries are process-global and duplicate registration
// panics, so tests register once in init() — never in test bodies,
// which would crash under `go test -count=2`.
func init() {
	RegisterNamedOptions("test-duplicate", tenantOptions)
	RegisterNamedOptions("test-partners", tenantOptions)
	RegisterNamedOptions("test-tag-wins", func(context.Context) ([]NamedOption, error) {
		return []NamedOption{{Name: "From registry", Value: "registry"}}, nil
	})
}

// tenantOptionsKey stands in for whatever a real application uses to
// carry per-request state (tenant, DB handle) through the context.
type tenantOptionsKey struct{}

func tenantOptions(ctx context.Context) ([]NamedOption, error) {
	opts, ok := ctx.Value(tenantOptionsKey{}).([]NamedOption)
	if !ok {
		return nil, errors.New("no tenant options in context")
	}
	return opts, nil
}

// ctxPartnerID implements NamedOptionsContextProvider: the option list
// comes from the request context, not from the type.
type ctxPartnerID string

func (ctxPartnerID) NamedOptionsContext(ctx context.Context) ([]NamedOption, error) {
	return tenantOptions(ctx)
}

// dualProvider implements both the context-aware and the static
// provider interface. The context-aware one wins when it has an
// answer; (nil, nil) falls through to the static one.
type dualProvider string

func (dualProvider) NamedOptionsContext(ctx context.Context) ([]NamedOption, error) {
	opts, _ := ctx.Value(tenantOptionsKey{}).([]NamedOption)
	return opts, nil
}

func (dualProvider) NamedOptions() []NamedOption {
	return []NamedOption{{Name: "Static", Value: "static"}}
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected panic", name)
		}
	}()
	f()
}

func TestRegisterNamedOptions_Panics(t *testing.T) {
	mustPanic(t, "empty name", func() {
		RegisterNamedOptions("", tenantOptions)
	})
	mustPanic(t, "nil provider", func() {
		RegisterNamedOptions("test-nil-provider", nil)
	})
	// "test-duplicate" was registered in init().
	mustPanic(t, "duplicate name", func() {
		RegisterNamedOptions("test-duplicate", tenantOptions)
	})
}

// TestCollectOptions_RegistryIsPerRequest encodes the reason the
// registry exists: the same field renders different options depending
// on the request context, without the field type being involved.
func TestCollectOptions_RegistryIsPerRequest(t *testing.T) {
	tag := ParseFormTagString("options=test-partners")
	fieldType := reflect.TypeFor[string]()

	for _, tenant := range []string{"acme", "globex"} {
		ctx := context.WithValue(context.Background(), tenantOptionsKey{},
			[]NamedOption{{Name: "Partner of " + tenant, Value: tenant + "-1"}})
		opts, err := CollectOptions(ctx, reflect.Value{}, tag, fieldType)
		if err != nil {
			t.Fatalf("CollectOptions: %v", err)
		}
		if len(opts) != 1 || opts[0].Value != tenant+"-1" {
			t.Errorf("tenant %s: got %v", tenant, opts)
		}
	}
}

func TestCollectOptions_UnregisteredName(t *testing.T) {
	tag := ParseFormTagString("options=test-never-registered")
	_, err := CollectOptions(context.Background(), reflect.Value{}, tag, reflect.TypeFor[string]())
	if err == nil {
		t.Fatal("expected error for unregistered options name")
	}
	if !strings.Contains(err.Error(), "test-never-registered") {
		t.Errorf("error should name the missing entry: %v", err)
	}
}

func TestCollectOptions_ContextProvider(t *testing.T) {
	want := []NamedOption{{Name: "P1", Value: "p1"}}
	ctx := context.WithValue(context.Background(), tenantOptionsKey{}, want)
	target := struct{ Partner ctxPartnerID }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	opts, err := CollectOptions(ctx, value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "p1" {
		t.Errorf("got %v, want %v", opts, want)
	}

	// The provider's error must surface, not be swallowed.
	_, err = CollectOptions(context.Background(), value, FormTag{}, value.Type())
	if err == nil {
		t.Fatal("expected provider error to propagate")
	}
}

func TestCollectOptions_TagBeatsInterface(t *testing.T) {
	tag := ParseFormTagString("options=test-tag-wins")
	target := struct{ Partner ctxPartnerID }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	// ctx carries no tenant options, so the interface path would fail —
	// proving the registry entry was used instead.
	opts, err := CollectOptions(context.Background(), value, tag, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "registry" {
		t.Errorf("expected registry options to win over interface, got %v", opts)
	}
}

func TestCollectOptions_ContextBeatsStatic(t *testing.T) {
	target := struct{ Field dualProvider }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	ctx := context.WithValue(context.Background(), tenantOptionsKey{},
		[]NamedOption{{Name: "Dynamic", Value: "dynamic"}})
	opts, err := CollectOptions(ctx, value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "dynamic" {
		t.Errorf("context provider should win over static: %v", opts)
	}

	// (nil, nil) from the context provider falls through to the static
	// NamedOptions.
	opts, err = CollectOptions(context.Background(), value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "static" {
		t.Errorf("expected fallthrough to static options: %v", opts)
	}
}

func TestOptionsNeedContext(t *testing.T) {
	if !OptionsNeedContext(ParseFormTagString("options=x"), reflect.TypeFor[string]()) {
		t.Error("options tag must need context")
	}
	if !OptionsNeedContext(FormTag{}, reflect.TypeFor[ctxPartnerID]()) {
		t.Error("NamedOptionsContextProvider type must need context")
	}
	if OptionsNeedContext(FormTag{}, reflect.TypeFor[dualProvider]()) != true {
		t.Error("dual provider type must need context")
	}
	if OptionsNeedContext(FormTag{}, reflect.TypeFor[colorEnum]()) {
		t.Error("static enum type must not need context")
	}
}

// ptrCtxProvider implements NamedOptionsContextProvider with a POINTER
// receiver — reachable only through the addressable probe.
type ptrCtxProvider struct{ V string }

func (*ptrCtxProvider) NamedOptionsContext(ctx context.Context) ([]NamedOption, error) {
	return tenantOptions(ctx)
}

func TestCollectOptions_PointerReceiverProvider(t *testing.T) {
	want := []NamedOption{{Name: "P1", Value: "p1"}}
	ctx := context.WithValue(context.Background(), tenantOptionsKey{}, want)
	target := struct{ Field ptrCtxProvider }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	opts, err := CollectOptions(ctx, value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "p1" {
		t.Errorf("pointer-receiver provider not reached via addressable probe: %v", opts)
	}
	if !OptionsNeedContext(FormTag{}, reflect.TypeFor[ptrCtxProvider]()) {
		t.Error("OptionsNeedContext must see pointer-receiver implementations via *T")
	}
}

func TestCollectOptions_OptionsProviderStrings(t *testing.T) {
	target := struct{ Field staticOptionsField }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	opts, err := CollectOptions(context.Background(), value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 2 || opts[0].Value != "a" || opts[0].Name != "a" || opts[1].Value != "b" {
		t.Errorf("Options() []string not converted to name=value pairs: %v", opts)
	}
}

func TestCollectOptions_ZeroValueFallbackContextProvider(t *testing.T) {
	want := []NamedOption{{Name: "P1", Value: "p1"}}
	ctx := context.WithValue(context.Background(), tenantOptionsKey{}, want)
	// Invalid value: the enum-set element-type case — only the
	// zero-value fallback can find the provider.
	opts, err := CollectOptions(ctx, reflect.Value{}, FormTag{}, reflect.TypeFor[ctxPartnerID]())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "p1" {
		t.Errorf("zero-value fallback must reach NamedOptionsContextProvider: %v", opts)
	}
}

func TestCollectOptions_NilPointerFieldDoesNotPanic(t *testing.T) {
	want := []NamedOption{{Name: "P1", Value: "p1"}}
	ctx := context.WithValue(context.Background(), tenantOptionsKey{}, want)
	// A nil *ctxPartnerID field: probing must not dispatch the
	// value-receiver method through the nil pointer.
	target := struct{ Partner *ctxPartnerID }{}
	value := reflect.ValueOf(&target).Elem().Field(0)

	opts, err := CollectOptions(ctx, value, FormTag{}, value.Type())
	if err != nil {
		t.Fatalf("CollectOptions: %v", err)
	}
	if len(opts) != 1 || opts[0].Value != "p1" {
		t.Errorf("nil pointer field must probe a zero element instead: %v", opts)
	}
}
