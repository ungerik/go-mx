//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Questionnaire is a Go port of shadcn/ui's Questionnaire (added upstream
// 2026-08-05, shadcn-ui/ui#11414): a form that asks one question at a time,
// with choice cards, a free-text answer, progress and previous / skip / next /
// submit buttons. Class strings are reconstructed from questionnaire.tsx plus
// the style-vega.css .cn-questionnaire-* rules — see "Upstream restructure
// (July 2026)" in TODOS.md. Upstream's cn-font-heading on the title is dropped
// (a theme font variable this port does not define, as in empty.go).
//
// Upstream wraps a headless React primitive (@shadcn/react/questionnaire) that
// holds every item in the page and keeps the step, the answers and the
// validation in client state. This port moves all of that to the server: each
// step is a round trip, so the page only ever holds the active item and no
// client state exists to get out of sync.
//
// # Round trip
//
// A Questionnaire is a <form>. Give it hx.Post(url), hx.Target("this") and
// hx.Swap(hx.SwapOuterHTML) to step through htmx, or html.Method("post") and
// html.Action(url) to step through full page loads; the contract is the same.
// Pressing one of the navigation buttons posts:
//
//   - the active item's answers under the names the caller gave the
//     [QuestionnaireChoice] and [QuestionnaireInput] controls (a checkbox item
//     posts one value per checked choice);
//   - [QuestionnaireItemParam] ("questionnaire-item") with the name of the
//     active [QuestionnaireItem], from a hidden input the item renders;
//   - [QuestionnaireActionParam] ("questionnaire-action") with the
//     [QuestionnaireAction] of the button pressed: "previous", "next", "skip"
//     or "submit". Enter in the form presses the Next or else the Submit
//     button (see the keyboard section below).
//
// The server answers with the whole Questionnaire re-rendered:
//
//   - next / submit with an invalid answer: the same item again, with a
//     [QuestionnaireError] holding the message. The item then marks itself
//     and its controls invalid (see [QuestionnaireItem]);
//   - next with a valid answer: the following item, its progress and its
//     buttons; submit with a valid answer: whatever the flow ends with;
//   - previous: the preceding item with its stored answer pre-selected
//     (html.Checked on the choice, html.Value on the input);
//   - skip: the following item, with the active one recorded as unanswered.
//     The posted answers of a skipped item are to be ignored.
//
// Only the active item's answers are posted. Keep the earlier ones on the
// server (a session) or carry them in hidden inputs inside the Questionnaire,
// which keeps the flow stateless. The server decides which buttons a step
// shows: no [QuestionnairePrevious] on the first item, [QuestionnaireSkip] only
// on an optional one, [QuestionnaireSubmit] instead of [QuestionnaireNext] on
// the last. Each button sits in its own column of [QuestionnaireActions], so
// leaving one out does not move the others.
//
// The form carries novalidate, as upstream's does by default: validation is
// the server's, so its messages look like every other [QuestionnaireError]
// instead of a browser bubble.
//
// To move the focus to the new step after an htmx swap, render html.AutoFocus
// on its [QuestionnaireItem] (focusable through tabindex="-1", like
// upstream's): htmx focuses an autofocus element of the content it swaps in.
// Leave it off the first step of a page, where autofocus would scroll to the
// form on page load.
//
// # Keyboard
//
// One small inline script, defined once per page however many
// questionnaires the page holds, replaces upstream's key handling: Enter (or
// Ctrl/Cmd+Enter in a textarea) presses the Next button, or Submit where there
// is no Next, so Enter never posts "previous" by being the first submit button
// of the form; and with [QuestionnaireChoices] shortcuts a letter or number
// key checks the choice it labels, unless the focus is in a text field. A
// caller-supplied onkeydown on the Questionnaire replaces the handler.
//
// # Built-in text
//
// Upstream's parts fall back to English captions ("Previous", "Question 1 of
// 3", "Choose an answer to continue."). Here every caption is the caller's
// children, so that a translated page needs no extra wiring: pass the button
// labels, the progress text and the error message as children.

// QuestionnaireItemParam is the form parameter that carries the name of the
// active [QuestionnaireItem].
const QuestionnaireItemParam = "questionnaire-item"

