package mx

import (
	"reflect"
	"testing"
	"time"
)

type colorEnum string

func (colorEnum) EnumStrings() []string { return []string{"red", "green", "blue"} }

// paramEnums has an Enums method with a parameter, which the option
// collector could not call, so it is not an enum.
type paramEnums string

func (paramEnums) Enums(int) []paramEnums { return nil }

type hintedField string

func (hintedField) FormWidget() string { return "email" }

type detectStruct struct {
	Name      string
	Age       int
	Active    bool
	When      time.Time
	Pointer   *string
	Color     colorEnum
	NotEnum   paramEnums
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
		{"NotEnum", FieldKindString},
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

// hintedDate declares itself a date, so its fields need no widget tag.
type hintedDate string

func (hintedDate) FormWidget() string { return "date" }

// internalNote declares the skip widget, so its fields are never rendered.
type internalNote string

func (internalNote) FormWidget() string { return "skip" }

// A FormWidgetHint that decides a field's kind is returned as the tag's
// Widget, so renderers and parsers read one widget. An explicit widget or
// options tag still takes precedence, and the hint does not overwrite it.
func TestDetectField_HintBecomesTagWidget(t *testing.T) {
	type form struct {
		Date      hintedDate
		DatePtr   *hintedDate
		Email     hintedField
		AsText    hintedField `form:"widget=text"`
		AsOptions hintedDate  `form:"options=partners"`
		Note      internalNote
	}
	tests := []struct {
		field      string
		wantKind   FieldKind
		wantWidget string
	}{
		{"Date", FieldKindDateTime, "date"},
		{"DatePtr", FieldKindDateTime, "date"},
		{"Email", FieldKindString, "email"},
		{"AsText", FieldKindString, "text"},
		{"AsOptions", FieldKindEnum, ""},
		// A skip hint skips the field like form:"widget=skip".
		{"Note", FieldKindSkip, "skip"},
	}
	for _, tt := range tests {
		kind, tag := detectKind(t, &form{}, tt.field)
		if kind != tt.wantKind || tag.Widget != tt.wantWidget {
			t.Errorf("DetectField(%s) = %s with widget %q, want %s with widget %q",
				tt.field, kind, tag.Widget, tt.wantKind, tt.wantWidget)
		}
	}
}

// ptrHintedDate declares its widget on the pointer receiver.
type ptrHintedDate string

func (*ptrHintedDate) FormWidget() string { return "date" }

// A pointer-receiver FormWidgetHint reaches the tag like a value-receiver
// one, whether or not the field value is addressable.
func TestDetectField_PointerReceiverHint(t *testing.T) {
	type form struct {
		Due ptrHintedDate
	}
	if kind, tag := detectKind(t, &form{}, "Due"); kind != FieldKindDateTime || tag.Widget != "date" {
		t.Errorf("addressable: %s with widget %q, want datetime with widget \"date\"", kind, tag.Widget)
	}
	v := reflect.ValueOf(form{})
	field, _ := v.Type().FieldByName("Due")
	if kind, tag := DetectField("Due", field, v.Field(0)); kind != FieldKindDateTime || tag.Widget != "date" {
		t.Errorf("not addressable: %s with widget %q, want datetime with widget \"date\"", kind, tag.Widget)
	}
}

// registeredMoney is a registered single-value struct that is no time.
type registeredMoney struct {
	Cents    int
	Currency string
}

// definedStamp is defined as time.Time, without a widget tag.
type definedStamp time.Time

// soleTime embeds a time.Time as its only field; heldTime holds more.
type soleTime struct {
	time.Time
}

type heldTime struct {
	time.Time
	Valid bool
}

// namedTime holds its time in a named field; timePtr embeds a pointer to a
// time. Neither renders nor binds like a time.Time without a widget tag.
type namedTime struct {
	T time.Time
}

type timePtr struct {
	*time.Time
}

// dateOnlyTime embeds a time but writes only its date.
type dateOnlyTime struct {
	time.Time
}

func (d dateOnlyTime) MarshalText() ([]byte, error) { return []byte(d.Format(time.DateOnly)), nil }

// Rule 14 makes a date/time field only of time-like structs: time.Time, a
// type defined as it, or a struct whose only field is an embedded time. A
// registered single-value struct that holds more is a catch-all text input.
// localStamp and localSole are time-like types no test registers, so the
// detection below rests on their shape alone.
type localStamp time.Time

type localSole struct {
	time.Time
}

type embedsLocalSole struct {
	localSole
}

// A slice of a time-like type tagged repeatable would render rows with no
// cells, since a time's fields are unexported, and zero every value on
// save, and no input binds a slice of times, so the field is skipped.
func TestDetectField_RepeatableTimeLikeSlice(t *testing.T) {
	type form struct {
		Stamps []localStamp   `form:"repeatable"`
		Soles  []localSole    `form:"repeatable"`
		Dates  []dateOnlyTime `form:"repeatable"`
	}
	for _, name := range []string{"Stamps", "Soles", "Dates"} {
		if kind, _ := detectKind(t, &form{}, name); kind != FieldKindSkip {
			t.Errorf("DetectField(%s) = %s, want skip", name, kind)
		}
	}
}

// An embedded time-like struct is a date and time field, not fields to
// inline, for a caller that detects it directly; the walker skips embeds.
func TestDetectField_EmbeddedTimeLikeNotInline(t *testing.T) {
	field := reflect.TypeFor[embedsLocalSole]().Field(0)
	if kind, _ := DetectField("localSole", field, reflect.ValueOf(embedsLocalSole{}).Field(0)); kind != FieldKindDateTime {
		t.Errorf("DetectField(embedded localSole) = %s, want datetime", kind)
	}
}

func TestDetectField_TimeLikeStructs(t *testing.T) {
	SingleValueTypes.Register(reflect.TypeFor[registeredMoney]())
	SingleValueTypes.Register(reflect.TypeFor[soleTime]())
	SingleValueTypes.Register(reflect.TypeFor[heldTime]())
	SingleValueTypes.Register(reflect.TypeFor[namedTime]())
	SingleValueTypes.Register(reflect.TypeFor[timePtr]())
	SingleValueTypes.Register(reflect.TypeFor[dateOnlyTime]())
	type form struct {
		When  time.Time
		Stamp definedStamp
		Sole  soleTime
		Held  heldTime
		Named namedTime
		Ptr   timePtr
		Date  dateOnlyTime
		Money registeredMoney
	}
	for _, tt := range []struct {
		field string
		want  FieldKind
	}{
		{"When", FieldKindDateTime},
		{"Stamp", FieldKindDateTime},
		{"Sole", FieldKindDateTime},
		{"Held", FieldKindCatchAll},
		{"Named", FieldKindCatchAll},
		{"Ptr", FieldKindCatchAll},
		// Its own text methods are not time.Time's RFC 3339.
		{"Date", FieldKindCatchAll},
		{"Money", FieldKindCatchAll},
	} {
		if kind, _ := detectKind(t, &form{}, tt.field); kind != tt.want {
			t.Errorf("DetectField(%s) = %s, want %s", tt.field, kind, tt.want)
		}
	}
}
