package mx

import (
	"reflect"
	"testing"
	"time"
)

type colorEnum string

func (colorEnum) EnumStrings() []string { return []string{"red", "green", "blue"} }

type hintedField string

func (hintedField) FormWidget() string { return "email" }

type detectStruct struct {
	Name      string
	Age       int
	Active    bool
	When      time.Time
	Pointer   *string
	Color     colorEnum
	Email     hintedField
	Features  map[colorEnum]struct{}
	Tags      []string
	Notes     []byte
	NumStream []int
	Hidden    string `form:"hidden"`
	Skipped   string `form:"-"`
	Custom    string `form:"widget=textarea,help=Long description"`
	Section   string `form:"section=Accounting"`
	Anon      struct{ X int }
}

type embeddedStruct struct {
	Inner string
}

type withEmbed struct {
	embeddedStruct
	Top int
}

func detectKind(t *testing.T, target any, fieldName string) (FieldKind, FormTag) {
	t.Helper()
	v := reflect.ValueOf(target).Elem()
	f, ok := v.Type().FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field %q on %T", fieldName, target)
	}
	return DetectField(FieldPath(fieldName), f, v.FieldByName(fieldName))
}

func TestDetectField_BasicKinds(t *testing.T) {
	s := &detectStruct{}
	cases := []struct {
		field string
		want  FieldKind
	}{
		{"Name", FieldKindString},
		{"Age", FieldKindNumber},
		{"Active", FieldKindBool},
		{"When", FieldKindDateTime},
		{"Pointer", FieldKindString},
		{"Color", FieldKindEnum},
		{"Email", FieldKindString}, // hinted as email → still string kind
		{"Features", FieldKindEnumSet},
		{"Tags", FieldKindTextarea},
		{"Notes", FieldKindTextarea},
		{"Hidden", FieldKindHidden},
		{"Skipped", FieldKindSkip},
		{"Custom", FieldKindTextarea},
		{"Section", FieldKindSection},
		{"Anon", FieldKindCatchAll},
	}
	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			got, _ := detectKind(t, s, c.field)
			if got != c.want {
				t.Errorf("DetectField(%s) = %s, want %s", c.field, got, c.want)
			}
		})
	}
}

func TestDetectField_NumStream(t *testing.T) {
	s := &detectStruct{}
	got, _ := detectKind(t, s, "NumStream")
	// []int falls through to CatchAll — no special textarea support.
	if got != FieldKindCatchAll {
		t.Errorf("got %s, want %s", got, FieldKindCatchAll)
	}
}

func TestDetectField_AnonymousEmbed(t *testing.T) {
	v := reflect.ValueOf(&withEmbed{}).Elem()
	f, _ := v.Type().FieldByName("embeddedStruct")
	got, _ := DetectField("embeddedStruct", f, v.FieldByIndex(f.Index))
	if got != FieldKindInline {
		t.Errorf("embedded struct: got %s, want %s", got, FieldKindInline)
	}
}

func TestDetectField_TagParsing(t *testing.T) {
	s := &detectStruct{}
	_, tag := detectKind(t, s, "Custom")
	if tag.Help != "Long description" || tag.Widget != "textarea" {
		t.Errorf("got %+v, want Widget=textarea Help=Long description", tag)
	}
}

func TestSingleValueRegistry_DefaultsTime(t *testing.T) {
	if !SingleValueTypes.Has(reflect.TypeFor[time.Time]()) {
		t.Errorf("time.Time should be registered by default")
	}
	if !SingleValueTypes.Has(reflect.TypeFor[*time.Time]()) {
		t.Errorf("*time.Time should be recognized via pointer normalization")
	}
}

type myStruct struct{ A int }

func (myStruct) FormWidget() string { return "select" }

func TestSingleValueRegistry_FormWidgetHintAutoRegisters(t *testing.T) {
	if !SingleValueTypes.Has(reflect.TypeFor[myStruct]()) {
		t.Errorf("FormWidgetHint impls should be recognized as single-value")
	}
}