// QuestionnaireActionParam is the form parameter that carries the
// [QuestionnaireAction] of the navigation button pressed.
const QuestionnaireActionParam = "questionnaire-action"

// QuestionnaireAction is the step a navigation button asks the server for,
// posted as [QuestionnaireActionParam]. Read it with
// QuestionnaireAction(r.FormValue(shadcn.QuestionnaireActionParam)) and treat
// an invalid value with ValidOr, e.g. as [QuestionnaireActionNext].
type QuestionnaireAction string //#enum

const (
	// QuestionnaireActionPrevious goes back to the preceding item.
	QuestionnaireActionPrevious QuestionnaireAction = "previous"
	// QuestionnaireActionNext validates the active item and goes on.
	QuestionnaireActionNext QuestionnaireAction = "next"
	// QuestionnaireActionSkip leaves an optional item unanswered and goes on.
	QuestionnaireActionSkip QuestionnaireAction = "skip"
	// QuestionnaireActionSubmit validates the last item and finishes.
	QuestionnaireActionSubmit QuestionnaireAction = "submit"
)

// Valid indicates if q is any of the valid values for QuestionnaireAction
func (q QuestionnaireAction) Valid() bool {
	switch q {
	case
		QuestionnaireActionPrevious,
		QuestionnaireActionNext,
		QuestionnaireActionSkip,
		QuestionnaireActionSubmit:
		return true
	}
	return false
}

// Validate returns an error if q is none of the valid values for QuestionnaireAction
func (q QuestionnaireAction) Validate() error {
	if !q.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.QuestionnaireAction", q)
	}
	return nil
}

// ValidOr returns q if it is any of the valid values for QuestionnaireAction, else defaultVal
func (q QuestionnaireAction) ValidOr(defaultVal QuestionnaireAction) QuestionnaireAction {
	if q.Valid() {
		return q
	}
	return defaultVal
}

// Enums returns all valid values for QuestionnaireAction
func (QuestionnaireAction) Enums() []QuestionnaireAction {
	return []QuestionnaireAction{
		QuestionnaireActionPrevious,
		QuestionnaireActionNext,
		QuestionnaireActionSkip,
		QuestionnaireActionSubmit,
	}
}

// EnumStrings returns all valid values for QuestionnaireAction as strings
func (QuestionnaireAction) EnumStrings() []string {
	return []string{
		"previous",
		"next",
		"skip",
		"submit",
	}
}

// String implements the fmt.Stringer interface for QuestionnaireAction
func (q QuestionnaireAction) String() string {
	return string(q)
}

// QuestionnaireChoiceType selects whether a [QuestionnaireChoice] is one of
// several exclusive answers (a radio) or one of several combinable ones (a
// checkbox) — upstream's multiple prop on the item.
type QuestionnaireChoiceType string //#enum

const (
	// QuestionnaireRadio is an exclusive choice (the default).
	QuestionnaireRadio QuestionnaireChoiceType = "radio"
	// QuestionnaireCheckbox is a choice that combines with the others.
	QuestionnaireCheckbox QuestionnaireChoiceType = "checkbox"
)

// Valid indicates if q is any of the valid values for QuestionnaireChoiceType
func (q QuestionnaireChoiceType) Valid() bool {
	switch q {
	case
		QuestionnaireRadio,
		QuestionnaireCheckbox:
		return true
	}
	return false
}

// Validate returns an error if q is none of the valid values for QuestionnaireChoiceType
func (q QuestionnaireChoiceType) Validate() error {
	if !q.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.QuestionnaireChoiceType", q)
	}
	return nil
}

// ValidOr returns q if it is any of the valid values for QuestionnaireChoiceType, else defaultVal
func (q QuestionnaireChoiceType) ValidOr(defaultVal QuestionnaireChoiceType) QuestionnaireChoiceType {
	if q.Valid() {
		return q
	}
	return defaultVal
}

// Enums returns all valid values for QuestionnaireChoiceType
func (QuestionnaireChoiceType) Enums() []QuestionnaireChoiceType {
	return []QuestionnaireChoiceType{
		QuestionnaireRadio,
		QuestionnaireCheckbox,
	}
}

// EnumStrings returns all valid values for QuestionnaireChoiceType as strings
func (QuestionnaireChoiceType) EnumStrings() []string {
	return []string{
		"radio",
		"checkbox",
	}
}

