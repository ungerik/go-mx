// Command shadcn-gallery serves a go-mx rebuild of the shadcn/ui component docs
// (https://ui.shadcn.com/docs/components): a sidebar of every ported component
// and one page per component showing each example as a live preview next to the
// Go source that produced it.
//
// Tailwind v4 is supplied by the @tailwindcss/browser CDN build, which compiles
// the shadcn theme tokens and scans the rendered DOM at runtime — so the gallery
// needs no node toolchain or build step. An internet connection is required for
// the CDN script to load.
//
// The same executable either serves the gallery over HTTP or writes it as static
// HTML files:
//
//	go run ./cmd/shadcn-gallery                      # serve on http://localhost:8080
//	go run ./cmd/shadcn-gallery -out ./dist          # write static files to ./dist, then exit
//	go run ./cmd/shadcn-gallery -static-highlight …   # either mode, highlighting the Code tab server-side
//	go run ./cmd/shadcn-gallery -out ./dist -base /go-mx/gallery   # static export for a URL sub-path
//
// In both modes -static-highlight chooses server-side (highlight package) over
// client-side (Shiki) Code-tab highlighting. The static output (-out) links pages
// with root-absolute URLs, so serve the directory from a web root (e.g.
// `python3 -m http.server` inside it); -base prefixes those links for hosting
// under a URL sub-path such as a GitHub Pages project page.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ungerik/go-mx/cmd/shadcn-gallery/examples"
)

