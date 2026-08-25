package shadcn

import (
	"cmp"
	"context"

	"github.com/ungerik/go-mx"
)

// Labels holds every user-visible string that a component of this package
// renders on its own, without the caller passing it: the screen-reader text of
// the built-in dialog/sheet close button, the pagination captions, the
// ellipsis "more" texts, and the default aria-labels of the landmark <nav>s and
// the icon-only buttons.
//
// The strings are resolved from the render context, not baked into the element
// at construction time, so one process can serve several languages
// concurrently: install the request's Labels with [ContextWithLabels] and every
// component rendered with that context picks them up, with no change to any
// call site. [DefaultLabels] is used when a context carries none.
//
// Like [CalendarOptions] this ships the hooks, not the translations: the
// package carries the English defaults only and stays free of any locale
// database.
//
// Fields may be left empty: each one falls back to the corresponding
// [DefaultLabels] field and then to the package's built-in English, so a
// partial translation renders a default for the strings it does not cover
// instead of an empty label.
type Labels struct {
	// BreadcrumbNav is the aria-label of the [Breadcrumb] <nav> landmark.
	BreadcrumbNav string
	// BreadcrumbEllipsis is the screen-reader text inside [BreadcrumbEllipsis].
	BreadcrumbEllipsis string

	// CarouselPrevious and CarouselNext are the aria-labels of the
	// icon-only [CarouselPrevious] and [CarouselNext] buttons.
	CarouselPrevious string
	CarouselNext     string

	// DialogClose is the screen-reader text of the built-in top-right close
	// button rendered by [DialogContent] and [SheetContent].
	DialogClose string

	// NavigationMenuNav is the aria-label of the [NavigationMenu] <nav>
	// landmark.
	NavigationMenuNav string

	// PaginationNav is the aria-label of the [Pagination] <nav> landmark.
	PaginationNav string
	// PaginationPrevious and PaginationNext are the visible captions of the
	// [PaginationPrevious] and [PaginationNext] links.
	PaginationPrevious string
	PaginationNext     string
	// PaginationPreviousPage and PaginationNextPage are the aria-labels of
	// those same links, which name the destination rather than repeat the
	// caption.
	PaginationPreviousPage string
	PaginationNextPage     string
	// PaginationEllipsis is the screen-reader text inside
	// [PaginationEllipsis].
	PaginationEllipsis string

	// SidebarTrigger is the aria-label of the icon-only [SidebarTrigger]
	// button.
	SidebarTrigger string

	// Spinner is the aria-label of the [Spinner] status icon.
	Spinner string

	// FormClear is the caption of the checkbox that a reflected form renders
	// for a nullable field to clear it (see [FormDecider]).
	FormClear string
	// FormValueNotAvailable is the caption of the disabled placeholder option
	// a reflected select renders when the field's current value is not in the
	// per-request option list. It is deliberately generic — the filtered-out
	// value must not be echoed into the markup — so it carries no data and is
	// safe to translate.
	FormValueNotAvailable string
}

// DefaultLabels holds the strings used when the render context carries no
// [Labels]. It is also the per-field fallback for an installed Labels that
// leaves a field empty, and it starts out as the English [englishLabels].
//
// Assign to it to change the built-in labels for the whole program, and use
// [ContextWithLabels] to vary them per request. Assign during initialization,
// before the first render: it is a plain package variable read by every render,
// like [mx.AsComponent] and [mx.AsAttribs], so writing it while requests are in
// flight is a data race. A partial assignment is safe — fields left empty fall
// through to englishLabels rather than rendering empty.
var DefaultLabels = englishLabels

// englishLabels holds the built-in English strings, matching shadcn/ui
// upstream. It is the frozen last resort of the fallback chain, so that a
// partial [DefaultLabels] assignment can never strip a control of its
// accessible name.
var englishLabels = Labels{
	BreadcrumbNav:          "breadcrumb",
	BreadcrumbEllipsis:     "More",
	CarouselPrevious:       "Previous slide",
	CarouselNext:           "Next slide",
	DialogClose:            "Close",
	NavigationMenuNav:      "Main",
	PaginationNav:          "pagination",
	PaginationPrevious:     "Previous",
	PaginationNext:         "Next",
	PaginationPreviousPage: "Go to previous page",
	PaginationNextPage:     "Go to next page",
	PaginationEllipsis:     "More pages",
	SidebarTrigger:         "Toggle Sidebar",
	Spinner:                "Loading",
	FormClear:              "clear",
	FormValueNotAvailable:  "(current value not available)",
}

// ctxKeyLabels is the unexported render-context key under which
// [ContextWithLabels] stores the active [Labels].
type ctxKeyLabels struct{}

// ContextWithLabels returns a derived context that carries labels, so that the
// components rendered with it use those strings instead of [DefaultLabels].
// Wire it once per request to serve several languages from one process:
//
//	next.ServeHTTP(w, r.WithContext(shadcn.ContextWithLabels(r.Context(), labelsFor(lang))))
func ContextWithLabels(ctx context.Context, labels Labels) context.Context {
	return context.WithValue(ctx, ctxKeyLabels{}, labels)
}