// String implements the fmt.Stringer interface for QuestionnaireChoiceType
func (q QuestionnaireChoiceType) String() string {
	return string(q)
}

// QuestionnaireShortcuts selects the keys [QuestionnaireChoices] assigns to
// its choices. "" assigns none.
type QuestionnaireShortcuts string //#enum

const (
	// QuestionnaireShortcutsLetters assigns A to Z.
	QuestionnaireShortcutsLetters QuestionnaireShortcuts = "letters"
	// QuestionnaireShortcutsNumbers assigns 1 to 9.
	QuestionnaireShortcutsNumbers QuestionnaireShortcuts = "numbers"
)

// Valid indicates if q is any of the valid values for QuestionnaireShortcuts
func (q QuestionnaireShortcuts) Valid() bool {
	switch q {
	case
		QuestionnaireShortcutsLetters,
		QuestionnaireShortcutsNumbers:
		return true
	}
	return false
}

// Validate returns an error if q is none of the valid values for QuestionnaireShortcuts
func (q QuestionnaireShortcuts) Validate() error {
	if !q.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.QuestionnaireShortcuts", q)
	}
	return nil
}

// ValidOr returns q if it is any of the valid values for QuestionnaireShortcuts, else defaultVal
func (q QuestionnaireShortcuts) ValidOr(defaultVal QuestionnaireShortcuts) QuestionnaireShortcuts {
	if q.Valid() {
		return q
	}
	return defaultVal
}

// Enums returns all valid values for QuestionnaireShortcuts
func (QuestionnaireShortcuts) Enums() []QuestionnaireShortcuts {
	return []QuestionnaireShortcuts{
		QuestionnaireShortcutsLetters,
		QuestionnaireShortcutsNumbers,
	}
}

// EnumStrings returns all valid values for QuestionnaireShortcuts as strings
func (QuestionnaireShortcuts) EnumStrings() []string {
	return []string{
		"letters",
		"numbers",
	}
}

// String implements the fmt.Stringer interface for QuestionnaireShortcuts
func (q QuestionnaireShortcuts) String() string {
	return string(q)
}

// questionnaireKeyScript is the once-defined key handler of [Questionnaire],
// called from the form's onkeydown. Enter presses the visible Next button, or
// Submit, instead of the form's first submit button (which would be
// Previous); a shortcut key focuses and clicks the choice input whose card
// carries that data-shortcut. Like upstream it leaves composing input,
// modified keys and text fields alone.
const questionnaireKeyScript = /*js*/ `if(!window.mxQuestionnaireKey){window.mxQuestionnaireKey=function(e){var f=e.currentTarget,t=e.target,k=e.key;if(e.defaultPrevented||e.isComposing||e.keyCode===229||e.altKey||!t||!t.tagName||!k)return;var g=t.tagName,y=(t.type||'').toLowerCase(),m=e.metaKey||e.ctrlKey;if(k==='Enter'){if(e.shiftKey||g==='BUTTON'||g==='A'||g==='SELECT'||(!m&&(g==='TEXTAREA'||t.isContentEditable)))return;var b=f.querySelector('[data-slot=questionnaire-next]:not(:disabled)')||f.querySelector('[data-slot=questionnaire-submit]:not(:disabled)');if(!b)return;e.preventDefault();if(!e.repeat)b.click();return;}if(m||k.length!==1||g==='TEXTAREA'||g==='SELECT'||t.isContentEditable||(g==='INPUT'&&!/^(checkbox|radio|button|submit|reset)$/.test(y)))return;k=k.toUpperCase();if(!/^[A-Z1-9]$/.test(k))return;var c=f.querySelector('[data-slot=questionnaire-choice][data-shortcut="'+k+'"]>input:not(:disabled)');if(!c)return;e.preventDefault();if(!e.repeat){c.focus();c.click();}};}`

