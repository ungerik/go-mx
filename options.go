package mx

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
