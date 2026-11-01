package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
)

// questionnaireStep renders a complete step the way a server answers one
// round trip: progress, one item with choices and a free-text answer, and the
// buttons of a middle step.
func questionnaireStep(itemChildren ...any) mx.Component {
	children := append([]any{
		QuestionnaireTitle("What should we prototype next?"),
		QuestionnaireDescription("Choose a direction or write your own."),
		QuestionnaireChoices(QuestionnaireShortcutsLetters,
			QuestionnaireChoice("", "direction", "delegation",
				html.Span(html.Class("font-medium"), "Delegation"),
				QuestionnaireChoiceDescription("Show how work moves to a specialist."),
			),
			QuestionnaireChoice("", "direction", "questions", html.Checked, "Question prompts"),
			QuestionnaireInput(html.Name("direction-other"), html.Attrib("aria-label", "Another answer")),
		),
	}, itemChildren...)
	return Questionnaire(hx.Post("/survey"), hx.Target("this"), hx.Swap(hx.SwapOuterHTML),
		QuestionnaireProgress(2, 3, html.Attrib("aria-label", "Questionnaire progress"), "Question 2 of 3"),
		QuestionnaireItem("direction", children...),
		QuestionnaireActions(
			QuestionnairePrevious("", "", "Previous"),
			QuestionnaireNext("", "", "Next"),
		),
	)
}

func TestQuestionnaireComposition(t *testing.T) {
	out := render(t, questionnaireStep())
	for _, want := range []string{
		`<form data-slot="questionnaire"`, `hx-post="/survey"`, "novalidate",
		`onkeydown="mxQuestionnaireKey(event)"`, "if(!window.mxQuestionnaireKey)",
		`<fieldset data-slot="questionnaire-item"`, `tabindex="-1"`,
		`<legend data-slot="questionnaire-title"`,
		`data-slot="questionnaire-description"`, `id="questionnaire-direction-description"`,
		`aria-describedby="questionnaire-direction-description"`,
		`<input type="hidden" name="questionnaire-item" value="direction"/>`,
		`data-slot="questionnaire-choices"`, `data-shortcuts="letters"`,
		`data-slot="questionnaire-choice"`, `data-slot="questionnaire-input"`,
		`data-slot="questionnaire-progress"`, `role="progressbar"`,
		`data-slot="questionnaire-actions"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The hidden item input follows the content: a <legend> must be the
	// first child of its fieldset to be rendered as the fieldset's caption.
	if strings.Index(out, `name="questionnaire-item"`) < strings.Index(out, "<legend") {
		t.Errorf("hidden item input must come after the legend: %s", out)
	}
	// Class strings are reconstructed from the cn-* style sheet; a leaked
	// cn-* token would silently render unstyled.
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
	// Base UI's data attributes are state the headless primitive sets at
	// runtime; this port has no such runtime, so every one must have been
	// rewritten to a native selector.
	for _, baseUI := range []string{"data-checked", "data-disabled", "data-invalid:", "data-[type=", "data-[shortcut]", "data-open", "data-closed"} {
		if strings.Contains(out, baseUI) {
			t.Errorf("Base UI selector %q should have been rewritten: %s", baseUI, out)
		}
	}
	// A valid step carries no invalid marks.
	for _, unwanted := range []string{`data-invalid=`, `aria-invalid="true"`, `role="alert"`} {
		if strings.Contains(out, unwanted) {
			t.Errorf("valid step must not carry %q: %s", unwanted, out)
		}
	}
}

func TestQuestionnaireKeyHandler(t *testing.T) {
	// A caller onkeydown replaces the handler; the script stays defined.
	out := render(t, Questionnaire(html.OnKeyDown("myKeys(event)"), "x"))
	if !strings.Contains(out, `onkeydown="myKeys(event)"`) || strings.Contains(out, `onkeydown="mxQuestionnaireKey`) {
		t.Errorf("caller onkeydown should replace the default: %s", out)
	}
	if strings.Count(out, "onkeydown=") != 1 {
		t.Errorf("exactly one onkeydown expected: %s", out)
	}
	// The script never presses Previous for Enter: it looks up Next, then
	// Submit, by data-slot.
	for _, want := range []string{"[data-slot=questionnaire-next]", "[data-slot=questionnaire-submit]", "[data-shortcut="} {
		if !strings.Contains(questionnaireKeyScript, want) {
			t.Errorf("key script misses %q", want)
		}
	}
	if strings.Contains(questionnaireKeyScript, "questionnaire-previous") {
		t.Error("key script must not press the Previous button")
	}
}