func TestSingleValueRegistry_Register(t *testing.T) {
	type customSingle struct{ A int }
	if SingleValueTypes.Has(reflect.TypeFor[customSingle]()) {
		// The registry is process-global; under `go test -count=2` a
		// previous run in this process already registered the type.
		t.Skip("already registered by a previous run in this process")
	}
	SingleValueTypes.Register(reflect.TypeFor[customSingle]())
	if !SingleValueTypes.Has(reflect.TypeFor[customSingle]()) {
		t.Errorf("after Register, Has should be true")
	}
}

func TestImplementsTextMarshaler(t *testing.T) {
	if !ImplementsTextMarshaler(reflect.TypeFor[time.Time]()) {
		t.Errorf("time.Time implements TextMarshaler (value receiver)")
	}
	if ImplementsTextMarshaler(reflect.TypeFor[int]()) {
		t.Errorf("int does not implement TextMarshaler")
	}
}

func TestImplementsTextUnmarshaler(t *testing.T) {
	if !ImplementsTextUnmarshaler(reflect.TypeFor[time.Time]()) {
		t.Errorf("time.Time implements TextUnmarshaler via pointer receiver")
	}
	if ImplementsTextUnmarshaler(reflect.TypeFor[int]()) {
		t.Errorf("int does not implement TextUnmarshaler")
	}
}

type staticOptionsField string

func (staticOptionsField) Options() []string { return []string{"a", "b"} }

func TestDetectField_OptionsTagAndProviderInterfaces(t *testing.T) {
	type optionsForm struct {
		Partner    string              `form:"options=partners"`
		PartnerSet []string            `form:"options=partners"`
		PartnerMap map[string]struct{} `form:"options=partners"`
		WidgetWins string              `form:"widget=text,options=partners"`
		Static     staticOptionsField
		StaticSet  []staticOptionsField
		CtxBased   ctxPartnerID
		CtxSet     []ctxPartnerID
	}
	s := &optionsForm{}
	cases := []struct {
		field string
		want  FieldKind
	}{
		{"Partner", FieldKindEnum},
		{"PartnerSet", FieldKindEnumSet},
		{"PartnerMap", FieldKindEnumSet},
		{"WidgetWins", FieldKindString}, // explicit widget beats options tag
		{"Static", FieldKindEnum},
		{"StaticSet", FieldKindEnumSet},
		{"CtxBased", FieldKindEnum},
		{"CtxSet", FieldKindEnumSet},
	}
	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			got, _ := detectKind(t, s, c.field)
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}

func TestDetectField_ProviderInterfaceMapKey(t *testing.T) {
	type mapForm struct {
		CtxFeatures map[ctxPartnerID]struct{}
	}
	got, _ := detectKind(t, &mapForm{}, "CtxFeatures")
	if got != FieldKindEnumSet {
		t.Errorf("map with provider-interface key: got %s, want %s", got, FieldKindEnumSet)
	}
}

func TestDetectField_OptionsTagShapeEdges(t *testing.T) {
	type edgeForm struct {
		Bytes    []byte            `form:"options=partners"`
		BadMap   map[string]string `form:"options=partners"`
		Hinted   hintedField       `form:"options=partners"`
		PtrProv  *ctxPartnerID
		AnyField any
	}
	s := &edgeForm{}
	cases := []struct {
		field string
		want  FieldKind
	}{
		{"Bytes", FieldKindTextarea},    // tag ignored on []byte
		{"BadMap", FieldKindCatchAll},   // tag ignored on non-set maps
		{"Hinted", FieldKindEnum},       // explicit options tag beats FormWidgetHint
		{"PtrProv", FieldKindEnum},      // provider interface through pointer type
		{"AnyField", FieldKindCatchAll}, // interface field must not panic in derefType
	}
	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			got, _ := detectKind(t, s, c.field)
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}