// Questionnaire renders the questionnaire as a <form novalidate>: the
// container of a [QuestionnaireProgress], the active [QuestionnaireItem] and
// the [QuestionnaireActions]. See the package section above for the round
// trip it takes part in; the htmx or form attributes that send it go here.
func Questionnaire(attribsChildren ...any) *mx.Element {
	e := html.Form(append(attribsChildren, html.ScriptJS(questionnaireKeyScript))...)
	if e.AttribIndex("novalidate") < 0 {
		e.Attribs = append(e.Attribs, html.NoValidate)
	}
	if e.AttribIndex("onkeydown") < 0 {
		e.Attribs = append(e.Attribs, html.OnKeyDown("mxQuestionnaireKey(event)"))
	}
	return finish(e, "questionnaire", "flex w-full min-w-0 flex-col gap-6")
}

// QuestionnaireProgress renders the position of the active item as a <div
// role="progressbar"> with aria-valuenow current out of total (both clamped
// so that 1 <= current <= total; a total below 1 omits the values). The
// children are the visible text, e.g. "Question 2 of 5"; a single text child
// also becomes the aria-valuetext, so a screen reader reads the same words
// instead of a percentage. Give it an accessible name with an aria-label, as
// upstream does ("Questionnaire progress"): a progressbar takes no name from
// its content.
func QuestionnaireProgress(current, total int, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("progressbar"))
	}
	if e.AttribIndex("aria-live") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-live", "polite"))
	}
	if total >= 1 {
		current = min(max(current, 1), total)
		if e.AttribIndex("aria-valuemin") < 0 {
			e.Attribs = append(e.Attribs, html.Attrib("aria-valuemin", "1"))
		}
		if e.AttribIndex("aria-valuemax") < 0 {
			e.Attribs = append(e.Attribs, html.Attrib("aria-valuemax", strconv.Itoa(total)))
		}
		if e.AttribIndex("aria-valuenow") < 0 {
			e.Attribs = append(e.Attribs, html.Attrib("aria-valuenow", strconv.Itoa(current)))
		}
	}
	if e.AttribIndex("aria-valuetext") < 0 && len(e.Children) == 1 {
		if text, ok := e.Children[0].(mx.Text); ok {
			e.Attribs = append(e.Attribs, html.Attrib("aria-valuetext", string(text)))
		}
	}
	return finish(e, "questionnaire-progress",
		"min-h-[1lh] w-fit min-w-[14ch] font-medium text-muted-foreground tabular-nums text-xs")
}

// QuestionnaireItem renders one question as a <fieldset>, titled by its
// [QuestionnaireTitle] (the <legend>). name identifies the item: it is posted
// as [QuestionnaireItemParam] from a hidden input appended to the fieldset,
// and is normally also the name of the item's answer controls. It must
// contain only letters, digits, '-' and '_'; an invalid name is reported per
// [PanicOnInvalidID].
//
// The item wires what upstream's primitive wires at runtime, at construction:
//
//   - a [QuestionnaireDescription] and a [QuestionnaireError] among its
//     descendants get the ids questionnaire-{name}-description and
//     questionnaire-{name}-error (unless they carry an id) and become the
//     fieldset's aria-describedby;
//   - a [QuestionnaireError] marks the item invalid: the fieldset gets
//     data-invalid="true", which turns the choice cards' borders destructive,
//     and every [QuestionnaireChoice] and [QuestionnaireInput] control gets
//     aria-invalid="true". Render the error only for an invalid answer.
//
// Caller-supplied aria-describedby, data-invalid and tabindex attributes are
// left untouched. Pass html.Disabled to disable the whole item, html.AutoFocus
// to focus it after an htmx swap.
func QuestionnaireItem(name string, attribsChildren ...any) *mx.Element {
	if err := validateID(name); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.FieldSet(attribsChildren...)

	var (
		describedBy           []string
		description, errorMsg *mx.Element
	)
	walkQuestionnaireElements(e.Children, func(c *mx.Element) {
		switch slotOf(c) {
		case "questionnaire-description":
			if description == nil {
				description = c
			}
		case "questionnaire-error":
			if errorMsg == nil {
				errorMsg = c
			}
		}
	})
	if description != nil {
		describedBy = appendElementID(describedBy, description, "questionnaire-"+name+"-description")
	}
	if errorMsg != nil {
		describedBy = appendElementID(describedBy, errorMsg, "questionnaire-"+name+"-error")
		if e.AttribIndex("data-invalid") < 0 {
			e.Attribs = append(e.Attribs, html.DataAttr("invalid", "true"))
		}
		walkQuestionnaireElements(e.Children, func(c *mx.Element) {
			switch slotOf(c) {
			case "questionnaire-choice-input", "questionnaire-input":
				if c.AttribIndex("aria-invalid") < 0 {
					c.Attribs = append(c.Attribs, html.Attrib("aria-invalid", "true"))
				}
			}
		})
	}
	if len(describedBy) > 0 && e.AttribIndex("aria-describedby") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-describedby", strings.Join(describedBy, " ")))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.TabIndex(-1))
	}
	// After the content: a <legend> must be the fieldset's first child.
	e.Children = append(e.Children,
		html.Input(html.Type("hidden"), html.Name(QuestionnaireItemParam), html.Value(name)))
	return finish(e, "questionnaire-item",
		"group/questionnaire-item min-w-0 border-0 p-0 outline-none flex flex-col gap-5")
}