func TestQuestionnaireProgress(t *testing.T) {
	out := render(t, QuestionnaireProgress(2, 5, html.Attrib("aria-label", "Progress"), "Question 2 of 5"))
	for _, want := range []string{
		`role="progressbar"`, `aria-live="polite"`,
		`aria-valuemin="1"`, `aria-valuemax="5"`, `aria-valuenow="2"`,
		// The visible text, not a percentage, is what a screen reader reads.
		`aria-valuetext="Question 2 of 5"`,
		"tabular-nums", "text-xs", "Question 2 of 5",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}

	clamped := render(t, QuestionnaireProgress(9, 3))
	if !strings.Contains(clamped, `aria-valuenow="3"`) {
		t.Errorf("current must clamp to total: %s", clamped)
	}
	low := render(t, QuestionnaireProgress(0, 3))
	if !strings.Contains(low, `aria-valuenow="1"`) {
		t.Errorf("current must clamp to 1: %s", low)
	}
	none := render(t, QuestionnaireProgress(0, 0))
	if strings.Contains(none, "aria-value") {
		t.Errorf("no total, no values: %s", none)
	}
	// Markup children are not a value text.
	markup := render(t, QuestionnaireProgress(1, 2, html.Span("1"), "/2"))
	if strings.Contains(markup, "aria-valuetext") {
		t.Errorf("aria-valuetext only from a single text child: %s", markup)
	}
}

