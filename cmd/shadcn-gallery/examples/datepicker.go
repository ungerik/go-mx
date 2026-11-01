package examples

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func init() {
	addMonthFragments("date-picker",
		time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
		DatePickerCalendar)
	addMonthFragments("date-range-picker",
		time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		DateRangePickerCalendar)
}

// DatePickerDemo is shadcn's DatePicker recipe: a Popover whose trigger shows
// the chosen date and whose content is a Calendar. shadcn ships DatePicker as a
// copy-paste composition, not an exported component, so this is a composition
// example rather than a new shadcn primitive. The Calendar's Name turns on its
// select mode: picking a day writes it to a hidden "demo-date" input, shows
// its DayLabel in the trigger's CalendarText and closes the popover.
func DatePickerDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-datepicker",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[280px] justify-start text-left font-normal has-[[data-empty=true]]:text-muted-foreground"),
			// lucide calendar
			navIcon("M8 2v4", "M16 2v4", "M5 4h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z", "M3 10h18"),
			shadcn.CalendarText("demo-date", html.DataAttr("empty", "true"), "Pick a date"),
		),
		shadcn.PopoverContent("demo-datepicker", "",
			html.Class("w-auto p-0"),
			DatePickerCalendar(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)),
		),
	)
}

// DatePickerCalendar is the Calendar in the DatePicker's popover. MonthURL
// loads the other months of 2026, which the gallery pre-renders as static
// fragments; the calendar carries the picked day into them.
func DatePickerCalendar(month time.Time) mx.Component {
	return shadcn.CalendarWith(
		shadcn.CalendarOptions{
			Name:       "demo-date",
			DayLabel:   func(d time.Time) string { return d.Format("January 2, 2006") },
			StartMonth: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			EndMonth:   time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
			MonthURL:   monthURL("date-picker"),
		},
		month,
		time.Time{},
		html.Class("border-0"),
	)
}

// DateRangePickerDemo is upstream's Date Range Picker: two months side by
// side, the trigger showing the first and last day of the range, and the
// popover staying open until the second click ends it.
func DateRangePickerDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-rangepicker",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[300px] justify-start text-left font-normal"),
			navIcon("M8 2v4", "M16 2v4", "M5 4h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z", "M3 10h18"),
			shadcn.CalendarText("demo-range", "Jan 20, 2026 – Feb 9, 2026"),
		),
		shadcn.PopoverContent("demo-rangepicker", "",
			html.Class("w-auto p-0"),
			DateRangePickerCalendar(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
		),
	)
}

// DateRangePickerCalendar is the range Calendar in the Date Range Picker's
// popover, with the range upstream's example starts with, navigating from
// January to March 2026.
func DateRangePickerCalendar(month time.Time) mx.Component {
	return shadcn.CalendarWith(
		shadcn.CalendarOptions{
			Name:           "demo-range",
			Mode:           shadcn.CalendarRange,
			NumberOfMonths: 2,
			RangeEnd:       time.Date(2026, time.February, 9, 0, 0, 0, 0, time.UTC),
			DayLabel:       func(d time.Time) string { return d.Format("Jan 2, 2006") },
			StartMonth:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			EndMonth:       time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
			MonthURL:       monthURL("date-range-picker"),
		},
		month,
		time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC),
		html.Class("border-0"),
	)
}

// june2026 is the month the date picker recipes show; they stay in it, so
// their month buttons are disabled and need no fragments.
var june2026 = time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)

// juneCalendar is the select-mode Calendar of a date picker recipe, named
// name, bounded to June 2026.
func juneCalendar(name string, opts shadcn.CalendarOptions) mx.Component {
	opts.Name = name
	opts.StartMonth, opts.EndMonth = june2026, june2026
	return shadcn.CalendarWith(opts, june2026, time.Time{}, html.Class("border-0"))
}

// DatePickerBasic is upstream's Basic recipe: the date picker in a Field,
// labelled by a FieldLabel that points at the trigger.
func DatePickerBasic() mx.Component {
	return shadcn.Field("", html.Class("w-[280px]"),
		shadcn.FieldLabelFor("basic-date-trigger", "Date of appointment"),
		shadcn.Popover(
			shadcn.PopoverTrigger("basic-date",
				html.ID("basic-date-trigger"),
				html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" justify-start text-left font-normal"),
				shadcn.CalendarText("basic", html.DataAttr("empty", "true"), "Select date"),
			),
			shadcn.PopoverContent("basic-date", "", html.Class("w-auto p-0"),
				juneCalendar("basic", shadcn.CalendarOptions{
					DayLabel: func(d time.Time) string { return d.Format("January 2, 2006") },
				}),
			),
		),
	)
}