// QuestionnaireTitle renders an item's question as the <legend> of its
// [QuestionnaireItem], which makes it the accessible name of the item and of
// its group of choices. Pass it as the item's first child.
func QuestionnaireTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Legend(attribsChildren...), "questionnaire-title",
		"text-pretty text-base font-semibold [&:not(:has(~[data-slot=questionnaire-description]))]:mb-5")
}

// QuestionnaireDescription renders helper text under a [QuestionnaireTitle]
// as a <p>; the [QuestionnaireItem] around it references it from
// aria-describedby.
func QuestionnaireDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "questionnaire-description",
		"text-pretty text-muted-foreground text-sm")
}

// QuestionnaireChoices stacks an item's [QuestionnaireChoice] cards and its
// optional [QuestionnaireInput]. shortcuts assigns keyboard shortcuts to the
// choices in order — [QuestionnaireShortcutsLetters] A to Z or
// [QuestionnaireShortcutsNumbers] 1 to 9; "" assigns none. A choice with a
// shortcut shows the key in a badge, announces it through aria-keyshortcuts
// and is checked by the key through the [Questionnaire] script.
//
// The shortcuts go to choices passed as direct children (also inside an
// mx.Components, which mx.ForEach returns), in the order given. Choices past
// the last key get none.
func QuestionnaireChoices(shortcuts QuestionnaireShortcuts, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	var keys string
	switch shortcuts {
	case QuestionnaireShortcutsLetters:
		keys = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case QuestionnaireShortcutsNumbers:
		keys = "123456789"
	}
	if keys != "" {
		e.Attribs = append(e.Attribs, html.DataAttr("shortcuts", string(shortcuts)))
		i := 0
		var assign func(cs []mx.Component)
		assign = func(cs []mx.Component) {
			for _, c := range cs {
				switch c := c.(type) {
				case *mx.Element:
					if i < len(keys) && slotOf(c) == "questionnaire-choice" && c.AttribIndex("data-shortcut") < 0 {
						setQuestionnaireShortcut(c, keys[i:i+1])
						i++
					}
				case mx.Components:
					assign(c)
				}
			}
		}
		assign(e.Children)
	}
	return finish(e, "questionnaire-choices",
		"group/questionnaire-choices grid min-w-0 gap-3")
}

// setQuestionnaireShortcut gives the [QuestionnaireChoice] card choice the
// shortcut key: the card's data-shortcut (which the key script looks up), the
// input's aria-keyshortcuts and the badge showing the key.
func setQuestionnaireShortcut(choice *mx.Element, key string) {
	choice.Attribs = append(choice.Attribs, html.DataAttr("shortcut", key))
	for _, c := range choice.Children {
		if input, ok := c.(*mx.Element); ok && slotOf(input) == "questionnaire-choice-input" {
			if input.AttribIndex("aria-keyshortcuts") < 0 {
				input.Attribs = append(input.Attribs, html.Attrib("aria-keyshortcuts", key))
			}
		}
	}
	badge := html.Span(html.Attrib("aria-hidden", "true"), key)
	choice.Children = append(choice.Children, finish(badge, "questionnaire-choice-shortcut",
		"pointer-events-none ms-auto inline-flex shrink-0 border-input bg-background text-muted-foreground size-5 translate-y-[--spacing(0.45)] group-has-data-[slot=questionnaire-choice-description]/questionnaire-choice:translate-y-0.5 items-center justify-center rounded-md border font-mono text-[0.625rem] font-medium leading-none shadow-xs"))
}

