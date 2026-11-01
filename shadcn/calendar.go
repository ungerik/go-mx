//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Calendar is a Go port of shadcn/ui's Calendar. shadcn wraps react-day-picker;
// this port generates one month's grid server-side with Go's time package — no
// client runtime (only the select mode and the month navigation of
// [CalendarWith] add a small script). It renders the month containing `month`,
// with the day equal to `selected` marked (pass the zero time.Time for no
// selection). Its days are text, as react-day-picker renders them without a
// mode; [CalendarOptions].Name makes them buttons to pick from.
//
// Without [CalendarOptions].MonthURL the previous- and next-month buttons are
// plain buttons with no default behavior, so wire them with
// html.HRef("?month=…") or an hx.Get, or set MonthURL to have the calendar
// load the adjacent month itself (see the gallery's Calendar and DatePicker
// examples).
//
// react-day-picker's disabled-matcher is available as a per-day predicate on
// [CalendarWith] (see [CalendarOptions]), which also localizes the week
// start, weekday headings and month label, and picks one day, a range or
// several days in the browser.
func Calendar(month, selected time.Time, attribsChildren ...any) *mx.Element {
	return CalendarWith(CalendarOptions{}, month, selected, attribsChildren...)
}

// defaultWeekdayAbbr holds the English two-letter weekday abbreviations indexed
// by time.Weekday (Sunday..Saturday). It is the WeekdayName used when
// [CalendarOptions.WeekdayName] is nil, preserving the original grid headings.
var defaultWeekdayAbbr = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// CalendarMode selects what a [CalendarWith] with a Name picks, like
// react-day-picker's mode prop.
type CalendarMode string //#enum

const (
	// CalendarSingle picks one day, the default.
	CalendarSingle CalendarMode = "single"
	// CalendarRange picks a range of days: the first click starts it, the
	// second ends it, and a third starts a new one.
	CalendarRange CalendarMode = "range"
	// CalendarMultiple picks any number of days, each click toggling one.
	CalendarMultiple CalendarMode = "multiple"
)

