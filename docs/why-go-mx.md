---
title: Why go-mx
---

# Why go-mx — server-rendered HTML, and where it fits

An honest look at *why* you would render HTML on the server in Go in 2026,
what `go-mx` does well, what it does badly, and the use cases it is actually
built for. Sources are listed at the [bottom](#sources).

## The backdrop: the pendulum is swinging back

For a decade the default answer to "how do I build a web UI" was "a SPA in
React/Vue plus a JSON API." That default is now openly questioned — not for
nostalgic reasons, but over **complexity cost and security surface**. The
argument practitioners make today is: every line of client JavaScript brings
build tooling, `npm audit` noise, and another supply-chain risk, so cutting
the payload often improves performance *and* security at once. The slogan
that has emerged is **"HTML by default, JS only where it buys you
something"** [^ssr-pendulum][^newstack].

What gives this teeth right now is the **supply chain**. The September 2025
npm attack compromised 18 packages pulling roughly **2.6 billion
downloads/week** (eventually 200+ packages) via a single phishing email to a
maintainer, and shipped the **Shai-hulud worm** that steals cloud tokens and
self-propagates [^palo-alto][^armorcode]. npm is structurally exposed because
of three properties: **deep transitive dependency trees, install-time code
execution (`postinstall`), and permissive semver** — a compromised package
cascades automatically into thousands of `node_modules` folders.

A server-rendered Go stack has none of those three properties. `go get` runs
no install-time scripts, versions are pinned and checksum-verified (`go.sum`
plus the checksum database), and the dependency graph is comparatively
shallow. The attack class that defined frontend security in 2025 essentially
does not exist here.

## What server-rendered UI buys you (the general case)

| Advantage                            | Why it matters                              |
|--------------------------------------|---------------------------------------------|
| Faster First Contentful Paint        | Pre-rendered HTML shows content fast        |
| SEO & social previews                | Crawlers and social cards get real markup   |
| Smaller supply chain & build surface | Less client JS, less build tooling to audit |
| Lower client resource use            | Less memory/CPU; works on low-end devices   |
| Architectural simplicity             | One language; no duplicated API layer       |

[^dotcms][^newstack][^spa-seo]

## Where go-mx is strong

go-mx is in the same family as Go's [gomponents][gomponents] and
[templ][templ] — "HTML as type-safe Go, no template language" — but it makes
some distinctive choices:

- **No build step, no codegen.** Markup is ordinary Go function calls
  (`html.Div(html.Class("x"), html.P("hi"))`). templ requires a codegen step
  and its own `.templ` language; go-mx is just Go, so it composes, refactors,
  and type-checks with the rest of your code and works with hot-reload tools
  directly [^templ-vs-gomponents].

- **Zero client runtime *and* zero npm — even for shadcn.** The `shadcn/`
  package does not *wrap* React components; it **ports** them, reproducing the
  Tailwind markup in Go with faithful Go ports of `clsx`, `tailwind-merge` v3,
  and `cva`. You get shadcn-style components without pulling a single npm
  package into your build. This is the supply-chain argument made concrete.

- **Reflection-driven forms are a genuine differentiator.**
  `mx.ReflectFormHandler[T]` turns a struct into a full `http.Handler` —
  rendering, parsing, validation, and load-then-apply — in one call. It is
  **mass-assignment-safe by construction** (POST only parses fields the form
  actually rendered) and runs a richest-first validation chain. For the
  CRUD/admin/line-of-business work that is the bread and butter of
  server-rendered apps, this is exactly the right abstraction. Neither templ
  nor gomponents ships it.

- **Escaping and structural validation are built in.** `CheckedWriter`
  escapes text and attribute values and validates structure (for example,
  void elements cannot have children) as it walks the tree — XSS prevention by
  default, without a separate sanitizer.

- **HTMX as the interactivity layer.** The `hx/` package integrates HTMX so
  you get partial-page updates without a SPA, including conveniences like
  `hx.FieldDecider` wiring `hx-trigger="change"` onto live inputs.

## The honest disadvantages

**Of the server-rendered approach in general:**

- **The server does the work.** Every render is a server request — higher
  CPU/hosting cost, and the server is a bottleneck under load. SPAs amortize:
  after first load, interactions are near-instant on the client [^hostragons].
- **Interactivity has a latency floor.** With HTMX, every interaction is a
  network round-trip. Without measuring request frequency and response time
  you get slow or jittery UIs [^htmx-badly].
- **Cross-cutting UI updates are awkward.** The canonical example, even from
  HTMX's own author, is updating an unread-count badge in one corner when you
  act somewhere else; state that spans structural regions is where hypermedia
  "goes wobbly" [^htmx-when].
- **Rich client experiences are out of scope.** Offline-first, optimistic UI,
  real-time collaborative editing, drag-and-drop canvases, and native-feeling
  mobile gestures genuinely want client-side state and are weak spots.

**Of go-mx specifically:**

- **It is young and partially built.** `web/` and `doc/` are explicitly
  partially implemented, and the TODO list (slice-of-struct fields,
  OOB fragments, richer file upload, options registry) shows the form layer is
  still maturing. templ and gomponents are more battle-tested.
- **You still need Tailwind itself for the shadcn path.** The npm *runtime* is
  gone, but the shadcn components emit Tailwind classes, so you still run a
  Tailwind CSS build (or the CDN). The supply-chain win is real but not total
  on that route.
- **Function-call markup is noisier than JSX/templates** for some, and makes a
  deliberate trade-off: a non-`Component` value passed as a child stringifies to
  escaped text (so it can never inject markup) rather than erroring, which means
  a mistyped child renders as text with no compile-time error.
- **Go-only, smaller ecosystem.** No Figma-to-React handoff, far fewer
  off-the-shelf components than the React world, and your team has to be a Go
  team.
- **Requires Go 1.26** (very recent), which can be a non-starter for some
  shops.

## Optimal use cases

go-mx is a strong fit when most of these are true:

- **Internal tools, admin panels, dashboards, CRUD, and forms-heavy
  line-of-business apps** — `ReflectFormHandler` is purpose-built for this,
  and these apps rarely need SPA-grade interactivity.
- **Content, docs, and marketing sites** where SEO and fast first paint
  matter.
- **You are already a Go shop** and want one language end-to-end, no separate
  frontend build/deploy, and no JS supply chain to audit.
- **Security/compliance pressure** makes minimizing the client-side
  dependency surface a feature, not a preference.
- **HTMX-level interactivity is enough** — page-level and fragment-level
  updates cover the need.

It is the wrong tool when:

- You are building a **genuinely app-like, highly interactive client**
  (editor, design tool, trading terminal, collaborative whiteboard,
  offline-first PWA).
- You have a **dedicated frontend/design org** invested in the React/Figma
  toolchain.
- Your backend **is not Go**, or you need to serve many native client types
  from one JSON API anyway.

## Bottom line

The honest pitch is not "SPAs are dead." It is that the industry
over-applied the SPA default to a huge class of apps — CRUD, admin, content,
internal tools — that never needed it, and paid for the mismatch in build
complexity and supply-chain risk. go-mx is a well-aimed bet on that reality:
**type-safe HTML as plain Go, batteries (shadcn, HTMX, reflected forms)
included, and a near-zero client dependency surface.** Its main risks are its
youth and the inherent ceiling of server-driven interactivity — both
acceptable precisely for the use cases above, and disqualifying outside them.

## Sources

[gomponents]: https://github.com/oderwat/gomponents
[templ]: https://github.com/a-h/templ

[^ssr-pendulum]: ["Server-side rendering never really went away" — Hacker News discussion](https://news.ycombinator.com/item?id=43881502)
[^newstack]: [SPAs and React: You Don't Always Need Server-Side Rendering — The New Stack](https://thenewstack.io/spas-and-react-you-dont-always-need-server-side-rendering/)
[^palo-alto]: [Widespread npm Supply Chain Attack — Palo Alto Networks](https://www.paloaltonetworks.com/blog/cloud-security/npm-supply-chain-attack/)
[^armorcode]: [Inside the September 2025 npm Supply Chain Attack — ArmorCode](https://www.armorcode.com/blog/inside-the-september-2025-npm-supply-chain-attack)
[^dotcms]: [SPAs and Server Side Rendering: A Must, or a Maybe? — dotCMS](https://www.dotcms.com/blog/spas-and-server-side-rendering-a-must-or-a-maybe)
[^spa-seo]: [Why SPAs Still Struggle with SEO — Dev Tech Insights](https://devtechinsights.com/spas-seo-challenges-2025/)
[^templ-vs-gomponents]: [Templ vs Gomponents — Ewen Quimerc'h](https://ewen.quimerch.com/articles/14-templ-vs-gomponents/)
[^hostragons]: [Single-Page Application (SPA) vs Server-Side Rendering (SSR) — Hostragons](https://www.hostragons.com/en/blog/single-page-application-spa/)
[^htmx-badly]: [Why Most Developers Are Using HTMX Badly — Hex Shift](https://hexshift.medium.com/why-most-developers-are-using-htmx-badly-21a01e3223b3)
[^htmx-when]: [When Should You Use Hypermedia? — htmx.org](https://htmx.org/essays/when-to-use-hypermedia/)
