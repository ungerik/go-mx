---
title: Why Magic UI animates with CSS alone
---

# Why Magic UI animates with CSS alone

[Magic UI](https://magicui.design) is a set of React components. Its
animations run in the browser: some through Tailwind keyframes, many through
the [motion](https://motion.dev) library, which springs and staggers elements
from JavaScript, and a few through `Math.random()` and `useId()`. go-mx renders
HTML on the server and sends no client runtime. This page explains how the
`shadcn/magicui` port gets the same effects anyway, and what it gives up to do
so. For the how, see the [how-to guides](how-to.html); for each component's
exact differences, the
[component reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#differences-from-upstream).

## The problem

A server-rendered page has nothing that could run upstream's code:

- **No motion library.** A word that blurs in after the one before it is,
  upstream, a motion `staggerChildren` setting interpreted by JavaScript on
  every frame. Without that script the words would just sit there, or never
  appear if their starting state is invisible.
- **No `Math.random()` at render time in the browser.** Meteors, sparkles and
  light rays get their positions from it when React mounts the component. Done
  on the server per request, it would make every response different, so no
  page with a meteor shower could be cached or compared in a test.
- **No `useId()`.** The SVG components refer to their own masks and patterns
  by id. React hands out ids that are unique in the page; plain HTML has no
  such mechanism.

Shipping a JavaScript runtime would answer all three, and it would also bring
back the bundle, the hydration step and the client-side failure modes go-mx
exists to avoid (see [Why go-mx?](../../why-go-mx.html)).

## The approach

Each piece of upstream's runtime is replaced by something the browser already
does, or by work done once on the server:

```
upstream (React, browser)            port (Go, server + CSS)
-------------------------            ------------------------------------------
motion spring / tween          -->   @keyframes + Tailwind animate-* utility,
                                     springs sampled into a linear() easing
staggerChildren                -->   one element per segment, delay computed
                                     from a CSS variable (--i, the index)
whileInView                    -->   scroll-driven animation:
                                     animation-timeline: view(), in @supports
Math.random()                  -->   seeded generator on the server (Seed)
useId()                        -->   ID option, else mx.UniqueID()
prefers-reduced-motion ignored -->   every animation switched off,
                                     resting style = final state
```

**The animations are one stylesheet.** Every keyframe, `animate-*` utility and
reduced-motion rule lives in `magicui.CSS`, assembled from one file per
component. The components emit only class names and CSS variables. That keeps
the Go side small, but it moves one duty to the app: it has to add
`magicui.CSS` to its Tailwind input, and Tailwind has to see the component
source to emit the utilities (the `@source` line of the
[production guide](how-to.html#build-the-magic-ui-css-for-production)).

**Staggering is arithmetic in CSS.** `TextAnimate` splits its text on the
server into one `<span>` per word, character or line and gives each its index
as `--i`. The shared keyframe's delay is computed from `--i`, the start delay
and the stagger, so segment *i* starts at `Delay + i × Duration / segments`,
the same schedule motion's `staggerChildren` produces.

**Springs become easings.** A motion spring has no CSS keyword, but its curve
can be sampled into a CSS `linear()` easing function, which is what
`Text3DFlip` and `ShinyButton` use. `BorderBeam` moves along the border with
`offset-path: rect(…)` and a keyframe on `offset-distance`, where upstream
animates the position from JavaScript.

**Randomness is drawn once, on the server, from a seed.** `Meteors`,
`LightRays`, `SparklesText`, `WarpBackground`, a glowing `DotPattern` and a
randomly staggered `Text3DFlip` take a `Seed` and draw their positions and
timings from a deterministic generator. The same options give the same
markup, so a page is cacheable and golden-file testable.

**Reduced motion is a first-class state.** Upstream ignores
`prefers-reduced-motion`. The port switches every animation off under it, and
each component is built so that its resting style is its finished state: text
fully visible, a marquee standing still, a progress bar at its value. The rule
is enforced by tests in the package: every animation `magicui.CSS` defines
must be switched off in a reduced-motion block, every `@keyframes` must be
used, every `animate-*` class in the Go source must be defined, and none may
redefine one of Tailwind's own animations that `shadcn` components rely on.

**Options that become CSS are validated.** Colors and lengths from options go
into `style` as custom properties. A value that could end the declaration or
load a resource (`;`, braces, angle brackets, quotes, backslashes, comment
delimiters, unbalanced parentheses or brackets, `url(`) makes the component
render an error instead.

## Trade-offs

Every replacement above gives something up:

- **No exit animations.** CSS animates an element that exists. Motion's
  `AnimatePresence` animates one that React is about to remove, and nothing on
  a server-rendered page removes elements that way. `TextAnimate` and
  `BlurFade` therefore have entrances only.
- **Random stays still.** Sparkles keep their places; upstream moves each one
  every few seconds with `setInterval`. `AnimatedList` renders all items at
  once and staggers them in; upstream keeps appending forever.
- **Hover replaces one-shot events.** `Text3DFlip` flips while the pointer is
  on it and flips back when it leaves, where upstream flips once on
  `mouseenter`, because a CSS hover transition has no "fire once".
- **Scroll-driven animations are not everywhere.** Browsers without
  `animation-timeline` (Firefox, unless its flag is set) skip the `@supports`
  block and show the final state: `ScrollProgress` is hidden, `TextReveal`
  shows its text solid, `BlurFade` with `InView` shows its content at once.
- **Ids need care.** Without an `ID` option an SVG component draws its id from
  `mx.UniqueID`, a process-wide counter, so two renders differ. Byte-identical
  output needs a fixed id from the caller.
- **The stylesheet is the app's job.** A component without `magicui.CSS`
  renders fine and stands still. That failure is silent: no error, just no
  motion.
- **Some components need a script after all.** Pointer-tracking effects (Magic
  Card, Dock, Lens), canvas effects (Particles, Flickering Grid), number
  counters and typing effects have no CSS equivalent. They are not ported yet;
  [`shadcn/TODOS.md`](https://github.com/ungerik/go-mx/blob/main/shadcn/TODOS.md#magic-ui-magicui)
  lists them with the shared script they would use.

## Alternatives considered

- **Wrap the React components** and hydrate them on the client. That's the
  SPA model go-mx is an alternative to; it would make every Magic UI page
  depend on a JavaScript bundle.
- **Port the motion library to a small script.** Possible for a subset, but
  it would bring per-frame JavaScript and a Content Security Policy question
  to effects that CSS can do natively. The scripted batch planned in
  `shadcn/TODOS.md` is limited to effects CSS cannot express, such as
  following the pointer or drawing on a canvas.
- **Draw randomness in the browser** with a tiny inline script. It would keep
  sparkles moving, but give up deterministic markup, which caching and the
  gallery's golden output rely on.

## Related

- [Tutorial: build an animated landing hero](tutorial.html)
- [How-to guides](how-to.html)
- [Component reference and differences from upstream](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md)
- [How the `shadcn` port replaces Radix with web-platform primitives](../index.html#explanation)