func TestQuestionnaireItemInvalid(t *testing.T) {
	out := render(t, questionnaireStep(QuestionnaireError("Choose an answer to continue.")))
	for _, want := range []string{
		`data-invalid="true"`,
		`<p data-slot="questionnaire-error" role="alert"`, `id="questionnaire-direction-error"`,
		`aria-describedby="questionnaire-direction-description questionnaire-direction-error"`,
		// The choice cards turn destructive through the item group.
		"group-data-[invalid=true]/questionnaire-item:border-destructive",
		"Choose an answer to continue.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Both radios and the free-text input are announced as invalid.
	if n := strings.Count(out, `aria-invalid="true"`); n != 3 {
		t.Errorf("want aria-invalid on 3 controls, got %d: %s", n, out)
	}
}

func TestQuestionnaireItemCallerAttribs(t *testing.T) {
	out := render(t, QuestionnaireItem("q",
		html.Attrib("aria-describedby", "mine"), html.TabIndex(0), html.AutoFocus,
		QuestionnaireTitle("T"),
		QuestionnaireDescription(html.ID("own-desc"), "D"),
		QuestionnaireError("E"),
	))
	for _, want := range []string{`aria-describedby="mine"`, `tabindex="0"`, "autofocus", `id="own-desc"`, `id="questionnaire-q-error"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, `id="questionnaire-q-description"`) {
		t.Errorf("a caller id must not be replaced: %s", out)
	}

	// Without a caller aria-describedby, a caller id is the one referenced.
	own := render(t, QuestionnaireItem("q", QuestionnaireDescription(html.ID("own-desc"), "D")))
	if !strings.Contains(own, `aria-describedby="own-desc"`) {
		t.Errorf("caller description id should be referenced: %s", own)
	}
}

func TestQuestionnaireItemInvalidName(t *testing.T) {
	defer func(orig bool) { PanicOnInvalidID = orig }(PanicOnInvalidID)
	PanicOnInvalidID = false
	for _, e := range []*mx.Element{
		QuestionnaireItem(`a"b`, "x"),
		QuestionnaireChoice("", "", "v"),
	} {
		if e.Err == nil {
			t.Errorf("invalid name should defer an error: %#v", e)
		}
	}
}

func TestQuestionnaireChoiceRadio(t *testing.T) {
	out := render(t, QuestionnaireChoice("", "plan", "pro",
		html.Class("my-card"), html.ID("plan-pro"), html.Checked, html.Disabled,
		"Pro", QuestionnaireChoiceDescription("Unlimited projects."),
	))
	for _, want := range []string{
		`<label data-slot="questionnaire-choice" data-type="radio"`,
		// The caller class goes to the card, everything else to the input.
		"my-card",
		`<input data-slot="questionnaire-choice-input" id="plan-pro" checked="checked" disabled="disabled" type="radio" name="plan" value="pro"`, "absolute inset-0 z-10 size-full cursor-pointer opacity-0",
		`data-slot="questionnaire-choice-indicator" aria-hidden="true"`, "rounded-full",
		`data-slot="questionnaire-choice-indicator-dot"`,
		`data-slot="questionnaire-choice-label"`,
		`data-slot="questionnaire-choice-description"`, "Unlimited projects.",
		// Base UI's data-checked/data-disabled on the card, rewritten.
		"has-[>input:checked]:border-primary/40", "has-[>input:disabled]:opacity-50",
		"group-has-[>input:checked]/questionnaire-choice:bg-primary",
		"has-[>input:focus-visible]:ring-3",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, `data-slot="questionnaire-choice-indicator-check"`) {
		t.Errorf("a radio has a dot, not a check: %s", out)
	}
	if strings.Contains(out, `<label data-slot="questionnaire-choice" id=`) || strings.Count(out, "my-card") != 1 {
		t.Errorf("id belongs to the input, class to the card only: %s", out)
	}
	// Without shortcuts there is no badge.
	if strings.Contains(out, "questionnaire-choice-shortcut") {
		t.Errorf("no shortcut badge expected: %s", out)
	}
}

func TestQuestionnaireChoiceCheckbox(t *testing.T) {
	out := render(t, QuestionnaireChoice(QuestionnaireCheckbox, "features", "search", "Search"))
	for _, want := range []string{
		`data-type="checkbox"`, `type="checkbox"`, `name="features"`, `value="search"`,
		`data-slot="questionnaire-choice-indicator-check"`, "lucide-check", "rounded-[4px]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "indicator-dot") {
		t.Errorf("a checkbox has a check, not a dot: %s", out)
	}
	// An unknown type resolves to a radio.
	if bad := render(t, QuestionnaireChoice("toggle", "f", "v")); !strings.Contains(bad, `type="radio"`) {
		t.Errorf("unknown choice type should resolve to radio: %s", bad)
	}
}

func TestQuestionnaireChoicesShortcuts(t *testing.T) {
	choices := func(n int) mx.Components {
		values := make([]int, n)
		return mx.ForEach(values, func(int) *mx.Element { return QuestionnaireChoice("", "q", "v", "x") })
	}

	letters := render(t, QuestionnaireChoices(QuestionnaireShortcutsLetters,
		QuestionnaireChoice("", "q", "a", "A choice"),
		choices(2), // through the mx.Components that mx.ForEach returns
	))
	for _, key := range []string{"A", "B", "C"} {
		for _, want := range []string{`data-shortcut="` + key + `"`, `aria-keyshortcuts="` + key + `"`} {
			if !strings.Contains(letters, want) {
				t.Errorf("missing %q in %s", want, letters)
			}
		}
	}
	if n := strings.Count(letters, `data-slot="questionnaire-choice-shortcut" aria-hidden="true"`); n != 3 {
		t.Errorf("want 3 hidden badges, got %d: %s", n, letters)
	}

	numbers := render(t, QuestionnaireChoices(QuestionnaireShortcutsNumbers, choices(11)))
	if !strings.Contains(numbers, `data-shortcut="9"`) || strings.Count(numbers, "data-shortcut=") != 9 {
		t.Errorf("numbers stop at 9: %s", numbers)
	}

	none := render(t, QuestionnaireChoices("", choices(2)))
	if strings.Contains(none, "shortcut") {
		t.Errorf("no shortcuts expected: %s", none)
	}
}

func TestQuestionnaireInput(t *testing.T) {
	out := render(t, QuestionnaireInput(html.Name("other"), html.Class("font-mono"), html.Placeholder("Type…")))
	for _, want := range []string{
		`<div data-slot="questionnaire-input-wrapper"`,
		`<input data-slot="questionnaire-input" name="other" placeholder="Type…" type="text"`,
		"font-mono", "aria-invalid:border-destructive", "focus-visible:ring-3",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if email := render(t, QuestionnaireInput(html.Type("email"))); strings.Contains(email, `type="text"`) {
		t.Errorf("caller type must win: %s", email)
	}
}

func TestQuestionnaireButtons(t *testing.T) {
	tests := []struct {
		el      *mx.Element
		slot    string
		action  QuestionnaireAction
		variant ButtonVariant
		column  string
		enter   bool
	}{
		{QuestionnairePrevious("", "", "Previous"), "questionnaire-previous", QuestionnaireActionPrevious, ButtonOutline, "col-start-1", false},
		{QuestionnaireSkip("", "", "Skip"), "questionnaire-skip", QuestionnaireActionSkip, ButtonOutline, "col-start-2", false},
		{QuestionnaireNext("", "", "Next"), "questionnaire-next", QuestionnaireActionNext, ButtonDefault, "col-start-3", true},
		{QuestionnaireSubmit("", "", "Submit"), "questionnaire-submit", QuestionnaireActionSubmit, ButtonDefault, "col-start-3", true},
	}
	for _, tt := range tests {
		out := render(t, tt.el)
		for _, want := range []string{
			`<button data-slot="` + tt.slot + `"`, `type="submit"`,
			`name="questionnaire-action"`, `value="` + string(tt.action) + `"`,
			`data-variant="` + string(tt.variant) + `"`, `data-size="default"`,
			tt.column, "row-start-1",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q in %s", tt.slot, want, out)
			}
		}
		if got := strings.Contains(out, `aria-keyshortcuts="Enter"`); got != tt.enter {
			t.Errorf("%s: aria-keyshortcuts=Enter is %v, want %v", tt.slot, got, tt.enter)
		}
	}

	ghost := render(t, QuestionnaireSkip(ButtonGhost, SizeSM, html.Value("later"), "Later"))
	for _, want := range []string{`data-variant="ghost"`, `data-size="sm"`, `value="later"`} {
		if !strings.Contains(ghost, want) {
			t.Errorf("missing %q in %s", want, ghost)
		}
	}
}

func TestQuestionnaireActionValidOr(t *testing.T) {
	if got := QuestionnaireAction("skip").ValidOr(QuestionnaireActionNext); got != QuestionnaireActionSkip {
		t.Errorf("skip: got %q", got)
	}
	if got := QuestionnaireAction("").ValidOr(QuestionnaireActionNext); got != QuestionnaireActionNext {
		t.Errorf("empty: got %q", got)
	}
}

func TestQuestionnaireActionsAndError(t *testing.T) {
	actions := render(t, QuestionnaireActions("x"))
	if !strings.Contains(actions, "grid-cols-[minmax(0,1fr)_auto_auto]") {
		t.Errorf("actions grid missing: %s", actions)
	}
	errOut := render(t, QuestionnaireError(html.Role("status"), "Bad"))
	if !strings.Contains(errOut, `role="status"`) || strings.Contains(errOut, `role="alert"`) {
		t.Errorf("caller role must win: %s", errOut)
	}
}