// Valid indicates if c is any of the valid values for CalendarMode
func (c CalendarMode) Valid() bool {
	switch c {
	case
		CalendarSingle,
		CalendarRange,
		CalendarMultiple:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for CalendarMode
func (c CalendarMode) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.CalendarMode", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for CalendarMode, else defaultVal
func (c CalendarMode) ValidOr(defaultVal CalendarMode) CalendarMode {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for CalendarMode
func (CalendarMode) Enums() []CalendarMode {
	return []CalendarMode{
		CalendarSingle,
		CalendarRange,
		CalendarMultiple,
	}
}

// EnumStrings returns all valid values for CalendarMode as strings
func (CalendarMode) EnumStrings() []string {
	return []string{
		"single",
		"range",
		"multiple",
	}
}

// String implements the fmt.Stringer interface for CalendarMode
func (c CalendarMode) String() string {
	return string(c)
}

// CalendarCaptionLayout selects what a [CalendarWith]'s caption shows, like
// react-day-picker's captionLayout prop.
type CalendarCaptionLayout string //#enum

const (
	// CalendarCaptionLabel shows the month label, the default.
	CalendarCaptionLabel CalendarCaptionLayout = "label"
	// CalendarCaptionDropdown shows a month and a year dropdown.
	CalendarCaptionDropdown CalendarCaptionLayout = "dropdown"
	// CalendarCaptionDropdownMonths shows a month dropdown and the year.
	CalendarCaptionDropdownMonths CalendarCaptionLayout = "dropdown-months"
	// CalendarCaptionDropdownYears shows the month and a year dropdown.
	CalendarCaptionDropdownYears CalendarCaptionLayout = "dropdown-years"
)

// Valid indicates if c is any of the valid values for CalendarCaptionLayout
func (c CalendarCaptionLayout) Valid() bool {
	switch c {
	case
		CalendarCaptionLabel,
		CalendarCaptionDropdown,
		CalendarCaptionDropdownMonths,
		CalendarCaptionDropdownYears:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for CalendarCaptionLayout
func (c CalendarCaptionLayout) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.CalendarCaptionLayout", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for CalendarCaptionLayout, else defaultVal
func (c CalendarCaptionLayout) ValidOr(defaultVal CalendarCaptionLayout) CalendarCaptionLayout {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for CalendarCaptionLayout
func (CalendarCaptionLayout) Enums() []CalendarCaptionLayout {
	return []CalendarCaptionLayout{
		CalendarCaptionLabel,
		CalendarCaptionDropdown,
		CalendarCaptionDropdownMonths,
		CalendarCaptionDropdownYears,
	}
}

// EnumStrings returns all valid values for CalendarCaptionLayout as strings
func (CalendarCaptionLayout) EnumStrings() []string {
	return []string{
		"label",
		"dropdown",
		"dropdown-months",
		"dropdown-years",
	}
}

// String implements the fmt.Stringer interface for CalendarCaptionLayout
func (c CalendarCaptionLayout) String() string {
	return string(c)
}

// CalendarOptions configures [CalendarWith]. Its zero value reproduces
// [Calendar] exactly: a Sunday-first grid, English two-letter weekday headings,
// an English "January 2006" month label, no disabled days, and ISO 8601
// accessible day names.
//
// The hooks carry no locale data on purpose. Go's time package has no locale,
// and this library deliberately keeps golang.org/x/text (and any other locale
// database) out of the core module. A German — or any other localized —
// calendar is produced by supplying the name functions, from the caller or a
// future separate module: the component ships the hooks, not the translations.
type CalendarOptions struct {
	// WeekStart is the weekday shown in the first column of the grid. The
	// zero value is time.Sunday, preserving the current US-convention
	// behavior; set time.Monday for the ISO 8601 / continental-European
	// week. It shifts both the column order and which trailing/leading days
	// of the adjacent months fill the grid.
	WeekStart time.Weekday

	// WeekdayName returns the column heading for a weekday. nil uses the
	// current English two-letter abbreviations ("Su".."Sa").
	WeekdayName func(time.Weekday) string

	// WeekdayLabel returns the full name of a weekday, the accessible name
	// (aria-label) of its column heading, so a screen reader moving through
	// the grid announces "Sunday" rather than the WeekdayName abbreviation.
	// nil uses time.Weekday.String(), which is always English because Go's
	// time package has no locale, but only when WeekdayName is nil too: a
	// localized heading is not given an English name.
	WeekdayLabel func(time.Weekday) string

	// MonthLabel returns the header text for the displayed month. nil uses
	// month.Format("January 2006"), which is always English because Go's
	// time package has no locale.
	MonthLabel func(time.Time) string

	// MonthName returns the name of a month in the month dropdown of a
	// CaptionLayout with one. nil uses time.Month.String, which is always
	// English.
	MonthName func(time.Month) string

	// Disabled reports whether a day cannot be selected. nil disables
	// nothing. A disabled day has the muted, faded look, in select mode as a
	// disabled button with aria-disabled="true", and is never marked
	// selected (data-selected, plus aria-selected in select mode or the
	// selected label in a static calendar) even when it equals selected. This is what makes the component usable as a booking
	// calendar: grey out dates with no availability, dates in the past, or
	// dates outside a window.
	Disabled func(time.Time) bool

	// DayLabel returns the accessible name of a day: the aria-label of a day
	// button in select mode, the screen-reader text next to a static day's
	// number. nil uses the ISO 8601 date (2006-01-02), so a screen reader
	// announces the full date instead of a bare day number with no month or
	// year context.
	// A [CalendarText] shows the picked day's DayLabel. Today's name is
	// wrapped in [Labels].CalendarToday, and in select mode a selected
	// day's in [Labels].CalendarSelected, as react-day-picker names them.
	DayLabel func(time.Time) string

	// Today marks that day with the accent background and
	// aria-current="date", as react-day-picker marks today. The zero time
	// marks no day. It is an option rather than time.Now() so the caller
	// decides the time zone of "today" and the output stays deterministic.
	Today time.Time

	// StartMonth and EndMonth bound month navigation: the previous-month
	// button is disabled while the displayed month is StartMonth's month or
	// earlier, the next-month button while it is EndMonth's month or later.
	// A zero time leaves that side open. Like react-day-picker's startMonth
	// and endMonth they only limit navigation; Disabled makes days
	// unselectable.
	StartMonth, EndMonth time.Time

	// MonthURL makes the calendar load another month itself. It returns the
	// URL of the calendar showing month, which is always the first of a
	// month at midnight UTC. The previous- and next-month buttons, the
	// month and year dropdowns, PageUp and PageDown (with Shift a year,
	// stopping at StartMonth and EndMonth), and an arrow key at
	// the edge of the grid fetch that URL and replace the calendar with the
	// first calendar in the response, returning the focus to the control
	// that was used or the day the keys moved to. htmx is not needed; when
	// it is on the page, the new calendar is processed by htmx.process.
	//
	// The handler behind the URL renders the calendar with the same
	// options, typically by sharing the function that builds them:
	//
	//	mux.HandleFunc("GET /calendar/{month}", func(w http.ResponseWriter, r *http.Request) {
	//		month, err := time.Parse("2006-01", r.PathValue("month"))
	//		...
	//		shadcn.CalendarWith(calendarOptions(), month, time.Time{}).Render(r.Context(), mx.NewCheckedWriter(w))
	//	})
	//
	// In select mode the script carries the picked days into the new
	// month, so the handler need not know them; days it marks selected are
	// replaced by the picked ones. nil leaves the buttons to the caller.
	MonthURL func(month time.Time) string

	// CaptionLayout replaces the month label with native month and year
	// dropdowns, bounded by StartMonth and EndMonth. Without StartMonth the
	// years start 100 years before Today's (or else the displayed month's)
	// year, without EndMonth they end with that year, as
	// react-day-picker's defaults. Choosing a month loads it through
	// MonthURL; without MonthURL the dropdowns are disabled.
	CaptionLayout CalendarCaptionLayout

	// Name makes the calendar a form control, the date picker's job in
	// react-day-picker's selection modes. A hidden input with this name carries
	// the selected day as 2006-01-02 ("" for none), and one shared inline
	// script, emitted once per calendar, lets the user pick a day:
	//
	//   - A click (or Enter or Space on a focused day) selects the day,
	//     writes it to the hidden input and fires bubbling input and change
	//     events there, so htmx triggers and form listeners react.
	//   - Arrow keys move the focus by a day or a week, Home and End to the
	//     start or end of the week row, skipping disabled days and following
	//     the text direction. Only one day is a Tab stop: the selected day,
	//     else Today, else the first enabled day of the month, else the first
	//     enabled day of the grid.
	//   - Every [CalendarText] for this name in the same form shows the
	//     picked day's DayLabel.
	//   - A calendar inside an open popover closes it, returning the focus
	//     to the trigger.
	//
	// Mode picks a range or several days instead (see [CalendarMode]). A
	// range has a second hidden input, named EndName, for its last day;
	// several days have one hidden input with this name per day.
	//
	// Without MonthURL, month navigation stays a server round-trip of the
	// caller's: the arrow keys stop at the edge of the six-week grid, and a
	// request that re-renders the calendar for another month should include
	// the hidden input (hx-include) to keep the picked day. "" renders the
	// static grid.
	//
	// Disabled, StartMonth and EndMonth only shape the grid in the browser;
	// the hidden input submits whatever it holds, so the handler must check
	// a submitted day against the same rules. A selected day that is
	// disabled is not marked selected, but the hidden input keeps its date,
	// so an unchanged submit does not clear a stored value.
	Name string

	// Mode selects what the calendar picks when Name is set; "" is
	// CalendarSingle. In a static calendar it only decides how RangeEnd and
	// Selected are shown.
	Mode CalendarMode

	// RangeEnd is the last day of the range whose first day is the
	// selected argument, for CalendarRange. A RangeEnd before selected is
	// swapped with it; a zero RangeEnd marks only the first day.
	RangeEnd time.Time

	// EndName is the name of the hidden input that carries RangeEnd in
	// CalendarRange mode. "" uses Name + "-end".
	EndName string

	// ExcludeDisabled keeps a picked range free of disabled days, as
	// react-day-picker's excludeDisabled does: a second click that would
	// span a disabled day of the shown grid starts a new range there
	// instead, so a booking range cannot cross a booked night. Days of
	// months not shown are not seen, so the handler still checks a
	// submitted range.
	ExcludeDisabled bool

	// Selected are the selected days besides the selected argument, for
	// CalendarMultiple.
	Selected []time.Time

	// NumberOfMonths shows that many consecutive months side by side,
	// starting with the month argument, as react-day-picker's
	// numberOfMonths does for a range picker. They share the month
	// buttons, at the outer ends, the selection and the keyboard: the
	// arrow keys move across the months. The days of the adjacent months
	// are left out, so that no day is shown twice, and only the first
	// month's caption has a CaptionLayout's dropdowns. MonthURL's handler
	// renders the same number of months. 0 and 1 show one month.
	NumberOfMonths int
}

// calendarScript is the once-emitted client code of a calendar with
// [CalendarOptions].Name or MonthURL. The hidden inputs hold the picked days
// (with their DayLabel in data-label), and apply marks the grid from them, so
// a day picked before a month change is marked again in the new month.
// calendarSelect picks a day button, calendarKey moves the focus between day
// buttons by date (so across the months shown side by side, and into the
// adjacent month past the last one), calendarNav loads another month from a
// control's data-month-url, and calendarPick picks a day from outside the
// calendar.
const calendarScript = /*js*/ `if(!window.calendarSelect){(function(){
var all=function(r,q){return Array.prototype.slice.call(r.querySelectorAll(q));};
var root=function(e){return e.closest('[data-slot=calendar]');};
var flag=function(e,n,on){if(on)e.setAttribute('data-'+n,'true');else e.removeAttribute('data-'+n);};
var iso=function(t){return t.toISOString().slice(0,10);};
var addDays=function(s,n){return iso(new Date(Date.parse(s)+n*864e5));};
var addMonths=function(s,n){var p=s.split('-'),y=+p[0],m=+p[1]-1+n,d=+p[2],last=new Date(Date.UTC(y,m+1,0)).getUTCDate();return iso(new Date(Date.UTC(y,m,Math.min(d,last))));};
var state=function(r){return all(r,'[data-slot=calendar-input]').map(function(i){return [i.value,i.getAttribute('data-label')||i.value];});};
var setState=function(r,s){var c=r.querySelector('[data-slot=calendar-inputs]');if(c){all(c,'input').forEach(function(i){i.remove();});s.forEach(function(x){var i=document.createElement('input');i.type='hidden';i.name=c.getAttribute('data-name');i.setAttribute('data-slot','calendar-input');c.appendChild(i);});}all(r,'[data-slot=calendar-input]').forEach(function(i,k){var x=s[k]||['',''];i.value=x[0];if(x[0])i.setAttribute('data-label',x[1]);else i.removeAttribute('data-label');});};
var name=function(r,b,sel){var l=b.getAttribute('data-label');if(b.hasAttribute('aria-current'))l=r.getAttribute('data-today-label').replace('%s',l);if(sel)l=r.getAttribute('data-selected-label').replace('%s',l);return l;};
var apply=function(r,stop){var m=r.getAttribute('data-mode'),v=state(r).map(function(x){return x[0];}).filter(Boolean),from=v[0],to=v[1]||v[0],first=null,days=all(r,'[data-date]');
days.forEach(function(b){var d=b.getAttribute('data-date'),c=b.parentElement,s=false,rs=false,re=false,rm=false;if(!b.disabled&&from){if(m==='range'){rs=d===from;re=d===to;rm=d>from&&d<to;s=rs||re||rm;}else{s=v.indexOf(d)>=0;}}
flag(b,'selected',s&&!rm);flag(b,'range-start',rs);flag(b,'range-end',re);flag(b,'range-middle',rm);flag(c,'range-start',rs);flag(c,'range-end',re);flag(c,'range-middle',rm);if(s)c.setAttribute('aria-selected','true');else c.removeAttribute('aria-selected');b.setAttribute('aria-label',name(r,b,s));if(s&&!first)first=b;});
stop=stop||first;if(stop){days.forEach(function(b){b.tabIndex=-1;});stop.tabIndex=0;}};
var showText=function(r){var s=state(r).filter(function(x){return x[0];}),l=s.map(function(x){return x[1];}),f=r.closest('form'),t=r.getAttribute('data-mode')==='range'?l.join(' – '):l.join(', ');
all(f||document,'[data-calendar-text]').forEach(function(e){if(e.getAttribute('data-calendar-text')!==r.getAttribute('data-name'))return;if(!e.hasAttribute('data-placeholder'))e.setAttribute('data-placeholder',e.hasAttribute('data-empty')?e.textContent:'');if(e.tagName==='INPUT'){e.value=t;return;}if(t){e.textContent=t;e.removeAttribute('data-empty');}else{e.textContent=e.getAttribute('data-placeholder');e.setAttribute('data-empty','true');}});};
window.calendarSelect=function(b){var r=root(b);if(!r||b.disabled)return;var m=r.getAttribute('data-mode'),d=b.getAttribute('data-date'),x=[d,b.getAttribute('data-label')],s=state(r).filter(function(y){return y[0];}),done=true;
if(m==='multiple'){var k=s.findIndex(function(y){return y[0]===d;});if(k>=0)s.splice(k,1);else{s.push(x);s.sort(function(a,c){return a[0]<c[0]?-1:1;});}done=false;}
else if(m==='range'){var lo=s.length===1?(d<s[0][0]?d:s[0][0]):'',hi=s.length===1?(d<s[0][0]?s[0][0]:d):'';if(s.length!==1||r.hasAttribute('data-exclude-disabled')&&all(r,'[data-date]').some(function(c){var cd=c.getAttribute('data-date');return c.disabled&&cd>lo&&cd<hi;})){s=[x];done=false;}else if(d<s[0][0]){s=[x,s[0]];}else{s=[s[0],x];}}
else{s=[x];}
setState(r,s);apply(r,b);var e=r.querySelector('[data-slot=calendar-inputs]')||r.querySelector('[data-slot=calendar-input]');if(e){e.dispatchEvent(new Event('input',{bubbles:true}));e.dispatchEvent(new Event('change',{bubbles:true}));}showText(r);
var p=r.closest('[popover]');if(done&&p&&p.matches(':popover-open'))p.hidePopover();};
window.calendarNav=function(b,f,y){if(!b||b.disabled)return;var r=root(b),u=b.tagName==='SELECT'?b.options[b.selectedIndex].getAttribute('data-month-url'):b.getAttribute(y?'data-year-url':'data-month-url');if(!r||!u)return;var slot=b.getAttribute('data-slot'),s=state(r);
fetch(u,{headers:{Accept:'text/html'}}).then(function(x){if(!x.ok)throw new Error(x.status);return x.text();}).then(function(h){var t=document.createElement('template');t.innerHTML=h;var n=t.content.querySelector('[data-slot=calendar]');if(!n)return;r.replaceWith(n);
var d=f?n.querySelector('[data-date="'+f+'"]'):null;if(d&&d.disabled)d=null;if(n.hasAttribute('data-mode')){setState(n,s);apply(n,d);}if(window.htmx)window.htmx.process(n);
if(!d)d=f?n.querySelector("[data-date][tabindex='0']"):n.querySelector('[data-slot="'+slot+'"]');if(d&&d.disabled)d=n.querySelector("[data-date][tabindex='0']");if(d)d.focus();}).catch(function(){});};
window.calendarKey=function(e){var b=e.target.closest('[data-date]');if(!b)return;var r=root(b),d=b.getAttribute('data-date'),col=b.parentElement.cellIndex,step,t,c,n=0,rtl=getComputedStyle(r).direction==='rtl';
switch(e.key){case'PageUp':case'PageDown':e.preventDefault();var k=e.key==='PageUp'?-1:1;window.calendarNav(r.querySelector('[data-slot=calendar-'+(k<0?'prev':'next')+']'),addMonths(d,e.shiftKey?12*k:k),e.shiftKey);return;
case'ArrowLeft':step=rtl?1:-1;t=addDays(d,step);break;case'ArrowRight':step=rtl?-1:1;t=addDays(d,step);break;case'ArrowUp':step=-7;t=addDays(d,-7);break;case'ArrowDown':step=7;t=addDays(d,7);break;case'Home':step=1;t=addDays(d,-col);break;case'End':step=-1;t=addDays(d,6-col);break;default:return;}
e.preventDefault();while((c=r.querySelector('[data-date="'+t+'"]'))&&c.disabled&&t!==d&&n++<62){t=addDays(t,step);}
if(!c){window.calendarNav(r.querySelector('[data-slot=calendar-'+(t<d?'prev':'next')+']'),t);return;}
if(c.disabled)return;all(r,'[data-date]').forEach(function(x){x.tabIndex=-1;});c.tabIndex=0;c.focus();};
window.calendarPick=function(nm,d,l){var r=document.querySelector('[data-slot=calendar][data-name="'+nm+'"]');if(!r)return false;setState(r,d?[[d,l||d]]:[]);apply(r);var e=r.querySelector('[data-slot=calendar-inputs]')||r.querySelector('[data-slot=calendar-input]');if(e){e.dispatchEvent(new Event('input',{bubbles:true}));e.dispatchEvent(new Event('change',{bubbles:true}));}showText(r);return true;};
})();}`

// CalendarText renders the text of a control that shows the day picked in
// a [Calendar] whose [CalendarOptions].Name is name, such as the label of a
// date picker's [PopoverTrigger]. Pass the label of the current day as
// children, or a placeholder together with html.DataAttr("empty", "true"),
// which renders muted like upstream's empty date picker. When the user
// picks a day, the calendar script replaces the text with that day's
// DayLabel and removes data-empty; a range shows its first and last day
// joined by " – ", several days are joined by ", ", and a placeholder comes
// back when the last day is unpicked. The script only looks in the form
// that holds the calendar, or in the whole document when there is none.
//
// An <input> may carry data-calendar-text too (html.DataAttr("calendar-text",
// name)): its value is set instead, for a date picker typed into. Picking
// from outside the calendar, such as from that input or a list of presets,
// is the script's calendarPick(name, date, label): it picks date
// (2006-01-02, "" for none) as if clicked, shown as label, and fires the
// same events.
func CalendarText(name string, attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("data-calendar-text") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("calendar-text", name))
	}
	return finish(e, "calendar-text", "data-[empty=true]:text-muted-foreground")
}

// calendarDayName is the accessible name of a day: the DayLabel, wrapped in
// the context's [Labels].CalendarToday for today and CalendarSelected for a
// selected day, as react-day-picker's labelDayButton names them. It is the
// aria-label of a day button, and the screen-reader text of a static day.
type calendarDayName struct {
	label           string
	today, selected bool
}

// AttribName returns "aria-label".
func (calendarDayName) AttribName() string { return "aria-label" }

// AttribValue returns the day's name for the context's [Labels].
func (a calendarDayName) AttribValue(ctx context.Context) (string, error) {
	name := a.label
	if !a.today && !a.selected {
		return name, nil
	}
	labels := LabelsFromContext(ctx)
	if a.today {
		name = strings.Replace(labels.CalendarToday, "%s", name, 1)
	}
	if a.selected {
		name = strings.Replace(labels.CalendarSelected, "%s", name, 1)
	}
	return name, nil
}

// calendarDayClassCache maps the class strings a calendar day button is
// built from to their merged form. A day's classes take only a handful of
// values (the base, range, muted, faded and today variants), so merging
// each once instead of for each of 42 days per calendar keeps [Cn] out of
// the cost of rendering a month.
var calendarDayClassCache sync.Map

// calendarDayClasses returns [Cn] of cls, cached.
func calendarDayClasses(cls string) string {
	if merged, ok := calendarDayClassCache.Load(cls); ok {
		return merged.(string)
	}
	merged := Cn(cls)
	calendarDayClassCache.Store(cls, merged)
	return merged
}

// calendarMonth is one month of a [CalendarWith]: its first day and the six
// weeks of its grid, which days of them are shown and which are disabled.
type calendarMonth struct {
	first    time.Time
	days     [42]time.Time
	shown    [42]bool
	disabled [42]bool
}

// calendarDayMarks is how one day of a [CalendarWith] grid is selected.
type calendarDayMarks struct {
	selected, rangeStart, rangeEnd, rangeMiddle bool
}

// CalendarWith is the configurable variant of [Calendar]. A zero-value
// [CalendarOptions] renders exactly what [Calendar] renders, byte for byte; the
// options localize the grid (WeekStart, WeekdayName, WeekdayLabel, MonthLabel,
// MonthName, DayLabel),
// turn it into a booking calendar (Disabled, Today, StartMonth/EndMonth), let
// it load other months (MonthURL, CaptionLayout) or make it a form control
// picked in the browser (Name, Mode, see [CalendarOptions].Name). The
// Xxx / XxxWith split mirrors
// [ReflectFormHandler] / ReflectFormHandlerWith in the root package.
func CalendarWith(opts CalendarOptions, month, selected time.Time, attribsChildren ...any) *mx.Element {
	weekdayName := opts.WeekdayName
	if weekdayName == nil {
		weekdayName = func(wd time.Weekday) string { return defaultWeekdayAbbr[wd] }
	}
	weekdayLabel := opts.WeekdayLabel
	if weekdayLabel == nil && opts.WeekdayName == nil {
		weekdayLabel = time.Weekday.String
	}
	monthLabel := opts.MonthLabel
	if monthLabel == nil {
		monthLabel = func(m time.Time) string { return m.Format("January 2006") }
	}
	dayLabel := opts.DayLabel
	if dayLabel == nil {
		dayLabel = func(d time.Time) string { return d.Format("2006-01-02") }
	}
	mode := opts.Mode.ValidOr(CalendarSingle)

	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	sameDay := func(a, b time.Time) bool {
		ay, am, ad := a.Date()
		by, bm, bd := b.Date()
		return ay == by && am == bm && ad == bd
	}
	// The picked days, as 2006-01-02, which compare in date order.
	var picked []string
	if !selected.IsZero() {
		picked = append(picked, selected.Format(time.DateOnly))
	}
	switch mode {
	case CalendarRange:
		if len(picked) == 1 && !opts.RangeEnd.IsZero() {
			picked = append(picked, opts.RangeEnd.Format(time.DateOnly))
			slices.Sort(picked)
		}
	case CalendarMultiple:
		for _, day := range opts.Selected {
			if !day.IsZero() {
				picked = append(picked, day.Format(time.DateOnly))
			}
		}
		slices.Sort(picked)
		picked = slices.Compact(picked)
	}
	marks := func(day time.Time) calendarDayMarks {
		if len(picked) == 0 {
			return calendarDayMarks{}
		}
		d := day.Format(time.DateOnly)
		if mode == CalendarRange {
			from, to := picked[0], picked[len(picked)-1]
			m := calendarDayMarks{rangeStart: d == from, rangeEnd: d == to, rangeMiddle: d > from && d < to}
			m.selected = m.rangeStart || m.rangeEnd || m.rangeMiddle
			return m
		}
		return calendarDayMarks{selected: slices.Contains(picked, d)}
	}
	// The months shown, each a grid of six weeks from the WeekStart weekday
	// on or before the first of the month, with Disabled asked once per
	// shown day: a booking calendar's predicate may look up availability.
	// Side by side, the days of the adjacent months are left out, as
	// react-day-picker leaves them out, so that no day is shown twice.
	count := max(opts.NumberOfMonths, 1)
	months := make([]calendarMonth, count)
	for mi := range months {
		m := &months[mi]
		m.first = first.AddDate(0, mi, 0)
		startOffset := (int(m.first.Weekday()) - int(opts.WeekStart) + 7) % 7
		gridStart := m.first.AddDate(0, 0, -startOffset)
		for i := range m.days {
			m.days[i] = gridStart.AddDate(0, 0, i)
			m.shown[i] = count == 1 || m.days[i].Month() == m.first.Month()
			m.disabled[i] = m.shown[i] && opts.Disabled != nil && opts.Disabled(m.days[i])
		}
	}
	selectable := opts.Name != ""
	// In select mode one day is the grid's Tab stop: the first selected day,
	// else today, else the first enabled day of the month, else the first
	// enabled day of the grid, so a month with every day disabled stays
	// reachable.
	var tabStop time.Time
	if selectable {
		var selectedDay, today, firstInMonth, firstInGrid time.Time
		for _, m := range months {
			for i, day := range m.days {
				if !m.shown[i] || m.disabled[i] {
					continue
				}
				inMonth := day.Month() == m.first.Month()
				switch {
				case marks(day).selected:
					if selectedDay.IsZero() {
						selectedDay = day
					}
				case inMonth && !opts.Today.IsZero() && sameDay(day, opts.Today) && today.IsZero():
					today = day
				case inMonth && firstInMonth.IsZero():
					firstInMonth = day
				}
				if firstInGrid.IsZero() {
					firstInGrid = day
				}
			}
		}
		tabStop = cmp.Or(selectedDay, today, firstInMonth, firstInGrid)
	}

	head := html.Element("tr")
	for col := range 7 {
		wd := time.Weekday((int(opts.WeekStart) + col) % 7)
		th := html.Element("th", html.Class("text-muted-foreground w-8 pb-1 text-[0.8rem] font-normal"))
		if weekdayLabel != nil {
			th.Attribs = append(th.Attribs, html.Attrib("aria-label", weekdayLabel(wd)))
		}
		th.Children = append(th.Children, mx.Text(weekdayName(wd)))
		head.Children = append(head.Children, th)
	}

	tables := make([]*mx.Element, count)
	for mi, mo := range months {
		body := html.Element("tbody")
		for week := range 6 {
			row := html.Element("tr")
			for d := range 7 {
				day := mo.days[week*7+d]
				disabled := mo.disabled[week*7+d]
				if !mo.shown[week*7+d] {
					// A day of an adjacent month, left out side by side.
					empty := html.Element("td", html.Class("p-0"))
					if selectable {
						empty.Attribs = append(empty.Attribs, html.Role("gridcell"))
					}
					row.Children = append(row.Children, empty)
					continue
				}
				// A static day is text, so it has no hover, focus or disabled
				// state of a button.
				cls := "inline-flex size-8 items-center justify-center rounded-md text-sm font-normal data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground"
				if selectable {
					cls = "inline-flex size-8 items-center justify-center rounded-md text-sm font-normal outline-none hover:bg-accent hover:text-accent-foreground focus-visible:relative focus-visible:z-10 focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground data-[selected=true]:hover:bg-primary"
				}
				if mode == CalendarRange {
					// The range classes of upstream's CalendarDayButton, with
					// logical sides so the range also joins up right to left.
					cls += " data-[range-middle=true]:rounded-none data-[range-middle=true]:bg-accent data-[range-middle=true]:text-accent-foreground data-[range-middle=true]:hover:bg-accent data-[range-start=true]:rounded-e-none data-[range-end=true]:rounded-s-none data-[range-start=true]:data-[range-end=true]:rounded-md"
				}
				// Days of the adjacent months are muted, as upstream's outside
				// days; only a disabled day is faded too, so a pickable outside
				// day does not look disabled.
				if day.Month() != mo.first.Month() || disabled {
					cls += " text-muted-foreground"
				}
				if disabled {
					cls += " opacity-50"
				}
				isToday := !opts.Today.IsZero() && sameDay(day, opts.Today)
				if isToday {
					// finish merges the classes, so these replace a muted color.
					cls += " bg-accent text-accent-foreground"
				}
				// A disabled day is never selectable, so it is never marked
				// selected even when it equals selected.
				m := calendarDayMarks{}
				if !disabled {
					m = marks(day)
				}
				// In select mode a day is a button named by its DayLabel, and
				// the grid cell carries aria-selected, which a button may not.
				// A static calendar's day is text, as react-day-picker renders
				// a day without a mode: the number shown, and its DayLabel,
				// with the today and selected labels, for screen readers. The
				// day's data-selected carries the selected look in either mode.
				name := calendarDayName{label: dayLabel(day), today: isToday, selected: m.selected}
				var btn *mx.Element
				if selectable {
					btn = html.ButtonButton(name, strconv.Itoa(day.Day()))
				} else {
					btn = html.Span(
						html.Span(html.Attrib("aria-hidden", "true"), strconv.Itoa(day.Day())),
						html.Span(html.Class("sr-only"), mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
							text, err := name.AttribValue(ctx)
							if err != nil {
								return err
							}
							return w.EscapeText(text)
						})),
					)
				}
				cellCls := "p-0 text-center"
				if mode == CalendarRange {
					// Upstream's range_start/range_end cells: the accent band
					// behind the range, rounded at its ends.
					cellCls += " data-[range-start=true]:rounded-s-md data-[range-start=true]:bg-accent data-[range-end=true]:rounded-e-md data-[range-end=true]:bg-accent data-[range-middle=true]:bg-accent"
				}
				cell := html.Element("td", html.Class(cellCls))
				if selectable {
					cell.Attribs = append(cell.Attribs, html.Role("gridcell"))
				}
				switch {
				case disabled && selectable:
					btn.Attribs = append(btn.Attribs, html.Disabled, html.Attrib("aria-disabled", "true"))
				case m.selected:
					// The middle of a range is selected but shown in the accent
					// band, not the primary color of its ends.
					if !m.rangeMiddle {
						btn.Attribs = append(btn.Attribs, html.DataAttr("selected", "true"))
					}
					if selectable {
						cell.Attribs = append(cell.Attribs, html.Attrib("aria-selected", "true"))
					}
				}
				for _, r := range []struct {
					name string
					on   bool
				}{{"range-start", m.rangeStart}, {"range-end", m.rangeEnd}, {"range-middle", m.rangeMiddle}} {
					if r.on {
						btn.Attribs = append(btn.Attribs, html.DataAttr(r.name, "true"))
						cell.Attribs = append(cell.Attribs, html.DataAttr(r.name, "true"))
					}
				}
				if isToday {
					btn.Attribs = append(btn.Attribs, html.Attrib("aria-current", "date"))
				}
				if selectable {
					tabIndex := -1
					if !tabStop.IsZero() && sameDay(day, tabStop) {
						tabIndex = 0
					}
					btn.Attribs = append(btn.Attribs,
						html.DataAttr("date", day.Format(time.DateOnly)),
						html.DataAttr("label", dayLabel(day)),
						html.TabIndex(tabIndex),
						html.OnClick("calendarSelect(this)"),
					)
				}
				// The class is merged once per distinct string, not per day.
				dayButton := finish(btn, "calendar-day", "")
				dayButton.Attribs = append(dayButton.Attribs, mx.NewAttrib("class", calendarDayClasses(cls)))
				cell.Children = append(cell.Children, dayButton)
				row.Children = append(row.Children, cell)
			}
			body.Children = append(body.Children, row)
		}
		table := html.Element("table", html.Class("w-full border-collapse"),
			html.Element("thead", head),
			body,
		)
		if selectable {
			// A grid named by its month, like react-day-picker's, so a screen
			// reader announces the picked day and passes on the arrow keys.
			table.Attribs = append(table.Attribs,
				html.Role("grid"),
				html.Attrib("aria-label", monthLabel(mo.first)),
				html.OnKeyDown("calendarKey(event)"),
			)
			if mode != CalendarSingle {
				table.Attribs = append(table.Attribs, html.Attrib("aria-multiselectable", "true"))
			}
		} else {
			// The static grid has no keyboard handling, so it stays a plain
			// table, named by its month, so a screen reader reading a day's
			// cell knows the month the plain number belongs to.
			table.Attribs = append(table.Attribs, html.Attrib("aria-label", monthLabel(mo.first)))
		}
		tables[mi] = table
	}

	monthIndex := func(t time.Time) int { return t.Year()*12 + int(t.Month()) }
	navigates := opts.MonthURL != nil
	prev := html.ButtonButton(
		html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
		labelAriaLabel(func(l Labels) string { return l.CalendarPrevious }),
		iconChevronLeft())
	if !opts.StartMonth.IsZero() && monthIndex(first) <= monthIndex(opts.StartMonth) {
		prev.Attribs = append(prev.Attribs, html.Disabled)
	} else if navigates {
		// A year back stops at StartMonth, as Shift+PageUp does in
		// react-day-picker.
		yearBack := first.AddDate(-1, 0, 0)
		if start := opts.StartMonth; !start.IsZero() && monthIndex(yearBack) < monthIndex(start) {
			yearBack = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
		prev.Attribs = append(prev.Attribs,
			html.DataAttr("month-url", opts.MonthURL(first.AddDate(0, -1, 0))),
			html.DataAttr("year-url", opts.MonthURL(yearBack)),
			html.OnClick("calendarNav(this)"))
	}
	next := html.ButtonButton(
		html.Class(ButtonClasses(ButtonOutline, SizeIcon)+" size-7"),
		labelAriaLabel(func(l Labels) string { return l.CalendarNext }),
		iconChevronRight())
	last := months[count-1].first
	if !opts.EndMonth.IsZero() && monthIndex(last) >= monthIndex(opts.EndMonth) {
		next.Attribs = append(next.Attribs, html.Disabled)
	} else if navigates {
		// A year on stops where the last month shown is EndMonth's.
		yearOn := first.AddDate(1, 0, 0)
		if end := opts.EndMonth; !end.IsZero() {
			if lastFirst := time.Date(end.Year(), end.Month()-time.Month(count-1), 1, 0, 0, 0, 0, time.UTC); yearOn.After(lastFirst) {
				yearOn = lastFirst
			}
		}
		next.Attribs = append(next.Attribs,
			html.DataAttr("month-url", opts.MonthURL(first.AddDate(0, 1, 0))),
			html.DataAttr("year-url", opts.MonthURL(yearOn)),
			html.OnClick("calendarNav(this)"))
	}
	prevButton, nextButton := finish(prev, "calendar-prev", ""), finish(next, "calendar-next", "")
	var structure []any
	if count == 1 {
		structure = []any{
			html.DivClass("flex items-center justify-between pb-2",
				prevButton, calendarCaption(opts, first, monthLabel), nextButton),
			tables[0],
		}
	} else {
		// Side by side, the month buttons sit at the outer ends of the
		// captions, and only the first caption offers dropdowns.
		row := html.DivClass("flex flex-col gap-4 sm:flex-row")
		for mi, table := range tables {
			before, after := mx.Component(html.DivClass("size-7")), mx.Component(html.DivClass("size-7"))
			if mi == 0 {
				before = prevButton
			}
			if mi == count-1 {
				after = nextButton
			}
			captionOpts := opts
			if mi > 0 {
				captionOpts.CaptionLayout = CalendarCaptionLabel
			}
			row.Children = append(row.Children, html.Div(
				html.DivClass("flex items-center justify-between pb-2",
					before, calendarCaption(captionOpts, months[mi].first, monthLabel), after),
				table,
			))
		}
		structure = []any{row}
	}
	if selectable {
		structure = append(structure, calendarInputs(opts, mode, picked, dayLabel)...)
	}
	if selectable || navigates {
		structure = append(structure, html.ScriptJS(calendarScript))
	}
	// Structural children first, then any caller attribs/children; finish
	// merges a caller class into the base.
	root := html.Div(append(structure, attribsChildren...)...)
	if selectable {
		// What the script needs to pick days and name them again.
		root.Attribs = append(root.Attribs,
			html.DataAttr("mode", string(mode)),
			html.DataAttr("name", opts.Name),
		)
		if mode == CalendarRange && opts.ExcludeDisabled {
			root.Attribs = append(root.Attribs, html.DataAttr("exclude-disabled", "true"))
		}
		root.Attribs = append(root.Attribs,
			labelAttrib{name: "data-today-label", get: func(l Labels) string { return l.CalendarToday }},
			labelAttrib{name: "data-selected-label", get: func(l Labels) string { return l.CalendarSelected }},
		)
	}
	return finish(root, "calendar", "bg-background w-fit rounded-md border p-3")
}

// calendarInputs returns the hidden inputs of a calendar in select mode,
// holding the picked days (sorted 2006-01-02 dates) with their DayLabel in
// data-label for the [CalendarText]: one for a single day, two for a range
// (its first and last day) and, for several days, one per day in a
// container that names them for the days the script adds.
func calendarInputs(opts CalendarOptions, mode CalendarMode, picked []string, dayLabel func(time.Time) string) []any {
	input := func(name, value string) *mx.Element {
		in := html.Input(html.Type("hidden"), html.Name(name), html.Value(value),
			mx.NewAttrib("data-slot", "calendar-input"))
		if value != "" {
			if day, err := time.Parse(time.DateOnly, value); err == nil {
				in.Attribs = append(in.Attribs, html.DataAttr("label", dayLabel(day)))
			}
		}
		return in
	}
	value := func(i int) string {
		if i < len(picked) {
			return picked[i]
		}
		return ""
	}
	switch mode {
	case CalendarRange:
		endName := cmp.Or(opts.EndName, opts.Name+"-end")
		// A range of which only the first day is picked has no end yet.
		return []any{input(opts.Name, value(0)), input(endName, value(1))}
	case CalendarMultiple:
		inputs := html.Span(html.DataAttr("slot", "calendar-inputs"), html.DataAttr("name", opts.Name), html.Hidden)
		for _, day := range picked {
			inputs.Children = append(inputs.Children, input(opts.Name, day))
		}
		return []any{inputs}
	}
	return []any{input(opts.Name, value(0))}
}

// calendarCaption renders the caption between the month buttons: the month
// label, or the month and year dropdowns of opts.CaptionLayout.
func calendarCaption(opts CalendarOptions, first time.Time, monthLabel func(time.Time) string) *mx.Element {
	layout := opts.CaptionLayout.ValidOr(CalendarCaptionLabel)
	if layout == CalendarCaptionLabel {
		return html.DivClass("text-sm font-medium", monthLabel(first))
	}
	monthName := opts.MonthName
	if monthName == nil {
		monthName = time.Month.String
	}
	// The dropdowns' range, with react-day-picker's defaults: from 100 years
	// back to the end of the year of today, and always holding the
	// displayed month.
	ref := cmp.Or(opts.Today, first)
	lo := time.Date(ref.Year()-100, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !opts.StartMonth.IsZero() {
		lo = time.Date(opts.StartMonth.Year(), opts.StartMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	hi := time.Date(ref.Year(), time.December, 1, 0, 0, 0, 0, time.UTC)
	if !opts.EndMonth.IsZero() {
		hi = time.Date(opts.EndMonth.Year(), opts.EndMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	if first.Before(lo) {
		lo = first
	}
	if first.After(hi) {
		hi = first
	}
	option := func(value, label string, month time.Time, selected bool) *mx.Element {
		o := html.Option(html.Value(value), label)
		switch {
		case month.Before(lo) || month.After(hi):
			o.Attribs = append(o.Attribs, html.Disabled)
		case opts.MonthURL != nil:
			o.Attribs = append(o.Attribs, html.DataAttr("month-url", opts.MonthURL(month)))
		}
		if selected {
			o.Attribs = append(o.Attribs, html.Selected)
		}
		return o
	}
	dropdown := func(slot string, label func(Labels) string, shown string, options []any) *mx.Element {
		sel := html.Select(append([]any{
			html.Class("absolute inset-0 w-full opacity-0 bg-popover"),
			labelAriaLabel(label),
		}, options...)...)
		if opts.MonthURL != nil {
			sel.Attribs = append(sel.Attribs, html.OnChange("calendarNav(this)"))
		} else {
			sel.Attribs = append(sel.Attribs, html.Disabled)
		}
		// The select is transparent over its shown label, as
		// react-day-picker draws its dropdowns, so the caption keeps the
		// calendar's look and the platform still opens the native list.
		return html.SpanClass("relative rounded-md border border-input shadow-xs has-focus:border-ring has-focus:ring-[3px] has-focus:ring-ring/50",
			finish(sel, slot, ""),
			html.SpanClass("flex h-8 items-center gap-1 rounded-md ps-2 pe-1 text-sm font-medium [&>svg]:size-3.5 [&>svg]:text-muted-foreground",
				html.Attrib("aria-hidden", "true"), shown, iconChevronDown()),
		)
	}
	var parts []any
	if layout == CalendarCaptionDropdown || layout == CalendarCaptionDropdownMonths {
		var options []any
		for m := time.January; m <= time.December; m++ {
			month := time.Date(first.Year(), m, 1, 0, 0, 0, 0, time.UTC)
			options = append(options, option(month.Format("2006-01"), monthName(m), month, m == first.Month()))
		}
		parts = append(parts, dropdown("calendar-month-select", func(l Labels) string { return l.CalendarMonth }, monthName(first.Month()), options))
	} else {
		parts = append(parts, html.Span(monthName(first.Month())))
	}
	if layout == CalendarCaptionDropdown || layout == CalendarCaptionDropdownYears {
		var options []any
		for y := lo.Year(); y <= hi.Year(); y++ {
			// A year shows the displayed month, moved into the range at
			// its ends.
			month := time.Date(y, first.Month(), 1, 0, 0, 0, 0, time.UTC)
			if month.Before(lo) {
				month = lo
			}
			if month.After(hi) {
				month = hi
			}
			options = append(options, option(strconv.Itoa(y), strconv.Itoa(y), month, y == first.Year()))
		}
		parts = append(parts, dropdown("calendar-year-select", func(l Labels) string { return l.CalendarYear }, strconv.Itoa(first.Year()), options))
	} else {
		parts = append(parts, html.Span(strconv.Itoa(first.Year())))
	}
	return html.DivClass("flex items-center justify-center gap-1.5 text-sm font-medium", parts...)
}
