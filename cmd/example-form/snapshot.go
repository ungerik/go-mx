package main

import (
	"net/http"
	"net/url"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/internal/shadcntheme"
	"github.com/ungerik/go-mx/internal/snapshot"
	"github.com/ungerik/go-mx/shadcn"
)

// writeSnapshot writes the form as static pages for the documentation site:
// index.html is the form as a GET shows it, and invalid.html the answer to a
// POST that fails validation, with the errors next to their fields. Both
// forms come from handler, so they are the markup the server serves.
//
// The server answers with the bare <form>, as a handler embedded in an
// application's own layout would. The snapshot puts it into a page with the
// shadcn theme compiled by the Tailwind browser build, so it shows the way
// the classes the field decider emits are meant to look, and it keeps the
// form from submitting, because a static host has nothing to submit to.
func writeSnapshot(snap *snapshot.Options, handler http.Handler) error {
	form, err := snapshot.Get(handler, "/admin/profile")
	if err != nil {
		return err
	}
	if err := writeFormPage(snap, "index.html", form); err != nil {
		return err
	}

	form, err = snapshot.Post(handler, "/admin/profile", invalidSubmission())
	if err != nil {
		return err
	}
	return writeFormPage(snap, "invalid.html", form)
}

// writeFormPage writes the page holding form to name.
func writeFormPage(snap *snapshot.Options, name, form string) error {
	form, err := snapshot.Edit(form, `<form method="post"`, `<form onsubmit="return false" method="post"`, 1)
	if err != nil {
		return err
	}
	tab := func(href, label string) mx.Component {
		variant := shadcn.ButtonOutline
		if href == name {
			variant = shadcn.ButtonDefault
		}
		return html.AHRef(href, html.Class(shadcn.ButtonClasses(variant, shadcn.SizeSM)), label)
	}
	doc := html.NewDocument("go-mx form example",
		html.Main(html.Class("mx-auto max-w-2xl px-6 py-10"),
			html.H1Class("text-2xl font-semibold tracking-tight", "Reflected form"),
			html.PClass("mt-2 text-sm text-muted-foreground",
				"mx.ReflectFormHandler renders a Profile struct through shadcn.FieldDecider. ",
				"Submitting needs the live server, so the second page shows what it answers to an invalid submission."),
			html.DivClass("mt-4 mb-8 flex gap-2",
				tab("index.html", "Form"),
				tab("invalid.html", "After an invalid submission"),
			),
			// The bare form leaves the spacing between its fields to the
			// layout it is embedded in.
			html.DivClass("[&>form]:flex [&>form]:flex-col [&>form]:gap-6", html.Raw(form)),
		),
	)
	doc.Meta = map[string]string{"viewport": "width=device-width, initial-scale=1"}
	doc.HeadCustom = shadcntheme.Head()
	page, err := snapshot.Render(doc)
	if err != nil {
		return err
	}
	return snap.Write(name, page)
}

// invalidSubmission is a POST that fails the per-field validation chain: an
// empty required Name and an Age below its min. The handler answers it with
// the form again, each error next to its field, and never calls Save.
func invalidSubmission() url.Values {
	form := url.Values{}
	set := func(name string, values ...string) {
		form.Set(mx.PresentSentinelName(mx.FieldPath(name)), "1")
		form[name] = values
	}
	set("Name", "")
	set("Email", "ada@example.com")
	set("Age", "7")
	return form
}