// docs is the ordered component catalog, alphabetical like the shadcn/ui docs.
// Stage 2 grows this to all 35 shipped components.
func docs() []ComponentDoc {
	return []ComponentDoc{
		{
			Slug:        "accordion",
			Title:       "Accordion",
			Description: "A vertically stacked set of interactive headings that each reveal a section of content.",
			Examples: []Example{
				{Name: "Demo", Func: examples.AccordionDemo},
			},
		},
		{
			Slug:        "alert",
			Title:       "Alert",
			Description: "Displays a callout for user attention.",
			Examples: []Example{
				{Name: "Default", Func: examples.AlertDefault},
				{Name: "Destructive", Func: examples.AlertDestructive},
			},
		},
		{
			Slug:        "alert-dialog",
			Title:       "Alert Dialog",
			Description: "A modal dialog that interrupts the user with important content and expects a response.",
			Examples: []Example{
				{Name: "Demo", Func: examples.AlertDialogDemo},
			},
		},
		{
			Slug:        "aspect-ratio",
			Title:       "Aspect Ratio",
			Description: "Displays content within a desired ratio.",
			Examples: []Example{
				{Name: "Demo", Func: examples.AspectRatioDemo},
			},
		},
		{
			Slug:        "avatar",
			Title:       "Avatar",
			Description: "An image element with a fallback for representing the user.",
			Examples: []Example{
				{Name: "Default", Func: examples.AvatarDemo},
				{Name: "Fallback", Func: examples.AvatarFallbackDemo},
			},
		},
		{
			Slug:        "badge",
			Title:       "Badge",
			Description: "Displays a badge or a component that looks like a badge.",
			Examples: []Example{
				{Name: "Default", Func: examples.BadgeDefault},
				{Name: "Secondary", Func: examples.BadgeSecondary},
				{Name: "Destructive", Func: examples.BadgeDestructive},
				{Name: "Outline", Func: examples.BadgeOutline},
				{Name: "All variants", Func: examples.BadgeRow},
			},
		},
		{
			Slug:        "breadcrumb",
			Title:       "Breadcrumb",
			Description: "Displays the path to the current resource using a hierarchy of links.",
			Examples: []Example{
				{Name: "Default", Func: examples.BreadcrumbDemo},
				{Name: "With ellipsis", Func: examples.BreadcrumbWithEllipsis},
			},
		},
		{
			Slug:        "button",
			Title:       "Button",
			Description: "Displays a button or a component that looks like a button.",
			Examples: []Example{
				{Name: "Default", Func: examples.ButtonDefault},
				{Name: "Secondary", Func: examples.ButtonSecondary},
				{Name: "Destructive", Func: examples.ButtonDestructive},
				{Name: "Outline", Func: examples.ButtonOutline},
				{Name: "Ghost", Func: examples.ButtonGhost},
				{Name: "Link", Func: examples.ButtonLink},
				{Name: "Sizes", Func: examples.ButtonSizes},
				{Name: "Disabled", Func: examples.ButtonDisabled},
			},
		},
		{
			Slug:        "calendar",
			Title:       "Calendar",
			Description: "A date field component that allows users to enter and edit date.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CalendarDemo},
			},
		},
		{
			Slug:        "card",
			Title:       "Card",
			Description: "Displays a card with header, content, and footer.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CardDemo},
			},
		},
		{
			Slug:        "carousel",
			Title:       "Carousel",
			Description: "A carousel with motion and swipe built using CSS scroll-snap.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CarouselDemo},
			},
		},
		{
			Slug:        "checkbox",
			Title:       "Checkbox",
			Description: "A control that allows the user to toggle between checked and not checked.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CheckboxDemo},
				{Name: "Checked", Func: examples.CheckboxChecked},
				{Name: "Disabled", Func: examples.CheckboxDisabled},
			},
		},
		{
			Slug:        "collapsible",
			Title:       "Collapsible",
			Description: "An interactive component which expands and collapses content.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CollapsibleDemo},
			},
		},
		{
			Slug:        "combobox",
			Title:       "Combobox",
			Description: "Autocomplete input and command palette with a list of suggestions — a Popover composed with a Command.",
			Examples: []Example{
				{Name: "Demo", Func: examples.ComboboxDemo},
			},
		},
		{
			Slug:        "command",
			Title:       "Command",
			Description: "Fast, composable, unstyled command menu — type to filter.",
			Examples: []Example{
				{Name: "Demo", Func: examples.CommandDemo},
			},
		},
		{
			Slug:        "context-menu",
			Title:       "Context Menu",
			Description: "Displays a menu located at the pointer, triggered by a right-click.",
			Examples: []Example{
				{Name: "Demo", Func: examples.ContextMenuDemo},
			},
		},
		{
			Slug:        "data-table",
			Title:       "Data Table",
			Description: "Powerful table and datagrids — Table composed with a filter, column menu and pagination.",
			Examples: []Example{
				{Name: "Demo", Func: examples.DataTableDemo},
			},
		},
		{
			Slug:        "date-picker",
			Title:       "Date Picker",
			Description: "A date picker component with range and presets — a Popover composed with a Calendar.",
			Examples: []Example{
				{Name: "Demo", Func: examples.DatePickerDemo},
			},
		},
		{
			Slug:        "dialog",
			Title:       "Dialog",
			Description: "A window overlaid on either the primary window or another dialog window, rendering the content underneath inert.",
			Examples: []Example{
				{Name: "Demo", Func: examples.DialogDemo},
			},
		},
		{
			Slug:        "drawer",
			Title:       "Drawer",
			Description: "A drawer that slides up from the bottom and can be dragged down to dismiss.",
			Examples: []Example{
				{Name: "Demo", Func: examples.DrawerDemo},
			},
		},
		{
			Slug:        "dropdown-menu",
			Title:       "Dropdown Menu",
			Description: "Displays a menu to the user — such as a set of actions or functions — triggered by a button.",
			Examples: []Example{
				{Name: "Demo", Func: examples.DropdownMenuDemo},
			},
		},
		{
			Slug:        "form",
			Title:       "Form",
			Description: "Building forms with labels, descriptions, and validation messages.",
			Examples: []Example{
				{Name: "Demo", Func: examples.FormDemo},
				{Name: "With error", Func: examples.FormWithError},
			},
		},
		{
			Slug:        "hover-card",
			Title:       "Hover Card",
			Description: "For sighted users to preview content available behind a link.",
			Examples: []Example{
				{Name: "Demo", Func: examples.HoverCardDemo},
			},
		},
		{
			Slug:        "input",
			Title:       "Input",
			Description: "Displays a form input field or a component that looks like an input field.",
			Examples: []Example{
				{Name: "Default", Func: examples.InputDefault},
				{Name: "Disabled", Func: examples.InputDisabled},
				{Name: "File", Func: examples.InputFile},
				{Name: "With label", Func: examples.InputWithLabel},
			},
		},
		{
			Slug:        "input-otp",
			Title:       "Input OTP",
			Description: "Accessible one-time password component with copy paste functionality.",
			Examples: []Example{
				{Name: "Default", Func: examples.InputOTPDemo},
				{Name: "With separator", Func: examples.InputOTPWithSeparator},
			},
		},
		{
			Slug:        "label",
			Title:       "Label",
			Description: "Renders an accessible label associated with controls.",
			Examples: []Example{
				{Name: "Demo", Func: examples.LabelDemo},
			},
		},
		{
			Slug:        "menubar",
			Title:       "Menubar",
			Description: "A visually persistent menu common in desktop applications that provides quick access to a consistent set of commands.",
			Examples: []Example{
				{Name: "Demo", Func: examples.MenubarDemo},
			},
		},
		{
			Slug:        "navigation-menu",
			Title:       "Navigation Menu",
			Description: "A collection of links for navigating websites.",
			Examples: []Example{
				{Name: "Demo", Func: examples.NavigationMenuDemo},
			},
		},
		{
			Slug:        "pagination",
			Title:       "Pagination",
			Description: "Pagination with page navigation, next and previous links.",
			Examples: []Example{
				{Name: "Demo", Func: examples.PaginationDemo},
			},
		},
		{
			Slug:        "popover",
			Title:       "Popover",
			Description: "Displays rich content in a portal, triggered by a button.",
			Examples: []Example{
				{Name: "Demo", Func: examples.PopoverDemo},
			},
		},
		{
			Slug:        "progress",
			Title:       "Progress",
			Description: "Displays an indicator showing the completion progress of a task.",
			Examples: []Example{
				{Name: "Demo", Func: examples.ProgressDemo},
			},
		},
		{
			Slug:        "radio-group",
			Title:       "Radio Group",
			Description: "A set of checkable buttons where no more than one can be checked at a time.",
			Examples: []Example{
				{Name: "Demo", Func: examples.RadioGroupDemo},
			},
		},
		{
			Slug:        "resizable",
			Title:       "Resizable",
			Description: "Accessible resizable panel groups and layouts with keyboard support.",
			Examples: []Example{
				{Name: "Demo", Func: examples.ResizableDemo},
			},
		},
		{
			Slug:        "scroll-area",
			Title:       "Scroll Area",
			Description: "Augments native scroll functionality for custom, cross-browser styling.",
			Examples: []Example{
				{Name: "Demo", Func: examples.ScrollAreaDemo},
			},
		},
		{
			Slug:        "select",
			Title:       "Select",
			Description: "Displays a list of options for the user to pick from, triggered by a button.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SelectDemo},
			},
		},
		{
			Slug:        "separator",
			Title:       "Separator",
			Description: "Visually or semantically separates content.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SeparatorDemo},
			},
		},
		{
			Slug:        "sheet",
			Title:       "Sheet",
			Description: "Extends the Dialog component to display content that complements the main content of the screen.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SheetDemo},
			},
		},
		{
			Slug:        "sidebar",
			Title:       "Sidebar",
			Description: "A composable, themeable and customizable sidebar that collapses to icons.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SidebarDemo},
			},
		},
		{
			Slug:        "skeleton",
			Title:       "Skeleton",
			Description: "Use to show a placeholder while content is loading.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SkeletonDemo},
			},
		},
		{
			Slug:        "slider",
			Title:       "Slider",
			Description: "An input where the user selects a value from within a given range.",
			Examples: []Example{
				{Name: "Default", Func: examples.SliderDemo},
				{Name: "Range", Func: examples.SliderRange},
			},
		},
		{
			Slug:        "sonner",
			Title:       "Sonner",
			Description: "An opinionated toast component — imperative toast() pushed into a Toaster region.",
			Examples: []Example{
				{Name: "Demo", Func: examples.SonnerDemo},
			},
		},
		{
			Slug:        "switch",
			Title:       "Switch",
			Description: "A control that allows the user to toggle between checked and not checked.",
			Examples: []Example{
				{Name: "Default", Func: examples.SwitchDemo},
				{Name: "Disabled", Func: examples.SwitchDisabled},
			},
		},
		{
			Slug:        "table",
			Title:       "Table",
			Description: "A responsive table component.",
			Examples: []Example{
				{Name: "Demo", Func: examples.TableDemo},
			},
		},
		{
			Slug:        "tabs",
			Title:       "Tabs",
			Description: "A set of layered sections of content, known as tab panels, that are displayed one at a time.",
			Examples: []Example{
				{Name: "Demo", Func: examples.TabsDemo},
			},
		},
		{
			Slug:        "textarea",
			Title:       "Textarea",
			Description: "Displays a form textarea or a component that looks like a textarea.",
			Examples: []Example{
				{Name: "Default", Func: examples.TextareaDefault},
				{Name: "Disabled", Func: examples.TextareaDisabled},
				{Name: "With label", Func: examples.TextareaWithLabel},
			},
		},
		{
			Slug:        "toggle",
			Title:       "Toggle",
			Description: "A two-state button that can be either on or off.",
			Examples: []Example{
				{Name: "Default", Func: examples.ToggleDemo},
				{Name: "Outline", Func: examples.ToggleOutline},
				{Name: "Sizes", Func: examples.ToggleSizes},
				{Name: "Disabled", Func: examples.ToggleDisabled},
			},
		},
		{
			Slug:        "toggle-group",
			Title:       "Toggle Group",
			Description: "A set of two-state buttons that can be toggled on or off.",
			Examples: []Example{
				{Name: "Multiple", Func: examples.ToggleGroupDemo},
				{Name: "Single", Func: examples.ToggleGroupSingleDemo},
			},
		},
		{
			Slug:        "tooltip",
			Title:       "Tooltip",
			Description: "A popup that displays information related to an element when it receives focus or the mouse hovers over it.",
			Examples: []Example{
				{Name: "Demo", Func: examples.TooltipDemo},
			},
		},
	}
}

func main() {
	addr := flag.String("addr", ":8080", "listen address when serving")
	out := flag.String("out", "",
		"if set, write the gallery as static HTML files into this directory and exit, instead of serving")
	base := flag.String("base", "",
		"URL sub-path the static export (-out) is hosted under, e.g. /go-mx/gallery for a GitHub Pages project page; prefixed to every in-gallery link")
	flag.BoolVar(&staticHighlight, "static-highlight", false,
		"highlight the Code tab server-side with the highlight package instead of client-side Shiki")
	flag.Parse()

	linkBase = strings.TrimRight(*base, "/")

	reg := NewRegistry(docs())

	if *out != "" {
		if err := writeStatic(reg, *out); err != nil {
			log.Fatal(err)
		}
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		page(reg, "", "Components", indexContent(reg)).HandleHTTP(w, r)
	})
	mux.HandleFunc("GET /components/{slug}", func(w http.ResponseWriter, r *http.Request) {
		d := reg.Lookup(r.PathValue("slug"))
		if d == nil {
			http.NotFound(w, r)
			return
		}
		page(reg, d.Slug, d.Title, componentContent(d)).HandleHTTP(w, r)
	})

	fmt.Printf("listening on %s\n", *addr)
	fmt.Printf("open http://localhost%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
