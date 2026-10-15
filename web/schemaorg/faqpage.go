package schemaorg

// https://schema.org/FAQPage
// https://developers.google.com/search/docs/appearance/structured-data/faqpage

// FAQPage is a page of frequently asked questions, each with one answer
// written by the site itself. A page where users post questions and answers
// is a QAPage, which this package does not cover.
//
// https://schema.org/FAQPage
type FAQPage struct {
	WebPage

	// MainEntity are the questions on the page, in page order. Each needs
	// its AcceptedAnswer, and the text of both has to be visible on the page.
	MainEntity []*Question `json:"mainEntity,omitempty"`
}

// NewFAQPage returns an FAQPage with the given title, canonical URL and
// questions.
func NewFAQPage(name, url string, questions ...*Question) *FAQPage {
	return &FAQPage{
		WebPage:    *newWebPage("FAQPage", name, url),
		MainEntity: questions,
	}
}

// Question is one question of an [FAQPage]. Its Name is the full text of the
// question.
//
// https://schema.org/Question
type Question struct {
	Thing

	// AcceptedAnswer is the answer to the question.
	AcceptedAnswer *Answer `json:"acceptedAnswer,omitempty"`
}

// NewQuestion returns a Question with the given question text and its
// accepted answer. The Answer it creates carries no @context: unlike the
// Question it can never be the root of a script, so it would only repeat the
// context it is nested in.
func NewQuestion(question, answer string) *Question {
	q := &Question{
		Thing:          newThing("Question"),
		AcceptedAnswer: &Answer{Thing: Thing{Type: "Answer"}, Text: answer},
	}
	q.Name = question
	return q
}

// Answer is the answer to a [Question].
//
// https://schema.org/Answer
type Answer struct {
	Thing

	// Text is the full text of the answer. Google allows a few HTML tags in
	// it (links, lists, paragraphs, headings); html.JSONLD escapes them so
	// they cannot end the script element early.
	Text string `json:"text,omitempty"`
}
