package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnumAttribs(t *testing.T) {
	// Enum constants are usable directly as element attributes, and a conversion
	// of a dynamic value works too (Dir("rtl")).
	got := Div(DirRTL, SpellCheckFalse, Translate("no"), Class("box")).String()
	want := `<div dir='rtl' spellcheck='false' translate='no' class='box'></div>`
	require.Equal(t, want, got)

	// go-enum generated validation and value listing.
	require.True(t, DirRTL.Valid())
	require.False(t, Dir("bogus").Valid())
	require.Error(t, Dir("bogus").Validate())
	require.Equal(t, []string{"soft", "hard"}, Wrap("").EnumStrings())
}

func TestEnumAttribInvalidValueDefersError(t *testing.T) {
	// An out-of-set value is not silently emitted: AttribValue returns the
	// Validate error, so rendering the enclosing element fails loudly.
	got := Div(Dir("bogus")).String()
	want := `mx.Element.String: invalid value "bogus" for type html.Dir`
	require.Equal(t, want, got)
}

func TestPopoverEnums(t *testing.T) {
	// Enum constants render directly as attributes in the order given.
	require.Equal(t,
		`<button popovertarget='menu' popovertargetaction='show'>Open</button>`,
		Button(PopoverTarget("menu"), PopoverTargetActionShow, "Open").String(),
	)
	require.Equal(t, `<div popover='auto'></div>`, Div(PopoverAuto).String())
	require.Equal(t, `<div popover='hint'></div>`, Div(PopoverHint).String())

	// go-enum generated validation and value listing, checked symmetrically for
	// each type so a type-specific Validate message or value set can't regress.
	require.True(t, PopoverManual.Valid())
	require.False(t, Popover("bogus").Valid())
	require.EqualError(t, Popover("bogus").Validate(), `invalid value "bogus" for type html.Popover`)
	require.Equal(t, []string{"auto", "manual", "hint"}, Popover("").EnumStrings())

	require.True(t, PopoverTargetActionShow.Valid())
	require.False(t, PopoverTargetAction("bogus").Valid())
	require.EqualError(t, PopoverTargetAction("bogus").Validate(), `invalid value "bogus" for type html.PopoverTargetAction`)
	require.Equal(t, []string{"toggle", "show", "hide"}, PopoverTargetAction("").EnumStrings())

	// An out-of-set enum value defers the Validate error to render time.
	require.Equal(t,
		`mx.Element.String: invalid value "bogus" for type html.Popover`,
		Div(Popover("bogus")).String(),
	)
}