// DatePickerTime is upstream's Time Picker recipe: the date picker next to
// a native time input, each submitting its own field.
func DatePickerTime() mx.Component {
	return html.DivClass("flex gap-4",
		shadcn.Field("", html.Class("w-40"),
			shadcn.FieldLabelFor("time-date-trigger", "Date"),
			shadcn.Popover(
				shadcn.PopoverTrigger("time-date",
					html.ID("time-date-trigger"),
					html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" justify-between font-normal"),
					shadcn.CalendarText("time-date", html.DataAttr("empty", "true"), "Select date"),
				),
				shadcn.PopoverContent("time-date", "", html.Class("w-auto overflow-hidden p-0"),
					juneCalendar("time-date", shadcn.CalendarOptions{
						DayLabel: func(d time.Time) string { return d.Format("01/02/2006") },
					}),
				),
			),
		),
		shadcn.Field("", html.Class("w-32"),
			shadcn.FieldLabelFor("time-picker", "Time"),
			shadcn.Input(html.Type("time"), html.ID("time-picker"), html.Name("time"), html.Step("1"), html.Value("10:30:00"),
				html.Class("appearance-none bg-background [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none")),
		),
	)
}

// DatePickerRTL is upstream's RTL recipe, in Arabic: the calendar follows the
// text direction, its arrow keys included, and its names come from the
// localization hooks.
func DatePickerRTL() mx.Component {
	arMonths := [12]string{"يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو",
		"يوليو", "أغسطس", "سبتمبر", "أكتوبر", "نوفمبر", "ديسمبر"}
	arDays := [7]string{"ح", "ن", "ث", "ر", "خ", "ج", "س"} // indexed by time.Weekday
	return html.Div(html.Dir("rtl"), html.Lang("ar"),
		shadcn.Popover(
			shadcn.PopoverTrigger("rtl-date",
				html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[280px] justify-start text-start font-normal"),
				shadcn.CalendarText("rtl", html.DataAttr("empty", "true"), "اختر تاريخًا"),
			),
			shadcn.PopoverContent("rtl-date", "", html.Class("w-auto p-0"),
				juneCalendar("rtl", shadcn.CalendarOptions{
					WeekStart:   time.Saturday,
					WeekdayName: func(wd time.Weekday) string { return arDays[wd] },
					MonthLabel:  func(m time.Time) string { return arMonths[m.Month()-1] + m.Format(" 2006") },
					DayLabel: func(d time.Time) string {
						return d.Format("2 ") + arMonths[d.Month()-1] + d.Format(" 2006")
					},
				}),
			),
		),
	)
}

// DatePickerInput is upstream's Input recipe: a text input that takes a typed
// 2006-01-02 date and shows a picked one, with the calendar behind a button
// in the input group. The input carries data-calendar-text, so a pick writes
// its value, and its onchange hands a typed date to calendarPick.
func DatePickerInput() mx.Component {
	return shadcn.Field("", html.Class("w-[280px]"),
		shadcn.FieldLabelFor("input-date-text", "Subscription date"),
		shadcn.InputGroup(
			shadcn.InputGroupInput(html.ID("input-date-text"), html.Placeholder("2026-06-01"),
				html.DataAttr("calendar-text", "input-date"),
				html.OnChange(`var v=this.value.trim();calendarPick('input-date',/^\d{4}-\d{2}-\d{2}$/.test(v)?v:'',v)`)),
			shadcn.InputGroupAddon(shadcn.InputGroupAddonInlineEnd,
				shadcn.InputGroupButton("", shadcn.InputGroupButtonIconXS,
					html.PopoverTarget("input-date"), html.Attrib("aria-label", "Select date"),
					navIcon("M8 2v4", "M16 2v4", "M5 4h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z", "M3 10h18"),
				),
			),
		),
		shadcn.PopoverContent("input-date", "", html.Class("w-auto p-0"),
			juneCalendar("input-date", shadcn.CalendarOptions{}),
		),
	)
}

// DatePickerPresets is the old With Presets recipe: a select of relative
// days next to the calendar, each picked with calendarPick. "Today" is a
// fixed date so the pre-rendered gallery stays deterministic.
func DatePickerPresets() mx.Component {
	today := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	label := func(d time.Time) string { return d.Format("January 2, 2006") }
	presets := []any{
		html.OnChange(`calendarPick('presets',this.value,this.selectedOptions[0].dataset.label)`),
		shadcn.NativeSelectOption("", "Select"),
	}
	for _, p := range []struct {
		name string
		days int
	}{{"Today", 0}, {"Tomorrow", 1}, {"In 3 days", 3}, {"In a week", 7}} {
		d := today.AddDate(0, 0, p.days)
		presets = append(presets, shadcn.NativeSelectOption(d.Format(time.DateOnly), html.DataAttr("label", label(d)), p.name))
	}
	return shadcn.Popover(
		shadcn.PopoverTrigger("presets-date",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[280px] justify-start text-left font-normal"),
			shadcn.CalendarText("presets", html.DataAttr("empty", "true"), "Pick a date"),
		),
		shadcn.PopoverContent("presets-date", "", html.Class("flex w-auto flex-col gap-2 p-2"),
			shadcn.NativeSelect("", presets...),
			juneCalendar("presets", shadcn.CalendarOptions{Today: today, DayLabel: label}),
		),
	)
}