// questionnaireChoiceClasses is the card: upstream's skeleton plus
// .cn-questionnaire-choice. Base UI's data-checked: and data-disabled: on the
// card become has-[>input:checked]: and has-[>input:disabled]: (the native
// input is the card's direct child, and a disabled fieldset disables it too),
// and data-invalid: becomes group-data-[invalid=true]/questionnaire-item:,
// keyed off the item that holds a [QuestionnaireError]. A checked card keeps
// its primary border in an invalid item: has-[>input:checked] is the more
// specific selector.
const questionnaireChoiceClasses = "group/questionnaire-choice relative flex min-h-11 cursor-pointer items-start text-start transition-colors outline-none select-none" +
	" has-[>input:disabled]:pointer-events-none has-[>input:disabled]:cursor-not-allowed has-[>input:disabled]:opacity-50" +
	" border-input dark:bg-input/20 hover:bg-muted/50 has-[>input:checked]:border-primary/40 has-[>input:checked]:bg-muted dark:has-[>input:checked]:bg-muted group-data-[invalid=true]/questionnaire-item:border-destructive has-[>input:focus-visible]:border-ring has-[>input:focus-visible]:ring-ring/50 gap-3 rounded-md border bg-transparent px-4 py-3.5 text-sm shadow-xs has-[>input:focus-visible]:ring-3"

// questionnaireIndicatorClasses is the radio dot / checkbox box drawn next to
// the content (.cn-questionnaire-choice-indicator), with Base UI's
// group-data-checked/questionnaire-choice: rewritten to
// group-has-[>input:checked]/questionnaire-choice:. Upstream rounds it per
// type with group-data-[type=radio]/questionnaire-choice:rounded-full; the
// type is known when the card is built, so the port picks the rounding then.
const questionnaireIndicatorClasses = "pointer-events-none relative flex shrink-0 items-center justify-center border border-input dark:bg-input/30 group-has-[>input:checked]/questionnaire-choice:bg-primary dark:group-has-[>input:checked]/questionnaire-choice:bg-primary group-has-[>input:checked]/questionnaire-choice:text-primary-foreground group-has-[>input:checked]/questionnaire-choice:border-primary size-4 translate-y-[--spacing(0.45)] group-has-data-[slot=questionnaire-choice-description]/questionnaire-choice:translate-y-0.5"

// QuestionnaireChoice renders one answer as a selectable card: a <label>
// holding a native radio or checkbox stretched invisibly over the whole card
// (so the card is the click and focus target, and the input stays a real form
// control), a dot or check indicator, the children as the answer's content,
// and, with [QuestionnaireChoices] shortcuts, the key badge.
//
// choiceType may be "" for a radio. name is the form name the answer posts
// under, normally the [QuestionnaireItem]'s name; it must contain only
// letters, digits, '-' and '_' (reported per [PanicOnInvalidID]). value is the
// posted value.
//
// As upstream does with className, a caller class goes to the card; every
// other attribute — html.Checked for a stored answer, html.Disabled, an id —
// goes to the input. Pass the answer's title and a
// [QuestionnaireChoiceDescription] as children.
func QuestionnaireChoice(choiceType QuestionnaireChoiceType, name, value string, attribsChildren ...any) *mx.Element {
	if err := validateID(name); err != nil {
		return mx.NewErrElement(err)
	}
	choiceType = choiceType.ValidOr(QuestionnaireRadio)

	card := html.Label(attribsChildren...)
	var inputAttribs []mx.Attrib
	cardAttribs := card.Attribs[:0]
	for _, a := range card.Attribs {
		if a.AttribName() == "class" {
			cardAttribs = append(cardAttribs, a)
		} else {
			inputAttribs = append(inputAttribs, a)
		}
	}
	card.Attribs = cardAttribs

	input := html.Input(inputAttribs...)
	if input.AttribIndex("type") < 0 {
		input.Attribs = append(input.Attribs, html.Type(string(choiceType)))
	}
	if input.AttribIndex("name") < 0 {
		input.Attribs = append(input.Attribs, html.Name(name))
	}
	if input.AttribIndex("value") < 0 {
		input.Attribs = append(input.Attribs, html.Value(value))
	}
	input = finish(input, "questionnaire-choice-input",
		"absolute inset-0 z-10 size-full cursor-pointer opacity-0")

	var mark *mx.Element
	rounding := "rounded-[4px]"
	if choiceType == QuestionnaireRadio {
		rounding = "rounded-full"
		mark = finish(html.Span(), "questionnaire-choice-indicator-dot",
			"hidden rounded-full bg-primary-foreground size-2 group-has-[>input:checked]/questionnaire-choice:block")
	} else {
		mark = finish(iconCheck(), "questionnaire-choice-indicator-check",
			"hidden size-3.5 group-has-[>input:checked]/questionnaire-choice:block")
	}
	indicator := finish(html.Span(html.Attrib("aria-hidden", "true"), mark),
		"questionnaire-choice-indicator", questionnaireIndicatorClasses+" "+rounding)

	content := html.Span()
	content.Children = card.Children
	content = finish(content, "questionnaire-choice-label",
		"flex min-w-0 flex-1 flex-col leading-snug gap-1")

	card.Children = []mx.Component{input, indicator, content}
	card.Attribs = append(card.Attribs, html.DataAttr("type", string(choiceType)))
	return finish(card, "questionnaire-choice", questionnaireChoiceClasses)
}

