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
