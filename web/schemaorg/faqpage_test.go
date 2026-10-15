package schemaorg_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestFAQPage(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"FAQPage","name":"FAQ","url":"https://example.com/faq"}`,
		schemaorg.NewFAQPage("FAQ", "https://example.com/faq"),
	)

	tickets := schemaorg.NewQuestion("Where do I get tickets?", `At the <a href="https://example.com/tickets">box office</a>.`)
	tickets.Context = ""
	parking := schemaorg.NewQuestion("Is there parking?", "Yes, in the garage next door.")
	parking.Context = ""

	page := schemaorg.NewFAQPage("FAQ", "https://example.com/faq", tickets, parking)
	page.InLanguage = "en"
	requireGolden(t, "faqpage", page)
}

func TestQuestion(t *testing.T) {
	question := schemaorg.NewQuestion("Is there parking?", "Yes, in the garage next door.")

	// The question is root-ready, its answer can only ever be nested and so
	// does not repeat the context.
	require.Equal(t, schemaorg.Context, question.Context)
	require.Empty(t, question.AcceptedAnswer.Context)
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Question","name":"Is there parking?",`+
			`"acceptedAnswer":{"@type":"Answer","text":"Yes, in the garage next door."}}`,
		question,
	)
	requireGolden(t, "question", question)
}