// QuestionnaireChoiceDescription renders a secondary line under a
// [QuestionnaireChoice]'s title, as a muted <span>. It also nudges the card's
// indicator and shortcut badge to the title's line height.
func QuestionnaireChoiceDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "questionnaire-choice-description",
		"text-muted-foreground")
}

// QuestionnaireInput renders a free-text answer, alone or after the choices
// in [QuestionnaireChoices] ("Another answer"), as an <input> (type text
// unless the caller passes html.Type) in a wrapper <div>. All attributes go to
// the input: give it a name, and an accessible name through a visible label,
// an aria-label or aria-labelledby — a placeholder is not a label.
//
// Upstream posts a typed answer under the item's own name and drops the
// checked choice when the user types (and the typed text when a choice is
// checked). This port has no script for that: both are posted, so give the
// input a name of its own (e.g. "{item}-other") and let the server decide,
// typically preferring a non-empty typed answer.
func QuestionnaireInput(attribs ...mx.Attrib) *mx.Element {
	input := html.Input(attribs...)
	if input.AttribIndex("type") < 0 {
		input.Attribs = append(input.Attribs, html.Type("text"))
	}
	input = finish(input, "questionnaire-input",
		"min-h-11 w-full min-w-0 transition-[color,box-shadow,background-color] outline-none disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 sm:min-h-0"+
			" selection:bg-primary selection:text-primary-foreground placeholder:text-muted-foreground"+
			" dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 h-9 rounded-md border bg-transparent px-2.5 py-1 text-base shadow-xs focus-visible:ring-3 aria-invalid:ring-3 md:text-sm")
	return finish(html.Div(input), "questionnaire-input-wrapper",
		"group/questionnaire-input relative min-w-0 w-full")
}

// QuestionnaireError renders the validation message of the active item as a
// <p role="alert"> in the destructive color. Render it only when the posted
// answer was invalid: its presence marks the [QuestionnaireItem] around it
// invalid. The message is the children (upstream's English defaults are not
// built in). A caller-supplied role is left untouched.
func QuestionnaireError(attribsChildren ...any) *mx.Element {
	e := html.P(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("alert"))
	}
	return finish(e, "questionnaire-error", "text-destructive text-sm")
}

// QuestionnaireActions lays out the navigation buttons in a three-column
// grid: [QuestionnairePrevious] at the start, [QuestionnaireSkip] and
// [QuestionnaireNext] or [QuestionnaireSubmit] at the end. Every button places
// itself in its column, so a step that leaves one out keeps the others where
// they were.
func QuestionnaireActions(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "questionnaire-actions",
		"grid min-h-11 w-full grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 sm:min-h-9")
}

