package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
	"github.com/ungerik/go-mx/shadcn"
)

// A Questionnaire steps through htmx round trips: every button posts the
// form, and the server answers with the next step rendered. The gallery is
// static HTML with no server to answer, so each example below is one step of
// the same three-step flow as the server would render it, and
// questionnaireNoServer keeps the buttons from reloading the page. A real
// page leaves it out.
var questionnaireNoServer = html.OnSubmit("event.preventDefault()")

// QuestionnaireDemo renders the first step: a required single choice with
// letter shortcuts (press A, B or C), descriptions, and a free-text answer.
// The first step has no Previous button and no Skip, since the item is
// required.
func QuestionnaireDemo() mx.Component {
	return shadcn.Questionnaire(html.Class("max-w-md"),
		hx.Post("/survey"), hx.Target("this"), hx.Swap(hx.SwapOuterHTML), questionnaireNoServer,
		shadcn.QuestionnaireProgress(1, 3, html.Attrib("aria-label", "Questionnaire progress"), "Question 1 of 3"),
		shadcn.QuestionnaireItem("direction",
			shadcn.QuestionnaireTitle("What should we prototype next?"),
			shadcn.QuestionnaireDescription("Choose a direction or write your own."),
			shadcn.QuestionnaireChoices(shadcn.QuestionnaireShortcutsLetters,
				shadcn.QuestionnaireChoice("", "direction", "delegation",
					html.SpanClass("font-medium", "Delegation"),
					shadcn.QuestionnaireChoiceDescription("Show how work moves to a specialist."),
				),
				shadcn.QuestionnaireChoice("", "direction", "questions",
					html.SpanClass("font-medium", "Question prompts"),
					shadcn.QuestionnaireChoiceDescription("Show choices while the interface waits."),
				),
				shadcn.QuestionnaireChoice("", "direction", "both",
					html.SpanClass("font-medium", "Both together"),
				),
				shadcn.QuestionnaireInput(html.Name("direction-other"),
					html.Attrib("aria-label", "Another answer"), html.Placeholder("Type another answer…")),
			),
		),
		shadcn.QuestionnaireActions(
			shadcn.QuestionnaireNext("", "", "Next"),
		),
	)
}

// QuestionnaireMultiple renders the second step, a checkbox item with number
// shortcuts, as the server answers Next on the first: the stored answer of
// step one rides along in a hidden input, which keeps the flow stateless.
func QuestionnaireMultiple() mx.Component {
	return shadcn.Questionnaire(html.Class("max-w-md"),
		hx.Post("/survey"), hx.Target("this"), hx.Swap(hx.SwapOuterHTML), questionnaireNoServer,
		html.Input(html.Type("hidden"), html.Name("direction"), html.Value("delegation")),
		shadcn.QuestionnaireProgress(2, 3, html.Attrib("aria-label", "Questionnaire progress"), "Question 2 of 3"),
		shadcn.QuestionnaireItem("surfaces",
			shadcn.QuestionnaireTitle("Where should it show up?"),
			shadcn.QuestionnaireDescription("Choose every surface that applies."),
			shadcn.QuestionnaireChoices(shadcn.QuestionnaireShortcutsNumbers,
				shadcn.QuestionnaireChoice(shadcn.QuestionnaireCheckbox, "surfaces", "chat", html.Checked, "Chat"),
				shadcn.QuestionnaireChoice(shadcn.QuestionnaireCheckbox, "surfaces", "editor", html.Checked, "Editor"),
				shadcn.QuestionnaireChoice(shadcn.QuestionnaireCheckbox, "surfaces", "terminal", "Terminal"),
				shadcn.QuestionnaireChoice(shadcn.QuestionnaireCheckbox, "surfaces", "mobile", html.Disabled,
					"Mobile",
					shadcn.QuestionnaireChoiceDescription("Not available yet."),
				),
			),
		),
		shadcn.QuestionnaireActions(
			shadcn.QuestionnairePrevious("", "", "Previous"),
			shadcn.QuestionnaireNext("", "", "Next"),
		),
	)
}

// QuestionnaireSkipSubmit renders the last step: an optional item, so the
// server adds a Skip button, and Submit in place of Next.
func QuestionnaireSkipSubmit() mx.Component {
	return shadcn.Questionnaire(html.Class("max-w-md"),
		hx.Post("/survey"), hx.Target("this"), hx.Swap(hx.SwapOuterHTML), questionnaireNoServer,
		html.Input(html.Type("hidden"), html.Name("direction"), html.Value("delegation")),
		html.Input(html.Type("hidden"), html.Name("surfaces"), html.Value("chat")),
		html.Input(html.Type("hidden"), html.Name("surfaces"), html.Value("editor")),
		shadcn.QuestionnaireProgress(3, 3, html.Attrib("aria-label", "Questionnaire progress"), "Question 3 of 3"),
		shadcn.QuestionnaireItem("detail",
			shadcn.QuestionnaireTitle("How much detail should it include?"),
			shadcn.QuestionnaireDescription("Skip this if you are not sure yet."),
			shadcn.QuestionnaireChoices("",
				shadcn.QuestionnaireChoice("", "detail", "focused", "Focused"),
				shadcn.QuestionnaireChoice("", "detail", "complete", "Complete flow"),
			),
		),
		shadcn.QuestionnaireActions(
			shadcn.QuestionnairePrevious("", "", "Previous"),
			shadcn.QuestionnaireSkip("", "", "Skip"),
			shadcn.QuestionnaireSubmit("", "", "Submit"),
		),
	)
}

// QuestionnaireValidation renders the first step as the server answers Next
// without an answer: the same item with a QuestionnaireError, which marks the
// item and its controls invalid.
func QuestionnaireValidation() mx.Component {
	return shadcn.Questionnaire(html.Class("max-w-md"),
		hx.Post("/survey"), hx.Target("this"), hx.Swap(hx.SwapOuterHTML), questionnaireNoServer,
		shadcn.QuestionnaireProgress(1, 3, html.Attrib("aria-label", "Questionnaire progress"), "Question 1 of 3"),
		shadcn.QuestionnaireItem("direction",
			shadcn.QuestionnaireTitle("What should we prototype next?"),
			shadcn.QuestionnaireChoices(shadcn.QuestionnaireShortcutsLetters,
				shadcn.QuestionnaireChoice("", "direction", "delegation", "Delegation"),
				shadcn.QuestionnaireChoice("", "direction", "questions", "Question prompts"),
				shadcn.QuestionnaireChoice("", "direction", "both", "Both together"),
			),
			shadcn.QuestionnaireError("Choose an answer to continue."),
		),
		shadcn.QuestionnaireActions(
			shadcn.QuestionnaireNext("", "", "Next"),
		),
	)
}
