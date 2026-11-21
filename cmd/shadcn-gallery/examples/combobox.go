package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ComboboxDemo is upstream's Combobox demo: a button showing the picked
// framework that opens a searchable list. Type to filter, ArrowDown/ArrowUp
// and Enter to pick; picking the picked framework again unpicks it. The
// picked value is submitted as "framework" by an enclosing form.
func ComboboxDemo() mx.Component {
	return shadcn.Combobox(shadcn.ComboboxOptions{
		Name: "framework",
		Options: []mx.NamedOption{
			{Value: "next.js", Name: "Next.js"},
			{Value: "sveltekit", Name: "SvelteKit"},
			{Value: "nuxt.js", Name: "Nuxt.js"},
			{Value: "remix", Name: "Remix"},
			{Value: "astro", Name: "Astro"},
		},
		Placeholder:       "Select framework...",
		SearchPlaceholder: "Search framework...",
		Empty:             "No framework found.",
	}, html.Class("w-[200px]"))
}

// bookingAccounts is a small chart of accounts: too many entries to scroll
// through comfortably, found by typing part of the number or the name.
var bookingAccounts = []mx.NamedOption{
	{Value: "1000", Name: "1000 Cash"},
	{Value: "1200", Name: "1200 Bank"},
	{Value: "1400", Name: "1400 Accounts receivable"},
	{Value: "1600", Name: "1600 Accounts payable"},
	{Value: "1800", Name: "1800 Private withdrawals"},
	{Value: "4200", Name: "4200 Rent"},
	{Value: "4240", Name: "4240 Gas, electricity, water"},
	{Value: "4600", Name: "4600 Advertising"},
	{Value: "4920", Name: "4920 Telephone"},
	{Value: "4930", Name: "4930 Office supplies"},
	{Value: "4970", Name: "4970 Bank charges"},
	{Value: "8400", Name: "8400 Revenue 19% VAT"},
}

// ComboboxForm is a booking form with two comboboxes over the same chart of
// accounts. The debit account is required: submitting without one is
// blocked by the browser, which points at the trigger. The submitted values
// appear below the button; the gallery has no server, so the form's
// onsubmit shows them instead of sending them, which a real page leaves out.
func ComboboxForm() mx.Component {
	return shadcn.Form(html.Class("w-full max-w-sm"),
		html.OnSubmit("event.preventDefault();this.querySelector('output').textContent=new URLSearchParams(new FormData(this)).toString()"),
		shadcn.FieldGroup(
			shadcn.Field("",
				shadcn.FieldLabelFor("debit", "Debit account"),
				shadcn.Combobox(shadcn.ComboboxOptions{
					Name:              "debit",
					Options:           bookingAccounts,
					Placeholder:       "Select account...",
					SearchPlaceholder: "Number or name...",
					Empty:             "No account found.",
					Required:          true,
				}, html.Attrib("aria-describedby", "debit-help")),
				shadcn.FieldDescription(html.ID("debit-help"), "Type 4200 or rent to find the rent account."),
			),
			shadcn.Field("",
				shadcn.FieldLabelFor("credit", "Credit account"),
				shadcn.Combobox(shadcn.ComboboxOptions{
					Name:              "credit",
					Options:           bookingAccounts,
					Selected:          "1200",
					SearchPlaceholder: "Number or name...",
					Empty:             "No account found.",
				}),
			),
			shadcn.Field("",
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Type("submit"), "Book"),
				html.Output(html.Class("text-muted-foreground font-mono text-sm")),
			),
		),
	)
}