// QuestionnairePrevious renders the button back to the preceding item: a
// submit button posting [QuestionnaireActionPrevious]. variant may be "" for
// [ButtonOutline], size "" for the default; pass the caption as children.
func QuestionnairePrevious(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	return questionnaireButton("questionnaire-previous", QuestionnaireActionPrevious,
		variant.ValidOr(ButtonOutline), size, false,
		"col-start-1 row-start-1 min-h-11 justify-self-start sm:min-h-0", attribsChildren)
}

// QuestionnaireSkip renders the button that leaves an optional item
// unanswered: a submit button posting [QuestionnaireActionSkip]. Render it
// only on optional items. variant may be "" for [ButtonOutline], size "" for
// the default; pass the caption as children.
func QuestionnaireSkip(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	return questionnaireButton("questionnaire-skip", QuestionnaireActionSkip,
		variant.ValidOr(ButtonOutline), size, false,
		"col-start-2 row-start-1 min-h-11 justify-self-end sm:min-h-0", attribsChildren)
}

// QuestionnaireNext renders the button on to the following item: a submit
// button posting [QuestionnaireActionNext], which Enter presses too
// (aria-keyshortcuts="Enter"). variant and size may be "" for the defaults;
// pass the caption as children.
func QuestionnaireNext(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	return questionnaireButton("questionnaire-next", QuestionnaireActionNext,
		variant.ValidOr(ButtonDefault), size, true,
		"col-start-3 row-start-1 min-h-11 justify-self-end sm:min-h-0", attribsChildren)
}

// QuestionnaireSubmit renders the button that finishes the questionnaire on
// its last item: a submit button posting [QuestionnaireActionSubmit], which
// Enter presses where no [QuestionnaireNext] is rendered. It takes Next's
// column; render one or the other. variant and size may be "" for the
// defaults; pass the caption as children.
func QuestionnaireSubmit(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	return questionnaireButton("questionnaire-submit", QuestionnaireActionSubmit,
		variant.ValidOr(ButtonDefault), size, true,
		"col-start-3 row-start-1 min-h-11 justify-self-end sm:min-h-0", attribsChildren)
}

// questionnaireButton builds the navigation buttons: a [Button]-styled
// <button type="submit" name="questionnaire-action" value="{action}"> with
// upstream's data-variant and data-size, placed in its grid column of
// [QuestionnaireActions]. enter announces the Enter shortcut of Next and
// Submit. Caller-supplied type, name, value and aria-keyshortcuts are left
// untouched.
func questionnaireButton(slot string, action QuestionnaireAction, variant ButtonVariant, size ButtonSize, enter bool, placement string, attribsChildren []any) *mx.Element {
	size = size.ValidOr(SizeDefault)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("submit"))
	}
	if e.AttribIndex("name") < 0 {
		e.Attribs = append(e.Attribs, html.Name(QuestionnaireActionParam))
	}
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(string(action)))
	}
	if enter && e.AttribIndex("aria-keyshortcuts") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-keyshortcuts", "Enter"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", string(variant)),
		html.DataAttr("size", string(size)),
	)
	return finish(e, slot, Cn(ButtonClasses(variant, size), placement))
}

// walkQuestionnaireElements calls fn for every element among cs and their
// descendants, looking through mx.Components; other components are opaque.
func walkQuestionnaireElements(cs []mx.Component, fn func(*mx.Element)) {
	for _, c := range cs {
		switch c := c.(type) {
		case *mx.Element:
			fn(c)
			walkQuestionnaireElements(c.Children, fn)
		case mx.Components:
			walkQuestionnaireElements(c, fn)
		}
	}
}

// slotOf returns the data-slot of e, or "" if it has none.
func slotOf(e *mx.Element) string {
	i := e.AttribIndex("data-slot")
	if i < 0 {
		return ""
	}
	value, err := e.Attribs[i].AttribValue(context.Background())
	if err != nil {
		return ""
	}
	return value
}

// appendElementID appends the id of e to ids, first giving e the id fallback
// if it has none. An id e cannot resolve or that is empty is not referenced.
func appendElementID(ids []string, e *mx.Element, fallback string) []string {
	i := e.AttribIndex("id")
	if i < 0 {
		e.Attribs = append(e.Attribs, html.ID(fallback))
		return append(ids, fallback)
	}
	value, err := e.Attribs[i].AttribValue(context.Background())
	if err != nil || value == "" {
		return ids
	}
	return append(ids, value)
}