// LabelsFromContext returns the [Labels] installed by [ContextWithLabels] with
// every empty field filled from [DefaultLabels], or DefaultLabels itself if the
// context carries none. Callers therefore always get a complete set, whether
// they render a component of this package or their own.
func LabelsFromContext(ctx context.Context) Labels {
	labels, _ := ctx.Value(ctxKeyLabels{}).(Labels)
	return labels.orDefaults()
}

// orDefaults returns l with every empty field replaced by the corresponding
// [DefaultLabels] field, and then by the built-in englishLabels field, so a
// partial translation still renders a label everywhere. The zero Labels — what
// the type assertion in [LabelsFromContext] yields for a context without any —
// therefore becomes DefaultLabels.
//
// The englishLabels step is what makes a partial DefaultLabels assignment safe:
// without it, assigning DefaultLabels a struct that sets only some fields would
// make every other field fall back to that same empty field, and controls whose
// only accessible name is an aria-label would render aria-label="".
//
// Written out field by field rather than by reflection because this runs for
// every built-in label of every render; TestLabelsOrDefaultsCoversAllFields
// fails if a new field is not added here.
func (l Labels) orDefaults() Labels {
	return Labels{
		BreadcrumbNav:          cmp.Or(l.BreadcrumbNav, DefaultLabels.BreadcrumbNav, englishLabels.BreadcrumbNav),
		BreadcrumbEllipsis:     cmp.Or(l.BreadcrumbEllipsis, DefaultLabels.BreadcrumbEllipsis, englishLabels.BreadcrumbEllipsis),
		CarouselPrevious:       cmp.Or(l.CarouselPrevious, DefaultLabels.CarouselPrevious, englishLabels.CarouselPrevious),
		CarouselNext:           cmp.Or(l.CarouselNext, DefaultLabels.CarouselNext, englishLabels.CarouselNext),
		DialogClose:            cmp.Or(l.DialogClose, DefaultLabels.DialogClose, englishLabels.DialogClose),
		NavigationMenuNav:      cmp.Or(l.NavigationMenuNav, DefaultLabels.NavigationMenuNav, englishLabels.NavigationMenuNav),
		PaginationNav:          cmp.Or(l.PaginationNav, DefaultLabels.PaginationNav, englishLabels.PaginationNav),
		PaginationPrevious:     cmp.Or(l.PaginationPrevious, DefaultLabels.PaginationPrevious, englishLabels.PaginationPrevious),
		PaginationNext:         cmp.Or(l.PaginationNext, DefaultLabels.PaginationNext, englishLabels.PaginationNext),
		PaginationPreviousPage: cmp.Or(l.PaginationPreviousPage, DefaultLabels.PaginationPreviousPage, englishLabels.PaginationPreviousPage),
		PaginationNextPage:     cmp.Or(l.PaginationNextPage, DefaultLabels.PaginationNextPage, englishLabels.PaginationNextPage),
		PaginationEllipsis:     cmp.Or(l.PaginationEllipsis, DefaultLabels.PaginationEllipsis, englishLabels.PaginationEllipsis),
		SidebarTrigger:         cmp.Or(l.SidebarTrigger, DefaultLabels.SidebarTrigger, englishLabels.SidebarTrigger),
		Spinner:                cmp.Or(l.Spinner, DefaultLabels.Spinner, englishLabels.Spinner),
		FormClear:              cmp.Or(l.FormClear, DefaultLabels.FormClear, englishLabels.FormClear),
		FormValueNotAvailable:  cmp.Or(l.FormValueNotAvailable, DefaultLabels.FormValueNotAvailable, englishLabels.FormValueNotAvailable),
	}
}

// labelText returns a [mx.Component] that renders the Labels field selected by
// get as escaped text, resolved from the render context.
func labelText(get func(Labels) string) mx.Component {
	return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
		return w.EscapeText(get(LabelsFromContext(ctx)))
	})
}

// labelAttrib is an [mx.Attrib] whose value is the Labels field selected by get,
// resolved from the render context. This is what mx.Attrib.AttribValue's
// context parameter exists for.
type labelAttrib struct {
	name string
	get  func(Labels) string
}

var _ mx.Attrib = labelAttrib{}

// AttribName returns the static attribute name.
func (a labelAttrib) AttribName() string { return a.name }

// AttribValue returns the label for the context's [Labels] and a nil error.
func (a labelAttrib) AttribValue(ctx context.Context) (string, error) {
	return a.get(LabelsFromContext(ctx)), nil
}

// labelAriaLabel returns an aria-label [mx.Attrib] resolved from the render
// context's [Labels].
func labelAriaLabel(get func(Labels) string) mx.Attrib {
	return labelAttrib{name: "aria-label", get: get}
}
