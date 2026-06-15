package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/xml"
)

func TestComment(t *testing.T) {
	require.Equal(t, `<!-- hello -->`, render(t, xml.Comment("hello")))

	// Comments must not contain the "--" sequence.
	_, err := xml.String(xml.Comment("bad -- comment"))
	require.Error(t, err)
}
